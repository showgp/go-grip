# GoGrip macOS 状态栏 App 实现方案

> 状态（2026-10-05）：本文是旧示例 App 的设计记录，其拖拽/历史/Finder Sync/嵌入 Resources 方案已被替换（旧历史存储源与测试已随任务 09 移除，用户旧 50 项数据保留但不被新宿主读取）；当前宿主按 `openspec/changes/rebuild-macos-preview-app` 重建，能力与边界以 `docs/ARCHITECTURE.md` 第七节和 `README.md` 为准。

> 创建日期: 2026-06-03
> 最后更新: 2026-06-03（Review 修复后）
> 状态: 设计完成，待实现

## 一、项目概述

在 macOS 状态栏（菜单栏）放置一个轻量 App，嵌入 go-grip 二进制，用户通过拖拽目录/.md 文件快速打开预览，并管理历史打开记录。

### 核心价值

- 无需终端操作，拖拽即用
- 记住历史打开的目录/文件，一键重返
- 多实例并行管理，状态一目了然

---

## 二、技术选型

### 方案对比

| 方案 | 可行性 | 结论 |
|------|--------|------|
| **Swift/AppKit (NSStatusItem + NSPopover)** | ✅ | **推荐**。生产级方案，1Password、Bartender 等均用此模式 |
| SwiftUI MenuBarExtra | ⚠️ | `.window` 样式不可调大小、无 popover 三角箭头、Settings 在 macOS 14/15 有 bug |
| Wails v2 | ❌ | 无系统托盘 API，社区确认 "impossible" |
| Wails v3 | ⚠️ | Alpha，API 不稳定，NSPanel PR 未合并 |
| Go systray 库 (getlantern/menuet) | ❌ | 仅支持 NSMenu 菜单，无法显示自定义 popover 视图 |

### 最终方案

**Swift/AppKit (NSStatusItem + NSPopover) + 嵌入 Go 二进制子进程**

- Swift 层：状态栏图标、NSPopover 面板、拖拽接收、历史管理、进程管理
- Go 层：go-grip 新增 `--json` 输出模式，作为子进程启动，负责 Markdown 渲染 + 浏览器打开

---

## 三、go-grip 多实例可行性确认

| 问题 | 结论 |
|------|------|
| 是否有单例锁？ | ❌ 无。无 PID 文件、无文件锁、无 mutex |
| 端口如何分配？ | 默认 6419，被占用时自动 +1 递增（最多 100 次） |
| 全局状态冲突？ | `editLocks`/`importLocks` 均为进程内 sync.Map，无跨进程问题 |
| 能否同时运行多个？ | ✅ 完全支持，每个实例独立运行 |

---

## 四、项目结构

```
go-grip/
├── main.go, cmd/, internal/, pkg/    # 现有 Go 代码（新增 --json 输出模式）
├── macos/                             # 新增：Swift App
│   ├── GoGrip.xcodeproj/
│   ├── GoGrip/
│   │   ├── GoGripApp.swift           # @main 入口
│   │   ├── AppDelegate.swift          # NSStatusItem + NSPopover + 进程管理
│   │   ├── Models/
│   │   │   └── HistoryEntry.swift     # 历史记录数据模型
│   │   ├── Services/
│   │   │   └── ProcessManager.swift   # go-grip 子进程跟踪（ObservableObject）
│   │   ├── Views/
│   │   │   ├── PopoverView.swift      # 主面板内容（ZStack: 历史列表 + 拖拽蒙版）
│   │   │   ├── DropOverlayView.swift  # 拖拽蒙版（仅拖拽时显示）
│   │   │   ├── HistoryListView.swift  # 历史列表
│   │   │   └── HistoryRowView.swift   # 单行条目
│   │   └── Utilities/
│   │       └── Storage.swift          # UserDefaults 持久化
│   ├── Info.plist                     # LSUIElement = true（无 Dock 图标）
│   └── GoGrip.entitlements
├── .github/workflows/
│   ├── build.yml                      # 现有 CI（加 macOS job）
│   └── release.yml                    # 现有 Release（加 DMG 打包）
└── Makefile                           # 新增：统一构建入口
```

---

## 五、应用架构

