package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fairytale5571/cv_onopchenko/components"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/jung-kurt/gofpdf"
	"gopkg.in/yaml.v3"
)

const (
	// PDF settings
	PDFOrientation = "P"
	PDFUnit        = "mm"
	PDFSize        = "A4"
	PDFFont        = "helvetica"

	// Font sizes
	FontSizeTitle      = 24
	FontSizeSubtitle   = 18
	FontSizeHeading    = 16
	FontSizeSubheading = 12
	FontSizeNormal     = 10
	FontSizeSmall      = 8

	// Spacing
	LineHeightLarge  = 16
	LineHeightMedium = 12
	LineHeightSmall  = 10
	LineHeightText   = 8
	ParagraphSpacing = 16
	SectionSpacing   = 12
	ListItemIndent   = 6

	// Default port
	DefaultPort = "8081"

	// File paths
	ConfigPath = "config.yaml"
)

// Colors for the document - Modern color scheme matching website
var (
	HeaderBgColor    = props.Color{Red: 0, Green: 51, Blue: 102}    // Deep blue (matches website header)
	HeaderTextColor  = props.Color{Red: 255, Green: 255, Blue: 255} // White
	PrimaryColor     = props.Color{Red: 17, Green: 24, Blue: 39}    // Dark blue/gray
	SecondaryBgColor = props.Color{Red: 243, Green: 244, Blue: 246} // Light gray-blue
	BorderColor      = props.Color{Red: 229, Green: 231, Blue: 235} // Light gray
	GrayColor        = props.Color{Red: 107, Green: 114, Blue: 128} // Medium gray
	LinkColor        = props.Color{Red: 0, Green: 51, Blue: 102}    // Deep blue (matches website header)
	AccentColor      = props.Color{Red: 55, Green: 48, Blue: 163}   // Darker indigo for accents
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
		Category string `yaml:"category"`
		Items    []struct {
			Name string `yaml:"name"`
			Link string `yaml:"link,omitempty"`
		} `yaml:"items"`
	} `yaml:"skills"`
	Experience []struct {
		Company        string   `yaml:"company"`
		Position       string   `yaml:"position"`
		Location       string   `yaml:"location"`
		StartDate      string   `yaml:"startDate"`
		EndDate        string   `yaml:"endDate"`
		Description    []string `yaml:"description"`
		EmploymentType string   `yaml:"employmentType"`
		Technologies   []struct {
			Name string `yaml:"name"`
			Link string `yaml:"link,omitempty"`
		} `yaml:"technologies"`
	} `yaml:"experience"`
	Education []struct {
		Institution string `yaml:"institution"`
		Degree      string `yaml:"degree"`
		Location    string `yaml:"location"`
		StartDate   string `yaml:"startDate"`
		EndDate     string `yaml:"endDate"`
	} `yaml:"education"`
	Projects []struct {
		Name         string `yaml:"name"`
		Description  string `yaml:"description"`
		Technologies []struct {
			Name string `yaml:"name"`
			Link string `yaml:"link,omitempty"`
		} `yaml:"technologies"`
		Link string `yaml:"link"`
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

	// Export PDF endpoint using the new Maroto library
	http.Handle("/export-pdf", genMarotoPdf(resumeData))

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	log.Printf("Starting server on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// genMarotoPdf generates a PDF using the Maroto v2 library
func genMarotoPdf(resumeData *components.ResumeData) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем gofpdf для регистрации шрифта
		pdf := gofpdf.New(PDFOrientation, PDFUnit, PDFSize, "")
		// Регистрируем шрифт Montserrat-Medium
		// Примечание: он не будет использоваться ниже из-за ограничений Maroto
		pdf.AddFont("montserrat", "", "static/fonts/Montserrat-Medium.ttf")

		// Create a new maroto instance
		m := maroto.New()

		// Header section with name/title (reduced height)
		headerRow := row.New(16)
		headerCol := col.New(12)

		// Add name text with modern styling
		nameComponent := text.New(resumeData.Personal.Name, props.Text{
			Size:  FontSizeTitle + 2, // Larger font size for more impact
			Style: fontstyle.Bold,
			Align: align.Center,
			Color: &HeaderBgColor, // Use the vibrant blue color
		})
		headerCol.Add(nameComponent)
		headerRow.Add(headerCol)
		m.AddRows(headerRow)

		// Add title row (reduced height)
		titleRow := row.New(8)
		titleCol := col.New(12)

		// Add title/position text with modern styling
		titleComponent := text.New(resumeData.Personal.Title, props.Text{
			Size:  FontSizeSubtitle,
			Style: fontstyle.Normal,
			Align: align.Center,
			Color: &LinkColor, // Use the vibrant blue color
		})
		titleCol.Add(titleComponent)
		titleRow.Add(titleCol)
		m.AddRows(titleRow)

		// Add spacing
		m.AddRows(row.New(10))

		// Contact Information Section with modern styling
		sectionTitleRow := row.New(12) // Slightly taller row for better spacing
		sectionTitleCol := col.New(12)
		contactTitleComponent := text.New("Contact Information", props.Text{
			Size:  FontSizeHeading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &HeaderBgColor, // Use the vibrant blue color for consistency
			Top:   2, // Add a small top padding
		})
		sectionTitleCol.Add(contactTitleComponent)
		sectionTitleRow.Add(sectionTitleCol)
		m.AddRows(sectionTitleRow)

 	// Add a modern styled divider
 	dividerRow := row.New(2) // Slightly taller for better visibility
 	dividerCol := col.New(12)
 	dividerText := strings.Repeat("━", 60) // Thicker Unicode horizontal line for modern look
 	dividerComponent := text.New(dividerText, props.Text{
 		Size:  FontSizeSmall,
 		Align: align.Left,
 		Color: &HeaderBgColor, // Match section title color
 		Top:   0, // No top padding
 	})
 	dividerCol.Add(dividerComponent)
 	dividerRow.Add(dividerCol)
 	m.AddRows(dividerRow)
 	m.AddRows(row.New(2)) // Add minimal space after divider

		// Contact details
		addContactInfo(m, "Email:", resumeData.Personal.Email)
		addContactInfo(m, "Phone:", resumeData.Personal.Phone)
		addContactInfo(m, "Location:", resumeData.Personal.Location)
		addContactInfo(m, "LinkedIn:", resumeData.Personal.Linkedin)
		addContactInfo(m, "GitHub:", resumeData.Personal.Github)

		// Add spacing (reduced)
		m.AddRows(row.New(8))

		// Experience section with modern styling (reduced height)
		experienceTitleRow := row.New(10)
		experienceTitleCol := col.New(12)
		experienceTitleComponent := text.New("Experience", props.Text{
			Size:  FontSizeHeading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &HeaderBgColor, // Use vibrant blue for section title
			Top:   2, // Add a small top padding
		})
		experienceTitleCol.Add(experienceTitleComponent)
		experienceTitleRow.Add(experienceTitleCol)
		m.AddRows(experienceTitleRow)

		// Add a modern styled divider under the section title
		dividerRow = row.New(2)
		dividerCol = col.New(12)
		dividerText = strings.Repeat("━", 40) // Thicker Unicode horizontal line
		dividerComponent = text.New(dividerText, props.Text{
			Size:  FontSizeSmall,
			Align: align.Left,
			Color: &HeaderBgColor, // Match section title color
		})
		dividerCol.Add(dividerComponent)
		dividerRow.Add(dividerCol)
		m.AddRows(dividerRow)
		m.AddRows(row.New(2)) // Add minimal space after divider

		// Add subtitle with modern styling
		subtitleRow := row.New(6)
		subtitleCol := col.New(12)
		subtitleComponent := text.New("Professional work history", props.Text{
			Size:  FontSizeNormal,
			Style: fontstyle.Italic,
			Align: align.Left,
			Color: &AccentColor, // Use accent color for subtitles
			Top:   1, // Add a small top padding
		})
		subtitleCol.Add(subtitleComponent)
		subtitleRow.Add(subtitleCol)
		m.AddRows(subtitleRow)

		m.AddRows(row.New(5))

		// Add each experience
		for _, exp := range resumeData.Experience {
			addExperienceItem(m, exp.Position, exp.Company, exp.StartDate+" - "+exp.EndDate, exp.Location, exp.EmploymentType, strings.Join(exp.Description, "\n"))
		}

		// Education section if available
		if len(resumeData.Education) > 0 {
 		// Education section with modern styling (reduced spacing)
 		m.AddRows(row.New(8))
 		educationTitleRow := row.New(10)
			educationTitleCol := col.New(12)
			educationTitleComponent := text.New("Education", props.Text{
				Size:  FontSizeHeading,
				Style: fontstyle.Bold,
				Align: align.Left,
				Color: &HeaderBgColor, // Use vibrant blue for section title
				Top:   2, // Add a small top padding
			})
			educationTitleCol.Add(educationTitleComponent)
			educationTitleRow.Add(educationTitleCol)
			m.AddRows(educationTitleRow)

			// Add a modern styled divider under the section title
			dividerRow = row.New(2)
			dividerCol = col.New(12)
			dividerText = strings.Repeat("━", 40) // Thicker Unicode horizontal line
			dividerComponent = text.New(dividerText, props.Text{
				Size:  FontSizeSmall,
				Align: align.Left,
				Color: &HeaderBgColor, // Match section title color
			})
			dividerCol.Add(dividerComponent)
			dividerRow.Add(dividerCol)
			m.AddRows(dividerRow)
			m.AddRows(row.New(2)) // Add minimal space after divider

			// Add subtitle for education section with modern styling
			subtitleRow = row.New(6)
			subtitleCol = col.New(12)
			subtitleComponent = text.New("Academic background and qualifications", props.Text{
				Size:  FontSizeNormal,
				Style: fontstyle.Italic,
				Align: align.Left,
				Color: &AccentColor, // Use accent color for subtitles
				Top:   1, // Add a small top padding
			})
			subtitleCol.Add(subtitleComponent)
			subtitleRow.Add(subtitleCol)
			m.AddRows(subtitleRow)

			m.AddRows(row.New(5))

			// Add each education
			for _, edu := range resumeData.Education {
				addEducationItem(m, edu.Degree, edu.Institution, edu.StartDate+" - "+edu.EndDate, edu.Location)
			}
		}

		// Skills section with modern styling (reduced spacing)
		m.AddRows(row.New(8))
		skillsTitleRow := row.New(10)
		skillsTitleCol := col.New(12)
		skillsTitleComponent := text.New("Skills", props.Text{
			Size:  FontSizeHeading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &HeaderBgColor, // Use vibrant blue for section title
			Top:   2, // Add a small top padding
		})
		skillsTitleCol.Add(skillsTitleComponent)
		skillsTitleRow.Add(skillsTitleCol)
		m.AddRows(skillsTitleRow)

		// Add a modern styled divider under the section title
		dividerRow = row.New(2)
		dividerCol = col.New(12)
		dividerText = strings.Repeat("━", 40) // Thicker Unicode horizontal line
		dividerComponent = text.New(dividerText, props.Text{
			Size:  FontSizeSmall,
			Align: align.Left,
			Color: &HeaderBgColor, // Match section title color
		})
		dividerCol.Add(dividerComponent)
		dividerRow.Add(dividerCol)
		m.AddRows(dividerRow)
		m.AddRows(row.New(2)) // Add minimal space after divider

		// Add subtitle for skills section with modern styling
		subtitleRow = row.New(6)
		subtitleCol = col.New(12)
		subtitleComponent = text.New("Technical competencies and expertise", props.Text{
			Size:  FontSizeNormal,
			Style: fontstyle.Italic,
			Align: align.Left,
			Color: &AccentColor, // Use accent color for subtitles
			Top:   1, // Add a small top padding
		})
		subtitleCol.Add(subtitleComponent)
		subtitleRow.Add(subtitleCol)
		m.AddRows(subtitleRow)

		m.AddRows(row.New(5))

		// Add skills section
		addSkillsSection(m, resumeData.Skills)

		// Languages section if available
		if len(resumeData.Languages) > 0 {
			langTitleRow := row.New(10)
			langTitleCol := col.New(12)
			langTitleComponent := text.New("Languages", props.Text{
				Size:  FontSizeHeading,
				Style: fontstyle.Bold,
				Align: align.Left,
				Color: &PrimaryColor,
			})
			langTitleCol.Add(langTitleComponent)
			langTitleRow.Add(langTitleCol)
			m.AddRows(langTitleRow)

			// Languages list
			langsText := ""
			for i, lang := range resumeData.Languages {
				langsText += lang.Language + " (" + lang.Proficiency + ")"
				if i < len(resumeData.Languages)-1 {
					langsText += ", "
				}
			}

			langsRow := row.New(6)
			langsCol := col.New(12)
			langsComponent := text.New(langsText, props.Text{
				Size:  FontSizeNormal,
				Style: fontstyle.Normal,
				Align: align.Left,
				Color: &PrimaryColor,
			})
			langsCol.Add(langsComponent)
			langsRow.Add(langsCol)
			m.AddRows(langsRow)
		}

		// Generate PDF
		fileName := fmt.Sprintf("resume_%s.pdf", time.Now().Format("20060102_150405"))
		doc, err := m.Generate()
		if err != nil {
			http.Error(w, "Failed to generate PDF: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Set response headers for PDF download
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(doc.GetBytes())))

		// Write PDF to response
		_, err = w.Write(doc.GetBytes())
		if err != nil {
			http.Error(w, "Failed to write PDF to response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// addContactInfo adds a contact information row with modern styling
func addContactInfo(m core.Maroto, label, value string) {
	contactRow := row.New(7) // Slightly taller row for better spacing

	// Create label column (3 units wide)
	labelCol := col.New(3)
	labelComponent := text.New(label, props.Text{
		Size:  FontSizeNormal,
		Style: fontstyle.Bold, // Bold for labels
		Align: align.Left,
		Color: &HeaderBgColor, // Use vibrant blue for labels
	})
	labelCol.Add(labelComponent)

	// Create value column (9 units wide)
	valueCol := col.New(9)
	valueComponent := text.New(value, props.Text{
		Size:  FontSizeNormal,
		Style: fontstyle.Normal,
		Align: align.Left,
		Color: &PrimaryColor, // Dark color for values
	})
	valueCol.Add(valueComponent)

	// Add both columns to the row
	contactRow.Add(labelCol, valueCol)
	m.AddRows(contactRow)
}

// addExperienceItem adds a single experience item to the CV with modern styling
func addExperienceItem(m core.Maroto, position, company, period, location, employmentType, description string) {
	// Add minimal spacing before each experience item
	m.AddRows(row.New(3))

	// Create title row for position at company
	titleRow := row.New(10)

	// Position part with modern styling
	positionCol := col.New(8)
	positionComponent := text.New(position, props.Text{
		Top:   2,
		Size:  FontSizeSubheading,
		Style: fontstyle.Bold,
		Align: align.Left,
		Color: &HeaderBgColor, // Use vibrant blue for position title
	})
	positionCol.Add(positionComponent)

	// Period part (right aligned) with modern styling
	periodCol := col.New(4)
	periodComponent := text.New(period, props.Text{
		Top:   2,
		Size:  FontSizeNormal,
		Style: fontstyle.Bold,
		Align: align.Right,
		Color: &GrayColor, // Use gray for dates
	})
	periodCol.Add(periodComponent)

	titleRow.Add(positionCol, periodCol)
	m.AddRows(titleRow)

	// Create company and location row with modern styling
	companyRow := row.New(7)

	// Company part
	companyCol := col.New(8)
	companyComponent := text.New(company, props.Text{
		Top:   1,
		Size:  FontSizeNormal,
		Style: fontstyle.Bold, // Bold instead of italic for better readability
		Align: align.Left,
		Color: &PrimaryColor, // Dark color for company name
	})
	companyCol.Add(companyComponent)

	// Location part (right aligned)
	locationCol := col.New(4)
	locationComponent := text.New(location, props.Text{
		Top:   1,
		Size:  FontSizeNormal,
		Style: fontstyle.Italic,
		Align: align.Right,
		Color: &GrayColor, // Gray for location
	})
	locationCol.Add(locationComponent)

	companyRow.Add(companyCol, locationCol)
	m.AddRows(companyRow)

	// Employment Type row (if provided) with modern badge-like styling
	if employmentType != "" {
		employmentRow := row.New(7)
		employmentCol := col.New(12)
		// Create employment type as a badge
		employmentTypeText := "• " + employmentType + " •" // Add bullet points for badge-like appearance
		employmentComponent := text.New(employmentTypeText, props.Text{
			Top:   1,
			Size:  FontSizeNormal,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &HeaderBgColor, // Vibrant blue for employment type
		})
		employmentCol.Add(employmentComponent)
		employmentRow.Add(employmentCol)
		m.AddRows(employmentRow)
	}

	// Add descriptions as modern bullet points
	if description != "" {
		// Add minimal space before description
		m.AddRows(row.New(2))

		// Split description by newlines to create bullet points
		items := strings.Split(description, "\n")

		for _, item := range items {
			if strings.TrimSpace(item) == "" {
				continue
			}

			// Ensure each bullet point fits within column
			wrappedItem := ensureTextFitsInColumn(item, 80)
			bulletPoint := "• " + wrappedItem

			// Replace newlines in wrapped item with space and bullet
			bulletPoint = strings.Replace(bulletPoint, "\n", "\n  ", -1)

			descriptionRow := row.New(6) // Moderate height for readability while saving space
			descriptionCol := col.New(12)
			descriptionComponent := text.New(bulletPoint, props.Text{
				Top:   1.5,
				Size:  FontSizeNormal,
				Style: fontstyle.Normal,
				Align: align.Left,
				Color: &PrimaryColor, // Dark color for better readability
			})
			descriptionCol.Add(descriptionComponent)
			descriptionRow.Add(descriptionCol)
			m.AddRows(descriptionRow)
		}
	}

	// Add minimal space before separator
	m.AddRows(row.New(4))

	// Add a modern separator between experiences
	separatorRow := row.New(3)
	separatorCol := col.New(12)
	// Use a more elegant separator with dots
	separatorText := "• • • • •" // Five dots with spaces for a cleaner look
	separatorComponent := text.New(separatorText, props.Text{
		Size:  FontSizeSmall,
		Align: align.Center,
		Color: &GrayColor, // Gray for subtle appearance
		Style: fontstyle.Normal,
	})
	separatorCol.Add(separatorComponent)
	separatorRow.Add(separatorCol)
	m.AddRows(separatorRow)

	// Add minimal space after separator
	m.AddRows(row.New(4))
}

// addEducationItem adds a single education item to the CV with modern styling
func addEducationItem(m core.Maroto, degree, institution, period, location string) {
	// Add minimal spacing before education item
	m.AddRows(row.New(2))

	// Create degree row with modern styling
	degreeRow := row.New(8)

	// Degree part
	degreeCol := col.New(8)
	degreeComponent := text.New(degree, props.Text{
		Top:   1,
		Size:  FontSizeSubheading,
		Style: fontstyle.Bold,
		Align: align.Left,
		Color: &HeaderBgColor, // Use vibrant blue for degree title
	})
	degreeCol.Add(degreeComponent)

	// Period part (right aligned)
	periodCol := col.New(4)
	periodComponent := text.New(period, props.Text{
		Top:   1,
		Size:  FontSizeNormal,
		Style: fontstyle.Normal,
		Align: align.Right,
		Color: &GrayColor, // Gray for dates
	})
	periodCol.Add(periodComponent)

	degreeRow.Add(degreeCol, periodCol)
	m.AddRows(degreeRow)

	// Create institution and location row with modern styling
	institutionRow := row.New(7)

	// Institution part
	institutionCol := col.New(8)
	institutionComponent := text.New(institution, props.Text{
		Top:   1,
		Size:  FontSizeNormal,
		Style: fontstyle.Bold, // Bold instead of italic for better readability
		Align: align.Left,
		Color: &PrimaryColor, // Dark color for institution name
	})
	institutionCol.Add(institutionComponent)

	// Location part (right aligned)
	locationCol := col.New(4)
	locationComponent := text.New(location, props.Text{
		Top:   1,
		Size:  FontSizeNormal,
		Style: fontstyle.Italic,
		Align: align.Right,
		Color: &GrayColor, // Gray for location
	})
	locationCol.Add(locationComponent)

	institutionRow.Add(institutionCol, locationCol)
	m.AddRows(institutionRow)

	// Add a subtle separator after each education item
	separatorRow := row.New(4)
	separatorCol := col.New(12)
	separatorText := "· · ·" // Three dots for a subtle separator
	separatorComponent := text.New(separatorText, props.Text{
		Size:  FontSizeSmall,
		Align: align.Center,
		Color: &GrayColor, // Gray for subtle appearance
		Style: fontstyle.Normal,
	})
	separatorCol.Add(separatorComponent)
	separatorRow.Add(separatorCol)
	m.AddRows(separatorRow)
}

// addSkillsSection adds the skills section with categories and modern styling
func addSkillsSection(m core.Maroto, skills []struct {
	Category string `yaml:"category"`
	Items    []struct {
		Name string `yaml:"name"`
		Link string `yaml:"link,omitempty"`
	} `yaml:"items"`
}) {
	// Process each category
	for i, category := range skills {
		// Add minimal spacing before categories (except the first one)
		if i > 0 {
			m.AddRows(row.New(3))
		}

		// Category heading with modern styling
		categoryRow := row.New(8)
		categoryCol := col.New(12)
		categoryComponent := text.New(category.Category, props.Text{
			Top:   1,
			Size:  FontSizeSubheading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &HeaderBgColor, // Use vibrant blue for category headings
		})
		categoryCol.Add(categoryComponent)
		categoryRow.Add(categoryCol)
		m.AddRows(categoryRow)

		// Add a small divider under the category
		dividerRow := row.New(1)
		dividerCol := col.New(12)
		dividerText := strings.Repeat("━", 20) // Short divider for visual separation
		dividerComponent := text.New(dividerText, props.Text{
			Size:  FontSizeSmall,
			Align: align.Left,
			Color: &HeaderBgColor, // Match category heading color
		})
		dividerCol.Add(dividerComponent)
		dividerRow.Add(dividerCol)
		m.AddRows(dividerRow)
		m.AddRows(row.New(2))

		// Format skills in a more readable way
		var formattedSkills []string
		for _, item := range category.Items {
			// Add a bullet point before each skill for better readability
			formattedSkills = append(formattedSkills, "• "+item.Name)
		}

		// Create skill groups to display in multiple columns visually
		skillGroups := groupSkills(formattedSkills, 3) // Group skills into sets of 3

		for _, group := range skillGroups {
			skillsText := strings.Join(group, "   ")
			wrappedSkills := ensureTextFitsInColumn(skillsText, 90)

			skillsRow := row.New(6)
			skillsCol := col.New(12)
			skillsComponent := text.New(wrappedSkills, props.Text{
				Top:   1,
				Size:  FontSizeNormal,
				Style: fontstyle.Normal,
				Align: align.Left,
				Color: &PrimaryColor, // Dark color for better readability
			})
			skillsCol.Add(skillsComponent)
			skillsRow.Add(skillsCol)
			m.AddRows(skillsRow)
		}

		// Add minimal space after each category
		m.AddRows(row.New(1))
	}
}

// groupSkills groups skills into smaller sets for better visual presentation
func groupSkills(skills []string, groupSize int) [][]string {
	var groups [][]string

	for i := 0; i < len(skills); i += groupSize {
		end := i + groupSize
		if end > len(skills) {
			end = len(skills)
		}
		groups = append(groups, skills[i:end])
	}

	return groups
}

// ensureTextFitsInColumn splits text into multiple lines if it exceeds the available width
func ensureTextFitsInColumn(text string, maxCharsPerLine int) string {
	if len(text) <= maxCharsPerLine {
		return text
	}

	words := strings.Fields(text)
	var result strings.Builder
	currentLine := ""

	for _, word := range words {
		// If adding this word would exceed the max chars per line
		if len(currentLine)+len(word)+1 > maxCharsPerLine {
			// Add the current line to result and start a new line
			if currentLine != "" {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			} else {
				// If a single word is too long, just add it anyway
				currentLine = word
			}
		} else {
			// Add word to current line
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	// Add the last line
	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}

// addLanguageSection adds the languages section with modern styling
func addLanguageSection(m core.Maroto, languages []struct {
	Language    string `yaml:"language"`
	Proficiency string `yaml:"proficiency"`
}) {
	if len(languages) == 0 {
		return
	}

	// Add minimal spacing before languages section
	m.AddRows(row.New(3))

	// Create language section title with modern styling
	languageRow := row.New(10)
	languageCol := col.New(12)
	languageComponent := text.New("Languages", props.Text{
		Size:  FontSizeHeading,
		Style: fontstyle.Bold,
		Align: align.Left,
		Color: &HeaderBgColor, // Use vibrant blue for section title
		Top:   2, // Add a small top padding
	})
	languageCol.Add(languageComponent)
	languageRow.Add(languageCol)
	m.AddRows(languageRow)

	// Add a divider under the section title
	dividerRow := row.New(2)
	dividerCol := col.New(12)
	dividerText := strings.Repeat("━", 30) // Medium-length divider for visual separation
	dividerComponent := text.New(dividerText, props.Text{
		Size:  FontSizeSmall,
		Align: align.Left,
		Color: &HeaderBgColor, // Match section title color
	})
	dividerCol.Add(dividerComponent)
	dividerRow.Add(dividerCol)
	m.AddRows(dividerRow)
	m.AddRows(row.New(2)) // Add minimal space after divider

	// Create a more visually appealing language list
	// Display each language on its own row with a bullet point
	for _, lang := range languages {
		langRow := row.New(7)
		langCol := col.New(12)

		// Format as "• Language — Proficiency" with an em dash for better visual separation
		langText := "• " + lang.Language + " — " + lang.Proficiency

		langComponent := text.New(langText, props.Text{
			Size:  FontSizeNormal,
			Style: fontstyle.Normal,
			Align: align.Left,
			Color: &PrimaryColor, // Dark color for better readability
		})
		langCol.Add(langComponent)
		langRow.Add(langCol)
		m.AddRows(langRow)
	}

	// Add minimal space after languages section
	m.AddRows(row.New(3))
}

// generatePDF generates a PDF from the resume data
func generatePDF(resumeData ResumeData) (*bytes.Buffer, error) {
	// Create PDF
	m := maroto.New()

	// Add page header
	headerRow := row.New(32)

	// Name and title column
	nameCol := col.New(8)
	nameComponent := text.New(resumeData.Personal.Name, props.Text{
		Size:  FontSizeTitle,
		Style: fontstyle.Bold,
		Align: align.Left,
		Color: &PrimaryColor,
		Top:   3,
	})
	nameCol.Add(nameComponent)

	titleComponent := text.New(resumeData.Personal.Title, props.Text{
		Size:  FontSizeSubheading,
		Style: fontstyle.Italic,
		Align: align.Left,
		Color: &GrayColor,
		Top:   12,
	})
	nameCol.Add(titleComponent)

	// Summary if available
	if resumeData.Personal.Summary != "" {
		summaryComponent := text.New(resumeData.Personal.Summary, props.Text{
			Size:  FontSizeNormal,
			Style: fontstyle.Normal,
			Align: align.Left,
			Top:   20,
		})
		nameCol.Add(summaryComponent)
	}

	// Profile picture if available
	imgCol := col.New(4)
	// Here you would add a profile picture if available

	headerRow.Add(nameCol, imgCol)
	m.AddRows(headerRow)

	// Add contact information
	// Add each contact method
	addContactInfo(m, "Email:", resumeData.Personal.Email)
	addContactInfo(m, "Phone:", resumeData.Personal.Phone)
	addContactInfo(m, "Location:", resumeData.Personal.Location)
	if resumeData.Personal.Linkedin != "" {
		addContactInfo(m, "LinkedIn:", resumeData.Personal.Linkedin)
	}
	if resumeData.Personal.Github != "" {
		addContactInfo(m, "GitHub:", resumeData.Personal.Github)
	}

	// Add minimal spacing
	m.AddRows(row.New(5))

	// Experience section
	if len(resumeData.Experience) > 0 {
		// Add section title
		expTitleRow := row.New(8)
		expTitleCol := col.New(12)
		expTitleComponent := text.New("Experience", props.Text{
			Size:  FontSizeHeading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &PrimaryColor,
		})
		expTitleCol.Add(expTitleComponent)
		expTitleRow.Add(expTitleCol)
		m.AddRows(expTitleRow)

		// Add each experience item
		for _, exp := range resumeData.Experience {
			addExperienceItem(m, exp.Position, exp.Company, exp.StartDate+" - "+exp.EndDate, exp.Location, exp.EmploymentType, strings.Join(exp.Description, "\n"))
		}
	}

	// Education section
	if len(resumeData.Education) > 0 {
		// Add section title
		eduTitleRow := row.New(8)
		eduTitleCol := col.New(12)
		eduTitleComponent := text.New("Education", props.Text{
			Size:  FontSizeHeading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &PrimaryColor,
		})
		eduTitleCol.Add(eduTitleComponent)
		eduTitleRow.Add(eduTitleCol)
		m.AddRows(eduTitleRow)

		// Add each education item
		for _, edu := range resumeData.Education {
			addEducationItem(m, edu.Degree, edu.Institution, edu.StartDate+" - "+edu.EndDate, edu.Location)
		}
	}

	// Skills section
	if len(resumeData.Skills) > 0 {
		// Add section title
		skillsTitleRow := row.New(8)
		skillsTitleCol := col.New(12)
		skillsTitleComponent := text.New("Skills", props.Text{
			Size:  FontSizeHeading,
			Style: fontstyle.Bold,
			Align: align.Left,
			Color: &PrimaryColor,
		})
		skillsTitleCol.Add(skillsTitleComponent)
		skillsTitleRow.Add(skillsTitleCol)
		m.AddRows(skillsTitleRow)

		// Add skills section
		addSkillsSection(m, resumeData.Skills)
	}

	// Languages section if available
	if len(resumeData.Languages) > 0 {
		addLanguageSection(m, resumeData.Languages)
	}

	pdfBuffer, err := m.Generate()
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(pdfBuffer.GetBytes()), nil
}
