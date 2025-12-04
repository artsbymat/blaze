package markdown

import (
	"fmt"
	"strings"

	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type Page struct {
	Title       string
	RawContent  []byte
	HTMLContent string
	Metadata    map[string]string
}

func Parse(content []byte) (*Page, error) {
	converter := NewConverter()
	ctx := parser.NewContext()

	htmlContent, err := converter.ConvertWithContext(content, ctx)
	if err != nil {
		return nil, err
	}

	metaData := meta.Get(ctx)
	metadata := convertMetadata(metaData)

	title := metadata["title"]
	if title == "" {
		title = "Untitled"
	}

	bodyContent := extractBody(string(content))

	return &Page{
		Title:       title,
		RawContent:  []byte(bodyContent),
		HTMLContent: htmlContent,
		Metadata:    metadata,
	}, nil
}

func ExtractFrontmatter(textStr string) (map[string]string, string) {
	content := []byte(textStr)
	md := newGoldmark()

	ctx := parser.NewContext()
	p := md.Parser()
	reader := text.NewReader(content)
	_ = p.Parse(reader, parser.WithContext(ctx))

	metaData := meta.Get(ctx)
	metadata := convertMetadata(metaData)

	body := extractBody(textStr)

	return metadata, body
}

func extractBody(text string) string {
	if !strings.HasPrefix(text, "---\n") {
		return text
	}

	parts := strings.SplitN(text[4:], "\n---\n", 2)
	if len(parts) == 2 {
		return parts[1]
	}

	return text
}

func convertMetadata(metaData map[string]interface{}) map[string]string {
	metadata := make(map[string]string)
	for key, value := range metaData {
		metadata[key] = fmt.Sprint(value)
	}
	return metadata
}

type MarkdownTransformer struct{}

func NewTransformer() *MarkdownTransformer {
	return &MarkdownTransformer{}
}

func (t *MarkdownTransformer) Name() string {
	return "markdown"
}

func (t *MarkdownTransformer) Transform(content []byte) (string, map[string]string, error) {
	page, err := Parse(content)
	if err != nil {
		return "", nil, err
	}
	return page.HTMLContent, page.Metadata, nil
}
