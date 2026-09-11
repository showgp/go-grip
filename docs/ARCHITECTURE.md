# go-grip Architecture Reference

> 最后更新: 2026-05-27
>
> 本文档是 go-grip 工程的正式代码结构参考，面向后续迭代开发。内容基于源码直接分析，优先于其他文档。

## 一、工程概要

| 属性 | 值 |
|---|---|
| 模块路径 | `github.com/showgp/go-grip` |
| Go 版本 | 1.26 |
| 许可证 | MIT |
| 定位 | 本地 Markdown 渲染服务器，GitHub 风格预览，无外部 API 依赖 |
| 上游 | [chrishrb/go-grip](https://github.com/chrishrb/go-grip) 的 fork |
| fork 增量 | 多文件目录浏览、TOC 导航、端口回退、递归目录、HTML/PDF 导出、在线编辑 |

## 二、目录结构

```
go-grip/
├── main.go                     # 入口: cmd.Execute()
├── cmd/
│   └── root.go                 # Cobra CLI 定义、标志注册、启动逻辑
├── internal/
│   ├── server.go               # HTTP 服务核心：路由、模板渲染、API、编辑保存
│   ├── server_test.go
│   ├── parser.go               # goldmark 解析器封装：渲染管线、扩展注册、TOC 提取
│   ├── parser_test.go
│   ├── listener.go             # 端口监听：默认回退 / 严格模式
│   ├── listener_test.go
│   ├── target.go               # 服务目标解析：单文件 / 目录模式判定
│   ├── target_test.go
│   ├── articles.go             # 文章发现、排序、树/平展、导航、README 优先
│   ├── articles_test.go
│   ├── open.go                 # 跨平台打开浏览器
│   ├── cjk_emphasis.go         # CJK 粗体/强调 AST 修复
│   ├── export.go               # HTML 导出：模板填充、图片内联 base64、文件名推导
│   ├── export_test.go
│   ├── pdf.go                  # PDF 生成：chromedp 无头 Chrome、打印模板
│   ├── pdf_test.go
│   └── hotreload/
│       ├── hotreload.go        # 文件监控 + WebSocket 热重载中间件
│       ├── fd_unix.go          # EMFILE/ENFILE 判定
│       └── fd_other.go         # 非 unix 平台的空实现
├── pkg/
│   ├── alert/                  # > [!NOTE/TIP/IMPORTANT/WARNING/CAUTION] 区块
│   ├── details/                # <details> 折叠状态持久化 (sessionStorage)
│   ├── footnote/               # [^1] PHP Markdown Extra 脚注
│   ├── ghissue/                # #123 / owner/repo#123 自动链接 + git remote 探测
│   ├── highlighting/           # 代码块 chroma 语法高亮 + 复制按钮
│   ├── mathjax/                # $...$ / $$...$$ / ```math LaTeX 公式
│   └── tasklist/               # - [x] / - [ ] GFM 任务列表
├── defaults/
│   ├── embed.go                # //go:embed 嵌入模板 + 静态资源 (embed.FS)
│   ├── templates/
│   │   ├── layout.html         # 主页面模板（侧边栏 + 内容 + TOC + 编辑区）
│   │   ├── export.html         # 独立 HTML 导出模板
│   │   └── print.html          # PDF 打印模板（无头 Chrome）
│   └── static/
│       ├── css/ (9 files)      # GitHub Markdown 样式、布局、编辑器、主题切换等
│       └── js/  (11 files)     # 侧边栏、TOC、主题、剪贴板、编辑器、Mermaid、MathJax
├── docs/                       # 设计文档、需求、实现计划、工程状态
├── .github/workflows/
│   ├── build.yml               # CI: build + test + gofmt + golangci-lint
│   └── release.yml             # CD: 6 架构交叉编译 → GitHub Release
├── go.mod / go.sum
├── flake.nix / flake.lock      # Nix 构建
├── mise.toml                   # mise 工具版本管理
└── .pre-commit-config.yaml
```

## 三、启动流程

```
main.go: main()
  └─ cmd.Execute()                          // Cobra 解析标志
       ├─ --export 非空？
       │   ├─ internal.NewParser()
       │   ├─ parser.Render(data)           // Markdown → HTML + TOC
       │   ├─ internal.BuildExportHTML()    // 填充 export.html 模板 + 图片内联
       │   └─ 写入文件或 stdout → 退出
       │
       └─ 否则启动服务器:
           ├─ internal.NewParser()
           ├─ internal.NewServerWithOptions(ServerOptions{...})
           │   └─ 保存 host, port, boundingBox, browser, enableReload, strictPort, recursive, parser
           └─ server.Serve(file)
               ├─ internal.resolveServeTarget(file)  // 判定单文件/目录模式
               ├─ hotreload.New(rootDir, recursive)   // (若 enableReload) 启动 fsnotify 监控
               ├─ handler = s.newHandlerForTarget(target)
               │   └─ 注册路由:
               │       /static/*          → defaults.StaticFiles (embed.FS)
               │       /api/edit/*        → handleSave (POST 保存编辑)
               │       /api/raw/*         → handleRaw (GET 获取原始内容)
               │       /export            → handleExport (HTML 下载)
               │       /pdf               → handlePDF (PDF 下载)
               │       /*                 → 主处理: .md 渲染 / 静态文件 / 目录列表
               ├─ handler = reloadMiddleware.Handle(handler)  // (若 enableReload) 注入 WS 脚本
               ├─ listenOnPort(port, strictPort)              // 端口监听 (默认回退/严格模式)
               ├─ internal.Open(addr)                         // (若 browser) 打开浏览器
               └─ http.Serve(listener, handler)               // 阻塞服务
```

## 四、渲染管线

```
Markdown 原始文本 ([]byte)
  │
  ├─ 1. text.NewReader(input)                    // goldmark 输入封装
  │
  ├─ 2. md.Parser().Parse(reader)                // 生成 goldmark AST
  │     ├─ 扩展链 (按序):
  │     │   Linkify, Table, Strikethrough       (goldmark 内置)
  │     │   footnote.Footnote                   (自定义)
  │     │   tasklist.TaskList                   (自定义)
  │     │   emoji.Emoji                         (第三方)
  │     │   hashtag.Extender                    (第三方)
  │     │   alert.New()                         (自定义)
  │     │   highlighting.Highlighting           (自定义, 替换代码块 Renderer)
  │     │   mermaid.Extender                    (第三方)
  │     │   mathjax.MathJax                     (自定义)
  │     │   ghissue.New()                       (自定义)
  │     │   details.New()                       (自定义)
  │     └─ parser.WithAutoHeadingID()           // 自动生成标题 id 属性
  │
  ├─ 3. promoteCJKStrongEmphasis(doc, input)    // AST 后处理: CJK 粗体修复
  │
  ├─ 4. collectTOC(doc, input) → []TOCEntry     // 遍历 AST 提取目录
  │     └─ TOCEntry { Level int, Text string, ID string }
  │
  ├─ 5. md.Renderer().Render(&buf, input, doc)  // AST → HTML 字节
  │     └─ html.WithUnsafe()                    // 允许原始 HTML 通过
  │
  └─ 6. RenderedDocument { Content string, TOC []TOCEntry }
       │
       └─ 传入 serveTemplate → layout.html 模板填充
```

## 五、核心类型索引

### 5.1 入口与 CLI (`cmd/root.go`)

| 类型/变量 | 说明 |
|---|---|
| `rootCmd *cobra.Command` | Cobra 根命令，定义 Use/Short/RunE |
| 标志 `--browser/-b` | `bool`, 默认 `true`，启动后打开浏览器 |
| 标志 `--host/-H` | `string`, 默认 `"localhost"`，绑定地址 |
| 标志 `--port/-p` | `int`, 默认 `6419`，监听端口 |
| 标志 `--bounding-box` | `bool`, 默认 `true`，页面调试边框 |
| 标志 `--no-reload` | `bool`, 默认 `false`，禁用热重载 |
| 标志 `--recursive/-r` | `bool`, 默认 `false`，递归扫描子目录 |
| 标志 `--export` | `string`, 默认 `""`，导出 HTML 文件路径 |
| 标志 `--output` | `string`, 默认 `""`，导出输出路径 |

### 5.2 服务层 (`internal/server.go`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `Server` | `parser *Parser` | 无状态 Markdown 解析器 |
| | `boundingBox bool` | 是否添加调试边框 |
| | `host string` | 绑定地址 |
| | `port int` | 端口号 |
| | `browser bool` | 是否打开浏览器 |
| | `enableReload bool` | 是否启用热重载 |
| | `strictPort bool` | 端口是否严格模式 |
| | `recursive bool` | 是否递归目录 |
| | `rootDir string` | 服务根目录 |
| | `pdfGenOnce sync.Once` | PDF 生成器懒初始化 |
| | `pdfGen *PDFGenerator` | 无头 Chrome PDF 生成器实例 |
| `ServerOptions` | (同上对应字段) | 服务器配置结构体 |
| `htmlStruct` | `Content template.HTML` | 渲染后的 Markdown HTML |
| | `BoundingBox bool` | 边框标志 |
| | `CssCodeLight template.CSS` | Chroma light 主题 CSS |
| | `CssCodeDark template.CSS` | Chroma dark 主题 CSS |
| | `ShowSidebar bool` | 是否显示文章侧边栏 |
| | `SidebarTitle string` | 侧边栏标题（目录名） |
| | `Articles []Article` | 侧边栏文章列表 |
| | `PreviousArticle Article` | 上一篇导航 |
| | `NextArticle Article` | 下一篇导航 |
| | `TOC []TOCEntry` | 当前文章目录 |
| | `CurrentFile string` | 当前文件路径 |
| | `RawContent string` | 原始 Markdown 文本（供编辑器） |

### 5.3 解析器 (`internal/parser.go`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `Parser` | (空结构体) | 无状态解析器 |
| `RenderedDocument` | `Content string` | 渲染后的 HTML |
| | `TOC []TOCEntry` | 目录条目列表 |
| `TOCEntry` | `Level int` | 标题层级 (1-6) |
| | `Text string` | 纯文本标题 |
| | `ID string` | HTML 锚点 ID |

### 5.4 文章管理 (`internal/articles.go`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `Article` | `Title string` | 显示名称（文件名） |
| | `Path string` | URL 路径（已编码） |
| | `Filename string` | 相对于根目录的文件路径 |
| | `Active bool` | 是否为当前文章 |
| | `IsDirectory bool` | 是否为目录节点（递归模式） |
| | `Expanded bool` | 目录是否展开（子节点含活跃项） |
| | `Children []Article` | 子文章/子目录列表 |

### 5.5 目标解析 (`internal/target.go`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `serveTarget` | `mode serveMode` | `modeSingleFile` 或 `modeDirectory` |
| | `rootDir string` | 服务根目录 |
| | `initialFile string` | 初始文件（单文件模式） |
| `serveMode` | — | `int` 类型常量: 0=SingleFile, 1=Directory |

### 5.6 导出 (`internal/export.go`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `ExportData` | `Content template.HTML` | 渲染后的 Markdown HTML |
| | `CssLight template.CSS` | GitHub light 主题 CSS |
| | `CssCodeLight template.CSS` | Chroma light 语法高亮 CSS |
| | `CssMermaid template.CSS` | Mermaid 图表 CSS |
| | `CssMathJax template.CSS` | MathJax 公式 CSS |
| | `CssClipboard template.CSS` | 剪贴板复制按钮 CSS |
| | `IncludeJS bool` | 是否包含 Mermaid/MathJax JS |

### 5.7 PDF (`internal/pdf.go`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `PDFData` | `Content template.HTML` | 渲染后的 Markdown HTML |
| | `CssLight template.CSS` | GitHub light 主题 CSS |
| | `CssPrint template.CSS` | 剥离包装的打印 CSS |
| | `CssCodeLight template.CSS` | Chroma light 语法高亮 CSS |
| | `CssMermaid template.CSS` | Mermaid CSS |
| | `CssMathJax template.CSS` | MathJax CSS |
| `PDFGenerator` | (chromedp 上下文管理) | 无头 Chrome 实例池 |

### 5.8 热重载 (`internal/hotreload/`)

| 类型 | 字段 | 说明 |
|---|---|---|
| `Reloader` | `rootDir string` | 监控根目录 |
| | `recursive bool` | 是否递归（对齐 `--recursive`，非递归时只监听根目录） |
| | `endpoint string` | WebSocket 端点: `"/reload_ws"` |
| | `errorLog *log.Logger` | 错误日志 |
| | `Upgrader websocket.Upgrader` | WebSocket 升级器（可配置 CheckOrigin） |
| | `clients map[*client]bool` | 已连接客户端集合 |
| | `maxDirs int` | 目录监听上限（`maxWatchedDirs` = 4096），超出即降级并告警 |
| | `ready chan struct{}` | 初始监听集合注册完成后关闭 |
| | `done chan struct{}` + `stopOnce sync.Once` | `Stop()` 结束 watch 循环并释放描述符 |
| `client` | `conn *websocket.Conn` | WebSocket 连接 |

监听范围（`docDirs`）：
- 非递归 → 仅根目录。
- 递归 → 含 `.md` 的目录 + 其祖先目录；跳过 `ignoredDirs`（`node_modules`/`.git`/`dist`/`.venv` 等依赖与构建目录），运行期新建的这类目录同样跳过。扫描根始终入选，因此「启动时为空、之后才出现首个 `.md`」也能被捕获。
- 该策略把描述符开销从「仓库规模」降到「文档规模」——macOS kqueue 下每目录、每目录内每个条目各占一个 fd，全树监听必然撞上 `too many open files`。注意单个目录（尤其 kqueue 下的根目录）自身条目就可能撑满上限，故不保证绝不耗尽。
- `addDirectories` / `adopt` 在 fd 耗尽或超过 `maxDirs` 时记录日志并降级（已注册目录继续工作），而非终止整个 watcher；fd 耗尽时调用 `watch.Remove(dir)` 回滚该目录已建立的 kqueue 监听——fsnotify 的 kqueue 后端在 `Add` 失败时不会自行回滚，残留描述符会让进程一直贴着上限。
- `fd_unix.go` 的 `isFdExhausted()` 区分 EMFILE/ENFILE（`fd_other.go` 为非 unix 空实现）；软限制由 Go 运行时在 `syscall.init` 自行提升，本项目不再调用 `Setrlimit`。
- 新建目录时先注册、后扫描，并广播扫描到的首个 `.md`（fsnotify 会把注册时已存在的文件标记为已见而不发事件）。顺序不可颠倒：先扫描后注册会丢掉「扫描到注册之间」落盘的文档。

## 六、HTTP 路由表

| 方法 | 路径 | 处理函数 | 说明 |
|---|---|---|---|
| GET | `/static/*` | `http.FileServer(defaults.StaticFiles)` | 嵌入式静态资源 |
| POST | `/api/edit/*` | `handleSave` | 保存编辑的 Markdown 内容（原子写入） |
| GET | `/api/raw/*` | `handleRaw` | 获取原始 Markdown 内容（供编辑器） |
| GET | `/export?file=...` | `handleExport` | 下载独立 HTML 文件 |
| GET | `/pdf?file=...` | `handlePDF` | 下载 PDF 文件 |
| GET | `/*` | 主路由（见下） | — |

主路由 `/*` 逻辑：
1. 路径为空 → 目录模式重定向到初始文章；无文章则渲染空状态
2. 路径为 `.md` 文件 → 渲染 Markdown 页面
3. 路径为目录 → 清除缓存验证头 → 交由 `http.FileServer` 处理
4. 单文件模式下访问非目标文件 → 404

## 七、两种服务模式

| 特性 | 单文件模式 (`go-grip file.md`) | 目录模式 (`go-grip` / `go-grip .` / `go-grip docs`) |
|---|---|---|
| 侧边栏 | 隐藏 | 显示所有 `.md` 文件 |
| URL 访问范围 | 仅该文件（其他 .md 返回 404） | 目录下所有 `.md`（不允许跨越根目录） |
| 初始页面 | 直接渲染该文件 | 重定向到 README.md 或第一个文章 |
| 上一篇/下一篇 | 无 | 按排序导航（支持键盘 ← →） |
| 递归子目录 | N/A | 可选 (`--recursive`) |
| 侧边栏标题 | N/A | 使用目录名 |
| TOC | 页面内嵌 | 右侧独立 TOC 面板 |

## 八、Markdown 扩展体系

所有 7 个自定义扩展均实现 `goldmark.Extender` 接口，通过 `Extend(m goldmark.Markdown)` 注册。注册顺序在 `internal/parser.go:newMarkdown()` 中定义。

### 扩展注册顺序与详情

| 顺序 | 扩展 | 包路径 | 功能 | 设计模式 | 外部依赖 |
|---|---|---|---|---|---|
| 1 | Linkify | goldmark/extension | URL 自动链接 | 内置 | 无 |
| 2 | Table | goldmark/extension | GFM 表格 | 内置 | 无 |
| 3 | Strikethrough | goldmark/extension | ~~删除线~~ | 内置 | 无 |
| 4 | Footnote | `pkg/footnote` | `[^1]` 脚注 | Block Parser + Inline Parser + Transformer + Renderer | 无 |
| 5 | TaskList | `pkg/tasklist` | `- [x]` 任务列表 | Inline Parser + 2 Renderers (复用以有节点) | 无 |
| 6 | Emoji | goldmark-emoji | `:smile:` 表情短代码 | 第三方 | gomoji |
| 7 | Hashtag | goldmark/hashtag | `#tag` 主题标签 | 第三方 | 无 |
| 8 | Alert | `pkg/alert` | `> [!NOTE]` 等 5 种提示 | AST Transformer + HTML Renderer | 无 |
| 9 | Highlighting | `pkg/highlighting` | 代码语法高亮 + 复制按钮 | 替换 FencedCodeBlock Renderer | chroma/v2 |
| 10 | Mermaid | goldmark/mermaid | ` ```mermaid` 图表 | 第三方 (客户端渲染) | 无 |
| 11 | MathJax | `pkg/mathjax` | `$...$` / `$$...$$` / ` ```math` | Block Parser + Inline Parser + Transformer + 2 Renderers | 无 |
| 12 | GHIssue | `pkg/ghissue` | `#123` / `owner/repo#123` 链接 | Inline Parser + Transformer + Renderer + git remote 探测 | 无 |
| 13 | Details | `pkg/details` | `<details>` 折叠状态持久化 | Transformer (注入ID) + Renderer (注入JS) | 无 |

### 通用扩展架构模式

每个 `pkg/*` 扩展遵循一致的分层：

```
pkg/<name>/
├── <name>.go        # Extender 实现 + goldmark.New(WithExtensions(...)) 注册入口
├── ast.go           # (可选) 自定义 AST 节点类型定义
├── transformer.go   # AST Transformer: 解析后遍历/修改 AST
├── renderer.go      # HTML Renderer: 将 AST 节点渲染为 HTML
└── <name>_test.go   # testify 测试
```

- **Parser** 阶段：识别标记语法，插入自定义 AST 节点
- **Transformer** 阶段：遍历 AST，做语义转换（如注入 ID、收集信息）
- **Renderer** 阶段：将 AST 节点输出为 HTML 字符串

## 九、热重载机制

```
fsnotify.Watcher (文件系统事件监控)
  │
  ├─ 递归监控 rootDir 及其子目录（仅目录注册到 watcher）
  ├─ 事件过滤:
  │   ├─ Create → 注册新目录/文件
  │   ├─ Write  → 触发重载（仅 .md 文件，100ms debounce）
  │   └─ Rename/Remove → 移除监控
  │
  ├─ debounce(100ms) → 合并短时间内的多次写入
  │
  └─ 遍历所有 WebSocket 客户端 → 发送 "reload" 消息

WebSocket 端点: /reload_ws
客户端注入: 每个 text/html 响应在 </body> 前注入 <script> 片段
  └─ 自动重连 (exponential backoff, 最大 30s)
  └─ 收到 "reload" → location.reload()
```

版本号 `wsVersion = "2"` 通过 WebSocket URL 参数传递，用于版本不匹配时强制整页刷新。

## 十、端口回退机制 (`internal/listener.go`)

```
listenOnPort(port, strictPort):
  ├─ strictPort=true (用户显式指定 --port):
  │   └─ net.Listen 失败 → 直接报错
  │
  └─ strictPort=false (默认):
      ├─ 尝试 port (6419)
      ├─ 失败 → port+1
      ├─ 失败 → port+2
      ├─ ...
      └─ 最大尝试 100 次 → 全部失败则报错
```

## 十一、导出功能

### HTML 导出 (`--export`)

```
命令行: go-grip --export README.md --output README.html
  ├─ 读取 Markdown 文件
  ├─ parser.Render() → HTML
  ├─ embedLocalImages() → 本地图片转 base64 data URI
  ├─ 读取嵌入 CSS: light, mermaid, mathjax, clipboard, chroma-light
  ├─ 填充 templates/export.html → 独立 HTML
  └─ 写入 --output 或 stdout

浏览器导出: GET /export?file=README.md
  └─ IncludeJS=true (带 Mermaid/MathJax 脚本，用于在线场景)
  └─ Content-Disposition: attachment
```

### PDF 导出 (`GET /pdf?file=...`)

```
GET /pdf?file=README.md
  ├─ 读取并渲染 Markdown
  ├─ buildPDFMarkup() → 使用 templates/print.html
  │   ├─ embedLocalImages() → 图片内联
  │   └─ stripPrintCSSWrapper() → 剥离 @media print 包装
  ├─ PDFGenerator (chromedp, 懒初始化)
  │   └─ 无头 Chrome 实例池 (大小 2)
  │   └─ page.PrintToPDF() → PDF 字节
  └─ Content-Type: application/pdf
  └─ Content-Disposition: attachment
```

## 十二、编辑功能

### 前端

- `defaults/static/js/editor.js` + `defaults/static/css/editor.css`
- 在渲染页面中提供分屏编辑界面
- 使用 `marked.min.js` 在客户端做实时预览

### 后端 API

| API | 方法 | 说明 |
|---|---|---|
| `/api/raw/{file}` | GET | 返回原始 Markdown 文本 |
| `/api/edit/{file}` | POST | 保存编辑内容 |

安全措施：
- 路径遍历防护 (`..` 检测 + `filepath.EvalSymlinks` + 前缀检查)
- 仅 `.md` 文件可编辑
- 文件大小限制 10MB
- 权限检查 (`checkWritable`)
- 原子写入 (先写 `.tmp`，再 rename)
- 按文件路径的互斥锁 (`sync.Map` + `sync.Mutex`)

## 十三、CJK 强调修复 (`internal/cjk_emphasis.go`)

这是项目中最复杂的非标准逻辑，解决 goldmark 对中日韩文字与全角标点的强调解析问题。

```
promoteCJKStrongEmphasis(doc, source):
  └─ 遍历 AST 查找 Emphasis 节点 (level >= 2, 即粗体)
      └─ 检查是否为 CJK 场景:
          ├─ 左边界: 前一个字符是 CJK/全角标点 → 提升为粗体
          └─ 右边界: 后一个字符是 CJK/全角标点 → 提升为粗体
      └─ 修改节点 level 和分隔符长度
```

## 十四、依赖总览

```
goldmark v1.7.16 (Markdown 解析核心)
├── goldmark-emoji v1.0.6 (emoji 短代码)
├── goldmark/hashtag v0.4.0 (#标签)
├── goldmark/mermaid v0.6.0 (Mermaid 图表)
├── chroma/v2 v2.14.0 (语法高亮)
│   └── regexp2 v1.11.4 (增强正则)
├── cobra v1.8.1 (CLI 框架)
│   └── pflag v1.0.5
├── gorilla/websocket v1.5.3 (WebSocket 热重载)
├── fsnotify v1.8.0 (文件系统监控)
├── bep/debounce v1.2.1 (防抖)
├── gomoji v1.3.0 (emoji 辅助)
├── chromedp v0.15.1 (无头 Chrome PDF 生成)
│   ├── cdproto
│   └── sysutil
├── uniseg v0.4.7 (Unicode 分段)
└── testify v1.11.1 (测试)
```

## 十五、关键设计决策

1. **无状态 Parser** — `Parser` 是空结构体，所有方法无副作用，天然并发安全
2. **//go:embed 嵌入** — 所有模板和静态资源编译进二进制，运行时零外部依赖
3. **端口回退** — 默认端口被占用时自动递增，显式 `--port` 时严格报错
4. **CJK AST 修复** — 在解析后直接修改 AST 节点属性，而非前置文本替换
5. **文件名即标题** — 侧边栏显示文件名，避免解析所有文件提取 h1 的性能开销
6. **目录排序规则** — 目录优先于文件 → README 优先 → 不区分大小写按标题升序
7. **原子写入** — 编辑保存通过 tmp 文件 + rename 实现，避免写一半的脏数据
8. **WebSocket 热重载** — 中间件模式注入脚本，版本号机制防止新旧客户端不兼容
9. **PDF 懒初始化** — `sync.Once` 保证 chromedp 实例池只创建一次
10. **图片内联** — 导出时将本地图片转 base64 data URI，HTTP/HTTPS/data: 图片保持不变

## 十六、构建与发布

### 本地构建

```bash
go build -o bin/go-grip main.go    # 标准构建
go install .                        # 安装到 $GOPATH/bin
nix build                           # Nix Flakes
mise run build                      # mise 任务
```

### 测试

```bash
go test ./...                       # 运行所有测试
```

### CI (`build.yml`)

触发: push / PR

1. mise 设置 Go 环境
2. `go build`
3. `go test ./...`
4. `gofmt -d .` 格式检查
5. `golangci-lint` 静态分析

### CD (`release.yml`)

触发: 推送 `v*` tag

1. `go test ./...`
2. 6 架构交叉编译: darwin/linux/windows × amd64/arm64
3. `-trimpath -ldflags="-s -w"` 精简二进制
4. `.tar.gz`/`.zip` + `checksums.txt`
5. 创建 GitHub Release

## 十七、扩展指南

### 添加新的 Markdown 扩展

1. 在 `pkg/` 下创建新包，实现 `goldmark.Extender` 接口
2. 在 `internal/parser.go:newMarkdown()` 中按需注册
3. 编写 testify 测试 (`_test.go`)
4. 如有 CSS/JS，放在包目录下，通过 `defaults/embed.go` 嵌入

### 添加新的 CLI 标志

1. 在 `cmd/root.go:init()` 中注册 `rootCmd.Flags().XXX()`
2. 在 `ServerOptions` 中添加对应字段
3. 在 `Server` 中实现对应行为

### 添加新的 HTML 模板变量

1. 在 `htmlStruct` 中添加字段
2. 在 `newPageData()` 中填充值
3. 在 `templates/layout.html` 中使用 `{{.FieldName}}`

## 十八、已知改进方向

1. 侧边栏标题可从文件首个 `h1` 推导（当前使用文件名）
2. 隐藏文件策略（`.` 开头的文件和目录）
3. 环境变量配置支持（当前仅 CLI 标志）
4. 大型文档/目录性能基准测试
5. Docker 容器化方案
6. CJK 强调修复缺少基准测试
