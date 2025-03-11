//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fairytale5571/cv_onopchenko/components"
	"gopkg.in/yaml.v3"
)

// loadResumeData загружает данные резюме из файла конфигурации
func loadResumeData(configPath string) (*components.ResumeData, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var resumeData components.ResumeData
	err = yaml.Unmarshal(data, &resumeData)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	return &resumeData, nil
}

// copyDir копирует содержимое одной директории в другую
func copyDir(src, dst string) error {
	// Создаем целевую директорию
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// Читаем содержимое исходной директории
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Рекурсивно копируем поддиректории
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Копируем файл
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}

			err = os.WriteFile(dstPath, data, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	// Создаем директорию для статических файлов
	err := os.MkdirAll("dist", 0755)
	if err != nil {
		log.Fatalf("Failed to create dist directory: %v", err)
	}

	// Загружаем данные резюме
	resumeData, err := loadResumeData("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load resume data: %v", err)
	}

	// Создаем компонент резюме
	resumeComponent := components.Resume(*resumeData)

	// Рендерим HTML в буфер
	var buf bytes.Buffer
	err = resumeComponent.Render(context.Background(), &buf)
	if err != nil {
		log.Fatalf("Failed to render resume: %v", err)
	}

	// Записываем HTML в файл
	err = os.WriteFile("dist/index.html", buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Failed to write index.html: %v", err)
	}

	// Копируем статические файлы
	err = copyDir("static", "dist/static")
	if err != nil {
		log.Fatalf("Failed to copy static files: %v", err)
	}

	// Создаем файл .nojekyll для GitHub Pages
	err = os.WriteFile("dist/.nojekyll", []byte{}, 0644)
	if err != nil {
		log.Fatalf("Failed to create .nojekyll file: %v", err)
	}

	// Создаем файл CNAME для пользовательского домена (если нужно)
	// err = ioutil.WriteFile("dist/CNAME", []byte("your-domain.com"), 0644)
	// if err != nil {
	// 	log.Fatalf("Failed to create CNAME file: %v", err)
	// }

	fmt.Println("Static site generated successfully in the 'dist' directory")
}
