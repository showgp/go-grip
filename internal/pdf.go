package internal

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/showgp/go-grip/defaults"
)

// PDFData carries the data used by the print template (templates/print.html).
type PDFData struct {
	Content      template.HTML
	CssLight     template.CSS // github-markdown-light.css
	CssPrint     template.CSS // stripped github-print.css
	CssCodeLight template.CSS // Chroma light syntax highlighting
	CssMermaid   template.CSS // github-mermaid.css
	CssMathJax   template.CSS // mathjax.css
}

// buildPDFMarkup renders a standalone HTML page optimized for PDF generation
// via headless Chrome. It inlines only the CSS relevant to print output:
// light-theme content CSS, stripped print CSS (without the @media print wrapper),
// and Chroma light syntax highlighting.
func buildPDFMarkup(content template.HTML, rootDir string) (string, error) {
	htmlContent := string(content)
	if rootDir != "" {
		embedded, err := embedLocalImages(htmlContent, rootDir)
		if err != nil {
			return "", fmt.Errorf("embed images: %w", err)
		}
		htmlContent = embedded
	}
	content = template.HTML(htmlContent)

	cssLight, err := readCSS("static/css/github-markdown-light.css")
	if err != nil {
		return "", fmt.Errorf("read light css: %w", err)
	}

	rawPrint, err := readCSS("static/css/github-print.css")
	if err != nil {
		return "", fmt.Errorf("read print css: %w", err)
	}
	cssPrint := stripPrintCSSWrapper(rawPrint)

	cssMermaid, err := readCSS("static/css/github-mermaid.css")
	if err != nil {
		return "", fmt.Errorf("read mermaid css: %w", err)
	}

	cssMathJax, err := readCSS("static/css/mathjax.css")
	if err != nil {
		return "", fmt.Errorf("read mathjax css: %w", err)
	}

	data := PDFData{
		Content:      content,
		CssLight:     template.CSS(cssLight),
		CssPrint:     template.CSS(cssPrint),
		CssCodeLight: template.CSS(getCssCode("github")),
		CssMermaid:   template.CSS(cssMermaid),
		CssMathJax:   template.CSS(cssMathJax),
	}

	tmpl, err := template.ParseFS(defaults.Templates, "templates/print.html")
	if err != nil {
		return "", fmt.Errorf("parse print template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute print template: %w", err)
	}

	return buf.String(), nil
}

// stripPrintCSSWrapper removes the outer @media print { ... } wrapper from the
// print CSS and strips the URL annotation rules (a[href]:after) that would
// clutter PDF output.
//
// github-print.css is entirely wrapped in @media print { ... }.
// For chromedp PDF generation we apply these rules unconditionally, so we
// extract the inner content and remove the annotations that print the full
// URL after each link.
func stripPrintCSSWrapper(rawCSS string) string {
	// Find the opening brace of @media print {
	startIdx := strings.Index(rawCSS, "{")
	if startIdx == -1 {
		return rawCSS
	}

	// Brace-count to find the matching closing brace
	depth := 1
	endIdx := startIdx + 1
	for endIdx < len(rawCSS) && depth > 0 {
		switch rawCSS[endIdx] {
		case '{':
			depth++
		case '}':
			depth--
		}
		endIdx++
	}
	if depth != 0 {
		// Unbalanced braces — return original as fallback
		return rawCSS
	}

	// Extract content inside the @media print { ... }
	inner := rawCSS[startIdx+1 : endIdx-1]

	// Remove URL annotation rules that print href after every link
	inner = removeURLAnnotationBlocks(inner)

	return strings.TrimSpace(inner)
}

// removeURLAnnotationBlocks strips CSS blocks containing a[href]:after and
// companion a[href^="..."]:after selectors that print link URLs after anchors.
func removeURLAnnotationBlocks(css string) string {
	// Remove the main "Print URLs after links" comment + rule block
	css = removeBlockContaining(css, "a[href]:after")

	// Remove the companion rule for relative links / anchors
	css = removeBlockContaining(css, "a[href^=")

	return css
}

// ---------------------------------------------------------------------------
// Phase 2: Chromedp integration for PDF generation
// ---------------------------------------------------------------------------

// PDFGenerator manages a headless Chrome browser pool for PDF generation.
type PDFGenerator struct {
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	sem           chan struct{}
	mu            sync.Mutex
	chromePath    string
	maxConcurrent int
}

