# go-grip Architecture Reference

> 最后更新: 2026-10-04
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
│   ├── managed.go              # 隐藏 --managed 宿主集成契约：NDJSON、ready/fatal、所有权监视与有界善后
│   ├── managed_test.go
│   ├── parser.go               # goldmark 解析器封装：渲染管线、扩展注册、TOC 提取
│   ├── parser_test.go
│   ├── listener.go             # 端口监听：显式 bind 地址 + 默认回退 / 严格模式
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
│       ├── watch_kqueue.go     # darwin/BSD: 成本=1+条目数, 预算=RLIMIT_NOFILE-保留
│       ├── watch_linux.go      # linux: 成本=1, 预算=inotify 配额的一半
│       ├── watch_other.go      # 其余平台: 成本=1, 保守预算
│       ├── watch_budget.go     # 预算换算
│       ├── fd_unix.go          # EMFILE/ENFILE/ENOSPC 判定
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
       ├─ --managed 已设置（隐藏宿主集成）？ → internal.RunManaged(ManagedOptions{...})
       │   ├─ 校验代次与不适用组合（--export/--output/--json/--host/--port）→ fatal + 退出
       │   ├─ 建立 stdin 所有权监视（先于目标 I/O）+ 2s 善后 watchdog
       │   └─ managed serve:
       │       ├─ resolveServeTarget → verifyTargetAccess（单文件常规可读 / 目录根可读）→ hotreload.NewWithReporter（状态出口先于扫描安装）→ handler + /__gogrip/ready 路由
       │       ├─ listenOn("127.0.0.1", 0, true)      // 真实回环 + OS 分配端口
       │       ├─ stdout 发布 ready（v1 NDJSON 首帧：实际 URL + reload 快照与已知原因）
       │       ├─ 运行期真实状态变化 → reload-status / target-status（仅变化时）
       │       └─ http.Serve；owner loss → 有界关闭 + watcher 等待 + WS 关闭 + 已初始化 PDF 释放
       │
       ├─ --export 非空？（仅独立 CLI）
       │   ├─ internal.NewParser()
       │   ├─ parser.Render(data)           // Markdown → HTML + TOC
       │   ├─ internal.BuildExportHTML()    // 填充 export.html 模板 + 图片内联
       │   └─ 写入文件或 stdout → 退出
       │
       └─ 否则启动独立服务器:
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
               ├─ listenOn("", port, strictPort)              // 通配绑定 (默认回退/严格模式)
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
| 标志 `--json` | `bool`, 默认 `false`，启动时以 JSON 输出服务信息（独立 CLI） |
| 标志 `--managed` | `string`, 默认 `""`，隐藏；macOS 宿主集成模式，值为启动代次（见第七节） |

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
| | `targetStatus func(state, reason string)` | managed 目标可访问性出口；CLI 为 nil，不分配状态通道 |
| | `targetState string` + `targetStatusMu` | 真实访问观察到的 `available` / `unavailable`（仅状态变化时报告） |
| `ServerOptions` | (同上对应字段) | 服务器配置结构体；`TargetStatusReporter` 仅在 managed 路径设置 |
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

目录发现（`discoverArticles` / `discoverArticlesInDir`）递归读取目录：子目录读取失败且属权限类错误（`errors.Is(err, fs.ErrPermission)`，含 EACCES/EPERM）时跳过该子树、保留可读兄弟；不可读子树不进入侧边栏或初始文章，但可访问根仍正常服务可读内容，watcher 的 walk error 按既有覆盖降级机制报告。根目录自身的读取错误在进入递归前返回，因此所选根或所选单文件不可访问仍是 fatal `target-unavailable`（见“服务模式”）；其他错误类别保持原传播，不定义通用扫描容错。

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
| | `maxFDs int` | 监听预算（按平台资源计：kqueue 下为描述符、Linux 下为 inotify watch 配额） |
| | `watchCost int` | 已消耗预算，由 watch goroutine 独占 |
| | `incomplete` + `incompleteReason` | watch goroutine 独占：覆盖不完整的标记与首个原因 |
| | `report StateReporter` | 可选覆盖状态出口；`NewWithReporter` 在初始扫描前安装，ready 之前的变化由快照携带 |
| | `openSource func() (*watchIO, error)` + `watchIO` | watcher 打开 seam（默认 `fsnotify.NewWatcher` 包装），供测试确定性注入创建/注册/运行期失败 |
| | `state State` + `reason string` + `stateMu` | 当前覆盖快照与原因；`Status()` 返回两者 |
| | `ready chan struct{}` | 初始监听集合注册完成后关闭 |
| | `stopped chan struct{}` | watch goroutine 退出后关闭；`Stop()` 等待它 |
| | `done chan struct{}` + `stopOnce sync.Once` | `stop()` 结束 watch 循环 |
| `client` | `conn *websocket.Conn` | WebSocket 连接 |

`Stop()` 结束 watch 循环、等待 `stopped`（watcher 与描述符已释放）并关闭全部 WebSocket 客户端；可重复调用。启动失败（watcher 创建、walk、预算截断、注册失败）在初始注册后以 `degraded` + 首个原因进入快照/事件；运行期 watcher 错误或新增目录覆盖不足令 `active` 一次性迁移到 `degraded`（首个原因）；`State` 不再单独存在，调用方使用 `Status()`。这些出口只报告真实状态变化，不解析日志、不增加轮询。

监听范围（`docPlans`）：
- 非递归 → 仅根目录。
- 递归 → 含 `.md` 的目录 + 其祖先目录；跳过 `ignoredDirs`（`node_modules`/`.git`/`dist`/`.venv` 等依赖与构建目录），运行期新建的这类目录同样跳过。扫描根始终入选，因此「启动时为空、之后才出现首个 `.md`」也能被捕获。
- 预算按**真实成本**计，而非目录数量：`dirCost(entries)` 在 kqueue 平台返回 `1 + 条目数`（fsnotify 为目录本身及目录内每个条目各开一个描述符），在 Linux 返回 `1`（inotify 每目录一个 watch）。`watchBudget()` 取平台额度：kqueue 用 `RLIMIT_NOFILE` 软上限减 `fdReserve`(1024)，Linux 用 `/proc/sys/fs/inotify/max_user_watches` 的一半；两者不可读时退回 `fallbackBudget`。
- 这是必需的：一个含 648 个文件的目录在 kqueue 下要 649 个描述符，仅按目录数设限会误判容量（曾因此出现「8694 个目录超过 4096 上限」的截断）。
- `registerDirs` 在预算耗尽或平台资源报错（EMFILE/ENFILE/ENOSPC）时记录日志并降级：已注册目录继续工作，不再终止整个 watcher。
- fd 耗尽时调用 `watch.Remove(dir)` 回滚该目录已建立的 kqueue 监听——fsnotify 的 kqueue 后端在 `Add` 失败时不会自行回滚，残留描述符会让进程一直贴着上限。
- `fd_unix.go` 的 `isFdExhausted()` 区分 EMFILE/ENFILE/ENOSPC（`fd_other.go` 为非 unix 空实现）；软限制由 Go 运行时在 `syscall.init` 自行提升，本项目不调用 `Setrlimit`。
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

## 七、服务模式

| 特性 | 单文件模式 (`go-grip file.md`) | 目录模式 (`go-grip` / `go-grip .` / `go-grip docs`) |
|---|---|---|
| 侧边栏 | 隐藏 | 显示所有 `.md` 文件 |
| URL 访问范围 | 仅该文件（其他 .md 返回 404） | 目录下所有 `.md`（不允许跨越根目录） |
| 初始页面 | 直接渲染该文件 | 重定向到 README.md 或第一个文章 |
| 上一篇/下一篇 | 无 | 按排序导航（支持键盘 ← →） |
| 递归子目录 | N/A | 可选 (`--recursive`) |
| 侧边栏标题 | N/A | 使用目录名 |
| TOC | 页面内嵌 | 右侧独立 TOC 面板 |

### managed 宿主集成模式（隐藏 `--managed <generation>`）

`internal/managed.go` 为 macOS 宿主提供稳定机器契约；独立 CLI 行为不受影响。managed 模式：

