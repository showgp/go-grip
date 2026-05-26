# go-grip Documentation Navigation Implementation Plan

## Purpose

This document tracks the implementation plan and progress for improving go-grip from a single Markdown preview tool into a more convenient local documentation browser.

The planned work covers:

- Article TOC generation for the currently rendered Markdown file.
- Directory mode with a persistent sidebar listing Markdown articles.
- Separate behavior for single-file mode and directory mode.
- Default-port conflict handling.

This file should be updated during implementation with progress, decisions, and relevant notes.

## Current Confirmed Behavior

- Markdown files are rendered one at a time when their `.md` URL is requested.
- Heading IDs are generated, but there is no generated TOC UI.
- Running with no argument opens `README.md` if present.
- Running against a directory currently behaves like a file server plus Markdown rendering for clicked `.md` files.
- The default port is fixed at `6419`, so running another instance with the default settings can fail due to port conflicts.

## Target Behavior

### Single-File Mode

Triggered by:

```bash
go-grip README.md
```

Expected behavior:

- Render only the selected Markdown file.
- Generate a TOC for the current article.
- Do not show the directory article sidebar.
- Preserve existing Markdown rendering features.
- Preserve auto-reload behavior for edits.
- Apply automatic default-port fallback when the default port is occupied.

### Directory Mode

Triggered by:

```bash
go-grip
go-grip .
go-grip docs
```

Expected behavior:

- Start one local documentation browser for the target directory.
- Render a persistent sidebar containing Markdown article entries from the directory.
- Keep the sidebar visible while navigating between articles.
- Highlight the currently selected article in the sidebar.
- Render the selected article in the main content area.
- Generate a TOC for the selected article.
- Prefer `README.md` as the initial article when present.
- Otherwise open the first discovered Markdown file.
- If no Markdown files exist, show a useful empty state.

### Port Handling

Expected behavior:

- Continue to prefer port `6419` by default.
- If the default port is occupied, automatically try the next available port.
- Print and open the actual URL that was selected.
- If the user explicitly provides `--port`, keep that request strict and report an error if the port is unavailable.

## Implementation Scope

### In Scope

- Parser changes needed to return rendered HTML plus TOC metadata.
- Server routing changes to distinguish single-file mode from directory mode.
- Directory Markdown discovery for sidebar navigation.
- Template changes for article sidebar and TOC rendering.
- CSS additions for the documentation layout.
- Port availability detection and fallback for default port.
- Tests for parser metadata, routing behavior, directory discovery, and port selection helpers.
- README usage updates after implementation.

### Out of Scope For First Pass

- Full-text search.
- Client-side routing or single-page-app behavior.
- Recursive directory scanning by default.
- Custom sorting configuration.
- Multi-root documentation workspaces.

## Proposed Design

### Rendering Model

The parser should return a structured render result instead of only raw HTML.

Conceptual result:

```text
RenderedDocument
- HTML content
- TOC entries
```

Each TOC entry should contain:

- Heading level
- Plain heading text
- Anchor ID

The existing `MdToHTML` behavior can either be preserved as a compatibility wrapper or replaced carefully where call sites are updated.

### Server Mode Selection

The server should determine the target mode before creating routes:

- If the argument is an existing regular file, use single-file mode.
- If the argument is an existing directory, use directory mode rooted at that directory.
- If no argument is provided, use directory mode rooted at the current directory.

Important edge case:

- Current code uses `path.Dir(file)` and `path.Base(file)`, which works for a file but is not enough for directory mode. This should be replaced with explicit filesystem target resolution.

### Directory Article Discovery

For the first pass:

- Scan only the root directory.
- Include files ending in `.md` case-insensitively.
- Sort articles alphabetically.
- Put `README.md` first when present.
- Store each article with display title and URL path.

Recursive enhancement:

- Add `--recursive` / `-r` to include nested Markdown files.
- Render recursive results as a nested sidebar tree.

### Layout

Single-file mode:

```text
main content + article TOC
```

Directory mode:

```text
left sidebar: article list
main content: selected article
right or inline area: article TOC
```

The existing GitHub Markdown visual style should remain the baseline. New layout CSS should be minimal and should not disrupt rendered Markdown content.

### Routes

Suggested routes:

- `/static/` serves embedded assets as today.
- `/*.md` renders Markdown files.
- `/` redirects or renders the initial article in directory mode.
- In single-file mode, `/` can redirect to the selected file.

Directory mode should prevent navigation outside the selected root.

### Port Fallback

Suggested behavior:

- Track whether the user explicitly set `--port`.
- If not explicit, try `6419`, then increment until a port is available.
- If explicit, attempt only that port.
- Use an explicit listener instead of `http.ListenAndServe` so the actual selected port is known before opening the browser.

## Implementation Tasks

### Phase 1: Foundation

