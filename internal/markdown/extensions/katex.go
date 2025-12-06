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

type Math struct {
	gast.BaseInline
	IsDisplay bool
}

var KindMath = gast.NewNodeKind("Math")

func (n *Math) Kind() gast.NodeKind {
	return KindMath
}

func (n *Math) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

func NewMath(isDisplay bool) *Math {
	return &Math{
		IsDisplay: isDisplay,
	}
}

// -----------------------------------------------------------------------------
// Parser
// -----------------------------------------------------------------------------

type mathParser struct{}

var defaultMathParser = &mathParser{}

func NewMathParser() parser.InlineParser {
	return defaultMathParser
}

func (s *mathParser) Trigger() []byte {
	return []byte{'$'}
}

func (s *mathParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
	if len(line) == 0 {
		return nil
	}

	// Check for escaped dollar
	before := block.PrecendingCharacter()
	if before == '\\' {
		return nil
	}

	// Check for display math $$
	isDisplay := false
	delimiter := []byte{'$'}
	if len(line) > 1 && line[1] == '$' {
		isDisplay = true
		delimiter = []byte{'$', '$'}
	}

	// Find closing delimiter
	// If it's display math, we need to check if it's empty $$ $$
	if isDisplay && len(line) > 2 && line[2] == '$' && line[3] == '$' {
		// Empty display math
		node := NewMath(true)
		block.Advance(4)
		return node
	}

	l, pos := block.Position()

	block.Advance(len(delimiter))

	node := NewMath(isDisplay)

	for {
		line, segment := block.PeekLine()
		if line == nil {
			break
		}

		cursor := 0
		foundInLine := false
		matchedIdx := -1

		for {
			next := bytes.Index(line[cursor:], delimiter)
			if next == -1 {
				break
			}
			realIdx := cursor + next

			isEscaped := false
			// Count backslashes
			bsCount := 0
			for i := realIdx - 1; i >= 0; i-- {
				if line[i] == '\\' {
					bsCount++
				} else {
					break
				}
			}
			if bsCount%2 != 0 {
				isEscaped = true
			}

			if !isEscaped {
				foundInLine = true
				matchedIdx = realIdx
				break
			}
			cursor = realIdx + 1
		}

		if foundInLine {
			// Found it!
			pc.Set(KatexContextKey, true)
			seg := segment.WithStop(segment.Start + matchedIdx)
			node.AppendChild(node, gast.NewTextSegment(seg))
			block.Advance(matchedIdx + len(delimiter))
			return node
		}

		// Not found in this line
		node.AppendChild(node, gast.NewTextSegment(segment))
		block.Advance(len(line))
	}

	// EOF
	block.SetPosition(l, pos)
	return nil
}

// -----------------------------------------------------------------------------
// HTML Renderer
// -----------------------------------------------------------------------------

type mathRenderer struct {
	html.Config
}

func newMathRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &mathRenderer{
		Config: html.NewConfig(),
	}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *mathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMath, r.renderMath)
}

func (r *mathRenderer) renderMath(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	mathNode := n.(*Math)
	if entering {

		if mathNode.IsDisplay {
			_, _ = w.WriteString("$$")
		} else {
			_, _ = w.WriteString("$")
		}

		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*gast.Text).Segment
			value := segment.Value(source)
			_, _ = w.Write(value)
		}

		if mathNode.IsDisplay {
			_, _ = w.WriteString("$$")
		} else {
			_, _ = w.WriteString("$")
		}

		return gast.WalkSkipChildren, nil
	}
	return gast.WalkContinue, nil
}

// -----------------------------------------------------------------------------
// Extension
// -----------------------------------------------------------------------------

type katex struct{}

var Katex = &katex{}

var KatexContextKey = parser.NewContextKey()

func (e *katex) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewMathParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newMathRenderer(), 500),
	))
}
