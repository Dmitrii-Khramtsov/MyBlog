// src/myblog/internal/infrastructure/config/load_SQL.go
package config

import (
	"os"
	"strings"
)

const beginSQL = "# SQL Commands"

func LoadSQLTables(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// разделяем строку на две части: до и после "# SQL Commands"
	parts := strings.SplitAfterN(string(content), beginSQL, 2)

	querySQL := parts[1]

	querySQL = strings.TrimSpace(querySQL)
	querySQL = strings.TrimSuffix(querySQL, ";")

	return querySQL, nil
}
