# Design

## Context

动机与产品边界见 [proposal.md](proposal.md)。行为以已批准的四份 delta spec 为准：[Finder 服务](specs/macos-finder-service/spec.md)、[预览会话](specs/macos-preview-sessions/spec.md)、[菜单栏 App](specs/macos-menu-bar-app/spec.md)、[候选打包](specs/macos-app-packaging/spec.md)。本文件只提出技术实现，不扩大这些规格。

当前实现的约束：

- `macos/GoGrip/` 是有缺陷的示例。主 plist 未声明 Services；工程仍嵌入 Finder Sync。旧进程管理以原始路径为键，丢弃 stderr，启动超时或 EOF 回退到 6419；不能保留这些行为。
- `internal/server.go` 已能解析目录/文件、生成带初始文章路径的 URL，并通过 `--json` 输出启动信息，但输出发生在进入 HTTP 服务循环之前。`Stop()` 只关闭 listener；热重载失败只写日志。
- `internal/listener.go` 当前绑定 `:port`，展示 URL 中的 `localhost` 不限制实际监听。宿主集成必须显式指定真实回环监听地址。
- `macos/project.yml` 与已提交的 Xcode 工程是两个配置来源；实际 Make/CI 使用后者。Go 构建脚本已合并两种架构，但不能据此认定宿主也是 universal。
- macOS 13 是 API 下限。候选产物不要求 Developer ID 或公证，本阶段不能据此免除实际功能验收，也不能宣称正式分发就绪。

## Goals / Non-Goals

**Goals:**

- 一个宿主协调器拥有目标身份、会话状态及所有子进程；入口和 UI 不另存一套运行事实。
- 分离三件事：所有权已经建立、HTTP 服务已经可用、系统已经接受浏览器打开请求。
- managed 集成与独立 CLI 明确区分，使用现有 Go 渲染器，不增加运行时依赖。
- 将平台风险落实到具体集成验收位置，不用权限提升、默认地址或假运行状态遮蔽失败。

**Non-Goals:**

- 不引入 XPC/launchd 服务、特权 helper、进程名清扫、父 PID 轮询或自动恢复框架。
- 不引入浏览器特定自动化、渲染界面重构、文件监视轮询或新的目录发现政策。
- 不为遗留 Finder Sync、URL 转发、旧进程管理器保留运行兼容层。
- 不把所有权管道泛化为任意进程树管理；它管理的是宿主直接启动的 Go 预览服务，不宣称解决任意 Chrome 等后代进程的所有退出语义。

## Decisions

### 1. AppKit 管生命周期，SwiftUI 只渲染面板

采用 AppKit 入口与 `NSApplicationDelegate`，由强持有的应用根对象创建协调器、Services provider、`NSStatusItem` 和 `NSPopover`；面板用 `NSHostingController` 承载 SwiftUI。主 plist 设置 `LSUIElement = true`，保留 bundle ID `com.showgp.GoGrip`。不保留旧的空 Settings scene，不建立管理主窗口。

协调器在注册 provider 前完成初始化；通过 `NSApp.servicesProvider` 注册服务后立即可接收请求，不等待首次弹出面板。AppKit 的 provider 文档明确服务请求可能在注册后、launch 回调结束前到达，因此不能将面板初始化当作 Services 初始化。关闭 popover 只影响显示。

UI 与状态变更由一个 `@MainActor` 协调器负责，SwiftUI 使用 macOS 13 可用的 `ObservableObject`/`@Published`，不依赖 macOS 14 的 Observation API。阻塞文件元数据操作、进程管道读写和 HTTP 启动探测不在主线程执行。

**替代方案：** 单用 SwiftUI scene 管所有生命周期会把 Services 与面板出现时机耦合；单用 AppKit 手写全部列表增加无必要的 UI 代码。这里保留两种框架各自明确的职责，不复用旧示例的状态和管理逻辑。

### 2. 一个 Services 声明，一个目标接收路径

