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

- Exporting generated HTML.
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

Later enhancement:

- Add `--recursive` to include nested Markdown files.

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

## Implementation Notes

- `internal.Parser.Render` now returns `RenderedDocument`, including rendered HTML and TOC entries.
- `MdToHTML` remains as a compatibility wrapper around the new render result.
- Directory mode scans only the selected root directory for `.md` files in this pass.
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
| Display filenames in the article sidebar | Filename display is deterministic and avoids parsing every document just to build navigation. |
| Place TOC on the right at wide widths and stack it on narrow screens | This keeps reading space stable on desktop while preserving mobile usability. |

## Open Questions

- Should hidden files and directories be ignored by default?
- Should recursive directory scanning be added behind a `--recursive` option?
- Should the sidebar optionally derive article titles from each file's first `h1` heading?
