package handler

import (
	"net/http"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	service BalanceService
	config  *config.Config
	logger  *zap.Logger
}

type BalanceService interface{}

func NewBalanceHandler(service BalanceService, config *config.Config, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		service: service,
		config:  config,
		logger:  logger,
	}
}

// GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (h *BalanceHandler) APIUserGetBalance(res http.ResponseWriter, req *http.Request) {}

// GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
func (h *BalanceHandler) APIUserGetWithdrawals(res http.ResponseWriter, req *http.Request) {}

// POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (h *BalanceHandler) APIUserCreateWithdrawal(res http.ResponseWriter, req *http.Request) {}