主 App 声明单个 `NSServices`：

| 键 | 值/约定 |
|---|---|
| `NSMessage` | `openWithGoGrip`，provider 暴露 `openWithGoGrip:userData:error:` |
| `NSPortName` | `GoGrip`，与应用名一致 |
| `NSMenuItem.default` | `Open with GoGrip` |
| `NSSendTypes` | `public.file-url` |
| `NSRequiredContext.NSApplicationIdentifier` | `com.apple.finder` |

该服务只接收目标，不写回 pasteboard，不设默认键盘快捷键。以 `readObjects(forClasses:options:)` 读取全部 `NSURL`，设置 `.urlReadingFileURLsOnly`。不读任意文本或 shell 命令，不以 `urls.first` 丢掉多选，也不添加旧 sender 的兼容解析。

selector 同步取得 URL 值，交给协调器后返回；确认、启动和最终错误提示异步进行，不在 Services 调用中阻塞整个批次。`error` 指针只在 selector 返回前用于无法读取请求等即时错误，不能捕获到异步任务中。`NSOpenPanel` 同时允许目录与文件、多选，手动入口和最近记录也进入同一个目标处理路径。

`en.lproj/ServicesMenu.strings` 与 `zh-Hans.lproj/ServicesMenu.strings` 均以 plist 的精确默认字符串 `Open with GoGrip` 为键；普通 `Localizable.strings` 不代替 Services 菜单本地化。首次引导解释菜单与系统设置，提供再次查看入口。可调用 `NSUpdateDynamicServices()` 请求重扫声明，但不修改用户启用偏好，不把调用成功当成菜单出现的证明。

**替代方案：** Finder Sync 已被产品决策排除；只用文件 UTI 过滤菜单不能代替运行时分类，且会让混合选择的实际错误被系统菜单条件掩盖。使用 file URL 输入，应用统一确认目标类型。

### 3. 规范化身份与批次执行先后分离

目标准备在后台执行：标准化绝对 file URL、解析符号链接、取得实际类型及访问错误。目标身份使用规范化后的路径；不使用原始字符串，也不使用文件 inode/bookmark 作为跨移动的身份。保留用户所选路径作展示，真实打开使用已解析路径；不笼统转换大小写，以免破坏大小写敏感卷。

每批先准备和按身份去重，将无效项目收集成失败结果；超过 5 个不同有效目标时，在任何本批启动或浏览器打开前确认。取消不改变现有会话。确认后逐项异步处理，批内不建立额外并发调度框架；一个目标失败不停止余下目标，结束后一次汇总。不同请求可以在异步等待期间到达，但同一目标的启动只能由协调器的一份记录负责。

父目录、子目录和文件的规范化路径不同，保持独立。Go 启动仍是实际访问的最终判定者：准备后卷断开、权限变化或文件删除都作为该项失败处理，不依赖一次预检查保证后续成功。

**替代方案：** inode/bookmark 会引入跨移动跟踪语义；按包含关系合并会话会改变打开边界；先启动再确认会让取消不再无副作用。均不采用。

### 4. 一个会话事实源，回调携带启动代次

协调器按目标身份维护会话记录；进程适配器只报告事件，不自行管理另一份会话字典。每次实际启动分配一个 `generation` UUID，所有 stdout、退出、超时与浏览器回调都绑定该代次；旧代次回调不能删除或修改同路径的新会话。

会话执行阶段为 `starting → running → stopping → terminated`。启动失败或意外退出保存失败信息；进程未确认退出前不能仅删记录宣称已停止。`targetAvailability`、`reloadStatus` 和最后一次浏览器错误与阶段分开：服务仍运行但目标不可访问、监视降级或浏览器未打开，不等于服务已经退出。

