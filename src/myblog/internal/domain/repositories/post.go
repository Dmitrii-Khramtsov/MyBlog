// myblog/internal/domain/repositories/post.go (интерфейсы репозиториев)
package repositories

import "github.com/lonmouth/myblog/internal/domain/entities"

type Post interface {
	GetByID(id int) (*entities.Post, error)
	List(page, limit int) ([]*entities.Post, error)
	Create(post *entities.Post) error
	Count() (int, error)
	// Update(post *entities.Post) error
	// Delete(id int) error
}
