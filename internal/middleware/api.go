package middleware

import (
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func API(l *zap.Logger) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				// if r.URL.Path == "/api/user/orders" && r.Method == http.MethodPost {
				// 	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
				// 		l.Debug("не валидный заголовок Content-Type", zap.String("request_header", r.Header.Get("Content-Type")))
				// 		w.Header().Set("Content-Type", "application/json")
				// 		w.WriteHeader(http.StatusBadRequest)
				// 		return
				// 	}
				// }
				w.Header().Set("Content-Type", "application/json")
			}
			next.ServeHTTP(w, r)
		})
	}

}
