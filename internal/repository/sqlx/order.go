package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KV2013/gophermart-loyalty/internal/model"
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

func (r *OrderRepository) FindByNumber(ctx context.Context, number string) (*model.Order, error) {
	var order model.Order
	query := `
		SELECT o.id, o.user_id, u.uuid::text AS user_uuid, o.number, o.status, o.uploaded_at,
		       COALESCE(
			   		(
			   			SELECT SUM(tr.sum)
						FROM transactions AS tr
						WHERE order_id = o.id
						AND tr.sum >= 0
					),
					0
				) AS accrual
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.number = $1`
	err := r.db.GetContext(ctx, &order, query, number)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("OrderRepository.FindByNumber: %w", err)
	}
	return &order, nil
}

func (r *OrderRepository) Create(ctx context.Context, userID int64, number string) error {
	query := `INSERT INTO orders (user_id, number) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, userID, number)
	if err != nil {
		return fmt.Errorf("OrderRepository.Create: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	query := `
		SELECT
				o.id,
				o.user_id,
				o.number,
				o.status,
				o.uploaded_at,
		       	COALESCE(
			   		(
			   			SELECT SUM(tr.sum)
						FROM transactions AS tr
						WHERE order_id = o.id
						AND tr.sum >= 0
					),
					0
				) AS accrual
		FROM orders o
		WHERE o.user_id = $1
		ORDER BY o.uploaded_at DESC
		LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &orders, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("OrderRepository.GetAllByUserID: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) UnprocessedOrdersCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM orders WHERE status NOT IN ('INVALID', 'PROCESSED')`
	if err := r.db.GetContext(ctx, &count, query); err != nil {
		return 0, fmt.Errorf("OrderRepository.UnprocessedOrdersCount: %w", err)
	}
	return count, nil
}

func (r *OrderRepository) GetUnprocessedOrders(ctx context.Context, limit int) ([]model.Order, error) {
	var orders []model.Order
	query := `
		SELECT o.id, o.user_id, o.number, o.status, o.uploaded_at,
			COALESCE(
				(
					SELECT SUM(tr.sum)
					FROM transactions AS tr
					WHERE order_id = o.id
					AND tr.sum >= 0
				),
				0
			) AS accrual
		FROM orders o
		WHERE o.status NOT IN ('INVALID', 'PROCESSED')
		ORDER BY o.id
		LIMIT $1`
	if err := r.db.SelectContext(ctx, &orders, query, limit); err != nil {
		return nil, fmt.Errorf("OrderRepository.GetUnprocessedOrders: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		return fmt.Errorf("OrderRepository.UpdateOrderStatus: %w", err)
	}
	return nil
}
