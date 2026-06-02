# 编辑器图片导入功能 — 实施方案

> 最后更新: 2026-06-02
>
> 基于 [editor-image-import-plan.md](editor-image-import-plan.md) 的最终方案，拆解为可执行的代码实现步骤。

---

## 一、后端实现 (`internal/import.go`)

辅助函数放在独立文件 `internal/import.go` 中，与 `export.go`/`pdf.go` 的按功能拆分模式一致。

### 1.1 新增 `handleImport` handler

在 `internal/server.go` 的 `newHandlerForTarget` 中注册路由：

```go
mux.HandleFunc("/api/import", s.handleImport)
```

`/api/import` 不带尾部斜杠，使用 `?file=` 查询参数，与 `/export`、`/pdf` 模式一致。

签名模式参考 `handleExport`/`handlePDF`（无 `dir` 参数，使用 `s.rootDir`）：

```go
// 放在 handlePDF 之后
func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
        return
    }

    fileParam := r.URL.Query().Get("file")
    if fileParam == "" {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file parameter"})
        return
    }

    dir := http.Dir(s.rootDir)
    mdAbsPath, err := validateEditPath(dir, fileParam)
    if err != nil {
        writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
        return
    }

    mdDir := filepath.Dir(mdAbsPath)
    imagesDir := filepath.Join(mdDir, "images")

    if err := r.ParseMultipartForm(32 << 20); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
        return
    }
    defer func() { _ = r.MultipartForm.RemoveAll() }()

    file, header, err := r.FormFile("file")
    if err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing file in request"})
        return
    }
    defer func() { _ = file.Close() }()

    contentType := header.Header.Get("Content-Type")
    if !isAllowedImageMIME(contentType) {
        ext := strings.ToLower(filepath.Ext(header.Filename))
        if !isAllowedImageExt(ext) {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only image files are allowed"})
            return
        }
    }

    data, err := io.ReadAll(file)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to read file"})
        return
    }

    finalName, err := writeImageWithDedup(imagesDir, header.Filename, data)
    if err != nil {
        writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to save file"})
        return
    }

    writeJSON(w, http.StatusOK, map[string]string{"path": "images/" + finalName})
}
```

注意：`"time"` 需要在 `server.go` 中新增 import（用于 `uniqueFilename`）。

### 1.2 `internal/import.go` — 辅助函数

```go
package internal

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

var importLocks sync.Map

func isAllowedImageExt(ext string) bool {
    switch ext {
    case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".tiff", ".tif", ".avif":
        return true
    }
    return false
}

func isAllowedImageMIME(mime string) bool {
    switch strings.ToLower(mime) {
    case "image/jpeg", "image/png", "image/gif", "image/webp",
        "image/svg+xml", "image/bmp", "image/tiff", "image/avif":
        return true
    }
    return false
}

func writeImageWithDedup(imagesDir string, filename string, data []byte) (string, error) {
    if err := os.MkdirAll(imagesDir, 0755); err != nil {
        return "", err
    }

    cleanName := filepath.Base(filename)
    if cleanName == "" || cleanName == "." {
        cleanName = "image.png"
    }

    targetPath := filepath.Join(imagesDir, cleanName)
    if info, err := os.Stat(targetPath); err == nil {
        if info.Size() == int64(len(data)) {
            return cleanName, nil
        }
        cleanName = uniqueFilename(imagesDir, cleanName)
        targetPath = filepath.Join(imagesDir, cleanName)
    }

    if err := writeImageToDir(targetPath, data); err != nil {
        return "", err
    }
    return cleanName, nil
}

func uniqueFilename(dir string, filename string) string {
    ext := filepath.Ext(filename)
    base := strings.TrimSuffix(filename, ext)
    for i := 1; i < 1000; i++ {
        candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
        if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
            return candidate
        }
    }
    return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

func writeImageToDir(absPath string, data []byte) error {
    muRaw, _ := importLocks.LoadOrStore(absPath, &sync.Mutex{})
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
```

---

## 二、模板修改 (`defaults/templates/layout.html`)

### 2.1 编辑工具栏新增 Import Image 按钮