| 输入 | 协调器动作 |
|---|---|
| 同目标正在启动 | 等待同一启动结果，不启动第二个服务 |
| 同目标已经运行 | 使用已确认 URL 请求浏览器打开 |
| ready 已验证且进程未退出 | 进入 running，更新最近记录，再打开浏览器 |
| ready 无效、启动 EOF/超时或早退 | 结束本次子进程，保留错误；不猜地址 |
| 正在停止时再次请求相同目标 | 等到旧进程实际结束后才允许新代次启动 |
| 退出事件 | 按是否为用户停止区分正常结束与意外失败，无自动重启 |

失败信息包括每个会话的最后原因和最近一次操作/批次报告，不存 Markdown 内容或持久进程状态。显式关闭失败报告只是清除显示，不改变运行事实。

**替代方案：** 仅用 path 字典加异步 termination handler 有旧回调删除新会话的竞态；用一组布尔值隐式表达执行阶段难以处理启动中停止。代次与明确阶段覆盖已批准的复用和退出边界，不增加恢复机制。

### 5. managed 模式建立专用机器契约

增加隐藏的宿主集成选项 `--managed <generation>`。宿主直接用 `Process.executableURL` 和参数数组启动内置工具，不经过 shell；目录附加 `-r`，目标前使用 `--`。managed 模式禁止导出模式等不属于预览的组合，禁止浏览器由 Go 自动打开。

managed 模式在文件解析之前建立所有权监视，指定真实 bind 为 `127.0.0.1:0`，由 OS 分配端口，使用 listener 的实际端口生成 URL。这样无需扫描默认端口区间，也不产生 100 个候选端口的隐含宿主上限。监听函数显式接收 bind 地址；展示 Host 与 bind 不再在宿主路径中混淆。

独立 CLI 的现有参数、默认端口及显式端口策略保持可用，`--json` 继续是独立启动信息输出，而不是旧 App 的兼容层。独立模式不读取所有权 stdin；本变更不擅自修改其既有网络暴露政策。App-only 回环限制不能被误述为整个 CLI 的安全保证。

managed stdout 专用于 UTF-8 NDJSON；普通日志与诊断写 stderr。协议版本为 1，每条消息带 `generation`，同一进程有一个序列化输出路径，不能让热重载和 HTTP 回调并发拼接 JSON。字段与事件如下：

| 事件 | 关键字段/作用 |
|---|---|
| `ready` | `version`, `generation`, `url`, 当前 `reload` 状态；同代次只成功发布一次 |
| `reload-status` | `state = pending / active / degraded`、可用原因；覆盖初始及运行期间已知失败 |
| `target-status` | `state = available / unavailable`、原因；来自实际目标访问结果 |
| `fatal` | 可分类 `code` 与实际 `message`；启动或服务致命失败后进程退出 |

适配器连续读取完整行，保留跨读取分片的残余数据，不把一次 pipe read 当作一个消息。版本、代次或 required 字段错误导致本次集成失败并终止其子进程，不做降级解析。stdout 帧与 stderr 诊断缓冲各设 64 KiB 上限，避免坏反馈无限积累；错误展示保留可获得诊断，不将 stderr 当机器状态解析。

**替代方案：** 继续读取一个 port 再拼 URL会丢失文件路由和启动失败；解析日志判断监视状态容易随文字变化破坏契约；仅加 PID 字段不能证明归属或可用性。选择显式机器事件，同时保留真正仍在使用的独立 CLI 模式。

### 6. readiness 验证不重复渲染正文

Go 先确认目标可访问、建立处理器、绑定 listener 并启动 HTTP 服务，再发 ready；watcher 尚在初始扫描时可报告 `pending`，不阻塞所有权监视。单文件初始路径沿用现有 URL escaping，不能丢失中文、空格或文章路由。

managed 服务提供一个仅供启动确认的固定路由 `/__gogrip/ready`：返回 204 及 `X-GoGrip-Generation`，不访问目标目录、不执行 Markdown 渲染。收到 ready 后，宿主校验 URL 是本代次的 HTTP 回环地址，再对同一 origin 的该路由做一次 HEAD，确认代次匹配、HTTP 可响应及对应进程未退出；随后才显示 running 和请求默认浏览器打开实际 URL。不接受跳转到其他 origin，不增加周期健康轮询。

