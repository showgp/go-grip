package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	chroma_html "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/showgp/go-grip/defaults"
	"github.com/showgp/go-grip/internal/hotreload"
)

const maxEditSize = 10 << 20 // 10MB

var editLocks sync.Map

type Server struct {
	parser       *Parser
	boundingBox  bool
	host         string
	port         int
	browser      bool
	enableReload bool
	strictPort   bool
	recursive    bool
	rootDir      string
	pdfGenOnce   sync.Once
	pdfGen       *PDFGenerator
	pdfGenErr    error
}

type ServerOptions struct {
	Host         string
	Port         int
	BoundingBox  bool
	Browser      bool
	EnableReload bool
	StrictPort   bool
	Recursive    bool
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
		recursive:    opts.Recursive,
		parser:       opts.Parser,
	}
}

func (s *Server) Serve(file string) error {
	target, err := resolveServeTarget(file)
	if err != nil {
		return err
	}
	s.rootDir = target.rootDir

	var reloadMiddleware *hotreload.Reloader
	if s.enableReload {
		reloadMiddleware = hotreload.New(filepath.Clean(target.rootDir))
		reloadMiddleware.Upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}

	handler := s.newHandlerForTarget(target)

	if s.enableReload {
		handler = reloadMiddleware.Handle(handler)
		fmt.Printf("📡 Auto-reload enabled. Only .md files will trigger browser refresh.\n")
	} else {
		fmt.Printf("🔄 Auto-reload disabled. Use F5 to manually refresh.\n")
	}

	listener, actualPort, err := listenOnPort(s.port, s.strictPort)
	if err != nil {
		return err
	}

	initialPath, err := initialPathForTarget(target, s.recursive)
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
	s.rootDir = string(dir)
	target := serveTarget{
		mode:    modeDirectory,
		rootDir: string(dir),
	}
	return s.newHandlerForTarget(target)
}

func (s *Server) newHandlerForTarget(target serveTarget) http.Handler {
	s.rootDir = target.rootDir
	dir := http.Dir(target.rootDir)
	fileServer := http.FileServer(dir)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(defaults.StaticFiles)))
	mux.HandleFunc("/api/edit/", s.handleSave(dir))
	mux.HandleFunc("/api/raw/", s.handleRaw(dir))
	mux.HandleFunc("/export", s.handleExport)
	mux.HandleFunc("/pdf", s.handlePDF)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		requestFile := cleanRequestPath(r.URL.Path)
		if requestFile == "" {
			setNoCacheHeaders(w)
			initialPath, err := initialPathForTarget(target, s.recursive)
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

	page, err := s.newPageData(target, currentFile, template.HTML(rendered.Content), rendered.TOC, string(bytes))
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
	page, err := s.newPageData(target, "", content, nil, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := serveTemplate(w, page); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) newPageData(target serveTarget, currentFile string, content template.HTML, toc []TOCEntry, rawContent string) (htmlStruct, error) {
	var articles []Article
	var previousArticle Article
	var nextArticle Article
	if target.mode == modeDirectory {
		discovered, err := discoverArticles(target.rootDir, s.recursive)
		if err != nil {
			return htmlStruct{}, err
		}
		articles = articlesWithActive(discovered, currentFile)
		previousArticle, nextArticle = articleNavigation(discovered, currentFile)
	}

	return htmlStruct{
		Content:         content,
		BoundingBox:     s.boundingBox,
		CssCodeLight:    template.CSS(getCssCode("github")),
		CssCodeDark:     template.CSS(getCssCode("github-dark")),
		ShowSidebar:     target.mode == modeDirectory,
		SidebarTitle:    sidebarTitle(target),
		Articles:        articles,
		PreviousArticle: previousArticle,
		NextArticle:     nextArticle,
		TOC:             toc,
		CurrentFile:     currentFile,
		RawContent:      rawContent,
	}, nil
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	fileParam := r.URL.Query().Get("file")
	if fileParam == "" {
		http.Error(w, "missing file parameter", http.StatusBadRequest)
		return
	}

	// Prevent directory traversal (both raw and after normalization)
	if strings.Contains(fileParam, "..") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}
	cleaned := path.Clean("/" + fileParam)[1:]
	if cleaned == "" || strings.HasPrefix(cleaned, "/") || strings.Contains(cleaned, "..") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	dir := http.Dir(s.rootDir)
	bytes, err := readToString(dir, cleaned)
	if err != nil {
		http.Error(w, "file not found: "+fileParam, http.StatusNotFound)
		return
	}

	rendered, err := s.parser.Render(bytes)
	if err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	htmlContent, err := BuildExportHTML(template.HTML(rendered.Content), true, s.rootDir)
	if err != nil {
		http.Error(w, "export error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	outName := ExportFilename(fileParam)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, outName))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(htmlContent))
}

func (s *Server) handlePDF(w http.ResponseWriter, r *http.Request) {
	fileParam := r.URL.Query().Get("file")
	if fileParam == "" {
		http.Error(w, "missing file parameter", http.StatusBadRequest)
		return
	}

	if strings.Contains(fileParam, "..") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}
	cleaned := path.Clean("/" + fileParam)[1:]
	if cleaned == "" || strings.HasPrefix(cleaned, "/") || strings.Contains(cleaned, "..") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	dir := http.Dir(s.rootDir)
	bytes, err := readToString(dir, cleaned)
	if err != nil {
		http.Error(w, "file not found: "+fileParam, http.StatusNotFound)
		return
	}

	rendered, err := s.parser.Render(bytes)
	if err != nil {
		http.Error(w, "render error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Build PDF-optimized HTML
	htmlContent, err := buildPDFMarkup(template.HTML(rendered.Content))
	if err != nil {
		http.Error(w, "PDF markup error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Lazy-init PDF generator
	s.pdfGenOnce.Do(func() {
		s.pdfGen, s.pdfGenErr = NewPDFGenerator(2)
	})
	if s.pdfGenErr != nil {
		http.Error(w, "PDF generator: "+s.pdfGenErr.Error(), http.StatusInternalServerError)
		return
	}

	pdfBuf, err := s.pdfGen.GeneratePDF(htmlContent)
	if err != nil {
		http.Error(w, "PDF generation error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	outName := ExportFilename(fileParam)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, strings.Replace(outName, ".html", ".pdf", 1)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBuf)
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

func (s *Server) handleSave(dir http.Dir) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		file := strings.TrimPrefix(r.URL.Path, "/api/edit/")
		absPath, err := validateEditPath(dir, file)
		if err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}

		if err := checkWritable(absPath); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxEditSize)

		content, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Request body too large"})
			return
		}

		if err := writeToDir(absPath, content); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save file"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// handleRaw serves the raw Markdown content of a file for the editor.

func (s *Server) handleRaw(dir http.Dir) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		file := strings.TrimPrefix(r.URL.Path, "/api/raw/")
		absPath, err := validateEditPath(dir, file)
		if err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "File not found"})
			return
		}

		setNoCacheHeaders(w)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write(content)
	}
}

