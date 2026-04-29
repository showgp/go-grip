package internal

import (
	"bytes"
	"strings"

	"github.com/chrishrb/go-grip/pkg/alert"
	"github.com/chrishrb/go-grip/pkg/details"
	"github.com/chrishrb/go-grip/pkg/footnote"
	"github.com/chrishrb/go-grip/pkg/ghissue"
	"github.com/chrishrb/go-grip/pkg/highlighting"
	"github.com/chrishrb/go-grip/pkg/mathjax"
	"github.com/chrishrb/go-grip/pkg/tasklist"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-emoji"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/hashtag"
	"go.abhg.dev/goldmark/mermaid"
)

type Parser struct{}

type RenderedDocument struct {
	Content string
	TOC     []TOCEntry
}

type TOCEntry struct {
	Level int
	Text  string
	ID    string
}

func NewParser() *Parser {
	return &Parser{}
}

func (m Parser) MdToHTML(input []byte) ([]byte, error) {
	rendered, err := m.Render(input)
	if err != nil {
		return nil, err
	}
	return []byte(rendered.Content), nil
}

func (m Parser) Render(input []byte) (*RenderedDocument, error) {
	md := newMarkdown()
	reader := text.NewReader(input)
	doc := md.Parser().Parse(reader)
	toc := collectTOC(doc, input)

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, input, doc); err != nil {
		return nil, err
	}
	return &RenderedDocument{
		Content: buf.String(),
		TOC:     toc,
	}, nil
}

func newMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.Linkify,
			extension.Table,
			extension.Strikethrough,
			footnote.Footnote,
			tasklist.TaskList,
			emoji.Emoji,
			&hashtag.Extender{},
			alert.New(),
			highlighting.Highlighting,
			&mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true},
			mathjax.MathJax,
			ghissue.New(),
			details.New(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

func collectTOC(node gast.Node, source []byte) []TOCEntry {
	entries := make([]TOCEntry, 0)
	_ = gast.Walk(node, func(n gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		heading, ok := n.(*gast.Heading)
		if !ok {
			return gast.WalkContinue, nil
		}

		id := attributeString(heading, "id")
		text := strings.TrimSpace(string(heading.Text(source)))
		if id == "" || text == "" {
			return gast.WalkContinue, nil
		}

		entries = append(entries, TOCEntry{
			Level: heading.Level,
			Text:  text,
			ID:    id,
		})
		return gast.WalkSkipChildren, nil
	})
	return entries
}

func attributeString(node gast.Node, name string) string {
	value, ok := node.AttributeString(name)
	if !ok {
		return ""
	}

	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	default:
		return ""
	}
}
