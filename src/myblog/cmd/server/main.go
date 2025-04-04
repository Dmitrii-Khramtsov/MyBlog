// src/myblog/cmd/server/main.go

// ├── cmd/
// │   └── blog/
// │       └── main.go
// │
// ├── configs/
// │   └── credentials/
// │       └── admin_credentials.txt
// │
// ├── internal/
// │   ├── domain/
// │   │   ├── entities/
// │   │   │   └── post.go
// │   │   └── repositories/
// │   │       └── post.go
// │   ├── application/
// │   │   └── usecase/
// │   │       └── post.go
// │   ├── interfaces/
// │   │   └── http/
// │   │       ├── handler/
// │   │       │   ├── api.go
// │   │       │   ├── auth.go
// │   │       │   ├── html.go
// │   │       │   ├── model_inline_response_400.go
// │   │       │   └── model_template_data.go
// │   |       ├── middleware/
// │   │       │   ├── admin.go
// │   │       │   ├── cors.go
// │   │       │   ├── csrf.go
// │   │       │   ├── jwt.go
// │   │       │   └── logger.go
// │   │       └── router/
// │   │           └── router.go
// │   └── infrastructure/
// │       ├── config/
// │       │   ├── admin_credentials.go
// │       │   ├── db_credentials.go
// │       │   ├── jwt_credentials.go
// │       │   ├── load_admin_credentials.go
// │       │   ├── load_config.go
// │       │   ├── load_db_credentials.go
// │       │   ├── load_jwt_credentials.go
// │       │   └── load_SQL.go
// │       ├── logger/
// │       │   └── logger.go
// │       └── persistence/
// │           └── postgres/
// │               └── article_repository.go
// |
// ├── static/
// |   ├── css/
// |   │   └── main.css
// |   ├── images/
// |   │   ├── favicon_io/
// |   │   │   ├── android-chrome-192x192.png
// |   │   │   ├── android-chrome-512x512.png
// |   │   │   ├── apple-touch-icon.png
// |   │   │   └── favicon.ico
// |   │   ├── blog_icon.png
// |   │   └── user_icon.png
// │   │
// │   └── logo_template
// |
// ├── templates/
// |   ├── admin_create.html
// |   ├── authorization.html
// |   ├── base.html
// |   ├── index.html
// |   └── post.html
// |
// ├── go.mod
// ├── go.sum
// └── Makefile

package main

import (
	"log"
	"net/http"

	"github.com/lonmouth/myblog/internal/application/usecases"
	"github.com/lonmouth/myblog/internal/infrastructure/config"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
	"github.com/lonmouth/myblog/internal/infrastructure/logo"
	"github.com/lonmouth/myblog/internal/infrastructure/markdown"
	"github.com/lonmouth/myblog/internal/infrastructure/persistence/postgres"
	"github.com/lonmouth/myblog/internal/interfaces/http/handlers"
	"github.com/lonmouth/myblog/internal/interfaces/http/router"
	"go.uber.org/zap"
)

const (
	logoFilePath   = "internal/infrastructure/logo/amazing_logo.png"
	configFilePath = "configs/credentials/admin_credentials.txt"
	serverPort     = ":8888"
	logLevel       = "debug"       // уровень логирования
	env            = "development" // окружение (режим работы приложения)
)

func main() {
	appLogger, err := logger.New(env, logLevel)
	if err != nil {
		log.Fatalf("Failed to inicialize logger: %v", err)
	}
	defer appLogger.Sync()

	log := appLogger.WithModule("main")

	err = logo.CreateGeneratedLogo(logoFilePath)
	if err != nil {
		log.Fatal("Failed to generate logo:", zap.Error(err))
	}

	// загрузка учетных данных базы данных
	dbCredentials, err := config.LoadDBCredentials(configFilePath)
	if err != nil {
		log.Fatal("Failed to load DB credentials", zap.Error(err))
	}

	// создание нового репозитория для работы с постами
	postRepo, err := postgres.NewPostRepository(dbCredentials, appLogger)
	if err != nil {
		log.Fatal("Failed to create post repository", zap.Error(err))
	}

	// загрузка SQL-запроса для создания таблицы
	createTableQuery, err := config.LoadSQLTables(configFilePath)
	if err != nil {
		log.Fatal("Failed to load SQL tanble", zap.Error(err))
	}

	// создание таблицы в базе данных
	err = postRepo.CreateTable(createTableQuery)
	if err != nil {
		log.Fatal("Failed to create table", zap.Error(err))
	}

	mdConverter := markdown.NewConverter()

	postUseCase := usecases.NewPost(postRepo, appLogger, mdConverter)

	postHandler := handlers.NewPostHandler(postUseCase, appLogger)

	jwtCfg, err := config.LoadJWTConfig(configFilePath)
	if err != nil {
		log.Fatal("Failed to create JWT config", zap.Error(err))
	}
	adminCreds, err := config.LoadAdminCredentials(configFilePath)
	if err != nil {
		log.Fatal("Failed to create admin credentials", zap.Error(err))
	}

	authHandler := handlers.NewAuthHandler(jwtCfg, adminCreds, appLogger)

	r := router.SetupRouters(postHandler, authHandler, jwtCfg, appLogger)

	log.Info("Starting server", zap.String("port", serverPort))
	log.Fatal(http.ListenAndServe(serverPort, r).Error())
}

// ALTER TABLE posts ADD COLUMN content_description TEXT;