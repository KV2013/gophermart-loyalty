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
	Create(ctx context.Context, login string, passwordHash string) (*model.User, error)
}

// OrderRepository — интерфейс репозитория заказов, определён на стороне потребителя.
type OrderRepository interface{}

// BalanceRepository — интерфейс репозитория баланса, определён на стороне потребителя.
type BalanceRepository interface{}

type Service struct {
	Auth    *authService
	Order   *orderService
	Balance *balanceService
}

func New(userRepo UserRepository, orderRepo OrderRepository, balanceRepo BalanceRepository, logger *zap.Logger) *Service {
	return &Service{
		Auth:    NewAuthService(userRepo, logger),
		Order:   NewOrderService(orderRepo, logger),
		Balance: NewBalanceService(balanceRepo, logger),
	}
}
