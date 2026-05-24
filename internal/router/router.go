package router

import (
	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Init(
	h *handler.Handler,
	authService middleware.AuthService,
	logger *zap.Logger,
	cfg *config.Config,
) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.API(logger))
	r.Use(middleware.GzipCompression)

	r.Post("/api/user/register", h.Auth.APIUserRegister)
	r.Post("/api/user/login", h.Auth.APIUserLogin)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthJWT(authService, cfg, logger))

		r.Get("/api/user/orders", h.Order.APIUserGetOrders)
		r.Post("/api/user/orders", h.Order.APIUserCreateOrder)

		r.Get("/api/user/balance", h.Balance.APIUserGetBalance)
		r.Get("/api/user/withdrawals", h.Balance.APIUserGetWithdrawals)
		r.Post("/api/user/balance/withdraw", h.Balance.APIUserCreateWithdrawal)
	})
	return r
}
