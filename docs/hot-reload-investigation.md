# 热重载机制调查

## 一、现状描述

### 1.1 后端监视机制
- 文件监视由第三方库 `github.com/aarol/reload v1.2.0` 实现
- `internal/server.go:84-85` 创建 `reload.Reloader` 实例，传入根目录：
  ```go
  reloadMiddleware = reload.New(target.rootDir)
  ```
- `internal/server.go:96` 通过 `reloadMiddleware.Handle(handler)` 将 reload 中间件注入 HTTP 处理链
- `Handle()` 内部在首次调用时启动 `reload.WatchDirectories()`（`reload.go:103`），该方法使用 `fsnotify` 递归监视所有子目录
- `watch.go:67-94` 的事件循环捕获 `fsnotify.Write`、`fsnotify.Create`、`fsnotify.Rename`、`fsnotify.Remove` 事件。对于 Write 和 Create，直接调用 `debounce(callback(path.Base(e.Name)))`
- `callback()`（`watch.go:57-65`）不做任何过滤，直接调用 `reload.cond.Broadcast()` 唤醒所有等待的 WebSocket 连接

### 1.2 WebSocket 通知机制
- `reload.go:173-193` 定义了 `InjectedScript()`，生成一个 `<script>` 标签，内含 WebSocket 客户端代码
- `reload.go:108-137` 的 `Handle()` 中间件在每个 HTTP 响应的末尾注入该脚本（仅当 Content-Type 为 text/html 时）
- WebSocket 端点默认为 `/reload_ws`（`reload.go:89`）
- 当 `reload.cond.Broadcast()` 被调用后，`ServeWS()`（`reload.go:143-164`）从 `Wait()` 中返回，向 WebSocket 客户端发送 `"reload"` 文本消息，然后关闭连接
- 每发一次 reload 信号，旧的 WebSocket 连接就关闭；客户端的 `onclose` 回调会触发 `retry()`，在 1 秒后重新连接

### 1.3 前端重载处理
- 注入的脚本（`reload.go:176-193`）：
  ```javascript
  ws.onmessage = function(msg) {
      if(msg.data === "reload") {
          window.location.reload()
      }
  }
  ```
- 收到 `"reload"` 消息后立即执行 `window.location.reload()`，导致页面完全刷新
- `defaults/static/js/editor.js` 中没有与这个外部 reload 机制交互的逻辑，也没有阻止 reload 的代码（除了 `beforeunload` 事件处理器）

## 二、Issue 1: 非 .md 文件也触发刷新

### 2.1 根因
`aarol/reload` 库的 `WatchDirectories()` 在 `watch.go:71-93` 中处理所有 `fsnotify` 事件，**没有任何文件扩展名过滤**。该库不提供任何过滤机制。`server.go` 创建 Reloader 时也未添加任何过滤层。

具体调用链：
1. `fsnotify` 上报任意文件的 Write/Create 事件 → `watch.go:82-83`
2. `debounce(callback(path.Base(e.Name)))` → `watch.go:57-65`
3. `reload.cond.Broadcast()` 无条件下发 reload 信号 → `watch.go:63`
4. 所有 WebSocket 客户端收到 `"reload"` → 浏览器刷新

### 2.2 受影响的文件/行号
- `internal/server.go:84-85` — 创建 Reloader，未指定过滤
- `internal/server.go:96` — 使用未过滤的中间件包裹 handler
- `/home/ray/go/pkg/mod/github.com/aarol/reload@v1.2.0/watch.go:67-94` — 事件循环无扩展名过滤
- `/home/ray/go/pkg/mod/github.com/aarol/reload@v1.2.0/watch.go:57-65` — 回调函数无条件 broadcast

### 2.3 修复建议（不实施）
`aarol/reload` 库不支持文件扩展名过滤，因此有两种方案：

**方案 A（推荐）**：在 `internal/server.go` 中自定义文件监视，取代直接使用 `reload.Handle()`。
- 创建自己的 `fsnotify.Watcher`，在事件回调中检查 `e.Name` 是否以 `.md` 结尾
- 如果是 `.md` 文件，调用 `reloadMiddleware.cond.Broadcast()` 触发 reload
- 保留 `reloadMiddleware.ServeWS()` 和注入脚本的能力，仅替换监视逻辑

**方案 B**：Fork `aarol/reload` 库，在 `WatchDirectories()` 中添加 `IncludeExts []string` 配置，在事件循环中检查文件扩展名。

## 三、Issue 2: 编辑模式下自动刷新导致退出编辑

### 3.1 根因
存在**两条独立的 reload 路径**，都会导致 `window.location.reload()`，从而丢失编辑状态：