// NewPDFGenerator creates a new PDFGenerator with a headless Chrome allocator.
// maxConcurrent controls how many PDFs can be generated in parallel (1-4).
func NewPDFGenerator(maxConcurrent int) (*PDFGenerator, error) {
	if maxConcurrent < 1 {
		maxConcurrent = 1
	}
	if maxConcurrent > 4 {
		maxConcurrent = 4
	}

	chromePath, err := findChrome()
	if err != nil {
		return nil, err
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	if chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	return &PDFGenerator{
		allocCtx:      allocCtx,
		allocCancel:   allocCancel,
		sem:           make(chan struct{}, maxConcurrent),
		chromePath:    chromePath,
		maxConcurrent: maxConcurrent,
	}, nil
}

// findChrome locates the Chrome/Chromium binary on the system.
func findChrome() (string, error) {
	// Try common locations in priority order (path entries first, then known locations)
	common := []string{
		"google-chrome",
		"google-chrome-stable",
		"chromium-browser",
		"chromium",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/snap/bin/chromium",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium-browser",
		"/usr/bin/chromium",
	}
	for _, p := range common {
		if _, err := exec.LookPath(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("Chrome/Chromium not found on the system; install Chrome or set PATH")
}

// GeneratePDF renders an HTML string to PDF using headless Chrome.
// It returns the PDF bytes ready to serve.
func (g *PDFGenerator) GeneratePDF(htmlContent string) ([]byte, error) {
	// Acquire semaphore
	select {
	case g.sem <- struct{}{}:
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("PDF generation timeout: all workers busy")
	}
	defer func() { <-g.sem }()

	// Create per-request chromedp context
	ctx, cancel := chromedp.NewContext(g.allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	// Use data URI to avoid needing a running server
	dataURI := "data:text/html," + url.PathEscape(htmlContent)

	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURI),
		chromedp.WaitReady("body"),
		// Allow time for MathJax and Mermaid client-side rendering
		chromedp.Sleep(3*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPaperWidth(210.0 / 25.4).  // A4 width in inches
				WithPaperHeight(297.0 / 25.4). // A4 height in inches
				WithMarginTop(1.5 / 2.54).     // 1.5cm top margin
				WithMarginBottom(1.5 / 2.54).  // 1.5cm bottom margin
				WithMarginLeft(2.0 / 2.54).    // 2.0cm left margin
				WithMarginRight(1.5 / 2.54).   // 1.5cm right margin
				WithPreferCSSPageSize(true).   // respect CSS @page rules when set
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp PDF generation: %w", err)
	}

	return pdfBuf, nil
}

// Close shuts down the Chrome browser allocator and releases resources.
func (g *PDFGenerator) Close() {
	g.allocCancel()
}

// removeBlockContaining removes a CSS rule block starting from the line
// containing marker back to the preceding comment (or line), forward through
// the matching closing brace.
func removeBlockContaining(css, marker string) string {
	idx := strings.Index(css, marker)
	if idx == -1 {
		return css
	}

	// Walk back to find the start of the block:
	//   1. Look for a preceding "/*" comment and start from there
	//   2. Otherwise start from the beginning of the line containing marker
	blockStart := idx
	// Search backwards for "/*" within reasonable distance (40 chars back)
	searchBack := idx
	if searchBack > 40 {
		searchBack = idx - 40
	}
	commentStart := strings.LastIndex(css[searchBack:idx], "/*")
	if commentStart != -1 {
		blockStart = searchBack + commentStart
	} else {
		// Start from the beginning of the line
		lineStart := strings.LastIndex(css[:idx], "\n")
		if lineStart != -1 {
			blockStart = lineStart
		}
	}

	// Find the opening brace of this rule block after marker
	braceIdx := strings.Index(css[idx:], "{")
	if braceIdx == -1 {
		return css
	}
	braceIdx += idx

	// Brace-count forward to find matching closing brace
	depth := 1
	endIdx := braceIdx + 1
	for endIdx < len(css) && depth > 0 {
		switch css[endIdx] {
		case '{':
			// Only count braces that aren't inside string values
			// (url(), attr(), content values use { ... } patterns)
			// For our limited use case, this works because a[href]:after
			// blocks don't contain nested braces
			depth++
		case '}':
			depth--
		}
		endIdx++
	}
	if depth != 0 {
		return css
	}

	// Remove everything from blockStart to just after the closing brace
	result := css[:blockStart] + css[endIdx:]

	// Recursively remove remaining occurrences
	return removeBlockContaining(result, marker)
}
