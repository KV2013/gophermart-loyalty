package service

import "go.uber.org/zap"

type OrderRepository interface{}

type OrderService struct {
	repo   OrderRepository
	logger *zap.Logger
}

func NewOrderService(orderRepository OrderRepository, logger *zap.Logger) *OrderService {
	return &OrderService{
		repo:   orderRepository,
		logger: logger,
	}
}
