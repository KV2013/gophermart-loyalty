package repository

import (
	"errors"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	sqlxrepo "github.com/KV2013/gophermart-loyalty/internal/repository/sqlx"
	"go.uber.org/zap"
)

type Repository struct {
	User    *sqlxrepo.UserRepository
	Balance *sqlxrepo.BalanceRepository
	Order   *sqlxrepo.OrderRepository
}

func New(cfg *config.Config, logger *zap.Logger) (*Repository, error) {
	if cfg.DatabaseURI == "" {
		return nil, errors.New("no database URI provided")
	}

	db, err := sqlxrepo.NewConnection(cfg.DatabaseURI)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	logger.Debug("Успешное подключение к БД")

	if err := sqlxrepo.RunMigrations(db); err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	userRepo, err := sqlxrepo.NewUserRepository(db, logger)
	if err != nil {
		return nil, err
	}
	balanceRepo, err := sqlxrepo.NewBalanceRepository(db, logger)
	if err != nil {
		return nil, err
	}
	orderRepo, err := sqlxrepo.NewOrderRepository(db, logger)
	if err != nil {
		return nil, err
	}

	return &Repository{
		User:    userRepo,
		Balance: balanceRepo,
		Order:   orderRepo,
	}, nil
}
