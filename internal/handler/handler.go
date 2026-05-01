package handler

import (
	"net/http"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"go.uber.org/zap"
)

type URLHandler struct {
	config *config.Config
	logger *zap.Logger
}

func New(config *config.Config, logger *zap.Logger) *URLHandler {
	return &URLHandler{
		config: config,
		logger: logger,
	}
}

// POST /api/user/register — регистрация пользователя;
func (h *URLHandler) APIUserRegister(res http.ResponseWriter, req *http.Request) {}

// POST /api/user/login — аутентификация пользователя;
func (h *URLHandler) APIUserLogin(res http.ResponseWriter, req *http.Request) {}

// POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (h *URLHandler) APIUserCreateOrder(res http.ResponseWriter, req *http.Request) {}

// GET /api/user/orders — получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях;
func (h *URLHandler) APIUserGetOrders(res http.ResponseWriter, req *http.Request) {}

// GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (h *URLHandler) APIUserGetBalance(res http.ResponseWriter, req *http.Request) {}

// POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (h *URLHandler) APIUserCreateWithdrawal(res http.ResponseWriter, req *http.Request) {}

// GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
func (h *URLHandler) APIUserGetWithdrawals(res http.ResponseWriter, req *http.Request) {}
