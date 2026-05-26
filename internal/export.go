package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/showgp/go-grip/defaults"
)

// ExportData carries the data used by the export template (templates/export.html).
type ExportData struct {
	Content      template.HTML
	CssLight     template.CSS // github-markdown-light.css
	CssCodeLight template.CSS // Chroma light syntax highlighting
	CssMermaid   template.CSS // github-mermaid.css
	CssMathJax   template.CSS // mathjax.css
	CssClipboard template.CSS // github-clipboard.css
	IncludeJS    bool         // include <script> tags for Mermaid/MathJax
}

// BuildExportHTML renders a standalone HTML page from rendered Markdown content.
// When includeJS is true, <script> tags for Mermaid and MathJax are included
// (for use when the file will be opened while the server is running).
// When includeJS is false, no JS is emitted (for CLI export or fully static output).
func BuildExportHTML(content template.HTML, includeJS bool) (string, error) {
	cssLight, err := readCSS("static/css/github-markdown-light.css")
	if err != nil {
		return "", fmt.Errorf("read light css: %w", err)
	}

	cssMermaid, err := readCSS("static/css/github-mermaid.css")
	if err != nil {
		return "", fmt.Errorf("read mermaid css: %w", err)
	}

	cssMathJax, err := readCSS("static/css/mathjax.css")
	if err != nil {
		return "", fmt.Errorf("read mathjax css: %w", err)
	}

	cssClipboard, err := readCSS("static/css/github-clipboard.css")
	if err != nil {
		return "", fmt.Errorf("read clipboard css: %w", err)
	}

	data := ExportData{
		Content:      content,
		CssLight:     template.CSS(cssLight),
		CssCodeLight: template.CSS(getCssCode("github")),
		CssMermaid:   template.CSS(cssMermaid),
		CssMathJax:   template.CSS(cssMathJax),
		CssClipboard: template.CSS(cssClipboard),
		IncludeJS:    includeJS,
	}

	tmpl, err := template.ParseFS(defaults.Templates, "templates/export.html")
	if err != nil {
		return "", fmt.Errorf("parse export template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute export template: %w", err)
	}

	return buf.String(), nil
}

// ExportFilename derives the output filename from a Markdown file path.
// README.md -> README.html,  path/to/doc.md -> doc.html
func ExportFilename(mdPath string) string {
	base := path.Base(mdPath)
	if !strings.HasSuffix(strings.ToLower(base), ".md") {
		return base
	}
	base = base[:len(base)-3]
	if base == "" {
		return "article.html"
	}
	return base + ".html"
}

// readCSS reads a CSS file from the embedded static filesystem.
func readCSS(embedPath string) (string, error) {
	data, err := defaults.StaticFiles.ReadFile(embedPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