- 在目标解析、卷读取、watcher 与 listener 创建 **之前** 建立 stdin 所有权监视。宿主（或测试中的临时拥有者）仅持有写端：正常停止关闭写端，崩溃/强制终止由内核关闭，两者都表现为 EOF/读错误 → owner loss。单文件目标在 ready 前必须是可读的常规文件（FIFO/设备等非常规目标 fatal），目录访问由解析初始 URL 的文章扫描验证：可读根中的权限拒绝子树按 §5.4 跳过，不使整个会话失败。
- 使用 `listenOn("127.0.0.1", 0, true)` 真实绑定回环并由 OS 分配端口，URL 取实际端口；不使用默认端口拼接，不接受 `--host`/`--port`/`--export`/`--output`/`--json`（fatal 拒绝，不读取目标），也不由 Go 打开浏览器。
- stdout 专用于 UTF-8 v1 NDJSON，单序列化 writer（并发事件不拼接帧）；日志/诊断走 stderr。
- 目标可访问、handler 就绪、listener 绑定后，仅发布一次 `ready`，且它是协议的第一帧：

| 事件 | 字段 |
|---|---|
| `ready` | `version=1`, `generation`, `url`（完整实际 URL，含转义后的单文件路径）, `reload.state`；degraded 时带可获得 `reload.reason` |
| `reload-status` | `reload.state` = `pending` / `active` / `degraded`，degraded 带 `reload.reason`；只在 ready 之后、真实覆盖状态发生变化时报告一次（同一 degraded 保留首个原因，重复错误不再发事件） |
| `target-status` | `target.state` = `available` / `unavailable`，unavailable 带实际访问失败 `target.reason`；只在状态变化时报告 |
| `fatal` | `code`（`target-unavailable` / `listen-failed` / `serve-failed` / `inapplicable-flags` / `invalid-generation`）, `message` |

- 状态出口在 watcher 初始扫描前安装；ready 之前观察到的状态变化折叠进 ready 快照（发布快照与抑制共用同一把锁，不会既丢快照又丢事件），因此 ready 永远是首帧，且不会出现永久 pending 或假 active。`reload.state` 为真实覆盖：`pending`（初始扫描未完成）、`active`（watch 集完整）、`degraded`（watcher 创建失败、walk 错误、预算截断、注册失败或运行期 watcher 错误/新增目录覆盖不足）、`disabled`（`--no-reload`）。运行期错误只写日志与状态出口，不逐条驱动 UI。
- `target-status` 只来自实际访问：目录目标按根目录级可访问性判定（根读取失败才是 unavailable），单文件目标按所选文件可读常规文件判定；目录内单篇文章或无关资源（图片、不存在子页面）的失败保持原 HTTP 错误，不翻转目标状态，也不显示为空目录；可访问根中的权限拒绝子树同样不使根 unavailable：发现跳过该子树（§5.4），其缺少覆盖经 `reload-status degraded` 与可获得原因报告，可读内容及手动刷新保持可用。原路径再次访问成功可报告 `available`，但不后台寻找新位置、不轮询、不重连卷、不自动重启或恢复 watcher。
- managed 专属固定路由 `HEAD/GET /__gogrip/ready` 返回 204 + `X-GoGrip-Generation`，不访问目标、不渲染正文，供宿主启动确认。
- owner loss 后取消运行，并在独立 watchdog 的 2 秒善后期限后强制退出（绕过 defer），不等阻塞的文件系统操作或启动函数返回；正常取消路径依次执行有界 HTTP 关闭（1s Shutdown 后必要时 Close）、`hotreload.Stop()`（等待 watch 循环结束并关闭 WebSocket 客户端）、仅释放已初始化的 PDF 资源（停止不会创建 headless Chrome）、释放 listener。
- 已验证（任务 1.1/1.3 补正 14，2026-10-06，真实二进制）：可访问根含 `chmod 000` 子目录时，managed ready 为实际回环 URL（PID 37458，`127.0.0.1:54531`），根与嵌套文档 HTTP 200；ready 快照 `pending` 后经 `reload-status` 报 `degraded` 且 reason 为 walk error（stderr 同步记录），全程无 `target-status unavailable`；改写文档后同一服务手动 GET 得到新内容；关闭 ownership writer 后 exit 0、端口关闭；独立 CLI（PID 37463，`*:54538`，通配监听为既有政策）与另一 managed 会话不受影响。根目录 `chmod 000` 仍 fatal `target-unavailable`、无 ready（`internal/managed_test.go` 回归）。scoped TDD：`cmd` 真实进程回归先红（fatal `resolve initial preview path: … permission denied`）后绿；`go test ./... -count=1`、`go test -race ./internal/... ./cmd/...`（3 轮）、`go vet ./...`、`gofmt -l .` 通过；本机无 `golangci-lint`（已装 staticcheck 二进制因 go1.24 无法加载本模块），既有 `internal/server_test.go:339-340` errcheck 缺口未在本票触碰。临时 fixture 权限还原后删除。
- 限制：强制退出依赖内核回收 listener/watcher 描述符；不管理 Chrome 等任意后代进程；不对不可中断的内核 I/O 承诺精确硬截止。宿主侧的 descriptor 所有权与停止见下节。

### Swift 宿主进程适配器（macOS 端）

`macos/GoGrip/Services/ManagedProtocol.swift`（v1 解码/校验）与 `ManagedProcess.swift`（进程适配器）是生产 Foundation 启动路径：一次 launch 对应一个生成代次与一个确切 owned child，不另建会话字典，也不复用旧管理器的端口/回退逻辑。AppKit 应用根、面板、默认浏览器与退出准入已由开发 App 接通用（见下节“Swift 应用根”）；Finder Services 入口已由 07 接通；TCC/受保护目录与卷访问的实际授权主体、拒绝指导及卷不可访问行为已由 08 接通并验证（同一节 08 记录），候选包仍属后续任务。

- 启动：`Process.executableURL` + 参数数组 `--managed <generation>`（目录附 `-r`，目标前 `--`），不经 shell；独立 CLI 的参数、端口与 `--json` 政策不受影响。
- 每个 child 有独立 stdin Pipe；在进程级串行的 launch 临界区内创建 descriptor，并对宿主持有的非 stdio 端（stdin 写端、stdout/stderr 读端）设置并复核 `FD_CLOEXEC`；writer 不复制、不传给其他进程。
- 连续消费 UTF-8 v1 NDJSON：跨读取分片、跨多字节 UTF-8 字符与同次读取多帧；校验 `version=1`、`generation`、必要字段及嵌套 `reload`/`target` 形状，未知事件或非法状态直接判失败（不做降级解析）。stdout 未完成帧残余与单个完整帧、stderr 诊断留存各上限 64 KiB（保留最新尾部，失败时的诊断尾部为尽力而为），诊断不参与机器状态判断；ready 之后继续消费 reload/target/fatal/退出，运行期解码/校验违规以带代次事件报告并终止该 child，不静默停读。
- 启动自创建 child 起 15 秒内部期限，覆盖机器反馈与一次 same-origin HEAD（`/__gogrip/ready`）：仅当 URL 为 `http://127.0.0.1:<实际端口>`、HEAD 返回 204、`X-GoGrip-Generation` 匹配、响应 origin 未跳转且进程仍存活时才成功。无效/缺失反馈、EOF、超时、早退或 HEAD 失败均报告失败（不猜 6419）并收尾本次 child。
- 停止：关闭该 child 的专用 writer 并等待真实退出；4 秒后仍存活的确切 owned pid 使用 SIGKILL，结果携带实际退出状态。会话协调器（下节）在其上提供单个/全部、启动中停止与重复停止语义。
- 通道释放：宿主持有的三个通道端在不依赖实例释放的情况下关闭——stdin writer 在停止或 child 真实退出时关闭；stdout/stderr 读端在读到 EOF 后立即关闭；未产生 child 的启动失败（`Process.run` 抛错或 CLOEXEC 检查失败）显式释放本次全部 Pipe。真实 Go 启动失败路径按“保留每个失败实例”统计本进程 `/dev/fd` FIFO 计数，验证失败启动不累积通道 descriptor。
- 已用同一生产 launch 路径验证（真实 Go，任务 2.1/2.5）：目录、中文+空格单文件与空目录的实际 URL/内容；Go fatal 与适配器判定失败的清理；双会话停止只影响对应 child；独立 CLI 隔离；临时 Foundation 拥有者在 ready 前（spawned 回调内自 SIGKILL，确定性先于 readiness 校验）与运行中宿主 SIGKILL 后 child 退出、端口释放；双会话宿主 SIGKILL 后两个 owned child 均退出；`ps` 记录的 argv 与 `lsof` 记录的真实监听。
- 未验证/限制：已挂载网络卷的实际访问（任务 3.3 该场景本轮无共享，保持未完成）及候选 DMG/universal/支持环境（任务 6–7）仍未接入（2026-10-06 更新：候选打包已由任务 12 交付，见第十六节；支持环境实际矩阵仍属任务 13）；TCC 实际主体与受保护/外接卷访问已由 08 在本机开发构建验证（见“Swift 应用根”的 08 记录）；Finder Services 已由 07 的开发 App 接通（见“Swift 应用根”）。AppKit 应用根与确定内置路径（`Contents/MacOS/go-grip`）已由 05 的开发 App 接通；未在 Intel 或 macOS 13 实际运行；不对不可中断内核 I/O 承诺硬截止，不管理 Chrome 等后代进程。构建/测试入口 `make macos-test` 将 Go 工具复制到系统卷临时 bundle 后运行 xctest——仓库位于外置卷时 xctest 进程无法打开该卷文件（open 阻塞），且 xcodebuild 不转发自定义环境变量。

