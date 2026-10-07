# 07 — Finder Services 真实冷启动与统一入口

Status: done
Blocked by: 06-native-batch-preview
Covers OpenSpec tasks: 3.1 (Services-provider/registration contribution), 3.2 (Finder-entry/cross-entry contribution), 3.4 (Services-entry integration contribution)
Behavior source: ../specs/macos-finder-service/spec.md, ../specs/macos-preview-sessions/spec.md, ../specs/macos-menu-bar-app/spec.md, ../specs/macos-app-packaging/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

实际开发 App 提供单个 Finder Services 命令。宿主未运行、登录启动未开启时，用户在 Finder 选择目录或 Markdown 文件调用服务，系统启动确切的本轮 App，provider 接收全部 file URL 并立即交给 06 的统一生产批次；内置 managed Go 提供预览，默认浏览器显示对应内容。Finder/手动重开共享身份、PID/generation 和已验证实际 URL，多选确认/取消、部分失败一次汇总及浏览器恢复与手动入口一致。

这是“真实 Finder → 系统冷启动 App → 统一批次 → 内置 Go → 默认浏览器/原生恢复”的交付，不是 plist、selector、注册调用或模拟 pasteboard 转发完成。复用 05/06 的生产 App/批次/浏览器/退出控制，不复制其实现；首次引导与完整双语界面仍由 10，候选 App/DMG 与支持环境矩阵由 12/13。

05、06 范围已由用户确认，但均仍 open、未实施；06 是尚未满足的直接前置，05 是间接前置。本轮仅发布独立 07 供审阅，**不授权实施、编写测试或补做前置票**。实施前另行批准范围、scoped TDD seam、检查及实际 smoke。

`tasks.md` 是唯一进度账本。07 只交付 3.1、3.2、3.4 的剩余 Services 贡献：完整本票验收及独立 Standards/Spec 审阅完成后，还须核对 05 的 3.1 贡献、05/06 的 3.2/3.4 贡献及整项任务验证，才可勾选对应完整 checkbox。不能仅因三票已发布或票面确认而勾选；2.3 仍归 06，3.3 及任务4–7不变。

## Source mapping

批准 specs 是行为权威。下表保留原名，但只交付右列的开发 App/Services 贡献，不把完整历史、权限、语言或候选场景标为已验收。

| Requirement / scenarios | 07 贡献及后续边界 |
|---|---|
| macos-finder-service — Finder Services entry：`Open a selected folder from Finder`、`Invoke the service while the host is not running` | 实际系统可用入口、单目标和冷启动，不依赖登录项/先打开面板；系统控制菜单位置/启用。候选由13。 |
| macos-finder-service — Supported open targets：`Receive a Markdown file`、`Receive a directory outside the home folder`、`Receive an unsupported file` | 全部file URL进入06，目录/常规大小写不敏感.md按运行时类型处理，普通主目录外目标可用，不支持文件不转父目录。真实TCC/卷由08，候选由13。 |
| macos-finder-service — Batch deduplication and quantity confirmation：`Open exactly five distinct targets`、`Confirm six distinct targets`、`Cancel a large batch`、`Count aliases only once` | 真实Finder多选沿用06完整去重/5或6确认/取消规则，证明没有first截断或先启动再确认。候选由13，不重做06内部矩阵。 |
| macos-finder-service — Partial batch failures：`Some targets fail`、`All targets succeed` | 真实Finder混选继续成功目标、一次原生汇总且可查看，全成功无额外提示；不依赖系统UTI过滤掩盖混选。候选由13。 |
| macos-finder-service — Minimum necessary file authorization：`Access is denied`、`Open an ordinary accessible target` | 共享普通访问/失败报告，不假running/空目录、不默认要求FDA/辅助功能/自动化；实际授权主体/指导及受保护/挂载卷由08，候选由13。 |
| macos-preview-sessions — One session per normalized target：`Reopen a target through a symbolic link`、`Receive concurrent requests for the same target` | Finder和手动请求共用04事实源/06批次，实际跨入口/别名复用，再请求浏览器；沿用单飞，不新增入口字典。候选由13。 |
| macos-preview-sessions — Directory and single-file preview modes：`Preview nested documents`、`Preview a selected file`、`Keep an empty directory session`、`Add Markdown to a watched empty directory` | Finder打开目录递归、中文文件直接定位及空目录；沿用01/02监视，不改发现政策。候选完整文件事件矩阵由13。 |
| macos-preview-sessions — Verified startup and actual preview URL：`A nondefault port is used`、`Open a file with spaces and non-ASCII characters`、`Startup ends without valid readiness information` | 真实Finder成功使用03/04验证的实际URL/端口/文件路径；失败用共享报告，不猜地址或留下假会话。候选由13。 |
| macos-preview-sessions — Browser opening is separate from service lifetime：`Browser opening fails after service startup`、`Close the browser preview` | Services复用05/06浏览器结果/恢复，一次失败汇总但保留服务/URL；关闭浏览器不回收。完整面板由09，候选由13。 |
| macos-preview-sessions — Services end with their owning application：`Quit the application normally`、`Force termination of the host`、`Host exits during startup` | Services交接/异步处理遵守05/06退出准入及owned收尾；不绕过共享入口，已有所有权矩阵复用。候选由13。 |
| macos-menu-bar-app — Manual opening uses the same target contract：`Manually reopen a Finder-started target`、`Choose a new target manually` | 实际Finder→手动及手动→Finder重开同身份/服务/URL，不产生第二套手动或Services状态；最近入口由09，候选由13。 |
| macos-menu-bar-app — Opt-in launch at login：`First use without login startup`、`Change the login-start preference` | 仅证明无登录项的真实Finder冷启动；不注册/注销登录项或增加设置。系统登录项由11。 |
| macos-menu-bar-app — Action failures remain understandable and recoverable：`Open a stale recent target`、`Recover from browser opening failure`、`A service exits unexpectedly` | Services中的目标/启动/浏览器错误接入06一次原生报告及05真实状态/恢复，不依赖通知授权。历史/完整面板由09，候选由13。 |
| macos-app-packaging — Candidate DMG and Services-only Finder integration：`Produce an installable candidate`、`Use the new Finder integration` | 开发App接通Services，不重新引入05已删除的FinderSync/专用URL转发；本票不是DMG/候选安装验收。资源/引导由10，候选由12/13。 |

