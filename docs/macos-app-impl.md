# GoGrip macOS App 实施方案

> 创建日期: 2026-06-03
> 基于: [macos-app-plan.md](macos-app-plan.md)
> 状态: 待实施

---

## 实施总览

```
Phase 1: Go 基础设施 ──→ Phase 4: 进程管理 ──┐
                                              │
Phase 2: Xcode 项目 ──→ Phase 3: 状态栏 ──────┤
                                              │
                                    Phase 5: 拖拽 + 文件打开
                                              │
                              ┌────────────────┼────────────────┐
                              │                │                │
                        Phase 6: 历史    Phase 7: Badge   Phase 8: 开机自启
                              │                │                │
                              └────────────────┼────────────────┘
                                               │
                                       Phase 9: 错误处理
                                               │
                              ┌────────────────┴────────────────┐
                              │                                 │
                        Phase 10: CI/CD              Phase 11: 测试
```

预计总工时: 3-5 天（含调试）

---

## Phase 1: Go 基础设施

> 目标: go-grip 支持 `--json` 输出模式，为 Swift App 提供端口感知能力

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 1.1 | 新增 `--json` flag 到 `init()` | `cmd/root.go` | 5 min |
| 1.2 | `ServerOptions` 新增 `JSONOutput bool` | `internal/server.go` | 5 min |
| 1.3 | `Server` struct 新增 `jsonOutput bool` | `internal/server.go` | 5 min |
| 1.4 | `NewServerWithOptions` 传递 `jsonOutput` | `internal/server.go` | 5 min |
| 1.5 | `RunE` 中读取 `--json` flag 并传入 ServerOptions | `cmd/root.go` | 5 min |
| 1.6 | `Serve()` 中 line 115 后插入 JSON 输出逻辑 | `internal/server.go` | 15 min |
| 1.7 | 4 处 `fmt.Print*` 改为条件 stderr 重定向（line 99, 101, 115, 120） | `internal/server.go` | 15 min |
| 1.8 | 手动测试: `go run . --json README.md` 验证 stdout 仅输出 JSON | — | 10 min |
| 1.9 | 手动测试: 端口被占用时验证 JSON 输出正确端口 | — | 10 min |
| 1.10 | 单元测试: 验证 `--json` 模式下 stdout 输出 JSON、stderr 输出状态消息 | `internal/server_test.go` | 15 min |

### 验收标准

- [ ] `go run . --json README.md` → stdout 仅输出 `{"port":6419,...}` 一行 JSON
- [ ] `go run . --json README.md` → 状态消息输出到 stderr
- [ ] 端口 6419 被占用时 → JSON 输出 `{"port":6420,...}`
- [ ] 不传 `--json` 时 → 行为与现有完全一致
- [ ] `go test ./...` 全部通过（含新增 `--json` 模式单元测试）

---

## Phase 2: Xcode 项目 + App 生命周期

> 目标: 创建可编译运行的 macOS 空壳 App

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 2.1 | 用 Xcode 创建 macOS App 项目 (Swift, macOS 13+) | `macos/GoGrip.xcodeproj` | 15 min |
| 2.2 | 配置 Build Settings (deployment target, bundle ID, 禁用签名) | `.xcodeproj` | 10 min |
| 2.3 | 创建 Info.plist (LSUIElement=true 等) | `macos/GoGrip/Info.plist` | 10 min |
| 2.4 | 创建 GoGripApp.swift (@main App 入口) | `macos/GoGrip/GoGripApp.swift` | 10 min |
| 2.5 | 创建 AppDelegate.swift (NSApplicationDelegate) | `macos/GoGrip/AppDelegate.swift` | 15 min |
| 2.6 | Xcode Build Phase: 添加 go-grip 构建脚本 | `.xcodeproj` | 15 min |
| 2.7 | 验证: Cmd+B 编译通过，App 可运行（无 UI） | — | 10 min |

### 验收标准

- [ ] Xcode 项目可在 macOS 13+ 编译
- [ ] Build Phase 脚本能构建 go-grip universal binary 并放入 App Bundle
- [ ] App 启动后不出现在 Dock 和 App Switcher 中
- [ ] App 启动后在菜单栏显示（初始可为空白占位）

