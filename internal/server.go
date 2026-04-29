package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/aarol/reload"
	chroma_html "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/showgp/go-grip/defaults"
)

type Server struct {
	parser       *Parser
	boundingBox  bool
	host         string
	port         int
	browser      bool
	enableReload bool
	strictPort   bool
}

type ServerOptions struct {
	Host         string
	Port         int
	BoundingBox  bool
	Browser      bool
	EnableReload bool
	StrictPort   bool
	Parser       *Parser
}

func NewServer(host string, port int, boundingBox bool, browser bool, enableReload bool, parser *Parser) *Server {
	return NewServerWithOptions(ServerOptions{
		Host:         host,
		Port:         port,
		BoundingBox:  boundingBox,
		Browser:      browser,
		EnableReload: enableReload,
		Parser:       parser,
	})
}

func NewServerWithOptions(opts ServerOptions) *Server {
	if opts.Parser == nil {
		opts.Parser = NewParser()
	}
	return &Server{
		host:         opts.Host,
		port:         opts.Port,
		boundingBox:  opts.BoundingBox,
		browser:      opts.Browser,
		enableReload: opts.EnableReload,
		strictPort:   opts.StrictPort,
		parser:       opts.Parser,
	}
}

func (s *Server) Serve(file string) error {
	target, err := resolveServeTarget(file)
	if err != nil {
		return err
	}

	var reloadMiddleware *reload.Reloader
	if s.enableReload {
		reloadMiddleware = reload.New(target.rootDir)
		reloadMiddleware.DebugLog = log.New(io.Discard, "", 0)
		// Fix WebSocket CORS issues for development
		reloadMiddleware.Upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	handler := s.newHandlerForTarget(target)

	if s.enableReload {
		handler = reloadMiddleware.Handle(handler)
		fmt.Printf("📡 Auto-reload enabled. Files will trigger browser refresh.\n")
	} else {
		fmt.Printf("🔄 Auto-reload disabled. Use F5 to manually refresh.\n")
	}

	listener, actualPort, err := listenOnPort(s.port, s.strictPort)
	if err != nil {
		return err
	}

	initialPath, err := initialPathForTarget(target)
	if err != nil {
		_ = listener.Close()
		return err
	}
	addr := fmt.Sprintf("http://%s:%d%s", s.host, actualPort, initialPath)
	fmt.Printf("🚀 Starting server: %s\n", addr)

	if s.browser {
		err := Open(addr)
		if err != nil {
			fmt.Println("❌ Error opening browser:", err)
		}
	}

	return http.Serve(listener, handler)
}

func (s *Server) newHandler(dir http.Dir) http.Handler {
	target := serveTarget{
		mode:    modeDirectory,
		rootDir: string(dir),
	}
	return s.newHandlerForTarget(target)
}

func (s *Server) newHandlerForTarget(target serveTarget) http.Handler {
	dir := http.Dir(target.rootDir)
	fileServer := http.FileServer(dir)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(defaults.StaticFiles)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestFile := cleanRequestPath(r.URL.Path)
		if requestFile == "" {
			setNoCacheHeaders(w)
			initialPath, err := initialPathForTarget(target)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if initialPath != "/" {
				http.Redirect(w, r, initialPath, http.StatusFound)
				return
			}
			s.renderEmpty(w, target)
			return
		}

		if isMarkdownFile(requestFile) {
			if target.mode == modeSingleFile && requestFile != target.initialFile {
				http.NotFound(w, r)
				return
			}

			isFile, err := isRegularFile(dir, requestFile)
			if err == nil && isFile {
				s.renderMarkdown(w, dir, target, requestFile)
				return
			}
		}

		isDirectory, err := isDirectory(dir, requestFile)
		if err == nil && isDirectory {
			setNoCacheHeaders(w)
			stripCacheValidators(r)
		}

		fileServer.ServeHTTP(w, r)
	})

	return mux
}

func (s *Server) renderMarkdown(w http.ResponseWriter, dir http.Dir, target serveTarget, currentFile string) {
	setNoCacheHeaders(w)

	bytes, err := readToString(dir, currentFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rendered, err := s.parser.Render(bytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	page, err := s.newPageData(target, currentFile, template.HTML(rendered.Content), rendered.TOC)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := serveTemplate(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderEmpty(w http.ResponseWriter, target serveTarget) {
	setNoCacheHeaders(w)

	content := template.HTML(`<div class="docs-empty"><h1>No Markdown files found</h1><p>Add a Markdown file to this directory and refresh the page.</p></div>`)
	page, err := s.newPageData(target, "", content, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := serveTemplate(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) newPageData(target serveTarget, currentFile string, content template.HTML, toc []TOCEntry) (htmlStruct, error) {
	var articles []Article
	if target.mode == modeDirectory {
		discovered, err := discoverArticles(target.rootDir)
		if err != nil {
			return htmlStruct{}, err
		}
		articles = articlesWithActive(discovered, currentFile)
	}

	return htmlStruct{
		Content:      content,
		BoundingBox:  s.boundingBox,
		CssCodeLight: template.CSS(getCssCode("github")),
		CssCodeDark:  template.CSS(getCssCode("github-dark")),
		ShowSidebar:  target.mode == modeDirectory,
		Articles:     articles,
		TOC:          toc,
	}, nil
}

func readToString(dir http.Dir, filename string) ([]byte, error) {
	f, err := dir.Open(filename)
	if err != nil {
		return nil, err
	}
	//nolint:errcheck
	defer f.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type htmlStruct struct {
	Content      template.HTML
	BoundingBox  bool
	CssCodeLight template.CSS
	CssCodeDark  template.CSS
	ShowSidebar  bool
	Articles     []Article
	TOC          []TOCEntry
}

func serveTemplate(w http.ResponseWriter, html htmlStruct) error {
	w.Header().Set("Content-Type", "text/html")
	tmpl, err := template.ParseFS(defaults.Templates, "templates/layout.html")
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, html)
	return err
}

func initialPathForTarget(target serveTarget) (string, error) {
	if target.mode == modeSingleFile {
		return "/" + urlPathEscape(target.initialFile), nil
	}

	articles, err := discoverArticles(target.rootDir)
	if err != nil {
		return "", err
	}
	initial := initialArticle(articles)
	if initial == "" {
		return "/", nil
	}
	return "/" + urlPathEscape(initial), nil
}

func cleanRequestPath(requestPath string) string {
	cleaned := strings.TrimPrefix(path.Clean("/"+requestPath), "/")
	unescaped, err := url.PathUnescape(cleaned)
	if err != nil {
		return cleaned
	}
	return unescaped
}

func urlPathEscape(file string) string {
	return url.PathEscape(file)
}

func getCssCode(style string) string {
	buf := new(strings.Builder)
	formatter := chroma_html.New(chroma_html.WithClasses(true))
	s := styles.Get(style)
	_ = formatter.WriteCSS(buf, s)
	return buf.String()
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func stripCacheValidators(r *http.Request) {
	r.Header.Del("If-Modified-Since")
	r.Header.Del("If-None-Match")
}

func isDirectory(dir http.Dir, name string) (bool, error) {
	file, err := dir.Open(name)
	if err != nil {
		return false, err
	}
	//nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	return info.IsDir(), nil
}

func isRegularFile(dir http.Dir, name string) (bool, error) {
	file, err := dir.Open(name)
	if err != nil {
		return false, err
	}
	//nolint:errcheck
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	return !info.IsDir(), nil
}
