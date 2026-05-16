# 网页编辑功能 — 分阶段实现规划

> 基于 `editing-requirements.md` (需求) 和 `editing-spec.md` (规格)

---

## 总体概览

| 阶段 | 目标 | 预计工期 | 需求覆盖 |
|------|------|---------|----------|
| **Phase 1: MVP** | 基本编辑闭环：编辑→保存→刷新 | 0.5-1天 | FR-01~FR-16、NFR-01~NFR-04、NFR-06~NFR-09、SM-01~SM-04 |
| **Phase 2: 增强** | 键盘快捷键 + 外部变更检测 + 滚动恢复 | 0.5-1天 | FR-11、FR-13、FR-15、EC-01~EC-06、AC-ENH-05 |
| **Phase 3: 完善** | 安全加固 + 并发保护 + 跨平台验证 | 0.5-1天 | NFR-05、EC-07 |

**核心设计原则**: 每阶段独立可发布、可验证，不做过度工程。

---

## 目录结构总览

```
go-grip/
├── internal/
│   └── server.go               # ← 核心变更 (新增 ~140 行)
├── defaults/
│   ├── templates/
│   │   └── layout.html         # ← 模板变更 (新增 ~15 行)
│   └── static/
│       ├── js/
│       │   └── editor.js       # ← 新增文件 (~180 行)
│       └── css/
│           └── editor.css      # ← 新增文件 (~45 行)
└── internal/
    └── server_test.go          # ← 测试变更 (新增 ~90 行)
```

---

## Phase 1: MVP（基本编辑闭环）

> **目标**: 用户能在预览页面点击编辑按钮，切换到编辑模式修改 Markdown 源码，保存回文件后自动刷新预览。
>
> **验收**: AC-MVP-01 ~ AC-MVP-09（9 条）

### Task 1.1 — 扩展模板数据模型

**文件**: `internal/server.go`

**修改点**:

| 位置 | 变更 | 说明 |
|------|------|------|
| `:266` `htmlStruct` 结构体末尾 | 新增 2 个字段 | `CurrentFile string` — 当前文件相对路径 |
| | | `RawContent string` — 原始 Markdown 源码 |
| `:212` `newPageData()` 函数签名 | 新增参数 `rawContent string` | 接收原始 Markdown 文本 |
| `:235` 函数内 return 结构体 | 赋值新增的 2 个字段 | `CurrentFile: currentFile`<br>`RawContent: rawContent` |
| `:188` `newPageData()` 调用点 | 传入 `string(bytes)` | `bytes` 已在行 177 读取 |
| `:202` `renderEmpty()` 调用点 | 传入 `""` | 空页面时无原始内容 |

**设计说明**: `CurrentFile` 传递给前端用于保存 API 定位文件；`RawContent` 用于编辑器初始内容填充。

### Task 1.2 — 注册保存与读取 API 路由

**文件**: `internal/server.go`

**修改点**:

| 位置 | 变更 | 说明 |
|------|------|------|
| `:130` 之后 | 新增路由注册语句 | 注册两个新路由端点 |

**新增功能**:

| Handler | 功能 | 输入 | 输出 |
|---------|------|------|------|
| 保存处理器 | 接收文件路径和内容，写入磁盘 | 文件路径 + Markdown 内容 | 操作结果 |
| 读取处理器 | 读取文件的原始 Markdown 文本 | 文件路径 | 原始 Markdown 文本 |

**路径安全校验（内置于保存处理器）**:

```
1. 从请求中提取文件路径
2. 调用 cleanRequestPath() 清理路径（去除 ../ 和 URL 编码）
3. 在 url.PathUnescape 之后再次调用 path.Clean（防御双重编码绕过，如 %2e%2e → ..）
4. 调用 isMarkdownFile() 确保仅 .md 文件
5. 拼接 rootDir + 清理后的路径 → 取绝对路径
6. 调用 filepath.EvalSymlinks() 解析符号链接（防止符号链接指向 rootDir 外部）
7. 验证绝对路径前缀是否在 rootDir 内 ← 防路径穿越
8. 实施请求体大小上限检查（默认 10MB），拒绝超大请求体
9. 通过后执行写入操作
```

