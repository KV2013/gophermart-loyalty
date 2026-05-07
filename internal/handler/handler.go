package handler

import (
	"context"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type AuthService interface {
	LoginExists(ctx context.Context, login string) (bool, error)
	Register(ctx context.Context, login, password string) (*model.User, error)
	Authenticate(ctx context.Context, login, password string) (*model.User, error)
}

type BalanceService interface{}

type OrderService interface{}

type Handler struct {
	Auth    *AuthHandler
	Balance *BalanceHandler
	Order   *OrderHandler
}

func New(authSvc AuthService, balanceSvc BalanceService, orderSvc OrderService, config *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		Auth:    NewAuthHandler(authSvc, config, logger),
		Balance: NewBalanceHandler(balanceSvc, config, logger),
		Order:   NewOrderHandler(orderSvc, config, logger),
	}
}
