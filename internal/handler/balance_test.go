package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/handler/mock"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newBalanceHandler(svc *mock.MockBalanceService) *handler.BalanceHandler {
	return handler.NewBalanceHandler(svc, testConfig(), zap.NewNop())
}

func testBalance() *model.Balance {
	return &model.Balance{Current: 500.0, Withdrawn: 100.0}
}

func testTransactions() []model.Transaction {
	orderNumber := "79927398713"
	now := time.Now()
	return []model.Transaction{
		{ID: 1, OrderNumber: &orderNumber, UserID: 1, Sum: 50.0, CreatedAt: now},
		{ID: 2, OrderNumber: &orderNumber, UserID: 1, Sum: 30.0, CreatedAt: now},
	}
}

func TestAPIUserGetBalance(t *testing.T) {
	tests := []struct {
		name        string
		uuid        string
		setup       func(svc *mock.MockBalanceService)
		wantStatus  int
		wantBalance *model.Balance
	}{
		{
			name: "200 успешное получение баланса",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(testBalance(), nil)
			},
			wantStatus:  http.StatusOK,
			wantBalance: testBalance(),
		},
		{
			name:       "401 нет uuid в контексте",
			uuid:       "",
			setup:      func(svc *mock.MockBalanceService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 пользователь не найден",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 ошибка при поиске пользователя",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 ошибка при получении баланса",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetBalance(gomock.Any(), int64(1)).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockBalanceService(ctrl)
			tt.setup(svc)

			h := newBalanceHandler(svc)
			req := requestWithUserUUID(http.MethodGet, "/api/user/balance", "", tt.uuid)
			rr := httptest.NewRecorder()

			h.APIUserGetBalance(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantBalance != nil {
				var balance model.Balance
				if err := json.Unmarshal(rr.Body.Bytes(), &balance); err != nil {
					t.Errorf("не удалось десериализовать тело ответа: %v", err)
				}
				if balance.Current != tt.wantBalance.Current {
					t.Errorf("got Current %f, want %f", balance.Current, tt.wantBalance.Current)
				}
				if balance.Withdrawn != tt.wantBalance.Withdrawn {
					t.Errorf("got Withdrawn %f, want %f", balance.Withdrawn, tt.wantBalance.Withdrawn)
				}
			}
		})
	}
}

type withdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float64 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

func TestAPIUserGetWithdrawals(t *testing.T) {
	tests := []struct {
		name            string
		uuid            string
		setup           func(svc *mock.MockBalanceService)
		wantStatus      int
		wantWithdrawals bool
	}{
		{
			name: "200 список списаний",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserWithdrawals(gomock.Any(), int64(1)).Return(testTransactions(), nil)
			},
			wantStatus:      http.StatusOK,
			wantWithdrawals: true,
		},
		{
			name: "204 нет списаний",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserWithdrawals(gomock.Any(), int64(1)).Return([]model.Transaction{}, nil)
			},
			wantStatus:      http.StatusNoContent,
			wantWithdrawals: false,
		},
		{
			name:       "401 нет uuid в контексте",
			uuid:       "",
			setup:      func(svc *mock.MockBalanceService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 пользователь не найден",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 ошибка при поиске пользователя",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "500 ошибка при получении списаний",
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserWithdrawals(gomock.Any(), int64(1)).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockBalanceService(ctrl)
			tt.setup(svc)

			h := newBalanceHandler(svc)
			req := requestWithUserUUID(http.MethodGet, "/api/user/withdrawals", "", tt.uuid)
			rr := httptest.NewRecorder()

			h.APIUserGetWithdrawals(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantWithdrawals {
				var withdrawals []withdrawalResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &withdrawals); err != nil {
					t.Errorf("не удалось десериализовать тело ответа: %v", err)
				}
				if len(withdrawals) != 2 {
					t.Errorf("got %d withdrawals, want 2", len(withdrawals))
				}
				if withdrawals[0].Sum != 50.0 {
					t.Errorf("got Sum %f, want %f", withdrawals[0].Sum, 50.0)
				}
				if withdrawals[1].Sum != 30.0 {
					t.Errorf("got Sum %f, want %f", withdrawals[1].Sum, 30.0)
				}
			}
		})
	}
}

func TestAPIUserCreateWithdrawal(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		uuid       string
		setup      func(svc *mock.MockBalanceService)
		wantStatus int
	}{
		{
			name: "200 успешное списание",
			body: `{"order":"79927398713","sum":100}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(100)).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 нет uuid в контексте",
			body:       `{"order":"79927398713","sum":100}`,
			uuid:       "",
			setup:      func(svc *mock.MockBalanceService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 пользователь не найден",
			body: `{"order":"79927398713","sum":100}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 ошибка при поиске пользователя",
			body: `{"order":"79927398713","sum":100}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "400 невалидный JSON",
			body: `not-json`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "400 пустой номер заказа",
			body: `{"order":"","sum":100}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "400 отрицательная сумма",
			body: `{"order":"79927398713","sum":-10}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "400 нулевая сумма",
			body: `{"order":"79927398713","sum":0}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "422 неверный номер заказа",
			body: `{"order":"12345678901","sum":100}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "402 недостаточно средств",
			body: `{"order":"79927398713","sum":1000}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(1000)).
					Return(&model.ErrInsufficientBalance{})
			},
			wantStatus: http.StatusPaymentRequired,
		},
		{
			name: "500 ошибка сервиса",
			body: `{"order":"79927398713","sum":100}`,
			uuid: testUserUUID,
			setup: func(svc *mock.MockBalanceService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateWithdrawal(gomock.Any(), int64(1), "79927398713", float64(100)).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockBalanceService(ctrl)
			tt.setup(svc)

			h := newBalanceHandler(svc)
			req := requestWithUserUUID(http.MethodPost, "/api/user/balance/withdraw", tt.body, tt.uuid)
			rr := httptest.NewRecorder()

			h.APIUserCreateWithdrawal(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}
