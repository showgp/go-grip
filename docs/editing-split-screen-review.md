# 分屏预览需求文档 Review

## 一、总体评价

The document is well-structured, covers the majority of the feature surface, and makes a sound case for marked as the client-side parser. The FR/NFR/AC taxonomy is clear and traceable. However, there are several P0 acceptance criteria gaps, one internal priority contradiction (FR-SPLIT-12 P0 but AC-SPLIT-09 under P1+), and missing coverage for XSS testing, image-path resolution, and the logistics of embedding marked into the build pipeline.

---

## 二、问题清单

### 严重 (Blocking)

1. **FR-SPLIT-12 与其 AC 优先级矛盾** (`:112` vs `:252`)：FR-SPLIT-12 "预览面板应有独立的垂直滚动条" 标记为 P0，但它的验收标准 AC-SPLIT-09 却列在 "增强接受标准（对应 P1+）" 下 (`:251`)。P0 需求没有对应的 P0 AC 是不允许的。**修复**：将 AC-SPLIT-09 移至 9.1 MVP 接受标准，或者反过来将 FR-SPLIT-12 降级为 P1。

2. **NFR-SPLIT-08（XSS 防护）缺失 AC** (`:213`)：这是 P0 安全需求，但完全没有对应的验收标准。9.1 节中 6 条 AC 无一涉及 XSS 测试。**修复**：新增 AC-SPLIT-0X，覆盖至少以下场景：
   - 在 textarea 中输入 `<script>alert(1)</script>`，预览面板不应执行该脚本
   - 在 textarea 中输入 `<img src=x onerror=alert(1)>`，事件处理器不应执行
   - （注意：marked 默认允许 raw HTML 透传，作者需评估是否这是可接受的 self-XSS 边界）

3. **FR-LIVE-06（异步非阻塞渲染）缺失 AC** (`:137`)：这是 P0 需求（"渲染操作不得阻塞 textarea 的输入响应"），没有可验证的 AC。虽然性能测试较难自动化，但至少需要一个手动 checklist 条目来验证大文档下输入不卡顿。

---

### 主要 (Should Fix)

4. **`toggleUI()` 的变更范围不完整** (FR-INT-07, `:161`)：当前 `editor.js:167-177` 的 `toggleUI()` 只 toggle 四个元素：previewToolbar、editorToolbar、editorContainer、previewContent。分屏模式需要在 editorContainer 内新增 preview-panel 子元素，且需要独立的显示/隐藏逻辑（因为 Preview 按钮可在编辑模式内单独切换它）。文档未细分 `toggleUI(true)` 进入编辑模式 vs Preview 按钮切换面板的 DOM 操作差异。**修复**：FR-INT-08 已部分覆盖（独立逻辑），但缺少对 toggleUI 必须同时处理 editorContainer 内部布局模式从"全宽 textarea"切换到"flex-row 分屏"的描述。

5. **`.editor-container` CSS 架构冲突** (`editor.css:47-50, 309-313`)：当前 `.editor-container` 在 `[data-editing="true"]` 下是 `flex-direction: column`，textarea 是 `flex: 1`。分屏需要 `.editor-container` 变为 `flex-direction: row`（或新增 wrapper），这与现有 flex-column 假设冲突。文档未提及需要修改现有 CSS 规则。**修复**：建议新增一个 `.editor-split-container` wrapper，保持 `.editor-container` 不变，分屏面板作为兄弟元素而非子元素插入。

6. **图片相对路径问题未覆盖** (FR-SPLIT-07, `:106`)：FR-SPLIT-07 声称支持图片渲染。但如果用户写 `![img](./screenshot.png)`，marked 生成的 `<img src="./screenshot.png">` 将相对于页面 URL 解析（如 `/doc/readme.md`），而非相对于文件所在目录（如 `/doc/screenshot.png`）。文档未提及此限制。**修复**：在 FR-SPLIT-08 或边界声明中补充说明客户端预览中相对路径图片可能无法正常显示。

7. **marked 库的获取方式缺失** (FR-DIST-01~03) ：FR-DIST 定义了文件位置和引用方式，但未说明 marked 库物理文件如何进入 `defaults/static/js/` 目录——是从 CDN 手动下载、npm install 后复制、还是通过构建脚本自动获取？这影响可复现构建。**修复**：补充一条需求：构建前应有脚本从固定版本 URL 下载 marked 并验证完整性（如校验 SHA256）。

