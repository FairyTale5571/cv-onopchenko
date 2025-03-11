package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fairytale5571/cv_onopchenko/components"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/yaml.v3"
)

const (
	// PDF settings
	PDFOrientation = "P"
	PDFUnit        = "mm"
	PDFSize        = "A4"
	PDFFont        = "Helvetica"

	// Font styles
	FontStyleNormal = ""
	FontStyleBold   = "B"

	// Font sizes
	FontSizeTitle      = 24
	FontSizeSubtitle   = 18
	FontSizeHeading    = 18
	FontSizeSubheading = 14
	FontSizeNormal     = 12

	// Spacing
	LineHeightLarge  = 16
	LineHeightMedium = 12
	LineHeightSmall  = 10
	LineHeightText   = 8
	ParagraphSpacing = 16
	SectionSpacing   = 12
	ListItemIndent   = 0

	// List markers
	ListMarker      = "-"
	ListMarkerWidth = 6

	// Default port
	DefaultPort = "8080"

	// File paths
	ConfigPath = "config.yaml"
)

// ResumeData represents the structure of the resume data
type ResumeData struct {
	Personal struct {
		Name     string `yaml:"name"`
		Title    string `yaml:"title"`
		Email    string `yaml:"email"`
		Phone    string `yaml:"phone"`
		Location string `yaml:"location"`
		Linkedin string `yaml:"linkedin"`
		Github   string `yaml:"github"`
		Summary  string `yaml:"summary"`
	} `yaml:"personal"`
	Skills []struct {
		Category string   `yaml:"category"`
		Items    []string `yaml:"items"`
	} `yaml:"skills"`
	Experience []struct {
		Company        string   `yaml:"company"`
		Position       string   `yaml:"position"`
		Location       string   `yaml:"location"`
		StartDate      string   `yaml:"startDate"`
		EndDate        string   `yaml:"endDate"`
		Description    []string `yaml:"description"`
		EmploymentType string   `yaml:"employmentType"`
		Technologies   []string `yaml:"technologies"`
	} `yaml:"experience"`
	Education []struct {
		Institution string `yaml:"institution"`
		Degree      string `yaml:"degree"`
		Location    string `yaml:"location"`
		StartDate   string `yaml:"startDate"`
		EndDate     string `yaml:"endDate"`
	} `yaml:"education"`
	Projects []struct {
		Name         string   `yaml:"name"`
		Description  string   `yaml:"description"`
		Technologies []string `yaml:"technologies"`
		Link         string   `yaml:"link"`
	} `yaml:"projects"`
	Languages []struct {
		Language    string `yaml:"language"`
		Proficiency string `yaml:"proficiency"`
	} `yaml:"languages"`
}

// loadResumeData loads the resume data from the config file
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

