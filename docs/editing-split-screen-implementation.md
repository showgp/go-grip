# 分屏预览实施计划

> 基于 `editing-split-screen-requirements.md` (需求) 和 `editing-split-screen-review.md` (Review)
> 在现有 `editing-implementation-plan.md` Phase 1-3 之上构建，不修改后端代码

---

## 总体概览

| 阶段 | 目标 | 预计工期 | 需求覆盖 |
|------|------|---------|----------|
| **Phase 1: MVP** | 分屏实时预览核心闭环：marked 嵌入、分屏布局、防抖渲染、Preview 按钮行为变更 | 1-1.5天 | FR-SPLIT-01/03~12/14~16、FR-LIVE-01/02/06、FR-INT-01~08/11/12、FR-DIST-01~03、NFR-SPLIT-01/03~10、AC-SPLIT-01~09 |
| **Phase 2: 增强** | 大文档自适应防抖 + 可拖拽分隔条 | 0.5-1天 | FR-LIVE-04、FR-SPLIT-02、AC-SPLIT-10~11 |
| **Phase 3: 完善** | 空内容占位 + Ctrl+P 快捷键 | 0.5天 | FR-SPLIT-13、FR-INT-10、FR-LIVE-03（由 Task 1.5 togglePreview 展开时 renderPreview 满足，详见 §五）|

**核心设计原则**: 每阶段独立可发布、可验证。Phase 1 完成后即可交付分屏编辑预览的完整用户体验。不做过度工程，不修改后端。

---

## 目录结构总览

```
go-grip/
├── defaults/
│   ├── templates/
│   │   └── layout.html                # ← 修改 (~8 行变更)
│   └── static/
│       ├── js/
│       │   ├── marked.min.js           # ← 新增文件 (第三方库, ~40KB)
│       │   └── editor.js               # ← 修改 (~80 行新增/变更)
│       └── css/
│           └── editor.css              # ← 修改 (~120 行新增)
```

**注意**: `internal/server.go` 零变更——分屏预览是纯客户端功能。`RawContent` 字段已存在 (`server.go:246`)，`embed.FS` 自动包含 `defaults/static/js/` 下新文件。

---

## 一、总体策略

分四个维度推进：

1. **静态资源嵌入**：下载 marked.min.js → 放入 `defaults/static/js/` → layout.html 引用（defer 加载，在 editor.js 之前）
2. **DOM 结构调整**：`.editor-container` 内部新增 `.editor-split-wrapper` 包裹 textarea + 新增 preview div（解决 CSS flex-direction 冲突）
3. **CSS 分屏布局**：利用现有 `[data-editing="true"]` 前缀规则集，在 desktop >=1024px 时 wrapper 为 flex row（textarea/preview 各 50%），<768px 时 column 堆叠
4. **JS 渲染引擎**：`renderPreview()` 调用 `marked.parse()`，防抖 150ms，`requestAnimationFrame` 异步非阻塞；Preview 按钮切换面板可见性

所有 CSS 选择器使用 `.editor-split-*` 命名前缀（FR-SPLIT-16、NFR-SPLIT-07），不修改现有 `.editor-container` 的 flex column 方向。

---

## 二、现状基线

### 2.1 HTML 结构 (`layout.html:84-127`)

当前编辑模式相关的 DOM 结构：

```
div.container                                              (:84)
  div.preview-toolbar > button.editor-btn-edit             (:87-89)
  div.editor-toolbar                                       (:93-97)
    button.editor-btn-preview                              (:94)
    button.editor-btn-save                                 (:95)
    button.editor-btn-cancel                               (:96)
  div.editor-container                                     (:100-102)
    textarea.editor-textarea                               (:101)
  div.preview-content > .container-inner > {{ .Content }}  (:104-108)
  nav.docs-page-nav                                        (:109-126)
```

`[data-editing="true"]` 设置后，CSS 隐藏 `.preview-toolbar`、`.preview-content`、`.docs-sidebar`、`.docs-toc`、`.footer`、`.docs-page-nav`，显示 `.editor-toolbar`、`.editor-container`。

### 2.2 当前 CSS 编辑模式规则 (`editor.css:302-334`)

| 规则 | 行号 | 作用 |
|------|------|------|
| `[data-editing="true"] .docs-main .container` | `:303-307` | `display: flex; flex-direction: column; min-height: calc(100vh - 120px)` |
| `[data-editing="true"] .editor-container` | `:309-313` | `flex: 1; display: flex; flex-direction: column` ← **分屏需在此之上叠加**
| `[data-editing="true"] .editor-textarea` | `:315-318` | `flex: 1; min-height: 0` |
| `[data-editing="true"] .docs-sidebar, .docs-toc, .footer, .docs-page-nav` | `:321-326` | `display: none` |
| `[data-editing="true"] .docs-shell` | `:328-330` | `display: block` |
| `[data-editing="true"] .docs-main` | `:332-334` | `max-width: 100%` |

**关键约束**：`.editor-container` 的 `flex-direction: column` 保持不变（为 `.editor-split-wrapper` 提供 column 容器上下文），wrapper 内部使用 `flex-direction: row` 实现左右分屏。

### 2.3 当前 JS 编辑状态 (`editor.js`)

| 变量/函数 | 行号 | 作用 |
|-----------|------|------|
| `isEditing`, `isDirty`, `originalContent`, `pollTimer` | `:2-6` | 编辑模式全局状态 |
| `init()` | `:12-50` | 绑定按钮事件 — Preview 按钮绑定 `cancelEdit` (**需变更**) |
| `enterEditMode()` | `:52-82` | 加载内容 → textarea → `data-editing="true"` → `toggleUI(true)` |
| `saveContent()` | `:84-119` | POST 保存 → reload |
| `cancelEdit()` | `:121-128` | 未保存确认 → `exitEditMode()` |
| `exitEditMode()` | `:130-139` | 清除 `data-editing` → `toggleUI(false)` → 清除 pollTimer |
| `toggleUI(editMode)` | `:167-177` | 控制 4 个元素的 display |
| `handleKeydown()` | `:179-201` | Ctrl+S / Ctrl+Enter / Esc |
| `textarea.oninput` | `:70-72` | 仅设置 `isDirty` 标记 |