---

## Phase 3: 状态栏图标 + NSPopover

> 目标: 菜单栏显示图标，点击弹出 Popover 面板
> 前置依赖: Phase 2

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 3.1 | AppDelegate 中创建 NSStatusItem + SF Symbol 图标 | `AppDelegate.swift` | 15 min |
| 3.2 | 实现 togglePopover (点击切换 popover 显示/隐藏) | `AppDelegate.swift` | 10 min |
| 3.3 | 创建 PopoverView.swift (ZStack: 历史列表 + 拖拽蒙版) | `Views/PopoverView.swift` | 20 min |
| 3.4 | 创建 DropOverlayView.swift (拖拽蒙版 UI) | `Views/DropOverlayView.swift` | 15 min |
| 3.5 | Popover 尺寸配置 (320×480, minWidth) | `AppDelegate.swift` | 5 min |
| 3.6 | 验证: 点击图标弹出 popover，点击外部关闭 | — | 5 min |

### 验收标准

- [ ] 菜单栏显示 SF Symbol `doc.text` 图标
- [ ] 点击图标 → popover 弹出（带三角箭头）
- [ ] 再次点击图标 → popover 关闭
- [ ] 点击 popover 外部 → popover 自动关闭
- [ ] Popover 宽度 320pt，高度自适应

---

## Phase 4: 进程管理 + 端口检测

> 目标: 能启动/停止 go-grip 子进程，获取实际端口
> 前置依赖: Phase 1

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 4.1 | 创建 ProcessManager.swift (ObservableObject) | `Services/ProcessManager.swift` | 20 min |
| 4.2 | 实现 start(path:) — 启动子进程 + readPortFromStdout | `Services/ProcessManager.swift` | 30 min |
| 4.3 | 实现 readPortFromStdout — 逐行扫描 + 超时保护 | `Services/ProcessManager.swift` | 20 min |
| 4.4 | 实现 stop(path:) — terminate + terminationHandler | `Services/ProcessManager.swift` | 10 min |
| 4.5 | 实现 stopAll() — App 退出时清理 | `Services/ProcessManager.swift` | 10 min |
| 4.6 | terminationHandler 中 DispatchQueue.main.async 更新 @Published | `Services/ProcessManager.swift` | 5 min |
| 4.7 | 手动测试: 启动子进程 → 验证端口正确 → HTTP 请求可访问 | — | 15 min |
| 4.8 | 手动测试: 启动 2 个不同目录 → 验证端口不同 | — | 10 min |

### 验收标准

- [ ] `ProcessManager.start(path:)` 启动 go-grip 子进程
- [ ] `readPortFromStdout` 正确解析 JSON 获取实际端口
- [ ] 5 秒超时后 fallback 到 6419
- [ ] `stop(path:)` 终止子进程，instances 字典同步更新
- [ ] 子进程意外退出时 instances 自动清理（terminationHandler）
- [ ] 所有 @Published 更新在主线程

---

## Phase 5: 拖拽接收 + 文件打开

> 目标: 用户可通过拖拽或按钮打开目录/.md 文件
> 前置依赖: Phase 3 + Phase 4

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 5.1 | PopoverView 添加 .onDrop 修饰符 | `Views/PopoverView.swift` | 15 min |
| 5.2 | 实现 NSItemProvider URL 提取 + 路径规范化 | `Views/PopoverView.swift` | 15 min |
| 5.3 | 实现路径验证（目录 或 .md 文件） | `Views/PopoverView.swift` | 10 min |
| 5.4 | 拖拽进入/离开时切换 popover.behavior | `Views/PopoverView.swift` | 10 min |
| 5.5 | 创建"打开"按钮 + NSOpenPanel 集成 | `Views/PopoverView.swift` | 15 min |
| 5.6 | 拖拽/按钮打开 → ProcessManager.start + Storage.addToHistory | `Views/PopoverView.swift` | 10 min |
| 5.7 | 验证: 拖拽目录 → 蒙版出现 → 释放 → 浏览器打开 | — | 10 min |
| 5.8 | 验证: 拖拽 .md 文件 → 正确打开 | — | 5 min |
| 5.9 | 验证: 拖拽非 .md 文件 → 蒙版显示拒绝提示 | — | 5 min |
| 5.10 | 验证: 点击"打开"按钮 → NSOpenPanel → 选择文件/目录 | — | 5 min |