```html
<div class="editor-toolbar" style="display:none">
  <button class="editor-btn editor-btn-save">Save</button>
  <button class="editor-btn editor-btn-cancel">Cancel</button>
  <button class="editor-btn editor-btn-import" title="Import Image">
    <svg aria-hidden="true" viewBox="0 0 16 16" fill="currentColor" width="14" height="14">
      <path d="M6.5 1.75a.75.75 0 0 1 1.5 0V5h3.25a.75.75 0 0 1 0 1.5H8v3.25a.75.75 0 0 1-1.5 0V6.5H3.25a.75.75 0 0 1 0-1.5H5V1.75Z"/>
      <path d="M3.5 11.75a.75.75 0 0 1 1.5 0v.5c0 .69.56 1.25 1.25 1.25h3.5c.69 0 1.25-.56 1.25-1.25v-.5a.75.75 0 0 1 1.5 0v.5A2.75 2.75 0 0 1 9.75 15h-3.5A2.75 2.75 0 0 1 3.5 12.25v-.5Z"/>
    </svg>
    Import
  </button>
</div>
```

### 2.2 导入对话框（在 `.editor-container` 后、`.preview-content` 前）

```html
<div class="import-dialog-overlay" style="display:none">
  <div class="import-dialog" role="dialog" aria-label="Import Image">
    <div class="import-dialog-header">
      Import Image
      <button class="import-dialog-close" aria-label="Close">&times;</button>
    </div>
    <div class="import-dialog-tabs">
      <button class="import-tab active" data-tab="local">Local File</button>
      <button class="import-tab" data-tab="url">URL</button>
    </div>
    <div class="import-tab-content active" data-tab="local">
      <div class="import-dropzone" id="import-dropzone">
        <div class="import-dropzone-message">
          <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="17 8 12 3 7 8"/>
            <line x1="12" y1="3" x2="12" y2="15"/>
          </svg>
          <p>Drag & drop images or a folder here</p>
          <p class="import-dropzone-sub">or click to browse files</p>
        </div>
        <input type="file" class="import-file-input" multiple accept="image/*" hidden>
        <div class="import-preview-list" style="display:none"></div>
      </div>
    </div>
    <div class="import-tab-content" data-tab="url">
      <div class="import-url-form">
        <label class="import-url-label">Image URL</label>
        <input type="url" class="import-url-input" placeholder="https://example.com/image.png">
      </div>
    </div>
    <div class="import-dialog-footer">
      <button class="editor-btn import-btn-cancel">Cancel</button>
      <button class="editor-btn editor-btn-save import-btn-confirm">Import</button>
    </div>
  </div>
</div>
```

### 2.3 Pending Tray（在 `.editor-container` 后、对话框前）

```html
<div class="import-pending-tray" style="display:none">
  <div class="import-pending-header">
    <span class="import-pending-count">Pending images (0)</span>
    <button class="editor-btn import-pending-insert-all">Insert All</button>
    <button class="import-pending-close" aria-label="Close">&times;</button>
  </div>
  <div class="import-pending-list"></div>
</div>
```

---

## 三、CSS 样式 (`defaults/static/css/editor.css`)

### 3.1 Import Dialog

```css
.import-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  z-index: 9997;
  display: flex;
  align-items: center;
  justify-content: center;
}

.import-dialog {
  background: #ffffff;
  border-radius: 12px;
  width: 520px;
  max-width: 90vw;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.import-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  font-size: 16px;
  font-weight: 600;
  color: #1f2328;
  border-bottom: 1px solid #d0d7de;
}

.import-dialog-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #656d76;
  padding: 0;
  line-height: 1;
}

.import-dialog-close:hover {
  color: #1f2328;
}

.import-dialog-tabs {
  display: flex;
  border-bottom: 1px solid #d0d7de;
  padding: 0 20px;
}

.import-tab {
  padding: 10px 16px;
  font-size: 14px;
  font-weight: 500;
  color: #656d76;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  transition: color 0.15s, border-color 0.15s;
}

.import-tab:hover {
  color: #1f2328;
}

.import-tab.active {
  color: #0969da;
  border-bottom-color: #0969da;
}

.import-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 20px;
  border-top: 1px solid #d0d7de;
}
```

### 3.2 Tab Content & Dropzone

```css
.import-tab-content {
  display: none;
  padding: 20px;
  overflow-y: auto;
}

.import-tab-content.active {
  display: block;
  flex: 1;
  overflow-y: auto;
}

.import-dropzone {
  border: 2px dashed #d0d7de;
  border-radius: 8px;
  padding: 40px 20px;
  text-align: center;
  cursor: pointer;
  transition: border-color 0.15s, background-color 0.15s;
}

.import-dropzone:hover,
.import-dropzone.drag-over {
  border-color: #0969da;
  background-color: rgba(9, 105, 218, 0.05);
}

.import-dropzone-message svg {
  color: #656d76;
  margin-bottom: 8px;
}

.import-dropzone-message p {
  margin: 0;
  font-size: 14px;
  color: #1f2328;
}

.import-dropzone-sub {
  margin-top: 4px !important;
  color: #656d76 !important;
  font-size: 12px !important;
}
```