### 2.4 现有 .markdown-body 滚动行为 (`docs-layout.css:1-3`)

```css
.markdown-body {
  scroll-behavior: smooth !important;
}
```

此规则通过 `!important` 强制对所有 `.markdown-body` 后代元素生效，包括预览面板。分屏预览面板需显式覆盖为 `scroll-behavior: auto`。

---

## 三、Phase 1: MVP 核心实现（所有 P0 项）

> **目标**: 用户点击 Edit 进入编辑模式 → 左侧 textarea + 右侧 real-time 预览 → 输入时防抖渲染 → Preview 按钮切换面板可见性 → Save/Cancel 行为不变。
>
> **验收**: AC-SPLIT-01 ~ AC-SPLIT-09（9 条）

### Task 1.1 — 获取并嵌入 marked 库

- **对应需求**: FR-DIST-01, FR-DIST-02, FR-DIST-03, NFR-SPLIT-01, NFR-SPLIT-10
- **对应 AC**: AC-SPLIT-06
- **依赖**: 无
- **修改文件**: `defaults/static/js/marked.min.js`（新增）, `defaults/templates/layout.html`
- **当前代码位置**: `layout.html:62`, `<script src="/static/js/editor.js" defer></script>`
- **实现要点**:

  1. **下载 marked**: 从 jsdelivr CDN 下载固定版本
     ```bash
     curl -Lo defaults/static/js/marked.min.js \
       https://cdn.jsdelivr.net/npm/marked@12.0.2/marked.min.js
     ```
     版本：**marked 12.0.2**（MIT 许可证，~40KB 原始 / ~12KB gzip，GFM 完整支持）。

  2. **放置位置**: `defaults/static/js/marked.min.js` — 与 `editor.js`、`mermaid.min.js` 同级。Go 的 `embed.FS`（`defaults.StaticFiles`）会自动包含此文件，无需修改 `server.go`。

  3. **HTML 引用**: 在 `layout.html:62` **之前**插入，使用 `defer` 确保非阻塞加载：
     ```html
     <script src="/static/js/marked.min.js" defer></script>
     ```
     插入后顺序为：
     ```
     :61 <script src="/static/js/article-nav.js"></script>
     :62 <script src="/static/js/marked.min.js" defer></script>   ← 新增
     :63 <script src="/static/js/editor.js" defer></script>
     ```

  4. **标记为第三方文件**: `marked.min.js` 是 marked 官方发布产物，不做 go-grip 特定修改。文件头部保留原始注释。如需定制解析行为，在 `editor.js` 中通过 `marked.use()` 扩展 API 实现（当前版本无需定制）。

  5. **全局符号**: marked 在 `marked.min.js` 中注册全局 `marked.parse(markdownString)` 函数。`editor.js` 通过 `defer` 保证在 `marked` 可用之后执行。

  6. **提交到仓库**: `git add defaults/static/js/marked.min.js` — 确保离线可用（FR-DIST-01）。

### Task 1.2 — HTML 结构：新增 split-wrapper 和 preview div

- **对应需求**: FR-SPLIT-01, FR-SPLIT-15, FR-SPLIT-16, NFR-SPLIT-07
- **对应 AC**: AC-SPLIT-01（分屏编辑闭环中的布局部分）
- **依赖**: Task 1.1（可并行，但运行期需 marked 可用）
- **修改文件**: `defaults/templates/layout.html`
- **当前代码位置**: `layout.html:100-102`, `.editor-container` 区域
- **当前代码**:
  ```html
  <!-- 编辑器区域（默认隐藏） -->
  <div class="editor-container" style="display:none">
    <textarea class="editor-textarea" aria-label="Markdown editor"></textarea>
  </div>
  ```
- **实现要点**:

  1. 在 `.editor-container` 内部新增 `.editor-split-wrapper` 作为 flex row 容器，包裹 textarea 和新增的 preview div：
     ```html
     <!-- 编辑器区域（默认隐藏） -->
     <div class="editor-container" style="display:none">
       <div class="editor-split-wrapper">
         <textarea class="editor-textarea" aria-label="Markdown editor"></textarea>
         <div class="editor-split-preview" aria-label="Live Markdown preview"></div>
       </div>
     </div>
     ```

  2. **设计理由**（解决 Review 问题 5）：
     - `.editor-container` 保持 `flex-direction: column` 不变（`editor.css:311`）
     - `.editor-split-wrapper` 在 column 容器中作为唯一子元素，通过 `flex: 1` 填充高度
     - wrapper 内部 `flex-direction: row` 实现左右分屏，不破坏现有 textarea `flex: 1` 的 column 假设
     - 当 preview 面板隐藏时（`no-preview` class），wrapper 仍为 flex row 但 preview 为 `display: none`，textarea 自然占据 100%

  3. **命名隔离**: `.editor-split-wrapper` 和 `.editor-split-preview` 使用独立于 `.preview-content` 的命名前缀（NFR-SPLIT-07、FR-SPLIT-16），避免与现有服务端渲染容器产生 CSS 或 JS 选择器冲突。

  4. **ARIA**: preview div 添加 `aria-label="Live Markdown preview"` 提升可访问性。preview div 初始为空，进入编辑模式时由 JS 填充。

### Task 1.3 — CSS 分屏布局与预览面板样式

- **对应需求**: FR-SPLIT-01, FR-SPLIT-03, FR-SPLIT-04, FR-SPLIT-10, FR-SPLIT-12, FR-SPLIT-14, NFR-SPLIT-06
- **对应 AC**: AC-SPLIT-03, AC-SPLIT-04, AC-SPLIT-09
- **依赖**: Task 1.2（split-wrapper 和 preview div 已存在于 DOM）
- **修改文件**: `defaults/static/css/editor.css`
- **当前代码位置**: 
  - `editor.css:309-313` — `[data-editing="true"] .editor-container` 规则（column flex）
  - `editor.css:315-318` — `[data-editing="true"] .editor-textarea` 规则（`flex: 1`）
  - `editor.css:321-326` — hide sidebar/TOC/footer/page-nav 规则