8. **编辑器模式切换时 pending debounce 未处理**：FR-LIVE-01 定义了防抖策略，但未规定用户快速操作（如打字后立即点击 Save 或切换文件）时 pending 的渲染应如何处理。保存前应完成最后一次渲染还是直接取消？**修复**：补充说明 pending 渲染在保存/退出时应被取消或立即完成。

9. **FR-SPLIT-14 与预览面板 DOM 位置潜在冲突** (`:118`)：当前 `[data-editing="true"]` 隐藏 sidebar/TOC/footer（`editor.css:321-330`），且设置 `.docs-shell { display: block }`（`editor.css:328`）。如果预览面板是一个 fixed/sticky 元素或在 `.docs-shell` 外部，这些规则不会冲突。但如果它插入到 `.container` 内部，现有的 `.docs-main .container { max-width: 896px }`（`docs-layout.css:57`）和 `[data-editing="true"] .docs-main { max-width: 100% }`（`editor.css:332`）已正确扩展宽度。**确认点**：确保预览面板的 DOM 插入位置在 `.container` 内部、`.preview-content` 之前，以便利用现有的宽度扩展规则。

---

### 次要 (Nice to Fix)

10. **FR-SPLIT-13（空内容占位提示）P2 定位合理但有歧义** (`:114`)："预览将显示在此处" 是中文占位文案，需确认是否接受英文项目中出现中文硬编码，或应跟随模板语言。

11. **NFR-SPLIT-02 的性能指标过于严格** (`:202`)：1000 行文档在 16ms 内完成对于字节码解析+DOM 构建较为激进，尤其是在低端设备上。建议改为"不丢帧（<16ms per frame）且总耗时 <50ms"或标记为测量指标而非硬约束。

12. **FR-INT-09 引用了 `Ctrl+Enter`** (`:167`)：该快捷键在 `editor.js:190-193` 已实现（保存后触发 reload），但 `editing-requirements.md` 和 `editing-spec.md` 中均未将其列为需求——它仅出现在 `editing-implementation-plan.md` Task 2.1 中。如果 `Ctrl+Enter` 已被实现为 P1 特性，FR-INT-09 应注明引用来源。

13. **FR-SPLIT-10 预览 CSS 复用 .markdown-body** (`:111`)：文档说使用 `github-markdown-light.css` / `github-markdown-dark.css`。这两个文件的选择器以 `.markdown-body` 为前缀（如 `.markdown-body h1`），而 `<body>` 已有 `class="markdown-body"`。如果预览面板 DIV 同样加 `class="markdown-body"`，样式通过后代选择器自动继承。但需注意 `docs-layout.css:1` 的 `.markdown-body { scroll-behavior: smooth !important }` 也会作用于预览面板，可能影响独立滚动行为。建议显式覆盖预览面板的 `scroll-behavior`。

14. **FR-LIVE-05 100KB 阈值是否需要服务端配合** (`:135`)：FR-LIVE-05 说"超过阈值时提示"，但判断依据是 textarea.value.length 还是服务端元数据？如果依赖客户端字符串长度，UTF-8 多字节字符会使其不准确。建议明确测量方式。

---

### 建议补充 (Missing Coverage)

15. **编辑中网络断开的恢复流程**：当前文档未覆盖以下场景：用户已进入编辑模式（内容已加载），之后网络断开，用户仍在编辑，最后网络恢复。此时：
    - 实时预览继续工作（纯客户端操作）
    - Save 会因网络问题失败，这已被 EC-04 覆盖
    - 但 `pollTimer`（`editor.js:77`，5 秒轮询外部变更）在断网下会持续失败，是否应暂停轮询或给予用户提示？建议补充。

16. **多标签页同时分屏编辑**：两个标签页同时进入同一文件的分屏编辑模式。FR-INT-12 说 `beforeunload` 不变，NFR-SPLIT-05 说热重载不干扰。但如果两个标签页都做了修改，后保存的会覆盖先保存的（后写者胜出，见 EC-07）。此为预期行为但应在分屏文档中引用 EC-07。

17. **代码块语法高亮在预览中的呈现**：当前服务端用 chroma 做代码高亮，客户端 marked 仅生成 `<code class="language-xxx">` 标签但无高亮样式注入。用户会看到无高亮的代码块，与最终渲染差异显著。虽然 FR-SPLIT-08 允许差异，但建议在 AC 中明确此预期差异，避免将"代码块无高亮"误报为 bug。

