package sqlx

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type BalanceRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewBalanceRepository(db *sqlx.DB, logger *zap.Logger) (*BalanceRepository, error) {

	repo := &BalanceRepository{db: db, logger: logger}
	return repo, nil
}
