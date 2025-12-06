package extensions

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// -----------------------------------------------------------------------------
// Node Definition
// -----------------------------------------------------------------------------

type YoutubeNode struct {
	gast.BaseBlock
	VideoID string
}

var KindYoutube = gast.NewNodeKind("Youtube")

func (n *YoutubeNode) Kind() gast.NodeKind {
	return KindYoutube
}

func (n *YoutubeNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// -----------------------------------------------------------------------------
// Transformer
// -----------------------------------------------------------------------------

type YoutubeTransformer struct{}

var defaultYoutubeTransformer = &YoutubeTransformer{}

func (t *YoutubeTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		if n.Kind() != gast.KindImage {
			return gast.WalkContinue, nil
		}

		img := n.(*gast.Image)
		dest := string(img.Destination)
		videoID := extractVideoID(dest)

		if videoID != "" {
			yt := &YoutubeNode{
				VideoID: videoID,
			}

			parent := n.Parent()
			parent.ReplaceChild(parent, n, yt)
		}

		return gast.WalkContinue, nil
	})
}

func extractVideoID(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}

	if parsed.Host == "www.youtube.com" || parsed.Host == "youtube.com" {
		if parsed.Path == "/watch" {
			return parsed.Query().Get("v")
		}
		if strings.HasPrefix(parsed.Path, "/embed/") {
			return strings.TrimPrefix(parsed.Path, "/embed/")
		}
	} else if parsed.Host == "youtu.be" {
		return strings.TrimPrefix(parsed.Path, "/")
	}

	return ""
}

// Redefine YoutubeNode as Inline
type YoutubeInlineNode struct {
	gast.BaseInline
	VideoID string
}

var KindYoutubeInline = gast.NewNodeKind("YoutubeInline")

func (n *YoutubeInlineNode) Kind() gast.NodeKind {
	return KindYoutubeInline
}

func (n *YoutubeInlineNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// Update Transformer to use YoutubeInlineNode
func (t *YoutubeTransformer) TransformInline(node *gast.Document, reader text.Reader, pc parser.Context) {
	// ...
}

// Let's rewrite the Transform method to be cleaner and use the inline node
func (t *YoutubeTransformer) Transform2(node *gast.Document, reader text.Reader, pc parser.Context) {
	// ...
}

// -----------------------------------------------------------------------------
// Renderer
// -----------------------------------------------------------------------------

type YoutubeRenderer struct{}

func (r *YoutubeRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindYoutubeInline, r.Render)
}

func (r *YoutubeRenderer) Render(w util.BufWriter, source []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		n := node.(*YoutubeInlineNode)
		w.WriteString(fmt.Sprintf(`<iframe src="https://www.youtube.com/embed/%s" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`, n.VideoID))
	}
	return gast.WalkContinue, nil
}

// -----------------------------------------------------------------------------
// Extension
// -----------------------------------------------------------------------------

type youtube struct{}

var Youtube = &youtube{}

func (e *youtube) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&YoutubeASTTransformer{}, 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&YoutubeRenderer{}, 500),
	))
}

type YoutubeASTTransformer struct{}

func (t *YoutubeASTTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		if n.Kind() == gast.KindImage {
			img := n.(*gast.Image)
			dest := string(img.Destination)
			videoID := extractVideoID(dest)

			if videoID != "" {
				yt := &YoutubeInlineNode{
					VideoID: videoID,
				}
				n.Parent().ReplaceChild(n.Parent(), n, yt)
			}
		}
		return gast.WalkContinue, nil
	})
}
