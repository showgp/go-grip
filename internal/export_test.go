package internal

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"README.md", "README.html"},
		{"guide.md", "guide.html"},
		{"path/to/doc.md", "doc.html"},
		{"doc.MD", "doc.html"},
		{".md", "article.html"},
		{".MD", "article.html"},
		{"readme", "readme"},
	}

	for _, tc := range tests {
		got := ExportFilename(tc.input)
		if got != tc.expected {
			t.Errorf("ExportFilename(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestBuildExportHTML_ContainsExpectedElements(t *testing.T) {
	t.Parallel()

	content := template.HTML("<h1>Hello</h1><p>Test content</p>")
	result, err := BuildExportHTML(content, false, "")
	if err != nil {
		t.Fatalf("BuildExportHTML returned error: %v", err)
	}

	// Should contain basic structure
	for _, want := range []string{
		"<!doctype html>",
		`<title>go-grip export</title>`,
		`<body class="markdown-body">`,
		"<h1>Hello</h1>",
		"<p>Test content</p>",
		`<style>`, // CSS is inlined
		"container",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}

	// Should NOT contain chrome elements
	for _, notWant := range []string{
		"docs-sidebar",
		"docs-toc",
		"theme-toggle",
		"docs-page-nav",
		"class=\"footer\"",
		`<link rel="stylesheet"`,
	} {
		if strings.Contains(result, notWant) {
			t.Errorf("expected output NOT to contain %q", notWant)
		}
	}

	// Should NOT include JS when includeJS=false
	if strings.Contains(result, "<script>") {
		t.Errorf("expected no script tags when includeJS=false")
	}
}

func TestBuildExportHTML_IncludeJS(t *testing.T) {
	t.Parallel()

	content := template.HTML("<p>Test</p>")
	result, err := BuildExportHTML(content, true, "")
	if err != nil {
		t.Fatalf("BuildExportHTML returned error: %v", err)
	}

	// Should include JS references
	for _, want := range []string{
		"mermaid.min.js",
		"tex-mml-chtml.js",
		"mathjax-options.js",
		"mermaid-init.js",
		"clipboard-copy.js",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("expected output to contain %q when includeJS=true", want)
		}
	}
}

func TestBuildExportHTML_ContainsCssFiles(t *testing.T) {
	t.Parallel()

	content := template.HTML("<p>Hello</p>")
	result, err := BuildExportHTML(content, false, "")
	if err != nil {
		t.Fatalf("BuildExportHTML returned error: %v", err)
	}

	// Verify CSS content is inlined (spot-check for known CSS classes)
	for _, cssClass := range []string{
		"markdown-body",
		"chroma",         // Chroma highlighting
		"mermaid-error",  // Mermaid CSS
		"clipboard-copy", // Clipboard CSS
	} {
		if !strings.Contains(result, cssClass) {
			t.Errorf("expected output to contain CSS class %q", cssClass)
		}
	}

	// Dark theme CSS should NOT be present
	if strings.Contains(result, "github-markdown-dark") || strings.Contains(result, "prefers-color-scheme: dark") {
		t.Errorf("expected NO dark theme CSS in export")
	}
}

func TestEmbedLocalImages_EmbedLocalPNG(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create a minimal valid 1x1 red PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG header
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1 pixel
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, // 8-bit RGB
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00, // (compressed)
		0x00, 0x00, 0x03, 0x00, 0x01, 0x26, 0xE0, 0xFE, // ...
		0x87, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, // IEND chunk
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "test.png"), pngData, 0o644); err != nil {
		t.Fatalf("write test.png: %v", err)
	}

	htmlContent := `<p><img src="test.png" alt="test"></p>`
	result, err := embedLocalImages(htmlContent, tmpDir)
	if err != nil {
		t.Fatalf("embedLocalImages returned error: %v", err)
	}

	// Should now contain a data URI
	if !strings.Contains(result, "data:image/png;base64,") {
		t.Fatalf("expected data URI in output, got: %s", result)
	}
	// Original src should be replaced
	if strings.Contains(result, `src="test.png"`) {
		t.Fatalf("expected original src to be replaced, got: %s", result)
	}
}