### 验收标准

- [ ] 拖拽文件到 popover → 显示半透明蒙版
- [ ] 释放 → 蒙版消失 → 浏览器打开对应页面
- [ ] 拖拽离开 → 蒙版消失，无操作
- [ ] 拖拽 .txt 文件 → 蒙版显示"不支持的文件类型"
- [ ] 点击"打开" → NSOpenPanel 可选目录和 .md 文件
- [ ] 路径含空格/Unicode 时正确处理

---

## Phase 6: 历史记录

> 目标: 持久化存储历史打开记录，展示在 popover 中
> 前置依赖: Phase 5（可与 Phase 7、8 并行）

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 6.1 | 创建 HistoryEntry.swift (Codable 数据模型) | `Models/HistoryEntry.swift` | 5 min |
| 6.2 | 创建 Storage.swift (UserDefaults 完整实现) | `Utilities/Storage.swift` | 20 min |
| 6.3 | 创建 HistoryListView.swift (ScrollView 列表) | `Views/HistoryListView.swift` | 15 min |
| 6.4 | 创建 HistoryRowView.swift (单行: 名称 + 状态指示器) | `Views/HistoryRowView.swift` | 15 min |
| 6.5 | 集成: PopoverView 绑定 Storage 数据 | `Views/PopoverView.swift` | 10 min |
| 6.6 | 实现点击历史条目 → 打开/激活 | `Views/HistoryRowView.swift` | 10 min |
| 6.7 | 实现删除单条历史 ( swipe 或按钮 ) | `Views/HistoryRowView.swift` | 10 min |
| 6.8 | 实现清空历史按钮 | `Views/PopoverView.swift` | 5 min |
| 6.9 | 验证: 打开文件 → 历史列表更新 → 重启 App 历史保留 | — | 10 min |
| 6.10 | 验证: 同一路径重复打开 → 移到顶部 + accessCount++ | — | 5 min |

### 验收标准

- [ ] 打开文件后历史列表立即显示新条目
- [ ] 重启 App 后历史记录保留
- [ ] 同一路径再次打开 → 条目移到顶部
- [ ] 最多显示 50 条，超出时自动移除最旧
- [ ] 点击历史条目 → 打开或激活对应实例
- [ ] 可删除单条或清空全部历史

---

## Phase 7: Badge 显示

> 目标: 状态栏图标右下角显示运行实例数
> 前置依赖: Phase 5（可与 Phase 6、8 并行）

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 7.1 | 实现 updateBadge(count:) — NSImage 叠加绘制 | `AppDelegate.swift` | 20 min |
| 7.2 | 监听 ProcessManager.instances 变化 → 调用 updateBadge | `AppDelegate.swift` | 10 min |
| 7.3 | count=0 时隐藏 badge，>0 时显示数字 | `AppDelegate.swift` | 5 min |
| 7.4 | 验证: 启动 3 个实例 → badge 显示 "3" | — | 5 min |
| 7.5 | 验证: 关闭 1 个 → badge 变为 "2" | — | 5 min |

### 验收标准

- [ ] 运行实例数 > 0 时，图标右下角显示红色圆形 + 白色数字
- [ ] 运行实例数 = 0 时，badge 隐藏
- [ ] 深色/浅色模式下 badge 均清晰可见
- [ ] 实例启动/停止时 badge 实时更新

---

## Phase 8: 开机自启

> 目标: 用户可设置开机自动启动
> 前置依赖: Phase 5（可与 Phase 6、7 并行）

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 8.1 | 实现 toggleLoginItem(enabled:) | `AppDelegate.swift` | 10 min |
| 8.2 | 实现 isLoginItemEnabled() | `AppDelegate.swift` | 5 min |
| 8.3 | 创建 SettingsView.swift (设置面板: 开机自启开关) | `Views/SettingsView.swift` | 20 min |
| 8.4 | Popover 中 ⚙️ 按钮 → 弹出 SettingsView | `Views/PopoverView.swift` | 10 min |
| 8.5 | 验证: 开启后重启 Mac → App 自动启动 | — | 5 min |