```
┌──────────────────────────────────────────────────┐
│                   macOS 进程                      │
│                                                  │
│  ┌─────────────┐      ┌────────────────────────┐ │
│  │ NSStatusItem│      │     NSPopover          │ │
│  │  (状态栏图标) │─click─▶│  ZStack:              │ │
│  │  badge: "3" │      │  ├─ HistoryListView   │ │
│  └─────────────┘      │  └─ DropOverlayView   │ │
│                       │     (拖拽时覆盖)        │ │
│  ProcessManager       │  [ 打开 ]              │ │
│  ├─ PID 1234: ~/docs │                        │ │
│  ├─ PID 1235: ~/notes│                        │ │
│  └─ PID 1236: a.md   └────────────────────────┘ │
│                                                  │
│  Storage (UserDefaults)                          │
└──────────────────────────────────────────────────┘
          │              │              │
          ▼              ▼              ▼
     ┌────────┐   ┌────────┐   ┌────────┐
     │go-grip │   │go-grip │   │go-grip │  ← 子进程
     │:6419   │   │:6420   │   │:6421   │
     │~/docs  │   │~/notes │   │a.md    │
     └───┬────┘   └───┬────┘   └───┬────┘
         ▼             ▼             ▼
     浏览器标签1    浏览器标签2    浏览器标签3
```

---

## 六、Popover UI 设计

### 状态 1: 无历史记录（空状态 = 拖拽蒙版样式直接展示）

```
┌─────────────────────────────┐
│  GoGrip                 ⚙️  │
├─────────────────────────────┤
│                             │
│     📂                      │
│                             │
│   拖拽目录或 .md 文件到此处    │
│   或点击下方按钮打开           │
│                             │
│        [ 打  开 ]            │
│                             │
├─────────────────────────────┤
│  最近打开（空）               │
└─────────────────────────────┘
```

### 状态 2: 有历史记录（正常浏览）

```
┌─────────────────────────────┐
│  GoGrip                 ⚙️  │
├─────────────────────────────┤
│  📋 最近打开                 │
│  ├── ~/docs/project     ●  │  ● 绿色 = 运行中
│  ├── ~/notes/meeting       │  ○ 灰色 = 已停止
│  └── ~/README.md           │
├─────────────────────────────┤
│        [ 打  开 ]            │
└─────────────────────────────┘
```

### 状态 3: 拖拽中（覆盖蒙版）

```
┌─────────────────────────────┐
│  GoGrip                 ⚙️  │
├─────────────────────────────┤
│  ╔═══════════════════════╗  │
│  ║                       ║  │  半透明蒙版
│  ║     📂                ║  │  覆盖在历史列表上方
│  ║                       ║  │
│  ║    释放以打开           ║  │
│  ║                       ║  │
│  ╚═══════════════════════╝  │
│                             │
├─────────────────────────────┤
│  📋 最近打开（被蒙版遮挡）     │
└─────────────────────────────┘
```

### Popover 尺寸

```swift
width: 320
minHeight: 200   // 空状态
maxHeight: 480   // 有历史时限制高度，内部 ScrollView
```

### 拖拽蒙版样式

```swift
背景: 半透明黑色 overlay (opacity 0.85)
边框: 虚线圆角矩形 (蓝色 accent color)
图标: 大号 SF Symbol "arrow.down.doc"
文字: "释放以打开"
动画: opacity 从 0 到 1, easeInOut(duration: 0.2)
```

### 拖拽期间 Popover 行为

```swift
// 拖拽进入时：阻止 popover 自动关闭
popover.behavior = .applicationDefined

// 拖拽结束（释放或离开）后：恢复自动关闭
popover.behavior = .transient
```

### SwiftUI 拖拽接线（PopoverView.swift）

```swift
struct PopoverView: View {
    @Binding var isDragOver: Bool
    @State private var history: [HistoryEntry] = []
    let processManager: ProcessManager

    var body: some View {
        ZStack {
            // 底层：历史列表
            HistoryListView(history: history, processManager: processManager)

            // 顶层：拖拽蒙版（仅拖拽时显示）
            if isDragOver {
                DropOverlayView()
            }
        }
        .onDrop(of: [.fileURL], isTargeted: $isDragOver) { providers in
            guard let provider = providers.first else { return false }
            provider.loadItem(forTypeIdentifier: UTType.fileURL.identifier, options: nil) { item, _ in
                guard let data = item as? Data,
                      let url = URL(dataRepresentation: data, relativeTo: nil) else { return }
                let path = url.path
                // 验证：是目录 或 .md 文件
                let isDir = (try? FileManager.default.attributesOfItem(atPath: path))?[.type] as? FileAttributeType == .typeDirectory
                let isMarkdown = path.hasSuffix(".md")
                guard isDir || isMarkdown else { return }

                DispatchQueue.main.async {
                    Task { await processManager.start(path: path) }
                    Storage().addToHistory(path: path)
                    history = Storage().load()
                }
            }
            return true
        }
    }
}
```