启动从成功创建子进程起使用 15 秒内部期限，覆盖机器反馈和一次 HTTP 确认；超时按启动失败清理，不请求默认端口。这个数值是初始工程设置，不提供配置 UI，不对阻塞的内核文件操作承诺精确时间；若实际验收需要调整，只调整实现参数，不改变超时失败的语义。

浏览器通过 `NSWorkspace.shared.open(url)` 打开，返回 false 则报告操作失败并保留服务；返回 true 只表示系统接受打开，不宣称页面已渲染或标签页已聚焦。复制操作使用同一个已验证 URL。

**替代方案：** 仅凭 bind 成功或收到 port 尚未验证 HTTP 路径；反复 GET 正文会重复解析文档。一次轻量启动路由验证是内部集成手段，不扩展浏览器界面或建立健康检查系统。

### 7. 独立所有权管道覆盖宿主异常退出

每个 Go 子进程有专用 `Pipe` 作为 `Process.standardInput`；宿主仅持有其写端，Go 仅获得 stdin 的读端。与 stdout 机器反馈分开，管道不承担心跳、重连或命令队列。

启动适配器在串行的 descriptor 建立与 launch 临界区设置并检查非 stdio 端的 `FD_CLOEXEC`，尤其是所有权写端；不复制写端、不传给其他进程。将 `Pipe` 对象本身交给 `standardInput`，使 Foundation 按文档在 launch 后关闭宿主读端。必须实际验证子进程及其他 Go 会话没有继承写端；写端遗漏在任一进程中都会掩盖 EOF。

managed Go 在任何目标解析、网络卷读取、watcher 或 listener 创建之前，启动独立 stdin EOF 监视和取消上下文。宿主主动停止时关闭写端；宿主崩溃或 SIGKILL 时由内核关闭其 descriptor。EOF 或所有权通道读错误按 owner loss 处理，取消服务运行，并由独立 watchdog 在最多 2 秒的有界善后后强制退出；不能等文件读取或主启动函数返回才退出。

正常取消使用 HTTP 服务的有界关闭、停止并等待 watcher 结束、关闭 WebSocket 客户端，并释放已经初始化的 PDF 资源，不在停止时创建 PDF generator。必须允许强制退出绕过 Go defer，以免阻塞文件系统操作让服务继续无人管理；内核会回收 Go 进程的 listener 和 watcher descriptors。App 仍在运行而停止迟迟未完成时，适配器在 4 秒后对自己确切拥有、尚未确认退出的子进程使用 SIGKILL，并等待退出结果；不得按名字或扫描未知 PID 清理。

正常退出 App 使用 `applicationShouldTerminate` 的延后完成路径，禁止新启动，停止包括 starting 在内的全部记录，确认后回复允许退出。SIGKILL 不依赖这个回调，而依赖 EOF。独立 CLI 未启用 managed 时完全不进入该所有权路径。

**替代方案：** `applicationWillTerminate` 不会在 SIGKILL 后执行；普通 `Process.terminate()` 和进程组不保证父进程死亡时清理；launchd helper 或 PID 轮询增加不必要生命周期与身份管理。管道直接表示当前拥有者是否还存在，且覆盖 ready 之前的窗口。

### 8. 结构化访问与热重载状态，不增加扫描政策

在现有访问/发现错误分支增加宿主事件出口。所选根目录或单文件的访问失败报告 unavailable；不能将普通图片 404 或一个不存在的子页面一律认作整个目录目标失效。访问再次成功时可以报告实际观察到的 available，但不后台寻找新路径或重连卷。浏览器继续使用现有 HTTP 错误页面，原可访问空目录继续使用现有空状态。