### 验收标准

- [ ] ⚙️ 按钮可访问设置面板
- [ ] 开机自启开关可切换
- [ ] 开启后重启 Mac → App 自动启动且菜单栏显示图标
- [ ] SMAppService.register() 失败时 → Alert 提示用户（如权限问题）

---

## Phase 9: 错误处理

> 目标: 覆盖所有异常场景的用户提示
> 前置依赖: Phase 5（可与 Phase 6、7、8 并行）

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 9.1 | go-grip 二进制不存在 → popover 显示错误提示 | `Views/PopoverView.swift` | 10 min |
| 9.2 | go-grip 启动失败 → Alert 提示用户 | `Views/PopoverView.swift` | 10 min |
| 9.3 | go-grip 进程意外退出 → 移除实例 + toast 提示 | `Services/ProcessManager.swift` | 10 min |
| 9.4 | 拖拽非支持文件 → 蒙版显示拒绝提示 | `Views/DropOverlayView.swift` | 5 min |
| 9.5 | 端口全占满（100 个）→ 启动失败处理 | `Services/ProcessManager.swift` | 5 min |
| 9.6 | 验证: 删除 go-grip 二进制 → 打开时显示错误 | — | 5 min |

### 验收标准

- [ ] 二进制不存在 → 用户看到明确错误提示
- [ ] 启动失败 → Alert 提示原因
- [ ] 进程 crash → 自动清理 + 用户可感知
- [ ] 不支持的文件类型 → 拖拽蒙版拒绝提示

---

## Phase 10: CI/CD

> 目标: GitHub Actions 自动构建 + DMG 打包
> 前置依赖: Phase 9 完成后

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 10.1 | `release.yml` 新增 `build-macos-app` job | `.github/workflows/release.yml` | 15 min |
| 10.2 | 同步更新所有 go-version 引用为 1.26 | `release.yml`, `build.yml`, `mise.toml` | 10 min |
| 10.3 | `build.yml` 新增 macOS CI job (PR 时仅编译 Swift，不重复 Go lint) | `.github/workflows/build.yml` | 15 min |
| 10.4 | 验证: 推送 tag → Release 包含 DMG | — | 15 min |

### 验收标准

- [ ] 推送 `v*` tag → GitHub Release 包含 `GoGrip.dmg`
- [ ] PR → CI 自动编译 Swift App（检查编译错误）
- [ ] DMG 包含 `GoGrip.app` + Applications 快捷方式（未签名，有 Gatekeeper 警告）

---

## Phase 11: 测试

> 目标: 核心功能测试覆盖（≥80% 覆盖率）
> 前置依赖: Phase 10 完成后

### 任务

| # | 任务 | 文件 | 预估 |
|---|------|------|------|
| 11.1 | Storage 单元测试 (load/save/add/remove/clear) | `GoGripTests/StorageTests.swift` | 20 min |
| 11.2 | ProcessManager 单元测试 (start/stop/stopAll/isRunning/port) | `GoGripTests/ProcessManagerTests.swift` | 20 min |
| 11.3 | readPortFromStdout 单元测试 (正常/混合/EOF/超时) | `GoGripTests/PortReaderTests.swift` | 20 min |
| 11.4 | 集成测试: 启动子进程 → 读端口 → HTTP 验证 | `GoGripTests/IntegrationTests.swift` | 20 min |
| 11.5 | 手动 UI 测试: 拖拽/打开/历史/badge 全流程 | — | 20 min |
| 11.6 | Edge case 测试: 路径含空格、Unicode、符号链接 | — | 10 min |

### 验收标准

- [ ] 单元测试全部通过
- [ ] 集成测试: go-grip 子进程可启动并响应 HTTP
- [ ] UI 全流程: 拖拽 → 打开 → 历史 → badge → 关闭 → 退出清理

---

## 进度追踪

