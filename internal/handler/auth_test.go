package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/handler/mock"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

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

func TestAPIUserRegister(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mock.MockAuthService)
		wantStatus int
	}{
		{
			name: "200 успешная регистрация",
			body: `{"login":"alice","password":"secret"}`,
			setup: func(svc *mock.MockAuthService) {
				svc.EXPECT().LoginExists(gomock.Any(), "alice").Return(false, nil)
				svc.EXPECT().Register(gomock.Any(), "alice", "secret").Return(testUser(), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 невалидный JSON",
			body:       `not-json`,
			setup:      func(svc *mock.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 пустой login",
			body:       `{"login":"","password":"secret"}`,
			setup:      func(svc *mock.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 пустой password",
			body:       `{"login":"alice","password":""}`,
			setup:      func(svc *mock.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "409 логин уже занят",
			body: `{"login":"alice","password":"secret"}`,
			setup: func(svc *mock.MockAuthService) {
				svc.EXPECT().LoginExists(gomock.Any(), "alice").Return(true, nil)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "500 ошибка при регистрации",
			body: `{"login":"alice","password":"secret"}`,
			setup: func(svc *mock.MockAuthService) {
				svc.EXPECT().LoginExists(gomock.Any(), "alice").Return(false, nil)
				svc.EXPECT().Register(gomock.Any(), "alice", "secret").Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockAuthService(ctrl)
			tt.setup(svc)

			h := handler.NewAuthHandler(svc, testConfig(), zap.NewNop())
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			h.APIUserRegister(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestAPIUserLogin(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(svc *mock.MockAuthService)
		wantStatus int
	}{
		{
			name: "200 успешная аутентификация",
			body: `{"login":"alice","password":"secret"}`,
			setup: func(svc *mock.MockAuthService) {
				svc.EXPECT().Authenticate(gomock.Any(), "alice", "secret").Return(testUser(), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 невалидный JSON",
			body:       `not-json`,
			setup:      func(svc *mock.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 пустой login",
			body:       `{"login":"","password":"secret"}`,
			setup:      func(svc *mock.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "400 пустой password",
			body:       `{"login":"alice","password":""}`,
			setup:      func(svc *mock.MockAuthService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "401 неверные учётные данные",
			body: `{"login":"alice","password":"wrong"}`,
			setup: func(svc *mock.MockAuthService) {
				svc.EXPECT().Authenticate(gomock.Any(), "alice", "wrong").Return(nil, errors.New("invalid credentials"))
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockAuthService(ctrl)
			tt.setup(svc)

			h := handler.NewAuthHandler(svc, testConfig(), zap.NewNop())
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()

			h.APIUserLogin(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
