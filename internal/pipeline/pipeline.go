package pipeline

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"blaze/internal/config"
	"blaze/internal/renderer"
	"blaze/internal/utils"

	"golang.org/x/sync/errgroup"
)

type Transformer interface {
	Name() string
	Transform(content []byte) (string, map[string]string, error)
}

type Pipeline struct {
	config       *config.Config
	renderer     *renderer.HTMLRenderer
	transformers map[string]Transformer
}

func NewPipeline(cfg *config.Config, renderer *renderer.HTMLRenderer) *Pipeline {
	return &Pipeline{
		config:       cfg,
		renderer:     renderer,
		transformers: make(map[string]Transformer),
	}
}

func (p *Pipeline) RegisterTransformer(ext string, transformer Transformer) {
	p.transformers[ext] = transformer
}

func (p *Pipeline) Process(contentDir, outputDir string) error {
	if err := p.renderer.RegenerateExplorer(contentDir); err != nil {
		return fmt.Errorf("failed to generate explorer: %w", err)
	}

	var g errgroup.Group

	sem := make(chan struct{}, 20)

	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(contentDir, path)
		if p.shouldIgnore(relPath) {
			return nil
		}

		sem <- struct{}{}

		g.Go(func() error {
			defer func() { <-sem }()
			return p.processFile(path, relPath, outputDir)
		})

		return nil
	})

	if err != nil {
		return err
	}

	return g.Wait()
}

func (p *Pipeline) ProcessTemplates(templateDir, outputDir string) error {
	return filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".html") {
			return nil
		}

		relPath, _ := filepath.Rel(templateDir, path)
		return p.copyStatic(path, relPath, outputDir)
	})
}

func (p *Pipeline) processFile(sourcePath, relPath, outputDir string) error {
	ext := filepath.Ext(sourcePath)
	transformer, ok := p.transformers[ext]

	if !ok {
		return p.copyStatic(sourcePath, relPath, outputDir)
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	htmlContent, metadata, err := transformer.Transform(content)
	if err != nil {
		return fmt.Errorf("failed to transform %s: %w", sourcePath, err)
	}

	if p.config.PublishMode == "explicit" {
		if val, ok := metadata["publish"]; !ok || val != "true" {
			return nil
		}
	}

	finalHTML, err := p.renderer.RenderPage(htmlContent, metadata)
	if err != nil {
		return err
	}

	dir := filepath.Dir(relPath)
	sluggedDir := utils.SlugifyPath(dir)
	slug := utils.PathToSlug(relPath)
	outputPath := filepath.Join(outputDir, sluggedDir, slug+".html")

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, []byte(finalHTML), 0644); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", outputPath)
	return nil
}

func (p *Pipeline) copyStatic(sourcePath, relPath, outputDir string) error {
	dir := filepath.Dir(relPath)
	base := filepath.Base(relPath)
	sluggedDir := utils.SlugifyPath(dir)
	outputPath := filepath.Join(outputDir, sluggedDir, base)

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

func (p *Pipeline) shouldIgnore(relPath string) bool {
	pathParts := strings.Split(filepath.ToSlash(relPath), "/")
	baseName := filepath.Base(relPath)

	for _, pattern := range p.config.IgnorePatterns {
		for _, part := range pathParts {
			matched, err := filepath.Match(pattern, part)
			if err == nil && matched {
				return true
			}
			if part == pattern {
				return true
			}
		}

		matched, err := filepath.Match(pattern, baseName)
		if err == nil && matched {
			return true
		}
	}
	return false
}