- **实现要点**:

  1. **Split wrapper 基础规则** — 在 `editor.css:318` 之后（`[data-editing="true"] .editor-textarea` 规则块之后）新增：
     ```css
     /* --- Split-screen preview layout --- */
     [data-editing="true"] .editor-split-wrapper {
       display: flex;
       flex: 1;
       flex-direction: row;
       min-height: 0;
     }

     [data-editing="true"] .editor-split-wrapper .editor-textarea {
       flex: 1;
       min-width: 0;
       min-height: 0;
       resize: none;
       border-top-right-radius: 0;
       border-bottom-right-radius: 0;
     }

[data-editing="true"] .editor-split-preview {
        flex: 1;
        min-width: 0;
        overflow-y: auto;
        overflow-x: hidden;
        scroll-behavior: auto !important;   /* 覆盖 docs-layout.css:2 的 smooth !important (Review 问题 13) */
        padding: 16px;
        border: 1px solid #d0d7de;
        border-left: none;
        border-radius: 0 6px 6px 0;
        background: #ffffff;
        color: #1f2328;
        word-wrap: break-word;
      }
     ```

  2. **Preview 面板可见性切换** — Preview 按钮隐藏面板时给 wrapper 添加 `no-preview` class：
     ```css
     /* Preview hidden: textarea full width */
     [data-editing="true"] .editor-split-wrapper.no-preview .editor-split-preview {
       display: none;
     }

     [data-editing="true"] .editor-split-wrapper.no-preview .editor-textarea {
       resize: vertical;
       border-top-right-radius: 6px;
       border-bottom-right-radius: 6px;
     }
     ```

3. **Preview 按钮激活态样式** — 仅在编辑模式下生效：
      ```css
      [data-editing="true"] .editor-btn-preview.active {
        background-color: #ddf4ff;
        border-color: #54aeff;
        color: #0969da;
      }
      ```

  4. **响应式规则** — `<768px` 堆叠到下方（FR-SPLIT-03）：
     ```css
     @media (max-width: 767px) {
       [data-editing="true"] .editor-split-wrapper {
         flex-direction: column;
       }

       [data-editing="true"] .editor-split-wrapper .editor-textarea {
         border-top-right-radius: 6px;
         border-bottom-right-radius: 0;
         border-bottom-left-radius: 0;
         resize: vertical;
       }

       [data-editing="true"] .editor-split-preview {
         border-left: 1px solid #d0d7de;
         border-top: none;
         border-radius: 0 0 6px 6px;
         min-height: 200px;
         max-height: 50vh;
       }
     }
     ```
     ≥768px 不添加额外 media query（现有 flex row 在 768-1023 范围内自动保持可用，不断裂）。

5. **深色/浅色主题覆盖** — 由 Task 1.7 统一处理。预览面板内的 Markdown 元素样式通过 `<body class="markdown-body">` 后代选择器自动继承 GitHub 主题 CSS（`github-markdown-light.css` / `github-markdown-dark.css`），面板容器的 `background`/`color`/`border-color` 跟随 `[data-theme]` 和 `prefers-color-scheme` 切换，详见 Task 1.7。

  6. **Preview 面板不响应交互**（FR-SPLIT-09）：通过 CSS 禁用链接点击和选择：
     ```css
     [data-editing="true"] .editor-split-preview a {
       pointer-events: none;
       color: inherit;
       text-decoration: underline;
     }
     ```

  7. **不修改现有规则**: `[data-editing="true"] .editor-container`（`:309-313`）的 `flex: 1; display: flex; flex-direction: column` 保持不变。split-wrapper 作为其唯一子元素继承 column 布局的 flex 空间分配。原有 sidebar/TOC/footer 隐藏规则（`:321-326`）和 `.docs-main` 全宽规则（`:332-334`）完全不触及。

### Task 1.4 — 实时渲染引擎（renderPreview + 防抖）

- **对应需求**: FR-SPLIT-06, FR-LIVE-01, FR-LIVE-02, FR-LIVE-06, NFR-SPLIT-03, NFR-SPLIT-09
- **对应 AC**: AC-SPLIT-01（渲染闭环）, AC-SPLIT-08（异步不阻塞输入）
- **依赖**: Task 1.1（marked 全局符号可用）, Task 1.2（preview div 存在于 DOM）
- **修改文件**: `defaults/static/js/editor.js`
- **当前代码位置**:
  - `editor.js:6` — 模块变量声明区（`pollTimer = null` 之后）
  - `editor.js:70-72` — `textarea.oninput` 处理函数
  - `editor.js:121-139` — `cancelEdit()` + `exitEditMode()`（`exitEditMode` 的分屏清理修改统一归入 Task 1.4）
