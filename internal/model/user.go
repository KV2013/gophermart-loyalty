package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int64     `db:"id" json:"id"`
	UUID      uuid.UUID `db:"uuid" json:"uuid"`
	Login     string    `db:"login" json:"login"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
