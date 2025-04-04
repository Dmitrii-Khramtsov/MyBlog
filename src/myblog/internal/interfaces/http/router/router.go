// myblog/internal/interfaces/http/router/router.go
package router

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/lonmouth/myblog/internal/infrastructure/config"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
	"github.com/lonmouth/myblog/internal/interfaces/http/handlers"
	"github.com/lonmouth/myblog/internal/interfaces/http/middleware"
	"golang.org/x/time/rate"
)

func SetupRouters(
	postHandler *handlers.PostHandler,
	authHandler *handlers.AuthHandler,
	jwtCfg config.JWTConfig,
	logger *logger.AppLogger,
) *mux.Router {
	r := mux.NewRouter()

	r.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusNoContent)
	})

	// Создаем RateLimiter с лимитом 100 запросов в секунду
	rateLimiter := middleware.NewRateLimiter(rate.Every(time.Second/100), 100)

	// Общий middleware
	r.Use(
		middleware.LoggingMiddleware(logger),
		middleware.CSRFProtectionMiddleware,
		middleware.JWTAuthMiddleware(jwtCfg.SecretKey, logger),
		rateLimiter.RateLimit,
	)

	// Публичные маршруты
	r.HandleFunc("/login", authHandler.ShowLoginForm).Methods("GET")
	r.HandleFunc("/login", authHandler.Login).Methods("POST")
	r.HandleFunc("/logout", authHandler.Logout).Methods("POST")

	// Основные маршруты
	r.HandleFunc("/", postHandler.HandleHTMLGetPosts).Methods("GET")
	r.HandleFunc("/posts", postHandler.HandleHTMLGetPosts).Methods("GET")
	r.HandleFunc("/posts/{id:[0-9]+}", postHandler.HandleHTMLGetPost).Methods("GET")

	// API маршруты
	api := r.PathPrefix("/api/v1").Subrouter()
	api.Use(middleware.AdminOnlyMiddleware)
	api.HandleFunc("/posts", postHandler.HandleAPICreatePost).Methods("POST")
	api.HandleFunc("/posts/{id}", postHandler.HandleAPIGetPostByID).Methods("PUT")

	// Админские HTML маршруты
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AdminOnlyMiddleware)
	admin.HandleFunc("/posts/create", postHandler.HandleAdminCreatePostForm).Methods("GET")
	admin.HandleFunc("/posts/create", postHandler.HandleAdminCreatePost).Methods("POST")

	// Статические файлы
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	return r
}
