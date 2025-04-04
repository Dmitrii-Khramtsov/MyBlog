// myblog/internal/interfaces/http/middleware/admin.go
package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value("user").(*jwt.Token)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
