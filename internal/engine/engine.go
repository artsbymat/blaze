package engine

import (
	"fmt"
	"os"

	"blaze/internal/config"
	"blaze/internal/markdown"
	"blaze/internal/pipeline"
	"blaze/internal/renderer"
)

type SSG struct {
	ContentDir  string
	TemplateDir string
	OutputDir   string
	ConfigPath  string
	config      *config.Config
	pipeline    *pipeline.Pipeline
}

func NewSSG(contentDir, templateDir, outputDir, configPath string) (*SSG, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	htmlRenderer, err := renderer.NewHTMLRenderer(templateDir, cfg)
	if err != nil {
		return nil, err
	}

	p := pipeline.NewPipeline(cfg, htmlRenderer)
	p.RegisterTransformer(".md", markdown.NewTransformer())

	return &SSG{
		ContentDir:  contentDir,
		TemplateDir: templateDir,
		OutputDir:   outputDir,
		ConfigPath:  configPath,
		config:      cfg,
		pipeline:    p,
	}, nil
}

func (s *SSG) Build() error {
	fmt.Println("Building site...")

	if err := os.RemoveAll(s.OutputDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(s.OutputDir, 0755); err != nil {
		return err
	}

	if err := s.pipeline.Process(s.ContentDir, s.OutputDir); err != nil {
		return err
	}

	return s.pipeline.ProcessTemplates(s.TemplateDir, s.OutputDir)
}
