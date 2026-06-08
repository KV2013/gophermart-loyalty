package service

import (
	"context"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/zap"
)

type balanceService struct {
	repo     BalanceRepository
	userRepo UserRepository
	logger   *zap.Logger
}

func NewBalanceService(balanceRepository BalanceRepository, userRepository UserRepository, logger *zap.Logger) *balanceService {
	return &balanceService{
		repo:     balanceRepository,
		userRepo: userRepository,
		logger:   logger,
	}
}

func (s *balanceService) FindUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
	return s.userRepo.FindByUUID(ctx, uuid)
}

func (s *balanceService) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("balanceService.GetBalance: %w", err)
	}
	return balance, nil
}

func (s *balanceService) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Transaction, error) {
	withdrawals, err := s.repo.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("balanceService.GetUserWithdrawals: %w", err)
	}
	return withdrawals, nil
}

func (s *balanceService) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float32) error {
	balance, err := s.repo.GetBalance(ctx, userID)
	if err != nil {
		return fmt.Errorf("balanceService.CreateWithdrawal: %w", err)
	}

	if balance.Current < sum {
		return &model.ErrInsufficientBalance{}
	}

	if err := s.repo.CreateWithdrawal(ctx, userID, orderNumber, sum); err != nil {
		return fmt.Errorf("balanceService.CreateWithdrawal: %w", err)
	}

	return nil
}
