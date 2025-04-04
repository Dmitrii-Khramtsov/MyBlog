// github.com/lonmouth/myblog/internal/infrastructure/logo
package logo

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
)

const (
	width          = 300
	height         = 300
	FileName       = "amazing_logo.png"
	PathToFile     = "amazing_logo.png"
	PathToTemplate = "static/logo_template/logo_300_2.txt"
	// PathToTemplate = "static/logo_template/betman_300.txt"
)

func OpenTemplate() (*bufio.Reader, *os.File) {
	file, err := os.Open(PathToTemplate)
	if err != nil {
		log.Fatalln("open template", err)
	}
	rd := bufio.NewReader(file)
	return rd, file
}

func GenerateLogo() *image.RGBA {
	// создаём изображение размером 300x300 пикселей
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// black := color.RGBA{0, 0, 0, 255}
	// yellow := color.RGBA{255, 221, 0, 255}
	black := color.RGBA{0, 0, 0, 255}
	gray := color.RGBA{217, 217, 217, 100}

	// открываем шаблон
	rd, file := OpenTemplate()
	defer file.Close() // закрываем файл после завершения чтения

	// рисуем логотип
	for i := 72; i < height; i++ {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Reached end of file")
				break
			}
			log.Fatalln("read template line", err)
		}

		for j := 0; j < width; j++ {
			switch {
			// case j < len(line) && line[j] == '0':
			case j < len(line) && (line[j] != '@' && line[j] != ' '):
				img.Set(j, i, gray)
			// case j < len(line) && (line[j] >= '0' && line[j] <= '9' && line[j] != '7'):
			case j < len(line) && (line[j] == '@'):
				img.Set(j, i, black)
			}
		}
	}

	return img
}

func CreateGeneratedLogo(PathToFile string) error {
	// создаём файл для сохранения изображения
	file, err := os.Create(PathToFile)
	if err != nil {
		return err
	}
	defer file.Close()

	img := GenerateLogo()
	// сохраняем изображение в файл
	if err := png.Encode(file, img); err != nil {
		return err
	}

	log.Println("логотип успешно создан и сохранен как", FileName)
	return nil
}

// https://www.asciiart.eu/image-to-ascii - генерация ASCII-арт изображения
