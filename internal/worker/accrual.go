package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

type OrderRepository interface {
	UnprocessedOrdersCount(ctx context.Context) (int, error)
	GetUnprocessedOrders(ctx context.Context, limit int) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
}

type BalanceRepository interface {
	CreateAccrualTransaction(ctx context.Context, orderID int64, userID int64, sum float32) error
}

//easyjson:json
type accrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float32 `json:"accrual"`
}

type orderResult struct {
	OrderID int64
	UserID  int64
	Status  string
	Accrual float32
	Err     error
}

type AccrualWorker struct {
	orderRepo   OrderRepository
	balanceRepo BalanceRepository
	logger      *zap.Logger
	accrualAddr string
	httpClient  *http.Client
}

func New(orderRepo OrderRepository, balanceRepo BalanceRepository, logger *zap.Logger, accrualAddress string) *AccrualWorker {
	return &AccrualWorker{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
		logger:      logger,
		accrualAddr: accrualAddress,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *AccrualWorker) Start(ctx context.Context) {
	const (
		tickInterval = 3 * time.Second
		batchSize    = 50
		numWorkers   = 5
	)

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	w.logger.Info("AccrualWorker: запуск очереди начисления баллов")
	queue := newOrderQueue(batchSize)

	for i := 1; i <= numWorkers; i++ {
		go w.processQueueJobs(i, ctx, queue)
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("AccrualWorker: остановка очереди начисления баллов")
			return
		case <-ticker.C:
			count, err := w.orderRepo.UnprocessedOrdersCount(ctx)
			if err != nil {
				w.logger.Error("AccrualWorker: ошибка получения количества необработанных заказов", zap.Error(err))
				continue
			}
			if count == 0 {
				continue
			}

			orders, err := w.orderRepo.GetUnprocessedOrders(ctx, batchSize)
			if err != nil {
				w.logger.Error("AccrualWorker: ошибка получения необработанных заказов", zap.Error(err))
				continue
			}

			for _, order := range orders {
				queue.AddJob(QueueJob{
					OrderID:     order.ID,
					UserID:      order.UserID,
					Type:        "FETCH_ACCRUAL",
					OrderNumber: order.Number,
				})
			}
		}
	}
}

func (w *AccrualWorker) processQueueJobs(id int, ctx context.Context, queue *OrderQueue) {
	w.logger.Debug("AccrualWorker: запуск воркера", zap.Int("id", id))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			job := queue.GetJob()
			switch job.Type {
			case "FETCH_ACCRUAL":
				accrualResp, err := w.fetchAccrual(job.OrderNumber)
				if err != nil {
					w.logger.Error("AccrualWorker: ошибка получения начислений", zap.Error(err), zap.Int64("OrderID", job.OrderID))
					continue
				}
				if accrualResp == nil {
					continue
				}
				queue.AddJob(QueueJob{
					OrderID: job.OrderID,
					UserID:  job.UserID,
					Type:    "UPDATE_ORDER",
					OrderResult: &orderResult{
						OrderID: job.OrderID,
						UserID:  job.UserID,
						Accrual: accrualResp.Accrual,
						Status:  accrualResp.Status,
					},
				})
			case "UPDATE_ORDER":
				w.updatePoints(ctx, job.OrderResult)
			}
		}
	}
}

func (w *AccrualWorker) fetchAccrual(orderNumber string) (*accrualResponse, error) {
	url := fmt.Sprintf("%s/api/orders/%s", w.accrualAddr, orderNumber)

	resp, err := w.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, fmt.Errorf("заказ не зарегистрирован в системе начисления баллов")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("превышен лимит запросов к сервису начисления баллов)")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("от сервиса начисления баллов получен статус ошибки %d", resp.StatusCode)
	}

	var result accrualResponse
	if err := easyjson.UnmarshalFromReader(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("декоидирование ответа от сервиса начисления баллов вызвало ошику: %w", err)
	}

	return &result, nil
}

func (w *AccrualWorker) updatePoints(ctx context.Context, result *orderResult) {
	switch result.Status {
	case "PROCESSED":
		if err := w.orderRepo.UpdateOrderStatus(ctx, result.OrderID, "PROCESSED"); err != nil {
			w.logger.Error("AccrualWorker: ошибка обновления статуса заказа",
				zap.Int64("orderID", result.OrderID),
				zap.Error(err),
			)
			return
		}
		if result.Accrual > 0 {
			if err := w.balanceRepo.CreateAccrualTransaction(ctx, result.OrderID, result.UserID, result.Accrual); err != nil {
				w.logger.Error("AccrualWorker: ошибка создания транзакции начисления",
					zap.Int64("orderID", result.OrderID),
					zap.Float32("accrual", result.Accrual),
					zap.Error(err),
				)
			}
		}
	case "INVALID":
		if err := w.orderRepo.UpdateOrderStatus(ctx, result.OrderID, "INVALID"); err != nil {
			w.logger.Error("AccrualWorker: ошибка обновления статуса заказа",
				zap.Int64("orderID", result.OrderID),
				zap.Error(err),
			)
		}
	case "PROCESSING":
		if err := w.orderRepo.UpdateOrderStatus(ctx, result.OrderID, "PROCESSING"); err != nil {
			w.logger.Error("AccrualWorker: ошибка обновления статуса заказа",
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
	OrderID     int64
	UserID      int64
	Type        string
	OrderNumber string
	OrderResult *orderResult
}