### 3.3 URL Tab

```css
.import-url-form {
  padding: 20px 0;
}

.import-url-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: #1f2328;
  margin-bottom: 8px;
}

.import-url-input {
  width: 100%;
  padding: 8px 12px;
  font-size: 14px;
  color: #1f2328;
  background: #ffffff;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  outline: none;
  box-sizing: border-box;
}

.import-url-input:focus {
  border-color: #0969da;
  box-shadow: 0 0 0 3px rgba(9, 105, 218, 0.15);
}
```

### 3.4 Preview List

```css
.import-preview-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.import-preview-item {
  position: relative;
  width: 64px;
  height: 64px;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  overflow: hidden;
}

.import-preview-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.import-preview-item-name {
  font-size: 10px;
  color: #656d76;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
```

### 3.5 Pending Tray

```css
.import-pending-tray {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 9995;
  background: #ffffff;
  border-top: 1px solid #d0d7de;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
  max-height: 120px;
  display: flex;
  flex-direction: column;
}

.import-pending-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid #f0f0f0;
  font-size: 12px;
  color: #656d76;
}

.import-pending-count {
  flex: 1;
}

.import-pending-insert-all {
  font-size: 12px !important;
  padding: 2px 10px !important;
}

.import-pending-close {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  color: #656d76;
  padding: 0 4px;
}

.import-pending-list {
  display: flex;
  gap: 8px;
  padding: 8px 16px;
  overflow-x: auto;
  flex: 1;
  align-items: center;
}

.import-pending-item {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
  position: relative;
  transition: border-color 0.15s;
}

.import-pending-item:hover {
  border-color: #0969da;
}

.import-pending-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.import-pending-item-name {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  font-size: 8px;
  color: #ffffff;
  background: rgba(0, 0, 0, 0.5);
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 1px 2px;
}
```

### 3.6 Textarea Drop Overlay

```css
.textarea-drop-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(9, 105, 218, 0.1);
  border: 2px dashed #0969da;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
  pointer-events: none;
}

.textarea-drop-msg {
  font-size: 16px;
  font-weight: 500;
  color: #0969da;
  background: #ffffff;
  padding: 8px 16px;
  border-radius: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}
```

### 3.7 Dark / Light 主题覆盖

遵循 editor.css 现有模式（仅使用 `[data-theme="dark"]` 和 `[data-theme="light"]`，不使用 `@media prefers-color-scheme`，与 `.editor-dialog-overlay`/`.editor-dialog` 一致）。

```css
/* Dialog overlay */
[data-theme="dark"] .import-dialog-overlay {
  background: rgba(0, 0, 0, 0.6);
}

[data-theme="dark"] .import-dialog {
  background: #161b22;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
}

[data-theme="dark"] .import-dialog-header {
  color: #e6edf3;
  border-bottom-color: #30363d;
}

[data-theme="dark"] .import-dialog-close {
  color: #8b949e;
}

[data-theme="dark"] .import-dialog-close:hover {
  color: #e6edf3;
}

[data-theme="dark"] .import-dialog-tabs {
  border-bottom-color: #30363d;
}

[data-theme="dark"] .import-tab {
  color: #8b949e;
}

[data-theme="dark"] .import-tab:hover {
  color: #e6edf3;
}

[data-theme="dark"] .import-tab.active {
  color: #58a6ff;
  border-bottom-color: #58a6ff;
}

[data-theme="dark"] .import-dialog-footer {
  border-top-color: #30363d;
}

/* Dropzone */
[data-theme="dark"] .import-dropzone {
  border-color: #30363d;
}

[data-theme="dark"] .import-dropzone:hover,
[data-theme="dark"] .import-dropzone.drag-over {
  border-color: #58a6ff;
  background-color: rgba(88, 166, 255, 0.1);
}

[data-theme="dark"] .import-dropzone-message svg {
  color: #8b949e;
}

[data-theme="dark"] .import-dropzone-message p {
  color: #e6edf3;
}

[data-theme="dark"] .import-dropzone-sub {
  color: #8b949e !important;
}

/* URL Tab */
[data-theme="dark"] .import-url-label {
  color: #e6edf3;
}

[data-theme="dark"] .import-url-input {
  color: #e6edf3;
  background: #0d1117;
  border-color: #30363d;
}

[data-theme="dark"] .import-url-input:focus {
  border-color: #58a6ff;
  box-shadow: 0 0 0 3px rgba(88, 166, 255, 0.15);
}

/* Preview list */
[data-theme="dark"] .import-preview-item {
  border-color: #30363d;
}

[data-theme="dark"] .import-preview-item-name {
  color: #8b949e;
}

/* Pending tray */
[data-theme="dark"] .import-pending-tray {
  background: #161b22;
  border-top-color: #30363d;
  box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.4);
}

[data-theme="dark"] .import-pending-header {
  border-bottom-color: #21262d;
  color: #8b949e;
}

[data-theme="dark"] .import-pending-close {
  color: #8b949e;
}

[data-theme="dark"] .import-pending-item {
  border-color: #30363d;
}

[data-theme="dark"] .import-pending-item:hover {
  border-color: #58a6ff;
}

/* Drop overlay */
[data-theme="dark"] .textarea-drop-msg {
  background: #161b22;
  color: #58a6ff;
}

[data-theme="dark"] .textarea-drop-overlay {
  background: rgba(88, 166, 255, 0.1);
  border-color: #58a6ff;
}

/* Light theme overrides (match default, follow existing pattern) */
[data-theme="light"] .import-dialog {
  background: #ffffff;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}
```

