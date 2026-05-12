package service

import (
	"context"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type orderService struct {
	repo     OrderRepository
	userRepo UserRepository
	logger   *zap.Logger
}

func NewOrderService(orderRepository OrderRepository, userRepository UserRepository, logger *zap.Logger) *orderService {
	return &orderService{
		repo:     orderRepository,
		userRepo: userRepository,
		logger:   logger,
	}
}

func (s *orderService) FindUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
	return s.userRepo.FindByUUID(ctx, uuid)
}

func (s *orderService) GetUserOrders(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error) {
	orders, err := s.repo.GetAllByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("orderService.GetUserOrders: %w", err)
	}
	return orders, nil
}

func (s *orderService) CreateOrder(ctx context.Context, userID int64, number string) error {
	existing, err := s.repo.FindByNumber(ctx, number)
	if err != nil {
		return fmt.Errorf("orderService.CreateOrder: %w", err)
	}

	if existing != nil {
		if existing.UserID == userID {
			return &model.ErrOrderOwnedByUser{Number: number}
		}
		return &model.ErrOrderOwnedByOther{Number: number}
	}

	if err := s.repo.Create(ctx, userID, number); err != nil {
		return fmt.Errorf("orderService.CreateOrder: %w", err)
	}

	return nil
}
