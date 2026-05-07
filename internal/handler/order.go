package handler

import (
	"net/http"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"go.uber.org/zap"
)

type OrderHandler struct {
	service OrderService
	config  *config.Config
	logger  *zap.Logger
}

func NewOrderHandler(service OrderService, config *config.Config, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		config:  config,
		logger:  logger,
	}
}

// GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
func (h *OrderHandler) APIUserGetOrders(res http.ResponseWriter, req *http.Request) {}

// POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (h *OrderHandler) APIUserCreateOrder(res http.ResponseWriter, req *http.Request) {}