### 历史条目状态指示器

```swift
if isRunning(path) {
    Circle().fill(.green)           // ● 运行中
        .onTapGesture { stop(path) } // 点击关闭
} else {
    Circle().fill(.gray.opacity(0.3)) // ○ 已停止
}
```

---

## 七、核心交互流程

### 拖拽打开

```
用户拖拽目录/.md → NSPopover 区域接收 NSItemProvider
  → isDragOver = true → 显示拖拽蒙版 (动画)
  → 松开释放 → 解析文件路径 (URL / String)
  → 验证: 是目录 或 .md 文件
  → 检查 ProcessManager: 该路径是否已在运行?
    → 已运行: 用该实例的 port 调用 open http://localhost:{port} 激活浏览器
    → 未运行: ProcessManager.start(path) → 启动 go-grip --json 子进程
             → 从 stdout 读取 {"port": 6419} 获取实际端口
             → 存储 RunningInstance(path, process, port)
  → Storage.addToHistory(path) → 更新历史（已存在则移到顶部 + accessCount++）
  → isDragOver = false → 蒙版消失
  → 更新 badge 数字
```

### 拖拽离开

```
拖拽离开 Popover 区域
  → isDragOver = false → 蒙版动画消失
  → 不执行任何操作
```

### 按钮打开

```
点击 [打开] 按钮
  → 弹出 NSOpenPanel
    → allowsMultipleSelection = false
    → canChooseDirectories = true
    → canChooseFiles = true
    → allowedContentTypes = [.folder, .markdown]
  → 用户选中 → 同拖拽打开流程
  → NSOpenPanel 显示期间，popover 保持打开（不 dismiss）
```

### 点击历史条目

```
点击历史条目 → 同拖拽打开流程（检查运行状态 → 启动或激活）
```

### 关闭实例

```
点击条目旁 ● (绿色圆点)
  → ProcessManager.stop(path) → terminate 子进程
  → 更新 badge 数字
```

### 退出 App

```
applicationWillTerminate
  → ProcessManager.stopAll() → terminate 所有子进程
```

---

## 八、数据模型

```swift
// MARK: - 历史记录条目

struct HistoryEntry: Codable, Identifiable {
    let id: UUID
    let path: String           // 绝对路径
    let displayName: String    // 目录名 或 文件名
    let isDirectory: Bool
    var lastOpened: Date
    var accessCount: Int       // 打开次数
}

// MARK: - 运行中实例

struct RunningInstance {
    let path: String
    let process: Process       // Foundation.Process
    let port: Int              // go-grip 监听的端口
    var pid: Int32 { process.processIdentifier }
}
```

---

## 九、go-grip `--json` 输出模式

### 需求

Swift App 需要知道 go-grip 子进程的实际监听端口，以便在用户点击"已运行"条目时用 `open http://localhost:{port}` 激活浏览器。

### stdout 协议

当 `--json` 启用时：
1. **所有现有状态消息重定向到 stderr**（`fmt.Fprintf(os.Stderr, ...)`）
2. **stdout 仅输出一行 JSON**，格式为：
   ```json
   {"port": 6419, "host": "localhost", "url": "http://localhost:6419/README.md"}
   ```
3. JSON 输出后立即 `os.Stdout.Sync()` flush
4. **端口值必须是 JSON number**（不是字符串），Swift 用 `json["port"] as? Int` 读取

### 代码执行顺序

go-grip `Serve()` 方法的实际顺序（`internal/server.go`）：
```
line 104: actualPort, err := listenOnPort(...)  // 绑定端口
line 109: initialPath, err := initialPathForTarget(...)  // 确定初始路径
line 115: fmt.Printf("🚀 Starting server: %s\n", addr)
  ← 在此处插入 JSON 输出（line 115 之后、line 117 之前）
line 117: if s.browser { Open(addr) }  // 打开浏览器
line 124: return http.Serve(listener, handler)  // 阻塞，永不返回
```

