// myblog/internal/infrastructure/config/config.go
package config

import "errors"

type JWTConfig struct {
	SecretKey string `yaml:"secret_key"`
	ExpiresIn int    `yaml:"expires_in"` // В минутах
}

func (c *JWTConfig) Validate() error {
	if c.SecretKey == "" {
		return errors.New("JWT secret key is required")
	}
	if c.ExpiresIn <= 0 {
		return errors.New("JWT expiration time must be positive")
	}
	return nil
}
