package logger

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// итерфейс Logger определяет методы для логирования на разных уровнях
type Logger interface {
	Debug(msg string, fields ...zap.Field) // логирование отладочной информации
	Info(msg string, fields ...zap.Field)  // логирование информационных сообщений
	Warn(msg string, fields ...zap.Field)  // логирование предупреждений
	Error(msg string, fields ...zap.Field) // логирование ошибок, приводящих к завершению программы
	Fatal(msg string, fields ...zap.Field) // логирование критических ошибок
	With(fields ...zap.Field) *zap.Logger  // создание нового логгера с добавленными полями
	Named(name string) *zap.Logger         // создание нового логгера с указанным именем
	Sync() error
	WithRequestID(requestID string) Logger
	WithModule(module string) Logger
	WithContext(ctx context.Context) context.Context
}

// структура, реализующая интерфэйс Logger
type AppLogger struct {
	*zap.Logger
}

type contextKey struct{}

var loggerKey = &contextKey{}

func New(env, level string) (*AppLogger, error) {
	// var config zap.Config
	config := zap.NewProductionConfig()

	// настройка конфигурации логгера в зависимости от окружения
	if env == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// настройки для всех окружений
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	// установка уровня логирования
	logLevel := parseLogLevel(level)
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// создание логгера с заданной конфигурацией
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &AppLogger{logger}, nil
}

// преобразует строку в уровень логирования
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// добавляет ID запроса (request_id) к логам
func (l *AppLogger) WithRequestID(requestID string) *AppLogger {
	return &AppLogger{l.Logger.With(zap.String("request_id", requestID))}
}

// создает новый логгер с указанным именем модуля
func (l *AppLogger) WithModule(module string) *AppLogger {
	return &AppLogger{l.Logger.Named(module)}
}

// context (контекст) в Go — это механизм для передачи данных между обработчиками запросов,
// аутентификационные данные, логгеры, тайм-аутов и отмены длительных операций.
// в веб-приложениях контекст часто используется для передачи данных между middleware и обработчиками
func (l *AppLogger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func FromContext(ctx context.Context) *AppLogger {
	if logger, ok := ctx.Value(loggerKey).(*AppLogger); ok {
		return logger
	}
	return &AppLogger{zap.NewNop()}
}

// многие системы логирования используют буферизацию для повышения производительности.
// Это означает, что логи не сразу записываются в файл или отправляются по сети,
// а накапливаются в памяти и записываются партиями.
// Вызов Sync гарантирует, что все логи, которые были записаны до этого момента, будут физически сохранены

func (l *AppLogger) Sync() error {
	return l.Logger.Sync()
}
