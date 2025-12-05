package extensions

import (
	"blaze/internal/utils"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

type WikilinkNode struct {
	gast.BaseInline
	Target   []byte
	Fragment []byte
	Embed    bool
}

var KindWikilink = gast.NewNodeKind("Wikilink")

func (n *WikilinkNode) Kind() gast.NodeKind {
	return KindWikilink
}

func (n *WikilinkNode) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// -----------------------------------------------------------------------------
// Parser
// -----------------------------------------------------------------------------

type wikilinkParser struct{}

var (
	_open      = []byte("[[")
	_embedOpen = []byte("![[")
	_pipe      = []byte{'|'}
	_hash      = []byte{'#'}
	_close     = []byte("]]")
)

func (p *wikilinkParser) Trigger() []byte {
	return []byte{'!', '['}
}

func (p *wikilinkParser) Parse(_ gast.Node, block text.Reader, _ parser.Context) gast.Node {
	line, seg := block.PeekLine()
	stop := bytes.Index(line, _close)
	if stop < 0 {
		return nil
	}

	var embed bool

	switch {
	case bytes.HasPrefix(line, _open):
		seg = text.NewSegment(seg.Start+len(_open), seg.Start+stop)
	case bytes.HasPrefix(line, _embedOpen):
		embed = true
		seg = text.NewSegment(seg.Start+len(_embedOpen), seg.Start+stop)
	default:
		return nil
	}

	n := &WikilinkNode{Target: block.Value(seg), Embed: embed}
	if idx := bytes.Index(n.Target, _pipe); idx >= 0 {
		n.Target = n.Target[:idx]
		seg = seg.WithStart(seg.Start + idx + 1)
	}

	if len(n.Target) == 0 || seg.Len() == 0 {
		return nil
	}

	if idx := bytes.LastIndex(n.Target, _hash); idx >= 0 {
		n.Fragment = n.Target[idx+1:]
		n.Target = n.Target[:idx]
	}

	n.AppendChild(n, gast.NewTextSegment(seg))
	block.Advance(stop + 2)
	return n
}

// -----------------------------------------------------------------------------
// Resolver
// -----------------------------------------------------------------------------

type WikilinkResolver interface {
	ResolveWikilink(*WikilinkNode) (destination []byte, err error)
}

type slugResolver struct {
	contentDir string
	index      map[string]string
	mediaIndex map[string]string
}

func NewSlugResolver(contentDir string) WikilinkResolver {
	r := &slugResolver{
		contentDir: contentDir,
		index:      make(map[string]string),
		mediaIndex: make(map[string]string),
	}
	r.buildIndex()
	return r
}

func (r *slugResolver) buildIndex() {
	filepath.Walk(r.contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(r.contentDir, path)
		ext := filepath.Ext(path)

		if isImage(relPath) {
			dir := filepath.Dir(relPath)
			base := filepath.Base(path)
			nameWithoutExt := strings.TrimSuffix(base, ext)

			slug := utils.PathToSlug(nameWithoutExt)
			normalizedExt := strings.ToLower(ext)
			normalizedBase := slug + normalizedExt

			sluggedDir := utils.SlugifyPath(dir)

			var mediaPath string
			if sluggedDir == "" || sluggedDir == "." {
				mediaPath = "/" + normalizedBase
			} else {
				mediaPath = "/" + sluggedDir + "/" + normalizedBase
			}

			key := strings.ToLower(base)
			r.mediaIndex[key] = mediaPath

			nameKey := strings.ToLower(nameWithoutExt)
			r.mediaIndex[nameKey] = mediaPath

			normalizedKey := strings.ToLower(normalizedBase)
			r.mediaIndex[normalizedKey] = mediaPath

			fullPathKey := strings.ToLower(filepath.ToSlash(relPath))
			r.mediaIndex[fullPathKey] = mediaPath

			return nil
		}

		if ext != ".md" {
			return nil
		}

		dir := filepath.Dir(relPath)
		sluggedDir := utils.SlugifyPath(dir)
		slug := utils.PathToSlug(relPath)

		base := filepath.Base(path)
		nameWithoutExt := strings.TrimSuffix(base, ext)
		key := strings.ToLower(nameWithoutExt)

		var urlPath string
		if key == "index" {
			if sluggedDir == "" || sluggedDir == "." {
				urlPath = "/"
			} else {
				urlPath = "/" + sluggedDir
			}
		} else {
			// Untuk file biasa
			if sluggedDir == "" || sluggedDir == "." {
				urlPath = "/" + slug
			} else {
				urlPath = "/" + sluggedDir + "/" + slug
			}
		}

		r.index[key] = urlPath

		fullPathKey := strings.ToLower(strings.TrimSuffix(relPath, ext))
		r.index[fullPathKey] = urlPath

		if key == "index" && dir != "." {
			folderKey := strings.ToLower(filepath.Base(dir))
			r.index[folderKey] = urlPath
			parentDirKey := strings.ToLower(dir)
			r.index[parentDirKey] = urlPath
		}

		return nil
	})
}