- [x] Add mode/target resolution for file vs directory input.
- [x] Add article discovery for directory mode.
- [x] Add tests for target resolution and article discovery.

### Phase 2: TOC Generation

- [x] Change parser rendering to return HTML plus TOC metadata.
- [x] Extract headings from the parsed Markdown AST.
- [x] Ensure generated TOC anchors match rendered heading IDs.
- [x] Add parser tests for nested headings and duplicate headings.

### Phase 3: Templates And Layout

- [x] Extend template data to include sidebar articles, current article, and TOC entries.
- [x] Add sidebar markup for directory mode.
- [x] Add TOC markup for current article.
- [x] Add CSS for documentation layout.
- [x] Verify single-file mode does not show the article sidebar.

### Phase 4: Routing

- [x] Update server routes for single-file mode.
- [x] Update server routes for directory mode.
- [x] Default directory mode to `README.md` when available.
- [x] Add empty-state page for directories without Markdown files.
- [x] Add tests for routing and template output.

### Phase 5: Port Handling

- [x] Detect whether `--port` was explicitly supplied.
- [x] Add default-port fallback.
- [x] Switch server startup to use a listener so the selected port is known before opening the browser.
- [x] Add tests for port selection helper behavior where practical.

### Phase 6: Documentation And Verification

- [x] Update README usage examples.
- [x] Run `go test ./...`.
- [x] Manually test single-file mode.
- [x] Manually test directory mode with multiple Markdown files.
- [x] Manually test default-port conflict fallback.

### Phase 7: TOC Interaction Polish

- [x] Add smooth scrolling when clicking right-side TOC entries.
- [x] Add active highlighting for the current right-side TOC entry while scrolling.
- [x] Preserve reduced-motion preferences.
- [x] Add template coverage for the TOC interaction script.

### Phase 8: Recursive Directory Navigation

- [x] Add `--recursive` / `-r` for nested Markdown discovery in directory mode.
- [x] Preserve root-only directory discovery by default.
- [x] Render nested directories as a sidebar tree.
- [x] Add tests for recursive article discovery and nested Markdown routing.

## Acceptance Criteria

- Running `go-grip README.md` renders the selected file with a current-article TOC and no article-list sidebar.
- Running `go-grip`, `go-grip .`, or `go-grip docs` opens a directory documentation view with a persistent sidebar.
- Clicking sidebar entries renders different Markdown files without restarting the server.
- The current sidebar item is visibly active.
- Each rendered article has a TOC based on its own headings.
- Existing Markdown extensions continue to work, including Mermaid, MathJax, alerts, task lists, syntax highlighting, and issue links.
- Starting a second default instance does not fail only because port `6419` is occupied.
- Explicit `--port` conflicts are reported clearly.
- Automated tests pass with `go test ./...`.

## Progress Log

| Date | Status | Notes |
| --- | --- | --- |
| 2026-04-29 | Planned | Initial implementation plan created. |
| 2026-04-29 | Implemented | Added TOC metadata generation, directory sidebar navigation, single-file/directory mode routing, default-port fallback, README updates, and automated tests. |
| 2026-04-29 | Polished | Added smooth TOC scrolling and current-section highlighting for the right-side article TOC. |
| 2026-05-08 | Implemented | Added opt-in recursive directory navigation with `--recursive` / `-r` and nested sidebar rendering. |

## Implementation Notes

- `internal.Parser.Render` now returns `RenderedDocument`, including rendered HTML and TOC entries.
- `MdToHTML` remains as a compatibility wrapper around the new render result.
- Directory mode scans only the selected root directory for `.md` files in this pass.
- Recursive directory mode is opt-in with `--recursive` / `-r` and includes nested `.md` files in a tree sidebar.
- Directory mode shows filenames in the sidebar and marks the current file as active.
- Single-file mode does not show the article sidebar and returns `404` for other Markdown files in the same directory.
- Default-port fallback is implemented through an explicit listener before browser launch, so the printed/opened URL uses the actual selected port.
- Explicit `--port` remains strict and reports a bind error if unavailable.
- Right-side TOC interactions are handled by `defaults/static/js/toc-active.js`.
- Smooth scrolling is CSS-backed and also triggered from TOC clicks; reduced-motion preference disables CSS and scripted smooth scrolling.
- The active TOC item uses `aria-current="location"` so the visual state has an accessibility signal.
- Active TOC detection is based on document scroll position and heading offsets, avoiding viewport observer lag near heading boundaries.

## Decisions

