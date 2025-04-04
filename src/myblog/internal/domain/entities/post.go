// myblog/internal/domain/entities/post.go (Бизнес-сущности - модели)
package entities

import (
	"errors"
	"fmt"
	"time"
)

type Post struct {
	ID            int       `json:"id"`
	Author        string    `json:"author"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	HTMLContent   string    `json:"html_content"`
	ContentDescr  string    `json:"content_descr"`
	HTMLContDescr string    `json:"html_cont_descr"`
	CreationTime  time.Time `json:"creation_time"`
}

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (p *Post) Validate() error {
	const (
		maxTitleLength = 1000
		maxDescLength  = 10000
	)

	switch {
	case len(p.Content) == 0:
		return errors.New("поле 'Текст' обязательно для заполнения")
	case p.Content == "":
		return errors.New("content required")
	case p.Title == "":
		return errors.New("title required")
	case len(p.Title) > maxTitleLength:
		return fmt.Errorf("заголовок слишком длинный (максимум %d символов)", maxTitleLength)
	case len(p.ContentDescr) > maxDescLength:
		return fmt.Errorf("описание слишком длинное (максимум %d символов)", maxDescLength)
	}

	return nil
}
