# 网页编辑功能可行性调查报告

> 目标：在当前预览页面添加编辑按钮，点击后可在浏览器内编辑 Markdown 并保存回文件

---

## 结论摘要

**完全可行。** 三个关键子系统的调查均给出肯定结论，技术风险低，实现难度中等（预计 2-3 天 MVP）。

### 一键三联的闭环

```
[浏览器点击"Edit"] → 加载原始 Markdown 到编辑器
[用户编辑完成点击"Save"]
    ↓
[POST /save {path, content}] → 服务端校验路径 → 原子写入磁盘
    ↓
[fsnotify 检测到 Write 事件] → 100ms 防抖 → cond.Broadcast()
    ↓
[所有浏览器 WebSocket 收到 "reload"] → location.reload()
    ↓
[go-grip 重新读取 .md → goldmark 渲染 → 返回最新 HTML]
```

---

## 一、后端 API 可行性 — ✅ 可行

### 路由系统

| 现状 | 评估 |
|---|---|
| 使用标准库 `http.ServeMux` (Go 1.22+) | ✅ 原生支持 `"POST /save"` 方法路由 |
| 现有 `/` 通配 catch-all | ✅ 更具体的 `/save` 路径优先匹配 |
| 无 POST/PUT 处理器 | ✅ 新增一个 handler 即可 |

### 新增端点

```go
// 注册（在 newHandlerForTarget 中）
mux.HandleFunc("POST /save", s.handleSave(target))
mux.HandleFunc("GET /raw/", s.handleRaw(target))
```

### 保存逻辑

```
POST /save
  Body: {"path": "README.md", "content": "# 新内容"}
  ↓
1. cleanRequestPath(path)           -- 规范化路径
2. 检查 isMarkdownFile(path)        -- 仅允许 .md
3. safeJoin(rootDir, path)          -- 路径穿越防护
4. atomicWriteFile(absPath, content) -- 原子写入
5. 返回 {"status": "ok"}
```

### 安全防护（四层）
1. 路径穿越: `path.Clean` → `filepath.Abs` → `strings.HasPrefix` 验证在 rootDir 内
2. 文件类型: 仅允许 `.md` 扩展名
3. 请求限制: `http.MaxBytesReader(w, r.Body, 10MB)`
4. 并发安全: per-file `sync.Mutex` + 原子写入(写临时文件 → `os.Rename`)

---

## 二、热重载联动 — ✅ 自动触发

### 已有机制（零改动）

`github.com/aarol/reload v1.2.0` 的完整工作链：

```
保存 .md 文件到磁盘
  → fsnotify 捕获 Write 事件 (亚毫秒)
    → 100ms debounce 防抖
      → cond.Broadcast() 唤醒所有 WebSocket 协程
        → 每个浏览器连接收到 "reload" 消息
          → window.location.reload()
            → 重新请求 go-grip → goldmark 渲染 → 新 HTML
```

整个过程约 **100~150ms**，所有已打开的浏览器标签页**同步刷新**。

### 注意事项

- `reload` 库**无公开的编程式触发 API**（`cond` 字段未导出），但通过实际写入文件会自动触发 fsnotify，完全不需要额外编程
- 原子写入（临时文件 + `os.Rename`）可能触发 `IN_MOVED_TO` 而非 `IN_MODIFY`，但 reload 库的 `Rename` 事件处理也存在 (`WatchDirectories` 中)
- 写入后双重 `no-cache` 头保障：reload 中间件设置 `Cache-Control: no-cache` + go-grip 自身设置 `no-store`

---

## 三、前端编辑 UI — ✅ 可行

### 模板架构

| 维度 | 评估 |
|---|---|
| 唯一模板 `layout.html` (131行) | ✅ 易修改，内容区可自由添加元素 |
| 数据模型 `htmlStruct` | ✅ 可扩展字段（Editable, CurrentFile） |
| 主题系统 `data-theme` + `themechange` | ✅ 编辑器可直接复用 |
| 现有 JS 均 IIFE 封装 | ✅ 无全局冲突风险 |

### 推荐方案

**MVP (1天)：用 `<textarea>`**
- `htmlStruct` 新增 `Editable bool` + `CurrentFile string`
- `layout.html` 内容区添加"Edit/Save"按钮 + `<textarea>`（默认隐藏）
- 新 JS 用 fetch 获取原始内容 / 保存
- 新 CSS 加编辑区样式

**进阶 (2-3天)：集成 CodeMirror 6**
- 通过 esbuild 打包为单文件放入 `static/js/`
- `//go:embed static` 自动包含，零额外构建配置
- 体积约 300KB，项目已含 ~4MB 三方 JS，可接受
- 监听 `themechange` 事件自动切换深色/浅色主题
- 键盘快捷键 Ctrl+S 保存

### 新增/修改文件清单

| 文件 | 操作 | 说明 |
|---|---|---|
| `internal/server.go` | 修改 | `htmlStruct` 加字段；`newPageData` 传值；`newHandlerForTarget` 注册新路由 |
| `internal/server.go` | 新增 | `handleSave()`, `handleRaw()`, `safeJoin()`, `atomicWriteFile()` |
| `defaults/templates/layout.html` | 修改 | 内容区添加编辑按钮和编辑器容器 |
| `defaults/static/js/editor.js` | 新增 | 编辑/保存交互逻辑 |
| `defaults/static/css/editor.css` | 新增 | 编辑器样式（可无） |

### 边界与兼容性

- `article-nav.js` 已排除 `isContentEditable` 和 `input/textarea` 区域 → 编辑模式下键盘导航自动禁用
- 所有新 CSS 使用 `.docs-editor-*` 命名空间，不与现有选择器冲突
- 单文件模式（无侧边栏）也可添加编辑按钮

---

## 四、实施建议

### Phase 1 — MVP
1. 后端: POST `/save` + GET `/raw/` 端点（含路径安全校验）
2. 模板: "Edit" 按钮切换 textarea 编辑模式
3. 前端 JS: fetch 加载原始内容 → textarea 编辑 → fetch POST 保存
4. 保存后利用已有 reload 机制自动刷新预览

### Phase 2 — 增强
1. 集成 CodeMirror 6（语法高亮、主题切换）
2. Ctrl+S 快捷键保存
3. 未保存更改离开警告
4. 编辑状态与预览状态切换动画

### Phase 3 — 完善
1. 保存成功/失败 toast 提示
2. 并发编辑冲突检测
3. 可选的只读/编辑权限控制
