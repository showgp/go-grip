package internal

import (
	"html/template"
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
	result, err := BuildExportHTML(content, false)
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
		`<style>`,   // CSS is inlined
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
	result, err := BuildExportHTML(content, true)
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
	result, err := BuildExportHTML(content, false)
	if err != nil {
		t.Fatalf("BuildExportHTML returned error: %v", err)
	}

	// Verify CSS content is inlined (spot-check for known CSS classes)
	for _, cssClass := range []string{
		"markdown-body",
		"chroma",       // Chroma highlighting
		"mermaid-error", // Mermaid CSS
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

func TestBuildExportHTML_BadContent(t *testing.T) {
	t.Parallel()

	// Empty content should still produce valid HTML
	result, err := BuildExportHTML("", false)
	if err != nil {
		t.Fatalf("BuildExportHTML with empty content returned error: %v", err)
	}
	if !strings.Contains(result, "<!doctype html>") {
		t.Errorf("expected valid HTML even with empty content")
	}
}