**安全函数位置**: 在 `server.go` 约 `:300` 附近新增独立的安全校验函数，供保存处理器开头调用。

### Task 1.3 — 文件写入实现

**文件**: `internal/server.go`

**新增函数位置**: `:253` 之后（紧接 `readToString()` 之后）

**写入策略**: 先写入临时文件，再原子重命名为目标文件

```
1. 检查目标文件是否存在（os.Stat）→ 若不存在返回"文件已不存在"的专用错误
2. 创建临时文件（路径 = 目标路径 + ".tmp" 后缀）
3. 将内容写入临时文件
4. 若写入失败 → 删除临时文件 → 返回错误
5. 若写入成功 → os.Rename(临时文件, 目标文件) ← 原子操作
6. 返回成功
```

**并发控制**: 使用全局的 per-file 互斥锁（map[string]*sync.Mutex），同一文件同时只有一个写入操作。

### Task 1.4 — 模板变更：编辑按钮 + 编辑器容器

**文件**: `defaults/templates/layout.html`

**修改点**:

| 位置 | 变更 | 说明 |
|------|------|------|
| `:63` `<body>` 标签 | 新增 `data-current-file` 属性 | 将 `CurrentFile` 传递到前端 |
| `:51` 之后 | 新增 CSS 引用 | 引入 `editor.css` |
| `:60` 之后 | 新增 JS 引用 | 引入 `editor.js` |
| `:81-84` `<main>` 内部 | 新增编辑工具栏 | 在 `.container` 内、`{{ .Content }}` 之上 |
| `:82` 之后、`:83` 之前 | 新增编辑器容器 | 初始状态隐藏 |

**模板中新增的 HTML 结构**:

```html
<!-- 编辑工具栏 -->
<div class="editor-toolbar" style="display:none">
  <button class="editor-btn editor-btn-preview">预览</button>
  <button class="editor-btn editor-btn-save">保存</button>
  <button class="editor-btn editor-btn-cancel">取消</button>
</div>

<!-- 预览工具栏 -->
<div class="preview-toolbar">
  <button class="editor-btn editor-btn-edit">编辑</button>
</div>

<!-- 编辑器容器（默认隐藏） -->
<div class="editor-container" style="display:none">
  <textarea class="editor-textarea"></textarea>
</div>

<!-- 预览内容 -->
<div class="preview-content">{{ .Content }}</div>
```

### Task 1.5 — 前端编辑逻辑

**文件**: `defaults/static/js/editor.js`（新增）

**核心交互流程**:

```
1. 页面加载:
   - 从 <body data-current-file> 读取当前文件路径
   - 检测当前页面是否为 .md 文件（通过 data-current-file 扩展名判断）
     - 若非 .md 文件或 data-current-file 为空 → 隐藏"编辑"按钮
   - 绑定"编辑"按钮点击事件

2. 点击"编辑"按钮:
   - 进入编辑模式前检查文件大小（通过读取 API 的响应头获取字节数）
     - 若超限 → 弹窗提示"文件过大，建议使用本地编辑器" → 拒绝进入编辑
   - 向 <body> 添加 data-editing 属性（用于 article-nav.js 全局禁用导航）
   - 隐藏预览内容 (.preview-content)
   - 隐藏预览工具栏
   - 显示编辑工具栏 (.editor-toolbar)
   - 显示编辑器容器 (.editor-container)
   - 向服务器请求原始 Markdown 文本
   - 将文本填充到 <textarea>
   - 聚焦编辑器

3. 点击"保存"按钮:
   - 从 <textarea> 读取内容
   - 构建请求体（文件路径 + 内容）
   - "保存"按钮变为禁用状态 + 显示"保存中..."
   - 发送保存请求到服务器
   - 等待服务器响应:
     - 成功(200) → 显示"已保存"(1-3秒 toast) → 记录当前滚动位置到 sessionStorage → 主动触发页面刷新
     - 失败 → 显示错误信息（使用用户可理解的语言，不暴露内部路径）→ 保持编辑器内容不变 → 恢复按钮可用状态

4. 检测热重载是否启用:
   - 若启用 (正常模式): 等待 WebSocket 连接存在 → 保存成功后依赖自动刷新
   - 若未启用 (--no-reload 模式): WebSocket 不可用时，保存成功后主动刷新；若手动刷新不便，显示"已保存，请手动刷新页面"提示

5. 点击"取消"按钮:
   - 检测内容是否有修改（与原始文本对比）
   - 若有修改 → 弹出确认对话框
   - 若无修改 → 直接返回预览模式
   - 移除 <body> 上的 data-editing 属性
   - 显示预览内容，隐藏编辑器

6. 点击"预览"按钮:
   - 同"取消"逻辑（不保存，直接返回预览）

7. 编辑中点击侧边栏切换到其他文件:
   - 若内容有未保存修改 → 阻止导航 → 弹出三选项自定义对话框：
     a. "保存并切换" → 保存 → 成功后导航到新文件
     b. "放弃并切换" → 丢弃修改 → 导航到新文件
     c. "继续编辑" → 关闭对话框，留在当前编辑状态
   - 若无修改 → 直接导航
```

**封装方式**: 使用 IIFE 模式 `(function() { ... })()` 与现有 JS 风格一致。

**错误处理**: 网络请求失败时显示错误提示，恢复按钮可用状态。

### Task 1.6 — 编辑器样式

**文件**: `defaults/static/css/editor.css`（新增）

**样式需求**:

| 元素 | 样式要求 |
|------|---------|
| `.editor-toolbar` | 固定在内容区顶部，背景与主题一致 |
| `.editor-btn` | 仿 GitHub 按钮风格，支持 hover/active 状态 |
| `.editor-container` | 全宽，最小高度 300px |
| `.editor-textarea` | 等宽字体，全宽，最小高度 300px，带内边距，圆角边框 |
| 深色主题 | 通过 `[data-theme="dark"]` 选择器覆盖背景/文字/边框颜色 |
| 平板适配 | `@media (min-width: 768px) and (max-width: 1023px)` 缩小内边距、调整字体大小 |
| 桌面适配 | `@media (min-width: 1024px)` 编辑器撑满主要内容区可用高度 |

### Task 1.7 — 测试

**文件**: `internal/server_test.go`

**新增测试**:

| 测试场景 | 验证项 |
|---------|--------|
| 保存正常 Markdown 文件 | 文件内容变更 + 响应确认 |
| 保存非 .md 扩展名 | 返回拒绝 |
| 路径穿越攻击（含 `../`） | 返回拒绝 |
| 路径穿越攻击（含符号链接） | 返回拒绝 |
| 保存到不存在的目录 | 返回错误 |
| 保存已被外部删除的文件 | 返回"文件已不存在"错误 |
| 空内容保存 | 文件被清空 |
| 大文件保存（超限） | 返回拒绝，保留编辑器内容 |
| 单文件模式编辑 | 编辑入口可用、CurrentFile 正确 |
| 目录模式编辑 | 侧边栏高亮文件可编辑 |
| 跨浏览器验证 | Chrome/Firefox/Edge 最新两个大版本功能正常 |

### Task 1.8 — 侧边栏状态保持

**文件**: `defaults/static/js/editor.js`

**功能**: 保存触发页面刷新时，保持侧边栏的展开/折叠状态。