---

## 四、JavaScript 实现 (`defaults/static/js/editor.js`)

在 IIFE 内部新增以下内容。

### 4.1 新增模块级变量

```javascript
var pendingImages = [];
```

无需声明其他变量：对话框、dropzone、tray 等元素用 `document.querySelector` 实时获取。

### 4.2 `init()` 中新增事件绑定

`init()` 中一次性初始化，避免重复绑定：

```javascript
// 在现有 init 函数末尾

// Import Image 按钮
var importBtn = document.querySelector(".editor-btn-import");
if (importBtn) {
  importBtn.addEventListener("click", function() {
    openImportDialog();
  });
}

// 对话框关闭按钮
document.querySelector(".import-dialog-close")?.addEventListener("click", closeImportDialog);
document.querySelector(".import-btn-cancel")?.addEventListener("click", closeImportDialog);

// Tab 切换
document.querySelectorAll(".import-tab").forEach(function(tab) {
  tab.addEventListener("click", function() {
    switchImportTab(this.dataset.tab);
  });
});

// 确认按钮
document.querySelector(".import-btn-confirm")?.addEventListener("click", handleImportConfirm);

// Pending Tray 按钮
document.querySelector(".import-pending-insert-all")?.addEventListener("click", insertAllPending);
document.querySelector(".import-pending-close")?.addEventListener("click", clearPendingTray);

// 初始化和导入相关的交互
initDropzone();
initTextareaDrop();
initPasteHandler();
```

### 4.3 对话框打开/关闭

```javascript
function openImportDialog() {
  var overlay = document.querySelector(".import-dialog-overlay");
  if (!overlay) return;
  overlay.style.display = "";
  resetDropzone();
  switchImportTab("local");
}

function closeImportDialog() {
  var overlay = document.querySelector(".import-dialog-overlay");
  if (!overlay) return;
  overlay.style.display = "none";
  var input = document.querySelector(".import-file-input");
  if (input) input.value = "";
}

function resetDropzone() {
  var list = document.querySelector(".import-preview-list");
  if (list) { list.style.display = "none"; list.innerHTML = ""; }
  var msg = document.querySelector(".import-dropzone-message");
  if (msg) msg.style.display = "";
}
```

### 4.4 Tab 切换

```javascript
function switchImportTab(tabName) {
  document.querySelectorAll(".import-tab").forEach(function(t) {
    t.classList.toggle("active", t.dataset.tab === tabName);
  });
  document.querySelectorAll(".import-tab-content").forEach(function(tc) {
    tc.classList.toggle("active", tc.dataset.tab === tabName);
  });
  var btn = document.querySelector(".import-btn-confirm");
  if (btn) {
    btn.textContent = tabName === "local" ? "Import" : "Insert";
  }
}
```

### 4.5 确认按钮逻辑

```javascript
function handleImportConfirm() {
  var activeTab = document.querySelector(".import-tab.active");
  if (!activeTab) return;

  if (activeTab.dataset.tab === "url") {
    var urlInput = document.querySelector(".import-url-input");
    var url = urlInput ? urlInput.value.trim() : "";
    if (!url) {
      showToast("Please enter an image URL", "error");
      return;
    }
    insertTextAtCursor("![](" + url + ")\n");
    closeImportDialog();
    return;
  }

  var previewItems = document.querySelectorAll(".import-preview-item");
  if (previewItems.length === 0) {
    showToast("No images selected", "error");
    return;
  }

  var confirmBtn = this;
  confirmBtn.disabled = true;
  confirmBtn.textContent = "Importing...";

  var filesToImport = [];
  previewItems.forEach(function(item) {
    filesToImport.push({
      file: item._file,
      name: item._fileName
    });
  });

  importAllFiles(filesToImport, function(successCount) {
    confirmBtn.disabled = false;
    confirmBtn.textContent = "Import";
    if (successCount > 0) {
      closeImportDialog();
    }
  });
}
```

