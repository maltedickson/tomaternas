package templates

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
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

func (r *Renderer) Render(w io.Writer, name string, data any) error {
	devMode := true // TODO: set to false in production
	if devMode {
		if err := r.loadTemplates(); err != nil {
			return fmt.Errorf("failed to reload templates: %w", err)
		}
	}

	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	return tmpl.ExecuteTemplate(w, name+".html", data)
}
