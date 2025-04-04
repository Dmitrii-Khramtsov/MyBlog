// src/myblog/internal/infrastructure/config/load_config.go
package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadConfig(filePath string, prefixes map[string]string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed tu open file: %w", err)
	}
	defer file.Close()

	configMap := make(map[string]string)
	scanner := bufio.NewScanner(file)

	// scanner.Scan() - читает данные из источника (например, файла) построчно, возвращает true, если успешно прочитал очередную строку, и false, если достиг конца входных данных или произошла ошибка
	for scanner.Scan() {
		// scanner.Text() - возвращает последнюю прочитанную строку, которая была сканирована с помощью scanner.Scan()
		line := scanner.Text()
		for key, prefix := range prefixes {
			// HasPrefix - проверяет, начинается ли строка line с префикса "AdminUsername="
			if strings.HasPrefix(line, prefix) {
				// TrimPrefix - удаляет префикс prefix из строки line
				configMap[key] = strings.TrimPrefix(line, prefix)
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return configMap, nil
}
