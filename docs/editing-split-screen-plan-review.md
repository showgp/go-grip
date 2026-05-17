# 分屏预览实施计划 Review

## 一、总体评价

计划整体结构清晰，Phase 1-3 划分合理，与需求文档及 Review 文档的闭环完整。绝大部分 P0 需求都有对应的任务覆盖，`.editor-split-wrapper` 方案巧妙地避免了 CSS flex-direction 冲突。存在一个阻塞级技术错误（`scroll-behavior` 无法正确覆盖）、一个 Phase 3 需求无对应任务（FR-LIVE-03），以及依赖关系图不完整等问题。

---

## 二、需求覆盖审查

### 2.1 P0 需求 → Task 映射表

| P0 需求 | 对应任务 | 覆盖？ | 备注 |
|---------|---------|--------|------|
| FR-SPLIT-01 分屏布局 | Task 1.2, 1.3, 1.6 | ✅ | |
| FR-SPLIT-03 响应式 | Task 1.3 | ✅ | |
| FR-SPLIT-04 面板收起/展开 | Task 1.5, 1.3 | ✅ | |
| FR-SPLIT-05 默认展开 | Task 1.6 | ✅ | |
| FR-SPLIT-06 实时渲染 | Task 1.1, 1.4 | ✅ | |
| FR-SPLIT-07 GFM 语法 | Task 1.1 | ✅ | marked 内置 |
| FR-SPLIT-08 允许差异 | Task 1.8 | ✅ | 注释文档 |
| FR-SPLIT-09 非交互预览 | Task 1.3 | ✅ | pointer-events |
| FR-SPLIT-10 GitHub CSS | Task 1.7 | ✅ | 继承 .markdown-body |
| FR-SPLIT-11 主题跟随 | Task 1.7 | ✅ | |
| FR-SPLIT-12 独立滚动条 | Task 1.3 | ⚠️ | 见问题 1 |
| FR-SPLIT-14 隐藏行为不退化 | Task 1.3, 1.6 | ✅ | |
| FR-SPLIT-15 .preview-content 隐藏 | Task 1.6 | ✅ | toggleUI 不变 |
| FR-SPLIT-16 命名隔离 | Task 1.2 | ✅ | .editor-split-* |
| FR-LIVE-01 防抖 | Task 1.4 | ✅ | |
| FR-LIVE-02 延迟可感知 | Task 1.4 | ✅ | 150ms |
| FR-LIVE-06 异步非阻塞 | Task 1.4 | ✅ | rAF + debounce |
| FR-INT-01 Preview 按钮行为变更 | Task 1.5 | ✅ | |
| FR-INT-02 按钮视觉标识 | Task 1.5, 1.3 | ✅ | .active |
| FR-INT-03 切换不丢内容 | Task 1.5 | ✅ | |
| FR-INT-04 进入默认显示 | Task 1.6 | ✅ | |
| FR-INT-05 退出清空 | Task 1.4 step5, 1.6 | ✅ | |
| FR-INT-06 退出恢复 | Task 1.6 | ✅ | toggleUI |
| FR-INT-07 toggleUI 适配 | Task 1.6 | ✅ | CSS 驱动 |
| FR-INT-08 toggleUI 独立 | Task 1.6 | ✅ | |
| FR-INT-09 快捷键不变 | Task 1.6 | ✅ | |
| FR-INT-11 侧边栏切换 | Task 1.6 | ✅ | |
| FR-INT-12 beforeunload | Task 1.6 | ✅ | |
| FR-DIST-01 离线内置 | Task 1.1 | ✅ | |
| FR-DIST-02 目录位置 | Task 1.1 | ✅ | |
| FR-DIST-03 脚本引用 | Task 1.1 | ✅ | defer |
| NFR-SPLIT-01 不阻塞首屏 | Task 1.1 | ✅ | defer |
| NFR-SPLIT-03 不阻塞输入 | Task 1.4 | ✅ | rAF |
| NFR-SPLIT-04 浏览器兼容 | — | ❌ | 无任务覆盖 |
| NFR-SPLIT-05 热重载 | Task 1.6 | ✅ | |
| NFR-SPLIT-06 主题切换 | Task 1.7 | ✅ | |
| NFR-SPLIT-07 命名隔离 | Task 1.2 | ✅ | |
| NFR-SPLIT-08 XSS 防护 | Task 1.8 | ✅ | self-XSS 边界 |
| NFR-SPLIT-09 代码集中 | — | ✅ | 全在 editor.js |
| NFR-SPLIT-10 三方库隔离 | Task 1.1 | ✅ | 不合入修改 |

