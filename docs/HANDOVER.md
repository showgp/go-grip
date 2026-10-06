# GoGrip macOS App — 交接文档

> 状态（2026-10-05）：本文描述的旧示例 App 已被替换。旧运行时链路（`ProcessManager`、拖拽入口、Finder Sync 扩展、`project.yml`）已从工程移除；新宿主使用独立版本化最近键（至多 20 个规范化身份），旧 50 项历史源与测试已随任务 09 移除，用户旧数据保留在 `go-grip-history` 但新宿主不读、不写、不迁移、不删除。宿主正按 `openspec/changes/rebuild-macos-preview-app` 重建。当前开发 App 的菜单栏面板支持多选批次手动入口（内置 `go-grip` → 默认浏览器；2026-10-05 由任务 2.3 接通）、最近目标与完整状态面板（任务 4.1–4.3），Finder Services 入口（服务冷启动、跨入口复用，任务 3.1/3.2/3.4 的 07 贡献）已于 2026-10-05 接通并经真实 Finder smoke 验证（登录项与候选包的旧状态：登录项已于 2026-10-06 由任务 5.3 接通，见本段末尾；候选包尚未接通；2026-10-06 更新：候选包已由任务 12 交付，见下方候选打包说明），首次引导与完整简中/英文本地化已由任务 5.1/5.2 于 2026-10-06 接通（权限 gate 已由 08 在本机开发构建验证：实际授权主体、拒绝指导与外接/受保护访问；网络卷访问仍缺环境，任务 3.3 未完整关闭），详见 `docs/ARCHITECTURE.md` 第七节与 `README.md`。（2026-10-06 更新：07 票重开以修正共享提示呈现——应用级提示改为非模态单槽，失败报告与数量确认共用一个提示且均可关闭，提示显示期间到达的 Finder Services 请求不再丢失；开发构建 `eead0216…` 已安装到 `/Applications`，09 构建 `9c4ddf9b…` 备份于 `~/Desktop/GoGrip-backups/GoGrip-09-backup.app`；两项 native 复核（提示关闭后面板报告列表、另一前台 App 时提示是否前置可见）已于当日下午补测通过，并按首次补测发现的渲染问题在展示前补 `alert.layout()`；最终构建 CDHash `7a2c78fa…` 已安装到 `/Applications`，09 构建 `9c4ddf9b…` 备份保留在 `~/Desktop/GoGrip-backups/GoGrip-09-backup.app`。详见架构文档与 07 票；任务 5.1/5.2 于 2026-10-06 交付首次/可再次查看引导与完整简中/英文本地化（其他语言回退英文，含 ServicesMenu 名称），`/Applications` 现为本票最终开发构建 `3949aca2…`（07 构建 `7a2c78fa…` 备份于 `~/Desktop/GoGrip-backups/GoGrip-07-backup.app`），语言、服务启用开关、首次与最近键均已记录并还原、用户 `go-grip-history` 未动，详见架构文档第七节 10 记录。任务 5.3（2026-10-06）接通用户可选登录启动：`SMAppService.mainApp` 真实状态、仅用户动作 register/unregister、注册失败/需批准不假成功、失败沿用根级一次报告，未注册默认关闭且不自动注册/注销；登录项关闭且宿主未运行时的真实 Finder Services 冷启动已复核；`/Applications` 现为本票最终开发构建 `d22aee2e…`（10 构建 `3949aca2…` 备份于 `~/Desktop/GoGrip-backups/GoGrip-10-backup.app`），登录项已恢复未注册，详见架构文档第七节 5.3 记录。）

> 候选打包（2026-10-06，任务 12）：`make macos-candidate` 从显式 archive（`macos/.build/GoGrip.xcarchive`）生成自包含 universal 候选 App/DMG（`macos/.build/candidate/`），宿主与包内 Go 均为 arm64/x86_64 并按各自 identifier ad-hoc 签名；macOS CI 配置为把 App zip/DMG 作为候选 workflow artifact 上传（本轮未运行远端，仅本地等价验证，见任务 12 关闭记录），tag 流水线不再向正式 Release 上传 App/DMG（独立 CLI 资产流程不变）。本轮 smoke 已把候选装到 `/Applications/GoGrip.app`（替换前的 11 开发构建备份于 `~/Desktop/GoGrip-backups/GoGrip-11-backup.app`）；Finder Services 冷启动需在系统设置重新启用该服务的启动项（ad-hoc 重建改变代码身份，旧启用状态不继承）。正式 Developer ID 签名、公证与公开发布仍未开始；本文其余章节为历史记录，当前能力以 `openspec/changes/rebuild-macos-preview-app`、`docs/ARCHITECTURE.md` 第七/十六节与 `README.md` 为准。