### 4.6 批量导入文件

```javascript
function importAllFiles(files, callback) {
  var completed = 0;
  var successCount = 0;
  var total = files.length;
  var pendingAdditions = [];

  files.forEach(function(f) {
    importSingleFile(f.file, f.name, function(path, fileObj) {
      completed++;
      if (path) {
        successCount++;
        pendingAdditions.push({
          name: path.split("/").pop(),
          path: path,
          file: fileObj
        });
      }
      if (completed === total) {
        // FileReader 全部完成后统一更新 tray
        var readerCount = pendingAdditions.length;
        if (readerCount === 0) {
          if (callback) callback(successCount);
          return;
        }
        pendingAdditions.forEach(function(item) {
          var reader = new FileReader();
          reader.onload = function(e) {
            pendingImages.push({
              name: item.name,
              path: item.path,
              dataUrl: e.target.result
            });
            readerCount--;
            if (readerCount === 0) {
              updatePendingTray();
              if (callback) callback(successCount);
            }
          };
          reader.readAsDataURL(item.file);
        });
      }
    });
  });
}
```

### 4.7 单文件导入 API 调用

```javascript
function importSingleFile(file, fileName, callback) {
  var formData = new FormData();
  formData.append("file", file, fileName);

  var url = "/api/import?file=" + encodeURIComponent(currentFile);

  fetch(url, {
    method: "POST",
    body: formData
  })
  .then(function(resp) {
    if (!resp.ok) {
      return resp.json().then(function(data) {
        throw new Error(data.error || "Import failed");
      });
    }
    return resp.json();
  })
  .then(function(data) {
    callback(data.path, file);
  })
  .catch(function(err) {
    showToast("Import failed: " + err.message, "error");
    callback(null, null);
  });
}
```

### 4.8 Pending Tray 管理

```javascript
function updatePendingTray() {
  var tray = document.querySelector(".import-pending-tray");
  var list = document.querySelector(".import-pending-list");
  var count = document.querySelector(".import-pending-count");
  if (!tray || !list) return;

  if (pendingImages.length === 0) {
    tray.style.display = "none";
    return;
  }

  tray.style.display = "";
  if (count) count.textContent = "Pending images (" + pendingImages.length + ")";

  list.innerHTML = "";
  pendingImages.forEach(function(img, index) {
    var item = document.createElement("div");
    item.className = "import-pending-item";
    item.title = img.name;

    var imgEl = document.createElement("img");
    imgEl.src = img.dataUrl;
    imgEl.alt = img.name;
    item.appendChild(imgEl);

    var nameEl = document.createElement("div");
    nameEl.className = "import-pending-item-name";
    nameEl.textContent = img.name;
    item.appendChild(nameEl);

    item.addEventListener("click", function() {
      insertImageAtIndex(index);
    });

    list.appendChild(item);
  });
}

function insertImageAtIndex(index) {
  if (index < 0 || index >= pendingImages.length) return;
  var img = pendingImages[index];
  insertTextAtCursor("![](" + img.path + ")\n");
  pendingImages.splice(index, 1);
  updatePendingTray();
}

function insertAllPending() {
  if (pendingImages.length === 0) return;
  var text = pendingImages.map(function(img) {
    return "![](" + img.path + ")";
  }).join("\n") + "\n";
  insertTextAtCursor(text);
  pendingImages = [];
  updatePendingTray();
}

function clearPendingTray() {
  pendingImages = [];
  updatePendingTray();
}
```

### 4.9 在光标位置插入文本

```javascript
function insertTextAtCursor(text) {
  var textarea = document.querySelector(".editor-textarea");
  if (!textarea) return;

  var start = textarea.selectionStart;
  var end = textarea.selectionEnd;

  var before = textarea.value.substring(0, start);
  var after = textarea.value.substring(end);

  textarea.value = before + text + after;

  var newPos = start + text.length;
  textarea.setSelectionRange(newPos, newPos);
  textarea.focus();

  var event = new Event("input", { bubbles: true });
  textarea.dispatchEvent(event);
}
```

### 4.10 对话框 Dropzone 拖拽支持

