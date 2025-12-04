package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func newGoldmark() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.Footnote,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

type Converter struct {
	md goldmark.Markdown
}

func NewConverter() *Converter {
	return &Converter{md: newGoldmark()}
}

func (c *Converter) ConvertWithContext(source []byte, ctx parser.Context) (string, error) {
	var buf bytes.Buffer
	if err := c.md.Convert(source, &buf, parser.WithContext(ctx)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