### 2.2 AC 覆盖

所有 P0 AC (AC-SPLIT-01~09) 均在 Phase 1 验证清单 (计划 §十) 中有对应检查项。✅

### 2.3 缺失的需求覆盖

| 需求 | 优先级 | 严重度 | 说明 |
|------|--------|--------|------|
| FR-LIVE-03 手动刷新触发 | P2 | 主要 | Phase 3 概述声称覆盖，但无对应 Task 3.x。Task 1.5 中的 togglePreview 展开时 `renderPreview()` 仅间接实现，缺失明确的"手动刷新预览"按钮或入口 |
| FR-LIVE-05 大文档性能提示 | P2 | 次要 | 整个计划中从未出现。可在 Phase 3 新增 Task 3.3 或明确声明降级/不做 |
| NFR-SPLIT-04 跨浏览器兼容 | P0 | 主要 | 无任何 Task 或验证项覆盖 Chrome/Firefox/Edge 兼容性测试 |
| NFR-SPLIT-02 性能指标测量 | P1 | 次要 | Task 2.1 自适应防抖间接改善性能，但无明确的 16ms/100ms 验证步骤 |

### 2.4 隐含任务

| 隐含任务 | 是否覆盖 | 说明 |
|----------|---------|------|
| 构建流程变更 | ✅ 无需 | `embed.FS` 自动包含 marked.min.js，server.go 零修改 |
| 自动化测试（JS） | ❌ | 无 renderPreview / togglePreview 单元测试，仅有手动验证清单 |
| 部署/打包 | ✅ 无需 | 纯客户端功能，二进制自然包含静态文件 |

---

## 三、技术正确性审查

### Task 1.1 — marked 库嵌入 ✅

- **file:line 准确性**: `layout.html:62`（`<script src="/static/js/editor.js" defer>`）— 确认正确。在行 62 之前插入 `<script src="/static/js/marked.min.js" defer></script>`，defer 执行顺序得以保持。
- **embed.FS**: `server.go:137` 对 `/static/` 路径使用 `http.FileServer(http.FS(defaults.StaticFiles))`。新增文件在 `defaults/static/js/` 下会**自动包含**，无需修改 server.go。✅
- **marked 版本**: 12.0.2 发布于 2023-12，非最新版。GFM 核心功能在该版本完整支持，但建议在计划中注明锁定此版本的理由（如：已知稳定、兼容性验证等）。⚠️ 次要

### Task 1.2 — HTML 结构 ✅

- **DOM 插入位置**: `.editor-split-wrapper` 放入 `.editor-container` 内部作为唯一子元素。`.editor-container` 保持 `flex-direction: column`（由 `editor.css:311` 规则控制），wrapper 内部 `flex-direction: row`。✅
- **与现有结构的兼容**: textarea 从 `.editor-container` 的直接子元素变为 `.editor-split-wrapper` 的子元素。现有 CSS `[data-editing="true"] .editor-textarea`（`editor.css:315-318`）通过后代选择器仍匹配 textarea。但新的 `.editor-split-wrapper .editor-textarea` 规则（特定性更高）会在分屏模式下正确覆盖。✅
- **命名隔离**: `.editor-split-*` 前缀不与现有 `.preview-content` / `.editor-*` 冲突。✅

### Task 1.3 — CSS 布局

**🔴 阻塞级问题: `scroll-behavior` 无法正确覆盖**

计划提议的规则：
```css
[data-editing="true"] .editor-split-preview {
  scroll-behavior: auto;   /* "覆盖 docs-layout.css:2 的 smooth !important" */
}
```

但 CSS 级联规则是:
1. `github-markdown-light.css:41` — `.markdown-body { scroll-behavior: auto !important }`
2. `docs-layout.css:2` — `.markdown-body { scroll-behavior: smooth !important }`（**后加载，同 `!important` 故胜出**）
3. 计划的 `.editor-split-preview { scroll-behavior: auto }` — 无 `!important`，**必输**

**修复**: 必须使用 `scroll-behavior: auto !important;`（或修改 `docs-layout.css` 排除预览面板，但计划声称不修改该文件）。

**其他 CSS 正确性问题**:

