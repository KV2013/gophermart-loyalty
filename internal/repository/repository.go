package repository

import (
	"errors"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	sqlxrepo "github.com/KV2013/gophermart-loyalty/internal/repository/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
}

func New(cfg *config.Config, logger *zap.Logger) (Repository, error) {
	if cfg.DatabaseURI != "" {
		return sqlxrepo.NewRepository(cfg.DatabaseURI, logger)
	}
	return nil, errors.New("no database URI provided")
}
