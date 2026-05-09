package internal

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Article struct {
	Title       string
	Path        string
	Filename    string
	Active      bool
	IsDirectory bool
	Expanded    bool
	Children    []Article
}

func discoverArticles(rootDir string, recursive bool) ([]Article, error) {
	return discoverArticlesInDir(rootDir, "", recursive)
}

func discoverArticlesInDir(rootDir string, relDir string, recursive bool) ([]Article, error) {
	dirPath := rootDir
	if relDir != "" {
		dirPath = filepath.Join(rootDir, filepath.FromSlash(relDir))
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	articles := make([]Article, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			if !recursive {
				continue
			}

			childRelDir := path.Join(relDir, entry.Name())
			children, err := discoverArticlesInDir(rootDir, childRelDir, recursive)
			if err != nil {
				return nil, err
			}
			if len(children) == 0 {
				continue
			}

			articles = append(articles, Article{
				Title:       entry.Name(),
				Filename:    childRelDir,
				IsDirectory: true,
				Children:    children,
			})
			continue
		}

		if !isMarkdownFile(entry.Name()) {
			continue
		}

		name := path.Join(relDir, entry.Name())
		articles = append(articles, Article{
			Title:    entry.Name(),
			Path:     "/" + urlPathEscape(name),
			Filename: name,
		})
	}

	sort.SliceStable(articles, func(i, j int) bool { return articleLess(articles[i], articles[j]) })

	return articles, nil
}

func articlesWithActive(articles []Article, currentFile string) []Article {
	result := make([]Article, len(articles))
	for i := range result {
		result[i] = articles[i]
		if len(articles[i].Children) > 0 {
			result[i].Children = articlesWithActive(articles[i].Children, currentFile)
			result[i].Expanded = articlesContainActive(result[i].Children)
		}
		result[i].Active = !result[i].IsDirectory && result[i].Filename == currentFile
	}
	return result
}

func initialArticle(articles []Article) string {
	for _, article := range articles {
		if article.IsDirectory {
			if initial := initialArticle(article.Children); initial != "" {
				return initial
			}
			continue
		}
		return article.Filename
	}
	return ""
}

func articleNavigation(articles []Article, currentFile string) (Article, Article) {
	flatArticles := flattenArticles(articles)
	for i, article := range flatArticles {
		if article.Filename != currentFile {
			continue
		}

		var previous Article
		var next Article
		if i > 0 {
			previous = flatArticles[i-1]
		}
		if i < len(flatArticles)-1 {
			next = flatArticles[i+1]
		}
		return previous, next
	}
	return Article{}, Article{}
}

func flattenArticles(articles []Article) []Article {
	result := make([]Article, 0)
	for _, article := range articles {
		if article.IsDirectory {
			result = append(result, flattenArticles(article.Children)...)
			continue
		}
		result = append(result, article)
	}
	return result
}

func isMarkdownFile(name string) bool {
	return strings.EqualFold(filepath.Ext(name), ".md")
}

func articleLess(left Article, right Article) bool {
	if left.IsDirectory != right.IsDirectory {
		return left.IsDirectory
	}

	leftReadme := isReadmeArticle(left)
	rightReadme := isReadmeArticle(right)
	if leftReadme != rightReadme {
		return leftReadme
	}

	leftTitle := strings.ToLower(left.Title)
	rightTitle := strings.ToLower(right.Title)
	if leftTitle != rightTitle {
		return leftTitle < rightTitle
	}
	return strings.ToLower(left.Filename) < strings.ToLower(right.Filename)
}

func isReadmeArticle(article Article) bool {
	return !article.IsDirectory && strings.EqualFold(path.Base(article.Filename), "README.md")
}

func articlesContainActive(articles []Article) bool {
	for _, article := range articles {
		if article.Active || article.Expanded {
			return true
		}
	}
	return false
}