### Swift 会话协调器（macOS 端）

`macos/GoGrip/Services/PreviewSessionCoordinator.swift`（唯一主 actor 会话事实源）、`TargetPreparation.swift`（目标身份准备）与 `ManagedProcessLaunching.swift`（生产适配器 seam 与 `FoundationManagedProcessFactory`）组成会话协调层。它不新建第二份进程或会话模型：适配器只按代次报告事件，入口与面板读取协调器，不自行维护运行事实。

- 目标准备在非隔离 async 中执行（文件元数据与符号链接解析可阻塞）：标准化绝对 file URL、解析符号链接、按实际类型分类；身份使用解析后的路径，用户所选路径保留为展示路径；`.md` 扩展名（按解析后路径，已获裁决）大小写不敏感且只接受常规文件（目录、FIFO 等其他类型按实际类型处理）；不支持文件与元数据失败返回目标与可获得原因，不回退父目录、不创建假 running、不使用 inode/bookmark 或路径小写化。Go 实际访问仍是最终判定。
- 批次准备（任务 2.3）：`prepareBatch(_:)` 对全部所选输入按选择顺序在后台逐个准备，只保留首次出现的规范化身份（别名/符号链接归一，展示路径取首次选择），把不支持/准备失败按目标与原因收集为失败项；此阶段不建立记录、不启动 child、不请求浏览器。执行阶段由 `open(prepared:)` 复用同一条记录/代次/单飞/停止路径（`open(_:)` 即“准备后调用它”），批次因此不重复元数据准备、也不产生第二份身份或运行模型。
- 每个身份一条会话记录，执行阶段为 `starting → running → stopping → terminated`；每次实际新启动分配 generation，事件（spawned/reload/target/fatal/违规/退出）按身份 + generation 校验后应用，旧代次回调不能覆盖新代次。启动成功只在同代次仍处于 starting、未请求停止且子进程仍在运行时提交；无有效 URL 或已退出不会成为成功会话。
- 同目标并发请求加入同一次启动（单飞）；running 复用同一实际 URL；停止期间的重开等待旧子进程真实结束后才建立新 generation。父目录、子目录与文件是不同身份，各自拥有独立会话。
- 停止：经适配器关闭该 child 的 ownership writer 并等待真实退出，仍存活的确切 owned child 沿用 4 秒 SIGKILL 兜底；重复停止共享同一次真实收尾；停止失败保留实际失败与管理入口（可重试），不以删行伪装退出；停止未确认期间的重开只重试收尾或报告失败、不启动重叠 generation；不自动重启、不恢复 watcher、不迁移路径、不影响独立 CLI。
- 显式停止不记为意外退出；运行期退出结束 running 并保留可获得原因（fatal/协议违规原因在退出前保存）。`target-status`/`reload-status` 与执行阶段分开：目标不可访问或热重载降级不删除仍运行会话的 URL 与管理入口，也不伪造空目录。
- 已用临时 Foundation 拥有者调用生产协调器 → 生产 `ManagedProcess` → 当前 Go 验证（任务 2.2/2.4）：原路径与符号链接并发复用同一会话/PID/generation；父/子/文件三个独立会话与递归边界；单个停止后对应 PID 退出且端口不再服务、其他会话继续；启动中停止（观察到 spawned 且协调器仍 starting）等待真实退出且迟到启动结果不恢复 running；真实子进程意外退出后不再 running、原因可查看且不自动重启；目标被移动后保留会话并真实报告 unavailable；Markdown 命名的 FIFO 在准备阶段按实际类型拒绝、不可读的常规 Markdown 目标得到真实 Go fatal 原因；独立 CLI 在 stop all 前后继续提供内容。
- 未验证/限制：候选包（任务 6–7）仍未接通（2026-10-06 更新：候选打包已由任务 12 交付，见第十六节）；登录项已由任务 5.3 接通（见“Swift 应用根”的 5.3 记录）；首次引导与完整简中/英文本地化已由任务 5.1/5.2 接通（见“Swift 应用根”的 10 记录）；最近记录/完整面板已由任务 4.1–4.3 接通（见“Swift 应用根”记录）；TCC/受保护与卷访问已由 08 接通（见“Swift 应用根”的 08 记录）；Finder 入口已由 07 接通并复用本协调器。元数据准备尚未完成、还未成为会话的请求不在 `stopAll` 取消范围内——退出路径的“禁止新启动”由 05 的准入关闭（`stopAcceptingNewOpens`，含准备完成后才到达的请求）处理，批次先 `prepareBatch` 再 `open(prepared:)` 也不改变这一点；不管理 Chrome 后代；未在 Intel 或 macOS 13 实际运行。

### Swift 应用根、面板与真实退出（macOS 端）

`macos/GoGrip/GoGripApp.swift`（AppKit 入口）、`AppDelegate.swift`（强持有根）、`FinderServiceProvider.swift`（唯一 Finder Services provider）、`PreviewAppModel.swift`（浏览器/退出协调）、`WorkspaceBrowserOpening.swift`（NSWorkspace seam）与 `Views/PopoverView.swift` 组成开发 App 的原生层。SwiftUI 只渲染弹窗内容；应用没有 SwiftUI Scene、空 Settings 或独立管理主窗口，主 plist 设 `LSUIElement`。