技术依据为 [design Decisions 1–4、6–7、9–10](../design.md#decisions)。[Apple provider文档](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/SysServices/Articles/providing.html) 明确请求可能在注册后、launch回调退出前到达；[Services属性文档](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/SysServices/Articles/properties.html) 定义声明/菜单匹配及本地化。它们支持已批准技术路线，不替代本变更的specs或实际系统证明。

## Scope and implementation handoff

当前 `AppDelegate.swift` 仍持有旧ProcessManager，selector只取urls.first，另有gogrip URL转发；`Info.plist` 无NSServices。它们不是可复用的可靠服务。05负责旧链路整体删除，06负责完整批次；实施07前重新读实际完成的05/06，不把本票描述当作已存在API，不重复清理或提前改动其已确认范围。

- **注册时序**：在05的真实AppKit根中强持有唯一协调器及06批次处理所需依赖，完全就绪后通过 `NSApp.servicesProvider` 注册一个provider；provider可以由现有根担任，不强制新包装类。不能等待第一次popover展示或把注册成功当作菜单可见。保留LSUIElement、bundle ID `com.showgp.GoGrip` 与既有App生命周期，不另建窗口/运行字典。
- **唯一声明**：主App的NSServices只有一项，`NSMessage = openWithGoGrip`、`NSPortName = GoGrip`、`NSMenuItem.default = Open with GoGrip`、`NSSendTypes = [public.file-url]`、`NSRequiredContext.NSApplicationIdentifier = com.apple.finder`。provider暴露 `openWithGoGrip:userData:error:`；不设返回类型/默认快捷键、不写回pasteboard、不加文本/shell/旧sender兼容解析，不用仅Markdown UTI限制阻止混合输入。10再以精确默认键配置en/zh-Hans的ServicesMenu及全套本地化，本票不提前实现引导/语言设置。
- **同步取得输入、异步处理**：selector用 `readObjects(forClasses:options:)` 读取全部NSURL，设置 `.urlReadingFileURLsOnly`，在返回前取得本次URL值，立即交同一生产批次后返回。文件元数据、数量确认、启动/HTTP及最终报告由06异步执行；不在selector中重做分类/逐项启动或等待整个批次。error指针只在返回前用于无法读取请求等即时错误；不得捕获到Task、用于稍后的目标/启动/浏览器错误或把pasteboard留到异步阶段重新读取。不新增超时/重试/入站队列框架。
- **统一事实/恢复/退出**：Services只增加输入入口，不复制06去重/确认/一次报告或05浏览器/退出行为；手动入口保持同路。原生操作报告与真实阶段分开，浏览器失败保留服务/实际URL及重开/复制，目标失败不猜端口/转父目录；Services请求不能绕过退出准入。provider工作由强持有根及共享协调器管理，不依赖面板可见，不增加自动恢复/重启。
- **系统发现与开发验证**：沿用05实际内置工具/本地ad-hoc签名及唯一工程。为真实Finder验收使用系统可发现的本轮开发App，记录确切bundle路径/代码身份及系统实际启动的可执行路径，识别同bundle ID旧副本，不能起错包后宣称本轮通过。可按批准design调用NSUpdateDynamicServices请求重扫；不自动改服务偏好、启用登录项、重启Finder、重置全局索引或覆盖/删除用户已安装包。需要安装替换或调整用户环境时先取得相应许可；菜单不出现先记录/排查实际系统状态，不能改走旧扩展、终端或selector直接调用冒充Finder。

## Proposed scoped TDD and real verification

发布不授权先写测试。新增永久测试只提出一项有独立用户影响的时序seam：**注册后立即收到首个服务请求、面板尚未首次展示时，实际生产根已可用，请求不丢失/崩溃并产生可管理的正确会话结果**。使用最小受控注册时序和既有启动seam，断言目标/会话结果，而非provider赋值、selector存在、参数转发或mock echo；若只能做声明/wiring断言则不新增该永久测试，由真实/临时smoke承担。实施前另行批准此seam。复用03–06的身份/单飞/批次/浏览器/准入回归，不重做内部矩阵，不新增plist/本地化字符串/签名拷贝/源文本/默认值或非空测试。

本票实施后必须记录以下本轮证据，不能引用旧04/05/06结果替代：

1. **真实Finder冷启动**：在用户已启用服务、未开启登录启动且宿主确实未运行的测试状态，从Finder选择普通目录及中文/空格 `.MD` 文件调用服务。观察系统启动本轮确切开发App，默认浏览器显示目录递归或所选单文件内容/实际URL，空目录也有可用会话；无需先弹面板、启用旧扩展、执行终端打开或运行时Go/PATH/源码回退。安装/索引准备可以先进行，但随后必须真正退出宿主再做冷启动；不擅自取消用户已有登录注册或启用其服务。
2. **真实跨入口及多选**：Finder启动目标后通过NSOpenPanel重开原路径/符号链接，手动启动另一目标后从Finder重开，记录相同有效身份、PID/generation/实际URL及正确内容，并再次请求默认浏览器。Finder实际多选5项直接、6项先确认、取消无本批child/浏览器且原会话/HTTP内容继续、别名去重为5无需数量确认；观察全部目标而非first。若系统选择交接规范化了别名，记录实际URL交接并用同一provider/生产批次的真实Go临时smoke补充6路径→5机制，不能把补充当作已观察的Finder选择行为。
3. **真实混选/失败恢复**：Finder选择有效目标与不支持文件及实际无法打开目标，观察成功保留/后项继续，一次原生报告含目标和可获得原因、面板可查看，全成功无额外提示。复制URL定位原内容，关闭浏览器服务继续。浏览器拒绝可用受控结果观察实际原生表面，须标明来源；正常默认浏览器必须实际观察页面。共享错误不能在selector返回后写其error指针，不把NSWorkspace true或mock failure当成页面/真实访问失败证明。
4. **交接与生命周期补充**：临时smoke用真实pasteboard、生产provider和同一批次验证输入在selector返回前取得、返回不等待数量确认/启动/最终报告，即时不可读请求仅在返回前报错；不新建永久纯转发测试。尽早请求/退出恢复可复用上述seam及05/06准入覆盖；实际Finder请求建立的会话通过面板停止/退出后owned PID/端口释放，独立CLI仍返回其原内容，不仅删行或发信号，不重做03所有权继承/4秒矩阵。

运行受影响Swift回归、现有真实Go集成及实际开发App构建，记录命令/结果/skip；Go无改动不为确认旧errcheck重跑lint或宣称修复。系统菜单、原生操作或默认浏览器环境缺失时保持对应验收未完成，说明缺少的实际观察；不能用单元测试、NSPerformService/直接selector调用或临时Foundation拥有者替代真实Finder冷启动。随证据更新 `README.md` 当前Finder/手动入口、系统启用/菜单检查、跨入口/多选及错误恢复操作和 `docs/ARCHITECTURE.md` 注册时序/交接/权限与候选边界，清理本次临时文件及确切owned资源。

## Prerequisites and non-goals

唯一直接blocker为 [06](06-native-batch-preview.md)，其真实批次及依赖 [05](05-native-finder-preview.md) 的开发App/浏览器/退出必须先完整实施并通过独立审阅。01–04已done，05/06只有范围确认；本票发布不解除此前置或授予实施权限。当前发布状态及历史编号以 [覆盖地图](../coverage-plan.md) 为准，不重写已有票。

不做3.3完整TCC/实际授权主体/受保护目录/外接与网络卷gate（08），不做新历史/完整面板/首次引导/本地化/登录项/universal/archive/DMG/CI/候选支持环境（09–13）；不要求菜单第一层/目录空白处或宣称系统必然显示。普通目标实际访问失败仍须明确报告；若证据否定内置工具路线则保留阻塞并请求设计裁决，不隐藏到08、不默认授予FDA/辅助功能/自动化、不detach或预埋helper。无新Go契约/发现政策、轮询/自动恢复/重启/重连/移动追踪、拖拽/标签页控制/高级参数、XPC/launchd、进程名/未知PID/Chrome后代清扫、旧扩展/URL兼容层。提交/推送、主规格同步、归档及正式签名公证/分发仍未授权。

## Acceptance

- [x] 实施前核对06及其前置05已完整实施/验证/独立审阅，另行批准本票范围、首请求时序seam、检查/真实smoke及实施；仅为独立批准消费者回归记录red/green，复用03–06，无plist/wiring/字段复制/mock echo/文案/默认值/资源存在或仅isRunning测试。
- [x] 注册前05真实App根已强持有唯一协调器、06批次依赖和可处理请求的provider；通过NSApp.servicesProvider只注册一个对象，不等待首次面板展示。首请求时序seam或实际/临时smoke证明早请求不丢失/崩溃且有正确可管理会话，不以赋值/声明存在替代。
- [x] 实际产物具有批准的单个NSServices声明及精确selector/PortName/default菜单键/public.file-url/Finder上下文，读取全部file URL而非first，不写回、不解析文本/shell或旧sender，不以仅Markdown UTI菜单过滤掩盖混选，不保留旧扩展/专用URL兼容路径；构建/系统发现证据而非永久源文本测试。
- [x] selector返回前取得本次URL值并立即交06，返回不等待异步分类/确认/启动/浏览器/报告；error指针仅即时不可读请求在返回前使用，无异步捕获/迟到写入或异步重读pasteboard。生产交接临时smoke有结果，不新增纯转发测试/重试或调度框架。
- [x] 真实Finder在宿主未运行、登录启动未开启时启动本轮确切开发App，从普通目录/主目录外可访问目标、中文空格.MD及空目录得到对应默认浏览器内容/实际URL；记录系统启动的bundle/可执行路径、菜单及页面，不依赖先打开面板、旧扩展、终端或生产PATH/源码回退，不以签名/plist或直接selector调用冒充冷启动。
- [x] Finder→手动和手动→Finder实际重开共用同一身份/PID/generation/已验证URL，别名不重复服务且再次请求浏览器；真实Finder多选观察5直接、6先确认、取消无本批启动/重开/浏览器且原会话保持、别名去重，明确系统实际交接和补充smoke边界，不能以first截断或预先启动让取消失效。
- [x] 真实Finder混合有效/不支持/无法打开目标继续后项并保留成功，只用06一次原生汇总且可查看具体目标/原因，全成功无额外提示；浏览器拒绝保留真实running/URL及重开/复制恢复，关闭浏览器不收尾，失败不转父目录/猜端口/假running/孤儿服务，受控和实际系统证据分别标明。
- [x] Services进入共享退出准入/管理路径，不建立第二套事实/启动入口；实际Finder会话可通过面板复制/重开/停止/退出，owned PID/端口真实释放且独立CLI继续。遵守系统菜单/启用状态，不自动启用服务/登录项、改偏好/重置全局索引/重启Finder或覆盖用户安装；必要环境变更单独取得许可。
- [x] 受影响Swift回归、真实Go集成及开发App构建有实际命令/结果/skip，真实Finder/浏览器/原生表面与PID/URL/资源释放有本轮证据；更新README/架构的当前Services能力、启用/菜单检查、注册/交接及未验证边界，清理临时文件/确切owned资源，不改旧50项用户数据、Go/独立CLI政策，不重跑旧errcheck确认已知失败。
- [x] 整票验收及独立只读Standards/Spec审阅通过且无未决范围裁决后才关闭07；按05+07的完整3.1和05+06+07的完整3.2/3.4贡献及整项验证决定勾选，不凭发布/确认自动勾选，2.3、3.3及其他任务不变。不自动发布/实施08、提交/推送、同步主规格、归档或正式分发。

## Change-wide coverage reference

完整26任务、30 requirement、68 scenario见 [coverage-plan.md](../coverage-plan.md)。01–04 done，05/06 open且范围已确认但未实施；本轮仅新增07，open待审阅、实施被06及其05前置阻塞；08–13未发布。现有票面的“本轮/未发布/待审阅”保持历史记录，以覆盖地图中的当前状态为准；开发App真实Finder证明不是12/13候选包、完整双语或跨架构/最低系统运行证明。

## Closure record

- 关闭：2026-10-05，状态 `done`；10 项验收全部按实际证据通过（上方发布时语句“05/06 open、未实施/不授权实施”等为发布当时记录，已由 05 于 2026-10-04、06 于 2026-10-05 关闭取代；`coverage-plan.md` 的发布时快照按 05/06 先例留待下一次发布轮更新）。实施批准依据：用户在实施前显式批准本票范围、scoped TDD seam（方案 A：provider 输入 seam 永久测试）与环境准备（替换 `/Applications` 副本、注销同 bundle ID 旧副本，不删除用户数据、不改服务偏好/登录项）；Spec 轴 scope decision（冷启动失败提示时机）由用户裁决取方案 A 并实施。`tasks.md`：本票贡献使 **3.1**（05+07）、**3.2**（05+06+07）、**3.4**（05+06+07）的完整描述具备整合验证，三处 checkbox 已勾选；2.3 归 06 已勾选；3.3 及其余任务不变。
- 变更面：新增 `macos/GoGrip/FinderServiceProvider.swift`（唯一 Services provider：`readObjects(forClasses:options:)` 同步取齐全部 file URL、返回前立即交同一 `openBatch`、`error` 仅用于返回前即时错误）与 `macos/GoGripTests/FinderServiceProviderTests.swift`（2 项 provider 输入 seam 回归）；修改 `macos/GoGrip/AppDelegate.swift`（协调器/批次依赖就绪后注册唯一 provider；根级失败提示订阅与 `presentFailureReport`）、`macos/GoGrip/Views/PopoverView.swift`（移除面板本地 `.alert`/状态，保留报告清单）、`macos/GoGrip/Info.plist`（唯一 NSServices 声明）、`macos/GoGrip.xcodeproj/project.pbxproj`（target 成员）、`README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`。无 Go 改动。
- Scoped TDD 证据（批准 seam）：以 `Cannot find 'FinderServiceProvider' in scope` 编译失败为 red；实现后 2 项绿（真实 NSPasteboard 三个目标——目录、`说明 文章.MD`、目录——经生产 provider 全部成为 running 会话且 displayPath 集合匹配、3 次浏览器请求、无报告；selector 返回即清空 pasteboard 仍不丢请求；空请求返回前写 `error` 且 0 会话/0 child/0 浏览器请求）。变异检验：临时取 `urls.first` 或改为异步重读 pasteboard 均实测失败后还原，证明回归非空断言。
- 自动检查（最终树）：`make macos-test` **80 tests / 0 failures**（基线 78 + 本票 2）；`make macos` Release `BUILD SUCCEEDED`、`codesign --verify --deep --strict` 通过；产物 `Info.plist` 唯一 NSServices（`openWithGoGrip`/`GoGrip`/`Open with GoGrip`/`public.file-url`/`com.apple.finder`）、工具仅在 `Contents/MacOS/go-grip`（identifier `com.showgp.GoGrip.go-grip`）、无 `.appex`、`LSUIElement=true`；`/Applications/GoGrip.app` 为唯一已注册副本、最终 CDHash `02d35053…`。Go 零改动，未重跑已知 errcheck。
- 真实 smoke（用户操作 + 终端侧 PID/端口/HTTP 记录，2026-10-05；宿主未运行、无登录项）：Finder 服务调用由系统冷启动本轮确切构建（宿主 44058 由 launchd 启动，PPID=1）；普通目录（递归嵌套文档）、中文空格 `.MD`、空目录实际内容正确；面板重开与符号链接 `alias-普通目录` 复用同一会话 44060/51434，手动新开 `t3` 后 Finder 重开仍只有 44262 一个 child；真实多选 `alias-t1、t1…t5`（6 路径）无数量确认且 `t1` 唯一 child、`t1…t6` 先确认 6 后 `t6` 新建、再次多选取消无副作用；混选 `photo.png`+`无法访问目录`+有效目录继续成功目标并一次原生汇总（两个失败目标均无 child）；受控（带标记、已还原）浏览器拒绝轮观察到原生弹窗与面板行内错误、`Open in Browser` 恢复并清除、关闭标签不收尾、Stop 只释放目标、Quit 释放全部 owned PID/端口；独立 CLI 全程不受影响。冷启动整批失败提示经方案 A 修复后由用户复验：`photo.png` 冷启动立即一次提示、开面板无第二弹；面板内 `无法访问目录` 触发一次提示、重开面板报告仍在。细节见 `docs/ARCHITECTURE.md` 的 07 记录。
- 独立只读审阅：首轮 Standards（**0 阻塞**、2 advisory：README 混选说明、注册时序仅由真实/临时 smoke 覆盖 → 均已修复；1 scope decision——实施批准，已由用户实施前显式批准解决并记录）与首轮 Spec（**0 阻塞**、4 advisory、1 scope decision——冷启动失败提示时机，用户裁决方案 A 并实施，其余 advisory 已修复或记录）；增量复审 Standards（**0 阻塞**、4 P3 advisory：GCD→`Task { @MainActor in }`、CDHash 取代关系、模态未串行化记录、关闭记录承接 → 均落实）与增量 Spec（**0 阻塞**、4 advisory：提示格式措辞与任务 10 的本地化范围、单项报告语义、面板交互观察、票面历史语句 → 分别改写记录/承接、按已声明语义接受、由用户复验观察、本记录说明）。审阅未改文件、未重跑检查。
- 保留的 advisory/限制：根提示与注册时序没有永久回归（由真实 smoke 观察；`make macos-test` 不构成首请求时序或原生表面的证明）；两个应用级模态（批次数量确认与根失败提示）未串行化，未观察到真实嵌套触发；任务 10 的“错误”本地化范围需覆盖 `presentFailureReport` 的 `GoGrip`/`OK` 与正文格式；单项报告（重开浏览器失败、退出未确认）由根呈现为已声明语义但未单独观察表面；候选包/universal/正式签名公证、TCC/受保护与卷访问（08）、完整面板/历史（09）、引导/本地化（10）、登录项（11）与 Intel/macOS 13 运行矩阵（12/13）仍未验证；不管理 Chrome 后代。
- 未覆盖/边界：未提交/推送、未同步主规格、未归档、未正式发布；未改动 Go/独立 CLI/发现政策；未读写或删除用户旧 50 项数据。

## Reopened 2026-10-06 — 共享原生提示呈现修正（已批准）

背景：09 验收实测首次触发本票已记录的残留“两个应用级模态（批次数量确认与根失败提示）未串行化”：根失败提示（`NSAlert.runModal`）显示期间到达的真实 Finder Services 请求在 pasteboard 交换超时（`Application com.showgp.GoGrip didn't return pasteboard data in time`，`NSPerformService` 返回 false）且请求丢失；随后出现 OK/Return 无法关闭的叠加 AXDialog，实例需强制重启。该缺陷违反 `macos-finder-service / Finder Services entry`（服务请求 SHALL 被处理）与 `Partial batch failures`（一次原生汇总、面板可查看、不逐项弹出）及 `macos-menu-bar-app / Action failures remain understandable and recoverable`，不能仅作限制保留。

修复范围（用户 2026-10-06 批准方案 A；含数量确认同修）：不改 Services provider/批次/会话事实/退出准入/最近记录；仅将应用级原生提示改为非模态、单槽串行呈现——新增单槽 presenter（非模态 `NSAlert` 窗口展示，主 run loop 保持默认 mode，提示显示期间 Services 可继续被处理）；失败报告与数量确认共用一槽，同一时刻仅一个提示，失败报告按到达顺序各展示一次；数量确认经 async 等待用户响应（取消不启动本批目标/不请求浏览器/不动既有会话）；`BatchQuantityConfirming.confirmOpening` 改 async；删除 `NativeBatchQuantityConfirmation`（不留 shim）。文案/按钮/汇总内容不变（本地化归 10）。浏览器失败、最近记录、退出准入与运行事实语义不变。环境批准：替换 `/Applications` 安装前先保存并核对当前回退副本（不删唯一回退副本）；专用临时目标与真实 Finder/AX 调用服务；必要时终止本 App 的实例；不改 Services 偏好/TCC/登录项/用户数据，不按进程名清扫，不实施 10/候选包，不提交/同步/归档。

验收（勾选以最终证据为准）：

- [x] 修正后，提示显示期间真实 Finder Services 有效目标请求端到端完成（会话/PID/端口/HTTP/浏览器），无 pasteboard 超时、不丢失请求
- [x] 提示可关闭（OK 与 Return），且不再出现叠加/不可关闭 AXDialog；应用保持可用
- [x] 每失败批次仍恰好一次原生汇总（目标+可得原因）且面板保留可查看报告；全成功不增提示；提示期间不再叠加（提示汇总、非叠加与全成功不增提示已实测；面板报告列表已在解锁后补测：提示关闭后仍保留同一条失败报告）
- [x] 数量确认显示期间 Services 请求仍被处理；取消不启动本批目标/不请求浏览器/不动既有会话；5/6 边界与别名去重不变（别名去重与 5/6 边界沿用 06/07 已观察路径，本轮未重演）
- [x] 冷启动（宿主未运行、面板从未展示）首次请求与提示仍有效；唯一 provider、selector 同步取齐全部 file URL、统一批次不变
- [x] 浏览器内容、Copy/重开/单停/全部停止/Quit 的 owned PID/端口真实释放与独立 CLI 隔离不变（Copy 当轮未重演；其路径未改动）
- [x] scoped TDD：presenter 单槽策略回归（一次一个、按序、响应路由）以编译失败为 red 后转绿；数量确认 seam async 适配后既有套件全绿；`make macos` 构建与签名通过
- [x] 文档与验证记录更新（保留历史观察，标明实际 native/受控结果与剩余限制）；独立 Standards/Spec 复审通过且无未决裁决（首轮与增量复审均 0 阻塞、无未决裁决）

边界：不重跑已完成强杀矩阵与 Go errcheck（无 Go 改动）；不改 09 已交付的最近/面板行为；不把本地化提前到本修正。

### 关闭记录（2026-10-06）

- 关闭：2026-10-06，状态 `done`；重开段 8 项验收全部按实际证据通过。环境（用户批准范围内）：替换 `/Applications` 前备份并核对 09 构建（`9c4ddf9b…` → `~/Desktop/GoGrip-backups/GoGrip-09-backup.app`，保留）；专用临时目标与真实 Finder/AX 调用；必要时终止本 App 实例；未改 Services 偏好/TCC/登录项/用户数据，未按进程名清扫，未提交/推送/同步/归档/实施 10。最终开发构建 CDHash `7a2c78fa…`（`/Applications` 当前为此构建）。
- 变更面：新增 `macos/GoGrip/Services/AppAlertPresenter.swift`（单槽、非模态展示、失败报告与数量确认共用，展示前 `alert.layout()`）与 `macos/GoGripTests/AppAlertPresenterTests.swift`；修改 `PreviewAppModel.swift`（async 确认 seam）、`AppDelegate.swift`（共享槽注入）、`PreviewAppModelTests.swift`（`ControlledConfirmation` async＋运行行 recency 断言）、`project.pbxproj`、`README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`、本票；删除 `NativeBatchQuantityConfirmation.swift`（无 shim）。Go/renderer/Services provider 未改。
- 检查（最终树）：`make macos-test` **90 tests / 0 failures**（87 → 90；修正后两次全量复跑一致）；`make macos` Release 构建 + `codesign --verify --deep --strict` 通过（`7a2c78fa…`）；red＝`cannot find 'AppAlertPresenter' in scope`；变异检验两处（槽 `current` 守卫、`openBrowser` recency 更新）按预期失败后还原。
- native（实际观察，证据截图在 `/tmp/gogrip-fix-smoke/evidence/`）：冷启动提示（面板未展示）；**提示显示期间**真实 Finder 请求端到端完成（`alpha` 生成 65659/58601、`beta` 生成 72816/52514，正文可辨），`log show` 无 pasteboard 超时；OK 与 Return 在最终构建上各关闭一次；两个失败批次提示一次一个、不叠加；数量确认显示期间真实请求完成、Cancel 零 child、Open 后 6 会话各有端口与内容；另一前台 App（Finder）时提示在其窗口之上可见且渲染完整；提示关闭后面板保留同一条失败报告；面板 Stop/Stop All/Quit 释放全部 owned child，独立 CLI（6419）存活。
- 复审：两轴互不合并，各两轮——首轮与增量复审 Standards **0 阻塞**（各 1 项 P3：断言归属、归档表述，均已按最终构建证据改正）＋ Spec **0 阻塞、0 advisory、0 裁决**两次；无未决裁决。
- 保留的 advisory/限制：槽策略回归经注入 seam 覆盖，不覆盖真实 AppKit 窗口关闭/事件路由（由 native 观察覆盖 OK/Return 与不叠加）；按钮 `target/action` 覆盖依赖 AppKit 未承诺的内部行为（复审指出、未实证失败）；Escape/窗口关闭按钮未单独观察；未重演 5/6、别名去重与 Copy URL（路径未改动，沿用 06/07/09 证据）；候选包/支持环境矩阵（12/13）、登录项/本地化（10/11）、网络卷（08 缺口）不在本票范围。
- 09 交叉引用：09 票 Closure record 所记“根失败提示为 `NSAlert.runModal`…Services 请求丢失、叠加无法关闭且需重启”的限制已由本修正消除；09 交付的最近/面板/清历史功能未改动。
- `tasks.md`：本修正属既有已勾选任务（2.3/3.1/3.2/3.4）的行为缺陷修复，无新增 checkbox；工作量/进度账本不变。

### 修正验证记录（2026-10-06）

- 实现与变更面：新增 `macos/GoGrip/Services/AppAlertPresenter.swift`（单槽、非模态展示、失败报告与数量确认共用；按钮 `target/action` 改指槽回调、`willClose` 视为关闭）、`macos/GoGripTests/AppAlertPresenterTests.swift`（3 项策略回归）；修改 `PreviewAppModel.swift`（`confirmOpening(count:) async`、`openBatch` await、默认参数与注释）、`AppDelegate.swift`（创建并注入共享槽，删除 `presentFailureReport`）、`PreviewAppModelTests.swift`（`ControlledConfirmation` 改 async；双目标用例新增运行行重开的前移断言）、`project.pbxproj`、`README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`；删除 `NativeBatchQuantityConfirmation.swift` 与工程引用（无 shim）。Go/renderer/Services provider 未改。
- red/green：`cannot find 'AppAlertPresenter' in scope` 编译失败（3 处）→ 聚焦 29 项全绿；变异检验：临时去掉槽的 `current` 守卫使 2 项策略回归失败、临时去掉 `openBrowser` 的 recency 更新使新断言失败，均已还原。
- 自动检查（最终树）：`make macos-test` **90 tests / 0 failures**（87 → 90）；`make macos` Release 构建成功、`codesign --verify --deep --strict` 通过（CDHash `eead0216…`）。
- 环境（用户批准范围内）：替换 `/Applications/GoGrip.app` 前备份 09 构建并核对一致（`9c4ddf9b…` → `~/Desktop/GoGrip-backups/GoGrip-09-backup.app`，保留）；smoke 专用目标在 `/tmp/gogrip-fix-smoke/`（证据截图 `evidence/01–06`）；smoke 期间写入的最近键已删除（其内容仅为本机 smoke 目标；旧 `go-grip-history` 未动）；App 实例与独立 CLI 已停止、端口释放；未改 Services 偏好/TCC/登录项，未按进程名清扫。
- native 实测（2026-10-06，开发构建 `eead0216…`）：冷启动（宿主未运行）对 `photo.png` 的服务请求立即出根提示且面板从未展示；**提示显示期间**真实 Finder 对 `alpha` 的服务请求端到端完成（child 65659、`127.0.0.1:58601`、正文 `alpha fix-smoke unique`、Chrome 打开 `…/alpha-notes.md`），`log show` 无 `didn't return pasteboard data in time`；OK 与 Return 各成功关闭一次；两个失败批次的提示一次只出现一个、第二个在前一个关闭后出现且可关闭（无叠加/不可关闭对话框）；数量确认显示期间真实 Finder 请求 `iota` 完成（child 65793、`127.0.0.1:58764`、正文 `iota fix-smoke unique`），Cancel 后 6 个批目标零 child，重新确认后 Open 使 6 个会话各自有端口与内容；面板 Stop（`gamma` 进程/端口归零）、Stop All、Quit 真实释放全部 owned child，独立 CLI（6419）全程存活。
- 补测与渲染修正（2026-10-06 解锁后）：① 另一前台 App（Finder）时模型化提示在其窗口之上可见且渲染完整（`evidence/10-alert-layout-fixed.png`、`11-alert-front-full.png`）；② 提示 OK 关闭后面板在最近列表下方保留同一条失败报告（红色原因行，`evidence/12-panel-report.png`），且提示期间真实 Finder 对 `beta` 的请求在最终构建上端到端完成（child 72816、`127.0.0.1:52514`）。补测中发现渲染缺陷并最小修正：绕过 `runModal`/`beginSheet` 后 `NSAlert` 的惰性布局不自动完成，首次补测的提示渲染异常（可见抑制复选框、空按钮、正文未排版，260×328）；`presentModelessly` 展示前调用 `alert.layout()` 后恢复常规单按钮提示（260×266，路径+原因+OK）。修正后复跑 `make macos-test` **90/0**、`make macos` 构建与 `codesign --verify --deep --strict` 通过，最终构建 CDHash `7a2c78fa…` 已安装（09 回退备份保留）；该构建上另复测 Return 关闭（提示出现后 `key code 36` → AXDialog 归零）。
- 复审：独立 Standards/Spec 首轮记录于下条；补测与增量复审通过后本票关闭（见关闭记录）。
- 复审（2026-10-06，独立只读两轴，互不合并）：首轮 Standards **0 阻塞**、1 P3 advisory（本轮新增的运行行 recency 断言需明确归属——已在测试注释标注其属任务 4.1 的“主动重开更新次序”语义并保留覆盖）；首轮 Spec **0 阻塞、0 advisory、0 裁决**。增量复审（补测后）Standards **0 阻塞**、1 P3 advisory（归档表述将 OK/Return 合并记为最终构建复测——已在最终构建补测 Return 并据实修正记录）；增量复审 Spec **0 阻塞、0 advisory、0 裁决**（确认面板报告保留与另一前台 App 时提示可见的覆盖已闭合、无新偏离）。无未决裁决；审阅未改文件、未重跑检查。