**关键**: `http.Serve()` 是阻塞调用，JSON 必须在它之前输出。

### 完整修改

#### 1. `cmd/root.go` — 新增 flag + 传递到 ServerOptions

```go
// init() 中新增：
rootCmd.Flags().Bool("json", false, "Output server info as JSON to stdout on startup")

// RunE 中新增：
jsonOutput, _ := cmd.Flags().GetBool("json")
server := internal.NewServerWithOptions(internal.ServerOptions{
    // ... 现有字段 ...
    JSONOutput: jsonOutput,
})
```

#### 2. `internal/server.go` — ServerOptions + Server 结构体新增字段

```go
// ServerOptions 新增：
type ServerOptions struct {
    // ... 现有字段 ...
    JSONOutput bool
}

// Server 新增：
type Server struct {
    // ... 现有字段 ...
    jsonOutput bool
}

// NewServerWithOptions 新增赋值：
func NewServerWithOptions(opts ServerOptions) *Server {
    // ... 现有逻辑 ...
    return &Server{
        // ... 现有字段 ...
        jsonOutput: opts.JSONOutput,
    }
}
```

#### 3. `internal/server.go` — Serve() 方法中的 JSON 输出 + stderr 重定向

```go
// 在 line 115 (fmt.Printf("🚀 Starting server...")) 之后插入：

// 当 --json 模式时，将状态消息重定向到 stderr
if s.jsonOutput {
    // 注意：上面的 fmt.Printf 已经输出了，这里只处理后续的
    // 实际实现中，需要将 line 99/101/115 的 fmt.Printf 改为：
    // if s.jsonOutput { fmt.Fprintf(os.Stderr, ...) } else { fmt.Printf(...) }
}

// JSON 输出（在 Open(browser) 之前）
if s.jsonOutput {
    info := map[string]interface{}{
        "port": actualPort,  // JSON number, not string
        "host": s.host,
        "url":  fmt.Sprintf("http://%s:%d%s", s.host, actualPort, initialPath),
    }
    jsonBytes, _ := json.Marshal(info)
    fmt.Println(string(jsonBytes))
    os.Stdout.Sync()
}
```

#### 4. stderr 重定向（已有状态消息）

将 `Serve()` 中现有的 3 处 `fmt.Printf` 改为条件输出：

```go
// line 99: 替换
if s.jsonOutput {
    fmt.Fprintf(os.Stderr, "📡 Auto-reload enabled. Only .md files will trigger browser refresh.\n")
} else {
    fmt.Printf("📡 Auto-reload enabled. Only .md files will trigger browser refresh.\n")
}

// line 101: 同理
// line 115: 同理
```

### 修改范围

修改 2 个文件：
- `cmd/root.go` — 新增 flag + ServerOptions 传递（+5 行）
- `internal/server.go` — 结构体字段 + JSON 输出 + stderr 重定向（+20 行）

---

## 十、go-grip 二进制嵌入与子进程启动

### Xcode Build Phase (Run Script)

```bash
set -euo pipefail  # 任何错误立即终止构建

# 构建 universal binary (arm64 + x86_64)
cd "${SRCROOT}/.."
GOOS=darwin GOARCH=arm64 go build -o /tmp/go-grip-arm64 .
GOOS=darwin GOARCH=amd64 go build -o /tmp/go-grip-amd64 .
lipo -create /tmp/go-grip-arm64 /tmp/go-grip-amd64 \
     -output "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"
chmod +x "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"
```

> 前提：构建机器需安装 Go 工具链。CI 通过 `setup-go` 处理，本地需手动安装。

### Swift 启动子进程

