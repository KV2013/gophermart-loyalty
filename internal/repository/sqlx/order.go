package sqlx

import (
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type OrderRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

func NewOrderRepository(db *sqlx.DB, logger *zap.Logger) (*OrderRepository, error) {

	repo := &OrderRepository{db: db, logger: logger}

	return repo, nil
}
