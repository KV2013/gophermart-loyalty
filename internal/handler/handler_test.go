package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/middleware"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/google/uuid"
)

const (
	validOrderNumber   = "79927398713" // проходит алгоритм Луна
	invalidOrderNumber = "12345678901" // не проходит алгоритм Луна
	testUserUUID       = "550e8400-e29b-41d4-a716-446655440000"
)

func requestWithUserUUID(method, path, body, uuid string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, uuid)
	return req.WithContext(ctx)
}

func testConfig() *config.Config {
	return &config.Config{JWTSecretKey: "test-secret-key"}
}

func testUser() *model.User {
	return &model.User{
		ID:        1,
		UUID:      uuid.New(),
		Login:     "alice",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