> 创建日期: 2026-06-03
> 用途: 交给其他 Agent 继续实施

---

## 一、项目目标

在 macOS 状态栏（菜单栏）放置一个轻量 App，嵌入 go-grip 二进制，用户通过拖拽目录/.md 文件快速打开预览，并管理历史打开记录。

**核心价值**: 无需终端操作，拖拽即用；记住历史，一键重返；多实例并行管理。

---

## 二、设计文档

完整设计文档（包含所有技术细节、代码示例、架构图）：

- **主方案**: `docs/macos-app-plan.md`（1002 行，17 个章节）
- **实施方案**: `docs/macos-app-impl.md`（412 行，11 个 Phase，76 个任务）

**⚠️ 实施前必须阅读这两个文档**，本交接文档仅覆盖当前进度和关键注意事项。

---

## 三、当前进度

### ✅ 已完成

#### Phase 1: Go --json 基础设施（全部完成并验证）

| 文件 | 状态 | 说明 |
|------|------|------|
| `cmd/root.go` | ✅ 已修改 | 新增 `--json` flag，传递 `JSONOutput` 到 ServerOptions |
| `internal/server.go` | ✅ 已修改 | 新增 `jsonOutput` 字段 + JSON 输出 + 4 处 stderr 重定向 + `Stop()` 方法 |
| `internal/server_test.go` | ✅ 已新增 | `TestServeJSONOutput` 测试 |
| `Makefile` | ✅ 已新增 | build/test/run targets |

**验证结果**:
- `go build .` ✅
- `go test ./...` ✅
- `go run . --json --no-reload README.md` → stdout 仅输出 `{"host":"localhost","port":6421,...}` ✅
- 状态消息输出到 stderr ✅

#### Phase 2: Xcode 项目（部分完成，缺少 project.pbxproj）

| 文件 | 状态 | 说明 |
|------|------|------|
| `macos/GoGrip/GoGripApp.swift` | ✅ 已创建 | @main 入口，@NSApplicationDelegateAdaptor |
| `macos/GoGrip/AppDelegate.swift` | ✅ 已创建 | NSStatusItem + NSPopover + togglePopover |
| `macos/GoGrip/Info.plist` | ✅ 已创建 | LSUIElement=true，macOS 13+ |
| `macos/Scripts/build-go-grip.sh` | ✅ 已创建 | universal binary 构建脚本（已 chmod +x） |
| `macos/GoGrip.xcodeproj/project.pbxproj` | ❌ **缺失** | 需要创建（见下方详细说明） |

### 🔄 未完成

- Phase 2 的 `project.pbxproj` 生成
- Phase 3–11 全部

---

## 四、Phase 2 剩余任务：生成 project.pbxproj

这是当前阻塞项。`project.pbxproj` 是 Xcode 项目的核心配置文件。

### 需要包含的内容

**文件引用（PBXFileReference）**:
1. `GoGrip/GoGripApp.swift`
2. `GoGrip/AppDelegate.swift`
3. `GoGrip/Info.plist`
4. `Scripts/build-go-grip.sh`
5. `GoGrip.app`（产品引用）

**构建阶段（PBXBuildFile）**:
1. GoGripApp.swift — Sources
2. AppDelegate.swift — Sources
3. build-go-grip.sh — PBXShellScriptBuildPhase

**Shell Script Build Phase**:
```bash
set -euo pipefail
cd "${SRCROOT}/.."
GOOS=darwin GOARCH=arm64 go build -o /tmp/go-grip-arm64 .
GOOS=darwin GOARCH=amd64 go build -o /tmp/go-grip-amd64 .
lipo -create /tmp/go-grip-arm64 /tmp/go-grip-amd64 \
     -output "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"
chmod +x "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"
```

### 关键构建设置（XCBuildConfiguration）

