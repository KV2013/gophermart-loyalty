package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/KV2013/gophermart-loyalty/internal/config"
	"github.com/KV2013/gophermart-loyalty/internal/service/auth"
	"go.uber.org/zap"
)

const tokenCookieName = "token"

type AuthError struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func AuthJWT(cfg *config.Config, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string
			var respErr error

			cookie, err := r.Cookie(tokenCookieName)
			if err != nil {
				respErr = writeError(w, AuthError{
					Status:  http.StatusUnauthorized,
					Message: "Требуется аутентификацйия",
				})
				if respErr != nil {
					logger.Error("writeError", zap.Error(respErr))
				}
				return
			}

			id, parseErr := auth.GetUserID(cookie.Value, cfg.JWTSecretKey)
			if parseErr != nil {
				respErr = writeError(w, AuthError{
					Status:  http.StatusUnauthorized,
					Message: parseErr.Error(),
				})
				if respErr != nil {
					logger.Error("writeError", zap.Error(respErr))
				}
				return
			}

			if id == "" {
				respErr = writeError(w, AuthError{
					Status:  http.StatusUnauthorized,
					Message: "Токен не содержит UserID",
				})
				if respErr != nil {
					logger.Error("writeError", zap.Error(respErr))
				}
				return
			}

			userID = id
			logger.Debug("middleware.AuthJWT", zap.String("userID", userID))
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeError(w http.ResponseWriter, err AuthError) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		return err
	}

	return nil
}
