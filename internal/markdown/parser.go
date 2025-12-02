package markdown

import (
	"bytes"
	"regexp"
	"strings"
)

type Page struct {
	Title    string
	Content  string
	Metadata map[string]string
}

var (
	h1Regex     = regexp.MustCompile(`(?m)^# (.+)$`)
	h2Regex     = regexp.MustCompile(`(?m)^## (.+)$`)
	h3Regex     = regexp.MustCompile(`(?m)^### (.+)$`)
	boldRegex   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex = regexp.MustCompile(`\*(.+?)\*`)
	codeRegex   = regexp.MustCompile("`([^`]+)`")
	linkRegex   = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

func Parse(content []byte) (*Page, error) {
	text := string(content)

	metadata, body := extractFrontmatter(text)

	html := convertToHTML(body)

	title := metadata["title"]
	if title == "" {
		title = "Untitled"
	}

	return &Page{
		Title:    title,
		Content:  html,
		Metadata: metadata,
	}, nil
}

func extractFrontmatter(text string) (map[string]string, string) {
	metadata := make(map[string]string)

	if !strings.HasPrefix(text, "---\n") {
		return metadata, text
	}

	parts := strings.SplitN(text[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return metadata, text
	}

	lines := strings.Split(parts[0], "\n")
	for _, line := range lines {
		if idx := strings.Index(line, ":"); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			metadata[key] = value
		}
	}

	return metadata, parts[1]
}

func convertToHTML(markdown string) string {
	var buf bytes.Buffer

	lines := strings.Split(markdown, "\n")
	inParagraph := false
	inCodeBlock := false

	for _, line := range lines {
		line = strings.TrimRight(line, " ")

		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				buf.WriteString("</code></pre>\n")
				inCodeBlock = false
			} else {
				if inParagraph {
					buf.WriteString("</p>\n")
					inParagraph = false
				}
				buf.WriteString("<pre><code>")
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			buf.WriteString(line + "\n")
			continue
		}

		if strings.HasPrefix(line, "# ") {
			if inParagraph {
				buf.WriteString("</p>\n")
				inParagraph = false
			}
			buf.WriteString("<h1>" + processInline(line[2:]) + "</h1>\n")
		} else if strings.HasPrefix(line, "## ") {
			if inParagraph {
				buf.WriteString("</p>\n")
				inParagraph = false
			}
			buf.WriteString("<h2>" + processInline(line[3:]) + "</h2>\n")
		} else if strings.HasPrefix(line, "### ") {
			if inParagraph {
				buf.WriteString("</p>\n")
				inParagraph = false
			}
			buf.WriteString("<h3>" + processInline(line[4:]) + "</h3>\n")
		} else if line == "" {
			if inParagraph {
				buf.WriteString("</p>\n")
				inParagraph = false
			}
		} else {
			if !inParagraph {
				buf.WriteString("<p>")
				inParagraph = true
			} else {
				buf.WriteString(" ")
			}
			buf.WriteString(processInline(line))
		}
	}

	if inParagraph {
		buf.WriteString("</p>\n")
	}

	return buf.String()
}

func processInline(text string) string {
	text = linkRegex.ReplaceAllString(text, `<a href="$2">$1</a>`)
	text = boldRegex.ReplaceAllString(text, `<strong>$1</strong>`)
	text = italicRegex.ReplaceAllString(text, `<em>$1</em>`)
	text = codeRegex.ReplaceAllString(text, `<code>$1</code>`)
	return text
}