- **实现要点**:

  1. **新增模块变量** — 在 `editor.js:6` (`pollTimer = null` 之后) 新增：
     ```javascript
     var debounceTimer = null;
     var DEBOUNCE_DELAY = 150; // ms, 默认防抖延迟
     ```

  2. **新增 `renderPreview()` 函数** — 在 `editor.js` 中 `saveContent()` 之前（约 `:83` 之前）插入：
     ```javascript
     function renderPreview() {
         var preview = document.querySelector(".editor-split-preview");
         var textarea = document.querySelector(".editor-textarea");
         if (!preview || !textarea) return;

         // XSS 边界说明:
         // marked.parse() 默认透传 raw HTML（如 <script>、事件处理器）。
         // 由于内容来自用户自己在 textarea 中的输入（非外部不可信来源），
         // 此行为属于 self-XSS，可接受。若未来引入协作编辑，
         // 需在 marked.parse() 结果上包装 DOMPurify 等清理步骤。
         // 参考: NFR-SPLIT-08, AC-SPLIT-07, Review 问题 2/17

         // 图片相对路径说明 (Review 问题 6):
         // marked 将 ![img](./screenshot.png) 渲染为 <img src="./screenshot.png">。
         // 浏览器按页面 URL 解析此相对路径（如 /doc/readme.md），
         // 而非文件所在目录（如 /doc/screenshot.png），因此预览中相对路径图片可能不可见。
         // 最终渲染以服务端 goldmark 输出为准（FR-SPLIT-08）。
         preview.innerHTML = marked.parse(textarea.value);
     }
     ```

  3. **新增 `scheduleRender()` 防抖调度函数**:
     ```javascript
     function scheduleRender() {
         clearTimeout(debounceTimer);
         debounceTimer = setTimeout(function () {
             requestAnimationFrame(renderPreview);
         }, DEBOUNCE_DELAY);
     }
     ```
     - `clearTimeout` 清除前次等待（连续输入只触发最后一次渲染）
     - `requestAnimationFrame` 将渲染推迟到下一帧，避免阻塞 textarea 输入事件处理（FR-LIVE-06）
     - 默认 150ms 延迟在用户感知上足够"实时"（FR-LIVE-02）

  4. **修改 `textarea.oninput`** (`editor.js:70-72`) — 在 `isDirty` 赋值后追加 `scheduleRender()` 调用：
     ```javascript
     textarea.oninput = function () {
         isDirty = textarea.value !== originalContent;
         scheduleRender();
     };
     ```

  5. **在 `exitEditMode()` 中清理定时器** (`editor.js:130-139`) — Review 问题 8：
     ```javascript
     function exitEditMode() {
         isEditing = false;
         isDirty = false;
         document.body.removeAttribute("data-editing");
         toggleUI(false);
         if (pollTimer) {
             clearInterval(pollTimer);
             pollTimer = null;
         }
         // 新增：清除防抖定时器，防止退出后残留渲染
         if (debounceTimer) {
             clearTimeout(debounceTimer);
             debounceTimer = null;
         }
         // 新增：清空预览面板内容 (FR-INT-05)
         var preview = document.querySelector(".editor-split-preview");
         if (preview) preview.innerHTML = "";
         // 新增：重置 no-preview class
         var wrapper = document.querySelector(".editor-split-wrapper");
         if (wrapper) wrapper.classList.remove("no-preview");
     }
     ```

### Task 1.5 — Preview 按钮行为变更为切换预览面板

- **对应需求**: FR-INT-01, FR-INT-02, FR-INT-03, FR-SPLIT-04, FR-SPLIT-05
- **对应 AC**: AC-SPLIT-02
- **依赖**: Task 1.2（preview div 存在）, Task 1.4（renderPreview 函数定义）
- **修改文件**: `defaults/static/js/editor.js`
- **当前代码位置**: `editor.js:34-37` — Preview 按钮事件绑定
- **当前代码**:
  ```javascript
  var previewBtn = document.querySelector(".editor-btn-preview");
  if (previewBtn) {
      previewBtn.addEventListener("click", cancelEdit);
  }
  ```
- **实现要点**:

  1. **修改事件绑定** (`editor.js:35`) — 从 `cancelEdit` 改为 `togglePreview`:
     ```javascript
     previewBtn.addEventListener("click", togglePreview);
     ```

  2. **新增 `togglePreview()` 函数** — 在 `editor.js` 中 `exitEditMode()` 之前插入：
     ```javascript
     function togglePreview() {
         if (!isEditing) return;

         var wrapper = document.querySelector(".editor-split-wrapper");
         var previewBtn = document.querySelector(".editor-btn-preview");
         if (!wrapper) return;

         var isHidden = wrapper.classList.toggle("no-preview");

         if (!isHidden) {
             // 重新展开预览 → 立即渲染当前内容
             renderPreview();
         }

         // 更新按钮激活态样式 (FR-INT-02)
         if (previewBtn) {
             previewBtn.classList.toggle("active", !isHidden);
         }
     }
     ```
     - `classList.toggle("no-preview")` 切换 preview 面板的显示/隐藏（CSS 已定义对应规则）
     - 返回 `true` 表示 class 已添加（preview 隐藏），`false` 表示已移除（preview 显示）
     - 展开时立即调用 `renderPreview()` 刷新内容（覆盖 FR-LIVE-03 手动刷新场景）
     - 切换时 textarea 内容、光标位置、`isDirty` 标记均不改变（FR-INT-03）

### Task 1.6 — 集成 enterEditMode / exitEditMode / toggleUI

- **对应需求**: FR-INT-04, FR-INT-05, FR-INT-06, FR-INT-07, FR-INT-08, FR-INT-09, FR-INT-11, FR-INT-12, NFR-SPLIT-05
- **对应 AC**: AC-SPLIT-01（编辑闭环）, AC-SPLIT-05（零退化）
- **依赖**: Task 1.4（renderPreview + scheduleRender）, Task 1.5（togglePreview）
- **修改文件**: `defaults/static/js/editor.js`
- **当前代码位置**:
  - `editor.js:57-81` — `enterEditMode()` 函数体
  - `editor.js:130-139` — `exitEditMode()` 函数体
  - `editor.js:167-177` — `toggleUI(editMode)` 函数体
