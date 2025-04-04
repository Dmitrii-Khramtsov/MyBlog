package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

type contextKey struct{}

var csrfTokenKey = &contextKey{}

func CSRFProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			token := generateCSRFToken()
			http.SetCookie(w, &http.Cookie{
				Name:     "csrf_token",
				Value:    token,
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			ctx = context.WithValue(ctx, csrfTokenKey, token)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		var formToken string
		contentType := r.Header.Get("Content-Type")

		switch {
		case strings.Contains(contentType, "application/json"):
			tokenHeader := r.Header.Get("X-CSRF-Token")
			if tokenHeader == "" {
				http.Error(w, "CSRF token required in header", http.StatusForbidden)
				return
			}
			formToken = tokenHeader

		case strings.Contains(contentType, "x-www-form-urlencoded"):
			formToken = r.FormValue("csrf_token")

		default:
			http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
			return
		}

		cookie, err := r.Cookie("csrf_token")
		if err != nil || cookie.Value != formToken {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		newToken := generateCSRFToken()
		http.SetCookie(w, &http.Cookie{
			Name:     "csrf_token",
			Value:    newToken,
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
		})

		ctx = context.WithValue(ctx, csrfTokenKey, newToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Вспомогательная функция для получения токена из контекста
func CSRFTokenFromContext(ctx context.Context) string {
	if token, ok := ctx.Value(csrfTokenKey).(string); ok {
		return token
	}
	return ""
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("CSRF token generation failed: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)
}
