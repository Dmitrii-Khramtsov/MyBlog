// myblog/internal/interfaces/http/middleware/middleware.go
package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// middleware для логирования HTTP-запросов
func LoggingMiddleware(logger *logger.AppLogger) mux.MiddlewareFunc {
	// возвращаем функцию, которая принимает следующий обработчик в цепочке
	return func(next http.Handler) http.Handler {
		// возвращаем новый обработчик, который будет обрабатывать запросы
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// создаём новый логгер с добавленными полями: метод, путь, удалённый адрес
			log := logger.With(
				zap.String("method", r.Method),          // метод HTTP-запроса (GET, POST..)
				zap.String("path", r.URL.Path),          // путь запроса
				zap.String("remote_addr", r.RemoteAddr), // адрес клиента, отправившего запрос
			)

			// логируем начало обработки запроса
			log.Info("Request started")

			// передаём управление следующему обработчику в цепочке
			next.ServeHTTP(w, r)

			// логируем завершение обработки запроса
			log.Info("Request completed")
		})
	}
}
