package repository

import (
	"errors"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	sqlxrepo "github.com/KV2013/gophermart-loyalty/internal/repository/sqlx"
	"go.uber.org/zap"
)

type Repository1 interface{}

func NewRepository(cfg *config.Config, logger *zap.Logger) (*sqlxrepo.SQLXRepository, error) {
	if cfg.DatabaseURI == "" {
		return nil, errors.New("no database URI provided")
	}

	return sqlxrepo.NewRepository(cfg.DatabaseURI, logger)
}