| Decision | Rationale |
| --- | --- |
| Keep single-file and directory mode separate | Single-file previews should stay lightweight, while directory mode should behave like a local documentation browser. |
| Use a persistent sidebar in directory mode | It is more usable than a separate entry page because navigation remains visible while reading articles. |
| Generate TOC per current article | The TOC should represent the currently rendered Markdown file, not the entire directory. |
| Default port can auto-fallback, explicit port stays strict | This balances convenience with predictable user intent. |
| Do not merge all Markdown files into one page in the first pass | Separate article navigation avoids large pages, heading collisions, and slower rendering. |
| Scan only root-level Markdown files in the first pass | This keeps navigation predictable and leaves recursive discovery for a dedicated option. |
| Keep recursive scanning behind `--recursive` / `-r` | This preserves existing directory-mode behavior while making deeper documentation trees available when requested. |
| Display filenames in the article sidebar | Filename display is deterministic and avoids parsing every document just to build navigation. |
| Place TOC on the right at wide widths and stack it on narrow screens | This keeps reading space stable on desktop while preserving mobile usability. |

## Open Questions

- Should hidden files and directories be ignored by default?
- Should the sidebar optionally derive article titles from each file's first `h1` heading?



---

## HTML Export Feature Plan

### Feature Overview

Add the ability to export a rendered Markdown article as a self-contained standalone HTML file. The exported HTML must preserve all content and light-theme styling, exclude all navigation chrome, and be accessible both from the browser UI (download button) and the command line (`--export` flag). No new external dependencies required.

Reference docs: `docs/requirements-html-export.md`, `docs/research-html-export.md`

### New Files

| File | Purpose |
|------|---------|
| `internal/export.go` | Export logic: build standalone HTML, CSS inlining helpers, shared data struct |
| `defaults/templates/export.html` | Clean export template with inlined CSS, no chrome |

### Modified Files

| File | Changes |
|------|---------|
| `internal/server.go` | Add `/export/` route handler, wire into `newHandlerForTarget` mux |
| `cmd/root.go` | Add `--export` and `--output` flags, early-return branch before server start |
| `defaults/templates/layout.html` | Add "Export HTML" button to main content area |

### Phase 1: CSS Inlining Helpers + Export Data Struct (`internal/export.go`)

- [ ] Create `internal/export.go` with an `ExportData` struct for the export template

```go
type ExportData struct {
    Content       template.HTML
    CssLight      template.CSS   // github-markdown-light.css
    CssCodeLight  template.CSS   // Chroma light highlighting (same as CssCodeLight in server.go)
    CssMermaid    template.CSS   // github-mermaid.css
    CssMathJax    template.CSS   // mathjax.css
    CssClipboard  template.CSS   // github-clipboard.css
    IncludeJS     bool           // whether to include Mermaid/MathJax <script> tags
}
```

- [ ] Add a helper function `buildExportHTML(content template.HTML, includeJS bool) (string, error)` that:
  1. Reads CSS files from `defaults.StaticFiles` via `embed.FS.ReadFile("static/css/...")`
  2. Wraps each in `<style>` tags as `template.CSS`
  3. Calls `getCssCode("github")` for Chroma light highlighting
  4. Executes the export template with the assembled data
  5. Returns the rendered HTML string

CSS inclusion matrix:

| CSS File | Include? | Reason |
|----------|----------|--------|
| `github-markdown-light.css` | ✅ Yes | Main content light-theme styling |
| `github-markdown-dark.css` | ❌ No | Dark theme not needed for export |
| Chroma light (`getCssCode("github")`) | ✅ Yes | Syntax highlighting, already server-generated |
| `github-mermaid.css` (132 lines) | ✅ Yes | Mermaid diagram styling |
| `mathjax.css` (7 lines) | ✅ Yes | MathJax rendering support |
| `github-clipboard.css` | ✅ Yes | Code block copy button styling |
| `theme-switch.css` | ❌ No | No theme toggle in export |
| `docs-layout.css` | ❌ No | No sidebar/TOC layout |
| `github-print.css` | ❌ No | Print-only, for PDF export |

### Phase 2: Export Template (`defaults/templates/export.html`)

- [ ] Create `defaults/templates/export.html` with the following structure:

```html
<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>go-grip export</title>
  <style>{{ .CssLight }}</style>
  <style>{{ .CssCodeLight }}</style>
  <style>{{ .CssMermaid }}</style>
  <style>{{ .CssMathJax }}</style>
  <style>{{ .CssClipboard }}</style>
  {{ if .IncludeJS }}
  <script src="/static/js/mathjax-options.js"></script>
  <script src="/static/js/tex-mml-chtml.js"></script>
  <script src="/static/js/mermaid.min.js"></script>
  <script src="/static/js/mermaid-init.js"></script>
  <script src="/static/js/clipboard-copy.js"></script>
  {{ end }}
</head>
<body class="markdown-body">
  <div class="container">
    <div class="container-inner">
      {{ .Content }}
    </div>
  </div>
</body>
</html>
```

