package router

import (
	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Init(
	authHandler *handler.AuthHandler,
	orderHandler *handler.OrderHandler,
	balanceHandler *handler.BalanceHandler,
	logger *zap.Logger,
	cfg *config.Config,
) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.Api)
	r.Use(middleware.GzipCompression)

	r.Post("/api/user/register", authHandler.APIUserRegister)
	r.Post("/api/user/login", authHandler.APIUserLogin)

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthJWT(cfg, logger))

		r.Get("/api/user/orders", orderHandler.APIUserGetOrders)
		r.Post("/api/user/orders", orderHandler.APIUserCreateOrder)

		r.Get("/api/user/balance", balanceHandler.APIUserGetBalance)
		r.Get("/api/user/withdrawals", balanceHandler.APIUserGetWithdrawals)
		r.Post("/api/user/balance/withdraw", balanceHandler.APIUserCreateWithdrawal)
	})
	return r
}