热重载增加可选状态报告出口，覆盖 watcher 创建失败、walk error、预算跳过、注册失败及运行期 watcher 错误；提供初始状态快照，后续只报告状态变化及相关原因，不逐条日志驱动 UI。状态回调在 watcher 初始扫描开始前安装，不能丢掉 ready 之前的降级。独立 CLI 无宿主事件接收方时仍保留日志，不为所有文档请求额外分配状态通道。

不改变已有忽略目录、预算与 Markdown 发现规则，不加轮询。已知缺失覆盖标为 degraded；网络卷提示自动刷新不保证始终可靠，不声称能发现所有静默丢失的文件事件。原路径恢复后也不自动宣称原 watcher 已恢复，用户可停止后重新打开。

**替代方案：** 解析 stderr 掩盖类型及顺序；自动扫描恢复会新增未批准的重连和轮询行为。只扩展已批准会话状态所需的出口。

### 9. 最近记录、错误及登录项不混入进程持久化

使用独立的版本化 UserDefaults 键保存最多 20 个 `RecentTarget`：规范化目标路径、用户选择的展示路径、目标类型和最近使用时间。成功建立会话或主动重开运行目标时，按身份更新 recency；浏览器打开失败不抹掉已经成功建立的目标记录。最近项目重开定位其已记录的规范化目标，不借助 bookmark 跟踪移动。

不持久化 PID、端口、运行阶段或正文。清空最近记录只清该键，不接触会话字典。遗留 50 项原始路径记录不接入新运行代码，也不自动删除用户旧数据；新 schema 从自己的键开始，首次升级不假称已迁移旧历史。

登录项使用 `SMAppService.mainApp`，读取真实 status；首次未注册时默认关闭，仅用户操作才 register/unregister。系统要求批准或操作失败时显示真实状态和原因，不保持一个已失效的 UI toggle，不擅自注销用户已经主动注册的登录项。

错误展示有一份操作/批次报告及会话的实际最后原因。用户操作失败由 `NSAlert` 汇总一次，不依赖通知授权；运行期状态失效在面板明确展示。UI 使用标准 `en`/`zh-Hans` 资源与英文 development region，不另设语言开关。

**替代方案：** 保存 RunningInstance 或旧 50 项 schema 会恢复假的运行状态；动态文件 bookmark 与用户选择的按路径会话不一致；仅保存登录 toggle 无法反映系统拒绝。选用小型持久目标模型和系统实际状态。

### 10. 单一工程来源与自包含候选包

以已提交的 `macos/GoGrip.xcodeproj` 为唯一工程定义，沿用当前 Make/CI 的构建入口；切换时移除未被流水线使用的 `macos/project.yml`，避免旧 generator 再生成 Finder Sync，不新增 XcodeGen 依赖。项目仅保留宿主与行为测试目标，移除旧扩展及嵌入阶段。

宿主显式构建 `arm64 x86_64`、`ONLY_ACTIVE_ARCH=NO`、deployment target 13.0。Go 脚本构建 darwin arm64/amd64，使用现有工具链及 `CGO_ENABLED=0`，再合并两种架构。内置工具放入 `Contents/MacOS/go-grip`，宿主只从该确定位置加载；不从 Resources、PATH 或源码回退，不在运行时 chmod 修改包内容。

候选包采用本地 ad-hoc 签名：先给内置工具独立 code identifier 签名，再由 Xcode 签宿主 bundle，验证签名结构；不需要 Developer ID 私钥或公证凭据。ad-hoc 只是本地运行签名，不解决跨版本 TCC 身份稳定性，也不满足正式分发标准。现有全局 `CODE_SIGNING_ALLOWED=NO` 构建覆盖需随候选路径更新。

统一以显式 archive 输出取得 App 并生成含 Applications 安装入口的候选 DMG，不再假设 DerivedData/App 输出恰好位于旧 Makefile 的 `macos/build/Release`。检查宿主与 Go 工具的两种架构及部署目标；架构检查不代替实际运行。