| 检查项 | 状态 | 说明 |
|--------|------|------|
| `editor.css:309-313` `.editor-container` column 保持不变 | ✅ | |
| `editor.css:315-318` `.editor-textarea` 被 split 规则正确覆盖 | ✅ | 特定性 `[data-editing] .editor-split-wrapper .editor-textarea` (0,2,1) > `[data-editing] .editor-textarea` (0,1,1) |
| `editor.css:64` base `resize: vertical` 被 split 模式 `resize: none` 覆盖 | ✅ | |
| `no-preview` 态恢复 `resize: vertical` + 全宽 border-radius | ✅ | |
| 响应式 `<768px` 列堆叠 | ✅ | |
| 暗色主题 `@media (prefers-color-scheme: dark)` 规则缺少 `[data-editing]` 前缀 | ⚠️ | 见下一条 |
| 768-1023px 范围无显式 media query | ✅ | 行内 flex:row 在此范围自然保持 |

**`@media (prefers-color-scheme: dark)` 规则缺少 `[data-editing="true"]` 前缀**: 计划 Task 1.3 step5 中 `prefers-color-scheme` 暗色规则直接使用 `.editor-split-preview`，而 `[data-theme="dark"]` 规则则使用完整前缀 `[data-theme="dark"] .editor-split-preview`。虽然 `.editor-split-preview` 在非编辑模式下不可见（父元素 `display: none`），不会产生视觉问题，但**风格不一致**，且如果未来 `.editor-split-preview` 在非编辑模式下可见，会产生意外的暗色背景。建议统一加 `[data-editing="true"]` 前缀。

### Task 1.4 — 渲染引擎

- **防抖逻辑正确**: `clearTimeout(debounceTimer)` → `setTimeout(..., DEBOUNCE_DELAY)` → 内部 `requestAnimationFrame(renderPreview)`。✅
- **`requestAnimationFrame` 确保非阻塞**: `marked.parse()` 在 rAF 回调中执行，此时浏览器已完成当前帧的输入处理。✅
- **XSS 边界**: marked 默认透传 raw HTML（`<script>` 不滤除）。计划正确识别为 self-XSS（输入源是用户自己），风险可接受。注释中已记录。✅
- **`exitEditMode` 中的 cleanup**: 清除 `debounceTimer`、清空 `preview.innerHTML`、移除 `no-preview`。**注意**: 此修改在 Task 1.4 中描述，但 `exitEditMode` 的完整变更也出现在 Task 1.6 step2 中。这意味着**同一函数的同一修改在两个 Task 中重复描述**——实施时需注意只改一次，避免重复合并。⚠️ 次要

### Task 1.5 — Preview 按钮行为变更

- **`classList.toggle("no-preview")` 返回值逻辑**: 返回 `true` 表示 class 被添加（preview 隐藏），`!isHidden` 即预览可见时调用 `renderPreview()`。✅
- **按钮 active 态**: `classList.toggle("active", !isHidden)` — force=true 当预览可见时。与 CSS `.editor-btn-preview.active` 联动。✅
- **与现有 `cancelEdit` 的解绑**: `previewBtn.addEventListener("click", togglePreview)` 替换原有的 `cancelEdit` 绑定。此时 Preview 按钮从"取消编辑"变为"切换预览面板"，用户取消编辑的唯一方式是通过 Cancel 按钮或 Esc 键 — 行为符合 FR-INT-01。✅

### Task 1.6 — 集成

- **`enterEditMode` 修改点**: 在 `textarea.setSelectionRange(0, 0)` 之后（行 74 之后）新增。此时 textarea 已有内容、已聚焦、`toggleUI(true)` 已调用。✅
- **`toggleUI` 无需修改**: 核心逻辑仍通过 `editorContainer.style.display` 控制整个 `.editor-container`（包含 split-wrapper + preview）。通过 CSS `[data-editing="true"]` 属性后代选择器激活所有 split 布局。✅
- **依赖关系图的错误**: §七 依赖图仅显示 `Task 1.1 + 1.2 + 1.3 → Task 1.6`，但 Task 1.6 自身描述明确指出依赖 `Task 1.4` 和 `Task 1.5`。**依赖图不完整**。⚠️ 主要

### Task 1.7 — 主题兼容

- **`.markdown-body` 后代选择器**: preview div 在 `<body class="markdown-body">` 内，其内的 h1/p/pre/code 等元素通过 github-markdown-*.css 的 `.markdown-body h1` 等规则自动获得样式。✅
- **主题切换**: `theme-switch.js` 修改 `data-theme` 属性 + `<link media>`，CSS `[data-theme]` 选择器自动跟随。无需 JS 重新渲染。✅
- **插入位置验证**: `editor.css:112`（`@media (prefers-color-scheme: dark)` 块末），`editor.css:152`（`[data-theme="dark"]` 块末），`editor.css:191`（`[data-theme="light"]` 块末）。均正确。✅

