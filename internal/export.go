package internal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

// imgSrcPattern matches src="..." inside <img> tags.
var imgSrcPattern = regexp.MustCompile(`<img[^>]+src="([^"]+)"[^>]*>`)

// BuildExportHTML renders a standalone HTML page from rendered Markdown content.
// When includeJS is true, <script> tags for Mermaid and MathJax are included
// (for use when the file will be opened while the server is running).
// When includeJS is false, no JS is emitted (for CLI export or fully static output).
// rootDir is the base directory for resolving local image paths.
// Pass "" to skip image embedding.
func BuildExportHTML(content template.HTML, includeJS bool, rootDir string) (string, error) {
	htmlContent := string(content)

	if rootDir != "" {
		embedded, err := embedLocalImages(htmlContent, rootDir)
		if err != nil {
			return "", fmt.Errorf("embed images: %w", err)
		}
		htmlContent = embedded
	}

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
		Content:      template.HTML(htmlContent),
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

// embedLocalImages parses HTML content and replaces <img src="..."> attributes
// for local image files with base64-encoded data URIs.
// Images referenced by http://, https://, data:, or protocol-relative (//) URLs
// are left unchanged. Missing or unreadable image files are silently skipped.
func embedLocalImages(htmlContent string, rootDir string) (string, error) {
	return imgSrcPattern.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the src value
		submatch := imgSrcPattern.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		src := submatch[1]

		// Skip non-local URLs
		if strings.HasPrefix(src, "http://") ||
			strings.HasPrefix(src, "https://") ||
			strings.HasPrefix(src, "data:") ||
			strings.HasPrefix(src, "//") {
			return match
		}

		// URL-decode the path (e.g. %20 -> space)
		decoded, err := url.PathUnescape(src)
		if err != nil {
			return match
		}

		// Resolve path relative to rootDir
		imagePath := filepath.Join(rootDir, filepath.FromSlash(decoded))

		data, err := os.ReadFile(imagePath)
		if err != nil {
			// File not found or unreadable — leave the src as-is
			return match
		}

		// Determine MIME type from extension
		ext := strings.ToLower(filepath.Ext(decoded))
		mimeType := mime.TypeByExtension(ext)
		if mimeType == "" {
			switch ext {
			case ".svg":
				mimeType = "image/svg+xml"
			case ".ico":
				mimeType = "image/x-icon"
			default:
				// Unknown extension — skip
				return match
			}
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

		return strings.Replace(match, `src="`+src+`"`, `src="`+dataURI+`"`, 1)
	}), nil
}

// readCSS reads a CSS file from the embedded static filesystem.
func readCSS(embedPath string) (string, error) {
	data, err := defaults.StaticFiles.ReadFile(embedPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