现有 CLI release 资产流程不在本变更中重做；macOS App 输出作为候选 CI artifact，不自动上传为正式 GitHub Release。后续 Developer ID、公证和正式干净安装仍需另行批准。

**替代方案：** 保留两个工程配置会重复维护并重新引入旧扩展；只生成 universal Go 工具不能使宿主 universal；完全禁用签名会增加本地系统集成的身份不确定性。选择一个工程来源、一个内置工具路径和明确候选身份，不提前建设正式发布系统。

## Verification Strategy

以下是实现后的验证边界，不是 tasks 清单，也不代表本轮已完成这些验证。

- **状态与批次 seams：** 规范化及符号链接去重、同目标并发、旧代次回调、退出早于 ready、启动中停止、5/6 边界与取消、部分失败保留成功、浏览器失败不抹掉运行状态。测试验证外部状态与实际动作，不固定内部类名或简单转发。
- **持久化与系统 seams：** 隔离 UserDefaults 域中的 20 项淘汰、重复重开、重启后不恢复会话、清历史不停止；登录注册失败与真实状态；ServicesMenu 双语资源通过实际菜单确认，而非只搜字符串。
- **真实 Go 集成：** 临时目标含中文、空格、嵌套 Markdown 和空目录；实际 HTTP 内容/URL、回环 listener、owned 与 standalone 共存、管道错误、进程早退和 watch 降级。使用隔离目录、独立端口和明确的进程清理，不以 fake shell 输出 port 或 merely isRunning 作为完整集成证明。
- **所有权：** 从生产 Foundation launch seam 启动真实 managed Go；分别在 ready 前与运行中强制终止拥有者，观察该 Go PID 和 listener 消失；另一 owned 会话及独立 CLI 不受单会话停止影响。覆盖多会话 writer 不继承的情形，不能只证明正常 quit。
- **原生实际路径：** 从候选 App 在 Finder 调用已启用服务，包含冷启动、多选和拒绝权限；观察浏览器正确内容、面板状态、重开与停止、登录项以及实际系统授权主体。保护目录与外接卷需实际访问，不能由 plist 或源码推断；挂载网络卷的实际访问经批准延期至后续 change（见下"本阶段验收范围与延期"）。
- **候选包：** 使用 archive 中的 App/DMG，脱离源码目录与开发 PATH；检查两种宿主/工具架构，在可用的目标系统及架构实际运行并明确记录覆盖。当前机器的运行不能冒充 Intel 或 macOS 13 运行证明；Intel 与 macOS 13 实机运行经批准延期至后续 change（见下"本阶段验收范围与延期"），其覆盖状态在证据汇总中如实记录为未验证。

**本阶段验收范围与延期（2026-10-06 批准的范围调整）：**

- 本阶段必验范围＝最终候选（12 重新交付版本，身份见票 12/13）在当前 Apple Silicon 与当前 macOS 的实测集合，加上本机补做项：TCC 拒绝路径与真实授权主体、目标移动、服务关闭状态下再次查看 Help、简体中文界面与 Finder 中文服务名（含不支持语言回退）、登录项真实开启/关闭、外接卷只读访问。
- 已裁决：候选级"系统拒绝浏览器打开"分支不列为延期项，其验收以既有消费者行为回归的确定性结果为准（见 `specs/macos-app-packaging/spec.md` 的 `Acceptance evidence for the browser-open failure branch`）；受控辅助仅在既有回归不足以覆盖且另行批准时使用，不作为必需证据；候选级真实系统拒绝未触发须如实记录，不得表述为已通过系统级拒绝验证；本分支标准不替代正常浏览器打开路径的实际观察。
- 延期至后续 change（不属本阶段完成条件、不得勾选为完成）：Intel 实机运行、macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷；本阶段不准备对应设备/环境。
- 证据层级：开发构建与旧候选记录只作对照；本阶段声明以最终候选实测为准，延期项保持"未验证"。不得以 universal 构建、架构/部署静态检查或替代环境结果顶替实际运行证据，也不得把旧证据表述为最终候选证据。
- 延期项的行为要求不删除；所需环境、证据标准与完成标准由后续 change 承接，逐项映射见 tasks.md 延期清单与 coverage-plan。

