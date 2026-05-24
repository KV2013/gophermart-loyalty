package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type orderResult struct {
	OrderID int64
	UserID  int64
	Status  string
	Accrual float64
	Err     error
}

type balanceService struct {
	repo           BalanceRepository
	userRepo       UserRepository
	orderRepo      OrderRepository
	logger         *zap.Logger
	accrualAddress string
	httpClient     *http.Client
}

func NewBalanceService(balanceRepository BalanceRepository, userRepository UserRepository, orderRepository OrderRepository, logger *zap.Logger, accrualAddress string) *balanceService {
	svc := &balanceService{
		repo:           balanceRepository,
		userRepo:       userRepository,
		orderRepo:      orderRepository,
		logger:         logger,
		accrualAddress: accrualAddress,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}

	return svc
}

func (s *balanceService) FindUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
	return s.userRepo.FindByUUID(ctx, uuid)
}

func (s *balanceService) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("balanceService.GetBalance: %w", err)
	}
	return balance, nil
}

func (s *balanceService) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Transaction, error) {
	withdrawals, err := s.repo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("balanceService.GetUserWithdrawals: %w", err)
	}
	return withdrawals, nil
}

func (s *balanceService) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("balanceService.CreateWithdrawal: %w", err)
	}

	if balance.Current < sum {
		return &model.ErrInsufficientBalance{}
	}

	if err := s.repo.CreateWithdrawal(ctx, userID, orderNumber, sum); err != nil {
		return fmt.Errorf("balanceService.CreateWithdrawal: %w", err)
	}

	return nil
}

func (s *balanceService) StartPointsAccrualQueue(mainCtx context.Context) {
	const (
		tickInterval = 3 * time.Second
		batchSize    = 50
		numWorkers   = 5
	)

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	s.logger.Info("BalanceService: запуск очереди начисления баллов")
	queue := newOrderQueue(batchSize)
	for w := 1; w <= numWorkers; w++ {
		go s.processQueueJobs(w, mainCtx, queue)
	}
	for {
		select {
		case <-mainCtx.Done():
			s.logger.Info("BalanceService: остановка очереди начисления баллов")
			return
		case <-ticker.C:

			count, err := s.orderRepo.UnprocessedOrdersCount(mainCtx)
			if err != nil {
				s.logger.Error("BalanceService: ошибка получения количества необработанных заказов", zap.Error(err))
				continue
			}
			if count == 0 {
				continue
			}
			s.logger.Debug("BalanceService: заказов к обработке", zap.Int("count", count))

			orders, err := s.orderRepo.GetUnprocessedOrders(mainCtx, batchSize)
			if err != nil {
				s.logger.Error("BalanceService: ошибка получения необработанных заказов", zap.Error(err))
				continue
			}
			s.logger.Debug("BalanceService: размер пачки заказов к обработке", zap.Int("count", len(orders)))

			for _, order := range orders {
				job := QueueJob{
					OrderID: order.ID,
					UserID:  order.UserID,
					Type:    "FETCH_ACCRUAL",
					Payload: map[string]interface{}{"orderNumber": order.Number},
				}
				queue.AddJob(job)
			}
		}
	}
}

func (s *balanceService) processQueueJobs(id int, ctx context.Context, queue *OrderQueue) {
	s.logger.Debug("BalanceService: запуск воркера", zap.Int("id", id))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job := queue.GetJob()
			switch job.Type {
			case "FETCH_ACCRUAL":
				accrualResponse, err := s.fetchAccrual(job.Payload["orderNumber"].(string))
				if err != nil {
					s.logger.Error("BalanceService: ошибка получения начислений", zap.Error(err), zap.Int64("OrderID", job.OrderID))
					continue
				}
				if accrualResponse == nil {
					continue
				}
				newJob := QueueJob{
					OrderID: job.OrderID,
					UserID:  job.UserID,
					Type:    "UPDATE_ORDER",
					Payload: map[string]interface{}{"orderResult": &orderResult{
						OrderID: job.OrderID,
						UserID:  job.UserID,
						Accrual: accrualResponse.Accrual,
						Status:  accrualResponse.Status,
					}},
				}
				queue.AddJob(newJob)
			case "UPDATE_ORDER":
				orderResult, ok := job.Payload["orderResult"].(*orderResult)
				if !ok {
					s.logger.Error("BalanceService: ошибка преобразования orderResult", zap.Int64("OrderID", job.OrderID))
					continue
				}
				s.updatePoints(ctx, orderResult)
			}
		}
	}
}

func (s *balanceService) fetchAccrual(orderNumber string) (*accrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", s.accrualAddress, orderNumber)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("Заказ не зарегистрирован в системе начисления баллов")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("Превышен лимит запросов к сервису начисления баллов)")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("От сервиса начисления баллов получен статус ошибки %d", resp.StatusCode)
	}

	var result accrualResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { // TODO: переделать в easyjson
		return nil, fmt.Errorf("Декоидирование ответа от сервиса начисления баллов вызвало ошику: %w", err)
	}

	return &result, nil
}

func (s *balanceService) updatePoints(ctx context.Context, result *orderResult) {
	switch result.Status {
	case "PROCESSED":
		// TODO: обернуть оба запроса в транзакцию
		if err := s.orderRepo.UpdateOrderStatus(ctx, result.OrderID, "PROCESSED"); err != nil {
			s.logger.Error("BalanceService: ошибка обновления статуса заказа",
				zap.Int64("orderID", result.OrderID),
				zap.Error(err),
			)
			return
		}
		if result.Accrual > 0 {
			if err := s.repo.CreateAccrualTransaction(ctx, result.OrderID, result.UserID, result.Accrual); err != nil {
				s.logger.Error("BalanceService: ошибка создания транзакции начисления",
					zap.Int64("orderID", result.OrderID),
					zap.Float64("accrual", result.Accrual),
					zap.Error(err),
				)
			}
		}
	case "INVALID":
		if err := s.orderRepo.UpdateOrderStatus(ctx, result.OrderID, "INVALID"); err != nil {
			s.logger.Error("BalanceService: ошибка обновления статуса заказа",
				zap.Int64("orderID", result.OrderID),
				zap.Error(err),
			)
		}
	case "PROCESSING":
		if err := s.orderRepo.UpdateOrderStatus(ctx, result.OrderID, "PROCESSING"); err != nil {
			s.logger.Error("BalanceService: ошибка обновления статуса заказа",
				zap.Int64("orderID", result.OrderID),
				zap.Error(err),
			)
		}
	}
}

type OrderQueue struct {
	jobsCh chan QueueJob
}

func newOrderQueue(capacity int) *OrderQueue {
	return &OrderQueue{
		jobsCh: make(chan QueueJob, capacity),
	}
}

func (q *OrderQueue) AddJob(job QueueJob) {
	q.jobsCh <- job
}

func (q *OrderQueue) GetJob() QueueJob {
	return <-q.jobsCh
}

type QueueJob struct {
	OrderID int64
	UserID  int64
	Type    string
	Payload map[string]interface{}
}