18. **主题切换时预览面板内容是否需要重新渲染**：FR-SPLIT-11 说配色跟随主题。如果仅通过 CSS 变量或 `[data-theme]` 选择器切换（不重新调用 marked.parse），性能最优但需确保预览面板的 CSS 规则覆盖了所有 marked 输出的元素（code、pre、table、blockquote 等）。如果通过重新渲染实现，应提及短暂闪烁的处理方案。

19. **防抖实现与内存**：FR-LIVE-01 要求防抖。如果用户在编辑器中持续输入 10 分钟（如撰写长文），每次输入都设置/清除 timer，不会泄露。但文档未说明退出编辑模式时应清除所有 pending timer——这一点 FR-INT-05 有"释放 DOM 和内存资源"但未明确提及定时器。建议在 FR-INT-05 中补充"清除所有定时器（防抖、轮询）"。

---

## 三、优先级审查

| 编号 | 当前优先级 | 建议优先级 | 理由 |
|------|-----------|-----------|------|
| FR-SPLIT-02（可拖拽分隔条） | P1 | **维持 P1** | 正确。拖拽需要 mousedown/move/up 全套事件处理 + 百分比计算 + 持久化，复杂度高，MVP 阶段 1:1 固定比例足够。 |
| FR-SPLIT-04（预览面板收起/展开） | P0 | **维持 P0** | 正确。Preview 按钮（FR-INT-01 P0）依赖此能力，是编辑模式的核心 UI 操作。 |
| FR-SPLIT-12（独立滚动条） | P0 | **维持 P0** | 正确。无独立滚动条分屏基本不可用——编辑器和预览需独立浏览。但其 AC 需从 P1+ 移至 P0（见问题 1）。 |
| FR-SPLIT-13（空内容占位） | P2 | **维持 P2** | 正确。纯视觉增强，不影响功能。 |
| FR-LIVE-04（大文档自适应防抖） | P1 | **维持 P1** | 正确。P0 已有固定防抖，大文档优化可推迟。 |
| FR-LIVE-05（大文档性能提示） | P2 | **维持 P2** | 正确。纯提示信息，不阻塞用户操作。 |
| FR-LIVE-07/08（滚动同步） | P2 | **维持 P2** | 正确。滚动同步实现复杂度高（需遍历 rendered DOM 计算偏移），且在 Markdown 场景下精确度有限。 |
| FR-INT-10（Ctrl+P 快捷键） | P2 | **维持 P2** | 正确。Preview 按钮已覆盖此功能，快捷键为锦上添花。 |
| NFR-SPLIT-02（16ms/100ms 性能指标） | P1 | **维持 P1** | 建议放宽但优先级正确。 |

**结论**：除 FR-SPLIT-12 / AC-SPLIT-09 的 P0/P1+ 矛盾外，优先级标记整体合理。

---

## 四、AC 覆盖审查

### P0 AC 覆盖缺口

| 缺失覆盖的 P0 需求 | 严重程度 | 说明 |
|-------------------|---------|------|
| NFR-SPLIT-08（XSS 防护） | **严重** | 安全需求完全无 AC（见问题 2） |
| FR-SPLIT-12（独立滚动条） | **严重** | AC-SPLIT-09 错放在 P1+ 下（见问题 1） |
| FR-LIVE-06（异步非阻塞渲染） | **严重** | 性能需求无 AC（见问题 3） |
| NFR-SPLIT-01（脚本不阻塞首屏渲染） | 主要 | defer 属性可保证，但无 AC 验证 |
| NFR-SPLIT-03（渲染不阻塞输入） | 主要 | 与 FR-LIVE-06 重叠，但无独立 AC 测量 |
| NFR-SPLIT-07（CSS 命名前缀隔离） | 次要 | 已有 FR-SPLIT-16 覆盖功能层面，CSS 层面可 code review 替代 |
| FR-SPLIT-14/15/16（与现有 UI 共存） | 次要 | AC-SPLIT-05 "零退化" 可间接覆盖，但未列举具体检查项 |

### 可测试性问题

- **AC-SPLIT-05** (`:244`)："现有编辑功能...全部正常"——过于模糊。"全部" 是指 FR-10~FR-16 所有项？应拆分为具体可验证检查点。
- **AC-SPLIT-08** (`:251`)："编辑器不卡顿"——主观标准。建议改为 "textarea 输入延迟不超过 50ms（可通过 performance.now() 测量）" 或提供可观测指标。
- **AC-SPLIT-03** (`:243`)："平板端 (768-1023px) 分屏可用不断裂"——"不断裂" 主观。建议明确为 "所有 UI 元素可见、无水平溢出、无重叠"。

---

## 五、与现有功能交互审查

