package extensions

import (
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// CalloutNode represents a callout block
type CalloutNode struct {
	ast.BaseBlock
	CalloutType string
	Foldable    bool
	DefaultFold bool
}

var KindCallout = ast.NewNodeKind("Callout")

func (n *CalloutNode) Kind() ast.NodeKind {
	return KindCallout
}

func (n *CalloutNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// CalloutTitleNode represents the title part of a callout
type CalloutTitleNode struct {
	ast.BaseBlock
}

var KindCalloutTitle = ast.NewNodeKind("CalloutTitle")

func (n *CalloutTitleNode) Kind() ast.NodeKind {
	return KindCalloutTitle
}

func (n *CalloutTitleNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// CalloutBodyNode represents the body part of a callout
type CalloutBodyNode struct {
	ast.BaseBlock
}

var KindCalloutBody = ast.NewNodeKind("CalloutBody")

func (n *CalloutBodyNode) Kind() ast.NodeKind {
	return KindCalloutBody
}

func (n *CalloutBodyNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

type calloutTransformer struct{}

// Regex updated to be more robust:
// ^\s* : Allow leading whitespace (which might remain after blockquote processing in some edge cases or if not fully stripped)
// \[\! : Match [!
// ([a-zA-Z0-9-]+) : Match Type (Group 1)
// \] : Match ]
// ([+-]?) : Match Fold (Group 2)
// \s* : Skip spaces
// (.*?) : Match Title non-greedy (Group 3)
// \s* : Trailing spaces
// $ : End of line
var calloutRegex = regexp.MustCompile(`^\s*\[!([a-zA-Z0-9-]+)\]([+-]?)\s*(.*?)\s*$`)

func (t *calloutTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()

	// Multiple passes to handle nested callouts
	for i := 0; i < 20; i++ {
		changed := false

		_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}

			blockquote, ok := n.(*ast.Blockquote)
			if !ok {
				return ast.WalkContinue, nil
			}

			firstChild := blockquote.FirstChild()
			if firstChild == nil {
				return ast.WalkContinue, nil
			}

			paragraph, ok := firstChild.(*ast.Paragraph)
			if !ok {
				return ast.WalkContinue, nil
			}

			lines := paragraph.Lines()
			if lines.Len() == 0 {
				return ast.WalkContinue, nil
			}

			firstLine := lines.At(0)
			line := firstLine.Value(source)

			matches := calloutRegex.FindSubmatchIndex(line)
			if matches == nil {
				return ast.WalkContinue, nil
			}

			// matches indices (groups):
			// 0,1: full match
			// 2,3: type
			// 4,5: fold
			// 6,7: title

			calloutType := string(line[matches[2]:matches[3]])
			foldSymbol := ""
			if matches[4] != -1 {
				// Ensure index valid check
				if matches[5] > matches[4] {
					foldSymbol = string(line[matches[4]:matches[5]])
				}
			}

			// Setup Callout Node
			callout := &CalloutNode{
				CalloutType: strings.ToLower(calloutType),
				Foldable:    foldSymbol != "",
				DefaultFold: foldSymbol == "-",
			}

			// Capture title segment
			calloutTitle := &CalloutTitleNode{}

			// Check if title group matched and is not empty
			titleText := ""
			if matches[6] != -1 && matches[7] > matches[6] {
				titleText = string(line[matches[6]:matches[7]])
			}

			if titleText == "" {
				// No title provided, use Type as default title
				titleText = strings.Title(strings.ToLower(calloutType))
			}

			// Create text node for title
			titleNode := ast.NewString([]byte(titleText))
			calloutTitle.AppendChild(calloutTitle, titleNode)
			callout.AppendChild(callout, calloutTitle)

			calloutBody := &CalloutBodyNode{}
			callout.AppendChild(callout, calloutBody)

			// Handle multi-line paragraph (content lines after callout header)
			if lines.Len() > 1 {
				newPara := ast.NewParagraph()
				for j := 1; j < lines.Len(); j++ {
					seg := lines.At(j)
					newPara.Lines().Append(seg)
					// Create text node for this segment so it renders
					textNode := ast.NewTextSegment(seg)
					newPara.AppendChild(newPara, textNode)
				}
				blockquote.ReplaceChild(blockquote, paragraph, newPara)
			} else {
				blockquote.RemoveChild(blockquote, paragraph)
			}

			// Move children to callout body
			for child := blockquote.FirstChild(); child != nil; {
				next := child.NextSibling()
				blockquote.RemoveChild(blockquote, child)
				calloutBody.AppendChild(calloutBody, child)
				child = next
			}

			// Replace blockquote with callout
			parent := blockquote.Parent()
			parent.ReplaceChild(parent, blockquote, callout)
			changed = true

			return ast.WalkContinue, nil
		})

		if !changed {
			break
		}
	}
}

type calloutRenderer struct {
	html.Config
}

func newCalloutRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &calloutRenderer{Config: html.NewConfig()}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *calloutRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindCallout, r.renderCallout)
	reg.Register(KindCalloutTitle, r.renderCalloutTitle)
	reg.Register(KindCalloutBody, r.renderCalloutBody)
}

func (r *calloutRenderer) renderCallout(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*CalloutNode)

	if entering {
		_, _ = w.WriteString(`<div class="callout" data-callout="`)
		_, _ = w.WriteString(n.CalloutType)
		_, _ = w.WriteString(`"`)
		if n.Foldable {
			_, _ = w.WriteString(` data-callout-fold="`)
			if n.DefaultFold {
				_, _ = w.WriteString(`collapsed`)
			} else {
				_, _ = w.WriteString(`expanded`)
			}
			_, _ = w.WriteString(`"`)
		}
		_, _ = w.WriteString(`>`)
	} else {
		_, _ = w.WriteString(`</div>`)
	}
	return ast.WalkContinue, nil
}

func (r *calloutRenderer) renderCalloutTitle(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	// We need access to parent properties
	parent, ok := node.Parent().(*CalloutNode)
	if !ok {
		return ast.WalkContinue, nil
	}

	if entering {
		_, _ = w.WriteString(`<div class="callout-title">`)
		_, _ = w.WriteString(`<div class="callout-icon"></div>`)
		_, _ = w.WriteString(`<div class="callout-title-inner">`)
		// Children (text) will be rendered after this
	} else {
		// Child text rendered.
		_, _ = w.WriteString(`</div>`) // close inner

		if parent.Foldable {
			_, _ = w.WriteString(`<div class="callout-fold"></div>`)
		}

		_, _ = w.WriteString(`</div>`) // close title
	}
	return ast.WalkContinue, nil
}

func (r *calloutRenderer) renderCalloutBody(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<div class="callout-content">`)
	} else {
		_, _ = w.WriteString(`</div>`)
	}
	return ast.WalkContinue, nil
}

type callout struct{}

var Callout = &callout{}

func (e *callout) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&calloutTransformer{}, 150),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(newCalloutRenderer(), 500),
		),
	)
}
