# Определяем имя бинарного файла
BINARY_NAME = myblog

# Определяем директории, которые нужно включить в ZIP-архив
STATIC_DIRS = static templates configs internal/infrastructure/logo/amazing_logo.png

# Определяем имя выходного ZIP-файла
ZIP_FILE = $(BINARY_NAME).zip

# Сборка Go-бинарника
build:
	go build -o $(BINARY_NAME) cmd/server/main.go

# Создание ZIP-архива
zip: build
	zip $(ZIP_FILE) $(BINARY_NAME) -r $(STATIC_DIRS)

# Очистка артефактов сборки
clean:
	rm -f $(BINARY_NAME) $(ZIP_FILE)

.PHONY: build zip clean
