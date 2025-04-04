// src/myblog/internal/infrastructure/config/load_admin_credentials.go
package config

import "fmt"

const (
	username = "admin_username="
	password = "admin_password="
)

func LoadAdminCredentials(filePath string) (AdminCredentials, error) {
	prefixes := map[string]string{
		"Username": username,
		"Password": password,
	}

	configMap, err := LoadConfig(filePath, prefixes)
	if err != nil {
		return AdminCredentials{}, err
	}

	credentials := AdminCredentials{
		Username: configMap["Username"],
		Password: configMap["Password"],
	}

	if err := credentials.Validate(); err != nil {
		return AdminCredentials{}, fmt.Errorf("missing credentials in file: %w", err)
	}

	return credentials, nil
}
