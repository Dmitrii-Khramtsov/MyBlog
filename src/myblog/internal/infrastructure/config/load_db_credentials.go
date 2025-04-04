// src/myblog/internal/infrastructure/config/load_db_credentials.go
package config

import "fmt"

const (
	dBName     = "db_name="
	dBUser     = "db_user="
	dBPassword = "db_password="
)

func LoadDBCredentials(filePath string) (DBCredentials, error) {
	prefixes := map[string]string{
		"DBName":     dBName,
		"DBUser":     dBUser,
		"DBPassword": dBPassword,
	}

	configMap, err := LoadConfig(filePath, prefixes)
	if err != nil {
		return DBCredentials{}, err
	}

	credentials := DBCredentials{
		DBName:     configMap["DBName"],
		DBUser:     configMap["DBUser"],
		DBPassword: configMap["DBPassword"],
	}

	if err := credentials.Validate(); err != nil {
		return DBCredentials{}, fmt.Errorf("missing credentials in file: %w", err)
	}

	return credentials, nil
}
