package sqlx

import (
	"context"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
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

func (r *BalanceRepository) GetBalance(ctx context.Context, userID int64) (*model.Balance, error) {
	var balance model.Balance
	query := `
		SELECT
		    COALESCE(SUM(t.sum), 0)::float8 AS current,
		    COALESCE(-SUM(t.sum) FILTER (WHERE t.sum < 0), 0)::float8 AS withdrawn
		FROM transactions t
		WHERE t.user_id = $1`
	if err := r.db.GetContext(ctx, &balance, query, userID); err != nil {
		return nil, fmt.Errorf("BalanceRepository.GetBalance: %w", err)
	}
	return &balance, nil
}

func (r *BalanceRepository) GetUserWithdrawals(ctx context.Context, userID int64) ([]model.Transaction, error) {
	var transactions []model.Transaction
	query := `
		SELECT id, order_id, order_number, user_id, ABS(sum) sum, created_at
		FROM transactions
		WHERE user_id = $1 AND sum < 0
		ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &transactions, query, userID); err != nil {
		return nil, fmt.Errorf("BalanceRepository.GetUserWithdrawals: %w", err)
	}
	return transactions, nil
}

func (r *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float32) error {
	query := `INSERT INTO transactions (order_number, user_id, sum) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, orderNumber, userID, -sum); err != nil {
		return fmt.Errorf("BalanceRepository.CreateWithdrawal: %w", err)
	}
	return nil
}

func (r *BalanceRepository) CreateAccrualTransaction(ctx context.Context, orderID int64, userID int64, sum float32) error {
	query := `INSERT INTO transactions (order_id, user_id, sum) VALUES ($1, $2, $3)`
	if _, err := r.db.ExecContext(ctx, query, orderID, userID, sum); err != nil {
		return fmt.Errorf("BalanceRepository.CreateAccrualTransaction: %w", err)
	}
	return nil
}
