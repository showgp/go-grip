# 05 — 原生单目标手动预览与真实退出

Status: done
Blocked by: 04-target-session-control
Covers OpenSpec tasks: 3.1 (app-root, controls, bundled-tool, legacy-cutover and quit contribution), 3.2 (single-target manual-entry contribution), 3.4 (single-target browser and error-recovery contribution)
Behavior source: ../specs/macos-menu-bar-app/spec.md, ../specs/macos-preview-sessions/spec.md, ../specs/macos-finder-service/spec.md, ../specs/macos-app-packaging/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

实际构建并本地签名的菜单栏开发 App，通过 NSOpenPanel 选择一个目录或 Markdown 文件，进入 04 的生产协调器及 ManagedProcess，用内置 Go 提供预览，使用已验证实际 URL 请求默认浏览器并观察正确内容。基础运行区域可重开、复制 URL、停止单个/全部及退出；浏览器失败保留可用服务和恢复入口，正常 App 退出先禁止新启动，确认包括 starting 在内的 owned 服务真实结束后才允许退出。

这是可独立使用和验收的“手动选择 → 实际 App → 内置 Go → 默认浏览器 → 原生控制/真实退出”纵向交付，不是 AppKit 或进程创建占位。本票只接通单目标手动选择；多选/批次由 06，Finder Services 注册及冷启动由 07，未接通的入口不能声明可用。

用户已批准将原大票拆成三张。本票保留已发布的稳定 ID/文件名 `05-native-finder-preview`，但标题和交付以本次修订为准；旧的 Finder/批次整票范围由本次修订替换，不保留第二张旧票或旧 API 兼容层。这里只修订发布 05；06、07 和后续票均未发布、未获准实施。

发布不授予实施许可，修订后的范围、TDD seams 和实际 smoke 仍须另行审阅/批准。`tasks.md` 是唯一进度账本；本票对 3.1、3.2、3.4 仅作部分贡献，全部本票验收及独立审阅完成后可关闭本票，**不能因此勾选任何上述完整任务**。3.1 还需 07 的 Services/provider；3.2 还需 06 的手动多选及 07 的 Finder；3.4 还需 06 的批次汇总及 07 的入口集成。2.3 完整归 06。本次不拆分或改写 task checkbox，不改变行为规格。

## Source mapping

批准 specs 是行为权威。下表列出 requirement 与全部原名 scenarios，但只交付“05 贡献”列；列出一个场景不代表本票验收其未来入口、历史或候选部分。

