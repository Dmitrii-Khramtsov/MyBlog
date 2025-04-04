// myblog/internal/interfaces/http/handler/model_template_data.go
package handlers

import (
	"html/template"

	"github.com/lonmouth/myblog/internal/application/usecases"
	"github.com/lonmouth/myblog/internal/domain/entities"
)

// структурa данных для шаблона
type TemplateData struct {
	Title         string // Добавляем поле для заголовка
	Post          *entities.Post
	Posts         []*entities.Post
	Pagination    usecases.Pagination
	Authenticated bool
	Username      string
	Role          string
	CSRFToken     string
	CurrentPage   int
	ContentHTML   template.HTML
	Error         string
}