```
1. 页面加载时，从 sessionStorage 读取侧边栏状态并恢复
2. 用户手动展开/折叠侧边栏目录节点时，将状态写入 sessionStorage
3. 保存操作前（Task 1.5 步骤 3），将当前侧边栏状态写入 sessionStorage
4. 页面刷新后（重新加载），sidebar-active.js 读取 sessionStorage 恢复状态
```

**设计说明**: 使用 sessionStorage（非 localStorage），仅在同一标签页会话中有效，避免跨标签页的状态污染。

---

## Phase 2: 增强（键盘快捷键 + 外部变更检测 + 滚动恢复）

> **目标**: 提升编辑效率，处理边界场景。
>
> **验收**: AC-ENH-01 ~ AC-ENH-06（6 条）

### Task 2.1 — 键盘快捷键

**文件**: `defaults/static/js/editor.js`

**新增快捷键**:

| 快捷键 | 操作 | 备注 |
|--------|------|------|
| `Ctrl+S` / `Cmd+S` | 保存 | 阻止浏览器默认行为（保存网页） |
| `Esc` | 取消编辑 | 返回预览模式（含未保存确认） |
| `Ctrl+Enter` | 保存并留在编辑模式 | 快速迭代编辑 |

**实现要点**:
- 在编辑模式下全局监听 `keydown` 事件
- `Ctrl+S` 需要 `e.preventDefault()` 阻止浏览器默认行为
- 仅在编辑模式激活时注册，退出时移除
- 与 `article-nav.js` 的键盘事件互不干扰：编辑器 `<textarea>` 已被 `isEditableTarget()` 排除

### Task 2.2 — 外部文件变更检测

**文件**: `defaults/static/js/editor.js`

**场景**: 用户在编辑器中编辑时，外部程序修改了同一文件。

**处理逻辑**:

```
1. 进入编辑模式时，记录初始文件内容
2. 定期检查服务器上的文件内容（轮询或通过 WebSocket）
3. 若检测到外部变更：
   a. 弹出提示对话框："文件已被外部修改，是否覆盖？"
   b. 用户选择"覆盖" → 保持当前编辑内容不变
   c. 用户选择"重新加载" → 丢弃当前编辑，加载新内容到编辑器
```

### Task 2.3 — 未保存离开提醒

**文件**: `defaults/static/js/editor.js`

**处理逻辑**:

```
1. 跟踪是否有未保存的修改
2. 通过 window.onbeforeunload 事件拦截页面关闭/刷新
3. 显示浏览器原生确认对话框
```

### Task 2.4 — 保存后滚动位置恢复

**文件**: `defaults/static/js/editor.js`

**功能**: 保存触发全页刷新后，恢复到刷新前用户的浏览位置。

```
1. 保存前（Phase 1 Task 1.5 步骤 3 已完成），将 document.documentElement.scrollTop 写入 sessionStorage
2. 页面加载时，从 sessionStorage 读取存储的滚动位置
3. 若存在且在合理范围内，通过 window.scrollTo() 恢复
4. 恢复后清除 sessionStorage 中的记录，避免下次非保存刷新误恢复
```

**设计说明**: 利用 Phase 1 已写入 sessionStorage 的 scrollTop 值，页面重新渲染后恢复到近似位置。

---

## Phase 3: 完善（安全加固 + 并发保护 + 跨平台验证）

> **目标**: 生产级安全与稳定性。
>
> **验收**: NFR-05、EC-07

### Task 3.1 — Per-file 并发锁

**文件**: `internal/server.go`

**新增**: 使用 `sync.Map` 的 `LoadOrStore` 方法管理 per-file 互斥锁。

```
保存逻辑中的并发保护:
1. 对目标文件路径调用 sync.Map.LoadOrStore → 获取或创建 *sync.Mutex
2. Lock() → 执行原子写入（临时文件 + os.Rename） → Unlock()
3. 使用 defer 确保异常时也会释放锁
```

### Task 3.2 — 跨平台热重载验证

**验证目标**: 确认原子写入（临时文件 + os.Rename）在不同操作系统上均能正确触发 reload 的浏览器刷新。