### Task 2.2 — 可拖拽分隔条

- **事件监听器泄漏风险**: 计划描述了 mousedown → 注册 document mousemove/mouseup，mouseup → 注销。但未提及如果用户在 `mousedown` 后切换到另一个标签页、或按 Esc 取消拖拽、或浏览器失去焦点时，监听器是否会被清理。实际实现需在 document 上注册 `mouseleave` 或 `blur` 兜底清理。⚠️ 次要
- **sessionStorage API**: 计划写的是 `sessionStorage("go-grip-split-ratio")`，应为 `sessionStorage.setItem("go-grip-split-ratio", ...)`。这是伪代码简写，非实际代码。✅（文档层面可接受）
- **`preview.style.flex = "1"`**: 拖拽后的恢复。如果用户拖拽后隐藏预览（no-preview），再展开，百分比会被重置。计划应明确展开时是否恢复 sessionStorage 中的比例。⚠️ 次要

---

## 四、任务依赖与排序审查

### 4.1 依赖图问题

计划 §七 的依赖图 vs Task 自身描述的依赖声明存在以下差异：

| 关系 | §七 依赖图 | Task 自身描述 | 差异 |
|------|-----------|-------------|------|
| Task 1.6 → Task 1.4 | **未显示** | "依赖: Task 1.4" | **缺失** — Task 1.6 的 `enterEditMode` 调用 `renderPreview`（定义在 1.4）和 `scheduleRender` |
| Task 1.6 → Task 1.5 | **未显示** | "依赖: Task 1.5" | **缺失** — Task 1.6 的 `enterEditMode` 通过移除 `no-preview` class 依赖 1.5 定义的概念 |
| Task 1.7 → Task 1.6 | **显示** | "依赖: Task 1.3" | **过度约束** — 主题 CSS 不需要 JS 集成，仅依赖 1.3 的布局 CSS 即可 |

**建议修正的依赖关系**:
```
Task 1.1 (marked)
  └── Task 1.4 (渲染引擎)
Task 1.2 (HTML)
  ├── Task 1.3 (CSS 布局) ──── Task 1.7 (主题 CSS)
  ├── Task 1.4 (渲染引擎)
  └── Task 1.5 (按钮行为)
Task 1.2 + 1.4 + 1.5
  └── Task 1.6 (集成) ← 新增依赖 1.4/1.5
```

### 4.2 无循环依赖

✅ 确认依赖图为有向无环图。

### 4.3 并行化机会

计划已识别的并行路径: Task 1.1 ∥ 1.2、Task 1.3 ∥ 1.4。✅

**额外可并行**:
- Task 1.7（主题 CSS）可与 Task 1.5/1.6 并行（仅依赖 1.3）
- Task 1.8（文档注释）实质是 Task 1.4 的子任务（在同一个 `renderPreview` 函数中），可合并
- Task 3.1 ∥ 3.2（互不依赖）

### 4.4 Phase 排序

Phase 1 (P0) → Phase 2 (P1) → Phase 3 (P2) 正确。无需将 P1 提升为 P0。

---

## 五、风险评估

### 5.1 风险最高的任务（Top 3）

| 排名 | 任务 | 风险 | 缓解措施 |
|------|------|------|---------|
| 1 | **Task 1.3** CSS 布局 | **高** — 涉及最复杂的 CSS 特定性管理；`scroll-behavior: auto` 无 `!important` 导致规则不生效；`.editor-textarea` 在两套规则（基础 + data-editing + split-wrapper）之间可能产生意外交互 | 修复 `!important` 问题；在 3 个断点 (360/768/1440px) 全面测试 |
| 2 | **Task 2.2** 拖拽分隔条 | **中高** — 事件监听器生命周期管理复杂；`mousemove` 高频触发可能导致布局抖动；与 `no-preview` 状态的交互未完全定义 | 添加 `mouseleave`/`blur` 兜底清理；在 `<768px` 堆叠模式禁用拖拽（已计划） |
| 3 | **Task 1.6** 集成 | **中** — 修改 `enterEditMode` 和 `exitEditMode` 两个关键路径；任何未捕获异常都会破坏编辑模式的基础功能 | 所有 DOM 操作加 null guard（计划已有）；充分的回归测试 |

