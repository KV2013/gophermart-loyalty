package service

import "go.uber.org/zap"

type orderService struct {
	repo   OrderRepository
	logger *zap.Logger
}

func NewOrderService(orderRepository OrderRepository, logger *zap.Logger) *orderService {
	return &orderService{
		repo:   orderRepository,
		logger: logger,
	}
}
