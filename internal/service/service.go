package service

import (
	"context"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

// UserRepository — интерфейс репозитория пользователей, определён на стороне потребителя.
type UserRepository interface {
	FindByLogin(ctx context.Context, login string) (*model.User, error)
	FindByCredentials(ctx context.Context, login string, passwordHash string) (*model.User, error)
	FindByUUID(ctx context.Context, uuid string) (*model.User, error)
	UUIDExists(ctx context.Context, uuid string) (bool, error)
	Create(ctx context.Context, login string, passwordHash string) (*model.User, error)
}

// OrderRepository — интерфейс репозитория заказов, определён на стороне потребителя.
type OrderRepository interface {
	FindByNumber(ctx context.Context, number string) (*model.Order, error)
	Create(ctx context.Context, userID int64, number string) error
	GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error)
	UnprocessedOrdersCount(ctx context.Context) (int, error)
	GetUnprocessedOrders(ctx context.Context, limit int) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
}

// BalanceRepository — интерфейс репозитория баланса, определён на стороне потребителя.
type BalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Transaction, error)
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float32) error
	CreateAccrualTransaction(ctx context.Context, orderID int64, userID int64, sum float32) error
}

type Service struct {
	Auth    *authService
	Order   *orderService
	Balance *balanceService
}

func New(userRepo UserRepository, orderRepo OrderRepository, balanceRepo BalanceRepository, logger *zap.Logger, accrualAddress string) *Service {
	return &Service{
		Auth:    NewAuthService(userRepo, logger),
		Order:   NewOrderService(orderRepo, userRepo, logger),
		Balance: NewBalanceService(balanceRepo, userRepo, orderRepo, logger, accrualAddress),
	}
}