```swift
// ProcessManager.swift 中的方法
func start(path: String) async {
    guard let url = Bundle.main.url(forResource: "go-grip", withExtension: nil) else {
        // 二进制不存在 → 在主线程显示错误
        return
    }
    try? FileManager.default.setAttributes(
        [.posixPermissions: 0o755], ofItemAtPath: url.path
    )

    let process = Process()
    process.executableURL = url
    process.arguments = ["--no-reload", "--json", "--browser=true", path]
    // --no-reload: 避免 fsnotify watcher 开销（App 管理生命周期）
    // --json: 启动后 stdout 仅输出一行 JSON（状态消息走 stderr）
    // --browser=true: 首次启动自动打开浏览器

    let pipe = Pipe()
    process.standardOutput = pipe
    process.standardError = FileHandle.nullDevice  // 静默 stderr

    // 设置 terminationHandler（必须在主线程更新 @Published）
    process.terminationHandler = { [weak self] _ in
        DispatchQueue.main.async {
            self?.instances.removeValue(forKey: path)
        }
    }

    try? process.run()

    // 从 stdout 逐行读取 JSON 获取实际端口
    let port = await readPortFromStdout(pipe: pipe)

    let instance = RunningInstance(path: path, process: process, port: port)
    instances[path] = instance
}

/// 从 go-grip stdout 逐行读取，找到 JSON 行并提取端口
private func readPortFromStdout(pipe: Pipe) async -> Int {
    return await withCheckedContinuation { continuation in
        var resumed = false
        let lock = NSLock()

        // 超时保护：5 秒后强制返回 fallback
        DispatchQueue.global().asyncAfter(deadline: .now() + 5) {
            lock.lock()
            if !resumed {
                resumed = true
                lock.unlock()
                continuation.resume(returning: 6419)
            } else {
                lock.unlock()
            }
        }

        DispatchQueue.global().async {
            let handle = pipe.fileHandleForReading
            var buffer = Data()

            while true {
                let chunk = handle.availableData
                if chunk.isEmpty { break } // EOF
                buffer.append(chunk)

                // 逐行扫描，找到 JSON 对象
                while let newlineRange = buffer.range(of: Data("\n".utf8)) {
                    let lineData = buffer.subdata(
                        in: buffer.startIndex..<newlineRange.lowerBound
                    )
                    buffer.removeSubrange(buffer.startIndex...newlineRange.lowerBound)

                    if let json = try? JSONSerialization.jsonObject(with: lineData) as? [String: Any],
                       let port = json["port"] as? Int {
                        lock.lock()
                        if !resumed {
                            resumed = true
                            lock.unlock()
                            continuation.resume(returning: port)
                            return
                        } else {
                            lock.unlock()
                            return
                        }
                    }
                }
            }

            // EOF 未找到 JSON → fallback
            lock.lock()
            if !resumed {
                resumed = true
                lock.unlock()
                continuation.resume(returning: 6419)
            } else {
                lock.unlock()
            }
        }
    }
}
```

### Info.plist 关键配置

```xml
<key>LSUIElement</key>
<true/>  <!-- 无 Dock 图标，无 App Switcher 条目 -->
```

### 错误处理

| 场景 | 处理方式 |
|------|----------|
| go-grip 二进制不存在 | popover 中显示错误提示："未找到 go-grip 二进制文件" |
| go-grip 启动失败（port 冲突等） | 弹出 Alert 提示用户 |
| go-grip 进程意外退出 | 从 instances 移除，badge -1，可选 toast 提示 |
| 拖拽的不是目录/.md 文件 | 拖拽蒙版显示"不支持的文件类型" |

---

## 十一、历史记录持久化

### 存储方式

UserDefaults + JSON 编码

```swift
// Storage.swift
struct Storage {
    private let key = "go-grip-history"
    private let maxEntries = 50

    func load() -> [HistoryEntry] {
        guard let data = UserDefaults.standard.data(forKey: key),
              let entries = try? JSONDecoder().decode([HistoryEntry].self, from: data) else {
            return []
        }
        return entries
    }

    func save(_ entries: [HistoryEntry]) {
        guard let data = try? JSONEncoder().encode(entries) else { return }
        UserDefaults.standard.set(data, forKey: key)
    }

    func addToHistory(path: String) {
        var entries = load()
        let isDir = (try? FileManager.default.attributesOfItem(atPath: path))?[.type] as? FileAttributeType == .typeDirectory
        let displayName = URL(fileURLWithPath: path).lastPathComponent

        // 去重：已存在则移到顶部 + accessCount++
        if let index = entries.firstIndex(where: { $0.path == path }) {
            var existing = entries.remove(at: index)
            existing.lastOpened = Date()
            existing.accessCount += 1
            entries.insert(existing, at: 0)
        } else {
            let entry = HistoryEntry(
                id: UUID(), path: path, displayName: displayName,
                isDirectory: isDir, lastOpened: Date(), accessCount: 1
            )
            entries.insert(entry, at: 0)
        }

        // 容量限制：超出时移除最久未打开的
        if entries.count > maxEntries {
            entries = Array(entries.prefix(maxEntries))
        }

        save(entries)
    }

    func removeFromHistory(id: UUID) {
        var entries = load()
        entries.removeAll { $0.id == id }
        save(entries)
    }

    func clearHistory() {
        UserDefaults.standard.removeObject(forKey: key)
    }
}
```

