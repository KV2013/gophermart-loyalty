package service

import "go.uber.org/zap"

type balanceService struct {
	repo   BalanceRepository
	logger *zap.Logger
}

func NewBalanceService(balanceRepository BalanceRepository, logger *zap.Logger) *balanceService {
	return &balanceService{
		repo:   balanceRepository,
		logger: logger,
	}
}
