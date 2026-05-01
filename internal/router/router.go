package router

import (
	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/handler"
	"github.com/KV2013/gophermart-loyalty/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Init(handler *handler.URLHandler, logger *zap.Logger, cfg *config.Config) *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.ZapLogger(logger))
	r.Use(middleware.GzipCompression)
	r.Use(middleware.AuthJWT(cfg, logger))

	r.Post("/api/user/register", handler.APIUserRegister)
	r.Post("/api/user/login", handler.APIUserLogin)
	r.Post("/api/user/orders", handler.APIUserCreateOrder)
	r.Get("/api/user/orders", handler.APIUserGetOrders)
	r.Get("/api/user/balance", handler.APIUserGetBalance)
	r.Post("/api/user/balance/withdraw", handler.APIUserCreateWithdrawal)
	r.Get("/api/user/withdrawals", handler.APIUserGetWithdrawals)

	return r
}