func (r *slugResolver) ResolveWikilink(n *WikilinkNode) ([]byte, error) {
	target := string(n.Target)

	if isImage(target) {
		key := strings.ToLower(target)
		if mediaPath, found := r.mediaIndex[key]; found {
			return []byte(mediaPath), nil
		}

		baseName := strings.ToLower(filepath.Base(target))
		if mediaPath, found := r.mediaIndex[baseName]; found {
			return []byte(mediaPath), nil
		}

		if !strings.HasPrefix(target, "/") {
			return []byte("/" + target), nil
		}
		return []byte(target), nil
	}

	key := strings.ToLower(target)
	urlPath, found := r.index[key]

	if !found {
		slug := utils.PathToSlug(target)
		if slug == "" {
			return nil, nil
		}
		urlPath = "/" + slug
	}

	var dest bytes.Buffer
	dest.WriteString(urlPath)

	if len(n.Fragment) > 0 {
		dest.WriteString("#")
		fragment := string(n.Fragment)
		normalizedFragment := utils.Slugify(fragment)
		dest.WriteString(normalizedFragment)
	}

	return dest.Bytes(), nil
}

func isImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".apng", ".avif", ".gif", ".jpg", ".jpeg", ".jfif", ".pjpeg", ".pjp", ".png", ".svg", ".webp":
		return true
	}
	return false
}

// -----------------------------------------------------------------------------
// Renderer
// -----------------------------------------------------------------------------

type WikilinkRenderer struct {
	Resolver WikilinkResolver
	once     sync.Once
	hasDest  sync.Map
}

func NewWikilinkRenderer(resolver WikilinkResolver) renderer.NodeRenderer {
	return &WikilinkRenderer{
		Resolver: resolver,
	}
}

func (r *WikilinkRenderer) init() {
	r.once.Do(func() {
		if r.Resolver == nil {
			r.Resolver = NewSlugResolver("content")
		}
	})
}

func (r *WikilinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindWikilink, r.Render)
}

func (r *WikilinkRenderer) Render(w util.BufWriter, src []byte, node gast.Node, entering bool) (gast.WalkStatus, error) {
	r.init()

	n, ok := node.(*WikilinkNode)
	if !ok {
		return gast.WalkStop, fmt.Errorf("unexpected node %T, expected *WikilinkNode", node)
	}

	if entering {
		return r.enter(w, n, src)
	}

	r.exit(w, n)
	return gast.WalkContinue, nil
}

func (r *WikilinkRenderer) enter(w util.BufWriter, n *WikilinkNode, src []byte) (gast.WalkStatus, error) {
	dest, err := r.Resolver.ResolveWikilink(n)
	if err != nil {
		return gast.WalkStop, fmt.Errorf("resolve %q: %w", n.Target, err)
	}
	if len(dest) == 0 {
		return gast.WalkContinue, nil
	}

	img := resolveAsImage(n)
	if !img {
		r.hasDest.Store(n, struct{}{})
		_, _ = w.WriteString(`<a href="`)
		_, _ = w.Write(util.URLEscape(dest, true))
		_, _ = w.WriteString(`" class="internal">`)
		return gast.WalkContinue, nil
	}

	_, _ = w.WriteString(`<img src="`)
	_, _ = w.Write(util.URLEscape(dest, true))

	var width, height []byte

	if n.ChildCount() == 1 {
		label := nodeText(src, n.FirstChild())

		labelText := string(label)
		if isNumeric(labelText) {
			width = []byte(labelText)
		} else if parts := strings.Split(labelText, "x"); len(parts) == 2 && isNumeric(parts[0]) && isNumeric(parts[1]) {
			width = []byte(parts[0])
			height = []byte(parts[1])
		} else {
			if !bytes.Equal(label, n.Target) {
				_, _ = w.WriteString(`" alt="`)
				_, _ = w.Write(util.EscapeHTML(label))
			}
		}
	}

	if len(width) > 0 {
		_, _ = w.WriteString(`" width="`)
		_, _ = w.Write(width)
	}
	if len(height) > 0 {
		_, _ = w.WriteString(`" height="`)
		_, _ = w.Write(height)
	}

	_, _ = w.WriteString(`">`)
	return gast.WalkSkipChildren, nil
}