### 排序规则

1. 按 `lastOpened` 降序（最近打开的在最上面）
2. 同时间按 `accessCount` 降序（打开次数多的优先）

### 去重行为

同一路径再次打开时：
- 移到列表顶部（更新 `lastOpened`）
- `accessCount += 1`
- 不创建新条目

### 容量限制

最多 50 条。超出时移除最久未打开的条目。

---

## 十二、进程管理

### ProcessManager（完整实现）

```swift
// ProcessManager.swift — 放在 Services/ 目录（ObservableObject，非纯 Model）
class ProcessManager: ObservableObject {
    @Published var instances: [String: RunningInstance] = [:]  // path → instance

    var count: Int { instances.count }

    func start(path: String) async {
        // 见第十节的完整实现（含 readPortFromStdout）
    }

    func stop(path: String) {
        guard let instance = instances[path] else { return }
        instance.process.terminate()  // SIGTERM
        // terminationHandler 会在回调中从 instances 移除
    }

    func stopAll() {
        for (_, instance) in instances {
            instance.process.terminate()
        }
        // 等待最多 2 秒，超时则强制 kill
        DispatchQueue.global().asyncAfter(deadline: .now() + 2) { [weak self] in
            guard let self = self else { return }
            DispatchQueue.main.async {
                for (_, instance) in self.instances {
                    if instance.process.isRunning {
                        instance.process.interrupt()  // SIGINT (kill -9 等效)
                    }
                }
                self.instances.removeAll()
            }
        }
    }

    func isRunning(path: String) -> Bool {
        instances[path]?.process.isRunning ?? false
    }

    func port(for path: String) -> Int? {
        instances[path]?.port
    }
}
```

### 生命周期管理

- **启动**: `Process()` + `process.run()` + `terminationHandler` 回调
- **terminationHandler**: 必须在 `DispatchQueue.main.async` 中更新 `@Published`，否则 SwiftUI 会 crash
- **终止**: `process.terminate()` (SIGTERM)，2 秒超时后 `interrupt()` (SIGINT)
- **App 退出**: `NSApplicationDelegate.applicationWillTerminate` → `stopAll()`

### Badge 更新

NSStatusItem 没有原生 badge API。通过在图标上叠加绘制数字实现：

```swift
func updateBadge(count: Int) {
    guard let button = statusItem.button else { return }

    let image = NSImage(systemSymbolName: "doc.text", accessibilityDescription: "GoGrip")!

    if count > 0 {
        // 在图标右下角绘制红色圆形 + 数字
        let badgeImage = NSImage(size: NSSize(width: 18, height: 18), flipped: false) { rect in
            let circleRect = NSRect(x: 8, y: 8, width: 18, height: 18)
            NSColor.red.setFill()
            NSBezierPath(ovalIn: circleRect).fill()

            let text = "\(count)" as NSString
            let attrs: [NSAttributedString.Key: Any] = [
                .font: NSFont.systemFont(ofSize: 11, weight: .bold),
                .foregroundColor: NSColor.white,
            ]
            let textSize = text.size(withAttributes: attrs)
            let textRect = NSRect(
                x: circleRect.midX - textSize.width / 2,
                y: circleRect.midY - textSize.height / 2,
                width: textSize.width,
                height: textSize.height
            )
            text.draw(in: textRect, withAttributes: attrs)
            return true
        }
        image.lockFocus()
        badgeImage.draw(at: .zero, from: .zero, operation: .sourceOver, fraction: 1.0)
        image.unlockFocus()
    }

    image.isTemplate = true  // 支持深色/浅色模式自动适配
    button.image = image
}
```

---

## 十三、构建与分发

### Xcode 项目配置

创建 `macos/GoGrip.xcodeproj` 时的关键设置：