func validateEditPath(dir http.Dir, file string) (string, error) {
	cleaned := strings.TrimPrefix(path.Clean("/"+file), "/")
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("invalid file path")
	}
	if !isMarkdownFile(cleaned) {
		return "", fmt.Errorf("only .md files can be edited")
	}

	absRoot, err := filepath.Abs(string(dir))
	if err != nil {
		return "", fmt.Errorf("internal error")
	}
	absPath := filepath.Join(absRoot, filepath.FromSlash(cleaned))
	resolved, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If symlink resolution fails, fall back to clean absPath
		resolved = absPath
	}
	cleanResolved := filepath.Clean(resolved)
	cleanRoot := filepath.Clean(absRoot)
	if !strings.HasPrefix(cleanResolved, cleanRoot+string(filepath.Separator)) && cleanResolved != cleanRoot {
		return "", fmt.Errorf("path traversal detected")
	}

	if _, err := os.Stat(cleanResolved); os.IsNotExist(err) {
		return "", fmt.Errorf("file no longer exists")
	}

	return cleanResolved, nil
}

func writeToDir(absPath string, data []byte) error {
	muRaw, _ := editLocks.LoadOrStore(absPath, &sync.Mutex{})
	mu := muRaw.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	tmpFile := absPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		_ = os.Remove(tmpFile)
		return err
	}
	return os.Rename(tmpFile, absPath)
}

func checkWritable(absPath string) error {
	f, err := os.OpenFile(absPath, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: cannot write to file")
		}
		return fmt.Errorf("cannot open file for writing")
	}
	_ = f.Close()
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data map[string]string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

type htmlStruct struct {
	Content         template.HTML
	BoundingBox     bool
	CssCodeLight    template.CSS
	CssCodeDark     template.CSS
	ShowSidebar     bool
	SidebarTitle    string
	Articles        []Article
	PreviousArticle Article
	NextArticle     Article
	TOC             []TOCEntry
	CurrentFile     string
	RawContent      string
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

func initialPathForTarget(target serveTarget, recursive bool) (string, error) {
	if target.mode == modeSingleFile {
		return "/" + urlPathEscape(target.initialFile), nil
	}

	articles, err := discoverArticles(target.rootDir, recursive)
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
	parts := strings.Split(file, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func sidebarTitle(target serveTarget) string {
	if target.mode != modeDirectory {
		return ""
	}

	title := filepath.Base(filepath.Clean(target.rootDir))
	if title == "." || title == string(filepath.Separator) {
		return "Articles"
	}
	return title
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