```javascript
var dropzoneInitialized = false;

function initDropzone() {
  if (dropzoneInitialized) return;
  var dropzone = document.querySelector(".import-dropzone");
  var fileInput = document.querySelector(".import-file-input");
  if (!dropzone || !fileInput) return;
  dropzoneInitialized = true;

  dropzone.addEventListener("click", function(e) {
    if (e.target.closest(".import-preview-item")) return;
    fileInput.click();
  });

  fileInput.addEventListener("change", function() {
    if (this.files && this.files.length > 0) {
      handleFilesSelected(this.files);
    }
    this.value = "";
  });

  ["dragenter", "dragover"].forEach(function(event) {
    dropzone.addEventListener(event, function(e) {
      e.preventDefault();
      e.stopPropagation();
      dropzone.classList.add("drag-over");
    });
  });

  ["dragleave", "drop"].forEach(function(event) {
    dropzone.addEventListener(event, function(e) {
      e.preventDefault();
      e.stopPropagation();
      dropzone.classList.remove("drag-over");
    });
  });

  dropzone.addEventListener("drop", function(e) {
    var items = e.dataTransfer.items;
    if (items && items.length > 0) {
      handleDroppedItems(items);
    } else if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFilesSelected(e.dataTransfer.files);
    }
  });
}
```

### 4.11 处理拖拽的目录（webkitGetAsEntry）

```javascript
function handleDroppedItems(items) {
  var allFiles = [];

  function processEntry(entry, callback) {
    if (entry.isFile) {
      entry.file(function(file) {
        allFiles.push(file);
        callback();
      }, callback);
    } else if (entry.isDirectory) {
      var reader = entry.createReader();
      var MAX_DEPTH = 10;
      readEntries(reader, entry.fullPath.split("/").length, function() {
        callback();
      });
    } else {
      callback();
    }
  }

  function readEntries(reader, startDepth, callback, depth) {
    depth = depth || 0;
    if (depth >= MAX_DEPTH) { callback(); return; }
    reader.readEntries(function(entries) {
      if (entries.length === 0) {
        callback();
        return;
      }
      var pending = entries.length;
      entries.forEach(function(entry) {
        processEntry(entry, function() {
          pending--;
          if (pending === 0) {
            readEntries(reader, startDepth, callback, depth + 1);
          }
        });
      });
    }, callback);
  }

  var fileEntries = [];
  for (var i = 0; i < items.length; i++) {
    var entry = items[i].webkitGetAsEntry ? items[i].webkitGetAsEntry() : null;
    if (entry) {
      fileEntries.push(entry);
    }
  }

  if (fileEntries.length === 0) {
    if (items.length > 0 && items[0].getAsFile) {
      var files = [];
      for (var i = 0; i < items.length; i++) {
        var f = items[i].getAsFile();
        if (f) files.push(f);
      }
      if (files.length > 0) handleFilesSelected(asFileList(files));
    }
    return;
  }

  var pending = fileEntries.length;
  fileEntries.forEach(function(entry) {
    processEntry(entry, function() {
      pending--;
      if (pending === 0) {
        handleFilesSelected(asFileList(allFiles));
      }
    });
  });
}

function asFileList(files) {
  return {
    length: files.length,
    item: function(i) { return files[i]; }
  };
}
```

### 4.12 处理选择的文件

```javascript
function handleFilesSelected(fileList) {
  var imageFiles = [];
  for (var i = 0; i < fileList.length; i++) {
    var file = fileList[i];
    if (file.type && file.type.startsWith("image/")) {
      imageFiles.push(file);
    }
  }

  if (imageFiles.length === 0) {
    showToast("No image files found", "error");
    return;
  }

  showPreviewItems(imageFiles);
}

function showPreviewItems(files) {
  var list = document.querySelector(".import-preview-list");
  var msg = document.querySelector(".import-dropzone-message");
  if (!list) return;

  msg.style.display = "none";
  list.style.display = "";
  list.innerHTML = "";

  files.forEach(function(file) {
    var item = document.createElement("div");
    item.className = "import-preview-item";

    var img = document.createElement("img");
    var objectUrl = URL.createObjectURL(file);
    img.src = objectUrl;
    item.appendChild(img);

    var name = document.createElement("div");
    name.className = "import-preview-item-name";
    name.textContent = file.name;
    item.appendChild(name);

    item._file = file;
    item._fileName = file.name;
    item._objectUrl = objectUrl;

    list.appendChild(item);
  });
}
```

### 4.13 Textarea 拖拽支持

