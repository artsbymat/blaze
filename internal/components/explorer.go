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

func (e *Explorer) isPublished(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	metadata, _ := markdown.ExtractFrontmatter(string(content))
	return metadata["publish"] == "true"
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
			html += fmt.Sprintf("<li><details open><summary>%s</summary>%s</details></li>", entry.Name(), subHtml)
		} else {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}

			if e.config.PublishMode == "explicit" && !e.isPublished(path) {
				continue
			}

			relPath := strings.TrimPrefix(path, e.root+"/")
			relPath = strings.TrimSuffix(relPath, ".md")
			slugPath := utils.SlugifyPath(relPath) + ".html"
			title := strings.TrimSuffix(entry.Name(), ".md")
			html += fmt.Sprintf("<li><a href=\"/%s\">%s</a></li>", slugPath, title)
		}
	}

	html += "</ul>"
	return template.HTML(html), nil
}
