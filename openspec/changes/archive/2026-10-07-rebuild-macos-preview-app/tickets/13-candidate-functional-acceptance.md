# 13 — 候选 App/DMG 跨组件真实功能与支持环境验收

Status: done
Blocked by: 12-universal-candidate
Covers OpenSpec tasks: 7.1, 7.2, 7.3
Behavior source: ../specs/macos-app-packaging/spec.md, ../specs/macos-finder-service/spec.md, ../specs/macos-menu-bar-app/spec.md, ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

对 12 交付的**同一最终候选 App/DMG**实际运行已落地的功能，完成 Finder→宿主冷启动→默认浏览器正确内容、目标身份/批次/面板/最近记录、真实停止与异常退出、访问/卷/watch 降级和支持环境的跨组件验收。记录每个场景的候选身份、环境、用户/native/browser 操作、实际 URL/内容、owned PID/端口及结果；区分静态检查、真实观察、受控辅助和仍缺证据，不把构建或自动测试通过等同于候选路径已经工作。

本票完整覆盖 **7.1、7.2、7.3**，是批准任务第 7 组的**验证-only 例外票**：不承担前置票遗漏的实现、行为测试或使用文档，不新增产品行为。发现组件不满足批准规格时回到对应票修正、复核并由 12 重新交付候选，再重验受影响路径；不边补实现边把当前候选判为通过。

**发布核对时状态**：01–07、09、10 为 done；10 关闭记录明确 5.1/5.2 完整交付并经独立 Standards/Spec 审阅，两项已勾。08 整票 open、真实网络共享待补证、3.3 未勾。11（5.3）、12（6.1–6.4）仍 open、对应任务未勾；用户确认继续发布 13 不等于其实施/验收完成。当前账本 **17/26**，7.1–7.3 全部未勾。本轮不重新执行或冒认 10 记录的开发构建/语言 smoke；其证据不提升为本票候选证明。

**真正实施前置**：依据 tasks 第 7 组“依赖第 6 组”，12 的完整交付和独立 Standards/Spec 审阅须先完成，并提供明确来源的最终候选、实际构建/签名/打包记录及使用文档；间接包含 11 的最终功能/语言资源。需要实际平台/卷环境和具体操作许可，不能凭票据发布或交叉编译推断已具备。Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷验收经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），本票不准备；08 网络缺口一并延期、保持未验证，不阻塞本票按调整后范围关闭。本票验证获准且可执行的路径，不以缺环境为由删除场景。

**发布不是实施许可**。本轮只新增本票并刷新 [完整覆盖地图](../coverage-plan.md)，不修改批准 artifacts、任务勾选、既有票、代码/测试或用户环境。执行仍须另行批准本票验收范围、现有套件、候选/native/browser 观察、必要的受控辅助手段和具体用户数据/安装/权限/卷操作；不授权提交、推送、打 tag、触发发布、主规格同步或归档。

**当前范围与关闭条件（2026-10-07 范围对齐，新增；依据已批准的 packaging spec、design、tasks）**：

- **延期（未验证、不属本票关闭条件）**：Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷验收——仍为后续 change 承接的未验证延期验收项，但不阻塞本票及 7.1–7.3 按调整后范围完成（见 [tasks.md](../tasks.md) 延期清单）；不得以静态检查、构建或替代环境结果顶替其实测证据，不得勾选为完成。
- **浏览器失败分支（已批准 A）**：以既有消费者行为回归的确定性结果验收；受控辅助不作为该分支的必需证据（仅在既有回归不足以覆盖且另行批准时使用，使用则须标注、还原并与最终候选证据分开）；候选级真实系统拒绝未触发须如实记录，不冒充系统级拒绝验证；该标准不替代正常浏览器成功路径的实际观察。
- **本机剩余验收（仍须完成）**：TCC 真实拒绝及授权主体、目标移动、服务关闭状态下再次查看 Help、简体中文界面与 Finder 中文服务名（含不支持语言回退）、登录项真实开启/关闭、最终候选外接卷只读访问。
- **历史与证据层级**：两份执行记录、最终候选身份与各级证据边界保留、不回写；本次为材料对齐，不表示任何未执行/缺环境项已通过；主规格同步、归档及正式分发另行批准。

## Source mapping

- [proposal.md](../proposal.md)：完整原生 Finder/菜单栏/预览功能与自包含候选；真实使用路径验收；所有权、状态、目标权限及最小设置边界；正式签名、公证、安装包验证和发布后置。
- [design.md](../design.md)：Decisions 1–10 的共享宿主/批次/身份/启动/所有权/状态/资源/候选契约；Verification Strategy 的真实 Go、所有权、原生表面与候选环境；Risks 的 Services/TCC、ad-hoc、内核阻塞与跨编译限制；Migration Plan 的 08 缺口、组件顺序、升级/回退及正式分发边界。
- [tasks.md](../tasks.md)：物理 checkbox **7.1、7.2、7.3** 的完整描述及验收；第 7 组只检查已落地组件和记录证据，不接收前面各组的遗漏。apply JSON 顺序 ID 不是 task 标签。

| Task / 行为源 | Requirement 与 scenario 对应 | 本票贡献 |
|---|---|---|
| 7.1 / [macos-finder-service](../specs/macos-finder-service/spec.md) | `Finder Services entry`（`Open a selected folder from Finder`、`Invoke the service while the host is not running`）；`Supported open targets`；`Batch deduplication and quantity confirmation`（5/6、取消、别名计数）；`Partial batch failures`（`Some targets fail`、`All targets succeed`） | 对最终候选的实际 Finder/手动/native/browser 路径验收，不重复实现 provider 或批次。首次/再次查看指导和系统服务启用边界消费已完成 10。 |
| 7.1 / [macos-preview-sessions](../specs/macos-preview-sessions/spec.md) | `One session per normalized target`、`Containment does not merge target identities`、`Directory and single-file preview modes`、`Verified startup and actual preview URL`、`Loopback-only application previews`、`Browser opening is separate from service lifetime` | 最终候选实际内容/身份/URL/回环/浏览器与服务生命周期；包括中文空格单文件、空目录新增 Markdown、跨入口/别名复用、父子/文件独立，不依赖 fake 服务或固定端口。 |
| 7.1/7.2 / [macos-menu-bar-app](../specs/macos-menu-bar-app/spec.md) | `Persistent native menu bar host`、`Visible sessions and explicit controls`、`Manual opening uses the same target contract`、`Bounded persistent recent targets`、`Clearing history does not stop sessions`、`Action failures remain understandable and recoverable` | 实际面板与最近消费者、复制/重开/清除/停止/退出、跨启动仅恢复记录、错误与真实会话一致；不把旧开发 UI 或纯 store 测试当候选表面。 |
| 7.2 / macos-preview-sessions、macos-finder-service | `Explicit stopping confirms service termination`（`Stop one of several sessions`、`Stop all application sessions`）；`Services end with their owning application`（正常退出、强制退出、启动期间退出）；`No automatic session restoration or restart`；`Unavailable paths are not empty directories`；`Visible hot-reload degradation`；`Independent CLI operation remains available`；`Minimum necessary file authorization` | 实际候选的 owned 清理/隔离、异常状态、拒绝/失效、已知 watch 降级与手动恢复（网络卷维度与候选级断卷经批准延期至后续 change，见 [tasks.md](../tasks.md) 延期清单），保留 08 的权限/网络证据归属，不新增自动恢复、权限绕过或 Chrome 清扫。 |
| 7.3 / macos-menu-bar-app、macos-finder-service | `Opt-in launch at login`（`First use without login startup`、`Change the login-start preference`）；`System-language localization`（`Use Simplified Chinese`、`Use English or an unsupported language`）；`First-use service guidance`（`Read first-use instructions`、`Revisit guidance for a disabled service`） | 最终候选消费 10/11 的完整功能和资源，实际 native/Finder 观察纳入证据汇总，不以资源存在、App-only Locale 或 SDK 调用返回替代。 |
| 7.1–7.3 / [macos-app-packaging](../specs/macos-app-packaging/spec.md) | `Universal self-contained application`（无开发工具、同一候选在支持架构运行）；`Candidate DMG and Services-only Finder integration`（可安装候选、新 Services 入口）；`Functional acceptance exercises actual user paths`（真实 Finder-to-browser、复用/停止、异常宿主退出）；`Explicitly deferred formal distribution`（候选评审、完成本阶段） | 消费 12 的实际产物，完成候选级真实路径与当前环境支持证明（Intel/macOS 13 实机维度经批准延期至后续 change，保持未验证，见 [tasks.md](../tasks.md) 延期清单）。静态架构/部署/签名检查只作其本身证据；不降低或提前完成正式分发标准。 |

