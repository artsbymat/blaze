package extensions

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// AnchorNode represents an anchor link in a heading
type AnchorNode struct {
	ast.BaseInline
	ID    []byte
	Level int
	Value []byte
}

var KindAnchor = ast.NewNodeKind("Anchor")

func (n *AnchorNode) Kind() ast.NodeKind {
	return KindAnchor
}

func (n *AnchorNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// AnchorTransformer adds anchor links to headings
type anchorTransformer struct{}

func (t *anchorTransformer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		idattr, ok := h.AttributeString("id")
		if !ok {
			return ast.WalkSkipChildren, nil
		}

		id, ok := idattr.([]byte)
		if !ok {
			return ast.WalkSkipChildren, nil
		}

		anchorNode := &AnchorNode{
			ID:    id,
			Level: h.Level,
			Value: []byte("#"),
		}
		anchorNode.SetAttributeString("class", []byte("anchor"))

		if h.ChildCount() == 0 {
			h.AppendChild(h, anchorNode)
		} else {
			h.InsertAfter(h, h.LastChild(), anchorNode)
		}

		return ast.WalkSkipChildren, nil
	})
}

// AnchorRenderer renders anchor nodes as HTML links
type anchorRenderer struct {
	html.Config
}

func newAnchorRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &anchorRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *anchorRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAnchor, r.renderAnchor)
}

func (r *anchorRenderer) renderAnchor(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	anchorNode := n.(*AnchorNode)

	_, _ = w.WriteString(" <a")
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.LinkAttributeFilter)
	}
	_, _ = w.WriteString(` href="#`)
	_, _ = w.Write(anchorNode.ID)
	_, _ = w.WriteString(`">`)
	_, _ = w.Write(anchorNode.Value)
	_, _ = w.WriteString("</a>")

	return ast.WalkContinue, nil
}

type anchor struct{}

var Anchor = &anchor{}

func (e *anchor) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&anchorTransformer{}, 200),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newAnchorRenderer(), 500),
		),
	)
}