func main() {
	// Load resume data
	resumeData, err := loadResumeData(ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load resume data: %v", err)
	}

	// Create resume component
	resumeComponent := components.Resume(*resumeData)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Define routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := resumeComponent.Render(context.Background(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Export PDF endpoint
	http.Handle("/export-pdf", genPdf(resumeData))

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	log.Printf("Starting server on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// MustReadFile reads a file and panics if there's an error
func MustReadFile(filename string) []byte {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return data
}

func genPdf(resumeData *components.ResumeData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a new PDF document with built-in fonts
		pdf := gofpdf.New(PDFOrientation, PDFUnit, PDFSize, "")

		// Add standard fonts
		pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)

		pdf.AddPage()

		// Set font - increased sizes for better readability
		pdf.SetFont(PDFFont, FontStyleBold, FontSizeTitle)

		// Add content to PDF
		pdf.Cell(0, LineHeightText, resumeData.Personal.Name)
		pdf.Ln(LineHeightLarge)

		pdf.SetFont(PDFFont, FontStyleNormal, FontSizeSubtitle)
		pdf.Cell(0, LineHeightSmall, resumeData.Personal.Title)
		pdf.Ln(LineHeightLarge)

		// Add contact information
		pdf.SetFont(PDFFont, FontStyleBold, FontSizeHeading)
		pdf.Cell(0, LineHeightSmall, "Contact Information")
		pdf.Ln(LineHeightMedium)

		pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
		pdf.Cell(0, LineHeightText, "Email: "+resumeData.Personal.Email)
		pdf.Ln(LineHeightText)
		pdf.Cell(0, LineHeightText, "Phone: "+resumeData.Personal.Phone)
		pdf.Ln(LineHeightText)
		pdf.Cell(0, LineHeightText, "Location: "+resumeData.Personal.Location)
		pdf.Ln(LineHeightText)
		pdf.Cell(0, LineHeightText, "LinkedIn: "+resumeData.Personal.Linkedin)
		pdf.Ln(LineHeightText)
		pdf.Cell(0, LineHeightText, "GitHub: "+resumeData.Personal.Github)
		pdf.Ln(ParagraphSpacing)

		// Add experience
		pdf.SetFont(PDFFont, FontStyleBold, FontSizeHeading)
		pdf.Cell(0, LineHeightSmall, "Experience")
		pdf.Ln(LineHeightMedium)

		for _, exp := range resumeData.Experience {
			pdf.SetFont(PDFFont, FontStyleBold, FontSizeSubheading)
			pdf.Cell(0, LineHeightText, exp.Position)
			pdf.Ln(LineHeightText)

			pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
			companyInfo := exp.Company + ", " + exp.Location
			if exp.EmploymentType != "" {
				companyInfo += " (" + exp.EmploymentType + ")"
			}
			pdf.Cell(0, LineHeightText, companyInfo)
			pdf.Ln(LineHeightText)
			pdf.Cell(0, LineHeightText, exp.StartDate+" - "+exp.EndDate)
			pdf.Ln(LineHeightSmall)

			for _, desc := range exp.Description {
				// Use dash instead of bullet point
				pdf.Cell(ListMarkerWidth, LineHeightText, ListMarker)
				pdf.SetX(pdf.GetX() + ListItemIndent)
				pdf.MultiCell(0, LineHeightText, desc, "", "", false)
			}

			// Add technologies if available
			if len(exp.Technologies) > 0 {
				pdf.Ln(LineHeightSmall)
				pdf.SetFont(PDFFont, FontStyleBold, FontSizeNormal)
				pdf.Cell(0, LineHeightText, "Technologies:")
				pdf.Ln(LineHeightText)

				pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
				var techNames []string
				for _, tech := range exp.Technologies {
					techNames = append(techNames, tech.Name)
				}
				techText := strings.Join(techNames, ", ")
				pdf.MultiCell(0, LineHeightText, techText, "", "", false)
			}

			pdf.Ln(LineHeightMedium)
		}

		// Add education section only if there is data
		if len(resumeData.Education) > 0 {
			pdf.SetFont(PDFFont, FontStyleBold, FontSizeHeading)
			pdf.Cell(0, LineHeightSmall, "Education")
			pdf.Ln(LineHeightMedium)

			for _, edu := range resumeData.Education {
				pdf.SetFont(PDFFont, FontStyleBold, FontSizeSubheading)
				pdf.Cell(0, LineHeightText, edu.Degree)
				pdf.Ln(LineHeightText)

				pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
				pdf.Cell(0, LineHeightText, edu.Institution+", "+edu.Location)
				pdf.Ln(LineHeightText)
				pdf.Cell(0, LineHeightText, edu.StartDate+" - "+edu.EndDate)
				pdf.Ln(LineHeightMedium)
			}
		}

		// Add projects section only if there is data
		if len(resumeData.Projects) > 0 {
			pdf.SetFont(PDFFont, FontStyleBold, FontSizeHeading)
			pdf.Cell(0, LineHeightSmall, "Projects")
			pdf.Ln(LineHeightMedium)

			for _, project := range resumeData.Projects {
				pdf.SetFont(PDFFont, FontStyleBold, FontSizeSubheading)
				pdf.Cell(0, LineHeightText, project.Name)
				pdf.Ln(LineHeightText)

				pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
				pdf.MultiCell(0, LineHeightText, project.Description, "", "", false)
				pdf.Ln(LineHeightSmall)

				if len(project.Technologies) > 0 {
					pdf.SetFont(PDFFont, FontStyleBold, FontSizeNormal)
					pdf.Cell(0, LineHeightText, "Technologies:")
					pdf.Ln(LineHeightText)

					pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
					var techNames []string
					for _, tech := range project.Technologies {
						techNames = append(techNames, tech.Name)
					}
					techText := strings.Join(techNames, ", ")
					pdf.MultiCell(0, LineHeightText, techText, "", "", false)
					pdf.Ln(LineHeightSmall)
				}

				pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
				pdf.Cell(0, LineHeightText, "Link: "+project.Link)
				pdf.Ln(LineHeightMedium)
			}
		}

		// Add languages
		pdf.SetFont(PDFFont, FontStyleBold, FontSizeHeading)
		pdf.Cell(0, LineHeightSmall, "Languages")
		pdf.Ln(LineHeightMedium)

		for _, lang := range resumeData.Languages {
			pdf.SetFont(PDFFont, FontStyleBold, FontSizeSubheading)
			pdf.Cell(0, LineHeightText, lang.Language)
			pdf.Ln(LineHeightText)

			pdf.SetFont(PDFFont, FontStyleNormal, FontSizeNormal)
			pdf.SetX(pdf.GetX() + ListMarkerWidth)
			pdf.Cell(0, LineHeightText, lang.Proficiency)
			pdf.Ln(LineHeightMedium)
		}

		// Set response headers for PDF download
		w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
		w.Header().Set("Content-Type", "application/pdf")

		// Output PDF to response writer
		err := pdf.Output(w)
		if err != nil {
			http.Error(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