### 5.2 最可能回归的现有功能

| 功能 | 风险等级 | 原因 |
|------|---------|------|
| 编辑模式 textarea 可用性 | **中** | 新的 split-wrapper 规则 override 了 textarea 的 `resize: vertical`；在 no-preview 模式下 textarea 被 split-wrapper 的 flex:1 规则约束而非直接由 `.editor-container` 控制 |
| 编辑模式退出后的 DOM 清理 | **中** | `exitEditMode` 新增了 3 个清理步骤，如果 preview/wrapper 元素不存在且 null guard 缺失，会导致 JS 错误 |
| 768-1023px 范围内布局 | **低** | 计划未添加此范围的显式 media query，依赖 flex:row 自然适应。但如果 panel 内容过宽，可能出现水平溢出 |
| 侧边栏切换文件 | **低** | `interceptSidebarLinks`（`editor.js:218`）不感知分屏；`showNavDialog`（`:238`）中的保存逻辑直接调用 `/api/edit/`，不涉及 preview panel |

### 5.3 测试覆盖缺口

| 缺口 | 严重度 | 说明 |
|------|--------|------|
| 无 JS 单元测试 | 次要 | renderPreview / togglePreview / exitEditMode (split 部分) 无自动化测试 |
| NFR-SPLIT-04 无验证 | 主要 | 跨浏览器兼容性测试未出现在任何 Task 或验证清单中 |
| NFR-SPLIT-02 无测量 | 次要 | 16ms/100ms 性能指标无明确的测量方法或工具建议 |
| FR-SPLIT-12 独立滚动的验证缺失 | 主要 | scroll-behavior 规则如不修复，独立滚动条行为将失效（smooth 滚动仍生效），但验证清单不会发现此问题，因为"各自独立滚动条"不等于"instant scroll" |
| `sessionStorage` 持久化的重启行为 | 次要 | 拖拽比例存入 sessionStorage，但退出编辑/刷新后再进入是否恢复比例未定义 |

---

## 六、问题清单

### 严重 (Blocking)

1. **`scroll-behavior: auto` 无法覆盖 `!important`** (Task 1.3, 计划行 225)

   `docs-layout.css:2` 的规则 `.markdown-body { scroll-behavior: smooth !important }` 使用 `!important`，且后于 `github-markdown-light.css` 加载。计划的 `.editor-split-preview { scroll-behavior: auto }`（无 `!important`）在 CSS 级联中必定落败。预览面板将保持 smooth 滚动，违反 FR-SPLIT-12（独立滚动条）。

   **修复**: 改为 `scroll-behavior: auto !important;`

2. **FR-LIVE-03 无对应 Task** (Phase 3 概述 vs Task 列表)

   计划 §一 "总体概览" 表中 Phase 3 声称覆盖 FR-LIVE-03（手动刷新预览），但 Phase 3 只有 Task 3.1（空内容占位）和 Task 3.2（Ctrl+P），没有一个 Task 实现"工具栏按钮手动触发即时刷新"。Task 1.5 中的展开时 `renderPreview()` 只能算间接覆盖。

   **修复**: 在 Phase 3 新增 Task 3.3 或明确声明 FR-LIVE-03 由 Task 1.5 的 togglePreview 行为覆盖，并在验证清单中增加对应检查项。

3. **Task 1.6 依赖关系图不完整** (§七 依赖图)

   依赖图显示 `Task 1.1 + 1.2 + 1.3 → Task 1.6`，但 Task 1.6 自身描述声明依赖 `Task 1.4`（renderPreview + scheduleRender）和 `Task 1.5`（togglePreview）。依赖图中缺失这两条边。

   **修复**: 添加 `Task 1.4 → Task 1.6` 和 `Task 1.5 → Task 1.6`。

### 主要 (Should Fix)

4. **NFR-SPLIT-04 跨浏览器兼容无任务覆盖** (P0)

   无任何 Task 或验证项覆盖 Chrome/Firefox/Edge 兼容性测试。作为 P0 需求，至少应在 Phase 1 验证清单中增加对应的手动测试项。

   **修复**: 在验证清单中添加 Chrome/Firefox/Edge 最新两个大版本的功能验证项。

