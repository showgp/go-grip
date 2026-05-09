package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverArticlesSortsReadmeFirstAndIgnoresNonMarkdown(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	files := map[string]string{
		"zeta.md":    "# Zeta\n",
		"README.md":  "# Readme\n",
		"alpha.MD":   "# Alpha\n",
		"notes.txt":  "notes\n",
		"script.mdx": "# Not included\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "nested.md"), 0o755); err != nil {
		t.Fatalf("mkdir nested.md: %v", err)
	}

	nestedDir := filepath.Join(tmpDir, "nested")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "guide.md"), []byte("# Nested\n"), 0o644); err != nil {
		t.Fatalf("write nested guide: %v", err)
	}

	articles, err := discoverArticles(tmpDir, false)
	if err != nil {
		t.Fatalf("discover articles: %v", err)
	}

	want := []string{"README.md", "alpha.MD", "zeta.md"}
	if len(articles) != len(want) {
		t.Fatalf("expected %d articles, got %d: %#v", len(want), len(articles), articles)
	}
	for i, wantName := range want {
		if articles[i].Filename != wantName {
			t.Fatalf("article %d: expected %q, got %q", i, wantName, articles[i].Filename)
		}
	}
}

func TestDiscoverArticlesRecursiveBuildsDirectoryTree(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	files := map[string]string{
		"README.md":                 "# Readme\n",
		"zeta.md":                   "# Zeta\n",
		"docs/guide.md":             "# Guide\n",
		"docs/api/reference.md":     "# Reference\n",
		"docs/中文 guide.md":          "# Localized\n",
		"docs/assets/ignored.txt":   "ignored\n",
		"empty-directory/notes.txt": "ignored\n",
	}
	for name, content := range files {
		path := filepath.Join(tmpDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	articles, err := discoverArticles(tmpDir, true)
	if err != nil {
		t.Fatalf("discover recursive articles: %v", err)
	}

	if len(articles) != 3 {
		t.Fatalf("expected 3 root articles, got %d: %#v", len(articles), articles)
	}
	docs := articles[0]
	if !docs.IsDirectory || docs.Title != "docs" {
		t.Fatalf("expected docs directory first, got %#v", docs)
	}
	if len(docs.Children) != 3 {
		t.Fatalf("expected 3 docs children, got %d: %#v", len(docs.Children), docs.Children)
	}
	if !docs.Children[0].IsDirectory || docs.Children[0].Title != "api" {
		t.Fatalf("expected api directory first under docs, got %#v", docs.Children[0])
	}
	if got := docs.Children[0].Children[0].Path; got != "/docs/api/reference.md" {
		t.Fatalf("expected nested escaped path, got %q", got)
	}
	if got := docs.Children[2].Path; got != "/docs/%E4%B8%AD%E6%96%87%20guide.md" {
		t.Fatalf("expected escaped localized path, got %q", got)
	}
	if articles[1].Filename != "README.md" {
		t.Fatalf("expected root README first among files, got %#v", articles[1])
	}
	if articles[2].Filename != "zeta.md" {
		t.Fatalf("expected zeta last, got %#v", articles[2])
	}
}

func TestArticlesWithActiveMarksCurrentArticle(t *testing.T) {
	t.Parallel()

	articles := []Article{
		{Filename: "README.md"},
		{
			Title:       "docs",
			IsDirectory: true,
			Children: []Article{
				{Filename: "docs/guide.md"},
			},
		},
	}
	active := articlesWithActive(articles, "docs/guide.md")

	if active[0].Active {
		t.Fatalf("expected README.md to be inactive")
	}
	if !active[1].Expanded {
		t.Fatalf("expected docs directory to be expanded")
	}
	if !active[1].Children[0].Active {
		t.Fatalf("expected docs/guide.md to be active")
	}
}

func TestArticleNavigationUsesFlattenedSidebarOrder(t *testing.T) {
	t.Parallel()

	articles := []Article{
		{
			Title:       "docs",
			IsDirectory: true,
			Children: []Article{
				{Title: "guide.md", Filename: "docs/guide.md", Path: "/docs/guide.md"},
				{Title: "reference.md", Filename: "docs/reference.md", Path: "/docs/reference.md"},
			},
		},
		{Title: "README.md", Filename: "README.md", Path: "/README.md"},
		{Title: "zeta.md", Filename: "zeta.md", Path: "/zeta.md"},
	}

	previous, next := articleNavigation(articles, "README.md")
	if previous.Filename != "docs/reference.md" {
		t.Fatalf("expected previous article from nested directory, got %#v", previous)
	}
	if next.Filename != "zeta.md" {
		t.Fatalf("expected next article from root files, got %#v", next)
	}

	previous, next = articleNavigation(articles, "docs/guide.md")
	if previous.Filename != "" {
		t.Fatalf("expected first article to have no previous article, got %#v", previous)
	}
	if next.Filename != "docs/reference.md" {
		t.Fatalf("expected next nested article, got %#v", next)
	}
}