| Requirement / scenarios | 05 贡献及剩余边界 |
|---|---|
| macos-menu-bar-app — Persistent native menu bar host：`Open and dismiss the panel` | AppKit 应用根、status item/popover 与 SwiftUI 基础运行区域；LSUIElement、无 Dock/终端/独立管理主窗口；关闭面板不停止服务。候选由 13。 |
| macos-menu-bar-app — Visible sessions and explicit controls：`Distinguish targets with the same name`、`Copy an active preview address`、`Exit through the panel` | 唯一协调器的路径、实际阶段/URL及可获得错误；明确重开/复制/单个停止/全部停止/手动打开/退出。最近区域与完整访问性/降级布局由 09，候选由 13。 |
| macos-menu-bar-app — Manual opening uses the same target contract：`Manually reopen a Finder-started target`、`Choose a new target manually` | 单目标 NSOpenPanel 与生产目标身份/会话行为；不建立另一套手动运行字典。多选由 06，Finder 跨入口复用由 07，最近入口由 09，候选由 13。 |
| macos-menu-bar-app — Action failures remain understandable and recoverable：`Open a stale recent target`、`Recover from browser opening failure`、`A service exits unexpectedly` | 单项目标/启动/浏览器失败有原生提示及可查看原因，浏览器失败不抹掉服务，意外退出不显示 running。批次由 06、Services 接入由 07、最近错误由 09、候选由 13。 |
| macos-finder-service — Supported open targets：`Receive a Markdown file`、`Receive a directory outside the home folder`、`Receive an unsupported file` | 手动入口共享目标契约：目录及常规大小写不敏感 .md、普通主目录外目标、拒绝不支持文件且不转父目录。不是 Finder 证明；Finder 由 07，真实卷/TCC 由 08，候选由 13。 |
| macos-finder-service — Minimum necessary file authorization：`Access is denied`、`Open an ordinary accessible target` | 普通可访问目标不要求 FDA/辅助功能/自动化；已有准备/Go 访问失败按目标和原因报告，不假称成功或空目录。实际授权主体、指导及卷访问 gate 由 08，候选由 13。 |
| macos-preview-sessions — One session per normalized target：`Reopen a target through a symbolic link`、`Receive concurrent requests for the same target` | 手动打开/重开使用 04 的身份、单飞和复用；复用实际 URL 再请求浏览器。批次别名由 06、Finder 跨入口由 07、候选由 13。 |
| macos-preview-sessions — Containment does not merge target identities：`Open a child directory and a file under an active parent` | 分别手动打开父/子/文件，保留三个独立生产会话和浏览边界，不按递归包含关系复用。候选由 13。 |
| macos-preview-sessions — Directory and single-file preview modes：`Preview nested documents`、`Preview a selected file`、`Keep an empty directory session`、`Add Markdown to a watched empty directory` | 手动原生成功路径沿用 01/02 的递归目录、直接文件、空目录和有效监视；本票实际验证对应内容，不重做 Go 发现政策。Finder 由 07，候选完整文件事件/浏览器矩阵由 13。 |
| macos-preview-sessions — Verified startup and actual preview URL：`A nondefault port is used`、`Open a file with spaces and non-ASCII characters`、`Startup ends without valid readiness information` | 原生成功只使用 03/04 的已验证完整实际 URL，中文/空格文件定位正确；失败不猜端口/假 running/遗留服务。Finder 由 07，候选由 13。 |
| macos-preview-sessions — Loopback-only application previews：`Reach a preview from the same machine`、`Attempt to use a nonloopback interface` | 手动 App 只走 production managed 路线，观察真实回环 listener，不增加共享入口或改变独立 CLI。候选实际接口覆盖由 13。 |
| macos-preview-sessions — Browser opening is separate from service lifetime：`Browser opening fails after service startup`、`Close the browser preview` | NSWorkspace 默认浏览器结果与服务生命周期分开，失败保留服务/URL及重开/复制；重复打开再请求浏览器，关闭浏览器不回收服务。完整面板由 09，候选由 13。 |
| macos-preview-sessions — Explicit stopping confirms service termination：`Stop one of several sessions`、`Stop all application sessions` | 原生单个/全部停止调用 04，等待实际结束；失败保留管理入口，不删行伪装退出，不影响其他 owned/CLI。候选由 13。 |
| macos-preview-sessions — Services end with their owning application：`Quit the application normally`、`Force termination of the host`、`Host exits during startup` | App 延后退出与准备中请求的启动准入，收尾 starting/running；沿用 01/03 所有权，记录实际 App 正常/异常退出的 owned 资源释放。06 补确认恢复准入，07 接入 Services 请求，13 验证候选矩阵。 |
| macos-preview-sessions — No automatic session restoration or restart：`Relaunch with recent targets`、`A preview process exits unexpectedly` | 新根不恢复旧运行状态，退出原因可查看且不自动重启，允许主动手动重开。最近列表由 09，候选由 13。 |
| macos-preview-sessions — Independent CLI operation remains available：`Launch the renderer without the application` | 实际 App 控制和退出只收尾自己拥有的 managed 实例，独立 CLI 仍提供其原目标内容，政策不改。候选由 13。 |
| macos-app-packaging — Universal self-contained application：`Use the application without development tools`、`Use the same candidate on supported architectures` | 开发构建确定内置工具路径及本地 ad-hoc 签名；无生产 Resources/PATH/源码回退或 runtime chmod。完整 universal/archive 由 12，无开发工具/支持环境候选由 13。 |
| macos-app-packaging — Candidate DMG and Services-only Finder integration：`Produce an installable candidate`、`Use the new Finder integration` | 删除与旧运行入口一并废弃的 Finder Sync/嵌入/专用 URL 转发和 generator，不保留兼容入口；本票尚无 Finder 功能，07 接通 Services，10 引导/本地化，12 DMG，13 候选实际安装路径。 |