- 入口与所有权：`@main` enum 直接创建 `NSApplication`、强持有 `AppDelegate` 并 `run`；`applicationDidFinishLaunching` 依次创建唯一生产协调器与批次入口（`PreviewAppModel`）、唯一 `FinderServiceProvider` 并赋值 `NSApp.servicesProvider`，随后才建立 status item/popover。关闭面板只影响显示，不停止会话；本层不新增第二份运行事实源。注册顺序是有意的：系统可能在 provider 注册后立即投递首个服务请求（早于首次面板展示），因此注册时协调器与批次依赖必须已可用，且请求处理不依赖面板可见。
- Finder Services 入口（任务 3.1/3.2/3.4 的 Services 贡献）：`Info.plist` 仅一项 `NSServices` 声明——`NSMessage=openWithGoGrip`、`NSPortName=GoGrip`、`NSMenuItem.default=Open with GoGrip`、`NSSendTypes=[public.file-url]`、`NSRequiredContext.NSApplicationIdentifier=com.apple.finder`；不设返回类型或快捷键、不写回 pasteboard、不解析文本/shell、不保留旧 Finder Sync 或专用 URL 转发。`openWithGoGrip:userData:error:` 用 `readObjects(forClasses:[NSURL.self], options:[.urlReadingFileURLsOnly:true])` 在返回前同步取得本次全部 file URL（不只 first），随即把 URL 值（而非 pasteboard）交给同一 `PreviewAppModel.openBatch` 并立即返回；`error` 指针只在返回前用于“请求中没有任何可读文件”这类即时错误，目标/启动/浏览器错误仍归批次的一次原生报告，异步阶段不重读 pasteboard。去重、5/6 数量确认、部分失败继续与汇总、默认浏览器、退出准入全部复用 05/06 的同一路径，Services 只增加输入入口、不建立第二套事实源或启动入口。
- 内置工具：宿主只从 `Bundle.main.bundleURL/Contents/MacOS/go-grip` 加载；开发构建脚本 `macos/Scripts/build-go-grip.sh` 构建双架构工具到该确定位置并 ad-hoc 签名（独立 code identifier `com.showgp.GoGrip.go-grip`），Xcode 再 ad-hoc 签宿主 bundle。运行期不回退 Resources/PATH/源码，不 runtime chmod。该开发构建是开发/验收入口，不代表 universal/签名公证/候选 DMG 或支持环境证明（任务 6–7；2026-10-06 更新：候选打包与静态架构检查已由任务 12 完成，见第十六节）。
- 批次入口（任务 2.3）：面板 `Open…` 用 `NSOpenPanel`（目录/`.md` 多选）把全部所选 URL 交给唯一生产批次入口 `PreviewAppModel.openBatch`。该入口先经协调器 `prepareBatch` 在后台准备并按规范化身份去重、收集不可打开项；去重后的不同有效目标超过 5 个时，才用原生确认（`AppAlertPresenter.confirmOpening`，显示实际数量、非模态并以 continuation 等待响应）询问，用户取消或退出准入已关闭时不建立任何记录、子进程或浏览器请求，也不停止既有会话。获准后按选择顺序逐项 `await` 协调器 `open(prepared:)`（同一记录/代次/单飞/停止机制，不重复准备、不建立批内并发或第二份运行模型），仅对已验证 running 会话请求 `NSWorkspace` 打开实际 URL；单项失败不阻断后项。
- 报告与恢复（任务 2.3）：批次结束时把不支持、准备/启动及浏览器失败汇总成一份 `lastOperationFailures`（目标 + 可获得原因）。该报告由**应用根**在每次非空发布时用一次原生 NSAlert 呈现（经一次主队列转派交共享的单槽 `AppAlertPresenter`），与面板是否可见无关——冷启动的 Services 请求在面板从未展示时也能立即感知失败；面板会话列表下方保留同一份可查看清单。失败报告与数量确认共用该槽：同一时刻最多一个提示，失败报告按到达顺序各呈现一次，提示为非模态窗口而不进入 app-modal 会话（见下方 2026-10-06 修正记录）。全部成功不发布报告也不额外提示，报告不改变会话阶段/URL/PID/管理入口。单项操作（面板重开浏览器、退出未确认，以及退出准入期间中断批次后已收集的失败）仍发布单条/汇总报告并同样由根呈现一次。被拒绝的浏览器请求照旧记录为会话 `browserFailure` 并保留 running/URL，重开成功即清除；`NSWorkspace.open` 返回 true 只表示系统接受请求，不证明页面已渲染。
- 退出准入：`applicationShouldTerminate` 先执行 `PreviewAppModel.beginTermination()`：置 `isTerminating` 并调用协调器 `stopAcceptingNewOpens()`（一次性；准备完成后才到达的打开请求同样被拒绝），随后 `completeTermination()` 等待 `stopAll`（含 starting）确认真实退出，全部 terminated 才回复允许退出；未确认的停止保留会话与管理入口并拒绝退出。强杀不经过该回调，由每个 child 的 ownership 管道负责。
- 最近目标与完整面板（任务 4.1–4.3）：`Models/RecentTarget.swift` 是持久记录（规范化身份、用户选择展示路径、`ManagedProcess.TargetMode`、最近使用时间），`Utilities/RecentTargetsStore.swift` 用独立版本化键 `go-grip-recent-targets-v1` 保存至多 20 个不同身份——数组按最近使用排序、同一身份去重并前移、超出淘汰最旧；JSON 解码加载不探测路径、不启动任何东西。旧 `go-grip-history`（50 项原始路径）不被读取、写入或删除；`Utilities/Storage.swift`、`Models/HistoryEntry.swift`、`GoGripTests/StorageTests.swift` 与工程引用已随本任务移除，无兼容层。`PreviewAppModel` 持有 store 并只在**已验证 running** 的打开结果之后更新 recency（`openBatch` 覆盖 Finder/手动/最近入口与复用运行会话；运行行 `Open in Browser` 的显式重开同样更新），随后才请求浏览器：准备/类型/启动失败、取消与浏览器拒绝都不撤销或伪造历史。最近行 Open 以记录的规范化身份定位（准备身份路径），并把记录的用户选择路径作为展示/失败上下文以 `openBatch` 的展示覆盖传入（同一准备/身份解析/复用/批次/退出准入/一次报告）；符号链接被改指或删除后也不会重定位到新目标，而是在原身份上重开或按记录路径报 unavailable；不使用 bookmark/inode、不建第二启动器。`clearRecentTargets()` 只清新键与发布列表，不触碰协调器、已验证 URL 或 owned 资源。面板 `Views/PopoverView.swift`（固定 420×460、内部滚动）运行会话区在前并显示实际 phase、target `unavailable` 原因、reload `pending`/`degraded`/`off` 与浏览器/退出失败原因，最近区在后提供逐行 Open 与 Clear；全部动作为显式按钮，无隐式状态点。
- 已验证（任务 4.1–4.3，2026-10-05 开发构建 `2195a0bc…` 临时安装于 `/Applications`，完成后已还原 08 构建 `5db6f61d…`）：`make macos-test` **87 tests / 0 failures**（81 → 87：3 项隔离域 store 回归——21 个身份保留具体 20 个与淘汰对象、重开较旧身份前移且不重复并保留最新展示路径、重建 store 恢复相同次序、clear 只清新键并保留同域哨兵与旧 `go-grip-history` 键哨兵；3 项模型回归——批次内失败目标不进入历史且浏览器拒绝不撤销已验证目标、从最近入口与真实符号链接别名重开运行目标复用同一 PID/generation/URL 且只保留一条历史并更新次序、清空后运行会话及真实停止能力保持；另有改指链接仍按记录身份重开的断言并入别名回归；裁决修正与批处理测试去竞争后连续 3 次全量 `make macos-test` 均 87/0）。真实开发 App 面板/浏览器观察（面板文本与动作经 AX 读取/触发，进程/端口/HTTP 由终端配对）：Services 与面板 `Open…` 多目标打开得到 5 个 owned child，各自 127.0.0.1 随机端口与可辨认内容（目录、中文+空格 `.MD`、真实空目录、两个不同父目录下的同名 `docs`）；运行区先于最近区且同名完整路径可区分；`Copy URL` 得到会话实际 URL（HTTP 200）；`Open in Browser` 打开同一 URL 且不新建服务；`Stop` 只释放对应 child/端口（`lsof` 归零），其余会话与独立 CLI（6419）继续；`Clear` 在有活动会话时清后 HTTP 与管理动作可用、真实默认域新键消失而 `go-grip-history` 保留；面板 Quit 释放全部 owned PID/端口后退出；重启仅恢复最近列表（无 child/端口/新浏览器标签），点最近行 Open 才启动新 child；移动专用最近目标后重开显示原路径、真实 missing 原因与 08 指导且无假 running；移动运行中目标后实际请求触发 `Target unavailable: open <原路径>: no such file or directory` 并在恢复路径后消失；运行期注入 `mkdir -m 000` 受限目录得到 `Hot reload degraded: walk error at …: permission denied`（与 phase 分离，管理入口保留）；`kill -9` owned child 得到 `Stopped` + `The preview service exited unexpectedly (status 9)`、无自动重启、最近入口可主动重开，显式 Stop 不误报该原因；带标注临时受控拒绝构建（CDHash `8474d42f…`，观察后已还原并以 `2195a0bc…` 重建）观察到一次根提示、会话保持 running/URL 与最近记录，第三次请求真实打开页面并清除行内失败。
- 已验证（任务 4.3 最终构建 `9c4ddf9b…`，2026-10-05）：经别名 `alias-alpha` 打开 `alpha`（记录 identity=`alpha`、display=`alias-alpha`）并退出后，把别名改指 `gamma` 再重启，点最近行 Open 得到的新 child 仍为 `data/alpha`（HTTP 内容为 alpha），记录保持一条且 display 仍为别名——裁决的实现语义（身份定位 + 保留用户选择路径）在真实 App 上复核通过。
- 首次引导与完整本地化（任务 5.1/5.2，2026-10-06 开发构建 CDHash `3949aca2…`，本机 macOS 27.0.1）：`Utilities/FirstUseGuidanceStore.swift` 用独立键 `go-grip-first-use-guidance-shown` 只记录“是否已自动展示过引导”，与 `go-grip-recent-targets-v1` 及旧 `go-grip-history` 完全分离（不读、不迁移、不清）；`AppDelegate` 在根就绪、provider 注册后判断并展示一次，面板 footer 的 Help 经 `AppAlertPresenter.presentServiceGuidance()` 随时重看——引导与失败报告、数量确认共用同一非模态单槽，不建第二弹窗/队列。引导内容为 Finder 选中目录/`.md` → Services 命令 → 默认浏览器、菜单不出现时的 **System Settings → Keyboard → Keyboard Shortcuts… → Services** 检查、条件式权限说明（普通目标无需额外授权；系统询问时允许 GoGrip 访问）与面板手动入口；不调用重扫、不修改服务启用偏好、不宣称菜单可见。资源：`GoGrip/en.lproj/{Localizable,ServicesMenu}.strings` 与 `GoGrip/zh-Hans.lproj/{Localizable,ServicesMenu}.strings`，经 Xcode PBXVariantGroup + Resources phase 进入实际构建，`knownRegions` 增 `zh-Hans`、development region 保持 en；`ServicesMenu.strings` 以 Info.plist 精确默认键 `Open with GoGrip` 提供名称（zh-Hans 为“用 GoGrip 打开”）。本地化边界：应用自有标签、错误包装与行动指导经 `NSLocalizedString`（键=英文原文，zh-Hans 翻译）或 SwiftUI `LocalizedStringKey` 呈现；Go code/message、target/reload reason、协议/进程 detail、系统 `localizedDescription`、path/URL 保留原文。验证：`make macos-test` **92 tests / 0 failures**（87 → 92：+2 隔离域 `FirstUseGuidanceStore` 回归——首次展示一次且重建实例保持、与最近/旧键分离；其余为 07 修正带入）；`make macos` Release 构建 + `codesign --verify --deep --strict` 通过，产物含四份 lproj 资源且 `CFBundleDevelopmentRegion=en`。真实观察（语言经全局 `AppleLanguages` 切换并已还原 `en-US`；Services 启用开关、首次键与最近键均记录并还原，旧 `go-grip-history` 未动）：en 首次冷启动自动显示引导，且同一时刻 Finder 服务已由 launchd 启动宿主（PPID=1）并建立 owned child/回环端口/HTTP 内容——引导不阻断 provider 或批次；第二次启动不自动显示、面板 Help 可重看且不改变运行会话；zh-Hans 环境自动引导与面板/6 目标数量确认/批次汇总/陈旧目标/浏览器失败/运行期退出原因均为中文包装并保留原始 path/URL/reason；真实 Finder Services 菜单在 zh 显示“用 GoGrip 打开”、不支持语言（ja）回退 “Open with GoGrip”，两者都经真实调用进入同一内置 Go pipeline；用户在 System Settings 关闭服务后引导仍可查看、App 未自动启用，按说明恢复后菜单重现；受控浏览器拒绝构建（标注、源码还原后重装最终构建）观察到 zh 浏览器失败提示且会话/URL 保持；面板 Quit 释放 owned PID/端口，`Stop All` 不影响独立 CLI（6419 继续 HTTP 200）。证据时序：Finder 菜单/服务开关/面板/确认/失败/退出等观察完成于 `e01603e1…` 构建；随后仅按审阅修改引导措辞（补“预览在默认浏览器打开”与“系统设置→键盘→键盘快捷键…→服务”路径），并在最终 `3949aca2…` 构建上重新观察 en/zh 首次引导与 en Help 实际表面、重跑 `make macos-test` 92/0——除引导措辞外无行为差异。advisory（留待候选验收观察）：自动首次引导经 `NSApp.activate` 展示；本轮未观察到其阻断 provider/批次/浏览器流程，若任务 13 候选 smoke 发现浏览器焦点被抢，可让自动实例不激活展示（Help 实例不变）。限制：登录项（任务 5.3）、Intel/macOS 13、网络卷与候选包（任务 6–7）未验证；开发构建证明不提升为候选验收。
- 用户可选登录启动（任务 5.3，2026-10-06 最终开发构建 CDHash `d22aee2e…`，本机 macOS 27.0.1）：`Services/LoginItemRegistering.swift` 是唯一登录项适配器（协议 + `MainAppLoginItem`：`SMAppService.mainApp.status`、`register()`、`unregister()`、`openSystemSettingsLoginItems()`），`PreviewAppModel` 注入它并只发布系统真实 `status`（模型初始化与每次面板打开时经 `AppDelegate.togglePopover` 重读；无持久化期望 bool、无 helper/LaunchAgent/其他登录项管理）；`setLoginItemEnabled(_:)` 仅在用户按钮动作时调用 SDK，随后重读真实状态：抛错时以一次 `OperationFailure`（本地化包装 + 系统 `localizedDescription` 原文）沿用既有根级一次原生报告与面板可查看行，绝不保留假成功。面板行 `Views/PopoverView.swift`：`Not set`（`notRegistered`/`notFound`，附 `Turn On`）、`Enabled`（附 `Turn Off`）、`Waiting for approval`（`requiresApproval`，附“System Settings → General → Login Items”可执行指导、`Open Login Items…`（用户触发打开系统设置）与 `Turn Off`）。裁决与实测（2026-10-06 用户批准）：当前 macOS 上从未注册的主 App 返回 `.notFound`（rawValue 3）而非 `.notRegistered`，且该状态下 `register()` 可成功，故 `notFound` 与 `notRegistered` 同为未注册呈现并提供明确开启；真实结果以 `register()` 返回/抛错后重读的状态为准（票面该句已同步修订）。验证：`make macos-test` **95 tests / 0 failures**（92 → 95：3 项确定性回归——register 抛错后显示实际状态并发布原因、register 成功但需批准时呈 `requiresApproval` 而非 enabled 且非失败、unregister 抛错保持 enabled 并发布原因；变异检验把“重读真实状态”改为“假定请求生效”后 3 项全部失败并已还原）；`make macos` Release 构建 + `codesign --verify --deep --strict` 通过。真实 native smoke（用户操作 + 终端/AX 观察；`/Applications` 为本票最终构建，10 构建 `3949aca2…` 备份于 `~/Desktop/GoGrip-backups/GoGrip-10-backup.app`）：首次未注册状态显示 `Not set` + `Turn On`，启动/打开面板不自行注册（BTM 无 GoGrip 条目）；`Turn On` 后状态重读为 `Enabled`，`sfltool dumpbtm` 显示 `GoGrip / Type: app / Disposition: [enabled, allowed, notified] / Identifier 2.com.showgp.GoGrip / URL /Applications/GoGrip.app`，`Turn Off` 后回到 `Not set` 且条目移除；**登录项关闭、宿主未运行**的真实 Finder Services 调用仍由 launchd 冷启动宿主（PID 86023，PPID=1）+ owned child 86035（`-r -- /tmp/gogrip-11-smoke`，127.0.0.1:64631，正文 `cold-start-marker-11 unique`），全程面板显示 `Not set`——登录启动不是服务前提；活动会话（child 86035）期间 `Turn On` 不改变 PID/端口/URL 与最近记录，关闭面板后继续服务，面板 Quit 释放宿主与 owned PID/端口，且退出后 BTM 仍保留用户注册（退出不注销），独立 CLI（6419）在 App 退出前后持续 HTTP 200 与自身内容；`requiresApproval` 与操作失败两个难触发表面以标注临时受控构建（`667481e7…`，适配器仅报告状态并抛 Service-Management 风格错误，源码已还原并重装最终构建）观察——面板显示 `Waiting for approval` + 指导 + `Open Login Items…`/`Turn Off`，按 `Turn Off` 得到一次根提示（`Launch at login` / `Could not change the login startup setting: The operation couldn’t be completed. (SMAppServiceErrorDomain error 1.)`）且面板保持真实状态与可查看失败行，`Open Login Items…` 实际打开 System Settings 的 Login Items 面板；zh-Hans 下真实表面为 `登录时启动 / 已开启 / 未开启`、指导与失败包装（`无法更改登录启动设置：…`）为中文且系统原因原文保留，不支持语言（ja）回退英文；语言切换（全局 `AppleLanguages`）已还原 `en-US`，登录项已恢复未注册、System Settings 已关闭。限制：`requiresApproval`/操作失败仅为受控补证（本轮无真实系统拒绝/审批成功）；未执行真实注销/重启登录会话的自动启动观察，不据此宣称已证；候选包（12/13）、Intel/macOS 13 未验证。
- 受控/限制（2026-10-05 实测，属 07 已记录的两个应用级模态未串行化）：根失败提示为 `NSAlert.runModal`；其显示期间到达的 Services 请求在 pasteboard 交换超时（`Application com.showgp.GoGrip didn't return pasteboard data in time`，`NSPerformService` 返回 false）并丢失该请求；同一失败提示随后出现无法通过 OK/Return 关闭的 AXDialog 叠加，该实例需强制重启（owned child 经 ownership EOF 正常退出）。本任务沿用批准策略、未改变该行为。
- 共享提示呈现修正（2026-10-06，07 票重开，用户批准方案 A）：上一条限制已修复。根失败提示与批次数量确认不再使用 `NSAlert.runModal`，改由 `Services/AppAlertPresenter.swift` 的单槽呈现：两类提示都展示为非模态 `NSAlert` 窗口（`NSApp.activate(ignoringOtherApps: true)` + `center()` + `makeKeyAndOrderFront`；按钮 `target/action` 改指槽回调，`willClose` 视为关闭），主 run loop 保持默认 mode，因此提示显示期间的 Services/pasteboard 交接继续被处理；同一时刻最多一个提示——失败报告按到达顺序各呈现一次（每个失败批次仍恰好一次汇总），数量确认改经 `BatchQuantityConfirming.confirmOpening(count:) async` 用 continuation 等待响应（取消仍不启动本批、不请求浏览器、不动既有会话），`NativeBatchQuantityConfirmation.swift` 删除、无 shim。文案/按钮/汇总内容不变（本地化仍归 10）；provider、批次、会话事实、退出准入与最近记录语义未改。TDD：以 `cannot find 'AppAlertPresenter' in scope` 编译失败为 red；`AppAlertPresenterTests` 3 项策略回归（失败报告一次一个且按序、确认响应只在展示后恢复、确认排在可见报告之后）转绿；变异检验临时去掉槽的 `current` 守卫后其中 2 项失败、已还原；`PreviewAppModelTests` 的 `ControlledConfirmation` 改 async 并新增运行行重开的前移断言（变异去掉 `openBrowser` 的 recency 更新后该断言失败、已还原）。`make macos-test` **90 tests / 0 failures**（87 → 90），`make macos` Release 构建 + `codesign --verify --deep --strict` 通过（开发构建 `eead0216…`）。native（2026-10-06；替换 `/Applications` 前先备份 09 构建 `9c4ddf9b…` 至 `~/Desktop/GoGrip-backups/`）：冷启动（宿主未运行）Services 请求即出根提示且面板未展示；**提示显示期间**真实 Finder Services 有效目标请求端到端完成（child 65659 → `127.0.0.1:58601`，HTTP 正文 `alpha fix-smoke unique`，Chrome 打开 `…/alpha-notes.md`），`log show` 无 `didn't return pasteboard data in time`；OK 与 Return 各关闭一次；两个失败批次的提示一次只出现一个、第二个在前一个关闭后出现且可关闭；数量确认显示期间真实 Services 请求（`iota` → child 65793、`127.0.0.1:58764`、正文 `iota fix-smoke unique`）完成，Cancel 后 6 个批目标零 child、Open 后 6 个会话各自端口与内容正确；面板 Stop（`gamma` 端口/进程归零）、Stop All 与 Quit 释放全部 owned child，独立 CLI（6419）全程存活。解锁后补测（当日下午完成）：① 另一前台 App（Finder）时模型化提示在其窗口之上可见且渲染完整（截图 `11-alert-front-full.png`、`10-alert-layout-fixed.png`）；② 提示关闭后面板仍在最近列表下方保留同一条失败报告（红色原因行，截图 `12-panel-report.png`）。补测中发现并最小修正渲染缺陷：绕过 `runModal`/`beginSheet` 后 `NSAlert` 的惰性布局不会自动完成，首次补测的提示窗口渲染异常（可见抑制复选框、空按钮、正文未排版，窗口 260×328）；`presentModelessly` 展示前调用 `alert.layout()`（AppKit 头文件即为此用途）后恢复为常规单按钮提示（260×266，路径+原因+OK），并在该构建上复测“提示显示期间真实 Finder 请求端到端完成、OK 与 Return 各关闭一次、一次汇总”。最终构建 CDHash `7a2c78fa…`（`/Applications` 当前为此构建；09 回退备份保留）。
- 环境说明（2026-10-05，如实披露）：本轮清理阶段发现 08 构建的唯一副本已随备份移除，无法恢复其 `5db6f61d…` 构建；`/Applications/GoGrip.app` 现为本任务最终开发构建 `9c4ddf9b…`（同 bundle ID、ad-hoc 签名）。ad-hoc 身份变化可能使 08 时期的 TCC 授权需要重新授予（既有实测限制）；08 整票状态与网络共享缺口不受影响。
- 已验证（2026-10-04 开发构建）：`make macos-test` 68 tests / 0 failures（含 7 项新增宿主 seam：浏览器拒绝保留 running/URL、失败不产生假运行或浏览器动作、重开恢复、退出准入拒绝迟到准备/启动以及停在停止之后才恢复的重开、停止未确认拒绝退出、无会话直接放行）；Release/Debug 开发 App `make macos`/`make macos-run` 构建成功并 ad-hoc 签名（`codesign --verify --deep --strict` 通过；工具与宿主均为双架构）、`go-grip` 仅位于 `Contents/MacOS`、无 `.appex`、`LSUIElement=true`；`open` 后 `lsappinfo` 显示 `type="UIElement"` 且进程稳定运行。
- 已验证（2026-10-04 本机手动 smoke，含终端侧 PID/端口观察）：NSOpenPanel 单选普通目录（嵌套文档）、中文+空格 `.MD`、可访问空目录、父/子/单文件独立目标及同名目标路径区分、重复打开与符号链接别名复用同一 PID/URL；Copy URL 得到实际地址、重开浏览器、关闭浏览器标签后服务继续；单停与 Stop All 释放 owned PID 与监听端口（仅 127.0.0.1）且不影响其他会话或独立 CLI；主目录外（`/tmp`）目标正确服务；`.png` 与不可读目录显示原生错误且不建会话；面板 Quit 与宿主 `kill -9` 后 owned 子进程与端口释放、独立 CLI 全程 HTTP 200 且不出现在面板；空列表/单行面板布局问题修复后复验通过。
- 已验证（2026-10-05 批次开发构建）：`make macos-test` **78 tests / 0 failures**（69 → 78：新增 9 项批次回归——5 个不同有效目标直接打开不发确认、6 个先以实际数量 6 确认、6 条路径经真实符号链接归一为 5 个目标、不支持项不计入有效数量、取消不启动也不改动既有会话的 generation/URL/PID、混合批次失败后继续且只发布一份含目标与原因的汇总报告、全部成功不发布报告、退出准入在数量确认与逐项启动等待窗口内不产生后项 child/浏览器请求；05 的 7 项宿主回归已迁移到统一批次入口 `openBatch`）。`make macos` Release 构建成功、`codesign --verify --deep --strict` 通过、工具仍只在 `Contents/MacOS/go-grip`、`LSUIElement=true`。
- 已验证（2026-10-05 Services seam，任务 3.1/3.2/3.4 的 07 贡献）：`make macos-test` **80 tests / 0 failures**（78 → 80：新增 2 项 provider 输入 seam——真实 NSPasteboard 写入“文档 资料”目录、`说明 文章.MD`、`other` 目录三个目标后，selector 返回即清空 pasteboard 仍由生产 provider → 同一批次得到全部三个 running 会话与三次浏览器请求（变异检验：临时取 `urls.first` 或改为异步重读均使该回归失败后还原）；空请求在返回前写 `error` 且不建会话、不建 child、不请求浏览器）。`make macos` Release 构建成功、`codesign --verify --deep --strict` 通过；产物 `Info.plist` 含唯一 `NSServices` 声明且键值为 `openWithGoGrip`/`GoGrip`/`Open with GoGrip`/`public.file-url`/`com.apple.finder`，工具仍只在 `Contents/MacOS/go-grip`、无 `.appex`、`LSUIElement=true`。
- 真实 Finder 冷启动与内容（2026-10-05 本机，用户操作 + 终端观察；宿主未运行、无登录项，`/Applications/GoGrip.app` 为本轮 Release 构建、ad-hoc、CDHash `4b0c900c…`；同 bundle ID 旧副本（含 FinderSync 的旧示例）已按用户批准的环境准备移入废纸篓（`~/.Trash/GoGrip.app-legacy-2026-10-05`）并从 LaunchServices 注销，未删除用户数据、未改服务偏好/登录项；服务库仅 `Open with GoGrip (0x238)`、`message=openWithGoGrip`、`port=GoGrip`、`send types=public.file-url`）：用户经 **服务 → Open with GoGrip** 调用后，宿主 **44058 由 launchd（PPID=1）于 16:01:33 系统启动**，可执行路径 `/Applications/GoGrip.app/Contents/MacOS/GoGrip`；三个 owned child 44060（`--managed … -r -- /tmp/gogrip-07-smoke/普通 目录`，127.0.0.1:51434）、44067（`说明 文章.MD`，51475）、44070（`empty-空目录`，51496）分别提供目录递归内容（嵌套文档 `嵌套内容` HTTP 200）、中文空格单文件（`你好，GoGrip 单文件预览`）与空目录空状态（`No Markdown`），默认浏览器实际打开页面；独立 CLI（6419）全程 HTTP 200 且不出现在面板。
- 真实跨入口、多选与失败恢复（2026-10-05 同上）：面板 `Open…` 重开 `普通 目录` 及其符号链接 `alias-普通目录` 均复用同一会话 **44060/51434**；面板新开 `t3` 后从 Finder 服务重开 `t3` 仍只有 **44262/51894** 一个 child。Finder 多选 `alias-t1、t1…t5`（6 路径）→ **无数量确认**、`t1` 唯一 child 44273（别名经规范化身份去重）；多选 `t1…t6` → 确认显示数量 **6**、接受后 `t6` 新建（44304/52032）；再次多选 `t1…t6` 并**取消** → 无新建/重开/浏览器动作，既有会话 PID/启动时间不变。混选 `photo.png`、`无法访问目录`（chmod 000）、`普通 目录` → `普通 目录` 复用，结束后**一次**原生汇总同时列出两个失败目标及可获得原因（非逐项弹窗），面板保留该报告；`photo.png`/`无法访问目录` 均未产生 child。
- 冷启动失败提示（2026-10-05 同上，方案 A scope decision 由用户裁决后实施）：独立审阅发现“宿主冷启动、面板从未展示时整批失败的汇总提示”缺失（唯一提示通道原是面板内容的 SwiftUI `.alert`，popover body 未渲染时不出现）。经用户裁决取方案 A：失败汇总改由应用根呈现——`AppDelegate` 订阅 `model.$lastOperationFailures`，每次非空发布经一次主 actor 转派（`Task { @MainActor in }`，避免在发布调用栈内模态）后执行 `presentFailureReport`（单个 NSAlert；标题为单条报告时的目标路径、多条时为 `GoGrip`，正文逐条列出“路径 + 原因”以空行连接，一个 `OK` 按钮——与方案 A 之前面板内 SwiftUI alert 的布局相同；其 `GoGrip`/`OK` 与正文格式随任务 5.2 的“错误”范围一并本地化），与面板可见性无关；`PopoverView` 移除本地 `.alert` 与 `presentedFailures` 状态，只保留会话列表下方的同一份报告清单（该变更归属 07，理由：Services 冷启动请求无法依赖面板渲染，规格“一次原生提示”不含“先打开面板”条件）。验证：`make macos-test` 仍 **80/0**、`make macos` 构建成功、`codesign --verify --deep --strict` 通过；首版修复（CDHash `fec72e55…`）经用户对 `photo.png` 冷启动观察“立即一次提示、开面板无第二弹窗”；按 Standards 复审 advisory 将转派改为 `Task { @MainActor in }` 后重建（安装副本最终 CDHash `02d35053…`，取代 `4b0c900c…`/`fec72e55…`），用户复验同一行为一致，并追加面板交互路径观察：面板打开时经 `Open…` 选 `无法访问目录` 触发失败仍只弹一次提示，重开面板报告仍在且不重复弹；无 child/端口残留，Quit 后宿主退出；独立 CLI 全程不受影响。
- 浏览器拒绝表面与生命周期（2026-10-05 同上；带标记的临时受控拒绝构建 CDHash `865021c7…`，拒绝前两次请求后已还原并重装当时正式构建 `4b0c900c…`，源码无残留）：从 Finder 服务冷启动受控构建（宿主 **44696 亦由 launchd 启动**）后，`t4`（44698/52466）与 `t5`（44703/52507）保持 running 与实际 URL；用户观察到一次原生弹窗与面板行内错误（"The system could not open the default browser for the preview"）；对 `t4` 执行 `Open in Browser`（第三次请求）实际打开页面并清除该行错误；关闭 `t4` 浏览器标签后 52466 继续 HTTP 200；面板 `Stop` 仅释放 `t5`（44703/52507），`t4` 保持；面板 `Quit` 后宿主与全部 owned PID/端口释放（9 个会话端口逐一确认关闭），独立 CLI 不受影响。
- 真实 Go 混合批次 smoke（同一生产入口 `PreviewAppModel.openBatch` → 生产协调器 → `ManagedProcess` → 内置 Go，仅浏览器/确认受控，临时用例已删除）：4 个有效目标不发数量确认；选择顺序第 2 位的不可读 `.md` 得到真实 Go 失败 `target-unavailable: access target: … permission denied`，其后的中文/空格单文件与空目录仍各自启动并返回真实 HTTP 200（正文含“你好”）；结束时一份报告含 3 个失败目标与可获得原因；`stopAll` 释放全部 owned PID。
- 已验证（2026-10-05 本机手动 smoke，用户操作 + 终端侧 PID/端口/HTTP 观察）：面板多选 5 个目标直接打开；6 个目标先显示数量 6 的确认、接受后打开；6 个目标取消既不新建也不重开，既有会话保持；`five/` 的 6 条路径（含 `alias-d1`→`d1`）归一为 5 个目标且不发确认；`Open in Browser`/`Copy URL`/`Stop` 行为符合预期。终端侧配对观察：一轮 2 目标批次的两个 owned child（38360/38361）分别在 60579/60582 提供 `HTTP 200`（正文 `target6 root`/`target7 root`），面板 `Quit` 后同一采样点 `app_alive=no children=[]` 且无端口继续服务；受控拒绝构建一轮的 owned child（39084/39085）在 60804/60806 服务，经正常退出路径结束后 app、child 与端口全部释放；独立 CLI（6419）在批次与退出前后均返回 `HTTP 200` 与自己的内容，且不出现在面板。
- 批次错误汇总的原生表面（受控拒绝，已还原、无源码残留）：带标注的临时一次性拒绝构建（前两次浏览器请求按系统拒绝返回 false）观察到**一次**原生提示同时列出两个目标及可获得原因，不是逐项弹窗，且失败未阻止后项；对一行 `Open in Browser` 成功后该行错误消失、其他行保留。面板会话列表下方显示同一报告数组（本轮未单独复核其版面）；真实系统拒绝的表面见 05 的受控拒绝记录，正常默认浏览器页面在各轮均实际打开。
- 浏览器失败的原生表面：用带标注的临时一次性拒绝构建（受控拒绝、非真实系统失败；已还原，源码无残留）观察——原生错误弹窗、会话保持 running 与实际 URL、重试成功后在会话行清除浏览器失败；恢复后的“最近一次失败”横幅按 2026-10-04 裁决保留为历史（规范未规定清除时机）。
- 真实授权主体与最小权限（任务 3.3/08，2026-10-05 本机 macOS 27.0.1 开发构建；用户操作 Finder 服务与系统提示，终端侧 tccd/进程/端口/HTTP 观察）：普通位置（目录递归、中文+空格 `.MD`、真实空目录）无任何系统授权提示与 FDA/辅助功能/自动化前置，实际 URL 仅 `127.0.0.1` 随机端口。受保护目录（`~/Documents` 专用目标）：提示 `"GoGrip.app" would like to access files in your Documents folder.`，tccd `AUTHREQ_PROMPTING` 主体为 `com.showgp.GoGrip`（`/Applications/GoGrip.app`）；拒绝记为 `TCCDEvent: type=Create … identifier_type=Bundle ID, identifier=com.showgp.GoGrip`；内置工具子进程（`com.showgp.GoGrip.go-grip`）的请求以 responsible=App、subject=`com.showgp.GoGrip` 判定——**授权主体是 App，内置工具由 App 覆盖而非独立身份**。拒绝后无遗留 child/端口/假 running；系统设置打开 GoGrip 的 Documents 项（`type=Modify`）后重开即恢复。外接 USB/APFS 卷（专用目录）与一次性磁盘映像测试卷（detach）：允许时经 native→Go 得到正确内容（`kTCCServiceSystemPolicyRemovableVolumes` 提示同为 App 主体）；专用目标改名/删除与卷 detach 后原路径请求得到真实 `HTTP 500`（`open …: no such file or directory`，非空目录），会话与管理/停止入口保留，无自动迁移/重启/重连；真实空目录仍为可用空状态。单停/Quit 仅释放对应 owned child 与全部 owned PID/端口，独立 CLI（`*:6419`）全程 HTTP 200。
- 访问失败指导（任务 3.3 用户批准的最小修正）：失败报告新增条件式 `accessCheckGuidance`，仅追加到 `describe(TargetPreparationFailure.unavailable)` 与 `describe(ManagedLaunchFailure.fatal)` 且 `code == "target-unavailable"`；内容为“目标权限、卷挂载/共享状态、以及在 macOS 已请求授权时 System Settings → Privacy & Security → Files and Folders（GoGrip 条目）”，不断言 TCC、不统一导向 FDA，不支持文件等其他失败保持原样（该指导句为用户可见文本，随任务 10 的本地化“错误”范围一并覆盖）。scoped TDD：`testAccessFailuresCarryConditionalGuidanceWithoutCoveringOtherCauses` 先红（81 tests / 2 failures，恰为两类访问失败缺指导）后绿（81/0）；重建后用户在实际拒绝报告中逐字复核该句，恢复复核通过。
- 身份限制（ad-hoc，实测）：重建改变 CDHash（`02d35053…` → `5db6f61d…`）后 tccd 报 `Failed to match existing code requirement`，Documents 与 RemovableVolumes 均按新构建重新请求授权——开发/候选阶段不承诺跨构建授权稳定。
- 未验证/限制：候选打包已由任务 12 交付，见第十六节；登录项（任务 5.3）已接通并有本机 native smoke（见上）；首次引导/可再次查看帮助与简中/英文本地化已由任务 5.1/5.2 接通并有本机 native smoke（见上）；最近记录/完整面板已由任务 4.1–4.3 接通并有本机 native smoke（见上）；权限 gate 已由 08 接通并验证（见上）；`NSOpenPanel` 按 `[.folder, markdown]` 过滤，面板无法选入不支持文件——Finder Services 混选与跨入口复用见上面的 07 记录，混合不支持/准备失败输入的批次语义由同一入口的真实 Go smoke 证明；注册时序（协调器/批次依赖先于 provider 注册、注册后第一个请求早于首次面板展示）由实现结构与真实/临时 smoke 覆盖，`make macos-test` 的永久回归只覆盖 provider 输入通道（同步取齐全部 file URL、返回前取齐、空请求即时报错），不构成注册时序证明；根失败提示同样无自动回归（单次呈现/无第二次弹窗由真实 smoke 观察）；两个应用级提示的串行化缺失已由 2026-10-06 的单槽非模态修正解决（见本节“共享提示呈现修正”，旧限制不再适用）；Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷验收经批准延期至后续 change，保持未验证（2026-10-07 范围对齐）；批准范围内的最终候选验收已完成，08/13 已关闭（见归档变更 `openspec/changes/archive/2026-10-07-rebuild-macos-preview-app/`）；候选已在当前机器安装并实际运行（见票 13 记录），正式干净安装、Developer ID 签名与公证尚未验收；不管理 Chrome 后代。
- 提交前验证（2026-10-07）：既有 `TestManagedServePublishesReadyAndReleasesPortOnOwnerLoss` 仅读取 ready 后停止消费无缓冲 stdout pipe，后续状态事件可阻塞 watcher 与退出清理；测试现持续消费协议流，与宿主行为一致，不修改生产退出逻辑或放宽断言。修正后该测试连续 20 次通过，`go test ./... -count=1`、`go vet ./...`、变更 Go 文件格式检查及四份主规格严格验证通过。

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

