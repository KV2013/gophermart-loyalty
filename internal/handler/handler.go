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

type BalanceService interface {
	FindUserByUUID(ctx context.Context, uuid string) (*model.User, error)
	GetBalance(ctx context.Context, userID int64) (*model.Balance, error)
	GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Transaction, error)
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error
}

type OrderService interface {
	FindUserByUUID(ctx context.Context, uuid string) (*model.User, error)
	CreateOrder(ctx context.Context, userID int64, number string) error
	GetUserOrders(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error)
}

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
