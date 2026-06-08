package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/model"
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
func (h *OrderHandler) APIUserGetOrders(res http.ResponseWriter, req *http.Request) {
	const logPrefix = "OrderHandler.APIUserGetOrders"

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

	const defaultSize = 50
	const defaultPage = 1

	pageSize := defaultSize
	if s := req.URL.Query().Get("page[size]"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			pageSize = v
		}
	}

	pageNumber := defaultPage
	if n := req.URL.Query().Get("page[number]"); n != "" {
		if v, err := strconv.Atoi(n); err == nil && v > 0 {
			pageNumber = v
		}
	}

	offset := (pageNumber - 1) * pageSize

	orders, err := h.service.GetUserOrders(req.Context(), user.ID, pageSize, offset)
	if err != nil {
		h.logger.Error(logPrefix+" получение заказов", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := json.Marshal(orders)
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

// POST /api/user/orders — загрузка пользователем номера заказа для расчёта;
func (h *OrderHandler) APIUserCreateOrder(res http.ResponseWriter, req *http.Request) {
	const logPrefix = "OrderHandler.APIUserCreateOrder"

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

	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		h.logger.Error(logPrefix+" чтение тела запроса", zap.Error(err))
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	number := strings.TrimSpace(string(body))
	if !luhnValid(number) {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	err = h.service.CreateOrder(req.Context(), user.ID, number)
	if err != nil {
		var ownedByUser *model.ErrOrderOwnedByUser
		if errors.As(err, &ownedByUser) {
			res.WriteHeader(http.StatusOK)
			return
		}

		var ownedByOther *model.ErrOrderOwnedByOther
		if errors.As(err, &ownedByOther) {
			res.WriteHeader(http.StatusConflict)
			return
		}

		h.logger.Error(logPrefix+" создание заказа", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// luhnValid проверяет номер заказа по алгоритму Луна.
func luhnValid(number string) bool {
	if len(number) == 0 {
		return false
	}
	sum := 0
	double := false
	for i := len(number) - 1; i >= 0; i-- {
		ch := number[i]
		if ch < '0' || ch > '9' {
			return false
		}
		digit := int(ch - '0')
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}