## 十、端口监听与回退机制 (`internal/listener.go`)

```
listenOn(bind, port, strictPort):
  ├─ bind="" (独立 CLI): 监听通配地址（与既有行为一致）
  ├─ bind="127.0.0.1" (managed): 真实回环地址
  ├─ strictPort=true (用户显式指定 --port / managed port 0):
  │   └─ net.Listen 失败 → 直接报错
  │
  └─ strictPort=false (默认):
      ├─ 尝试 port (6419)
      ├─ 失败 → port+1
      ├─ 失败 → port+2
      ├─ ...
      └─ 最大尝试 100 次 → 全部失败则报错
```

managed 路径固定 `listenOn("127.0.0.1", 0, true)`：端口由 OS 分配，URL 使用 listener 实际端口；独立 CLI 的通配绑定与网络暴露政策不变。

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

### macOS 宿主与候选 App/DMG

```bash
make macos-candidate        # 双架构候选 DMG（per-arch archive → 签名/架构检查 → 打包）
make macos-archive          # 按 ARCHS=arm64 / ARCHS=x86_64 分别 archive 到 macos/.build/GoGrip-<arch>.xcarchive
make macos-dmg              # 逐个 archive 生成含 /Applications 入口的 GoGrip-arm64.dmg 与 GoGrip-x86_64.dmg
make macos-candidate-check  # 逐包校验 ad-hoc 签名、identifier、本架构单切片（另一架构必须失败）
make macos-test             # Swift 宿主行为套件（真实 Go）
make macos                  # 开发用 Release 构建
```