```javascript
var textareaDropInitialized = false;

function initTextareaDrop() {
  if (textareaDropInitialized) return;
  var textarea = document.querySelector(".editor-textarea");
  if (!textarea) return;
  textareaDropInitialized = true;

  var dropOverlay = document.createElement("div");
  dropOverlay.className = "textarea-drop-overlay";
  dropOverlay.innerHTML = '<div class="textarea-drop-msg">Drop images to import</div>';
  dropOverlay.style.display = "none";
  textarea.parentNode.appendChild(dropOverlay);

  ["dragenter", "dragover"].forEach(function(event) {
    textarea.addEventListener(event, function(e) {
      if (!e.dataTransfer.types || !Array.from(e.dataTransfer.types).includes("Files")) return;
      e.preventDefault();
      e.stopPropagation();
      dropOverlay.style.display = "";
    });
  });

  ["dragleave", "drop"].forEach(function(event) {
    textarea.addEventListener(event, function(e) {
      dropOverlay.style.display = "none";
    });
  });

  textarea.addEventListener("drop", function(e) {
    e.preventDefault();
    e.stopPropagation();

    var items = e.dataTransfer.items;
    if (!items || items.length === 0) return;

    // 统一处理所有文件：先收集再判断数量
    if (items[0].webkitGetAsEntry) {
      // Chrome/Edge: 通过 entry API 获取所有文件
      var allFiles = [];
      var pending = items.length;
      var collected = false;

      for (var i = 0; i < items.length; i++) {
        var entry = items[i].webkitGetAsEntry();
        if (!entry) { pending--; continue; }
        collectFilesFromEntry(entry, allFiles, function() {
          pending--;
          if (pending === 0) {
            collected = true;
            processTextareaDropFiles(allFiles);
          }
        });
      }
      // 如果没有 entry 可处理（fallback），走 files 路径
      if (pending === items.length) {
        processTextareaDropFiles(e.dataTransfer.files);
      }
    } else {
      // Firefox 降级
      processTextareaDropFiles(e.dataTransfer.files);
    }
  });
}

function collectFilesFromEntry(entry, result, callback) {
  if (entry.isFile) {
    entry.file(function(file) {
      result.push(file);
      callback();
    }, callback);
  } else if (entry.isDirectory) {
    var reader = entry.createReader();
    var MAX_DEPTH = 10;
    readDirEntries(reader, result, callback, 0);
  } else {
    callback();
  }
}

function readDirEntries(reader, result, callback, depth) {
  if (depth >= 10) { callback(); return; }
  reader.readEntries(function(entries) {
    if (entries.length === 0) {
      callback();
      return;
    }
    var pending = entries.length;
    entries.forEach(function(entry) {
      collectFilesFromEntry(entry, result, function() {
        pending--;
        if (pending === 0) {
          readDirEntries(reader, result, callback, depth + 1);
        }
      });
    });
  }, callback);
}

function processTextareaDropFiles(files) {
  var imageFiles = [];
  for (var i = 0; i < files.length; i++) {
    var file = files[i];
    if (file.type && file.type.startsWith("image/")) {
      imageFiles.push(file);
    }
  }
  if (imageFiles.length === 0) return;

  var dropPos = document.querySelector(".editor-textarea").selectionStart;

  if (imageFiles.length === 1) {
    importSingleFile(imageFiles[0], imageFiles[0].name, function(path) {
      if (path) {
        var textarea = document.querySelector(".editor-textarea");
        textarea.focus();
        textarea.selectionStart = textarea.selectionEnd = dropPos;
        insertTextAtCursor("![](" + path + ")\n");
      }
    });
  } else {
    importAllFiles(imageFiles.map(function(f) {
      return { file: f, name: f.name };
    }), function() {});
  }
}
```

### 4.14 剪贴板粘贴处理