旧测试中固定 EOF/超时返回 6419、50 项原始历史等与规格冲突的断言应移除，不重新固化。永久测试只围绕批准的行为、边界及竞态；机器 API 声明、资源搬运和纯转发以构建/临时 smoke 验证，不新增文字或 wiring 测试。

本轮仅做了一项临时机制 smoke：Swift/Foundation 宿主持有设置 `FD_CLOEXEC` 的所有权 Pipe 写端，`/bin/cat` 子进程等待读端；对宿主发 SIGKILL 后，子进程消失。它证明当前环境下该 descriptor 配置的 EOF 行为，不证明尚未实现的 Go watchdog、App UI、Services、TCC 或候选包。没有创建仓库实现文件。

## Risks / Trade-offs

- [任一进程继承所有权写端会阻止 EOF] → launch 临界区检查 CLOEXEC，禁止 writer 复制，并用生产 launch seam 的多会话/SIGKILL 场景验证，不以 Foundation 文档未明确保证的部分作假设。
- [网络文件系统可阻塞启动、请求甚至进程拆除] → 元数据操作离开主线程；owned watchdog 独立于这些操作，执行有界善后和强制退出。不对不可中断的内核 I/O 承诺精确硬截止。
- [Services 索引、用户偏好及菜单条件受系统控制] → 声明与 provider 均完整配置，提供引导并验证实际 Finder 冷启动，不用修改用户偏好或强制重启 Finder 代替验收。
- [TCC 对内置 Go 的实际授权主体不能由打包关系推断] → 使用非沙盒 App 的正常父子 launch，不 detach、不提升权限；在原生入口与 Go 访问首次接通时优先验证受保护目标。若实际证据否定当前直接内置工具路线，停止该集成并与用户重评设计，不预埋 helper 或 FDA 兜底。
- [ad-hoc 签名的代码身份随构建变化，权限可能需重新授予] → 固定 bundle ID 和可验证签名结构，说明候选阶段限制；不承诺跨构建授权稳定，正式签名阶段另行安排。
- [EOF 强制退出不执行全部 Go defer，也不自动管理任意后代] → 正常关闭现有已初始化资源；进程退出回收直接服务的内核资源。Chrome 等后代生命周期不宣称由本管道保证，也不扩大为通用进程组清扫。
- [新历史键不显示旧示例的 50 项记录] → 切换说明明确这一点，保留用户旧存储不自动删除，不把旧格式接入新运行代码。
- [跨编译不等于最低系统或 Intel 实际兼容] → 显式 deployment/architecture 检查及真实运行矩阵；缺少环境只记录未验证，不能据此声明全部平台已通过。
- [延期实机/卷验收与支持声明边界] → README/架构、证据汇总与候选说明明示未验证环境；不得以 universal 构建、部署信息检查或替代环境结果顶替实际运行证据；后续 change 取得实机证据前不改变该边界。
- [归档后主规格包含本阶段未实机验证的平台/卷场景] → 已批准的处理方式：保留要求文本、明示延期并由后续 change 补齐证据；这是"要求保留 + 声明边界"，不是删除功能要求。若要求 durable specs 不含未验证承诺，需另行批准收窄规格或推迟归档。

## Migration Plan

