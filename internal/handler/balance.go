package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/model"
	"github.com/mailru/easyjson"
	"go.uber.org/zap"
)

//easyjson:json
type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float32 `json:"sum"`
}

//easyjson:json
type WithdrawalResponse struct {
	Order       string  `json:"order"`
	Sum         float32 `json:"sum"`
	ProcessedAt string  `json:"processed_at"`
}

type BalanceHandler struct {
	service BalanceService
	config  *config.Config
	logger  *zap.Logger
}

func NewBalanceHandler(service BalanceService, config *config.Config, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		service: service,
		config:  config,
		logger:  logger,
	}
}

// GET /api/user/balance — получение текущего баланса счёта баллов лояльности пользователя;
func (h *BalanceHandler) APIUserGetBalance(res http.ResponseWriter, req *http.Request) {
	const logPrefix = "BalanceHandler.APIUserGetBalance"

	userUUID, ok := userUUIDFromRequest(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.service.FindUserByUUID(req.Context(), userUUID)
	if err != nil || user == nil {
		h.logger.Error(logPrefix+" поиск пользователя", zap.Error(err))
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	balance, err := h.service.GetBalance(req.Context(), user.ID)
	if err != nil {
		h.logger.Error(logPrefix+" получение баланса", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(balance)
	if err != nil {
		h.logger.Error(logPrefix+" сериализация ответа", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write(body); err != nil {
		h.logger.Error(logPrefix+" запись ответа", zap.Error(err))
	}
}

// GET /api/user/withdrawals — получение информации о выводе средств с накопительного счёта пользователем.
func (h *BalanceHandler) APIUserGetWithdrawals(res http.ResponseWriter, req *http.Request) {
	const logPrefix = "BalanceHandler.APIUserGetWithdrawals"

	userUUID, ok := userUUIDFromRequest(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.service.FindUserByUUID(req.Context(), userUUID)
	if err != nil || user == nil {
		h.logger.Error(logPrefix+" поиск пользователя", zap.Error(err))
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	transactions, err := h.service.GetUserWithdrawals(req.Context(), user.ID)
	if err != nil {
		h.logger.Error(logPrefix+" получение списаний", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(transactions) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	responses := make([]WithdrawalResponse, len(transactions))
	for i, t := range transactions {
		order := ""
		if t.OrderNumber != nil {
			order = *t.OrderNumber
		}
		responses[i] = WithdrawalResponse{
			Order:       order,
			Sum:         t.Sum,
			ProcessedAt: t.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		}
	}

	var jsonElements []json.RawMessage
	for _, resp := range responses {
		data, err := easyjson.Marshal(&resp)
		if err != nil {
			h.logger.Error(logPrefix+" сериализация ответа", zap.Error(err))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		jsonElements = append(jsonElements, json.RawMessage(data))
	}

	body, err := json.Marshal(jsonElements)
	if err != nil {
		h.logger.Error(logPrefix+" сериализация ответа", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	if _, err = res.Write(body); err != nil {
		h.logger.Error(logPrefix+" запись ответа", zap.Error(err))
	}
}

// POST /api/user/balance/withdraw — запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа;
func (h *BalanceHandler) APIUserCreateWithdrawal(res http.ResponseWriter, req *http.Request) {
	const logPrefix = "BalanceHandler.APIUserCreateWithdrawal"

	userUUID, ok := userUUIDFromRequest(req)
	if !ok {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	user, err := h.service.FindUserByUUID(req.Context(), userUUID)
	if err != nil || user == nil {
		h.logger.Error(logPrefix+" поиск пользователя", zap.Error(err))
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		h.logger.Error(logPrefix+" чтение тела запроса", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	var withdrawalRequest WithdrawalRequest
	if err := easyjson.Unmarshal(reqBody, &withdrawalRequest); err != nil {
		h.logger.Error(logPrefix+" разбор запроса", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if withdrawalRequest.Order == "" || withdrawalRequest.Sum <= 0 {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if !luhnValid(withdrawalRequest.Order) {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = h.service.CreateWithdrawal(req.Context(), user.ID, withdrawalRequest.Order, withdrawalRequest.Sum)
	if err != nil {
		var insufficientBalance *model.ErrInsufficientBalance
		if errors.As(err, &insufficientBalance) {
			res.WriteHeader(http.StatusPaymentRequired)
			return
		}

		h.logger.Error(logPrefix+" создание списания", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
