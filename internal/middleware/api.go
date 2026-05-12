package middleware

import (
	"net/http"
	"strings"
)

func Api(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			if !(r.Method == http.MethodPost && r.URL.Path == "/api/user/orders") {
				if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
		}
		next.ServeHTTP(w, r)
	})
}