- **实现要点**:

  1. **`enterEditMode()`** (`editor.js:57-81`) — 加载内容后初始化预览：
     - 在 `textarea.setSelectionRange(0, 0)` 之后（`:74` 之后）新增：
       ```javascript
       // 初始化分屏预览：默认显示预览面板 (FR-INT-04)
       var previewBtn = document.querySelector(".editor-btn-preview");
       if (previewBtn) previewBtn.classList.add("active");

       // 设定默认防抖延迟（Phase 2 会动态调整）
       DEBOUNCE_DELAY = 150;

       // 初始渲染 + 重置 no-preview 状态
       var wrapper = document.querySelector(".editor-split-wrapper");
       if (wrapper) wrapper.classList.remove("no-preview");
       renderPreview();
       ```
     - 现有 `toggleUI(true)` 调用（`:67`）已处理 toolbar/container 显示

  2. **`exitEditMode()`** — 已在 Task 1.4 步骤 5 中完整描述（清除 debounceTimer、清空 preview innerHTML、移除 no-preview class）。Task 1.6 不再重复修改 `exitEditMode()`，实施时需确保将 Task 1.4 步骤 5 的 `exitEditMode()` 修改作为独立提交单元，避免与其他 Task 的修改产生合并冲突。

  3. **`toggleUI(editMode)`** (`editor.js:167-177`) — 无需修改核心逻辑：
     - 现有 `editorContainer.style.display` 切换（`:175`）仍然控制整个 `.editor-container` 的显示/隐藏
     - `.editor-split-wrapper` 和 `.editor-split-preview` 作为 `.editor-container` 的子元素，随父容器一起显示/隐藏
     - `[data-editing="true"]` CSS 规则通过后代选择器自动激活所有 split 布局样式

  4. **键盘快捷键不变** (`editor.js:179-201`)：
     - Ctrl+S 保存 → `saveContent()` → 触发 reload（与分屏无交互）
     - Esc 取消 → `cancelEdit()` → `exitEditMode()`（已在步骤 2 处理清理）
     - Ctrl+Enter 保存并停留 → `saveContent()` → reload（同 Ctrl+S）
     - FR-INT-09 全覆盖，零退化

  5. **侧边栏切换** (`editor.js:218-236`) 和 **beforeunload** (`editor.js:41-45`)：
     - `interceptSidebarLinks()` 和 `showNavDialog()` 完全不变 — 它们的逻辑仅依赖 `isEditing` 和 `isDirty`（FR-INT-11）
     - `beforeunload` 仅检查 `isDirty`，不感知分屏（FR-INT-12）

### Task 1.7 — 主题兼容性

- **对应需求**: FR-SPLIT-10, FR-SPLIT-11, NFR-SPLIT-06
- **对应 AC**: AC-SPLIT-04
- **依赖**: Task 1.3（CSS 基础布局规则已就位）
- **修改文件**: `defaults/static/css/editor.css`
- **当前代码位置**: 
  - `editor.css:74-113` — `@media (prefers-color-scheme: dark)` 块
  - `editor.css:115-191` — `[data-theme="dark"]` + `[data-theme="light"]` 块
- **实现要点**:

  1. **继承 `.markdown-body` 样式**: 预览面板在 `<body class="markdown-body">` 内，`.markdown-body h1`、`.markdown-body p`、`.markdown-body pre` 等后代选择器自动作用于 preview div 内的 marked 输出，无需额外配置。github-markdown-light/dark.css 通过 `<link media>` + `[data-theme]` 属性自动切换（Review 技术可行性 §4/§5）。

2. **prefers-color-scheme 暗色** — 在 `editor.css:112` 之后（`@media (prefers-color-scheme: dark)` 块的 `}` 之前）追加：
      ```css
      [data-editing="true"] .editor-split-preview {
        background: #0d1117;
        color: #e6edf3;
        border-color: #30363d;
      }

      [data-editing="true"] .editor-btn-preview.active {
        background-color: #0c2d6b;
        border-color: #58a6ff;
        color: #58a6ff;
      }
      ```
      使用 `[data-editing="true"]` 前缀确保预览面板和按钮样式仅在编辑模式下生效，与非编辑模式下的基础编辑器暗色样式不冲突。

  3. **`[data-theme="dark"]` 覆盖** — 在 `editor.css:152` 之后追加：
     ```css
     [data-theme="dark"] .editor-split-preview {
       background: #0d1117;
       color: #e6edf3;
       border-color: #30363d;
     }

     [data-theme="dark"] .editor-btn-preview.active {
       background-color: #0c2d6b;
       border-color: #58a6ff;
       color: #58a6ff;
     }
     ```

  4. **`[data-theme="light"]` 覆盖** — 在 `editor.css:191` 之后追加：
     ```css
     [data-theme="light"] .editor-split-preview {
       background: #ffffff;
       color: #1f2328;
       border-color: #d0d7de;
     }

     [data-theme="light"] .editor-btn-preview.active {
       background-color: #ddf4ff;
       border-color: #54aeff;
       color: #0969da;
     }
     ```

  5. **主题切换时的渲染**: `theme-switch.js` 通过修改 `<body data-theme="...">` 和 `<link media>` 切换 CSS。preview 面板的 CSS 使用 `[data-theme]` 选择器跟随，无需 JS 重新调用 `marked.parse()` — 渲染的 HTML 内容不变，仅样式切换（Review 问题 18）。

### Task 1.8 — XSS 边界文档 + Review 发现的其他注意事项

- **对应需求**: NFR-SPLIT-08, FR-SPLIT-07, FR-SPLIT-08
- **对应 AC**: AC-SPLIT-07
- **依赖**: Task 1.4（renderPreview 中已添加代码注释）
- **修改文件**: `defaults/static/js/editor.js`（仅注释）
- **当前代码位置**: `editor.js` 中 `renderPreview()` 函数体（Task 1.4 已添加注释）
- **实现要点**:

  1. **XSS 边界**（已在 Task 1.4 的 `renderPreview()` 中通过注释记录）：
     - marked 默认透传 raw HTML（`<script>`、`onerror` 等事件处理器不会被过滤）
     - 输入源是用户自己在 textarea 中的内容 → **self-XSS**，风险可接受
     - 验收标准 AC-SPLIT-07 要求的 `<script>alert(1)</script>` 不会被过滤的行为是**预期行为**，不是 bug
     - 若未来引入协作编辑或外部内容注入，必须引入 DOMPurify 或 `sanitize-html` 包裹 `marked.parse()` 结果

  2. **图片相对路径限制**（已在 Task 1.4 的 `renderPreview()` 中通过注释记录）：
     - `![img](./screenshot.png)` → 浏览器按页面 URL 解析，非文件目录
     - 最终渲染以服务端 goldmark 为准（FR-SPLIT-08）

  3. **代码块无语法高亮**（Review 问题 17）：
     - 服务端 goldmark + chroma 生成高亮 HTML，marked 仅生成 `<code class="language-xxx">` 标签
     - 预览面板中代码块无高亮是预期差异，FR-SPLIT-08 已声明允许此差异
     - 无需在 JS 中注入高亮逻辑

  4. **防抖定时器清理**（Review 问题 8）：已在 Task 1.4 步骤 5 中通过 `exitEditMode()` 内的 `clearTimeout(debounceTimer)` 处理。保存（`saveContent()`）触发 `window.location.reload()`（`:111`），reload 后 JS 全量重置，无需额外清理。

