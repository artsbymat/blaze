package extensions

import (
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

type Highlight struct {
	gast.BaseInline
}

var KindHighlight = gast.NewNodeKind("Highlight")

func (n *Highlight) Kind() gast.NodeKind {
	return KindHighlight
}

func NewHighlight() *Highlight {
	return &Highlight{}
}

// Required to implement ast.Node
func (n *Highlight) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// -----------------------------------------------------------------------------
// Delimiter Processor (==)
// -----------------------------------------------------------------------------

type highlightDelimiterProcessor struct{}

func (p *highlightDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '='
}

func (p *highlightDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *highlightDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return NewHighlight()
}

var defaultHighlightDelimiterProcessor = &highlightDelimiterProcessor{}

// -----------------------------------------------------------------------------
// Inline Parser
// -----------------------------------------------------------------------------

type highlightParser struct{}

var defaultHighlightParser = &highlightParser{}

func NewHighlightParser() parser.InlineParser {
	return defaultHighlightParser
}

func (s *highlightParser) Trigger() []byte {
	return []byte{'='}
}

func (s *highlightParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()

	node := parser.ScanDelimiter(line, before, 1, defaultHighlightDelimiterProcessor)
	if node == nil || node.OriginalLength > 2 || before == '=' {
		return nil
	}

	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (s *highlightParser) CloseBlock(parent gast.Node, pc parser.Context) {}

// -----------------------------------------------------------------------------
// HTML Renderer
// -----------------------------------------------------------------------------

type HighlightHTMLRenderer struct {
	html.Config
}

func NewHighlightHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &HighlightHTMLRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

var HighlightAttributeFilter = html.GlobalAttributeFilter

func (r *HighlightHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindHighlight, r.renderHighlight)
}

func (r *HighlightHTMLRenderer) renderHighlight(
	w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {

	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<mark")
			html.RenderAttributes(w, n, HighlightAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<mark>")
		}
	} else {
		_, _ = w.WriteString("</mark>")
	}
	return gast.WalkContinue, nil
}

// -----------------------------------------------------------------------------
// Extension
// -----------------------------------------------------------------------------

type highlight struct{}

var ObsidianHighlight = &highlight{}

func (e *highlight) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewHighlightParser(), 501),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewHighlightHTMLRenderer(), 501),
	))
}
