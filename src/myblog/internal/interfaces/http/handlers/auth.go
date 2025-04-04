// myblog/internal/interfaces/http/handlers/auth.go
package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lonmouth/myblog/internal/infrastructure/config"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
	"github.com/lonmouth/myblog/internal/interfaces/http/middleware"
	"go.uber.org/zap"
)

type AuthHandler struct {
	jwtCfg    config.JWTConfig
	adminCred config.AdminCredentials
	logger    *logger.AppLogger
}

func NewAuthHandler(jwtCfg config.JWTConfig, adminCred config.AdminCredentials, logger *logger.AppLogger) *AuthHandler {
	return &AuthHandler{
		jwtCfg:    jwtCfg,
		adminCred: adminCred,
		logger:    logger,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		CSRFToken string `json:"csrf_token"`
	}

	contentType := r.Header.Get("Content-Type")

	switch {
	case strings.Contains(contentType, "application/json"):
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			h.logger.Warn("JSON decode error", zap.Error(err))
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

	case strings.Contains(contentType, "x-www-form-urlencoded"):
		if err := r.ParseForm(); err != nil {
			h.logger.Warn("Form parse error", zap.Error(err))
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		request.Username = r.FormValue("username")
		request.Password = r.FormValue("password")
		request.CSRFToken = r.FormValue("csrf_token")

	default:
		h.logger.Warn("Unsupported content type", zap.String("type", contentType))
		http.Error(w, "Unsupported media type", http.StatusUnsupportedMediaType)
		return
	}

	h.logger.Debug("Login attempt data",
		zap.String("username", request.Username),
		zap.Bool("password_provided", request.Password != ""),
	)

	if request.Username != h.adminCred.Username || request.Password != h.adminCred.Password {
		h.logger.Warn("Invalid credentials", zap.String("username", request.Username))
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": request.Username,
		"role":     "admin",
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	// Логирование данных пользователя для проверки роли
	claims := token.Claims.(jwt.MapClaims)
	h.logger.Info("Пользователь вошел в систему", zap.String("username", claims["username"].(string)), zap.String("role", claims["role"].(string)))

	tokenString, err := token.SignedString([]byte(h.jwtCfg.SecretKey))
	if err != nil {
		h.logger.Error("Token generation failed", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "jwt",
		Value:   tokenString,
		Expires: time.Now().Add(2 * time.Hour),
		// 	Expires:  time.Now().Add(15 * time.Minute),
		Path:     "/",
		Secure:   false, // Для HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Редирект вместо JSON-ответа
	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Добавить проверку наличия куки
	cookieToken, err := r.Cookie("csrf_token")
	if err != nil {
		http.Error(w, "CSRF token missing", http.StatusForbidden)
		return
	}

	formToken := r.FormValue("csrf_token")
	if cookieToken.Value != formToken {
		http.Error(w, "Invalid CSRF token", http.StatusForbidden)
		return
	}

	// Удаляем JWT cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/posts", http.StatusSeeOther)
}

func (h *AuthHandler) ShowLoginForm(w http.ResponseWriter, r *http.Request) {
	// Получаем токен из контекста
	csrfToken := middleware.CSRFTokenFromContext(r.Context())
	if csrfToken == "" {
		h.logger.Error("CSRF token missing in context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/authorization.html",
		"templates/partials/user_panel.html",
	))

	err := tmpl.Execute(w, map[string]interface{}{
		"Title":         "Авторизация",
		"Authenticated": false,
		"CSRFToken":     csrfToken,
	})

	if err != nil {
		h.logger.Error("Failed to render login form",
			zap.Error(err),
			zap.String("path", r.URL.Path),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