---

## 四、Phase 2: 增强（P1 项）

> **目标**: 大文档下自动调节防抖延迟，避免卡顿；可拖拽调整分屏比例。
>
> **验收**: AC-SPLIT-10, AC-SPLIT-11

### Task 2.1 — 大文档自适应防抖

- **对应需求**: FR-LIVE-04, NFR-SPLIT-02
- **对应 AC**: AC-SPLIT-11
- **依赖**: Phase 1 完成
- **修改文件**: `defaults/static/js/editor.js`
- **当前代码位置**: 
  - `editor.js:7` — `DEBOUNCE_DELAY` 常量（Task 1.4 新增）
  - `editor.js:70-72` — `textarea.oninput`（调用 `scheduleRender()`）
  - Task 1.4 的 `scheduleRender()` 函数
- **实现要点**:

  1. **修改 `scheduleRender()`** — 在设置 timer 前根据文本长度动态计算延迟：
     ```javascript
     function scheduleRender() {
         clearTimeout(debounceTimer);

         // 大文档自适应防抖 (FR-LIVE-04):
         //   < 5000 行  → 150ms (默认)
         //   5000+ 行   → 300ms
         var textarea = document.querySelector(".editor-textarea");
         var lines = textarea ? textarea.value.split("\n").length : 0;
         var delay = lines > 5000 ? 300 : DEBOUNCE_DELAY;

         debounceTimer = setTimeout(function () {
             requestAnimationFrame(renderPreview);
         }, delay);
     }
     ```
     阈值 5000 行对应 FR-LIVE-04 的精确要求。仅影响防抖延迟，不改变渲染逻辑。

### Task 2.2 — 可拖拽分隔条（Draggable Splitter）

- **对应需求**: FR-SPLIT-02
- **对应 AC**: AC-SPLIT-10
- **依赖**: Phase 1 完成
- **修改文件**: `defaults/templates/layout.html`, `defaults/static/css/editor.css`, `defaults/static/js/editor.js`
- **当前代码位置**:
  - `layout.html:100-105` — `.editor-split-wrapper` 内部（Task 1.2 新增）
  - `editor.css` — split layout 规则（Task 1.3 新增）
  - `editor.js` — 模块顶部变量区
- **实现要点**:

  1. **HTML**: 在 `.editor-split-wrapper` 内、textarea 和 preview 之间插入分隔条：
     ```html
     <div class="editor-split-wrapper">
       <textarea class="editor-textarea" aria-label="Markdown editor"></textarea>
       <div class="editor-split-handle" aria-label="Resize panels" role="separator" tabindex="0"></div>
       <div class="editor-split-preview" aria-label="Live Markdown preview"></div>
     </div>
     ```

  2. **CSS**: 分隔条样式：
     ```css
     .editor-split-handle {
       width: 6px;
       min-width: 6px;
       cursor: col-resize;
       background: #d0d7de;
       transition: background-color 0.15s ease;
       flex-shrink: 0;
     }

     .editor-split-handle:hover,
     .editor-split-handle.dragging {
       background: #0969da;
     }

     /* 拖拽时禁止 iframe/text selection */
     body.editor-resizing {
       user-select: none;
       cursor: col-resize;
     }
     ```

  3. **JS**: 拖拽事件处理逻辑：
     ```
     1. mousedown on .editor-split-handle:
        a. 记录起始鼠标 X 坐标 + 当前 wrapper 宽度 + 当前 textarea 百分比宽度
        b. 给 body 添加 "editor-resizing" class（禁用 user-select）
        c. 给 handle 添加 "dragging" class
        d. 注册 document mousemove 和 mouseup 监听器

     2. document mousemove:
        a. 计算 deltaX = 当前 mouseX - 起始 mouseX
        b. 新 textarea 百分比 = 起始百分比 + (deltaX / wrapperWidth) * 100
        c. 约束在 20% ~ 80% 之间
        d. 设置 textarea.style.flex = "0 0 " + 新百分比 + "%"
        e. 设置 preview.style.flex = "1"（填充剩余空间）

     3. document mouseup:
        a. 移除 body "editor-resizing" class
        b. 移除 handle "dragging" class
        c. 注销 mousemove/mouseup 监听器
        d. 将比例写入 sessionStorage("go-grip-split-ratio")
     ```

4. **响应式约束**:
      - `<768px` 堆叠模式下禁用拖拽（`pointer-events: none`）
      - `sessionStorage` 持久化比例（同标签页内保留，跨标签页独立）

   5. **边界场景处理**:
      - 拖拽中浏览器窗口失焦（`window blur`）：触发 `mouseup` 清理逻辑，移除 `editor-resizing` / `dragging` class，注销监听器
      - 拖拽中按 `Esc` 键：同上清理逻辑
      - `sessionStorage` 比例恢复：`enterEditMode()` 初始化时读取 `sessionStorage("go-grip-split-ratio")`，若有值则应用到 textarea/preview 的 flex 比例；预览面板从 `no-preview` 状态展开时也恢复先前比例

---

## 五、Phase 3: 完善（P2 项）

