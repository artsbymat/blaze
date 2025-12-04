package renderer

import (
	"bytes"
	"html/template"
	"path/filepath"

	"blaze/internal/components"
	"blaze/internal/config"
	"blaze/internal/markdown"
)

type HTMLRenderer struct {
	template         *template.Template
	config           *config.Config
	componentFactory *components.ComponentFactory
	explorerCache    template.HTML
}

func NewHTMLRenderer(templateDir string, cfg *config.Config) (*HTMLRenderer, error) {
	tmplPath := filepath.Join(templateDir, "layout.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}

	return &HTMLRenderer{
		template:         tmpl,
		config:           cfg,
		componentFactory: components.NewComponentFactory(cfg),
	}, nil
}

func (r *HTMLRenderer) RegenerateExplorer(contentDir string) error {
	explorer := r.componentFactory.CreateExplorer(contentDir)
	explorerHTML, err := explorer.Generate()
	if err != nil {
		return err
	}
	r.explorerCache = explorerHTML
	return nil
}

func (r *HTMLRenderer) Render(page *markdown.Page) (string, error) {
	data := map[string]any{
		"PageTitle":       r.config.PageTitle,
		"PageTitleSuffix": r.config.PageTitleSuffix,
		"Locale":          r.config.Locale,
		"Content":         template.HTML(page.HTMLContent),
		"Explorer":        r.explorerCache,
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

func (r *HTMLRenderer) RenderPage(htmlContent string, metadata map[string]string) (string, error) {
	data := map[string]any{
		"PageTitle":       r.config.PageTitle,
		"PageTitleSuffix": r.config.PageTitleSuffix,
		"Locale":          r.config.Locale,
		"Content":         template.HTML(htmlContent),
		"Explorer":        r.explorerCache,
		"GraphView":       "",
		"TableOfContent":  "",
		"Backlinks":       "",
	}

	for k, v := range metadata {
		data[k] = v
	}

	var buf bytes.Buffer
	if err := r.template.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