| Phase | 名称 | 状态 | 开始日期 | 完成日期 |
|-------|------|------|----------|----------|
| 1 | Go 基础设施 | ⬜ 待开始 | — | — |
| 2 | Xcode 项目 + App 生命周期 | ⬜ 待开始 | — | — |
| 3 | 状态栏图标 + NSPopover | ⬜ 待开始 | — | — |
| 4 | 进程管理 + 端口检测 | ⬜ 待开始 | — | — |
| 5 | 拖拽接收 + 文件打开 | ⬜ 待开始 | — | — |
| 6 | 历史记录 | ⬜ 待开始 | — | — |
| 7 | Badge 显示 | ⬜ 待开始 | — | — |
| 8 | 开机自启 | ⬜ 待开始 | — | — |
| 9 | 错误处理 | ⬜ 待开始 | — | — |
| 10 | CI/CD | ⬜ 待开始 | — | — |
| 11 | 测试 | ⬜ 待开始 | — | — |

状态: ⬜ 待开始 | 🔄 进行中 | ✅ 已完成 | ❌ 阻塞

---

## 依赖关系

```
Phase 1 (Go --json) ──→ Phase 4 (进程管理)
                              │
Phase 2 (Xcode) ──→ Phase 3 (状态栏) ──→ Phase 5 (拖拽)
                                              │
                                        Phase 6 (历史)
                                              │
                                        Phase 7 (Badge)
                                              │
                                        Phase 8 (开机自启)
                                              │
                                        Phase 9 (错误处理)
                                              │
                                        Phase 10 (CI/CD)
                                              │
                                        Phase 11 (测试)
```

Phase 1 和 Phase 2 可并行执行。Phase 3 依赖 Phase 2。Phase 4 依赖 Phase 1。Phase 5 依赖 Phase 3 + Phase 4。Phase 6、7、8 均依赖 Phase 5，三者可并行。Phase 9 依赖 Phase 5，可与 6/7/8 并行。Phase 10 在全部功能完成后执行。Phase 11 在 Phase 10 完成后执行（CI 环境可用）。

---

## 文件变更清单

### Go 侧（Phase 1）

| 文件 | 变更类型 | 行数 |
|------|----------|------|
| `cmd/root.go` | 修改 | +5 |
| `internal/server.go` | 修改 | +20 |
| `internal/server_test.go` | 修改（新增 --json 测试） | +30 |
| `Makefile` | 新建（统一构建入口） | +20 |
| **合计** | | **+75** |

### Swift 侧（Phase 2-9）

| 文件 | 变更类型 |
|------|----------|
| `macos/GoGrip.xcodeproj/` | 新建 |
| `macos/GoGrip/GoGripApp.swift` | 新建 |
| `macos/GoGrip/AppDelegate.swift` | 新建 |
| `macos/GoGrip/Models/HistoryEntry.swift` | 新建 |
| `macos/GoGrip/Services/ProcessManager.swift` | 新建 |
| `macos/GoGrip/Views/PopoverView.swift` | 新建 |
| `macos/GoGrip/Views/DropOverlayView.swift` | 新建 |
| `macos/GoGrip/Views/HistoryListView.swift` | 新建 |
| `macos/GoGrip/Views/HistoryRowView.swift` | 新建 |
| `macos/GoGrip/Views/SettingsView.swift` | 新建 |
| `macos/GoGrip/Utilities/Storage.swift` | 新建 |
| `macos/GoGrip/Info.plist` | 新建 |

### 测试文件（Phase 11）

| 文件 | 变更类型 |
|------|----------|
| `macos/GoGripTests/StorageTests.swift` | 新建 |
| `macos/GoGripTests/ProcessManagerTests.swift` | 新建 |
| `macos/GoGripTests/PortReaderTests.swift` | 新建 |
| `macos/GoGripTests/IntegrationTests.swift` | 新建 |

### CI 侧（Phase 10）

| 文件 | 变更类型 |
|------|----------|
| `.github/workflows/release.yml` | 修改（新增 job + 更新 go-version） |
| `.github/workflows/build.yml` | 修改（新增 macOS job） |