```
MACOSX_DEPLOYMENT_TARGET = 13.0
PRODUCT_BUNDLE_IDENTIFIER = com.showgp.GoGrip
SWIFT_VERSION = 5.0
ENABLE_HARDENED_RUNTIME = NO
CODE_SIGN_IDENTITY = ""
CODE_SIGNING_ALLOWED = NO
INFOPLIST_FILE = GoGrip/Info.plist
LD_RUNPATH_SEARCH_PATHS = @executable_path/../Frameworks
GENERATE_INFOPLIST_FILE = NO
```

### 两种实现方式（选其一）

**方式 A（推荐）: 使用 xcodegen**

1. 安装: `brew install xcodegen`
2. 在 `macos/` 目录创建 `project.yml`:
```yaml
name: GoGrip
options:
  bundleIdPrefix: com.showgp
  deploymentTarget:
    macOS: "13.0"
  xcodeVersion: "16.0"
targets:
  GoGrip:
    type: application
    platform: macOS
    sources:
      - GoGrip
    info:
      path: GoGrip/Info.plist
    scripts:
      - name: Build go-grip
        script: |
          set -euo pipefail
          cd "${SRCROOT}/.."
          GOOS=darwin GOARCH=arm64 go build -o /tmp/go-grip-arm64 .
          GOOS=darwin GOARCH=amd64 go build -o /tmp/go-grip-amd64 .
          lipo -create /tmp/go-grip-arm64 /tmp/go-grip-amd64 \
               -output "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"
          chmod +x "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"
    settings:
      base:
        CODE_SIGN_IDENTITY: ""
        CODE_SIGNING_ALLOWED: NO
        ENABLE_HARDENED_RUNTIME: NO
```
3. 运行: `cd macos && xcodegen generate`
4. 验证: `xcodebuild -project GoGrip.xcodeproj -scheme GoGrip -configuration Debug CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO build`

**方式 B: 手写 pbxproj**

参考现有 Xcode 项目的 pbxproj 格式，需要生成：
- 24 字符十六进制 UUID（每个对象唯一）
- PBXBuildFile / PBXFileReference / PBXGroup / PBXNativeTarget / PBXProject / PBXShellScriptBuildPhase / XCBuildConfiguration / XCConfigurationList

### 验证命令

```bash
cd /Volumes/GW2T_PICE3/workspace/05-systems-rust-macos/go-grip
xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
  -configuration Debug \
  CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO build 2>&1 | tail -20
```

成功标志: `BUILD SUCCEEDED`

---

## 五、Phase 3–11 实施指南

完整任务列表见 `docs/macos-app-impl.md`。以下是关键实现要点：

### Phase 3: 状态栏 + Popover

- 在 `AppDelegate.swift` 中已创建 `statusItem` 和 `popover`
- 需要创建 `PopoverView.swift`（SwiftUI 视图）并设置为 popover 的 contentViewController
- 使用 `NSHostingController(rootView: PopoverView(...))` 桥接 SwiftUI 到 AppKit
- 创建 `DropOverlayView.swift`（拖拽蒙版 UI）

### Phase 4: 进程管理

完整 Swift 代码在 `docs/macos-app-plan.md` 第十节（约 line 480-580）：
- `ProcessManager.swift` — ObservableObject，管理子进程
- `readPortFromStdout` — 逐行扫描 + NSLock + 5 秒超时
- `RunningInstance` — 路径 + Process + 端口

### Phase 5: 拖拽

- `.onDrop(of: [.fileURL], isTargeted: $isDragOver)` 修饰符
- NSItemProvider URL 提取代码在 `docs/macos-app-plan.md` 约 line 215-233
- 拖拽期间 `popover.behavior = .applicationDefined`，释放后恢复 `.transient`

### Phase 6: 历史记录

- `Storage.swift` — UserDefaults + JSON 编码，代码在 `docs/macos-app-plan.md` 约 line 608-662
- `HistoryEntry.swift` — Codable 数据模型，line 319-326
- 最多 50 条，同一路径重复打开移到顶部 + accessCount++

### Phase 7: Badge

- NSImage 叠加绘制代码在 `docs/macos-app-plan.md` 约 line 744-778
- 监听 `ProcessManager.instances` 变化 → 调用 `updateBadge(count:)`

### Phase 8: 开机自启

- SMAppService 代码在 `docs/macos-app-plan.md` 约 line 927-947
- 创建 `SettingsView.swift`（设置面板 + ⚙️ 按钮）

### Phase 9: 错误处理

- 二进制不存在 → popover 错误提示
- 启动失败 → Alert
- 进程 crash → terminationHandler 自动清理