| 设置 | 值 |
|------|-----|
| `MACOSX_DEPLOYMENT_TARGET` | `13.0` |
| `PRODUCT_BUNDLE_IDENTIFIER` | `com.showgp.GoGrip` |
| `SWIFT_VERSION` | `5.0` |
| `ENABLE_HARDENED_RUNTIME` | `NO`（未签名构建） |
| `CODE_SIGN_IDENTITY` | `""`（空，不签名） |
| `INFOPLIST_FILE` | `GoGrip/Info.plist` |

### Info.plist 完整内容

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>LSUIElement</key>
    <true/>  <!-- 无 Dock 图标，无 App Switcher 条目 -->
    <key>CFBundleName</key>
    <string>GoGrip</string>
    <key>CFBundleDisplayName</key>
    <string>GoGrip</string>
    <key>CFBundleIdentifier</key>
    <string>com.showgp.GoGrip</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleExecutable</key>
    <string>GoGrip</string>
    <key>LSMinimumSystemVersion</key>
    <string>13.0</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
```

### GoGrip.entitlements

未签名构建不需要 entitlements 文件。如果未来需要签名/公证：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>com.apple.security.app-sandbox</key>
    <false/>  <!-- 沙盒关闭，允许启动子进程和访问文件系统 -->
</dict>
</plist>
```

### 本地构建

```bash
# 方式 1: Xcode
open macos/GoGrip.xcodeproj
# Cmd+B 构建

# 方式 2: 命令行
xcodebuild -project macos/GoGrip.xcodeproj \
  -scheme GoGrip -configuration Release \
  CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO \
  build
```

### GitHub Actions CI/CD

在现有 `release.yml` 中增加 macOS App job（与现有 `release` job 并行）：

```yaml
# 注意：需同步更新现有 job 的 go-version 从 "1.25.x" 到 "1.26"
# （go.mod 已声明 go 1.26）

build-macos-app:
  runs-on: macos-latest
  permissions:
    contents: write  # 需要写权限上传 Release assets
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.26'

    # 构建 Swift App（内含 go-grip universal binary）
    - name: Build GoGrip.app
      run: |
        xcodebuild -project macos/GoGrip.xcodeproj \
          -scheme GoGrip -configuration Release \
          CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO \
          archive -archivePath build/GoGrip.xcarchive

    # 创建 DMG（使用系统自带 hdiutil，无需第三方工具）
    - name: Create DMG
      run: |
        mkdir -p dmg-root
        cp -R build/GoGrip.xcarchive/Products/Applications/GoGrip.app dmg-root/
        ln -s /Applications dmg-root/Applications
        hdiutil create -volname "GoGrip" -srcfolder dmg-root \
          -ov -format UDZO GoGrip.dmg

    # 上传到 GitHub Release
    - name: Upload DMG
      uses: softprops/action-gh-release@v2
      with:
        files: GoGrip.dmg
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Gatekeeper 注意事项

- 未签名 App 用户首次打开需右键 → 打开
- 有 Apple Developer 账号可做公证（notarize），消除警告
- 公证需要: Developer ID Application 证书 + Apple ID + app-specific password

---

## 十四、功能范围

### 第一版 (V1)

| 功能 | 包含 | 说明 |
|------|------|------|
| 状态栏图标 | ✅ | SF Symbol `doc.text`，右下角运行实例数 badge（attributedTitle 自绘） |
| NSPopover 面板 | ✅ | 点击图标弹出，点击外部自动关闭 |
| 拖拽打开 | ✅ | 拖拽目录和 .md 文件到面板，拖拽时显示蒙版，拖拽期间阻止 dismiss |
| 按钮打开 | ✅ | 单个"打开"按钮，弹出 NSOpenPanel 选择文件/目录 |
| 历史记录 | ✅ | 最近 50 条，UserDefaults 持久化，按最近打开排序 |
| 多实例管理 | ✅ | 每个目录独立进程，● 标记运行中，点击可关闭 |
| 端口感知 | ✅ | go-grip 新增 `--json` 输出模式，Swift 解析实际端口 |
| 退出清理 | ✅ | App 退出时 kill 所有子进程 |
| 开机自启 | ✅ | Login Items (SMAppService)，仅 macOS 13+ |
| 错误处理 | ✅ | 二进制不存在/启动失败时弹出错误提示 |

### 开机自启实现（SMAppService）

```swift
import ServiceManagement

// AppDelegate.swift 或 SettingsView.swift
func toggleLoginItem(enabled: Bool) {
    guard let service = SMAppService.mainApp else { return }
    do {
        if enabled {
            try service.register()
        } else {
            try service.unregister()
        }
    } catch {
        print("Login item error: \(error)")
    }
}

