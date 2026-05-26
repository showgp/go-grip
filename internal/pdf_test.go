package internal

import (
	"html/template"
	"strings"
	"testing"
)

func TestStripPrintCSSWrapper_RemovesWrapper(t *testing.T) {
	t.Parallel()

	input := `@media print {
  * { background: transparent !important; }
  body { color: #000; }
}`
	result := stripPrintCSSWrapper(input)

	if strings.Contains(result, "@media print") {
		t.Errorf("result should not contain @media print wrapper")
	}
	if !strings.Contains(result, "background: transparent") {
		t.Errorf("result should contain inner CSS content")
	}
	if !strings.Contains(result, "body { color: #000; }") {
		t.Errorf("result should contain body rule")
	}
}

func TestStripPrintCSSWrapper_RemovesURLAnnotations(t *testing.T) {
	t.Parallel()

	input := `@media print {
  .markdown-body a { text-decoration: underline; }

  /* Print URLs after links */
  .markdown-body a[href]:after {
    content: " (" attr(href) ")";
    font-size: 9pt;
  }

  .markdown-body img { max-width: 100%; }
}`
	result := stripPrintCSSWrapper(input)

	if strings.Contains(result, "a[href]:after") {
		t.Errorf("result should not contain a[href]:after rules, got:\n%s", result)
	}
	if !strings.Contains(result, "max-width: 100%") {
		t.Errorf("result should still contain non-url-annotation rules")
	}
}

func TestStripPrintCSSWrapper_RemovesCompanionURLRules(t *testing.T) {
	t.Parallel()

	input := `@media print {
  .markdown-body a[href]:after {
    content: " (" attr(href) ")";
  }
  .markdown-body a[href^="#"]:after,
  .markdown-body a[href^="/"]:after,
  .markdown-body a[href*="javascript:"]:after {
    content: "";
  }
  .markdown-body p { margin: 0; }
}`
	result := stripPrintCSSWrapper(input)

	if strings.Contains(result, "a[href]:after") {
		t.Errorf("result should not contain a[href]:after, got:\n%s", result)
	}
	if strings.Contains(result, "a[href^=") {
		t.Errorf("result should not contain a[href^= rules, got:\n%s", result)
	}
	if !strings.Contains(result, ".markdown-body p") {
		t.Errorf("result should still contain unaffected rules")
	}
}

func TestStripPrintCSSWrapper_HandlesNestedBraces(t *testing.T) {
	t.Parallel()

	input := `@media print {
  @page {
    margin: 2cm;
  }
  .markdown-body { color: #000; }
}`
	result := stripPrintCSSWrapper(input)

	if !strings.Contains(result, "@page") {
		t.Errorf("result should preserve nested @page rule")
	}
	if !strings.Contains(result, "margin: 2cm") {
		t.Errorf("result should preserve @page content")
	}
}

func TestStripPrintCSSWrapper_ActualFile(t *testing.T) {
	t.Parallel()

	css, err := readCSS("static/css/github-print.css")
	if err != nil {
		t.Fatalf("readCSS failed: %v", err)
	}

	result := stripPrintCSSWrapper(css)

	// Should not contain @media print wrapper
	if strings.Contains(result, "@media print") {
		t.Errorf("result still contains @media print wrapper")
	}

	// Should not contain a[href]:after rules
	if strings.Contains(result, "a[href]:after") {
		t.Errorf("result still contains a[href]:after")
	}

	// Should contain key print rules
	for _, want := range []string{
		"page-break-inside",
		"max-width: 100%",
		"text-decoration: underline",
		"orphans",
		"widows",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("expected result to contain %q", want)
		}
	}
}

func TestBuildPDFMarkup_ContainsExpectedElements(t *testing.T) {
	t.Parallel()

	content := template.HTML("<h1>PDF Test</h1><p>Content</p>")
	result, err := buildPDFMarkup(content, "")
	if err != nil {
		t.Fatalf("buildPDFMarkup returned error: %v", err)
	}

	// Basic structure
	for _, want := range []string{
		"<!doctype html>",
		"<body class=\"markdown-body\">",
		"<h1>PDF Test</h1>",
		"<p>Content</p>",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("expected result to contain %q", want)
		}
	}

	// CSS inlined
	if !strings.Contains(result, "<style>") {
		t.Errorf("expected inlined CSS in result")
	}
	if !strings.Contains(result, "chroma") {
		t.Errorf("expected Chroma CSS in result")
	}
	if !strings.Contains(result, "background: #fff") {
		t.Errorf("expected white background override in result")
	}

	// JS for MathJax/Mermaid/clipboard
	for _, want := range []string{
		"tex-mml-chtml.js",
		"mermaid.min.js",
		"clipboard-copy.js",
		"MathJax",
		"mermaid.initialize",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("expected result to contain %q", want)
		}
	}

	// NO chrome elements
	for _, notWant := range []string{
		"docs-sidebar",
		"docs-toc",
		"theme-toggle",
		"docs-page-nav",
		"docs-footer",
		"github-clipboard.css",
		"github-markdown-dark",
	} {
		if strings.Contains(result, notWant) {
			t.Errorf("expected result NOT to contain %q", notWant)
		}
	}
}