### 与现有编辑器 (`editor.js`) 的交互

| 交互点 | 当前行为 | 分屏后影响 | 风险评估 |
|--------|---------|-----------|---------|
| `toggleUI(true)` (`:167`) | 显示 editorToolbar + editorContainer，隐藏 previewToolbar + previewContent | 需额外显示 preview-panel 并初始化渲染器 | **低风险**——新增分支即可 |
| `toggleUI(false)` (`:167`) | 反向操作 | 需额外隐藏 preview-panel 并取消 pending 渲染/清除内容 | **低风险** |
| `enterEditMode()` (`:52`) | 加载内容到 textarea，设置 data-editing | 需初始化 marked 渲染 + 在 textarea.oninput 中触发防抖渲染 | **中风险**——需修改 `enterEditMode` 和 textarea.oninput 逻辑 |
| `exitEditMode()` (`:130`) | 移除 data-editing，清除 pollTimer | 需额外清除防抖 timer，清空 preview-panel innerHTML | **低风险** |
| `cancelEdit()` (`:121`) | 检查 isDirty → confirm → exitEditMode | 不变 | **无影响** |
| `handleKeydown()` (`:179`) | Ctrl+S / Ctrl+Enter / Esc | 分屏不引入新快捷键（P2 的 Ctrl+P 在 P2） | **无影响** |
| `checkExternalChanges()` (`:141`) | 5s 轮询 | 不变 | **无影响** |
| `interceptSidebarLinks()` (`:218`) | 阻止侧边栏导航 + 三选项对话框 | 不变 | **无影响** |

### 与现有 CSS (`editor.css` + `docs-layout.css`) 的交互

| 规则 | 位置 | 分屏影响 |
|------|------|---------|
| `[data-editing="true"] .editor-container { flex: 1; display: flex; flex-direction: column }` | `editor.css:309-313` | **冲突**：分屏需要 row 方向，需改为 `flex-direction: row` 或新增 wrapper |
| `[data-editing="true"] .editor-textarea { flex: 1; min-height: 0 }` | `editor.css:315-318` | **需调整**：textarea 在 row 布局中不再是 `flex: 1`（需设固定或百分比宽度） |
| `[data-editing="true"] .docs-main .container { display: flex; flex-direction: column; min-height: calc(100vh - 120px) }` | `editor.css:303-307` | **兼容**：column 方向 + flex children 可容纳 row 方向的分屏容器 |
| `[data-editing="true"] .docs-sidebar, .docs-toc, .footer, .docs-page-nav { display: none }` | `editor.css:321-326` | **兼容**：分屏模式下仍需隐藏，不变 |
| `[data-editing="true"] .docs-main { max-width: 100% }` | `editor.css:332 | **兼容**：分屏需要全宽 |
| `[data-editing="true"] .docs-shell { display: block }` | `editor.css:328` | **兼容**：移除 grid 约束 |
| `.docs-main .container { max-width: 896px }` | `docs-layout.css:57` | **兼容**：`[data-editing="true"] .docs-main` 已 override 为 100% |

### 与模板 (`layout.html`) 的交互

现有 DOM 结构（`:84-108`）：
```
div.container
  div.preview-toolbar
  div.editor-toolbar
  div.editor-container
    textarea.editor-textarea
  div.preview-content
  nav.docs-page-nav
