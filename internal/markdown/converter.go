package markdown

import (
	"blaze/internal/markdown/extensions"
	"blaze/internal/markdown/highlighting"
	"bytes"

	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
)

func newGoldmark() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
			extension.Footnote,
			extensions.ObsidianHighlight,
			extensions.Mermaid,
			extensions.Katex,
			extensions.Wikilink(extensions.NewSlugResolver("content")),
			extensions.Youtube,
			extensions.HeadingShift,
			extensions.Anchor,
			extensions.Callout,
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(
					chromahtml.WithClasses(true),
				),
				highlighting.WithGuessLanguage(true),
			),
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