> **目标**: 空内容占位提示 + Ctrl+P 快捷键。
>
> **验收**: 视觉检查 + 功能测试
>
> **FR-LIVE-03 覆盖说明**：FR-LIVE-03（手动刷新预览）由 Task 1.5 的 `togglePreview()` 行为满足。当预览面板从隐藏状态重新展开时（用户通过 Preview 按钮或 Ctrl+P 快捷键切换回可见状态），`togglePreview()` 立即调用 `renderPreview()` 重新渲染当前 textarea 内容，实现"手动刷新"效果，无需额外 Task。
>
> **FR-LIVE-05 处置**：FR-LIVE-05（超过 100KB 时提示"文档较大，预览性能可能受影响"）为 P2 需求。当前 Phase 2 的自适应防抖（Task 2.1，5000+ 行 → 300ms 延迟）已降低大文档的输入感知延迟，暂不实现独立的大小提示 UI。如未来需要，可在 Phase 3 新增 Task 检测 `textarea.value.length` 并在预览面板顶部显示提示横幅。

### Task 3.1 — 空内容占位提示

- **对应需求**: FR-SPLIT-13
- **对应 AC**: 视觉检查
- **依赖**: Phase 1 完成
- **修改文件**: `defaults/static/js/editor.js`, `defaults/static/css/editor.css`
- **当前代码位置**: 
  - `editor.js` — `renderPreview()` 函数（Task 1.4 新增）
  - `editor.css` — `.editor-split-preview` 规则（Task 1.3 新增）
- **实现要点**:

  1. **JS**: 在 `renderPreview()` 开头判断内容是否为空，空时设置占位 HTML：
     ```javascript
     function renderPreview() {
         var preview = document.querySelector(".editor-split-preview");
         var textarea = document.querySelector(".editor-textarea");
         if (!preview || !textarea) return;

         var content = textarea.value.trim();
         if (!content) {
             preview.innerHTML = '<p class="editor-split-placeholder">Preview will appear here...</p>';
             return;
         }

         preview.innerHTML = marked.parse(textarea.value);
     }
     ```
     文案 "Preview will appear here..." 使用英文（go-grip 项目语言为英文，Review 问题 10）。

  2. **CSS**: 占位文本样式：
     ```css
     .editor-split-placeholder {
       margin: 0;
       color: #8b949e;
       font-size: 14px;
       font-style: italic;
       text-align: center;
       padding: 40px 0;
     }
     ```

### Task 3.2 — Ctrl+P 快捷键切换预览面板

- **对应需求**: FR-INT-10
- **对应 AC**: 功能测试
- **依赖**: Phase 1 完成（`togglePreview()` 已定义）
- **修改文件**: `defaults/static/js/editor.js`
- **当前代码位置**: `editor.js:179-201` — `handleKeydown()` 函数
- **实现要点**:

  1. 在 `handleKeydown()` 函数（`:179`）中，`Escape` 键处理之前（`:196` 之前）新增 Ctrl+P 处理：
     ```javascript
     if (isCtrl && e.key === "p") {
         e.preventDefault();
         togglePreview();
         return;
     }
     ```
     仅在 `isEditing === true` 时生效（函数开头已有此 guard）。仅在 textarea 获得焦点时触发（Ctrl+P 组合键在 textarea 内不会触发浏览器默认打印行为，因为 `e.preventDefault()` 已拦截）。

---

## 六、变更文件汇总

| 文件 | Phase 1 | Phase 2 | Phase 3 | 总新增/修改行 |
|------|---------|---------|---------|-------------|
| `defaults/static/js/marked.min.js` | 新增 (第三方库) | — | — | ~1 行 (二进制) |
| `defaults/templates/layout.html` | +4 行 (修改) | +2 行 (handle div) | — | +6 |
| `defaults/static/js/editor.js` | +60 行 (修改) | +40 行 (修改) | +15 行 (修改) | +115 |
| `defaults/static/css/editor.css` | +90 行 (新增) | +25 行 (新增) | +10 行 (新增) | +125 |

**无后端变更**。`internal/server.go` 零修改，`defaults.StaticFiles` embed.FS 自动包含 `marked.min.js`。

---

## 七、实现依赖关系

```
Task 1.1 (marked 库)
  └── Task 1.4 (renderPreview 渲染引擎)
Task 1.2 (HTML split-wrapper + preview div)
  ├── Task 1.3 (CSS 分屏布局)
  ├── Task 1.4 (渲染引擎)
  └── Task 1.5 (Preview 按钮行为变更)
Task 1.1 + Task 1.2 + Task 1.3 + Task 1.4 + Task 1.5
  └── Task 1.6 (enterEditMode/exitEditMode/toggleUI 集成)
Task 1.3
  └── Task 1.7 (主题兼容)
Task 1.4 (完成)
  └── Task 1.8 (XSS 边界文档)
Phase 1 (完成)
  ├── Task 2.1 (大文档自适应防抖)
  └── Task 2.2 (可拖拽分隔条)
Phase 2 (完成)
  ├── Task 3.1 (空内容占位)
  └── Task 3.2 (Ctrl+P 快捷键)
```

Task 1.1 + Task 1.2 可**并行开发**；Task 1.3 + Task 1.4 可**并行开发**（都依赖 Task 1.2 但彼此独立）。

---

## 八、关键设计决策与 Review 问题闭环