技术依据为 [design Decisions 1、3–4、6–7、9–10](../design.md#decisions)：AppKit 根/唯一事实源、后台目标身份、实际 URL/默认浏览器、真实退出、原生操作报告及内置工具。Decision 2 的 Services 接入由 07、Decision 3 的批次执行由 06；本票不实现或伪装这些剩余行为。

## Scope and implementation handoff

- **应用与工具**：切换 `macos/GoGrip/GoGripApp.swift` 的空 Settings 入口、`AppDelegate.swift` 的旧运行根为强持有 AppKit 根/生产协调器；保持 `com.showgp.GoGrip`，SwiftUI 只渲染运行区域。`Info.plist` 设置 LSUIElement；本票不新增 NSServices/provider/selector 占位。`macos/Scripts/build-go-grip.sh`、唯一已提交工程及必要 `Makefile` 开发构建/运行入口将真实 Go 放在 `Contents/MacOS/go-grip`，对工具与宿主本地 ad-hoc 签名，用实际构建产物运行。完整 universal/archive/DMG/CI 留给 12。
- **手动入口与事实**：实际 NSOpenPanel 可选一个目录或文件，复用 `TargetPreparation`、`PreviewSessionCoordinator` 与 `FoundationManagedProcessFactory`，真实准备/管道/HTTP 不阻塞主线程；保留已裁决的解析后路径 .md/常规文件判定。不同手动请求仍用同一事实源；不预建多选调度、数量确认、批次报告或 Finder provider 框架。06 在已有生产手动路径上交付完整多选，07 再让 Services 使用同一路径。
- **控制与错误**：替换 `Views/PopoverView.swift` 的旧历史主界面，用目标路径、阶段及可获得错误配合明确重开/复制/停止/退出动作；NSWorkspace.open false 是浏览器操作失败，不能结束仍可用服务；true 只表示系统接受请求，不证明页面渲染。错误与实际会话事实一致、可查看，成功不额外提示，不从 stderr 推断状态或丢单文件路径。不建第二份运行字典；必要的最小浏览器/操作错误事实不改变进程阶段。
- **退出**：`applicationShouldTerminate` 先禁止新启动，再延后回复、await 04 的真实全部收尾，包括 starting；准备或启动结果恢复后不能越过退出准入创建 child/打开浏览器，停止未确认不提前允许退出。退出不依赖仅 `applicationWillTerminate`、删行或信号发出；保留独占 ownership 管道、4 秒确切 owned 兜底和 Go watchdog，不清扫未知 PID。不对阻塞元数据/不可中断内核 I/O 承诺硬截止。
- **旧链路切换**：迁移全部被替换调用，移除 `ProcessManager`/`RunningInstance`、port/fallback/独立 JSON 宿主解析、runtime chmod、旧空 Settings、废弃旧视图/拖拽入口及冲突测试；移除 `macos/GoGripFinderSync/`、target/依赖/嵌入、专用 `gogrip` URL 转发及 `macos/project.yml`。不新增 shims，不留下旧扩展请求新宿主的断链。新 App 只声明当前手动能力；Finder 菜单重新可用须等待 07 的真实接通。
- **历史边界**：新宿主不读取/写入旧 `Storage`/`HistoryEntry` 的 50 项 schema，不迁移或删除用户旧 `go-grip-history` 数据，不实现新历史或历史清除。旧存储源/测试的最终切换与新最近 20 项仍由 09 处理，不重钉旧断言。07/09/10 等历史编号在已关闭票中的旧含义见覆盖地图对照，本票与未来发布以当前地图为准。
- **检查与说明**：复用现有工程、scheme、test target 及有效 01–04 行为回归，按需调整真正受影响的编译/构建调用；obsolete mock port、non-empty、default/wiring/资源存在和仅进程创建测试删除而非重钉。随实际证据更新 `README.md`、`docs/ARCHITECTURE.md` 与受切换影响的旧说明，明确开发 App 当前只有单目标手动入口、无 Finder/批次能力及 08–13 的后续边界。

### Proposed scoped TDD seams and real verification

以下待实施前批准，发布不授权先写测试。新增永久测试仅捕获批准的独立消费者回归，复用已有目标/代次/停止覆盖，不新增 plist/签名/资源搬运/转发/文案/默认值或重复同路径测试。

1. **单项浏览器/恢复 seam**：最小受控系统打开结果，证明启动成功而浏览器拒绝后仍有真实 running/实际 URL及可恢复动作；目标/启动失败无浏览器成功或假 running，成功不产生失败报告。断言用户可观察结果，不固定提示文字或只数调用次数；不要重复 03/04 的内部矩阵。
2. **原生退出 seam**：用最小受控准备恢复及 starting/running 收尾，证明退出准入后无新 child/浏览器副作用、未确认服务退出时不能允许 App 退出、迟到启动结果不恢复。验证实际结果而非 applicationShouldTerminate 的方法转发；06 的数量确认恢复测试不提前编写。
3. **实际开发 App smoke**：实际构建、本地签名并运行 App，通过真实 NSOpenPanel 分别打开普通目录、中文/空格/.MD 文件与空目录，观察默认浏览器正确内容及路径；分别手动打开父/子/文件并观察独立内容，原路径/符号链接重开复用实际 PID/URL。操作真实面板重开/复制/单个/全部停止及退出，记录本次 App/owned PID、generation、URL/HTTP内容及端口释放；运行期 App SIGKILL 沿用所有权证明，独立 CLI 在控制/退出前后仍服务。

浏览器失败可以用受控拒绝证明语义并观察实际原生错误表面，须明确其证据来源；正常默认浏览器路径必须实际观察页面，不能把 NSWorkspace true、04 的临时 Foundation 拥有者、mock echo 或 XCTest 通过当成本票 App smoke。原生交互环境缺失时保持相应验收未完成，说明需要的实际操作/证据，不自行授予 App FDA/辅助功能/自动化或改走 helper。

## Prerequisites and non-goals

唯一直接前置 [04](04-target-session-control.md) 已 done，01–03 间接满足，生产身份/单飞/generation/结构化事实/真实停止及 Go 契约可复用。04 的 61 passed/0 failed/0 skipped、真实链路 54 项 smoke 是前置证据，不是本票验证。其已记录的主线程同步 emitter/MainActor 前提及失败测试路径 advisory 不自动增加本票永久测试或行为保证。

不做 2.3 批次/多选/5或6确认/取消/一次汇总（06）；不做 Services 声明/provider/selector/Finder 冷启动或跨入口验收（07）；不做 3.3 受保护目录/TCC/外接或网络卷 gate（08）；不做任务4–7的新历史/完整面板/引导/本地化/登录项/universal/archive/DMG/候选CI/候选支持环境（09–13）。不新增 Go 行为、扫描政策、轮询/恢复/重启/移动跟踪/重连、拖拽、标签页管理、高级参数、helper/XPC/launchd、进程名清扫或 Chrome 后代管理。实际普通目标接通若否定已选工具/权限路线，保持相关验收阻塞并请求设计裁决，不把问题藏到 08 或绕过权限。提交/推送、主规格同步、归档、正式签名公证/安装/公开发布仍未授权。

## Acceptance

- [x] 实施前另行批准本次修订范围、两项 scoped TDD seams 与实际开发 App smoke；记录 red/green，每项新永久测试对应独立批准回归，复用01–04，不新增wiring/字段复制/mock echo/文案/默认值/资源存在或仅isRunning证明。
- [x] AppKit根强持有唯一生产协调器及status item/popover，使用macOS13可用API、LSUIElement；实际App无Dock/终端/独立管理主窗口，开关面板不停止服务，不留空Settings或第二份运行事实源；不新增未接通Services占位。
- [x] 实际开发构建将当前Go置于Contents/MacOS/go-grip并本地ad-hoc签名工具/宿主，运行从确定内置路径和实际产物启动；无生产Resources/PATH/源码回退或runtime chmod，记录构建/签名及运行结果，不冒充universal/TCC/候选证明。
- [x] 迁移全部本票替换调用并删除旧ProcessManager/RunningInstance、端口fallback/独立JSON宿主解析、旧视图/Settings/废弃拖拽、FinderSync源码/target/依赖/嵌入、专用URL转发、project.yml及冲突mock/default/wiring测试；新宿主不接入旧50项历史，不删除用户旧数据，不新增兼容层或提前实现09历史。
- [x] 真实NSOpenPanel单选目录/常规.MD文件，经生产后台准备/协调器/适配器启动内置Go；普通主目录外目标可用，中文/空格单文件在默认浏览器定位正确内容，可访问空目录仍有可用会话，失败显示目标与可获得原因且不转父目录、猜端口、假running或遗留本次服务。
- [x] 真实手动原路径/符号链接重开复用同一有效PID/generation/实际URL并再次请求浏览器，分别手动打开父/子/文件保持独立内容和浏览根；记录HTTP/实际回环listener，不重做03/04内部协议/期限矩阵或声明Finder跨入口已完成。
- [x] 浏览器失败保留实际running/URL和可查看错误，明确重开/复制可恢复；正常打开不额外提示成功，真实关闭浏览器后服务继续。受控拒绝与系统真实结果分别标注，正常浏览器页面已实际观察，不能用返回true代替内容证明。
- [x] 实际基础面板能通过路径/阶段/原因定位目标并重开、复制实际可用URL、停止单个/全部、手动打开及退出；复制地址定位原内容，停A确认PID/端口释放且B/独立CLI继续，停止失败保留管理入口，意外退出不假running/自动重启。
- [x] applicationShouldTerminate先关闭新启动准入并延后回复，确认包括starting的全部recorded owned服务真实结束才允许退出；受控准备/启动恢复证明退出后无新child/浏览器或迟到复活，真实App正常退出和运行期SIGKILL后观察owned PID/端口释放、独立CLI仍可用，不仅发信号/删行，不承诺内核硬截止。
- [x] 现有唯一工程/test target的受影响Swift回归和真实Go集成及开发App构建有实际命令/结果/skip，本轮原生表面/浏览器/PID/URL/资源释放有证据；Go无改动不重跑已知errcheck确认旧失败或宣称lint通过，契约变更先审批。
- [x] 随真实证据更新README/架构及受切换影响旧说明，清理本次临时文件和确切owned资源；明确当前仅单目标手动入口、06批次/07Finder与08–13/最低系统/Intel/TCC/卷/候选/内核/Chrome后代边界，不以旧smoke冒充本轮结果。
- [x] 全部本票验收及独立只读Standards/Spec审阅通过且无未解决范围裁决后才关闭05；3.1/3.2/3.4仅部分贡献，保持完整checkbox未勾选，2.3及其他任务不变。不得自动提交/推送、发布实施06/07、同步主规格、归档或正式发布。

## Change-wide coverage reference

完整26任务、30 requirement、68 scenario及原大票全部剩余范围见 [coverage-plan.md](../coverage-plan.md)。01–05 已发布且完成（05 于 2026-10-04 关闭）；06/07 已发布、open（状态以 coverage-plan 当前记录为准）；08–13 未发布规划。本票收敛不减少变更总范围、不重写已关闭票或批准artifacts，不自动授予实施许可。

## Closure record

- 关闭：2026-10-04，状态 `done`；12 项验收全部按实际证据通过。3.1/3.2/3.4 仅部分贡献，**保持未勾选**（须 05+07 / 05+06+07 共同验证完整描述）；2.3 未触碰；`tasks.md` 勾选状态不变。
- 变更面：新增 `macos/GoGrip/PreviewAppModel.swift`（单目标打开→生产协调器→仅对已验证 running 会话请求浏览器、操作失败报告、一次性退出准入）、`macos/GoGrip/Services/WorkspaceBrowserOpening.swift`（NSWorkspace seam）、`macos/GoGripTests/PreviewAppModelTests.swift`（7 项宿主回归）；重写 `GoGripApp.swift`（AppKit 入口，无 SwiftUI Scene/空 Settings）、`AppDelegate.swift`（强持有唯一协调器、status item/popover、延后退出）、`Views/PopoverView.swift`、`Info.plist`（LSUIElement）；`PreviewSessionCoordinator.swift` 新增 `browserFailure` 会话事实、`stopAcceptingNewOpens()`、可注入 `prepare`、`.terminated` 前准入复查与 `willWaitForStop` 测试确认点；`project.pbxproj` 移除 FinderSync target/嵌入/依赖与旧文件、加入新文件、app target 改 ad-hoc 签名；`macos/Scripts/build-go-grip.sh` 将工具落到 `Contents/MacOS/go-grip` 并以独立 identifier `com.showgp.GoGrip.go-grip` ad-hoc 签名（构建期，无 runtime chmod）；`Makefile` 显式 derivedDataPath 并移除失效的 `dmg` target；删除 `ProcessManager.swift`、旧视图×4、`GoGripFinderSync/`、`project.yml`、`IntegrationTests.swift`、`ProcessManagerTests.swift`；更新 `README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`、`docs/macos-app-impl.md`、`docs/macos-app-plan.md`；`Storage.swift`/`HistoryEntry.swift`/`StorageTests.swift` 按票面保留待 09。
- Scoped TDD 证据：S1/S2 均以缺少 `BrowserOpening`/`prepare` 参数的编译失败为 red，实现后 6 项与 7 项绿；复审发现的“停在停止后被退出准入拦下的重开”竞态以 `.terminated` 前准入复查修复，并改用协调器 stop-wait 确认点（生产 nil）消除调度假设：删除复查时该回归实测 `factory.count` 2≠1（红），恢复后 68/0。
- 自动检查：`make macos-test` **68 tests / 0 failures**（ManagedProtocolTests 17、ManagedProcessTests 16、ManagedDiagnosticsTailTests 2、ManagedProcessGoIntegrationTests 7、PreviewSessionCoordinatorTests 14、TargetPreparationTests 5、PreviewAppModelTests 7）；Release `make macos` 与 Debug `xcodebuild build` **BUILD SUCCEEDED**；`codesign --verify --deep --strict` 宿主与工具通过（工具 identifier `com.showgp.GoGrip.go-grip`、宿主 `com.showgp.GoGrip`，均双架构）；工具仅在 `Contents/MacOS/go-grip`、无 `.appex`、`LSUIElement=true`；`open` 后 `lsappinfo` 为 `type="UIElement"`。Go 零改动，未重跑已知 errcheck。
- 真实开发 App smoke（人工交互 + 终端侧 PID/端口记录）：NSOpenPanel 单选普通目录、中文+空格 `.MD`、可访问空目录、主目录外 `/tmp` 目标；父/子/单文件独立与同名路径区分；原路径与符号链接别名复用同一 PID/URL；Copy URL/重开浏览器/关标签后服务继续；停 A 释放其 PID/端口、B 与独立 CLI 继续，Stop All 释放全部 owned；png 与不可读目录原生错误、无遗留会话；正常 Quit 与宿主 `kill -9` 后 owned 子进程与端口释放；运行中 `kill -9` owned 子进程 → 面板 Stopped + 原因、`Open in Browser` 不可点、无自动重启；独立 CLI（6419）全程 HTTP 200 且不出现在面板；两处面板布局缺陷（空列表弹窗定位、单行居中）修复后由用户复验。
- 浏览器拒绝原生表面：用带标注的临时一次性拒绝构建观察（弹窗、会话保持 running+URL、重开恢复并清除会话行错误；底部“最近一次失败”横幅按 2026-10-04 裁决保留），补丁已还原、无源码残留。
- 独立只读审阅：Standards 三轮（P1 迟到重开 + 5 项 advisory → 修复；确定性回归经 stop-wait 确认点修复后 0 阻塞/0 advisory）；Spec 两轮（浏览器拒绝原生证据 gap → 修复后 0 阻塞；0 未决裁决）。范围裁决 2 项均已有裁决：`release.yml` 候选 DMG 发布留给 12；失败横幅保留为历史。
- 保留的 advisory/限制：`willWaitForStop` 为测试确认点、生产为 nil；Finder Services/多选/最近记录/权限/登录项/本地化/候选包与 Intel/macOS 13/TCC/真实卷未验证；不管理 Chrome 后代；不对不可中断内核 I/O 承诺硬截止。
- 未覆盖/边界：06（多选批次）、07（Services）、08–13 及候选矩阵见 coverage-plan；未提交/推送、未同步主规格、未归档、未正式发布。