func TestEmbedLocalImages_SkipsRemoteURLs(t *testing.T) {
	t.Parallel()

	htmlContent := `<p><img src="https://example.com/img.png" alt="remote"></p>
<img src="http://example.com/img.jpg" alt="http">
<img src="data:image/png;base64,abc" alt="data">
<img src="//cdn.example.com/img.gif" alt="protocol-relative">`
	result, err := embedLocalImages(htmlContent, "/tmp")
	if err != nil {
		t.Fatalf("embedLocalImages returned error: %v", err)
	}

	// All URLs should remain unchanged
	if result != htmlContent {
		t.Fatalf("expected remote URLs to remain unchanged:\ngot:  %s\nwant: %s", result, htmlContent)
	}
}

func TestEmbedLocalImages_SkipsMissingFile(t *testing.T) {
	t.Parallel()

	htmlContent := `<img src="nonexistent.png">`
	result, err := embedLocalImages(htmlContent, "/tmp")
	if err != nil {
		t.Fatalf("embedLocalImages returned error: %v", err)
	}

	// Should leave the src unchanged (silent skip)
	if result != htmlContent {
		t.Fatalf("expected missing image to be skipped unchanged:\ngot:  %s\nwant: %s", result, htmlContent)
	}
}

func TestEmbedLocalImages_MultipleImages(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create a minimal SVG
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"><rect width="1" height="1" fill="red"/></svg>`
	if err := os.WriteFile(filepath.Join(tmpDir, "icon.svg"), []byte(svgContent), 0o644); err != nil {
		t.Fatalf("write icon.svg: %v", err)
	}

	// Create a minimal PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x26, 0xE0, 0xFE,
		0x87, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "img.png"), pngData, 0o644); err != nil {
		t.Fatalf("write img.png: %v", err)
	}

	// Mix of local and remote
	htmlContent := `<img src="icon.svg"><img src="https://example.com/remote.png"><img src="img.png">`
	result, err := embedLocalImages(htmlContent, tmpDir)
	if err != nil {
		t.Fatalf("embedLocalImages returned error: %v", err)
	}

	// Local images should be embedded
	if !strings.Contains(result, "data:image/svg+xml;base64,") {
		t.Fatalf("expected SVG to be embedded, got: %s", result)
	}
	if !strings.Contains(result, "data:image/png;base64,") {
		t.Fatalf("expected PNG to be embedded, got: %s", result)
	}
	// Remote URL should remain
	if !strings.Contains(result, "https://example.com/remote.png") {
		t.Fatalf("expected remote URL to remain unchanged, got: %s", result)
	}
}

func TestEmbedLocalImages_BuildExportHTMLEndToEnd(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	// Create a small PNG
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x26, 0xE0, 0xFE,
		0x87, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "logo.png"), pngData, 0o644); err != nil {
		t.Fatalf("write logo.png: %v", err)
	}

	// BuildExportHTML with rootDir should embed the image
	content := template.HTML(`<p><img src="logo.png" alt="L"></p>`)
	result, err := BuildExportHTML(content, false, tmpDir)
	if err != nil {
		t.Fatalf("BuildExportHTML returned error: %v", err)
	}

	if !strings.Contains(result, "data:image/png;base64,") {
		t.Fatalf("expected embedded image in BuildExportHTML output, got: %s", result)
	}
}

func TestBuildExportHTML_BadContent(t *testing.T) {
	t.Parallel()

	// Empty content should still produce valid HTML
	result, err := BuildExportHTML("", false, "")
	if err != nil {
		t.Fatalf("BuildExportHTML with empty content returned error: %v", err)
	}
	if !strings.Contains(result, "<!doctype html>") {
		t.Errorf("expected valid HTML even with empty content")
	}
}
