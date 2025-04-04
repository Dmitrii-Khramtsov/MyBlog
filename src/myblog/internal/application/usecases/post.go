// myblog/internal/application/usecases/post.go
package usecases

import (
	"time"

	"github.com/lonmouth/myblog/internal/domain/entities"
	"github.com/lonmouth/myblog/internal/domain/repositories"
	"github.com/lonmouth/myblog/internal/infrastructure/logger"
	"github.com/lonmouth/myblog/internal/infrastructure/markdown"
	"go.uber.org/zap"
)

const (
	pageSize     = 3 // количество постов на странице
	visiblePages = 4 // всего видимых страниц (включая текущую)
	pagesAround  = 2 // страниц по бокам от текущей
)

// Определяем интерфейс для use-case постов
type PostUseCase interface {
	CreatePost(post *entities.Post) (int, int, error)
	GetPosts(page int) ([]*entities.Post, Pagination, error)
	GetPostByID(id int) (*entities.Post, error)
	Count() (int, error) // колличество постов
	generateHTMLContent(post *entities.Post) error
}

// структура, реализующая интерфейс PostUseCase
type Post struct {
	repo        repositories.Post
	logger      *logger.AppLogger
	mdConverter *markdown.Converter
}

// Pagination содержит информацию для отображения пагинации
type Pagination struct {
	CurrentPage int   // текущая страница
	TotalPages  int   // всего страниц
	Pages       []int // номера отображаемых страниц
	ShowFirst   bool  // показывать кнопку "В начало"
	ShowNext    bool  // показывать кнопку "Дальше"
	FirstPage   int   // номер первой страницы
	LastPage    int   // номер последней страницы
}

// конструктор для создания нового экземпляра Post
func NewPost(repo repositories.Post, logger *logger.AppLogger, converter *markdown.Converter) PostUseCase {
	return &Post{
		repo:        repo,
		logger:      logger.WithModule("post_usecase"), // создаём логер для post_handler модуля
		mdConverter: markdown.NewConverter(),
	}
}

// метод для создания нового поста
func (p *Post) CreatePost(post *entities.Post) (int, int, error) {
	log := p.logger.With(zap.Any("post", post)) // логирует данные и добавляет поле с объектом поста (.Any() - позволяет добавлять поля любого типа в лог)

	post.CreationTime = time.Now() // принудительная установка времени

	// генерация HTML контента
	if err := p.generateHTMLContent(post); err != nil {
		log.Error("Markdown conversion failed", zap.Error(err))
		return 0, 0, err
	}

	if err := post.Validate(); err != nil {
		log.Error("Post validation failed", zap.Error(err)) // логирование ошибки валидации
		return 0, 0, err
	}

	if post.CreationTime.IsZero() {
		post.CreationTime = time.Now() // установка времени создания, если оно не задано
	}

	if err := p.repo.Create(post); err != nil {
		log.Error("Failed to create post", zap.Error(err)) // логирование ошибки создания поста
		return 0, 0, err
	}

	// Получаем общее количество постов
	totalPosts, err := p.repo.Count()
	if err != nil {
		log.Error("Failed to count posts", zap.Error(err))
		return 0, 0, err
	}

	// Вычисляем номер страницы для нового поста
	pageNumber := (totalPosts) / pageSize // количество страниц
	if totalPosts%pageSize > 0 {
		pageNumber++
	}

	log.Info("Post created successfully",
		zap.Int("post_id", post.ID),
		zap.String("content", post.Content),
		zap.String("description", post.ContentDescr),
	)

	log.Info("Post created successfully") // логирование успешного создания поста
	return post.ID, pageNumber, nil
}