1. 先实现并验证 managed Go 的回环、事件、启动探测及所有权边界，保持独立 CLI 可用；这一集成是新宿主的真实基础，不先建设假进程 UI。
2. 在完整原生入口接通阶段验证 Services 冷启动及 TCC 目标访问，再完成协调器、面板、历史、本地化和登录项；每个批准 ticket 按约定 seams 做 scoped TDD 和真实 smoke。
   - **08 → 09 的阶段性顺序例外：** 用户批准本阶段将实际已挂载网络共享验证暂列为已知缺口。在 08 已交付的真实授权主体、受保护目录允许/拒绝、外接卷访问及失败指导已通过独立 Standards/Spec 审阅，且无未决路线裁决的前提下，允许先推进 09 的最近目标与完整面板，不再以尚缺网络共享环境作为本票实施的硬前置。该例外仅调整 09 的实施顺序，不代表完整权限/卷 gate 或网络卷路线已经通过。
   - **缺口归属与关闭边界（2026-10-06 范围调整后）：** 网络卷的规格与行为要求不删除、不缩减；其实际访问/拒绝、适用时已知监视降级与手动刷新及 owned 资源清理证据延期至后续 change，通过真实 native→内置 Go 路径在具备环境的候选上补齐并独立审阅。08 与 3.3 按调整后范围关闭（网络卷项列入 tasks 延期清单）；补齐前不得宣称网络卷支持已验证，09 完成不替代这些证据。
   - **路线风险仍保留：** 网络卷上的实际访问、监视及收尾仍未验证；该证据与路线重评责任随延期清单进入后续 change。若后续证据否定直接内置工具路线，停止相关集成并与用户重评设计，不自行添加 helper、detach 或 FDA 兜底。
   - **审批边界不变：** 先逐项批准并对齐 `tasks.md` 和 09 票的依赖/验收前置，再按单票流程批准 09 范围、scoped TDD seams、native smoke、具体数据/环境操作和实施。本设计修订不关闭 08、不勾选任何任务，也不授权编码、写测试或环境变更。
3. 在新宿主接通时移除旧 App 的相应管理逻辑、Finder Sync target/嵌入、URL 转发、旧测试与双工程配置，不同时发布新旧入口。
4. 候选包统一 archive/签名/DMG 路径，作为候选 artifact 交付；CLI 资产发布保持原流程，不代替用户执行推送或发布。
5. 升级前明确要求结束旧示例运行会话并退出旧宿主，避免系统选择另一个同 bundle ID 的安装副本；不按进程名自动清理可能属于用户独立 CLI 的服务。新 schema 不自动恢复旧服务或删除旧数据。
6. 回退时先停止新宿主拥有的服务并退出，再使用用户明确选择的旧版本或独立 CLI；不保留旧入口 shim，不把回退方案解释为旧示例已可靠。
7. 延期验收与后续承接：本阶段在最终候选上完成 Apple Silicon/当前 macOS 实测与本机补做项后，按调整后范围关闭 08/13 与 3.3/7.1–7.3；Intel 与 macOS 13 实机、已挂载网络卷、候选级断卷验收延期至后续 change（另行批准创建），由其承接所需环境、证据标准与完成标准。延期期间 README/架构与候选说明必须明示未验证边界；实现代码若变更，须重新交付候选并对受影响路径复验，既有证据保持原候选身份不改写。

同步主规格、归档以及正式分发仍需独立批准。本文件未创建实施 tickets 或 tasks，也不授权开始实现。

## References

- [Apple：Services provider 注册及可能立即到达的请求](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/SysServices/Articles/providing.html)
- [Apple：Services 属性与 ServicesMenu.strings](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/SysServices/Articles/properties.html)
- [Apple：Process.standardInput 与 Pipe 读端关闭](https://developer.apple.com/documentation/foundation/process/standardinput)
- [Apple：execve descriptor 的 close-on-exec 规则](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/execve.2.html)
- [Apple：macOS 文件访问授权](https://support.apple.com/guide/security/controlling-app-access-to-files-secddd1d86a6/web)
- [Apple TN3127：ad-hoc 身份及内置工具代码标识](https://developer.apple.com/documentation/technotes/tn3127-inside-code-signing-requirements)
