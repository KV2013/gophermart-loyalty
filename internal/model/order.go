package model

import "time"

type Order struct {
	ID         int64     `db:"id"          json:"id"`
	UserID     int64     `db:"user_id"     json:"-"`
	UserUUID   string    `db:"user_uuid"   json:"-"`
	Number     string    `db:"number"      json:"number"`
	Status     string    `db:"status"      json:"status"`
	Accrual    int       `db:"accrual"     json:"accrual"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}
