package components

import (
	"blaze/internal/config"
	"blaze/internal/utils"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
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
	for _, pattern := range e.config.IgnorePatterns {
		if strings.Contains(name, pattern) {
			return true
		}
		matched, err := filepath.Match(pattern, name)
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

	text := string(content)
	if !strings.HasPrefix(text, "---\n") {
		return false
	}

	parts := strings.SplitN(text[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return false
	}

	lines := strings.Split(parts[0], "\n")
	for _, line := range lines {
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			if key == "publish" && value == "true" {
				return true
			}
		}
	}

	return false
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
				return "", err
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
