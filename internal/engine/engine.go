package engine

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"blaze/internal/config"
	"blaze/internal/markdown"
	"blaze/internal/utils"
)

type SSG struct {
	ContentDir  string
	TemplateDir string
	OutputDir   string
	ConfigPath  string
	config      *config.Config
	renderer    *markdown.Renderer
}

func NewSSG(contentDir, templateDir, outputDir, configPath string) (*SSG, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	renderer, err := markdown.NewRenderer(templateDir, cfg)
	if err != nil {
		return nil, err
	}

	return &SSG{
		ContentDir:  contentDir,
		TemplateDir: templateDir,
		OutputDir:   outputDir,
		ConfigPath:  configPath,
		config:      cfg,
		renderer:    renderer,
	}, nil
}

func (s *SSG) shouldIgnore(relPath string) bool {
	for _, pattern := range s.config.IgnorePatterns {
		if strings.Contains(relPath, pattern) {
			return true
		}
		matched, err := filepath.Match(pattern, filepath.Base(relPath))
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (s *SSG) Build() error {
	fmt.Println("Building site...")

	if err := os.RemoveAll(s.OutputDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(s.OutputDir, 0755); err != nil {
		return err
	}

	if err := filepath.Walk(s.ContentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(s.ContentDir, path)

		if s.shouldIgnore(relPath) {
			return nil
		}

		if strings.HasSuffix(path, ".md") {
			return s.processMarkdown(path, relPath)
		}

		return s.copyStatic(path, relPath)
	}); err != nil {
		return err
	}

	return filepath.Walk(s.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".html") {
			return nil
		}

		relPath, _ := filepath.Rel(s.TemplateDir, path)
		return s.copyStatic(path, relPath)
	})
}

func (s *SSG) processMarkdown(sourcePath, relPath string) error {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	page, err := markdown.Parse(content)
	if err != nil {
		return err
	}

	html, err := s.renderer.Render(page)
	if err != nil {
		return err
	}

	dir := filepath.Dir(relPath)
	sluggedDir := utils.SlugifyPath(dir)
	slug := utils.PathToSlug(relPath)
	outputPath := filepath.Join(s.OutputDir, sluggedDir, slug+".html")
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

func (s *SSG) copyStatic(sourcePath, relPath string) error {
	dir := filepath.Dir(relPath)
	base := filepath.Base(relPath)
	sluggedDir := utils.SlugifyPath(dir)
	outputPath := filepath.Join(s.OutputDir, sluggedDir, base)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err == nil {
		fmt.Printf("Copied: %s\n", outputPath)
	}
	return err
}