- 唯一运行路径 `GoGrip.app/Contents/MacOS/go-grip`，运行期没有 Resources/PATH/源码回退。
- `macos/Scripts/build-go-grip.sh` 按 `$ARCHS` 以 `CGO_ENABLED=0` 构建 darwin arm64 或 amd64 单切片（双架构时保持 lipo 合并路径），先按独立 identifier `com.showgp.GoGrip.go-grip` ad-hoc 签名，再由 Xcode 以 `com.showgp.GoGrip` 签宿主（不依赖 Developer ID/公证凭据，运行期不 chmod/重签）。
- 候选 archive 使用 Release 配置，按 `ARCHS = arm64` / `ARCHS = x86_64` 各 archive 一次（`ONLY_ACTIVE_ARCH = NO`、deployment 13.0）；每个包内宿主与内置 Go 均为对应单切片，`make macos-candidate-check` 要求本架构 `lipo -verify_arch` 通过且另一架构失败（实测 `LC_BUILD_VERSION` minos：宿主 13.0，Go 工具 12.0）。
- 候选产物位于 `macos/.build/candidate/`：`GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`（各自含 `/Applications` 入口的安装镜像）；不再产出候选 App zip 与 universal DMG。
- 本阶段止于候选：正式 Developer ID 签名、公证、干净安装验证与公开发布另行处理。

### 测试

```bash
go test ./...                       # 运行所有测试
make macos-test                     # macOS 宿主行为套件（真实 Go 工具）
```

### CI (`build.yml`)

触发: push / PR

1. mise 设置 Go 环境
2. `go build`
3. `go test ./...`
4. `gofmt -d .` 格式检查
5. `golangci-lint` 静态分析
6. macOS job（PR）: `make macos-test` + `make macos-candidate`，把 `GoGrip-arm64.dmg`/`GoGrip-x86_64.dmg` 作为候选 workflow artifact 上传

### CD (`release.yml`)

触发: 推送 `v*` tag

1. `go test ./...`
2. 6 架构交叉编译: darwin/linux/windows × amd64/arm64
3. `-trimpath -ldflags="-s -w"` 精简二进制
4. `.tar.gz`/`.zip` + `checksums.txt`
5. 创建 GitHub Release（CLI 资产，流程不变）
6. macOS App job 仅按同一 `make macos-candidate` 构建候选并上传 workflow artifact；不再向 GitHub Release 上传 App/DMG

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
