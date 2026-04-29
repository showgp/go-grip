package internal

import (
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Article struct {
	Title    string
	Path     string
	Filename string
	Active   bool
}

func discoverArticles(rootDir string) ([]Article, error) {
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	articles := make([]Article, 0)
	for _, entry := range entries {
		if entry.IsDir() || !isMarkdownFile(entry.Name()) {
			continue
		}

		name := entry.Name()
		articles = append(articles, Article{
			Title:    name,
			Path:     "/" + url.PathEscape(name),
			Filename: name,
		})
	}

	sort.SliceStable(articles, func(i, j int) bool {
		iReadme := strings.EqualFold(articles[i].Filename, "README.md")
		jReadme := strings.EqualFold(articles[j].Filename, "README.md")
		if iReadme != jReadme {
			return iReadme
		}
		return strings.ToLower(articles[i].Filename) < strings.ToLower(articles[j].Filename)
	})

	return articles, nil
}

func articlesWithActive(articles []Article, currentFile string) []Article {
	result := make([]Article, len(articles))
	copy(result, articles)
	for i := range result {
		result[i].Active = result[i].Filename == currentFile
	}
	return result
}

func initialArticle(articles []Article) string {
	if len(articles) == 0 {
		return ""
	}
	return articles[0].Filename
}

func isMarkdownFile(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".md")
}
