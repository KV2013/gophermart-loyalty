package model

import "time"

type Transaction struct {
	ID          int64     `db:"id"`
	OrderID     *int64    `db:"order_id"`
	OrderNumber *string   `db:"order_number"`
	UserID      int64     `db:"user_id"`
	Sum         float64   `db:"sum"`
	CreatedAt   time.Time `db:"created_at"`
}
