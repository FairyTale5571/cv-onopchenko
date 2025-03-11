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
	"strings"

	"github.com/fairytale5571/cv_onopchenko/components"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/yaml.v3"
)

// PDF constants
const (
	PDFOrientation = "P"
	PDFUnit        = "mm"
	PDFSize        = "A4"
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

	// Генерируем PDF файл
	err = generatePDF("dist/static/resume.pdf", resumeData)
	if err != nil {
		log.Fatalf("Failed to generate PDF: %v", err)
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

// generatePDF generates the PDF file
func generatePDF(outputPath string, resumeData *components.ResumeData) error {
	pdf := gofpdf.New(PDFOrientation, PDFUnit, PDFSize, "")
	pdf.SetFont("Helvetica", "", 12)
	pdf.AddPage()

	// Add content to PDF
	pdf.SetFont("Helvetica", "B", 24)
	pdf.Cell(0, 8, resumeData.Personal.Name)
	pdf.Ln(16)

	pdf.SetFont("Helvetica", "", 18)
	pdf.Cell(0, 8, resumeData.Personal.Title)
	pdf.Ln(16)

	// Add contact information
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 8, "Contact Information")
	pdf.Ln(12)

	pdf.SetFont("Helvetica", "", 12)
	pdf.Cell(0, 8, "Email: "+resumeData.Personal.Email)
	pdf.Ln(8)
	pdf.Cell(0, 8, "Phone: "+resumeData.Personal.Phone)
	pdf.Ln(8)
	pdf.Cell(0, 8, "Location: "+resumeData.Personal.Location)
	pdf.Ln(8)
	pdf.Cell(0, 8, "LinkedIn: "+resumeData.Personal.Linkedin)
	pdf.Ln(8)
	pdf.Cell(0, 8, "GitHub: "+resumeData.Personal.Github)
	pdf.Ln(16)

	// Add experience section
	pdf.SetFont("Helvetica", "B", 18)
	pdf.Cell(0, 8, "Experience")
	pdf.Ln(12)

	for _, exp := range resumeData.Experience {
		pdf.SetFont("Helvetica", "B", 14)
		pdf.Cell(0, 8, exp.Position)
		pdf.Ln(8)

		pdf.SetFont("Helvetica", "", 12)
		companyInfo := exp.Company + ", " + exp.Location
		if exp.EmploymentType != "" {
			companyInfo += " (" + exp.EmploymentType + ")"
		}
		pdf.Cell(0, 8, companyInfo)
		pdf.Ln(8)
		pdf.Cell(0, 8, exp.StartDate+" - "+exp.EndDate)
		pdf.Ln(8)

		for _, desc := range exp.Description {
			pdf.Cell(6, 8, "-")
			pdf.SetX(pdf.GetX() + 4)
			pdf.MultiCell(0, 8, desc, "", "", false)
		}

		if len(exp.Technologies) > 0 {
			pdf.Ln(8)
			pdf.SetFont("Helvetica", "B", 12)
			pdf.Cell(0, 8, "Technologies:")
			pdf.Ln(8)

			pdf.SetFont("Helvetica", "", 12)
			var techNames []string
			for _, tech := range exp.Technologies {
				techNames = append(techNames, tech.Name)
			}
			techText := strings.Join(techNames, ", ")
			pdf.MultiCell(0, 8, techText, "", "", false)
		}

		pdf.Ln(12)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return pdf.OutputFileAndClose(outputPath)
}