```javascript
var pasteHandlerInitialized = false;

function initPasteHandler() {
  if (pasteHandlerInitialized) return;
  var textarea = document.querySelector(".editor-textarea");
  if (!textarea) return;
  pasteHandlerInitialized = true;

  textarea.addEventListener("paste", function(e) {
    var items = e.clipboardData && e.clipboardData.items;
    if (!items) return;

    var imageItems = [];
    for (var i = 0; i < items.length; i++) {
      if (items[i].type && items[i].type.startsWith("image/")) {
        imageItems.push(items[i]);
      }
    }

    if (imageItems.length === 0) return;

    e.preventDefault();

    if (imageItems.length === 1) {
      var item = imageItems[0];
      var file = item.getAsFile();
      if (!file) return;

      var now = new Date();
      var ts = now.getFullYear() +
        String(now.getMonth() + 1).padStart(2, "0") +
        String(now.getDate()).padStart(2, "0") + "_" +
        String(now.getHours()).padStart(2, "0") +
        String(now.getMinutes()).padStart(2, "0") +
        String(now.getSeconds()).padStart(2, "0");
      var fileName = file.name || "pasted_" + ts + ".png";

      importSingleFile(file, fileName, function(path) {
        if (path) {
          var pos = textarea.selectionStart;
          textarea.focus();
          textarea.selectionStart = textarea.selectionEnd = pos;
          insertTextAtCursor("![](" + path + ")\n");
        }
      });
    } else {
      var files = [];
      imageItems.forEach(function(item, idx) {
        var f = item.getAsFile();
        if (f) {
          var now = new Date();
          var ts = now.getFullYear() +
            String(now.getMonth() + 1).padStart(2, "0") +
            String(now.getDate()).padStart(2, "0") + "_" +
            String(now.getHours()).padStart(2, "0") +
            String(now.getMinutes()).padStart(2, "0") +
            String(now.getSeconds()).padStart(2, "0");
          var name = f.name || "pasted_" + ts + "_" + idx + ".png";
          files.push({ file: f, name: name });
        }
      });
      if (files.length > 0) {
        importAllFiles(files, function() {});
      }
    }
  });
}
```

### 4.15 `enterEditMode()` 中集成

不需要在 `enterEditMode()` 中初始化——`initDropzone()`、`initTextareaDrop()`、`initPasteHandler()` 已通过 `dropzoneInitialized` / `textareaDropInitialized` / `pasteHandlerInitialized` 标志在 `init()` 中一次性初始化。

`exitEditMode()` 中增加 pending 清理：

```javascript
// 在 exitEditMode 中（在现有清理代码后面添加）
pendingImages = [];
var tray = document.querySelector(".import-pending-tray");
if (tray) tray.style.display = "none";
```

### 4.16 释放 objectURL

为防止内存泄漏，在对话框关闭时释放预览缩略图的 `objectURL`：

```javascript
function closeImportDialog() {
  var overlay = document.querySelector(".import-dialog-overlay");
  if (!overlay) return;
  overlay.style.display = "none";
  var input = document.querySelector(".import-file-input");
  if (input) input.value = "";
  // 释放 preview item 的 objectURL
  document.querySelectorAll(".import-preview-item").forEach(function(item) {
    if (item._objectUrl) URL.revokeObjectURL(item._objectUrl);
  });
}
```

---

## 五、测试要点

### 5.1 后端测试

| 测试场景 | 方法 |
|---|---|
| 正常导入图片 | `POST /api/import?file=test.md` 带 multipart → 检查 `images/` 目录文件 |
| 重复文件名 + 相同 size | 两次导入同文件 → 返回相同 path，不创建副本 |
| 重复文件名 + 不同 size | 第二次导入 → 返回 `photo_1.png` |
| 非图片 MIME | 上传 `.pdf` → 400 error |
| 缺少 `?file` | 无参数 → 400 error |
| 路径穿越 `?file=../../etc` | → 403 error |
| 目录自动创建 | 删除 `images/` 后导入 → 目录重新创建 |

### 5.2 前端测试

| 测试场景 | 方法 |
|---|---|
| 点击 Import Image 按钮 | 对话框弹出 |
| Tab 切换 | Local / URL 切换正确 |
| 文件选择器选图 | 预览缩略图显示，点击 Import 后导入并关闭 |
| 拖拽单张图片到对话框 | 预览显示，可导入 |
| 拖拽目录到对话框 | 目录内图片递归扫描，预览列出 |
| 拖拽单张到 textarea | 自动导入并直接插入光标位置 |
| 拖拽多张到 textarea | 全部导入进 pending |
| URL 输入链接 | 插入 `![](url)\n` 到光标位置 |
| Pending 点击单张 | 插入到光标位置，列表减少 |
| Insert All | 全部插入，每行一张 |
| 粘贴截图 (Ctrl+V) | 单张自动导入并直接插入 |
| 粘贴多文件 | 多张进 pending |
| 关闭 pending tray | 清空列表 |
| 退出编辑模式 | pending 清空 |

---

## 六、实施顺序

```
Step 1: 后端 handleImport + internal/import.go 辅助函数 → 测试
Step 2: HTML 模板修改（按钮 + 对话框 + tray）
Step 3: CSS 样式（含 dark/light 主题）
Step 4: JS 核心逻辑（对话框 + API 调用）
Step 5: JS 拖拽/粘贴/插入逻辑
Step 6: 集成测试 + 边界情况处理
```
