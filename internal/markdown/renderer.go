package markdown

import (
	"bytes"
	"html/template"
	"path/filepath"

	"blaze/internal/config"
)

type Renderer struct {
	template *template.Template
	config   *config.Config
}

func NewRenderer(templateDir string, cfg *config.Config) (*Renderer, error) {
	tmplPath := filepath.Join(templateDir, "layout.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}

	return &Renderer{
		template: tmpl,
		config:   cfg,
	}, nil
}

func (r *Renderer) Render(page *Page) (string, error) {
	data := map[string]interface{}{
		"PageTitle":       r.config.PageTitle,
		"PageTitleSuffix": r.config.PageTitleSuffix,
		"BaseStyle":       "/base.css",
		"Content":         template.HTML(page.Content),
		"Explorer":        "",
		"GraphView":       "",
		"TableOfContent":  "",
		"Backlinks":       "",
	}

	for k, v := range page.Metadata {
		data[k] = v
	}

	var buf bytes.Buffer
	if err := r.template.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
