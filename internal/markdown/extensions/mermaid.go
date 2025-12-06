package extensions

import (
	"bytes"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// -----------------------------------------------------------------------------
// Node Definition
// -----------------------------------------------------------------------------

type MermaidBlock struct {
	gast.BaseBlock
}

var KindMermaidBlock = gast.NewNodeKind("MermaidBlock")

func (n *MermaidBlock) Kind() gast.NodeKind {
	return KindMermaidBlock
}

func (n *MermaidBlock) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

func NewMermaidBlock() *MermaidBlock {
	return &MermaidBlock{}
}

// -----------------------------------------------------------------------------
// AST Transformer
// -----------------------------------------------------------------------------

type mermaidTransformer struct{}

var defaultMermaidTransformer = &mermaidTransformer{}

var MermaidContextKey = parser.NewContextKey()

func (t *mermaidTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	var nodesToReplace []*gast.FencedCodeBlock

	gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		if n.Kind() != gast.KindFencedCodeBlock {
			return gast.WalkContinue, nil
		}

		fcb := n.(*gast.FencedCodeBlock)
		lang := fcb.Language(reader.Source())
		if bytes.Equal(lang, []byte("mermaid")) {
			nodesToReplace = append(nodesToReplace, fcb)
		}

		return gast.WalkContinue, nil
	})

	if len(nodesToReplace) > 0 {
		pc.Set(MermaidContextKey, true)

		for _, fcb := range nodesToReplace {
			// Create our custom node
			mermaidNode := NewMermaidBlock()

			// Copy lines from the code block to our new node
			lines := fcb.Lines()
			for i := 0; i < lines.Len(); i++ {
				line := lines.At(i)
				mermaidNode.Lines().Append(line)
			}

			// Replace the code block with our mermaid node
			parent := fcb.Parent()
			parent.ReplaceChild(parent, fcb, mermaidNode)
		}
	}
}

// -----------------------------------------------------------------------------
// HTML Renderer
// -----------------------------------------------------------------------------

type mermaidRenderer struct {
	html.Config
}

func newMermaidRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &mermaidRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *mermaidRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMermaidBlock, r.renderMermaidBlock)
}

func (r *mermaidRenderer) renderMermaidBlock(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<div class=\"mermaid\">")
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			_, _ = w.Write(line.Value(source))
		}
	} else {
		_, _ = w.WriteString("</div>")
	}
	return gast.WalkContinue, nil
}

// -----------------------------------------------------------------------------
// Extension
// -----------------------------------------------------------------------------

type mermaid struct{}

var Mermaid = &mermaid{}

func (e *mermaid) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(defaultMermaidTransformer, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newMermaidRenderer(), 500),
	))
}