func isLoginItemEnabled() -> Bool {
    SMAppService.mainApp?.status == .enabled
}
```

Popover 中的 ⚙️ 按钮可弹出设置面板，包含"开机自启"开关。

### 第二版 (V2)

| 功能 | 说明 |
|------|------|
| 右键菜单 | 右键状态栏图标显示菜单 (设置、关于、退出) |
| 全局快捷键 | ⌘⇧G 呼出 popover |
| 文件变更检测 | 监听子进程 stdout，检测文件保存事件 |

---

## 十五、已确认的设计决策

| 决策项 | 结论 | 依据 |
|--------|------|------|
| UI 框架 | Swift/AppKit NSStatusItem + NSPopover | 生产级方案，完全可控 |
| 端口感知 | go-grip 新增 `--json` 输出模式 | 最准确，改动量小（+12 行 Go 代码） |
| 最低 macOS 版本 | macOS 13+ | SMAppService 开机自启需要 |
| 状态栏图标 | SF Symbol `doc.text` + attributedTitle 数字 badge | 无需自定义图标资源 |
| 拖拽时 popover 行为 | 阻止 dismiss，释放后恢复 | 防止拖拽过程中 popover 意外关闭 |
| 历史记录去重 | 同一路径再次打开 → 移到顶部 + accessCount++ | 保持最近使用排序 |
| NSOpenPanel 多选 | 不支持多选（allowsMultipleSelection = false） | V1 简化交互 |
| Popover 宽度 | 固定 320pt | V1 简化，V2 可加可调 |
| go-grip 改动范围 | cmd/root.go + internal/server.go（+25 行，含 stderr 重定向） | 最小侵入性 |

---

## 十六、测试计划

| 测试类型 | 覆盖内容 |
|----------|----------|
| **Unit — Storage** | load/save/addToHistory（去重+排序+容量限制）/removeFromHistory/clearHistory |
| **Unit — ProcessManager** | start/stop/stopAll/isRunning/port，terminationHandler 回调清理 |
| **Unit — readPortFromStdout** | 正常 JSON 行、混合 stdout 内容、EOF 无 JSON、超时 fallback |
| **Integration** | Swift 启动 go-grip 子进程 → 读取端口 → HTTP 请求验证服务可用 |
| **UI — 拖拽** | 拖入显示蒙版、释放执行打开、拖离蒙版消失、非 .md 文件拒绝 |
| **UI — 历史** | 添加/删除/清空/去重/排序 |
| **Edge case** | 路径含空格/Unicode/符号链接、端口全占满（100 个实例）、App 退出时子进程清理 |

---

## 十七、风险与注意事项

1. **Gatekeeper**: 未签名 App 用户首次打开需右键 → 打开。有 Apple Developer 账号可做公证（notarize）消除警告
2. **Xcode 项目**: 需用 Xcode GUI 创建 `.xcodeproj`，按第十三节配置 deployment target 等关键设置
3. **macOS 版本**: 最低支持 macOS 13（SMAppService 开机自启需要）。NSPopover 支持 10.7+，无兼容性问题
4. **子进程残留**: App 崩溃时子进程可能残留。V2 可加 PID 文件清理机制
5. **内存占用**: 每个 go-grip 运行一个 HTTP server + fsnotify watcher，约占 20-30MB
6. **go-grip `--json` stdout 协议**: 启用 `--json` 时，状态消息走 stderr，stdout 仅输出一行 JSON。Swift 逐行读取直到找到 JSON 对象
7. **拖拽 dismiss 冲突**: 需在拖拽期间将 NSPopover behavior 从 `.transient` 切换为 `.applicationDefined`，释放后恢复
8. **`--browser=true` 语义**: go-grip 首次启动时自动打开浏览器（`--browser=true`），Swift 的 `open http://localhost:{port}` 仅用于激活已有 tab（点击已运行条目时）
9. **路径规范化**: 拖拽接收的路径需做 `URL.path` 标准化（解析符号链接、处理空格和 Unicode），再检查扩展名和 isDirectory
10. **CI Go 版本同步**: 现有 `release.yml` 的 `go-version: "1.25.x"` 需同步更新为 `"1.26"`（`go.mod` 已声明 1.26）