| Review 问题 | 严重程度 | 解决方式 | 对应任务 |
|-------------|---------|---------|---------|
| CSS flex-direction 冲突 (`:130`) | **严重** | `.editor-split-wrapper` 内部 row flex，`.editor-container` column 不变 | Task 1.2, 1.3 |
| FR-SPLIT-12/AC-SPLIT-09 P0/P1+ 矛盾 (`:13`) | **严重** | AC-SPLIT-09 提升至 P0 AC (已在前置 requirements 文档中修复) | Task 1.3 |
| NFR-SPLIT-08 XSS 缺失 AC (`:15`) | **严重** | 新增 AC-SPLIT-07 (已在前置 requirements 文档中修复)，code comment | Task 1.8 |
| FR-LIVE-06 缺失 AC (`:20`) | **严重** | 新增 AC-SPLIT-08 (已在前置 requirements 文档中修复) | Task 1.4 |
| toggleUI 变更范围不完整 (`:26`) | 主要 | FR-INT-08 覆盖 + Task 1.6 详述 DOM 操作差异 | Task 1.6 |
| 图片相对路径未覆盖 (`:30`) | 主要 | `renderPreview()` 注释记录限制 | Task 1.8 |
| marked 获取方式缺失 (`:32`) | 主要 | Task 1.1 明确下载 URL + 版本号 + 嵌入流程 | Task 1.1 |
| 防抖 pending 未处理 (`:34`) | 主要 | `exitEditMode()` 中 `clearTimeout(debounceTimer)` | Task 1.4（统一归属） |
| DOM 插入位置 (`:37`) | 主要 | `.editor-split-preview` 在 `.container` 内、`.preview-content` 前，利用现有全宽规则 | Task 1.2 |
| scroll-behavior 冲突 (`:48`) | 次要 | `.editor-split-preview { scroll-behavior: auto !important }` 覆盖 docs-layout.css 的 `smooth !important` | Task 1.3 |
| 代码块无高亮 (`:63`) | 次要 | FR-SPLIT-08 允许差异，Task 1.8 注释记录预期行为 | Task 1.8 |
| 主题切换渲染 (`:65`) | 次要 | CSS `[data-theme]` 选择器跟随，不重新渲染 | Task 1.7 |

---

## 九、与现有编辑功能交互确认

| 功能 | Phase 1 原有 | 分屏后 | 风险 |
|------|-------------|--------|------|
| 保存 (Ctrl+S / Save button) | POST → reload | **不变** | 无 |
| 取消 (Esc / Cancel button) | confirm → exitEditMode | **不变** — exitEditMode 中新增清理 | 低 |
| 侧边栏切换文件 | 三选项对话框 | **不变** | 无 |
| beforeunload 拦截 | 检查 isDirty | **不变** | 无 |
| 外部文件变更检测 (pollTimer) | 5s 轮询 | **不变** | 无 |
| 热重载 (保存 → reload) | WebSocket 触发 | **不变**（纯客户端功能） | 无 |
| 主题切换 | CSS `[data-theme]` | `[data-theme]` 选择器扩展覆盖 preview panel | 低 |
| 侧边栏状态保持 (sessionStorage) | 保存前写入 | **不变** | 无 |
| 滚动位置恢复 (sessionStorage) | 保存前写入 | **不变** | 无 |

---

## 十、验证清单

Phase 1 完成后逐项验证：

- [ ] 编辑模式下默认显示左侧 textarea + 右侧实时预览（AC-SPLIT-01）
- [ ] 在 textarea 输入 Markdown → 停止输入 ~150ms 后右侧自动更新渲染
- [ ] GFM 核心语法全部正确渲染：标题、粗体/斜体、删除线、链接、图片、列表、任务列表、表格、代码块、引用、分割线、自动链接（FR-SPLIT-07）
- [ ] 服务端特有扩展（CJK emphasis、mermaid、mathjax、emoji、高亮）在预览中不显示但也不破坏输出
- [ ] Preview 按钮高亮/取消高亮反映预览面板可见性（AC-SPLIT-02）
- [ ] 点击 Preview 按钮切换面板时编辑器内容不丢失（AC-SPLIT-02）
- [ ] >=1024px: 左右 50% 分屏（AC-SPLIT-03）
- [ ] <768px: 预览面板堆叠在编辑器下方（AC-SPLIT-03）
- [ ] 768-1023px: 分屏保持可用，不重叠、不溢出
- [ ] 浅色/深色/系统主题切换后预览面板配色跟随（AC-SPLIT-04）
- [ ] 编辑器 + 预览面板各自独立滚动条（AC-SPLIT-09）
- [ ] Save / Cancel / Esc / Ctrl+S / Ctrl+Enter 行为不变（AC-SPLIT-05）
- [ ] 侧边栏切换文件三选项对话框不变
- [ ] 热重载（保存后自动刷新所有标签页）不变
- [ ] marked.min.js 从本地加载，无 CDN 请求（AC-SPLIT-06）
- [ ] `<script>alert(1)</script>` 不被执行（self-XSS 边界，AC-SPLIT-07）
- [ ] 2000+ 行文档编辑时 textarea 输入无感知延迟（AC-SPLIT-08）
- [ ] 退出编辑模式后 preview div 内容清空，`.preview-content` 恢复显示
- [ ] Chrome / Firefox / Edge 最近两个大版本分屏功能正常（NFR-SPLIT-04）

---

*计划版本: v1.1，基于 editing-split-screen-requirements.md v1.0 / editing-split-screen-review.md v1.0 / editing-split-screen-plan-review.md v1.0 / editing-implementation-plan.md v1.0 / editor.js / editor.css / layout.html / docs-layout.css*

---

## 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|---------|
| v1.0 | 2026-05-17 | 初始版本 |
| v1.1 | 2026-05-17 | Review 修复：1) Task 1.3 step 5 主题 CSS 统一移交 Task 1.7，消除重复；2) `.editor-split-preview` 添加 `overflow-x: hidden` 防止水平溢出；3) Task 1.7 `prefers-color-scheme: dark` 规则添加 `[data-editing="true"]` 前缀及 `.editor-btn-preview.active` 覆盖；4) Task 1.6 `exitEditMode()` 修改统一归属 Task 1.4；5) 依赖图 Task 1.7 移除对 Task 1.6 的依赖；6) 第八章 Review 闭环表 `scroll-behavior` 描述补全 `!important`；7) Phase 3 新增 FR-LIVE-05 处置声明；8) 验证清单新增 NFR-SPLIT-04 跨浏览器验证项；9) Task 2.2 补充边界场景处理（窗口失焦/Esc/sessionStorage 恢复）；10) `.editor-btn-preview.active` 添加 `[data-editing="true"]` 前缀 |