Key design decisions:
- **No navigation chrome**: no `<aside class="docs-sidebar">`, no `<aside class="docs-toc">`, no theme toggle, no prev/next article nav, no footer
- **Light theme only**: only light-theme CSS is inlined; no `prefers-color-scheme` media queries needed
- **JS loads from server**: `<script>` tags reference live server paths; works when the server is running
- **`IncludeJS` guard**: allows CLI export to optionally omit JS for fully static output
- **No bounding box**: the `container-inner` border styling is included via CSS, not via `BoundingBox` template flag

### Phase 3: Server `/export/` Route (`internal/server.go`)

- [ ] Add a `content-type` and `content-disposition` helper or inline headers in the handler

- [ ] Add `handleExport` method to `*Server`:

```go
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
    fileParam := r.URL.Query().Get("file")
    if fileParam == "" || !strings.HasSuffix(strings.ToLower(fileParam), ".md") {
        http.Error(w, "missing or invalid file parameter", http.StatusBadRequest)
        return
    }

    // Security: prevent directory traversal
    cleaned := path.Clean("/" + fileParam)[1:]
    if cleaned == "" {
        http.Error(w, "invalid file path", http.StatusBadRequest)
        return
    }

    // Read file from the serve target's root directory
    target := serveTarget{mode: modeDirectory, rootDir: s.rootDir}
    dir := http.Dir(target.rootDir)
    bytes, err := readToString(dir, cleaned)
    if err != nil {
        http.Error(w, "file not found: "+fileParam, http.StatusNotFound)
        return
    }

    // Render Markdown
    rendered, err := s.parser.Render(bytes)
    if err != nil {
        http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Build standalone HTML
    htmlContent := buildExportHTML(template.HTML(rendered.Content), true)

    // Derive filename
    outName := strings.TrimSuffix(fileParam, ".md") + ".html"
    baseName := path.Base(outName)

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, baseName))
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(htmlContent))
}
```

Note: `s.rootDir` is not currently a field on `Server`. The server's root directory is determined during `Serve()` and stored in the `serveTarget`. The handler needs access to the target's root directory. Options:
  - Store `rootDir` as a field on `Server` (set during `Serve()`)
  - Or store the `serveTarget` on the `Server` struct
  - Or pass the directory through to the handler at registration time (closure)

Recommended: Store `rootDir string` on `Server` (set in `Serve()` via `resolveServeTarget`).

- [ ] Register the `/export/` handler in `newHandlerForTarget` (around server.go line 130):

```go
mux.HandleFunc("/export/", s.handleExport)
```

- [ ] (Optional) Reuse the existing `readToString` helper (server.go lines 239-253) — already available.

- [ ] Add `serveExportTemplate` function for use by `buildExportHTML`:

```go
func serveExportTemplate(w io.Writer, data ExportData) error {
    tmpl, err := template.ParseFS(defaults.Templates, "templates/export.html")
    if err != nil {
        return err
    }
    return tmpl.Execute(w, data)
}
```

### Phase 4: CLI `--export` + `--output` Flags (`cmd/root.go`)

- [ ] Add two new flags in `init()`:

```go
rootCmd.Flags().String("export", "", "Export a Markdown file as standalone HTML")
rootCmd.Flags().String("output", "", "Output file path (used with --export; default: stdout)")
```

- [ ] Add early-return branch at the top of `RunE`, before server start:

```go
RunE: func(cmd *cobra.Command, args []string) error {
    exportFile, _ := cmd.Flags().GetString("export")
    outputFile, _ := cmd.Flags().GetString("output")

    if exportFile != "" {
        // Read the file
        data, err := os.ReadFile(exportFile)
        if err != nil {
            return fmt.Errorf("read %q: %w", exportFile, err)
        }

        // Render
        parser := internal.NewParser()
        rendered, err := parser.Render(data)
        if err != nil {
            return fmt.Errorf("render %q: %w", exportFile, err)
        }

        // Build standalone HTML (no JS for CLI export)
        htmlContent := internal.BuildExportHTML(template.HTML(rendered.Content), false)

        // Write output
        if outputFile != "" {
            if err := os.WriteFile(outputFile, []byte(htmlContent), 0o644); err != nil {
                return fmt.Errorf("write %q: %w", outputFile, err)
            }
        } else {
            fmt.Print(htmlContent)
        }
        return nil
    }

    // ... existing server startup code ...
```

Usage examples:
```bash
go-grip --export README.md > README.html           # pipe to stdout
go-grip --export README.md --output article.html    # write to file
go-grip --export README.md | pbcopy                 # pipe to clipboard
```

Note: When `--export` is used, the positional `args[0]` is NOT consumed. The `--export` flag takes the file path directly, avoiding confusion between serve mode and export mode.

### Phase 5: Browser UI "Export HTML" Button

- [ ] Add an "Export HTML" button to `defaults/templates/layout.html` in the main content area

Placement: inside the `.container` div, near the top of the content area, as a subtle icon button. This ensures it's visible regardless of sidebar/TOC state.