### Phase 10: CI/CD

- 在 `.github/workflows/release.yml` 新增 `build-macos-app` job
- DMG 创建使用 `hdiutil`
- 同步更新 go-version 为 1.26

### Phase 11: 测试

- Storage 单元测试
- ProcessManager 单元测试
- readPortFromStdout 单元测试（正常/混合/EOF/超时）
- 集成测试

---

## 六、文件结构

### 现有文件（已修改）

```
cmd/root.go                    # +3 行（--json flag）
internal/server.go             # +49 行（JSON 输出 + stderr 重定向）
internal/server_test.go        # +84 行（新增测试）
Makefile                       # 新建（构建入口）
```

### 新建文件（Phase 2 已创建）

```
macos/GoGrip/GoGripApp.swift      # @main 入口
macos/GoGrip/AppDelegate.swift     # NSStatusItem + NSPopover
macos/GoGrip/Info.plist            # LSUIElement=true
macos/Scripts/build-go-grip.sh     # universal binary 构建脚本
```

### 待创建文件

```
macos/GoGrip.xcodeproj/project.pbxproj   # ← 当前阻塞项
macos/GoGrip/Views/PopoverView.swift
macos/GoGrip/Views/DropOverlayView.swift
macos/GoGrip/Views/HistoryListView.swift
macos/GoGrip/Views/HistoryRowView.swift
macos/GoGrip/Views/SettingsView.swift
macos/GoGrip/Models/HistoryEntry.swift
macos/GoGrip/Services/ProcessManager.swift
macos/GoGrip/Utilities/Storage.swift
macos/GoGripTests/StorageTests.swift
macos/GoGripTests/ProcessManagerTests.swift
macos/GoGripTests/PortReaderTests.swift
macos/GoGripTests/IntegrationTests.swift
```

---

## 七、关键注意事项

### 1. Go `--json` 协议

Swift 从 go-grip stdout 读取的 JSON 格式：
```json
{"port": 6419, "host": "localhost", "url": "http://localhost:6419/README.md"}
```
- **端口值必须是 JSON number**（不是字符串）
- 状态消息走 stderr（已实现）
- Swift 用 `json["port"] as? Int` 读取

### 2. 线程安全

- `ProcessManager` 是 `ObservableObject`，所有 `@Published` 更新必须在 `DispatchQueue.main.async` 中
- `terminationHandler` 在后台线程触发，必须 dispatch 到主线程
- `readPortFromStdout` 使用 `NSLock` 保护 continuation 状态

### 3. Popover 行为

- 拖拽期间设为 `.applicationDefined` 防止 dismiss
- 释放后恢复 `.transient`
- 点击外部关闭由 `NSEvent.addGlobalMonitorForEvents` 处理

### 4. go-grip 调用参数

```swift
process.arguments = ["--no-reload", "--json", "--browser=true", path]
```
- `--no-reload`: App 管理生命周期，不需要 fsnotify
- `--json`: 输出端口信息
- `--browser=true`: 首次启动自动打开浏览器

### 5. 构建脚本路径

`build-go-grip.sh` 中 `${SRCROOT}/..` 是 go-grip 项目根目录（macos/ 的上级）。

### 6. 环境要求

- Go 1.26+（`go.mod` 已声明）
- Xcode 16+（macOS SDK）
- macOS 13+ 部署目标

---

## 八、实施顺序建议

```
Phase 2 完成 (project.pbxproj)
    ↓
Phase 3 (状态栏 UI) + Phase 4 (进程管理) ← 可并行
    ↓
Phase 5 (拖拽)
    ↓
Phase 6 (历史) + Phase 7 (Badge) + Phase 8 (自启) + Phase 9 (错误处理) ← 可并行
    ↓
Phase 10 (CI/CD)
    ↓
Phase 11 (测试)
```

---

## 九、验证清单

每个 Phase 完成后：

1. `go build .` — Go 侧编译通过
2. `go test ./...` — Go 测试通过
3. `xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip -configuration Debug CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO build` — Swift 编译通过
4. 手动运行验证功能

全部完成后：

5. 拖拽目录 → 浏览器打开预览
6. 拖拽 .md 文件 → 正确打开
7. 历史列表显示 → 重启后保留
8. Badge 实时更新
9. 开机自启开关正常
10. 退出 App → 子进程清理
