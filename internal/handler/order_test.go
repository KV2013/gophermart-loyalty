package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/handler/mock"
	"github.com/KV2013/gophermart-loyalty/internal/middleware"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
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

func TestAPIUserCreateOrder(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		uuid       string
		setup      func(svc *mock.MockOrderService)
		wantStatus int
	}{
		{
			name: "202 новый заказ принят",
			body: validOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateOrder(gomock.Any(), int64(1), validOrderNumber).Return(nil)
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "200 заказ уже загружен этим пользователем",
			body: validOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateOrder(gomock.Any(), int64(1), validOrderNumber).
					Return(&model.ErrOrderOwnedByUser{Number: validOrderNumber})
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "400 пустое тело",
			body: "",
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "401 нет uuid в контексте",
			body:       validOrderNumber,
			uuid:       "",
			setup:      func(svc *mock.MockOrderService) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 пользователь не найден",
			body: validOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "401 ошибка при поиске пользователя",
			body: validOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "409 заказ загружен другим пользователем",
			body: validOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateOrder(gomock.Any(), int64(1), validOrderNumber).
					Return(&model.ErrOrderOwnedByOther{Number: validOrderNumber})
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "422 неверный формат номера заказа",
			body: invalidOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "500 внутренняя ошибка сервера",
			body: validOrderNumber,
			uuid: testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().CreateOrder(gomock.Any(), int64(1), validOrderNumber).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockOrderService(ctrl)
			tt.setup(svc)

			h := newOrderHandler(svc)
			req := requestWithUserUUID(http.MethodPost, "/api/user/orders", tt.body, tt.uuid)
			rr := httptest.NewRecorder()

			h.APIUserCreateOrder(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func newOrderHandler(svc *mock.MockOrderService) *handler.OrderHandler {
	return handler.NewOrderHandler(svc, testConfig(), zap.NewNop())
}

func testOrders() []model.Order {
	return []model.Order{
		{ID: 1, UserID: 1, Number: validOrderNumber, Status: "PROCESSED", Accrual: 500, UploadedAt: time.Now()},
		{ID: 2, UserID: 1, Number: "49927398716", Status: "NEW", Accrual: 0, UploadedAt: time.Now()},
	}
}

func TestAPIUserGetOrders(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		uuid        string
		setup       func(svc *mock.MockOrderService)
		wantStatus  int
		wantOrders  bool
	}{
		{
			name:  "200 список заказов (параметры по умолчанию)",
			query: "",
			uuid:  testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserOrders(gomock.Any(), int64(1), 50, 0).Return(testOrders(), nil)
			},
			wantStatus: http.StatusOK,
			wantOrders: true,
		},
		{
			name:  "200 список заказов (page[number]=2&page[size]=10)",
			query: "page[number]=2&page[size]=10",
			uuid:  testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserOrders(gomock.Any(), int64(1), 10, 10).Return(testOrders(), nil)
			},
			wantStatus: http.StatusOK,
			wantOrders: true,
		},
		{
			name:  "204 нет заказов",
			query: "",
			uuid:  testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserOrders(gomock.Any(), int64(1), 50, 0).Return([]model.Order{}, nil)
			},
			wantStatus: http.StatusNoContent,
			wantOrders: false,
		},
		{
			name:       "401 нет uuid в контексте",
			query:      "",
			uuid:       "",
			setup:      func(svc *mock.MockOrderService) {},
			wantStatus: http.StatusUnauthorized,
			wantOrders: false,
		},
		{
			name:  "401 пользователь не найден",
			query: "",
			uuid:  testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
			wantOrders: false,
		},
		{
			name:  "401 ошибка при поиске пользователя",
			query: "",
			uuid:  testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusUnauthorized,
			wantOrders: false,
		},
		{
			name:  "500 ошибка при получении заказов",
			query: "",
			uuid:  testUserUUID,
			setup: func(svc *mock.MockOrderService) {
				svc.EXPECT().FindUserByUUID(gomock.Any(), testUserUUID).Return(testUser(), nil)
				svc.EXPECT().GetUserOrders(gomock.Any(), int64(1), 50, 0).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantOrders: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock.NewMockOrderService(ctrl)
			tt.setup(svc)

			h := newOrderHandler(svc)
			target := "/api/user/orders"
			if tt.query != "" {
				target += "?" + tt.query
			}
			req := requestWithUserUUID(http.MethodGet, target, "", tt.uuid)
			rr := httptest.NewRecorder()

			h.APIUserGetOrders(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rr.Code, tt.wantStatus)
			}

			if tt.wantOrders {
				var orders []model.Order
				if err := json.Unmarshal(rr.Body.Bytes(), &orders); err != nil {
					t.Errorf("не удалось десериализовать тело ответа: %v", err)
				}
				if len(orders) == 0 {
					t.Error("ожидался непустой список заказов")
				}
			}
		})
	}
}