四份 spec 的全部 **30 个 requirement、68 个 scenario 原名、共同贡献者与当前状态**见 [完整覆盖地图](../coverage-plan.md#every-requirement-and-scenario)。组件已验证的契约可引用确切来源，候选级规定路径必须有该候选实际证据，不能以引用组件单元测试关闭本票。

## Acceptance handoff and boundaries

### Final candidate identity and current prerequisites

实施前重读当前 CLI/context 和 01–12，确认没有未决 scope decision；记录 12 的 archive/App/DMG 来源与候选标识、实际运行 App 绝对路径、宿主/包内工具标识、架构/部署信息及签名检查。各环境使用同一交付候选；不能在 Apple Silicon 跑一个开发 build、Intel 跑另一个单架构工具后合称“同一候选”。更新候选后明确旧证据归属并重验受影响路径，不在结果中混合不同版本。

当前 [Makefile](../../../../Makefile) 的 `macos`/`macos-run` 是开发 build，`macos-test` 使用真实 Go 的逻辑测试夹具；它们不是已完成的 12 候选交付或本票 native 路径证明。当前 [README.md](../../../../README.md) 和 [docs/ARCHITECTURE.md](../../../../docs/ARCHITECTURE.md) 记录开发能力及限制，实施使用 12 最终更新的候选命令/说明，不猜旧 `macos/build/Release`、复用旧 PID/端口或把测试临时 App 当安装候选。

10 的关闭记录与 [docs/HANDOVER.md](../../../../docs/HANDOVER.md) 中当前开发版本/备份可作为环境准备依据，但须重新核对实际安装与运行来源。**10 的 S-A3 自动引导激活仅是非阻塞观察项**：在首次 Finder 冷启动到默认浏览器时观察自动引导/Help 的实际焦点顺序、页面与服务是否可用，记录用户影响；不把它升格为“浏览器永不失焦/必须聚焦同一标签页”的新增要求。若发现规格缺陷回到 10；若仅是体验取舍，按仓库 YAGNI 规则与用户讨论可信触发、影响、简单替代和成本后另行批准，不在本票自动改激活策略或重开已关闭票。

### 7.1 — Real candidate user paths

| 场景 | 必须观察的最终候选结果 | 证据与边界 |
|---|---|---|
| Finder 冷启动与自包含 | 在宿主未运行、登录启动非前提的状态，从 Finder 对专用目录实际调用 Services；系统启动该候选，包内真实 Go 服务与默认浏览器显示对应内容，无 Dock/终端使用前提。 | 配对记录候选路径、宿主/child PID、实际 URL/目标内容及可见原生/浏览器表面；不以 provider 调用、声明/重扫、HTTP 200 或进程创建替代 Finder 和实际页面。候选置于源码/构建目录之外，运行环境不依赖开发 Go/额外 CLI/PATH；不需卸载用户 Go，不把受控 PATH 隔离冒称机器未安装 Go。 |
| 目录、单文件、真正空目录 | 可访问普通目录递归呈现嵌套 Markdown，中文/空格 `.MD` 单文件 URL 定位所选文件，主目录之外可访问目标仍可用；真正空目录建立服务并显示空状态，新增 Markdown 后按已知监视状态观察内容更新。 | 用可区分的实际正文和 URL 证明目标，文件不改成父目录；空状态不能代替拒绝访问。正常 watcher 场景与明确 degraded 场景分别记录，不承诺网络卷始终自动刷新。 |
| 别名、跨入口与包含关系 | Finder、手动与最近入口重开同一规范化目标/符号链接复用同一真实服务并再次请求浏览器；并发同目标仍单飞。父目录、子目录、其中单文件各有独立会话/内容。 | 记录 normalized identity、展示路径、generation/PID/URL 与实际内容；不以相同名字或“行数没增加”代替身份，不承诺重用/聚焦同一浏览器标签页。 |
| 5/6 边界、去重与取消 | 5 个不同有效目标直接处理；6 个先以实际数量确认；6 条路径归一为 5 个目标不确认；不支持项不计有效数量。接受前无本批启动/浏览器副作用；取消保持既有会话、地址和服务。 | 从 Finder/手动共享入口实际操作并配对现有行为回归；不只证明确认按钮或 mock 回调，不能通过减少输入覆盖来避开边界。 |
| 部分失败与提示并发 | 有效目标、不支持文件和真实打不开目标混选时，后项仍处理、成功保留，一次原生汇总含各失败目标/可获得原因，面板可查看；全部成功不额外提示。提示显示期间到达的实际 Services 请求仍处理，关闭一个提示不叠出无法关闭的窗口。 | 观察原生提示、各实际会话/页面和失败报告；保持 07 已批准非模态单槽，FIFO 按 presenter 入槽顺序，不新加根报告发布顺序的全局保证；不依赖先打开 popover 才感知冷启动失败。 |
| 面板、浏览器与明确控制 | 同名不同路径可区分，实际阶段/不可访问/已知降级/原因可查看；Copy URL 定位原目标，主动 Open 复用服务；关闭 popover/浏览器不停止，单停/全部停止有效。浏览器系统拒绝时服务/URL保留，可主动重开或复制。 | Native 可见动作必须对应真实内容/资源。正常/未知访问或 watcher 状态可沿用已接受的安静表面，不以新正向标签/措辞判失败。浏览器失败分支按已批准 A 以既有消费者行为回归的确定性结果验收：受控辅助非必需，候选级真实系统拒绝未触发如实记录、不冒充系统级拒绝验证，且该标准不替代正常浏览器成功路径的实际观察。 |
| 最近记录与清空 | 候选最近存储仍以规范化身份至多20项、主动重开更新次序不重复；按批准场景核对21项淘汰与别名/较早目标重开。退出重启只恢复列表、不访问或启动旧目标；活动会话下清空后仍可 Open/Copy/Stop，之后不再打开目标而重启时列表为空。 | 结合现有隔离域回归和实际候选最近/重开/清除/跨启动表面，不新增文字/store wiring 测试；用户现有最近键先保存，旧 `go-grip-history` 不读/写/迁移/删除，不清整个 defaults 域。 |
| 指导、语言与登录状态 | 候选首次引导与 Help 可用，关闭服务时仍可按指导检查，App不自行启用；简中/英文/不支持语言下观察完整 native、错误/行动指导及实际 Finder 名称；登录选项按真实系统状态，仅用户开启/关闭，无登录启动仍能 Finder 冷启动。 | 系统/Finder 语言与 App-only Locale 分清；保留底层诊断。注册/拒绝/需批准不假成功，已有注册不擅自取消；实际用户操作需许可，不把签名或 SDK 调用当启用证据，不要求注销登录、重启系统或重开10/11实现。 |

### 7.2 — Actual failure, ownership and isolation

| 场景 | 必须观察的结果与证据 | 不能替代 / 限制 |
|---|---|---|
| 正常停止与退出 | 多会话下单停释放对应 owned PID/监听且其他会话仍提供各自内容；Stop All 与正常 Quit 释放本宿主全部 owned 服务。独立 CLI 前后保持自己的地址/内容，不属于面板或停止范围。 | 不只是删除行/发信号，不按进程名扫杀。单停的“其他 owned 保持”和 Quit/宿主强杀的“同宿主全部 owned 结束”分别核对，不能误把同宿主剩余会话存活当强杀成功。 |
| 运行中强制终止宿主 | 在该候选同时持有真实多个 Go child 时，终止确切宿主 PID；这些 child 和其 listener 消失，不等下次 App 启动清理，独立 CLI 保持可用。 | 记录杀前后的 child/PID/端口和时间，不以 `/bin/cat`、正常 termination callback 或其他 Foundation harness 代替候选；多 child 也用于真实复核 writer 未相互继承的所有权结果。 |
| ready 前强制终止宿主 | 先观察实际 owned Go child 已创建且尚未完成 verified startup，再终止候选宿主，确认该 child/监听清理，无新浏览器或假运行。 | 仅 UI starting、尚无 child 或已 ready 后再杀不能记为 ready 前证明。用获准慢初始化目标/专用环境取得实际窗口；无法建立或观察时保留缺口，不修改生产为 sleep/fake ready。记录已有有界退出机制及不可中断内核 I/O限制，不凭一次结果承诺精确硬截止。 |
| 服务意外退出与启动失败 | 在多个服务运行时终止一个确切 owned child；对应行不再 running、原因可查看、不自动重启，其他服务/CLI继续；用户可从最近主动重开。实际准备/启动失败无假running、猜测URL、6419 fallback或本次遗留 child。 | 用户显式停止不误报意外退出；不以点击Stop证明异常退出。必要受控故障只能补充对应表面观察，不能伪称真实系统/进程事件或掩盖缺失 startup 证据。 |
| 实际权限主体、允许与拒绝 | 在获准专用受保护目标走候选 native→包内 Go，观察当前候选真实系统授权主体；允许时正确页面，实际拒绝时原生目标/原因/适用指导，无假空目录/成功。普通允许目标不需统一 FDA/辅助功能/自动化。 | 不继承旧 ad-hoc 构建的授权成功结论；父子结构/签名/chmod/Terminal 可读不替代真实 TCC 允许/拒绝。测试工具权限不转成产品默认前提。若否定内置工具路线，停止相关集成并请求设计裁决，无 helper/detach/FDA 兜底。 |
| 删除、移动、专用卷不可访问 | 对实际已打开的专用目标执行获准删除/移动（候选级断卷经批准延期至后续 change，见 [tasks.md](../tasks.md) 延期清单），再访问原URL；明确原路径 unavailable，保留真实查看/停止入口，既有服务不跟踪新位置、不自动重启/重连、不误为空目录。 | 只操作可丢弃目标/自己的测试卷，不卸载用户工作卷；实际检测/请求触发与状态到达分别记录，不强加不存在的即时发现保证。恢复可访问路径后可按指导主动刷新/停止/重开，不承诺 watch 自动恢复。 |
| 外接与已挂载网络卷、已知 watch 降级 | 真实外接卷的目录/文件在允许范围内正确预览，真实不可访问/共享拒绝反馈明确；已挂载网络卷维度经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单）。实际预算/注册失败/覆盖不完整使 watcher degraded 时，可查看真实状态/原因、预览仍可用、浏览器手动刷新反映新内容。 | 不用普通目录/别名/命名模拟外接或网络卷；未出现降级不虚构已测降级，网络卷无可靠事件不当普通 watcher。保留无轮询/自动重连策略；网络卷证据经批准延期至后续 change（未验证、不阻塞本票），本票不自动关闭08/3.3。 |

对不易触发的浏览器系统拒绝，按已批准 A 验收：以现有消费者行为回归的确定性结果为准，受控辅助不作为必需证据，候选级真实系统拒绝未触发须如实记录、不冒充系统级拒绝验证，且不替代正常浏览器成功路径的实际观察。对特定 watcher 故障等其余不易触发项，优先现有消费者行为回归与可安全执行的实际场景；确需临时受控辅助时先明确批准、标注所控制边界及产物差异，不能把 mock/受控输出冒充真实系统拒绝、权限主体、断卷或候选所有权证明。辅助观察不改变最终交付候选，结束后还原/清理并从12最终产物复核正常路径，结果归属不同候选时不得混记。

### 7.3 — Supported environments and evidence summary

使用同一候选，记录每次运行的实际硬件、OS版本、执行架构/是否翻译、App/Go来源、native/browser结果与清理。最低系统和两种硬件运行是独立覆盖维度，可以由同一真实环境同时满足，**不新增每个OS小版本或四格全组合运行保证**。

| 覆盖维度 | 所需证据 | 不能替代 |
|---|---|---|
| Apple Silicon / macOS 13+ | 在实际 Apple Silicon 环境运行候选宿主和包内 Go，实际入口到正确内容、面板操作和 owned 清理符合规格。 | arm64 Mach-O 存在、构建成功或另一App/工具运行不够。 |
| Intel / macOS 13+（**经批准延期**） | 该维度经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），保持未验证、不阻塞 7.3 按调整后范围完成；后续实际运行须记录对应架构的预览、操作和收尾。 | Apple Silicon 的 Rosetta 运行可另记辅助结果，但不冒充 Intel 硬件验证；x86_64 slice/交叉编译不够。 |
| 最低 macOS 13（**经批准延期**） | 该维度经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），保持未验证、不阻塞 7.3 按调整后范围完成；后续实际运行须记录宿主/工具及原生集成的实际结果。 | `MACOSX_DEPLOYMENT_TARGET=13.0`、Mach-O最低版本或较新系统成功不证明最低系统运行。 |
| 当前及其他已测环境 | 分别记录实际版本/架构、已执行场景及结果，与静态架构/部署/签名检查和未测环境区分。 | 当前机器的一轮不能覆盖所有支持环境；本机剩余验收未完成时 7.3/13 保持未完成，延期维度不阻塞 7.3 按调整后范围完成。 |

执行相关现有 Go/Swift 行为套件，记录实际命令、结果、skip及与12候选源版本的对应；不引用旧 tests 数量代替当前检查，不为确认已知错误重跑无关lint，不把失败或skip改名为通过。组件行为套件与候选实际运行结果须一致，分歧回到所属组件处理。本票不新建永久集成测试/框架、源码/文案/资源存在/拷贝/YAML/wiring/mock echo断言或重复内部矩阵。

证据写入本票实际执行/关闭记录及现有相关架构/交接说明，区分**本轮真正观察、引用前置记录、静态检查、受控辅助、未执行/缺环境**；记录候选标识、场景/操作、期望、实际结果、截图/输出出处、owned资源收尾及未完成原因。必要使用说明修正回到其所属票，不以本票汇总接收12未完成文档。不凭假设生成PASS，不能以“未见失败”关闭缺环境的场景。

## Native/environment approval and failure routing

- 安装/替换候选前核对已有App、运行宿主、owned会话及同bundle ID副本，保存明确可用回退副本/状态。退出/强杀仅限本次确切宿主/owned PID，保留独立CLI；不删除用户App/数据或批量扫杀。
- 保护目标、外接/测试卷、权限拒绝/允许、删除/移动仅用获准专用环境，保存并按计划还原；网络共享与 Intel/macOS 13 环境经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），本票不准备，不尝试枚举/修改用户环境或提前声称已具备。
- 若三种系统/Finder语言、服务启用开关、登录注册或首次/最近键场景需要状态变更，事先逐项许可、记录原状态、按约定恢复；不清整个UserDefaults域、不删除旧历史、不重置全局TCC、不默认强制重启Finder/pbs。前置票许可和以前的环境还原记录不是本轮操作许可。
- 暂缺许可/可观察窗口时完成其他获准检查并明确阻塞哪项（Intel/macOS 13/网络卷/断卷已批准延期，不在阻塞之列）；相应 checkbox 保持未勾直至按调整后范围完成，不缩为普通目录/本机编译。发现可信缺陷报告现象、候选身份、来源requirement和所属01–12票；修正按单票审批/必要scoped TDD/独立审阅执行，候选重新交付后复验，不在13增加实现或防御抽象。
- 本票不修改发现/预算规则、不轮询、不追踪移动、不自动恢复服务/卷/watch、不增加设置中心/登录helper，不管理任意Chrome后代；不重建CI/打包系统、不取消CLI release或改变CLI网络政策。Developer ID、公证、正式干净安装和公开发布另行安排，后续标准保留。

## Acceptance

- [x] 12的6.1–6.4完整交付并独立Standards/Spec确认，最终候选和来源/签名/架构/部署/使用说明可核对；13范围、现有套件、真实观察、受控辅助及具体安装/数据/系统操作另获批准，未把发布当执行许可。
- [x] 7.1从最终交付包取出的同一候选实际完成Finder冷启动到默认浏览器正确内容，观察确切宿主/包内Go/目标URL，无源码、开发Go/PATH、旧扩展、Dock或终端使用前提，不以声明/进程/HTTP/测试代替真实表面。
- [x] 目录递归、中文/空格单文件、主目录外允许目标、真正空目录及新增Markdown按实际watch状态正确；别名/并发/跨入口复用和父/子/文件独立有真实identity/PID/URL/内容证据，无固定端口fallback或文件转父目录。
- [x] 5/6边界、6路径归一为5、取消无启动/浏览器副作用、混合有效/不支持/真实失败仍保留成功且一次汇总、全成功不提示实际观察通过；共享非模态提示期间实际Services请求不丢失，FIFO不扩张为全局报告顺序保证。
- [x] 面板同名路径/实际状态/失败/已知降级可辨，Copy/Open/单停/Stop All/正常Quit作用于真实目标；关闭UI/浏览器不停止，浏览器失败分支按已批准 A 以既有消费者行为回归的确定性结果验收（受控辅助非必需；候选级真实系统拒绝未触发如实记录、不冒充系统级拒绝验证），失败不抹掉running/URL及主动恢复入口，正常或未知正向状态不被新增措辞标准误判。
- [x] 最近至多20项/别名去重/recency、实际最近重开、跨启动仅恢复记录及清空不停止/跨启动为空与组件回归一致；陈旧路径有目标/原因/适用指导，用户原键按许可保存还原，旧历史未迁移或删除。
- [x] 7.2正常退出、运行中多owned强杀宿主、ready前已有真实Go child时终止宿主、一个child意外退出及实际启动失败有配对进程/端口/状态证据；owned资源释放、其他会话/CLI按对应规则隔离，无假running、自动重启/孤儿服务或Chrome后代清扫，未将正常退出或无child的starting当强杀证明。
- [x] 候选实际系统授权主体、受保护允许/拒绝（含真实 TCC 拒绝）、专用目标删除/移动、真实外接卷与已知watch降级/手动刷新有对应证据，失败不误为空目录，不绕过权限/统一要求FDA，不自动追踪/重连/轮询；网络卷与候选级断卷经批准延期至后续 change（未验证、不阻塞本票），08/3.3缺口不被代签或自动关闭。
- [x] 最终候选指导/Help、服务关闭检查、简中/英文/不支持语言的完整native及真实Finder名称、真实登录状态与用户开启/关闭行为已核对；无登录启动仍冷启动有效，底层诊断保留、不假成功或擅自取消已有注册；S-A3仅记录非阻塞观察，不新增焦点保证或自动修改10。
- [x] 7.3同一候选有实际Apple Silicon运行覆盖，记录真实OS/硬件/执行架构/对应场景/收尾；Intel与最低macOS13实机维度经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），保持未验证但不阻塞7.3按调整后范围完成；静态或Rosetta辅助与实机结果分开，不新增全小版本/全组合保证。
- [x] 相关现有Go/Swift行为套件有实际命令/结果/skip并与候选路径一致，无新永久wiring/文字/资源/拷贝/mockecho或重复内部矩阵测试；如使用受控辅助，须另行批准、标注、还原、清理并与最终候选证据分开；浏览器失败分支的受控辅助不是必需证据；环境/用户数据按批准计划恢复。
- [x] 全部本票按调整后范围必要场景（延期项记录为未验证，不阻塞关闭）及证据汇总完成并经独立Standards/Spec审阅、无未决范围后才关闭13并勾选7.1–7.3；08/3.3按各自调整后条件独立关闭，票据全部发布不等于整变更完成，未宣称正式签名/公证/干净安装/公开发布或同步/归档已获许可。

## Whole-change coverage checkpoint

[完整覆盖地图](../coverage-plan.md)逐项保留**26个OpenSpec checkbox、30个requirement、68个scenario**与全部贡献者。本轮新增最终规划票 **13-candidate-functional-acceptance**，完整映射 **7.1–7.3**，状态open、待审阅；至此既定 **01–13均已发布**，没有剩余未发布切片或未分配范围。

当前01–07、09、10 done；08 open/网络共享待补证；11、12、13 open，13的12实施前置尚未满足。账本**17/26**，3.3、5.3、6.1–6.4、7.1–7.3保持未勾。本轮只刷新地图中的当前事实，不重写既有票的历史发布快照、不修改批准artifacts或任务进度，也不授予实施、环境操作或正式分发许可。

注（2026-10-07 范围对齐，新增）：上段为发布时快照，保留不回写；当前范围、延期项与关闭条件见上方「当前范围与关闭条件」及 Acceptance 调整后文本；延期项（Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷验收）保持未验证、不阻塞本票关闭；计数以 [coverage-plan.md](../coverage-plan.md) 更新后为准。

## Execution record — 2026-10-06（本轮执行；票保持 open）

**前置与台账复核**：12 `done`，关闭记录含独立 Standards（0 阻塞/1 P3 advisory）/Spec（0 阻塞）审阅，6.1–6.4 已勾；本票 `Blocked by: 12` 已满足。发布时快照（11/12 open、账本 17/26、coverage-plan 当前事实段）已过期，实际为 11/12 done、账本 **22/26**（3.3、7.1–7.3 未勾）；按仓库存例只报告、不重写历史快照，不在本票内刷新地图。

**候选身份（本轮运行对象）**：`/Applications/GoGrip.app`（宿主 `com.showgp.GoGrip`，LSUIElement=true，LSMinimumSystemVersion=13.0，无 `.appex`，资源 en/zh-Hans），包内工具 sha256 `7426a27f…` 与 `macos/.build/candidate/GoGrip.dmg`（`91af28d6…`）、`GoGrip.app.zip`（`f5c821df…`）、`GoGrip.xcarchive` 中同一文件逐字节一致；`diff -r` 挂载 DMG 载荷与已装 App 完全相同，`codesign --verify --strict` 通过，两个 Mach-O 均 `x86_64 arm64`（工具 slice minos 12.0，宿主 13.0），无源码/PATH/Resources 回退。机器：Mac14,3（M2）、macOS 27.0.1（26A434）arm64、Xcode 27.0、Go 1.26.3；默认浏览器 Chrome。

**批准范围**：用户批准 7.1+7.2 可执行路径全跑（Apple Silicon/macOS 27）；允许 DMG/Services 操作、AX 驱动真实 Finder/面板、启动/强杀确切宿主 PID、浏览器检查、专用可丢弃目标与卷操作、全局 AppleLanguages 切换并还原、TCC 专用目标允许/拒绝（含 per-app 重置）、登录项真实开启/关闭并还原未注册、Finder 重启以读中文服务名；受控辅助逐项报批（本轮未使用）；确认 Intel/macOS 13/网络共享不可用（记录阻塞）、外接卷只读不卸载、并发脚本按“真实慢启动目标”建立 ready 前窗口；裁决“目录含不可读子树→整会话失败”记为**缺陷**，回所属票修正、12 重新交付后复验。

**现有套件（最终树，git 295311a）**：`go test -count=1 ./...` 全部包 ok；`make macos-test` **95 tests / 0 failures**。二者与候选实际行为无分歧。

**7.1 实际路径证据（同候选）**：Finder 冷启动（宿主未运行、登录项 Not set）：launchd 启动宿主 PID 98266（PPID 1，路径 `/Applications/GoGrip.app/Contents/MacOS/GoGrip`），owned child `--managed C436BDB1-… -r -- <目录>`（PID 98268，`127.0.0.1:59641` 回环），Chrome 实际页面 `…/子目录/deep/深.md` 含 `DEEP-MARKER`（截图 01）。首启自动引导（“Using GoGrip from Finder”全文）实际出现，Help 可再次打开；S-A3 观察：浏览器为实际前台页面，引导窗口同时存在且未抢焦点，无新增焦点保证。内容模式：递归嵌套、中文/空格 `.MD` 单文件（URL 定位该文件、`FILE-MARKER`）、主目录外 `/Users/Shared` 目标、真正空目录显示 “No Markdown files” 空状态后新增 `新增.md` 由真实浏览器页面热更新显示（Chrome AX 文本实测）、符号链接复用同一 child、同目标并发两请求仅 1 个 child、父/子/文件三个独立 child+端口且子会话以自身为根（`parent.md` 在其下 404）。批次：5 目标无确认直接 5 会话；6 目标出现 “Open 6 previews?”（Cancel：无新会话/无浏览器请求/既有会话不变；Open：+6 浏览器请求、t6 启动、既有复用）；6 路径含符号链接归一为 5 不确认；混合 `m1`+`unsupported.png`+`locked` 保留 m1 且一次汇总命名两失败及原因/指导、全成功无提示；报告显示期间到达的 Services 请求仍被处理（n1 建会话，仅 1 个窗口、无叠窗）。面板：同名 Alpha/Beta 以完整路径区分；Copy URL 剪贴板得到该会话真实 URL；Open in Browser 复用服务并 +1 标签；单停释放对应 child/端口且其他会话继续 200；Stop All 与 Quit 释放全部 owned（0 go-grip listeners，独立 CLI 200）；关闭预览标签与关闭 popover 后服务继续 200；失败/已停会话在面板保留原因（“exited unexpectedly (status 9)”“target-unavailable…”）。最近：存储实测按规范化 identity（9→10 项、无别名双条目；打开符号链接复用同一 child 32102 且仅 1 条记录）；21 项淘汰/别名/次序由既有 `RecentTargetsTests` 覆盖并与候选表面一致；最近 Open 建会话、再次 Open 复用同一 child 并再请求浏览器；陈旧路径 Open 得原生错误（目标+原因+指导）且不建会话；清空在有活动会话时只清列表（Open/Copy/Stop 与内容保留），重启后列表为空；有记录重启只恢复列表、不访问目标。指导/语言/登录：首启与 Help 英文实测；zh-Hans 真实面板（会话/最近目标/清空/打开/开启）与真实 Finder 菜单名 **用 GoGrip 打开**（Finder 重启后实测；随后还原 en-US 并复测 `Open with GoGrip`）；ja 回退英文；登录项 Not set → Turn On → 系统状态 Enabled（BTM 条目 Name GoGrip / Identifier 2.com.showgp.GoGrip / URL /Applications/GoGrip.app）→ Turn Off → Not set。

**7.2 失败/所有权/权限证据**：正常退出 2 会话 → 全部 owned 释放、CLI 200；运行中 SIGKILL 宿主（22 个 owned child）→ 全部消失、端口释放、不依赖下次启动、独立 CLI 未受影响（log.md 配对记录；复核补证带时间戳：host 32733 于 18:19:55.440 被杀、0.35s 后 children 全空且端口释放、独立 CLI pid 32827 仍 200）；ready 前终止（USB 慢目标 61k 目录、ready 延迟实测 1.6s）：先观察到 owned child（PID 31977，无监听、无面板 Running 行、无浏览器请求），随即 SIGKILL 宿主 → child 约 1.5s 内消失、无监听残留、无新浏览器、无假运行；外部杀死单个 child → 行 “Stopped / exited unexpectedly (status 9)”、不自动重启、其他会话继续、可从最近记录重开被杀的 `targets/batch/t1`（得到新 child 662）；复核补证（log.md「Re-verification」段，文本留存）：alpha child 32752 于 18:19:31.020 被外部 SIGKILL，幸存 beta child 32763 连续 3 次请求均 HTTP 200 且页面含 `REV-BETA`，面板 alpha=Stopped（原因同上）/beta=Running+URL；此前一次检查日志中的 `u2 … 000` 为未复现的检查产物，已由上述配对证据取代；启动失败（locked 目录、删除后目标、TCC 拒绝）均无假 running/猜测 URL/6419 fallback/遗留 child。TCC：真实系统提示 ““GoGrip.app” would like to access files in your Documents folder.”（UserNotificationCenter）；拒绝 → 原生错误 + 适用指导（非空目录、无会话）；per-app 重置后允许 → 受保护目标会话 Running 且 `TCC-MARKER` 由包内 child 提供（授权主体为 App）。删除/移动：删除运行中目标 → 错误页（非空状态）、面板保留原路径与 unavailable 原因；外接卷上移动 → 404 “not found” 页、原路径 unavailable；路径恢复后同 child 200（无追踪/重连）。外接卷：真实 USB 卷（disk5s1，Protocol USB，2TB）目录/嵌套文件正常预览；断卷未执行（卷不可卸载，记录阻塞）。降级：真实预算耗尽（软/硬 fd=1200 → 176 budget），两个独立会话分别记录：`targets/degraded`（402 目录）——包内工具 stderr 与候选面板同报 `402 of 402 directories … exceed the 176 entry watch budget`（工具侧 degraded 文本留存于 `evidence/probe-degraded.txt`：同类条件复核运行 PROBE-13D 记录 `reload.state=degraded`、201/201 超出同一 176 budget 的完整文本）；`targets/degraded2`（180 目录，root cost 180>176）——候选面板报 “Hot reload degraded: 180 of 180 directories … 176 entry watch budget”（面板文本即工具 reload-status 事件转述）；两者预览仍可用，在未监视根新增 `NEW-LATE.md` 不自动刷新，手动浏览器刷新后出现（AX + 截图 18/19）。独立 CLI（源码构建，6419）在各停止/强杀/降级场景前后持续 200（正面文本：正常退出段 `cli http=200`、强杀段 `cli still up? http=200 pid=99327`、复核段 `cli serves repo root: http=200`；复核段另有三条 `cli http=404`，是以 CLI 服务根之外的固定路径请求的预期否定项，已在 log.md 注明）。

**未执行/阻塞（保持未完成，不代签）**：Intel 实机运行、macOS 13 实机运行（7.3）；已挂载网络共享（08/3.3 缺口）；外接卷断卷（不可卸载）；候选级“系统拒绝浏览器打开”未触发（消费者回归覆盖该分支，未使用受控辅助）；服务关闭状态下的 Help 复核未做（未申请切换服务偏好）。**发现（已按用户裁决标记为缺陷，待回所属票）**：可访问目录只要含一个不可读子树（`chmod 000` 子目录复现；同为 TCC 受限子目录的路径）即整体会话以 `target-unavailable: resolve initial preview path: … permission denied` 失败，而非“可用预览 + 降级”；对照 `macos-preview-sessions` 的 `Directory and single-file preview modes`/`Preview nested documents`（可访问目录应建立递归预览会话并服务其可读内容）与 `Visible hot-reload degradation`（含 design Decision 8 将 walk error 列为降级触发）两项 requirement；修复票须同时引用两者，不能只加降级标签而会话仍失败。13 内不实现修正。

**环境与数据还原**：全部专用测试目标、外接卷目录、61k 慢目标树、临时 CLI 与脚本产物已删除；App 自有新键（首启引导、最近目标）恢复为会话前缺失状态，旧 `go-grip-history` 键未动；per-app TCC 已重置回无记录；语言还原 en-US；登录项 Not set；Services 启用条目仍指向该候选；候选未被修改（工具哈希/签名复验一致）；宿主与 owned 子进程均已退出。证据留存于本记录及截图/日志（`~/GoGrip-acceptance-13/evidence/`、`state/`）；本轮与复核补证使用的只读驱动/驱动脚本留存于 `~/GoGrip-acceptance-13/tools/`（finder-service.sh、lib.sh、panel_*.py、readykill.sh、uictl），复核补证的文本记录追加于 `evidence/log.md`「Re-verification」段。

**账本**：7.1–7.3 仍不勾（7.1/7.2 尚缺网络共享/断卷/浏览器拒绝子项，7.3 缺 Intel 与 macOS 13）；本票保持 **open**，待缺陷修正与 12 重新交付后复验受影响路径，并与其余阻塞项一并复审、再按各自批准条件关闭。

## Execution record — 2026-10-06（新候选复验；票保持 open）

**候选身份核对（不重建）**：本轮对象为 12 Redelivery record 标识的新候选：zip sha256 `5062025121c62f0c16fab42135ce7446e0c9bea660645694f8ee79154dd688e2`、dmg `5dd330e802e39829a971c2b008aeb8c98805c95be698d873da21f0d2a0841d81`、archive 内工具 `61040c8225f996671a6c911e119bbb8c03532e3a9c8f05293f69d6b2db89dca2`、宿主 `09ab6b280ef7aa62e06d6344a19c1d9160fafe8fadfcf09efdb3f681e8a85878`。安装前 `/Applications/GoGrip.app` 为旧候选（工具 `7426a27f…`、宿主 `d6677502…`）；已备份至 `~/Desktop/GoGrip-backups/GoGrip-12old-backup.app`（两哈希复核一致，旧 dmg/zip 亦留存于 `macos/.build/candidate-12old-backup/`）。从新 dmg 挂载后 `ditto` 安装到 `/Applications`，安装后工具/宿主哈希与 archive 逐字节一致、`codesign --verify --strict` 通过、`x86_64 arm64`；`lsregister -f` 刷新；未重新构建。Services 条目仍为 `/Applications/GoGrip.app`（用户已启用），本轮**未需重勾**且未修改任何服务偏好。

**批准范围**：用户批准安装/替换 `/Applications`（含备份与回退）、必要时重勾 Services、TCC 允许/拒绝系统提示的真实点选、真实强杀确切宿主/子进程 PID、慢启动目标窗口、Finder AX 驱动与 Chrome 观察；并裁定**本轮跳过并标记未执行**语言（AppleLanguages/Finder 中文名）与登录项真实开关场景。

**环境**：Mac14,3（M2）、macOS 27.0.1（26A434）arm64、系统语言 en-US；独立 CLI 为 `/Users/ray/go/bin/go-grip`（sha256 `404ac666…`，本轮以 `--browser=false --port 7801` 常驻，作为隔离对照）。面板/对话框操作用扩展的 UI 驱动 `~/GoGrip-acceptance-13/tools-r2/uictl2`（由已有 `uictl.swift` 增加 `--xdesc` 精确描述与 `--nth` 序号 AXPress 后编译；**驱动仅操作现成 UI（AXPress/输入事件），不修改产品**；披露：坐标点击在本机多屏布局下不可靠，行级动作改用 AXPress 序号定位；NSOpenPanel 的 CJK 路径用 System Events 直接设置 Go-to 字段 AX 值）。

**7.1 证据（同一新候选）**：Finder 冷启动（宿主未运行）：Finder 选中 `/tmp/gogrip13/main` → Services → Open with GoGrip → 宿主 PID 40491（PPID 1，路径 `/Applications/GoGrip.app/Contents/MacOS/GoGrip`）约 3s 内启动；首启引导 alert（“Using GoGrip from Finder”）实际出现；owned child `--managed A937FEFE… -r -- /tmp/gogrip13/main`（PID 40493，`127.0.0.1:55853`）；**默认浏览器自动打开成功**（Chrome 标签页 `http://127.0.0.1:55853/sub/nested.md`，页面文本 `G13 nested marker`）；HTTP 200 与页面配对。递归目录（root.md/sub/nested.md 200）、符号链接复用（`main-link` → 同一 child 40493，浏览器再次打开）、并发同目标（两次同时请求 `empty` → 仅 1 child）、父/子/文件独立（`main` 40493、`main/sub` 40592 以自身为根：其下 root.md 404/nested.md 200、单文件 40559 三会话并存）、空目录（`empty` 40618：页面 “No Markdown files”；新增 `added.md` 后同一会话页面显示 `G13 added hot reload marker`，无重建）。CJK/空格单文件：经 **Finder 原生选择**成功（child 40559 `--managed … -- /tmp/gogrip13/main/sub/中文 笔记.md`，端口 55896，URL `…/%E4%B8%AD%E6%96%87%20%E7%AC%94%E8%AE%B0.md`）。批次：5 个不同目录直接打开 5 child 无确认；6 个 → 原生确认 “Open 6 previews?（…6 distinct targets…）”，Cancel 后 0 新 child、浏览器标签数不变、既有会话不变，Open 后 6 child；6 条路径含 `a5-alias` 归一为 5 → 无确认、5 child；混合 `m1.md + unsupported.png + locked(chmod 000 根)` → m1 会话保留（200），**一次**原生汇总列出两个失败目标及原因（不支持类型；“target-unavailable: access target: … permission denied”+ 授权检查指导），面板失败信息可查看；全成功批次无额外提示。面板：同名不同路径三行可区分（`same/alpha/docs`、`same/beta/docs`、`same/delta/docs`，各自 Running 与独立 URL）；Copy URL 写入实际 URL（剪贴板 = 该会话 URL）；Open in Browser 再次请求；单会话 Stop 释放对应 child/端口且其他会话继续 200；Stop All 一次释放全部 owned child（24→0）；Recent 重开已停止目标成功新建会话；有活动会话时 Clear 清空 Recent 而会话继续（清空后退出并 `open -a` 重启，Recent 仍为空）；非空历史重启：打开 `/tmp/gogrip13b` 后正常 Quit、`open -a` 重启 → 面板仅恢复 Recent 显示该目标（Sessions 为 “No preview sessions”）、无自动打开、浏览器标签数不变（14→14）。TCC 受保护目标：`~/Documents/gogrip13-tcc` 在用户于系统提示点选“允许”后可预览（child 41971，端口 56843，200 + marker）；随后已重置并删除。面板 NSOpenPanel 的 CJK 单文件选择亦成功（child 42094 单文件模式，端口 56924，200，浏览器打开对应 URL）。

**7.2 证据**：正常 Quit（多会话）→ 宿主退出（日志 `Termination complete. Exiting without sudden termination`）并释放全部 owned child/端口，独立 CLI 持续 200；SIGKILL 宿主（2 owned child，端口 56567/56570）→ 两者随宿主消失、端口关闭、CLI 不受影响；ready 前强杀（50k 文件慢目标：child 41511 已创建且 `lsof` 无监听）→ 杀宿主后 child 消失、Chrome 标签数不变（无新浏览器请求）；子进程意外退出（kill -9 41536）→ 面板行 `Stopped` + “The preview service exited unexpectedly (status 9)”，不自动重启，其他会话继续，Recent 重开得到新 child 41580 与新 URL；运行中删除目标 → 面板 “Target unavailable: open /tmp/gogrip13/todelete: no such file or directory” 且保留 Stop 入口，HTTP 404、4s 后仍 404（不重连），Stop 正常释放；单目标启动失败（`locked` 根）→ 0 新 child、无假 running，一次错误提示含目标/原因/指导；watch 降级（`partial` 含 `chmod 000` 子树）：面板显示 “Hot reload degraded: walk error at …/denied: permission denied”，可读文档 200、无 unavailable 误报（14 修复在候选层的复验）。

**12 轮三观察复核**：① 首次会话自动浏览器打开——本轮在 **installed/LaunchServices 注册的候选**上首次会话即观察成功；12 轮“未观察”属未注册 `/tmp` 副本 + `launchctl submit` 上下文，本记录不复述为缺陷。② CJK 单文件原生选择——Finder 选择与面板 NSOpenPanel 两条真实路径均成功；12 轮的合成按键/粘贴输入失败本轮未再复现，归为输入方式限制的观察（其历史成因未被本轮独立证实），不作为产品缺陷——与用户“不把合成输入失败判为缺陷”的裁定一致。③ 宿主退出原因——本轮所有宿主退出均由明确操作触发且日志可归因（正常 Quit/强杀/ready 前），未复现 12 轮的未确证自行退出（该次历史成因仍未被独立证实）。

**行为套件**：`go test ./... -count=1` 全部包 ok（cmd 5.3s、internal 4.8s、hotreload 4.3s、pkg/*）；`make macos-test` **95 tests / 95 passed / 0 failed / 0 skipped**（xcresult `Test-GoGrip-2026.10.06_21-48-22`）；`gofmt -l .` 干净。源码树与候选构建同源（HEAD 295311a + 14 修复），本轮无代码/构建改动。

**未执行/缺环境（保持未完成，不代签）**：Intel 实机与 macOS 13 实机运行（7.3）；已挂载网络共享卷（08/3.3 缺口）；外接卷断卷；候选级“系统拒绝浏览器打开”；服务关闭状态下的 Help 复核；登录项真实开启/关闭；简体中文界面与 Finder 中文服务名；TCC **拒绝**路径（用户本轮仅点选允许；POSIX 权限拒绝已由 chmod 000 场景与错误指导覆盖）；目标移动（本轮执行删除路径）；干净安装与 Developer ID/公证（后续阶段）。

**环境与数据还原**：宿主、owned child 与独立 CLI/`launchctl` 作业均已结束（无残留监听）；`/tmp/gogrip13`（含 chmod 000 目录与 50k 慢目标树）、`/tmp/gogrip13b`（非空历史重启用）与 `~/Documents/gogrip13-tcc` 均已删除，`tccutil reset SystemPolicyDocumentsFolder com.showgp.GoGrip` 恢复无记录；App 自有键（首启引导、最近目标、面板窗口帧）已删除，defaults 域恢复为仅 `go-grip-history`；System Services、登录项、系统语言与用户卷未改（新候选沿用已启用条目，无需重勾）；`/Applications/GoGrip.app` 保留为**新候选**（工具 61040c82…、签名 strict 通过），旧候选备份与旧 dmg/zip 保留；剪贴板现存本轮 Copy URL 测试值（会话前为用户文本，属已批准动作的副作用）；证据截图含浏览器窗口与少量文件名，外部分享前应裁剪/脱敏；证据/脚本留存 `~/GoGrip-acceptance-13/evidence-r2/` 与 `tools-r2/`（另有本轮 CLI 常驻脚本于 `tools/`）。

**账本**：7.1–7.3 仍不勾（仍缺网络/断卷/浏览器拒绝/TCC 拒绝/语言/登录项与 7.3 两环境），**13 保持 open**；本记录不代替 08/3.3 关闭，也不宣称整个变更完成。

注（2026-10-07 范围对齐，新增）：本记录上方「未执行/缺环境」按当时事实保留、不回写。当前分类：① 延期（未验证、不阻塞本票关闭）：Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷验收；② 浏览器失败分支按已批准 A 以既有消费者行为回归的确定性结果验收（受控辅助不是该分支必需证据；候选级真实系统拒绝未触发，不冒充系统级拒绝验证）；③ 本机剩余验收（仍须完成）：TCC 真实拒绝及授权主体、目标移动、服务关闭状态下再次查看 Help、简体中文界面与 Finder 中文服务名（含不支持语言回退）、登录项真实开启/关闭、最终候选外接卷只读访问；④ 干净安装与 Developer ID/公证属后续正式分发阶段，不属本机剩余必验或本票关闭条件。详见上方「当前范围与关闭条件」。

## Execution record — 2026-10-07（本机剩余补验：b/c/f1/f2/d/e 与 a 步骤 1；票保持 open）

**前置（候选身份与基线，只读）**：`/Applications/GoGrip.app` 工具 `61040c82…dca2`、宿主 `09ab6b28…5878`（与 12 交付 archive 一致）、`com.showgp.GoGrip`、adhoc、universal；基线快照存 `state-r3/`（`pbs-before.plist`、`servicesmenu-before.plist`、`applelanguages-before.txt`=("en-US")、`app-defaults-before.plist`=仅 `go-grip-history`、`candidate-identity.txt`、`session-locked-evidence.txt`）。起始会话为**锁屏**（`CGSSessionScreenIsLocked=1`、截图全黑 `litPixels=0%`、System Settings 无窗口），GUI 项无法执行；解锁后按 a1→b→c→f1→f2→d→e 执行，期间以 `caffeinate -d -i` 防显示器休眠（随进程失效）。未重建/重装候选、未改实现代码。

**a 步骤 1（只读确认；未执行任何 reset/点选）**：`sqlite3 -readonly` 直读用户 TCC.db 被拒（`unable to open database file`；本会话读取方未持 FDA，本轮不申请、不开启 FDA）；改经 System Settings → Privacy & Security → Files & Folders 只读核对：应用列表 25 行（含 Ghostty.app、iTerm.app 等相邻项）中**未显示 GoGrip/com.showgp.GoGrip**（`state-r3/ss-ax-faf3.txt`、`evidence-r3/11c-files-and-folders.png`）。**系统界面未显示 GoGrip；底层 TCC 记录及 Documents 初始授权状态尚未确认，不能据此确认处于未决定状态。** reset、拒绝点选与任何还原操作均未执行；初始状态未确认时继续遵守**不 reset** 的已批准边界，拒绝路径执行与还原方案**待单独批准**；不得以“无需前置 reset”或“reset 可恢复原状态”为前提（原状态未经确认）。

**b 目标移动（实际候选）**：`/tmp/gogrip13-r3/b-move/main`（root.md/sub/nested.md）经面板 Open…（NSOpenPanel Go-to）打开：child 53636 @127.0.0.1:53095，URL `…/sub/nested.md`，HTTP 200 与 `R3-B-ROOT-MARKER` 配对。10:15:21 对已打开目标执行 `mv` 移走：原 URL root.md/sub/nested.md 均 **404**，child 未变、无新 child、无重连；面板保留目标行并显示 “Target unavailable: open /tmp/gogrip13-r3/b-move/main: no such file or directory”，保留 Open in Browser/Copy URL/Stop，**未误报空目录**；3 分 17 秒后复核仍 404、child 仍 53636（不跟踪新位置）。面板 Stop → child 释放、端口关闭、行转 “Stopped”；fixture 删除。证据 `12a–12f`、`panel-b-*`。

**c 服务关闭状态 Help（真实勾选）**：现状=**启用**（Keyboard Shortcuts… → Services → Internet 组 “Open with GoGrip”，AX checkbox=1；`pbs` 无 GoGrip 条目）。取消勾选 → checkbox=0，`pbs` 出现显式禁用条目 `"com.showgp.GoGrip - Open with GoGrip - openWithGoGrip" {enabled_context_menu=0; enabled_services_menu=0; presentation_modes={ContextMenu=0;ServicesMenu=0}}`；Finder 服务子菜单（Finder 前台、选中目录）**不再出现** “Open with GoGrip”。面板 Help 实际再次出现完整指导（含“如果菜单里没有这个命令，请打开‘系统设置’→‘键盘’→‘键盘快捷键…’→‘服务’，确认 GoGrip 的服务已启用”）；Help 后 checkbox 仍 0、pbs 条目未变（**应用不自行启用**）。恢复：重勾 → checkbox=1，系统**移除**该 pbs 条目 → `pbs.plist` 与执行前逐字节一致；Finder 菜单恢复显示 “Open with GoGrip”。证据 `13a–13i`、`state-r3/pbs-*`、`help-while-disabled.txt`。

**f1/f2 最终候选外接卷只读访问**：f1 在 `/Volumes/GW2T_PICE3/gogrip-13r3-extvol`（卷 `/dev/disk5s1`）建 `main/root.md`、`main/sub/nested.md`、`file.md`（markers `R3-EXT-*`）。f2：面板 Open… 打开目录 → child 55482 @53698（cmd 含 `-r -- /Volumes/…/main`），HTTP 200 + `R3-EXT-ROOT-MARKER`，浏览器打开 `…/sub/nested.md`；单文件 `file.md` → child 55530 @53719（cmd 无 `-r`，单文件模式），HTTP 200 + `R3-EXT-FILE-MARKER`；面板两行均 Running、无降级提示；**全程未出现任何系统授权提示**；Stop All → 全部释放、端口关闭；fixture 删除。证据 `14a–14d`、`panel-f2-*`、`extvol-fixture-f1.txt`。

**d 语言（含还原）**：原值 `("en-US")`（键存在）。设 `("zh-Hans","en-US")` 并重启 Finder/App → 面板/引导/错误全中文（面板“会话/最近目标/登录时启动/未开启/打开…/帮助/全部停止/退出”；引导“在 Finder 中使用 GoGrip…”；stale recent 错误“未能打开文件…请检查目标的文件/文件夹权限…”），Finder 服务名 **“用 GoGrip 打开”**。设 `("ja")`（仅不支持语言）并重启 → 面板/引导英文回退、Finder 服务名 **“Open with GoGrip”**（ja 条目）。恢复 `("en-US")` 并重启 → 英文面板与 Finder 名恢复。证据 `15a–15g`、`help-zh.txt`、`help-ja.txt`、`panel-zh-*`、`panel-ja-open.txt`。

**e 登录项真实开关（含还原与残留）**：初始面板 **“Not set”**（SMAppService notRegistered）；系统 Login Items → “Open at Login” 3 项（Android File Transfer Agent.app、BuhoNTFSMenu、Google Drive），无 GoGrip；`launchctl` 仅应用自身作业与既有 `com.showgp.GoGrip.ServiceProvider`。面板 Turn On → 面板 **“Enabled”**+Turn Off；系统 “Open at Login” 出现 **GoGrip.app（Application）**。面板 Turn Off → 面板回到 **“Not set”**+Turn On；系统 “Open at Login” 中 GoGrip.app 移除。**注册/注销路径已验证；系统 “Open at Login” 已恢复执行前状态。残留（如实记录，未自行清理）**：Background App Activity 在 Turn On 后出现 “GoGrip.app” 行（toggle=1、状态随运行/退出变化），Turn Off 与退出 App 后仍保留，**因而环境未完全还原，残留性质待确认**——既不能认定为仍会登录启动，也不能认定为无害历史记录；`sfltool dumpbtm` 在本环境 30s 超时不可用（无输出），BTM 深读缺失；处理方式**待裁决**（不将全局 `sfltool resetbtm` 作为普通清理方案）。证据 `16a–16f`、`state-r3/ss-login-items-*`、`panel-e-*`。

**环境与数据还原**：`AppleLanguages` 恢复 `("en-US")`；`pbs.plist` 与执行前逐字节一致；`com.showgp.GoGrip` 删除本轮新建键（`go-grip-first-use-guidance-shown`、`go-grip-recent-targets-v1`、`NSNavPanelExpandedSizeForOpenMode`、`NSOSPLastRootDirectory`、`NSWindow Frame GoToSheet`）后与执行前一致（仅 `go-grip-history`）；测试目标（/tmp）、外接卷 fixture、临时目录均删除；本轮 Chrome 打开的 127.0.0.1 标签页（53095/53698/53719）已关闭；System Settings 退出；无宿主/child/监听残留；候选身份哈希复核一致。

**未执行/未完成（保持未完成，不代签）**：a 的 TCC 拒绝路径（触发 Documents 访问 + 系统提示点选）与还原方案**待单独批准**（初始状态未确认，遵守不 reset 边界）；延期四项（Intel 实机、macOS 13 实机、已挂载网络卷访问与热重载降级、候选级可卸载卷断卷）保持未验证；BTM 深读不可用；e 的 Background App Activity 残留性质与处理**待裁决**。

**账本**：7.1–7.3 仍不勾（本票关闭仍需：a 的 TCC 拒绝路径执行与还原方案获批（初始状态未确认，遵守不 reset 边界）、e 残留性质与处理裁决，以及独立 Standards/Spec 审阅与关闭批准；延期四项不阻塞）；**13 保持 open**；本记录不代替 08/3.3 关闭，也不宣称正式签名/公证/干净安装/公开发布。

## 证据强度纠正与 state-r4 只读调查（2026-10-07 追加；不重写上文原始观察）

本节按 2026-10-07 只读调查（留档 `~/GoGrip-acceptance-13/state-r4/`：`a-tcc-audit.txt`、`a-tccd-log-events.txt`、`e-btm-audit.txt`、`e-evidence-audit.txt`）对上一节 a、e 两项的证据强度与归因作明确纠正；原始执行观察与证据引用保持不变。调查为只读：未申请/未开启 FDA、未 reset、未触发任何授权提示、未变更系统状态。

**a — 路径与读取失败归因（纠正）**：经典路径 `~/Library/Application Support/com.apple.TCC/TCC.db` 在本机**不存在/不可见**（`ls`/`file`/`stat`/`dd` 一致 ENOENT）；上节 `sqlite3 -readonly` 的 `unable to open database file` **源于该路径不存在，不归因于缺少 FDA**。用户 TCC.db 实际位于受保护容器（tccd 打开句柄 `/private/var/containers/Data/ProtectedSystem/DF47B43C-…/…/com.apple.TCC/TCC.db`），普通进程只读访问被拒（`authorization denied`）；**FDA 能否解决该实际路径访问未经验证，不作保证**。系统库（`/Library/Application Support/com.apple.TCC/TCC.db`）可读且无 showgp 行，但不含 per-user（Files & Folders）记录，不能用于确认候选状态；另有 10-06 日志层证据（AUTHREQ→Create→Delete，见 `a-tccd-log-events.txt`），非当前库读取。据此重申：界面未显示 GoGrip ≠ 底层无记录，Documents 初始授权状态仍未确认。

**e — 证据覆盖与可审计性（纠正）**：`evidence-r3/16a–16f` 截图**未覆盖 Login Items 面板内容**（16e 为 GoGrip 面板，属应用侧证据）；`state-r3/ss-login-items-pre2.txt`、`ss-login-items-scan.txt` 为 Turn On 前状态、**无 GoGrip 行**；含 GoGrip 行的临时 UI 文本（/tmp/li-*.txt）已删除。故上节“Open at Login 出现/移除 GoGrip.app”“Background App Activity 行保留”属**当时运行观察**；现存可审计证据为 BTM 守护进程日志与 BTM 存储记录，**不以 BTM 存储替代历史 UI 转换证据**。

**BTM 存储记录（证据强度限定）**：uuid `4606E17B-ED6B-4552-BC3E-F29F9E8788EC`、identifier `2.com.showgp.GoGrip`、disposition=2、modificationDate=**2026-10-06 20:47:01（早于本轮）**→ 不声称该条目由本轮新增，也不声称已证明注销；disposition 位语义（11/10/2 对比及 enabled/allowed/notified 措辞）为**跨项对比＋守护进程日志措辞推断，非官方位表**。

## 本轮补验结果（2026-10-07 追加：e 当前状态只读核对、a 真实 TCC 拒绝路径；票保持 open）

**e 当前状态（只读；13:16 打开 Login Items 面板、13:22 退出 System Settings；未切换开关、未注册/注销、未 resetbtm、未登录/重启）**：
- Open at Login：Android File Transfer Agent.app、BuhoNTFSMenu、Google Drive（Kind=Application）——与执行前基线一致，**无 GoGrip**。
- Background App Activity：AX 完整列表 22 项，**未显示 GoGrip**（G 区 Ghostty.app 与 Google Drive 相邻，无 GoGrip 行）；截图覆盖面板顶部与 G 区并逐张核对（`state-r4/e-current-top-d2.png`、`e-current-gregion-d2.png`，正确显示器 -D 2；AX `ss-login-items-current.txt`、`ss-login-items-current-scrolled.txt`）。
- 状态差异（系统行为）：打开面板期间 13:16:16 `backgroundtaskmanagementd` 执行 gc 并移除 GoGrip 的 BTM 记录（`state-r4/e-btm-log-after.txt`：`gc … removing uuid=4606E17B-…, name=GoGrip, identifier=2.com.showgp.GoGrip`）；BTM 存储复读（13:16）已无 GoGrip 记录（41 条）。此为系统 GC，非人工删除；此前 11:07–12:06 尚存的记录已被系统清除。
- 裁决记录：接受保留现有 BTM/Background App Activity 记录、不要求清除系统记录；上述系统 GC 与该裁决不冲突（无人工清除动作）。当前 UI/存储均无 GoGrip 记录，不等于判定其“无害”或“已完全还原”。
- 本核对仅报告当前状态，**不用于补证历史开启阶段**。

**a 真实 TCC 拒绝路径（本轮授权执行；未 reset、未用 sudo/FDA、未申请 FDA）**：
- 前置（只读）：候选身份复核一致（宿主 `09ab6b28…5878`、工具 `61040c82…dca2`）；`pbs.plist` 无 GoGrip 条目（服务启用）；AppleLanguages=("en-US")；无 GoGrip 运行。
- 准备：`~/Documents/GoGrip-13-r4-deny/main`（root.md、sub/nested.md；权限 0644/0755；`state-r4/a-fixture.txt`）；准备过程**未出现任何授权提示**（tccd 无相关 AUTHREQ）。
- 触发（Finder Services，未用 NSOpenPanel）：`tools/finder-service.sh "Open with GoGrip" …`（rc=0）；13:23:20 GoGrip（pid 59744，`/Applications/GoGrip.app/Contents/MacOS/GoGrip`）启动；13:23:20.696 `AUTHREQ_PROMPTING: service=kTCCServiceSystemPolicyDocumentsFolder, subject=Sub:{com.showgp.GoGrip}Resp:{…GoGrip…}`（attribution：accessing=GoGrip pid 59744；requesting=sandboxd）。
- 提示核对与点选：提示属 UserNotificationCenter（pid 30967）窗口，文本 `“GoGrip.app”想访问“文稿”文件夹中的文件。`，按钮 不允许/允许（`state-r4/a-deny-prompt.png`、`a-deny-prompt-ax.txt`）→ 明确属 GoGrip 且为 Documents 访问 → 13:26:35 以 AXPress 触发**“不允许”**按钮（等效点选；提示截图先于点选保留）。
- 结果：13:26:35.287 `AUTHREQ_RESULT: msgID=45036.566, authValue=0, authReason=2`（拒绝）；`Publishing <TCCDEvent: type=Create, service=…DocumentsFolder, identifier=com.showgp.GoGrip>`（**拒绝记录已创建并保留**）。证据引用（2026-10-07 补充）：`state-r4/a-deny-tccd-log.txt` 为原始日志留档，含 PROMPTING/Create、**不含 AUTHREQ_RESULT**；完整 45036.566 请求链（REQUEST→CTX/ATTRIBUTION/SUBJECT→PROMPTING→RESULT→REPLY，逐字摘录）另见 `state-r4/a-deny-tccd-result-45036.566.txt`（提取自执行会话原始工具输出：会话 `01a11184-e3ed-73e5-858d-578a47f731d2`，toolResult 第 1569 行，工具结果时间 2026-10-07T05:28:23.238Z；非新采集、未重跑）。拒绝来自 TCC 层（AUTHREQ_PROMPTING→RESULT authValue=0），**非 POSIX 权限拒绝**（fixture 权限正常）。
- 应用拒绝表现：面板行状态 **“Stopped”**，错误 `The preview service reported target-unavailable: access target: open /Users/ray/Documents/GoGrip-13-r4-deny/main: operation not permitted Check the target's permissions, the volume's mount or sharing state, and — when macOS asked for authorization — System Settings → Privacy & Security → Files and Folders (GoGrip's entry).`（`state-r4/a-deny-error-text.txt`、`a-deny-panel-ax.txt`、`a-deny-panel.png`）；**无 Running/假成功**；Recent “No recent targets”，`go-grip-history` 内容与执行前一致（被拒目标未进入历史）；无 child 进程/监听（pgrep/lsof）、无浏览器标签（Chrome 127.0.0.1 计数 0）。
- 应用启动时显示首次使用引导（“Using GoGrip from Finder”，含授权检查路径说明），并写入 `go-grip-first-use-guidance-shown=1`（`state-r4/a-app-defaults-after.txt`）。**裁决（2026-10-07）**：接受保留该键，记录为本轮运行产生、经批准保留的应用偏好差异；**不删除该键，不声称应用偏好已完全恢复执行前状态**。
- 收尾：约 13:35 GoGrip 退出（`osascript quit` rc=0，无残留进程/监听/标签）；`rm -rf ~/Documents/GoGrip-13-r4-deny`（rc=0，无残留）；关闭 Finder 中本轮触发产生的窗口（2 个，~920×436 级联；判定依据：同尺寸级联、创建于触发时段、触发前 13:16 截图中不存在；`state-r4/a-finder-windows.txt`）。
- **状态与还原声明**：Documents 拒绝记录**已产生并保留**；本轮**未 reset**；初始授权状态未确认，**不称“已还原”**；Files & Folders 界面未复核（本轮未授权该读取）；底层 TCC 库仍未直读。**裁决（2026-10-07）**：Documents 拒绝状态继续保留；**不 reset、不授予 FDA**。

**未完成/未变（保持）**：延期四项（Intel 实机、macOS 13 实机、已挂载网络卷访问与热重载降级、候选级可卸载卷断卷）保持未验证；e 的 BTM 深读仍不可用；正式分发（Developer ID/公证/干净安装/公开发布）不在本阶段。**08/13 保持 open，Acceptance 与 tasks 未勾选；未提交/推送/同步/归档。**

## 证据出处注记（2026-10-07 追加：语言与登录项原始工具输出；不改写历史观察）

本节按 2026-10-07 定向检索（只读；未重跑、未新采集、未变更系统状态）补充 d、e 两项关键证据的出处与限定：相关原始工具输出仍留存于执行会话（`01a11184`）的压缩回放文件 `212.shake.log`；已逐条核对会话 ID、region 与原始内容，并逐字摘录留档（`state-r3/finder-service-name-raw.txt`、`state-r3/login-item-transition-raw.txt`；含来源、工具结果时间、toolCallId 与提取时间 2026-10-07 16:22 +08:00）。**相关截图不独立承担证明：`15c`/`15g` 未显示可读的服务子菜单文字；`16a–16f` 未覆盖 Login Items 面板内容。** 本节只补充出处，不改写上文原始观察与纠正段。

- **语言服务名（d）**：zh-Hans 菜单枚举实际包含 **“用 GoGrip 打开”**（region 115；工具结果 2026-10-07T03:02:47.188Z；toolCallId `call_00_9CXTQIalk1lf5YlWnOzi2632`）；ja 回退菜单枚举实际包含 **“Open with GoGrip”**（region 121；2026-10-07T03:03:58.241Z；`call_00_ET_ZkAZDtdnMtzeIW78gVu53184`）。
- **登录项历史转换（e）**：开启链＝面板 “Not set” → Turn On（region 133；03:06:58.832Z；`call_00_mAJOY4Y9vtO6ZRQ4SQlk4649`）→ 面板 “Enabled” 且系统 Login Items 出现 **GoGrip.app/Application** 行、Background App Activity 出现 GoGrip.app（region 134；03:07:10.573Z；`call_00_BymR3TRoTYo8JuOSzdLU6610`）；关闭链＝Turn Off 可用性检查（region 136；03:07:32.256Z；`call_00_gzA9JRkqMRsG8MPlnhQn6727`）→ 实际按下 Turn Off 并关闭成功：面板回到 “Not set”，系统面板仅剩 1 处 GoGrip＝Background App Activity 残留行、Open at Login 行已移除（region 137；03:07:46.783Z；`call_00_IKH7jG91r24G2pVmbE9a1532`）；开启前基线复核 gogrip 计数 0（region 138；03:07:58.019Z；`call_00_pXNTnrs3zok3NF1Qn6ZE7554`）；退出后残留仍在（region 139；03:09:24.025Z；`chatcmpl-tool-8dfb5f0b776828b0`）。**第一次关闭尝试（region 135；03:07:23.456Z；“only 0 matches”）为失败操作，不作为成功关闭证据。**
- **边界**：以上为当时运行观察的原始工具输出恢复引用；不表示相关场景重新执行或已通过独立复审，不改变 e 当前状态（r4）与系统 GC 的分别判读，也不改变本票 open 状态与 Acceptance/tasks 未勾选。
- **引用修正（2026-10-07）**：登录项关闭链的 region 引用修正——region 136＝关闭前 Turn Off 可用性检查；实际按下 Turn Off 与关闭后状态在 region 137。仅修正引用，不改写历史结论。

## 四项定向补验（2026-10-07：提示期间请求、跨入口复用、关浏览器不停止、降级后手动刷新）

按 2026-10-07 批准的修正最小方案执行；仅使用自建 /tmp fixture、本轮 owned 服务与本轮浏览器标签；执行前复核候选身份（宿主 `09ab6b28…5878`、包内工具 `61040c82…dca2`，与最终候选一致；未重建/重装）。四项均取得完整转换链（以下为本轮实施者观察，尚未经独立复审；未新增永久测试、未改实现）：

1. **提示显示期间 Services 请求**：17:03:27.524 Finder 一次选择 `/tmp/gogrip15/t1/validA`＋不支持文件 `note.png` 触发原生汇总提示（逐字：`/tmp/gogrip15/t1/note.png` / `Only directories and Markdown files can be previewed`；未用 chmod 000 子树触发），validA 会话同时打开（child 64002，127.0.0.1:58048）；17:03:52.695 在提示保持显示时经 Services 请求 validB → 目标实际被处理（child 64034，127.0.0.1:58065；页面 marker "G15 t1b marker" 与 "G15 t1a marker" 均经 Chrome AX 读取）；提示仍显示；17:04:17.472 AXPress OK 关闭成功（alert 计数 0、系统窗口计数 0，无残留窗口）。证据：`state-r5/t1-record.txt`、`evidence-r5/t1-{alert-open,alert-open-validB,after-close,amd-page}.png`。
2. **活动会话跨入口复用**：17:05:59.886 Finder 打开 `/tmp/gogrip15/t2`（child 64149，`--managed BFB2002B-…`，127.0.0.1:58098）；17:06:16–42 经面板 Open…（NSOpenPanel，Where: t2）重开同一规范化目标 → 未新建服务：child 计数 1、PID/UUID/端口/URL 均不变、owned children 集合不变；同 URL 浏览器标签 1→2（复用请求再次要求默认浏览器打开；服务未重复）；面板仍单行 Running。generation 无现成诊断，未记录。证据：`state-r5/t2-record.txt`、`evidence-r5/t3-before.png`。
3. **关闭浏览器预览不停止服务**：17:07:38.381 记录活动会话（child 64149、listen 58098、HTTP 200、标签数 2、Running）；17:07:38–41 仅关闭 URL 以 `http://127.0.0.1:58098` 开头的本轮标签（返回 2），未 Stop/Quit；17:07:42 复核：同一 child 存活、端口仍 LISTEN、HTTP 200、标签数 0、面板 Running。证据：`state-r5/t3-record.txt`、`evidence-r5/t3-{before,after}.png`。
4. **降级后浏览器手动刷新**：17:07:55.328 Finder 打开 `/tmp/gogrip15/t4`（可读 doc.md＋`locked/` chmod 000；child 64298，`--managed 0A532242-…`，127.0.0.1:58137）；候选 App 实际显示降级及原因（面板 AX 行，AX 值原文止于 `…open /tmp/gogrip15/t`，`…` 为摘要标记：`Hot reload degraded: walk error at /tmp/gogrip15/t4/locked: open /tmp/gogrip15/t…`）；浏览器显示 v1 marker；17:08:39.805 改写 doc.md→v2；Chrome 前台 ⌘R 主动刷新后 Chrome AX 显示 "G15 t4 marker v2"；同一 child、HTTP 200（补充）、面板仍 Running＋降级行。证据：`state-r5/t4-record.txt`、`state-r5/t4-panel-degraded.txt`、`evidence-r5/t4-{degraded,browser-v1,browser-v2}.png`（降级行与页面标题以原始 AX 输出证明；截图仅作表面补充）。

**收尾与还原（如实记录）**：Stop All→children 0；Quit 正常退出（17:10:46 voluntary，无崩溃报告）；本轮 5 个浏览器标签全部关闭（127 标签=0；Chrome 标签总数 17＝执行前）；7 个本轮 Finder 窗口全部关闭；fixture 权限恢复后删除；应用偏好按前置快照逐键恢复并验证与执行前完全一致（`state-r5/defaults-{before,after,restored}.plist`；本轮新增键 `go-grip-recent-targets-v1`、`NSNavPanelExpandedSizeForOpenMode`、`NSOSPLastRootDirectory`、`NSWindow Frame GoToSheet` 已删除；`go-grip-history` 字节不变；`go-grip-first-use-guidance-shown=1` 按批准保留）。**未恢复/保留差异**：Documents 拒绝记录（r4 产生，批准保留，未 reset/未授予 FDA）；系统 BTM/Background App Activity 历史未改动（本轮未涉及）；未声称环境完全还原至执行前状态。
**边界**：以上为本轮实施者观察并留档；不表示已通过独立复审；不改变本票 open 状态与 Acceptance/tasks 未勾选；延期四项（Intel 实机、macOS 13 实机、已挂载网络卷访问与热重载降级、候选级可卸载卷断卷）仍为未验证。

## Closure record — 2026-10-07

按调整后范围关闭本票：四项定向补验（提示期间 Services 请求、活动会话跨入口复用、关闭浏览器不停止、降级后手动刷新）经独立两轴**定向**复审（`ReviewStandards13R2`/`ReviewSpec13R2`）——Spec 轴 4 项 P2 **全部 closed**（0 阻塞/0 建议/0 待裁决）；Standards 轴 0 阻塞、3 项 P3 记录精度建议（state-08 复制时间、t1b marker 时间注记、t4 降级行摘要标注），**已按其最小修正落实**（不影响证据与裁定）。关票条件确认：(i) 4 项 P2 闭合；(ii) 无未解决阻塞/范围待裁决；(iii) 08 关键证据已持久化（state-08 逐字副本，sha256 一致）；(iv) 批准范围内条件满足。延期四项（Intel 实机、macOS 13 实机、已挂载网络卷访问与热重载降级、候选级可卸载卷断卷）保持**未验证**；浏览器失败分支按已批准 A（既有消费者行为回归；候选级真实系统拒绝未触发如实记录）；Documents 拒绝记录与 `go-grip-first-use-guidance-shown=1` 按批准保留；不声称环境完全还原。审阅限制（无本票单独 git 固定点、未重跑实施者检查、定向范围）如实保留于本记录。本票 `Status: done`；12 项 Acceptance 全部勾选；tasks **7.1–7.3** 已勾。08/3.3 按各自条件独立关闭；未宣称正式签名/公证/干净安装/公开发布，未执行同步/归档/提交。
