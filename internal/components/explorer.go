package components

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	"blaze/internal/config"
	"blaze/internal/markdown"
	"blaze/internal/utils"
)

type Explorer struct {
	config *config.Config
	root   string
}

func NewExplorer(cfg *config.Config, root string) *Explorer {
	return &Explorer{
		config: cfg,
		root:   root,
	}
}

func (e *Explorer) shouldIgnore(name string) bool {
	baseName := filepath.Base(name)

	for _, pattern := range e.config.IgnorePatterns {
		if name == pattern || baseName == pattern {
			return true
		}
		matched, err := filepath.Match(pattern, baseName)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func (e *Explorer) getMetadata(filePath string) (map[string]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	metadata, _ := markdown.ExtractFrontmatter(string(content))
	return metadata, nil
}

func (e *Explorer) Generate() (template.HTML, error) {
	return e.generateTree(e.root)
}

func (e *Explorer) generateTree(dir string) (template.HTML, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var html string
	html += "<ul>"

	for _, entry := range entries {
		if e.shouldIgnore(entry.Name()) {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			subHtml, err := e.generateTree(path)
			if err != nil {
				log.Printf("Failed to generate tree for %s: %v\n", path, err)
				continue
			}
			if strings.TrimSpace(string(subHtml)) == "<ul></ul>" {
				continue
			}

			relPath := strings.TrimPrefix(path, e.root+"/")

			html += fmt.Sprintf(
				`<li><details data-key="%s"><summary>%s</summary>%s</details></li>`,
				template.HTMLEscapeString(relPath),
				entry.Name(),
				subHtml,
			)
		} else {
			metadata, err := e.getMetadata(path)
			if err != nil {
				continue
			}

			if e.config.PublishMode == "explicit" && metadata["publish"] != "true" {
				continue
			}

			relPath := strings.TrimPrefix(path, e.root+"/")
			relPath = strings.TrimSuffix(relPath, ".md")
			slugPath := utils.SlugifyPath(relPath)
			title := metadata["title"]
			if title == "" {
				title = strings.TrimSuffix(entry.Name(), ".md")
			}
			html += fmt.Sprintf(`<li><a href="/%s">%s</a></li>`, slugPath, title)
		}
	}

	html += "</ul>"
	return template.HTML(html), nil
}