5. **`@media (prefers-color-scheme: dark)` 暗色规则缺少 `[data-editing="true"]` 前缀** (Task 1.3 step5)

   计划提议：
   ```css
   @media (prefers-color-scheme: dark) {
     .editor-split-preview { ... }  /* 无 [data-editing] 前缀 */
   }
   ```
   而对应的 `[data-theme="dark"]` 规则使用 `[data-theme="dark"] .editor-split-preview`。虽然当前 preview 仅在编辑模式下可见（父元素 `display: none`），但风格不一致，且为未来的意外可见留下隐患。

   **修复**: 统一为 `[data-editing="true"] .editor-split-preview`。

6. **`exitEditMode` 修改在两处重复描述** (Task 1.4 step5 和 Task 1.6 step2)

   Task 1.4 详细描述了 `exitEditMode()` 的修改（清除 debounceTimer、清空 preview、重置 no-preview），Task 1.6 又引用它。实施时如需分 Task 提交代码，可能导致 merge conflict。

   **修复**: 将 `exitEditMode` 修改明确归属到一个 Task（建议 Task 1.6），Task 1.4 仅描述 `renderPreview` + `scheduleRender` + `textarea.oninput` 修改。

7. **FR-LIVE-05 大文档性能提示无覆盖** (P2)

   需求 FR-LIVE-05（"超过 100KB 时提示"）在整个计划中从未出现。虽然优先级 P2，但应明确声明：是降级不做、还是有隐含覆盖、还是遗漏。

   **修复**: 在 §一 表或 Phase 3 中说明 FR-LIVE-05 决定不做（或新增 Task 3.x）。

### 次要 (Nice to Fix)

8. **Task 2.2 拖拽事件清理不完整**

   计划描述了 mousedown → mousemove → mouseup 的正常流程，但缺乏异常场景（浏览器失去焦点、用户按 Esc、标签页切换）下的监听器清理。建议添加 `document` 级别的 `mouseleave` 兜底。

9. **Task 2.2 sessionStorage 恢复时机未定义**

   拖拽比例写入 sessionStorage 后，用户退出编辑模式再进入时是否需要恢复比例？计划未说明。建议在 `enterEditMode` 中增加比例恢复逻辑。

10. **preview 面板缺少 `overflow-x` 控制** (Task 1.3)

    `.editor-split-preview` 设置了 `overflow-y: auto` 但未设置 `overflow-x`。如果 marked 生成的代码块或表格非常宽，可能超出 preview 面板导致水平溢出到容器外。建议添加 `overflow-x: auto` 或 `overflow-wrap: break-word`。

11. **Task 1.4 的 `marked.parse()` 在大文档上可能耗时 >100ms**

    `requestAnimationFrame` 将渲染推迟到下一帧，但如果 `marked.parse()` 本身耗时 >16ms，仍会掉帧。在 5000+ 行文档上未做分段渲染（Phase 2 仅调整防抖延迟，不减少 `marked.parse` 的执行时间）。可作为 Phase 2 的已知限制在文档中记录。

---

## 七、建议改进

1. **关键技术修正**: 修复 Task 1.3 的 `scroll-behavior: auto !important`，这是唯一的技术阻塞项。
2. **依赖关系图补全**: 添加 Task 1.4→1.6、Task 1.5→1.6 的依赖边。
3. **Phase 3 补全**: 为 FR-LIVE-03 新增一个 Task 或声明由 Task 1.5 覆盖。为 FR-LIVE-05 明确处置意见（纳入 Phase 3 或不做）。
4. **验证清单增强**: 添加跨浏览器验证项（NFR-SPLIT-04）、`scroll-behavior` 独立滚动验证项、以及拖拽比例 sessionStorage 恢复验证项。
5. **CSS 风格统一**: 所有 `.editor-split-preview` / `.editor-split-wrapper` 规则统一使用 `[data-editing="true"]` 前缀。
6. **Task 合并建议**: Task 1.8（XSS 文档注释）可合并入 Task 1.4（注释本来就写在 renderPreview 中），减少 Task 数量但不减少工作量。
7. **拖拽实现补充**: Task 2.2 需补充异常场景的事件监听器清理、展开预览时的比例恢复、以及 `<768px` 禁用拖拽的 CSS（`pointer-events: none`）。

---

*Review 日期：2026-05-17*
*基于: editing-split-screen-implementation.md v1.0 / editing-split-screen-requirements.md v1.0 / editing-split-screen-review.md v1.0 / editor.js / editor.css / layout.html / docs-layout.css / github-markdown-light.css / github-markdown-dark.css / server.go*