**路径 1 — 外部 WebSocket reload（无意破坏）**：
1. 任意文件变化（含 Issue 1 的误触发）→ `reload.cond.Broadcast()`
2. WebSocket 发送 `"reload"` → `reload.go:162`
3. 注入脚本执行 `window.location.reload()` → `reload.go:187`
4. 页面完全刷新，编辑器状态（textarea 内容、isDirty、isEditing）全部丢失

**路径 2 — saveContent() 后的主动 reload（有意但也退出编辑）**：
1. 用户点击 Save 或 Ctrl+S → `editor.js:121`
2. POST 到 `/api/edit/` 保存文件
3. 600ms 后执行 `window.location.reload()` → `editor.js:147-149`
4. 同样退出编辑模式

**保护机制失效**：
- `editor.js:44-48` 的 `beforeunload` 处理器理论上应在 `isDirty` 时阻止页面离开
- 但现代浏览器对 `beforeunload` 行为不一致：可能忽略 `preventDefault()`，或显示通用提示而非自定义消息
- 对于 `window.location.reload()` 的调用，部分浏览器可能不触发 `beforeunload` 事件
- 即使触发，用户若选择继续刷新，未保存的编辑内容（textare 中的文本）也会丢失

**额外问题**：`editor.js:90` 在进入编辑模式后启动 `pollTimer = setInterval(checkExternalChanges, 5000)` 每 5 秒轮询文件变化。但 WebSocket reload 的延迟远低于 5 秒，实际在 `checkExternalChanges()` 的 `confirm` 对话框弹出之前，WebSocket reload 就已经先触发了页面刷新。

### 3.2 受影响的文件/行号
- `defaults/static/js/editor.js:147-149` — `saveContent()` 保存后执行 `window.location.reload()`
- `defaults/static/js/editor.js:44-48` — `beforeunload` 处理器（保护机制但不可靠）
- `defaults/static/js/editor.js:90` — `pollTimer = setInterval(checkExternalChanges, 5000)`（5 秒轮询与即时 WebSocket 冲突）
- `/home/ray/go/pkg/mod/github.com/aarol/reload@v1.2.0/reload.go:187` — 注入脚本执行 `window.location.reload()`
- `/home/ray/go/pkg/mod/github.com/aarol/reload@v1.2.0/reload.go:162` — WebSocket 发送 `"reload"`

### 3.3 修复建议（不实施）

修复 Issue 1（仅 .md 触发 reload）是解决 Issue 2 的前提，因为非 .md 文件变化是编辑状态下最频繁的外部触发源。

**针对外部 reload（路径 1）**：
- 在 `aarol/reload` 注入脚本的 `onmessage` 回调中，在 `window.location.reload()` 之前检查编辑器状态。例如用 `document.body.getAttribute("data-editing")` 判断是否在编辑中
- 如果在编辑中，不执行 reload 而改用 Toast 提示"文件已更新，请手动刷新"；或者用 AJAX 更新预览区域而非整页刷新

**针对 save 后的 reload（路径 2）**：
- 保存后改为 AJAX 请求重新渲染 Markdown，用返回的 HTML 替换预览区内容，而不是 `window.location.reload()`
- 这样可以保留编辑模式、滚动位置和未保存的更改

**改进 beforeunload 保护**：
- 当前 `beforeunload` 只检查 `isDirty`，建议配合 `sessionStorage` 暂存 textarea 内容，在刷新后自动恢复

## 四、补充发现

1. **`aarol/reload` 库的 `OnReload` 回调无参数**：`reload.go:55-56` 定义了 `OnReload func()`，没有参数传递文件路径。即使想在服务端做过滤也无法知道是哪个文件触发的变化。这使得在现有 API 上做扩展名过滤不可行，必须绕过或修改库本身。

2. **注入脚本版本号检查**：`reload.go:144-151` 检查 WebSocket 连接 URL 中的 `v` 参数是否匹配 `wsCurrentVersion`。版本号硬编码为 `"1"`，不匹配时仅日志警告，不影响功能。

3. **编辑器轮询与 WebSocket 双重检测**：`editor.js:90` 的 `checkExternalChanges()` 每 5 秒轮询 `/api/raw/` 来检测外部修改，同时 `aarol/reload` 的 WebSocket 也在监听文件变化。这两个机制互相独立且可能冲突——WebSocket reload 会直接刷新页面，导致 `checkExternalChanges` 对话框永远不会显示。

4. **`saveContent()` 中的 reload 可以通过 `beforeunload` 避免**：如果保存后不执行 `window.location.reload()`，而是改为局部刷新渲染区，不仅能保留编辑状态，还能避免不必要的整页加载。

5. **`defaults/templates/layout.html` 中没有主动引用 reload 脚本**：因为 `aarol/reload` 是通过中间件自动注入脚本的，模板本身无需修改。这意味着如果要更换 reload 方案，只需调整 `internal/server.go` 中的中间件逻辑，不需要改模板。
