package handler

import (
	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/service"
	"go.uber.org/zap"
)

type Handler struct {
	Auth    *AuthHandler
	Balance *BalanceHandler
	Order   *OrderHandler
}

func New(service *service.Service, config *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		Auth:    NewAuthHandler(service.Auth, config, logger),
		Balance: NewBalanceHandler(service.Balance, config, logger),
		Order:   NewOrderHandler(service.Order, config, logger),
	}
}
