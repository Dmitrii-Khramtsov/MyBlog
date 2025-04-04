// src/myblog/internal/infrastructure/persistence/postgres/post_repository.go
package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // с подчеркиванием перед именем пакета - называется "blank import" (пустой импорт) и используется в моём случае для регистрации драйвера
	//_ github.com/lib/pq регистрирует себя как драйвер для PostgreSQL в системе database/sql. Это позволяет использовать строку подключения postgres в функции sql.Open.
	"github.com/lonmouth/myblog/internal/domain/entities"
	"github.com/lonmouth/myblog/internal/infrastructure/config"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// реализация репозитория для работы с постами
type PostRepository struct {
	// объект представляющий пул соединений с базой данных, используется для выполнения SQL-запросов
	db     *sql.DB
	logger *logger.AppLogger
}

func NewPostRepository(credentials config.DBCredentials, logger *logger.AppLogger) (*PostRepository, error) {
	// формирование строки подключения
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		credentials.DBUser, credentials.DBPassword, credentials.DBName)

	// установка соединения с базой данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Error("Failed to connect to the database", zap.Error(err))
		return nil, err
	}

	// проверка соединения
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping the database", zap.Error(err))
		return nil, err
	}

	return &PostRepository{db: db, logger: logger.WithModule("post_repository")}, nil
}

func (r *PostRepository) CreateTable(query string) error {
	_, err := r.db.Exec(query)

	if err != nil {
		r.logger.Error("Failed to create table", zap.Error(err))
		return fmt.Errorf("failed to create table: %w", err)
	}
	r.logger.Info("Table created successfully")
	return nil
}

func (r *PostRepository) Create(post *entities.Post) error {
	query := `INSERT INTO posts 
  (author, title, content, html_content, content_descr, html_cont_descr, creation_time) 
  VALUES ($1, $2, $3, $4, $5, $6, $7) 
  RETURNING id, html_content, html_cont_descr`

	err := r.db.QueryRow(
		query,
		post.Author,
		post.Title,
		post.Content,
		post.HTMLContent,
		post.ContentDescr,
		post.HTMLContDescr,
		post.CreationTime,
	).Scan(
		&post.ID,
		&post.HTMLContent,   // cканируем возвращенные значения
		&post.HTMLContDescr, // из БД
	)

	if err != nil {
		r.logger.Error("Failed to create post", zap.Error(err))
		return err
	}

	r.logger.Info("Post created successfully", zap.Int("post_id", post.ID))
	return nil
}

func (r *PostRepository) List(page, limit int) ([]*entities.Post, error) {
	offset := (page - 1) * limit

	rows, err := r.db.Query(`
    SELECT id, author, title, content, content_descr, html_cont_descr, creation_time 
    FROM posts 
    LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		r.logger.Error("Failed to retrieve posts", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var posts []*entities.Post

	// итерация по строкам результата, возвращаемого SQL-запросом
	for rows.Next() {
		var post entities.Post

		// cканирование значений из текущей строки в структуру post
		if err := rows.Scan(
			&post.ID,
			&post.Author,
			&post.Title,
			&post.Content,
			&post.ContentDescr,
			&post.HTMLContDescr,
			&post.CreationTime,
		); err != nil {
			r.logger.Error("Failed to scan post", zap.Error(err))
			return nil, err
		}

		posts = append(posts, &post)
	}

	r.logger.Info("Posts retrieved successfully", zap.Int("page", page), zap.Int("limit", limit))
	return posts, nil
}

// http://localhost:8080/posts?page=1
func (r *PostRepository) GetByID(id int) (*entities.Post, error) {
	query := `SELECT id, author, title, content, html_content, 
                   content_descr, html_cont_descr, creation_time 
            FROM posts WHERE id = $1`

	// SQL-запрос для выборки записи из таблицы posts по её ID
	row := r.db.QueryRow(query, id)

	var post entities.Post
	// сканирование значений из строки в структуру post
	err := row.Scan(
		&post.ID,
		&post.Author,
		&post.Title,
		&post.Content,
		&post.HTMLContent,
		&post.ContentDescr,
		&post.HTMLContDescr,
		&post.CreationTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("Post not found", zap.Int("id", id))
		} else {
			r.logger.Error("Failed to retrieve post by ID", zap.Error(err), zap.Int("id", id))
		}

		return nil, err
	}

	r.logger.Info("Post retrieved successfully", zap.Int("id", id))
	return &post, nil
}

func (r *PostRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&count)
	if err != nil {
		r.logger.Error("Failed to count posts", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// DROP TABLE posts;
