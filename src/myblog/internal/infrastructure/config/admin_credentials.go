// myblog/internal/infrastructure/config/admin_credentials.go
package config

import (
	"errors"
)

// AdminCredentials - структура для хранения учетных данных администратора
type AdminCredentials struct {
	Username string
	Password string
}

func (a *AdminCredentials) Validate() error {
	switch {
	case a.Username == "":
		return errors.New("username required")
	case a.Password == "":
		return errors.New("password required")
	}
	return nil
}
