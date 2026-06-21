package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/maltedickson/tomaternas/internal/middleware"
	"github.com/maltedickson/tomaternas/web"
)

type Renderer struct {
	templates map[string]*template.Template
	embedFS   embed.FS
}

func NewRenderer(embedFS embed.FS) (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]*template.Template),
		embedFS:   embedFS,
	}

	if err := r.loadTemplates(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Renderer) loadTemplates() error {
	templatesDir := "templates"
	layoutFiles, err := fs.Glob(r.embedFS, filepath.Join(templatesDir, "layouts", "*.html"))
	if err != nil {
		return fmt.Errorf("getting filenames of layouts: %w", err)
	}

	partialFiles, err := fs.Glob(r.embedFS, filepath.Join(templatesDir, "partials", "*.html"))
	if err != nil {
		return fmt.Errorf("getting filenames of partials: %w", err)
	}

	pageFiles, err := fs.Glob(r.embedFS, filepath.Join(templatesDir, "pages", "*.html"))
	if err != nil {
		return fmt.Errorf("getting filenames of pages: %w", err)
	}

	for _, pageFile := range pageFiles {
		templateName := strings.TrimPrefix(pageFile, filepath.Join(templatesDir, "pages")+string(filepath.Separator))
		templateName = strings.TrimSuffix(templateName, ".html")

		files := make([]string, 0)
		files = append(files, layoutFiles...)
		files = append(files, partialFiles...)
		files = append(files, pageFile)

		tmpl, err := template.New(pageFile).Funcs(r.funcMap()).Funcs(web.IconFuncMap()).ParseFS(r.embedFS, files...)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", templateName, err)
		}

		r.templates[templateName] = tmpl
	}

	return nil
}

func (r *Renderer) funcMap() template.FuncMap {
	return template.FuncMap{
		"toFloat": func(v any) float64 {
			switch val := v.(type) {
			case int:
				return float64(val)
			case float64:
				return val
			default:
				return 0.0
			}
		},
		"roundToDecimals": func(f float64, count int) float64 {
			exp := math.Pow10(count)
			return math.Round(f*exp) / exp
		},
		"printWith1Decimal": func(f float64) string {
			return fmt.Sprintf("%.1f", f)
		},
		"roundToHalf": func(f float64) float64 {
			return math.Round(f*2) / 2
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"formatDate": func(t any) string {
			return fmt.Sprintf("%v", t)
		},
		"proseDate": func(t time.Time) string {
			swedishMonths := map[time.Month]string{
				time.January:   "jan",
				time.February:  "feb",
				time.March:     "mar",
				time.April:     "apr",
				time.May:       "maj",
				time.June:      "jun",
				time.July:      "jul",
				time.August:    "aug",
				time.September: "sep",
				time.October:   "okt",
				time.November:  "nov",
				time.December:  "dec",
			}

			return fmt.Sprintf("%d %s %d", t.Day(), swedishMonths[t.Month()], t.Year())
		},
	}
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, templateName string, localData any) error {
	tmpl, ok := r.templates[templateName]
	if !ok {
		log.Printf("template %s not found", templateName)
		http.Error(w, "Något gick fel.", http.StatusInternalServerError)
		return nil
	}

	user, _ := middleware.GetUser(req)

	templateData := map[string]any{
		"G": map[string]any{
			"User": user,
			"Path": req.URL.Path,
		},
		"L": localData,
	}

	return tmpl.ExecuteTemplate(w, templateName+".html", templateData)
}

func (r *Renderer) RenderErr(w http.ResponseWriter, req *http.Request, statusCode int, message string) {
	templateName := "error"

	tmpl, ok := r.templates[templateName]
	if !ok {
		log.Printf("template %s not found", templateName)
		http.Error(w, "Något gick fel.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	user, _ := middleware.GetUser(req)

	templateData := map[string]any{
		"G": map[string]any{
			"User": user,
			"Path": req.URL.Path,
		},
		"L": map[string]any{
			"StatusCode": statusCode,
			"Message":    message,
		},
	}

	tmpl.ExecuteTemplate(w, templateName+".html", templateData)
}