// метод для получения списка постов с пагинацией
func (p *Post) GetPosts(page int) ([]*entities.Post, Pagination, error) {
	posts, err := p.repo.List(page, pageSize)
	if err != nil {
		p.logger.Error("Failed to retrieve posts", zap.Error(err)) // логирование ошибки получения постов
		return nil, Pagination{}, err
	}

	total, err := p.repo.Count() // общее количество постов  из базы
	if err != nil {
		p.logger.Error("Failed to get posts count", zap.Error(err))
		return nil, Pagination{}, err
	}

	totalPages := total / pageSize // количество страниц
	if total%pageSize > 0 {
		totalPages++
	}

	// корректируем номер страницы если он выходит за допустимые пределы
	if page < 1 {
		page = 1
	} else if page > totalPages {
		page = totalPages
	}

	// Рассчитываем параметры пагинации
	pagination := p.calculatePagination(page, totalPages)

	return posts, pagination, nil
}

// calculatePagination вычисляет параметры пагинации
func (p *Post) calculatePagination(current, totalPages int) Pagination {
	start := current - pagesAround
	end := current + pagesAround

	if start < 1 {
		start = 1
		end = start + visiblePages - 1
		if end > totalPages {
			end = totalPages
		}
	}

	if end > totalPages {
		end = totalPages
		start = end - visiblePages + 1
		if start < 1 {
			start = 1
		}
	}

	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		if i == end {
			i = totalPages
		} // посленяя страница в списке пагинации (end) ровна totalPages
		pages = append(pages, i)
	}

	return Pagination{
		CurrentPage: current,
		TotalPages:  totalPages,
		Pages:       pages,
		ShowFirst:   current > (pagesAround + 1),
		ShowNext:    current < totalPages,
		FirstPage:   1,
		LastPage:    totalPages,
	}
}

// метод для получения поста по ID
func (p *Post) GetPostByID(id int) (*entities.Post, error) {
	post, err := p.repo.GetByID(id)
	if err != nil {
		p.logger.Error("Failed to retrieve post by ID", zap.Error(err), zap.Int("id", id)) // логирование ошибки получения поста
		return nil, err
	}

	// Регенерация HTML если отсутствует
	if post.HTMLContent == "" {
		if err := p.generateHTMLContent(post); err != nil {
			p.logger.Warn("Failed to regenerate HTML content",
				zap.Int("post_id", post.ID),
				zap.Error(err))
		}
	}

	return post, nil
}

// метод для получения общего количества постов
func (p *Post) Count() (int, error) {
	count, err := p.repo.Count()
	if err != nil {
		p.logger.Error("Failed to count posts", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// // generateHTMLContent преобразует Markdown в HTML
// func (p *Post) generateHTMLContent(post *entities.Post) error {
// 	// Конвертация основного контента
// 	htmlContent, err := p.mdConverter.ToHTML(post.Content)
// 	if err != nil {
// 		return err
// 	}
// 	post.HTMLContent = htmlContent

// 	// Конвертация краткого описания (если необходимо)
// 	if post.ContentDescr != "" {
// 		htmlDescr, err := p.mdConverter.ToHTML(post.ContentDescr)
// 		if err != nil {
// 			return err
// 		}
// 		post.HTMLContDescr = htmlDescr
// 	}
// 	return nil
// }

func (p *Post) generateHTMLContent(post *entities.Post) error {
	p.logger.Debug("Converting Markdown",
		zap.String("input_content", post.Content),
		zap.String("input_desc", post.ContentDescr),
	)

	htmlContent, err := p.mdConverter.ToHTML(post.Content)
	if err != nil {
		p.logger.Error("HTML conversion failed for content",
			zap.Error(err),
			zap.String("content", post.Content),
		)
		return err
	}
	post.HTMLContent = htmlContent

	if post.ContentDescr != "" {
		htmlDescr, err := p.mdConverter.ToHTML(post.ContentDescr)
		if err != nil {
			p.logger.Error("HTML conversion failed for description",
				zap.Error(err),
				zap.String("description", post.ContentDescr),
			)
			return err
		}
		post.HTMLContDescr = htmlDescr
	}

	p.logger.Debug("HTML generated",
		zap.String("html_content", post.HTMLContent),
		zap.String("html_desc", post.HTMLContDescr),
	)

	return nil
}