```html
<div class="container">
  <div class="docs-export-bar">
    <button id="export-html" class="docs-export-btn" title="Export as HTML">
      <svg><!-- download icon --></svg>
      Export HTML
    </button>
  </div>
  <div {{if .BoundingBox }} class="container-inner" {{end}}>
    {{ .Content }}
  </div>
  ...
```

- [ ] Add minimal CSS for the export button (can go in `docs-layout.css` or a new small block in the template)

- [ ] Add JS handler — either inline in `layout.html` or in a new small JS file:

```js
document.getElementById('export-html')?.addEventListener('click', function() {
  // Get current file path from the URL
  const currentPath = window.location.pathname;
  if (!currentPath.endsWith('.md')) return;

  // Trigger download via navigation (simplest, reliable)
  window.location.href = '/export/?file=' + encodeURIComponent(currentPath.replace(/^\//, ''));
});
```

Alternative: Use `fetch()` + blob download for a more polished UX that keeps the user on the same page:

```js
document.getElementById('export-html')?.addEventListener('click', async function() {
  const currentPath = window.location.pathname;
  if (!currentPath.endsWith('.md')) return;
  try {
    const resp = await fetch('/export/?file=' + encodeURIComponent(currentPath.replace(/^\//, '')));
    if (!resp.ok) throw new Error('Export failed');
    const blob = await resp.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = currentPath.split('/').pop().replace(/\.md$/i, '.html');
    a.click();
    URL.revokeObjectURL(url);
  } catch (e) {
    alert('Export failed: ' + e.message);
  }
});
```

The `fetch` + blob approach is recommended — it doesn't navigate away from the current page and provides proper error feedback.

### Phase 6: Tests

- [ ] **Server handler test**: Test `/export/?file=README.md` returns `200` with `Content-Type: text/html` and `Content-Disposition: attachment`
  - File: `internal/server_test.go`
  - Pattern: similar to `TestMarkdownResponsesDisableCaching` (server_test.go:98-127)
  - Verify: response body contains `<!doctype html>`, contains rendered content, does NOT contain `docs-sidebar`, `docs-toc`, or `theme-toggle`

- [ ] **Server handler error tests**:
  - Request `/export/` without `?file=` param → expect 400
  - Request `/export/?file=nonexistent.md` → expect 404
  - Request `/export/?file=../etc/passwd` → expect 400 (directory traversal)

- [ ] **Export template smoke test**: Parse `export.html` with the Go template engine, execute it with sample data, verify output contains expected `<style>` blocks and `<body class="markdown-body">`

- [ ] **CSS inlining test**: Verify that `buildExportHTML` returns HTML containing `<style>` blocks for each included CSS file, and does NOT reference any `<link>` tags

- [ ] **CLI flag test** (if feasible in test environment):
  - Run `go run . --export README.md` in a test dir → verify stdout contains valid HTML
  - Run `go run . --export README.md --output out.html` → verify `out.html` exists and contains valid HTML

### Acceptance Criteria (HTML Export)

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC-HTML-1 | Browser download button is visible in the rendered page | Navigate to a Markdown file, inspect the page for an export button/icon |
| AC-HTML-2 | Clicking export downloads a `.html` file | Click the export button → browser downloads a file without leaving the page |
| AC-HTML-3 | Exported HTML contains all article content | Open exported file → all headings, text, code, tables, images are present |
| AC-HTML-4 | Exported HTML has light theme only | Inspect exported HTML → no dark theme CSS, no theme toggle |
| AC-HTML-5 | Exported HTML has no chrome | Exported file has no sidebar, no TOC, no prev/next nav, no footer |
| AC-HTML-6 | All CSS is in `<style>` blocks | Inspect exported HTML → no `<link rel="stylesheet">` tags |
| AC-HTML-7 | Syntax highlighting is present and light-themed | Code blocks in exported HTML have Chroma syntax highlighting classes |
| AC-HTML-8 | `--export README.md` writes HTML to stdout | Run command → stdout contains `<!doctype html>` with rendered content |
| AC-HTML-9 | `--export README.md --output out.html` writes to file | Run command → `out.html` exists with valid HTML |
| AC-HTML-10 | CLI export works without starting a server | Run `--export` when no server is running → succeeds immediately |
| AC-HTML-11 | HTTP response has correct headers | `/export/?file=test.md` returns `Content-Type: text/html` and `Content-Disposition: attachment` |
| AC-HTML-12 | Missing file → clear error | `/export/?file=nonexistent.md` returns 404; `--export nonexistent.md` returns error message |
| AC-HTML-13 | Export works in single-file mode and directory mode | Test both modes → export always renders the current article |
| AC-HTML-14 | Mermaid/MathJax script tags are present in export | Exported HTML contains `<script>` tags for `mermaid.min.js` and `tex-mml-chtml.js` |

### Progress Log

| Date | Status | Notes |
|------|--------|-------|
| 2026-05-26 | Implemented | HTML Export implementation complete. All tests pass. See Phase 1-6. |