| 平台 | 文件系统后端 | 验证项 |
|------|------------|--------|
| Linux | inotify | `os.Rename` 产生 `IN_MOVED_TO` → `Create` 事件 → 触发广播 |
| macOS | kqueue | 验证 kqueue 下 `os.Rename` 事件映射是否正确触发广播 |
| Windows | ReadDirectoryChangesW | 验证 Windows 下 `os.Rename` 事件映射是否正确触发广播 |

### Task 3.3 — 文件权限检查

**文件**: `internal/server.go` 保存处理器

**新增**: 写入前检查目标文件/目录的写权限，若不可写则返回明确错误信息。

---

## 变更文件汇总

| 文件 | Phase 1 | Phase 2 | Phase 3 | 总新增/修改行 |
|------|---------|---------|---------|-------------|
| `internal/server.go` | +140 行 (修改) | — | +20 行 (修改) | +160 |
| `defaults/templates/layout.html` | +15 行 (修改) | — | — | +15 |
| `defaults/static/js/editor.js` | +120 行 (新增) | +60 行 (修改) | — | +180 |
| `defaults/static/css/editor.css` | +45 行 (新增) | — | — | +45 |
| `internal/server_test.go` | +70 行 (新增) | — | +20 行 (新增) | +90 |

---

## 热重载联动确认

> 结论来自 `edit-hotreload-analysis` 子 Agent 报告。

### 现有机制完全支持

| 环节 | 状态 | 细节 |
|------|------|------|
| 文件写入 → fsnotify 检测 | ✅ 自动 | `reload.WatchDirectories()` 监听文件事件。原子写入（`os.Rename`）在 Linux 上产生 `Create` 事件（非 `Write`），reload 库的 `Create` 处理器会触发广播 |
| 事件防抖 | ✅ 100ms | `bep/debounce` 合并连续写入为一次广播 |
| 通知广播 | ✅ 全量 | `sync.Cond.Broadcast()` 唤醒所有 WebSocket 连接 |
| 浏览器刷新 | ✅ 全页 | `window.location.reload()` 整页刷新 |
| 渲染更新 | ✅ 最新 | `no-cache` 策略确保不缓存旧版本 |

### 已知限制

- 整页刷新会丢失 JS 状态（已通过 sessionStorage 在 Task 1.8/1.5 中补救侧边栏状态和滚动位置）
- 连续快速保存（间隔 < 100ms）会被防抖合并为一次刷新
- macOS (kqueue) 和 Windows 上 `os.Rename` 的事件映射需在 Task 3.2 中单独验证

### 不需要额外开发的内容

- 不需要手动触发 reload
- 不需要修改 reload 库的任何配置
- 不需要新增 WebSocket 逻辑

### 安全配置现状

- WebSocket 来源检查已设为允许所有（本地开发安全）
- 写入完成后浏览器收到 `Cache-Control: no-store, no-cache, must-revalidate` 头部

---

## 实现依赖关系

```
Task 1.1 (数据模型)
  └── Task 1.4 (模板变更)
Task 1.2 (路由注册+安全校验+大小限制)
  └── Task 1.3 (文件写入)
Task 1.2 + Task 1.3
  └── Task 1.5 (前端逻辑: 编辑/保存/取消/按钮可见性/nav禁用/文件切换)
Task 1.4 + Task 1.5
  ├── Task 1.6 (样式: 编辑器+主题+响应式)
  └── Task 1.8 (侧边栏状态保持)
Task 1.1 ~ 1.6
  └── Task 1.7 (测试: 服务端+浏览器兼容)
Task 1.5 (完成)
  └── Phase 2 (快捷键 + 外部变更检测 + 滚动恢复)
Phase 2 (完成)
  └── Phase 3 (并发锁加固 + 跨平台验证)
```

Phase 1 内部的 Task 1.1+1.4 和 Task 1.2+1.3 可以**并行开发**。
