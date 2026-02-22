package templates

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	templates *template.Template
}

func NewRenderer() (*Renderer, error) {
	templates, err := template.ParseGlob(filepath.Join("web", "templates", "**", "*.html"))
	if err != nil {
		return nil, err
	}
	return &Renderer{templates: templates}, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) error {
	return r.templates.ExecuteTemplate(w, name+".html", data)
}
