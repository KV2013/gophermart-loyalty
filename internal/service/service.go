package service

import (
	"github.com/KV2013/gophermart-loyalty/internal/repository"
	"github.com/KV2013/gophermart-loyalty/internal/service/auth"
	"go.uber.org/zap"
)

type Service struct {
	Auth    *auth.AuthService
	Order   *OrderService
	Balance *BalanceService
}

func New(repo *repository.Repository, logger *zap.Logger) *Service {
	return &Service{
		Auth:    auth.NewAuthService(repo.User, logger),
		Order:   NewOrderService(repo.Order, logger),
		Balance: NewBalanceService(repo.Balance, logger),
	}
}
