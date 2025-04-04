// src/myblog/internal/infrastructure/config/load_jwt_credentials.go
package config

import (
	"fmt"
	"strconv"
)

const (
	secretKeyPrefix = "jwt_secret="
	expiresInPrefix = "jwt_expires="
)

func LoadJWTConfig(filePath string) (JWTConfig, error) {
	prefixes := map[string]string{
		"SecretKey": secretKeyPrefix,
		"ExpiresIn": expiresInPrefix,
	}

	configMap, err := LoadConfig(filePath, prefixes)
	if err != nil {
		return JWTConfig{}, err
	}

	// преобразуем строку в число для expiresIn
	expiresIn, err := strconv.Atoi(configMap["ExpiresIn"])
	if err != nil {
		return JWTConfig{}, fmt.Errorf("invalid expires_in format: %w", err)
	}

	configJWT := JWTConfig{
		SecretKey: configMap["SecretKey"],
		ExpiresIn: expiresIn,
	}

	if err := configJWT.Validate(); err != nil {
		return JWTConfig{}, fmt.Errorf("invalid JWT config: %w", err)
	}

	// logger.Info("Loading JWT config")
	return configJWT, nil
}
