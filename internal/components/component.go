package components

import (
	"html/template"

	"blaze/internal/config"
)

type Component interface {
	Generate() (template.HTML, error)
}

type ComponentFactory struct {
	config *config.Config
}

func NewComponentFactory(cfg *config.Config) *ComponentFactory {
	return &ComponentFactory{config: cfg}
}

func (f *ComponentFactory) CreateExplorer(root string) Component {
	return NewExplorer(f.config, root)
}