func (r *WikilinkRenderer) exit(w util.BufWriter, n *WikilinkNode) {
	if _, ok := r.hasDest.LoadAndDelete(n); ok {
		_, _ = w.WriteString("</a>")
	}
}

func resolveAsImage(n *WikilinkNode) bool {
	if !n.Embed {
		return false
	}

	filename := string(n.Target)
	switch ext := filepath.Ext(filename); ext {
	case ".apng", ".avif", ".gif", ".jpg", ".jpeg", ".jfif", ".pjpeg", ".pjp", ".png", ".svg", ".webp":
		return true
	default:
		return false
	}
}

func nodeText(src []byte, n gast.Node) []byte {
	var buf bytes.Buffer
	writeNodeText(src, &buf, n)
	return buf.Bytes()
}

func writeNodeText(src []byte, dst io.Writer, n gast.Node) {
	switch n := n.(type) {
	case *gast.Text:
		_, _ = dst.Write(n.Segment.Value(src))
	case *gast.String:
		_, _ = dst.Write(n.Value)
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			writeNodeText(src, dst, c)
		}
	}
}

// -----------------------------------------------------------------------------
// Extension
// -----------------------------------------------------------------------------

type wikilink struct {
	resolver WikilinkResolver
}

func Wikilink(resolver WikilinkResolver) goldmark.Extender {
	return &wikilink{resolver: resolver}
}

func (e *wikilink) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&wikilinkParser{}, 199),
	))
	m.Parser().AddOptions(parser.WithASTTransformers(
		util.Prioritized(&LinkTransformer{}, 100),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewWikilinkRenderer(e.resolver), 199),
	))
}

// -----------------------------------------------------------------------------
// Transformer
// -----------------------------------------------------------------------------

type LinkTransformer struct{}

func (t *LinkTransformer) Transform(node *gast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *gast.Link:
			processLink(n)
		case *gast.Image:
			processImage(n, source)
		}
		return gast.WalkContinue, nil
	})
}

func processLink(n *gast.Link) {
	dest := string(n.Destination)
	if isExternal(dest) {
		n.SetAttribute([]byte("class"), []byte("external"))
	} else {
		n.SetAttribute([]byte("class"), []byte("internal"))
	}
}

func processImage(n *gast.Image, source []byte) {
	if len(n.Title) > 0 {
		title := string(n.Title)
		if idx := strings.LastIndex(title, "|"); idx >= 0 {
			size := title[idx+1:]
			if parseImageSize(n, size) {
				n.Title = n.Title[:idx]
				return
			}
		}
	}

	if n.ChildCount() > 0 {
		child := n.FirstChild()
		if textNode, ok := child.(*gast.Text); ok {
			content := string(textNode.Segment.Value(source))
			if idx := strings.LastIndex(content, "|"); idx >= 0 {
				size := content[idx+1:]
				if parseImageSize(n, size) {
					newContent := content[:idx]
					textNode.Segment = text.NewSegment(textNode.Segment.Start, textNode.Segment.Start+len(newContent))
				}
			}
		}
	}
}

func parseImageSize(n *gast.Image, size string) bool {
	parts := strings.Split(size, "x")
	if len(parts) == 1 {
		if isNumeric(parts[0]) {
			n.SetAttribute([]byte("width"), []byte(parts[0]))
			return true
		}
	} else if len(parts) == 2 {
		if isNumeric(parts[0]) && isNumeric(parts[1]) {
			n.SetAttribute([]byte("width"), []byte(parts[0]))
			n.SetAttribute([]byte("height"), []byte(parts[1]))
			return true
		}
	}
	return false
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func isExternal(url string) bool {
	return strings.HasPrefix(url, "http") || strings.HasPrefix(url, "//")
}
