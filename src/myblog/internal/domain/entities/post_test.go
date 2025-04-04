// myblog/internal/domain/entities/post_test.go    (go test ./internal/domain/entities/... -v)
package entities_test

import (
	"testing"
	"time"

	"github.com/lonmouth/myblog/internal/domain/entities"
	"github.com/stretchr/testify/assert"
)

// TestPost_Validate проверяет метод Validate структуры Post.
func TestPost_Validate(t *testing.T) {
	// Подтест для проверки корректного поста.
	t.Run("Valid Post", func(t *testing.T) {
		post := entities.Post{
			Author:       "John Doe",
			Title:        "Sample Title",
			Content:      "This is the content of the post.",
			CreationTime: time.Now(),
		}
		// Проверяем, что для корректного поста ошибок нет.
		assert.NoError(t, post.Validate())
	})

	// Подтест для проверки случая, когда автор поста отсутствует.
	t.Run("Missing Author", func(t *testing.T) {
		post := entities.Post{
			Title:        "Sample Title",
			Content:      "This is the content of the post.",
			CreationTime: time.Now(),
		}
		// Проверяем, что возвращается ошибка "author required".
		assert.EqualError(t, post.Validate(), "author required")
	})

	// Подтест для проверки случая, когда заголовок поста отсутствует.
	t.Run("Missing Title", func(t *testing.T) {
		post := entities.Post{
			Author:       "John Doe",
			Content:      "This is the content of the post.",
			CreationTime: time.Now(),
		}
		// Проверяем, что возвращается ошибка "title required".
		assert.EqualError(t, post.Validate(), "title required")
	})

	// Подтест для проверки случая, когда содержание поста отсутствует.
	t.Run("Missing Content", func(t *testing.T) {
		post := entities.Post{
			Author:       "John Doe",
			Title:        "Sample Title",
			CreationTime: time.Now(),
		}
		// Проверяем, что возвращается ошибка "content required".
		assert.EqualError(t, post.Validate(), "content required")
	})
}