### Open Questions (HTML Export)

- Should `--export` respect the `--recursive` flag for discovering files in subdirectories? (Current plan: no — `--export` takes a single file argument.)
- Should the "Export HTML" button be visible in single-file mode only, or in directory mode too? (Current plan: both modes. The button reads the current file from `window.location.pathname`.)
- Should the exported `<title>` be derived from the first heading of the article instead of a static "go-grip export"? (Nice-to-have enhancement.)
## PDF Export Feature Plan

### Feature Overview

Add server-side PDF export using chromedp (headless Chrome via CDP). The exported PDF must replicate the light-theme visual rendering of the article content with high fidelity, including all text, code blocks, syntax highlighting, tables, images, MathJax formulas, and Mermaid diagrams. Navigation chrome must be excluded. The PDF must be downloadable directly from the browser without opening a print dialog.

Reference docs: docs/requirements-pdf-export.md, docs/research-pdf-export.md

### New Files

| File | Purpose |
|------|---------|
| internal/pdf.go | PDF generation logic: stripped HTML markup builder, chromedp integration, browser pool management |
| defaults/templates/print.html | Print-optimized layout template (no sidebar/TOC/nav, CSS inlined) |

### Modified Files

| File | Changes |
|------|---------|
| internal/server.go | Add /pdf?file= route handler |
| go.mod | Add github.com/chromedp/chromedp and github.com/chromedp/cdproto |
| defaults/templates/layout.html | Add Export PDF button to browser UI |
| defaults/embed.go | No changes needed (CSS/JS already embedded) |

### Dependencies

- github.com/chromedp/chromedp (pure Go CDP client, ~13k stars)
- github.com/chromedp/cdproto (auto-generated CDP bindings)
- Chrome/Chromium binary on the system (auto-detected, no manual PATH config needed)
- No CGo or platform-specific dependencies

### Phase 1: Print Template + CSS Stripping (defaults/templates/print.html + internal/pdf.go)

- [ ] Create defaults/templates/print.html, a minimal HTML page with no chrome:

```html
<!doctype html>
<html>
<head>
  <meta charset="utf-8"/>
  <style>{{ .CssLight }}</style>
  <style>{{ .CssPrint }}</style>
  <style>{{ .CssCodeLight }}</style>
  <style>body { background: #fff; }</style>
  <!-- MathJax for client-side rendering inside headless Chrome -->
  <script>window.MathJax = { ... }</script>
  <script src="tex-mml-chtml.js"></script>
  <!-- Mermaid for client-side rendering inside headless Chrome -->
  <script src="mermaid.min.js"></script>
  <script>mermaid.initialize({ startOnLoad: true });</script>
</head>
<body class="markdown-body">
  <div class="container">
    <div class="container-inner">
      {{ .Content }}
    </div>
  </div>
</body>
</html>
```

- [ ] In internal/pdf.go, add helper function stripPrintCSSWrapper(rawCSS string) string that:

  1. Reads the existing defaults/static/css/github-print.css (374 lines, all rules wrapped in @media print { ... })
  2. Strips the outer @media print { ... } wrapper so rules apply unconditionally
  3. Removes or comments out the a[href]:after { content: " (" attr(href) ")"; } rule (would clutter PDF links)
  4. Returns clean CSS string for embedding as a <style> block

- [ ] In internal/pdf.go, add helper function buildPDFMarkup(renderedHTML template.HTML) (string, error) that:

  1. Reads github-markdown-light.css from defaults.StaticFiles
  2. Strips and processes github-print.css (via stripPrintCSSWrapper)
  3. Calls getCssCode("github") for light Chroma highlighting
  4. Assembles the data and executes the print template
  5. Returns the complete HTML string to feed to chromedp

CSS inclusion matrix for PDF:

| CSS File | Include? | Notes |
|----------|----------|-------|
| github-markdown-light.css | Yes | Main content light-theme styling |
| github-markdown-dark.css | No | Dark theme not needed for PDF |
| Chroma light (getCssCode("github")) | Yes | Syntax highlighting, already server-generated |
| github-print.css (stripped) | Yes | Print-optimized typography, page breaks, A4 layout |
| github-mermaid.css | Yes | Mermaid styling |
| mathjax.css | Yes | MathJax rendering (7 lines) |
| github-clipboard.css | No | Clipboard button styling not needed in PDF |
| theme-switch.css | No | No theme toggle |
| docs-layout.css | No | No sidebar/TOC layout |

### Phase 2: Chromedp Integration (internal/pdf.go)

- [ ] Add chromedp and cdproto to go.mod:

go get github.com/chromedp/chromedp@latest
go get github.com/chromedp/cdproto@latest

- [ ] Define a PDFGenerator struct in internal/pdf.go:

type PDFGenerator struct {
    allocCtx   context.Context
    cancel     context.CancelFunc
    sem        chan struct{}  // bounded semaphore for concurrency
    chromePath string
}

- [ ] Add NewPDFGenerator(maxConcurrent int) (*PDFGenerator, error):

  1. Creates a chromedp remote allocator context using chromedp.DefaultExecAllocatorOptions
  2. Adds a custom Chrome path option if findChrome() detects a non-standard location
  3. Initializes a buffered channel semaphore (maxConcurrent, default 2)
  4. Returns the generator (or error if no Chrome binary found)

- [ ] Add findChrome() helper:

  1. Checks common Chrome installation paths: /Applications/Google Chrome.app/... (macOS), Program Files (Windows), chromium/google-chrome (Linux PATH)
  2. Falls back to chromedp.FindExecPath() which handles common locations
  3. Returns path string or error with clear message: "Chrome/Chromium is required for PDF export. Install Chrome or chromium."

- [ ] Add generatePDF(ctx context.Context, htmlContent string) ([]byte, error) method:

  1. Acquires semaphore slot (respects context cancellation)
  2. Creates a new chromedp tab context from the allocator
  3. Navigates to data:text/html,<url-escaped HTML>
  4. Waits for page ready: chromedp.WaitReady("body")
  5. Waits for MathJax rendering: chromedp.Evaluate for window.MathJax?.startup?.documentReady, or chromedp.Sleep(3 * time.Second) as fallback
  6. Waits for Mermaid rendering: chromedp.WaitVisible(".mermaid svg", ...) or similar
  7. Calls page.PrintToPDF() with:

     WithPrintBackground(true)
     WithPaperWidth(210.0 / 25.4)   // A4 in inches
     WithPaperHeight(297.0 / 25.4)
     WithMarginTop(1.5 / 2.54)
     WithMarginBottom(1.5 / 2.54)
     WithMarginLeft(2.0 / 2.54)
     WithMarginRight(1.5 / 2.54)
     WithPreferCSSPageSize(true)

  8. Releases semaphore slot
  9. Returns PDF bytes

- [ ] Add Close() method to clean up chromedp resources

- [ ] Add renderWaitStrategy enum/option: Sleep, PollMathJax, PollMermaid, or All — Sleep is simplest for initial implementation

- [ ] (Future) Browser pool warmup: pre-allocate 1-2 chromedp tab contexts at server startup to avoid cold-start latency (~1-2s)

### Phase 3: Server /pdf Route + Integration (internal/server.go)

- [ ] Add a PDFGenerator field to Server struct:

type Server struct {
    parser       *Parser
    pdfGen       *PDFGenerator   // nil if chrome not available
    // ... existing fields ...
}

- [ ] In NewServerWithOptions, initialize PDFGenerator (handle nil gracefully — PDF export disabled if Chrome not found)

- [ ] Add handlePDFExport method on *Server:

func (s *Server) handlePDFExport(w http.ResponseWriter, r *http.Request) {
    if s.pdfGen == nil {
        http.Error(w, "PDF export unavailable: Chrome/Chromium not found", http.StatusServiceUnavailable)
        return
    }

    fileParam := r.URL.Query().Get("file")
    // ... same file resolution logic as HTML export handler ...

    // Read and render
    bytes, err := readToString(dir, cleaned)
    rendered, err := s.parser.Render(bytes)

    // Build stripped HTML for chromedp
    pdfHTML := buildPDFMarkup(template.HTML(rendered.Content))

    // Generate PDF
    pdfBytes, err := s.pdfGen.generatePDF(context.Background(), pdfHTML)

    // Return PDF
    outName := strings.TrimSuffix(fileParam, ".md") + ".pdf"
    w.Header().Set("Content-Type", "application/pdf")
    w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, path.Base(outName)))
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write(pdfBytes)
}

- [ ] Register /pdf route in newHandlerForTarget:

mux.HandleFunc("/pdf", s.handlePDFExport)

### Phase 4: Browser UI Export PDF Button (defaults/templates/layout.html)

- [ ] Add Export PDF button next to the Export HTML button, using the same placement in .container

- [ ] Add JS click handler:

document.getElementById('export-pdf')?.addEventListener('click', async function() {
    const currentPath = window.location.pathname;
    if (!currentPath.endsWith('.md')) return;
    try {
        const resp = await fetch('/pdf?file=' + encodeURIComponent(currentPath.replace(/^\//, '')));
        if (!resp.ok) throw new Error(await resp.text());
        const blob = await resp.blob();
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = currentPath.split('/').pop().replace(/\.md$/i, '.pdf');
        a.click();
        URL.revokeObjectURL(url);
    } catch (e) {
        alert('PDF export failed: ' + e.message);
    }
});

- [ ] Conditionally hide the PDF button if s.pdfGen is nil (Chrome not available) — pass a flag to template

### Phase 5: Concurrency, Hardening + Error Handling

- [ ] Add configurable max concurrent PDF generations: MaxConcurrentPDF int field on ServerOptions (default 2)

- [ ] Add timeout to chromedp context: 30-second timeout per PDF generation request

- [ ] Handle math rendering timeout gracefully: if MathJax times out, produce PDF with placeholder note rather than failing entirely

- [ ] Handle mermaid rendering timeout gracefully: if a Mermaid diagram fails to render, include a fallback text notice in the PDF

- [ ] Store the Chrome-not-found state and serve a friendly error page or disable the PDF button rather than crashing

- [ ] Add PDF generation timeout cleanup: if generatePDF exceeds deadline, cancel the chromedp context and release semaphore

### Phase 6: Docker/Deployment

- [ ] Add Dockerfile to the repo (if desired) that includes chromedp/headless-shell:

FROM chromedp/headless-shell:latest AS chrome
FROM golang:1.25-alpine
COPY --from=chrome /headless-shell /headless-shell
RUN apk add --no-cache ca-certificates tzdata
# ... build go binary ...
ENTRYPOINT ["/go-grip"]

- [ ] Update flake.nix (optional): add chromium to build inputs for nix-based deployments

### Phase 7: Tests

- [ ] Unit test: TestStripPrintCSSWrapper — verify @media print wrapper is removed, a[href]:after rule is removed, core print styles preserved

- [ ] Unit test: TestBuildPDFMarkup — verify output contains <style> blocks for light CSS, print CSS, and Chroma CSS; does NOT contain docs-layout.css, theme-switch.css, or sidebar/TOC markup

- [ ] Integration test: TestPDFExportRoute (skip if no Chrome) — start test server, hit /pdf?file=test.md, verify Content-Type: application/pdf, Content-Disposition: attachment; filename="test.pdf", response body starts with %PDF

- [ ] Integration test: TestPDFExportNoChrome — if Chrome not found, verify /pdf returns 503 with explanatory message

- [ ] Integration test: TestPDFExportMissingFile — /pdf?file=nonexistent.md returns 404

- [ ] Integration test: TestConcurrentPDFExports — fire 5 simultaneous PDF requests, verify all return valid PDFs

### Acceptance Criteria (PDF Export)

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC-PDF-1 | PDF download button is visible | Navigate to Markdown file → PDF button visible (hidden if no Chrome) |
| AC-PDF-2 | Clicking PDF button downloads .pdf file | Click → browser downloads PDF without print dialog |
| AC-PDF-3 | PDF contains all article content | Open PDF → all headings, text, code, tables, images present |
| AC-PDF-4 | Light theme only | PDF uses light background, light Chroma highlighting |
| AC-PDF-5 | No chrome in PDF | No sidebar, TOC, theme toggle, nav buttons in PDF |
| AC-PDF-6 | MathJax formulas rendered | PDF contains rendered math, not raw $...$ markup |
| AC-PDF-7 | Mermaid diagrams rendered | PDF contains rendered diagram, not raw mermaid source |
| AC-PDF-8 | A4 page size | PDF page dimensions are 210×297 mm |
| AC-PDF-9 | Hyperlinks clickable | Links in PDF are active/clickable |
| AC-PDF-10 | CJK text renders correctly | PDF shows Chinese/Japanese/Korean characters properly |
| AC-PDF-11 | File naming: README.md → README.pdf | Downloaded file named correctly |
| AC-PDF-12 | Error when Chrome not found | /pdf returns 503 with clear message |
| AC-PDF-13 | Error on missing file | /pdf?file=nonexistent.md returns 404 |
| AC-PDF-14 | Multiple concurrent exports work | 3 simultaneous PDF requests all succeed |
| AC-PDF-15 | Works fully offline | No network access required for PDF generation |
| AC-PDF-16 | Deterministic output | Same input + same settings → same PDF (byte-identical) |

### Decisions

| Decision | Rationale |
|----------|-----------|
| Use chromedp (not rod, not wkhtmltopdf) | chromedp is pure Go, actively maintained, largest community |
| Keep one persistent chromedp allocator | Avoids cold-start (~1-2s) on every request |
| Limit concurrency to 2-4 | Prevents Chrome from exhausting system memory |
| Strip @media print wrapper from CSS | chromedp doesn't trigger @media print; rules must apply unconditionally |
| Exclude a[href]:after URL annotations | Not useful in PDF; clutters output |
| A4 with 1.5-2cm margins | Matches existing @page rules in github-print.css |
| Chromedp.Sleep(2-3s) as default wait | Simpler than Evaluate polling; most stable across MathJax/Mermaid versions |
| No Dockerfile in initial pass | Can be added when container deployment is needed |

### Progress Log

| Date | Status | Notes |
|------|--------|-------|
| 2026-05-26 | Planned | PDF Export implementation plan written. |

