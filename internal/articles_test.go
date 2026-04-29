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

	articles, err := discoverArticles(tmpDir)
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

func TestArticlesWithActiveMarksCurrentArticle(t *testing.T) {
	t.Parallel()

	articles := []Article{
		{Filename: "README.md"},
		{Filename: "guide.md"},
	}
	active := articlesWithActive(articles, "guide.md")

	if active[0].Active {
		t.Fatalf("expected README.md to be inactive")
	}
	if !active[1].Active {
		t.Fatalf("expected guide.md to be active")
	}
}