```

**建议**：在 `.editor-container` 内新增 `.editor-split-preview` div（而非修改 `.editor-container` 的 flex 方向），以避免破坏现有的 textarea flex:1 column 布局的默认行为（在预览面板隐藏时仍是全宽 textarea）。

### 与 TOC (`toc-active.js`) 的交互

编辑模式下 sidebar/TOC 已通过 CSS 隐藏，TOC 高亮 JS 无需感知分屏。**无影响**。

### 与热重载的交互

NFR-SPLIT-05 要求不干扰热重载。当前保存触发 `window.location.reload()`（`editor.js:111`），reload 后整页刷新，分屏状态自然重置。**无影响**。

### 与主题切换 (`theme-switch.js`) 的交互

FR-SPLIT-11 要求预览配色跟随主题。`theme-switch.js` 通过修改 `<body data-theme="...">` 和 `<link media>` 切换实现。如果预览面板 CSS 使用 `[data-theme="dark"] .editor-split-preview` 选择器，可自动跟随。**注意**：如果主题切换中途预览面板正在渲染，可能存在短暂的样式不一致——建议先完成渲染再切换，或渲染函数中读取当前 theme 属性。

---

## 六、技术可行性补充说明

### 1. marked 的选择确认

marked (~12KB gzip) 是正确的选择。snarkdown 缺乏 GFM 是硬伤，markdown-it (~45KB) 对于 ~12KB 的代价来说体积过于庞大，micromark 的组装复杂度对于"辅助预览"场景过度工程化。

**但需注意**：
- marked v18+ 的 `marked.parse()` 默认**不** sanitize HTML。用户输入的 raw HTML（如 `<script>`）会被透传。考虑到这是本地开发工具且输入源是用户自己，self-XSS 风险可接受，但应在文档中明确声明此边界。
- 如果未来引入协作编辑，需在 marked 外层包裹 DOMPurify 或使用 `marked.parse(content, { sanitize: true })`（注意：marked 内置的 sanitizer 在 v5+ 已移除，需引入 `dompurify` 或 `sanitize-html`）。

### 2. 拖拽分隔条（FR-SPLIT-02, P1）的实现复杂度

如果需要实现，至少涉及：
- `mousedown` 在分隔条上注册，`mousemove`/`mouseup` 在 `document` 上监听（防止快速拖拽时鼠标离开分隔条）
- 百分比计算需考虑 `.editor-container` 的容器宽度和 min/max 约束
- 分隔条样式（hover 光标变化、拖拽态高亮）
- sessionStorage 持久化比例（可选）
- `<768px` 堆叠模式下应禁用拖拽

建议在 P2 实现计划中显式记录这些子任务。

### 3. 防抖 + 大文档的自适应策略

FR-LIVE-01（150-300ms debounce, P0）+ FR-LIVE-04（5000+ 行延长 debounce, P1）的实现建议：
- 使用 `requestIdleCallback()` 或 `requestAnimationFrame()` 包装 marked.parse 调用，使渲染在浏览器空闲时进行
- 超大文档时可分段渲染（渲染可见区域），但这引入了虚拟滚动的复杂度。如果文档 >5000 行且无分段渲染，单次 marked.parse 可能耗时 >500ms。文档应评估是否需要分段渲染或接受此限制。
- 防抖延迟可基于 `textarea.value.length` 动态计算：<10KB → 150ms, 10-50KB → 300ms, >50KB → 500ms

### 4. 预览面板 CSS 在 `.markdown-body` 下的具体方案

`github-markdown-light.css` 和 `github-markdown-dark.css` 的选择器格式为：
```css
.markdown-body h1 { ... }
.markdown-body p { ... }
.markdown-body pre { ... }
```

这些样式通过后代选择器生效，因此 `<body class="markdown-body">` 内的所有 h1/p/pre 元素都会自动应用。但预览面板内容是通过 `innerHTML` 设置的一段独立 HTML 片段，无需额外 class 即可继承 `.markdown-body` 样式。

**特殊情况**：预览面板如果被放在一个 `overflow: auto` 的容器中，该容器内的元素仍会正确继承 `.markdown-body` 样式——因为 CSS 后代选择器不关心 `overflow` 属性。

### 5. 服务端无需变更

审查 `server.go:392-405` 的 `htmlStruct` 结构体和 `server.go:197-204` 的 `renderMarkdown` 调用链后确认：分屏预览是纯客户端功能，服务端零变更。`RawContent` 字段已存在（`server.go:246`），模板已引用 `editor.js`（`layout.html:62`）。

### 6. 构建流程中嵌入 marked 的建议

当前项目使用 Go 1.21+ 的 `embed.FS`（`defaults.StaticFiles`），静态文件编译进二进制。marked 库的获取方式：
- **方案 A**：在 `defaults/static/js/` 下手动放置 `marked.min.js`，通过 git 提交（版本锁定）
- **方案 B**：在 Makefile 中添加 `make vendor` 目标，从 jsdelivr CDN 下载指定版本并校验 SHA256
- **方案 C**：使用 Go 的 `//go:embed` 的 `all:` 前缀，自动包含目录下所有文件

推荐方案 A + B 组合：提交 marked 到仓库以确保离线可用（FR-DIST-01），同时在 Makefile 中提供下载/更新脚本以便升级版本。

应补充一条构建需求：marked 文件必须来自官方发布的 release artifact，并记录版本号和来源 URL。

---

*Review 日期：2026-05-17*
*基于：editing-split-screen-requirements.md v1.0 / editing-requirements.md v2.0 / editing-spec.md v1.0 / editing-implementation-plan.md v1.0 / editor.js / editor.css / layout.html / server.go / docs-layout.css*
