# go-grip 工程现状说明

> 最后更新: 2026-05-16

## 一、项目概述

`go-grip` 是一个 Go 语言编写的 Markdown 渲染服务器，将 Markdown 文件渲染为 GitHub 风格的 HTML 页面并通过浏览器实时预览。它是 Python 版 [grip](https://github.com/joeyespo/grip) 的完整重写，消除了对外部 GitHub API 的依赖。本仓库是 [chrishrb/go-grip](https://github.com/chrishrb/go-grip) 的 fork，增加了多文件目录浏览能力。

- **模块路径**: `github.com/showgp/go-grip`
- **Go 版本**: 1.25
- **许可证**: MIT

---

## 二、项目结构

```
go-grip/
├── main.go                  # 入口点
├── cmd/
│   └── root.go              # Cobra CLI 定义与参数
├── internal/
│   ├── server.go            # HTTP 服务 + WebSocket 热重载
│   ├── server_test.go
│   ├── parser.go            # goldmark 解析器封装 + CJK 修复
│   ├── parser_test.go
│   ├── listener.go          # 端口监听与回退
│   ├── listener_test.go
│   ├── target.go            # 服务目标解析（文件/目录模式）
│   ├── target_test.go
│   ├── articles.go          # 文章发现、排序、导航
│   ├── articles_test.go
│   ├── open.go              # 跨平台打开浏览器
│   └── cjk_emphasis.go      # CJK 粗体强调修复
├── pkg/
│   ├── alert/               # GitHub Alert 区块 ([!NOTE] 等)
│   ├── details/             # <details> 折叠状态持久化
│   ├── footnote/            # PHP Markdown Extra 脚注
│   ├── ghissue/             # GitHub Issue/PR 引用 (#123)
│   ├── highlighting/        # 代码语法高亮 (chroma)
│   ├── mathjax/             # LaTeX 数学公式 ($...$ / $$...$$)
│   └── tasklist/            # GFM 任务列表 (- [x])
├── defaults/
│   └── embed.go             # //go:embed 嵌入模板和静态资源
├── docs/
│   └── implementation-plan.md  # 实现路线图
├── .github/workflows/
│   ├── build.yml            # CI 构建工作流
│   └── release.yml          # 发布工作流
├── go.mod / go.sum
├── flake.nix / flake.lock   # Nix 构建配置
├── mise.toml                # mise 工具版本管理
├── .pre-commit-config.yaml  # pre-commit 钩子
└── README.md
```

---

## 三、核心架构

### 渲染管线

```
Markdown 输入
  → goldmark.Parser.Parse() (生成 AST)
    → promoteCJKStrongEmphasis (CJK 粗体修复)
      → collectTOC (提取目录, []TOCEntry)
        → goldmark.Renderer.Render() (输出 HTML)
          → 注入 layout.html 模板 + 侧边栏 + 代码高亮 CSS + 导航
            → HTTP 响应给浏览器 (WebSocket 自动重载)
```

### 核心类型

| 类型 | 文件 | 职责 |
|---|---|---|
| `Server` | `internal/server.go:21` | HTTP 服务器，持有 Parser、配置、路由 |
| `Parser` | `internal/parser.go:25` | 无状态 Markdown 解析器 |
| `RenderedDocument` | `internal/parser.go:27` | 渲染结果（HTML + TOC） |
| `serveTarget` | `internal/target.go:16` | 服务模式（文件/目录） |
| `Article` | `internal/articles.go:11` | 侧边栏文章模型（树形结构） |
| `htmlStruct` | `internal/server.go:255` | HTML 模板数据模型 |

### 两种服务模式

| 特性 | 单文件模式 | 目录模式 |
|---|---|---|
| 侧边栏 | 隐藏 | 显示所有 `.md` 文件 |
| URL 访问范围 | 仅该文件 | 目录下所有 `.md` |
| 初始页面 | 直接显示文件 | 重定向到第一个文章 |
| 上一篇/下一篇 | 无 | 按排序导航 |
| 递归子目录 | N/A | 可选 (`--recursive`) |

---

## 四、CLI 配置

通过 Cobra 定义在 `cmd/root.go`：

| 标志 | 短形式 | 默认值 | 说明 |
|---|---|---|---|
| `--browser` | `-b` | `true` | 启动后自动打开浏览器 |
| `--host` | `-H` | `"localhost"` | 绑定地址 |
| `--port` | `-p` | `6419` | 监听端口（显式指定时 strict） |
| `--bounding-box` | — | `true` | 页面添加调试边框 |
| `--no-reload` | — | `false` | 禁用热重载 |
| `--recursive` | `-r` | `false` | 递归扫描子目录 |

---

## 五、Markdown 扩展体系

所有 7 个自定义扩展均实现 `goldmark.Extender` 接口，通过 `Extend(m goldmark.Markdown)` 注册。

| 扩展包 | 功能 | 设计模式 | 外部依赖 |
|---|---|---|---|
| `pkg/alert/` | `> [!NOTE/TIP/IMPORTANT/WARNING/CAUTION]` 区块 | AST Transformer + HTML Renderer | 无 |
| `pkg/details/` | `<details>` 折叠状态 sessionStorage 持久化 | Transformer(注入ID) + Renderer(注入JS) | 无 |
| `pkg/footnote/` | `[^1]` 脚注语法 | Block Parser + Inline Parser + Transformer + Renderer | 无 |
| `pkg/ghissue/` | `#123` / `owner/repo#123` 链接 | Inline Parser + Transformer + Renderer + git 探测 | 无 |
| `pkg/highlighting/` | 代码块 chroma 语法高亮 + 复制按钮 | 替换 FencedCodeBlock Renderer | chroma/v2 |
| `pkg/mathjax/` | `$...$` / `$$...$$` / ` ```math` 公式 | Block Parser + Inline Parser + Transformer + 2 Renderers | 无 |
| `pkg/tasklist/` | `- [x]` 任务列表 | Inline Parser + 2 Renderers (复用以有节点) | 无 |

另有 4 个第三方 goldmark 扩展：
- `goldmark-emoji` — Emoji 短代码
- `go.abhg.dev/goldmark/hashtag` — 主题标签
- `go.abhg.dev/goldmark/mermaid` — Mermaid 图表
- `extension.Linkify/Table/Strikethrough` — goldmark 内置

---

## 六、构建与部署

### 构建方式

| 方式 | 命令 |
|---|---|
| 标准 Go | `go build -o bin/go-grip main.go` |
| Go Install | `go install github.com/showgp/go-grip@latest` |
| Nix Flakes | `nix build` |
| mise | `mise run build` |

### 发布目标

跨平台编译 6 种二进制：
- darwin/amd64, darwin/arm64
- linux/amd64, linux/arm64
- windows/amd64, windows/arm64

构建标志: `-trimpath -ldflags="-s -w"`，打包为 `.tar.gz`/`.zip` + `checksums.txt`。

### CI/CD

| 工作流 | 触发 | 操作 |
|---|---|---|
| `build.yml` | push / PR | mise 设置环境 → build → test → gofmt 检查 → golangci-lint |
| `release.yml` | 推送 `v*` tag | setup-go → go test → 6 架构交叉编译 → 创建 GitHub Release |

### 开发工具链 (mise.toml)

| 工具 | 版本 |
|---|---|
| Go | 1.25 |
| golangci-lint | 2.7 |
| pre-commit | latest |

### 资源嵌入

所有静态资源（`templates/` 和 `static/`）通过 Go 1.16 `//go:embed` 编译时嵌入二进制，运行时不依赖外部文件。

---

## 七、关键设计决策

1. **无状态 Parser** — `Parser` 结构体不含状态，所有逻辑在方法中，简化并发和测试
2. **标准库优先** — 大量使用 `net/http`、`html/template`、`os/exec`、`embed`，减少外部依赖
3. **CJK 特殊处理** — `cjk_emphasis.go` 在 AST 阶段修复 goldmark 对中日韩文字与全角标点的强调解析问题，是项目中最复杂的非标准逻辑
4. **端口回退机制** — 默认端口被占用时自动尝试 +1~+99，仅当用户显式指定 `--port` 时才严格报错
5. **文件名即标题** — 侧边栏显示文件名而非从内容提取标题，避免解析所有文件的性能开销
6. **TOC 独立生成** — 每篇文章单独生成 TOC，不做多文件合并，保持结构和 URL 锚点一致性

---

## 八、实现完成度 (基于 `docs/implementation-plan.md`)

| 阶段 | 内容 | 状态 |
|---|---|---|
| Phase 1 | 文件/目录模式、文章发现 | ✅ 完成 |
| Phase 2 | TOC 元数据提取、标题锚点 | ✅ 完成 |
| Phase 3 | 侧边栏/TOC 模板、CSS 布局 | ✅ 完成 |
| Phase 4 | 路由重构、默认页、空状态 | ✅ 完成 |
| Phase 5 | 端口回退、浏览器 URL | ✅ 完成 |
| Phase 6 | README 更新、手动测试 | ✅ 完成 |
| Phase 7 | 平滑滚动、活跃高亮、减少动效 | ✅ 完成 |
| Phase 8 | 递归目录侧边栏树 | ✅ 完成 |

---

## 九、测试覆盖

| 包 | 测试框架 | 关键测试用例 |
|---|---|---|
| `internal/parser` | testify | CJK 强调 6 用例、基础渲染 |
| `internal/server` | testify | HTTP 响应、侧边栏、路由 |
| `internal/target` | testify | 文件/目录/空参数模式 |
| `internal/articles` | testify | 发现、排序、导航、递归 |
| `internal/listener` | testify | 端口监听与回退 |
| `pkg/alert` | testify | 5 类型 + 嵌套内容 + 边界 |
| `pkg/details` | testify | ID 生成 + 确定性 + 前缀配置 |
| `pkg/footnote` | testify/testutil | 数据驱动测试 + 6 种选项 |
| `pkg/ghissue` | testify | 外部/内部/混合/边界引用 |
| `pkg/highlighting` | testify | Go/C++/HTTP 高亮 + 复制按钮 |
| `pkg/mathjax` | testify | 行内/块级/代码块/回归 |
| `pkg/tasklist` | testify | 勾选/未勾选/有序列表 |

---

## 十、依赖总览

```
goldmark v1.7.16 (核心渲染引擎)
├── goldmark-emoji v1.0.6 (emoji)
├── go.abhg.dev/goldmark/hashtag v0.4.0 (#标签)
├── go.abhg.dev/goldmark/mermaid v0.6.0 (图表)
├── alecthomas/chroma/v2 v2.14.0 (语法高亮)
│   └── dlclark/regexp2 v1.11.4 (增强正则)
├── spf13/cobra v1.8.1 (CLI)
│   └── spf13/pflag v1.0.5
├── aarol/reload v1.2.0 (热重载)
│   ├── gorilla/websocket v1.5.3
│   ├── fsnotify/fsnotify v1.8.0
│   └── bep/debounce v1.2.1
├── forPelevin/gomoji v1.3.0 (emoji 辅助)
└── rivo/uniseg v0.4.7 (Unicode 分段)
```

---

## 十一、开放问题与改进方向

1. **侧边栏标题来源** — 目前显示文件名，可考虑从文件首个 `h1` 推导
2. **隐藏文件策略** — 是否默认忽略 `.` 开头的文件和目录
3. **没有环境变量配置** — 当前所有配置仅通过 CLI 标志，缺少 env var 支持
4. **测试未覆盖性能** — 缺少对大型文档/目录的性能基准测试
5. **无 Docker 支持** — 未提供 Dockerfile 或容器化方案
6. **CJK 强调修复缺少基准测试** — 这部分复杂逻辑仅有功能性测试
