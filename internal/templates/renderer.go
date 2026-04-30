package templates

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"recipe-web-server/internal/middleware"
	"strings"
)

type Renderer struct {
	templates map[string]*template.Template
}

func NewRenderer() (*Renderer, error) {
	r := &Renderer{
		templates: make(map[string]*template.Template),
	}

	if err := r.loadTemplates(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Renderer) loadTemplates() error {
	templatesDir := "web/templates"

	layoutFiles, err := filepath.Glob(filepath.Join(templatesDir, "layouts", "*.html"))
	if err != nil {
		return fmt.Errorf("failed to read layouts: %w", err)
	}

	partialFiles, err := filepath.Glob(filepath.Join(templatesDir, "partials", "*.html"))
	if err != nil {
		return fmt.Errorf("failed to read partials: %w", err)
	}

	pageFiles, err := filepath.Glob(filepath.Join(templatesDir, "pages", "*.html"))
	if err != nil {
		return fmt.Errorf("failed to read pages: %w", err)
	}

	for _, pageFile := range pageFiles {
		templateName := strings.TrimPrefix(pageFile, filepath.Join(templatesDir, "pages")+string(filepath.Separator))
		templateName = strings.TrimSuffix(templateName, ".html")

		files := make([]string, 0)
		files = append(files, layoutFiles...)
		files = append(files, partialFiles...)
		files = append(files, pageFile)

		tmpl, err := template.New(pageFile).Funcs(r.funcMap()).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %w", templateName, err)
		}

		r.templates[templateName] = tmpl
	}

	return nil
}

func (r *Renderer) funcMap() template.FuncMap {
	return template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"formatDate": func(t any) string {
			return fmt.Sprintf("%v", t)
		},
	}
}

func (r *Renderer) Render(w http.ResponseWriter, req *http.Request, templateName string, pageTitle string, localData any) error {
	devMode := true // TODO: set to false in production
	if devMode {
		if err := r.loadTemplates(); err != nil {
			return fmt.Errorf("failed to reload templates: %w", err)
		}
	}

	tmpl, ok := r.templates[templateName]
	if !ok {
		log.Printf("template (%s) not found", templateName)
		http.Error(w, "Något gick fel.", http.StatusInternalServerError)
		return nil
	}

	user, _ := middleware.GetUser(req)

	templateData := map[string]any{
		"G": map[string]any{
			"User":  user,
			"Path":  req.URL.Path,
			"Title": pageTitle,
		},
		"L": localData,
	}

	return tmpl.ExecuteTemplate(w, templateName+".html", templateData)
}

func (r *Renderer) RenderErr(w http.ResponseWriter, req *http.Request, statusCode int, message string) {
	devMode := true // TODO: set to false in production
	if devMode {
		if err := r.loadTemplates(); err != nil {
			log.Printf("failed to reload templates: %v", err)
		}
	}

	templateName := "error"

	tmpl, ok := r.templates[templateName]
	if !ok {
		log.Printf("template (%s) not found", templateName)
		http.Error(w, "Något gick fel.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	user, _ := middleware.GetUser(req)

	templateData := map[string]any{
		"G": map[string]any{
			"User":  user,
			"Path":  req.URL.Path,
			"Title": fmt.Sprintf("%d - %s", statusCode, message),
		},
		"L": map[string]any{
			"StatusCode": statusCode,
			"Message":    message,
		},
	}

	tmpl.ExecuteTemplate(w, templateName+".html", templateData)
}
