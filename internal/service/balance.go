package service

import "go.uber.org/zap"

type BalanceRepository interface{}

type BalanceService struct {
	repo   BalanceRepository
	logger *zap.Logger
}

func NewBalanceService(balanceRepository BalanceRepository, logger *zap.Logger) *BalanceService {
	return &BalanceService{
		repo:   balanceRepository,
		logger: logger,
	}
}
