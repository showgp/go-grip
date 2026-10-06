# 13 — 候选 App/DMG 跨组件真实功能与支持环境验收

Status: open
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

**真正实施前置**：依据 tasks 第 7 组“依赖第 6 组”，12 的完整交付和独立 Standards/Spec 审阅须先完成，并提供明确来源的最终候选、实际构建/签名/打包记录及使用文档；间接包含 11 的最终功能/语言资源。需要实际平台/卷环境和具体操作许可，不能凭票据发布或交叉编译推断已具备。08 网络缺口保持原归属，缺相关证据时不宣称完整权限/卷 gate 或整个变更已验收；本票可先验证获准且可执行的路径，受影响检查仍未完成，不能以缺环境为由删场景或关闭整票。

**发布不是实施许可**。本轮只新增本票并刷新 [完整覆盖地图](../coverage-plan.md)，不修改批准 artifacts、任务勾选、既有票、代码/测试或用户环境。执行仍须另行批准本票验收范围、现有套件、候选/native/browser 观察、必要的受控辅助手段和具体用户数据/安装/权限/卷操作；不授权提交、推送、打 tag、触发发布、主规格同步或归档。

## Source mapping

- [proposal.md](../proposal.md)：完整原生 Finder/菜单栏/预览功能与自包含候选；真实使用路径验收；所有权、状态、目标权限及最小设置边界；正式签名、公证、安装包验证和发布后置。
- [design.md](../design.md)：Decisions 1–10 的共享宿主/批次/身份/启动/所有权/状态/资源/候选契约；Verification Strategy 的真实 Go、所有权、原生表面与候选环境；Risks 的 Services/TCC、ad-hoc、内核阻塞与跨编译限制；Migration Plan 的 08 缺口、组件顺序、升级/回退及正式分发边界。
- [tasks.md](../tasks.md)：物理 checkbox **7.1、7.2、7.3** 的完整描述及验收；第 7 组只检查已落地组件和记录证据，不接收前面各组的遗漏。apply JSON 顺序 ID 不是 task 标签。

| Task / 行为源 | Requirement 与 scenario 对应 | 本票贡献 |
|---|---|---|
| 7.1 / [macos-finder-service](../specs/macos-finder-service/spec.md) | `Finder Services entry`（`Open a selected folder from Finder`、`Invoke the service while the host is not running`）；`Supported open targets`；`Batch deduplication and quantity confirmation`（5/6、取消、别名计数）；`Partial batch failures`（`Some targets fail`、`All targets succeed`） | 对最终候选的实际 Finder/手动/native/browser 路径验收，不重复实现 provider 或批次。首次/再次查看指导和系统服务启用边界消费已完成 10。 |
| 7.1 / [macos-preview-sessions](../specs/macos-preview-sessions/spec.md) | `One session per normalized target`、`Containment does not merge target identities`、`Directory and single-file preview modes`、`Verified startup and actual preview URL`、`Loopback-only application previews`、`Browser opening is separate from service lifetime` | 最终候选实际内容/身份/URL/回环/浏览器与服务生命周期；包括中文空格单文件、空目录新增 Markdown、跨入口/别名复用、父子/文件独立，不依赖 fake 服务或固定端口。 |
| 7.1/7.2 / [macos-menu-bar-app](../specs/macos-menu-bar-app/spec.md) | `Persistent native menu bar host`、`Visible sessions and explicit controls`、`Manual opening uses the same target contract`、`Bounded persistent recent targets`、`Clearing history does not stop sessions`、`Action failures remain understandable and recoverable` | 实际面板与最近消费者、复制/重开/清除/停止/退出、跨启动仅恢复记录、错误与真实会话一致；不把旧开发 UI 或纯 store 测试当候选表面。 |
| 7.2 / macos-preview-sessions、macos-finder-service | `Explicit stopping confirms service termination`（`Stop one of several sessions`、`Stop all application sessions`）；`Services end with their owning application`（正常退出、强制退出、启动期间退出）；`No automatic session restoration or restart`；`Unavailable paths are not empty directories`；`Visible hot-reload degradation`；`Independent CLI operation remains available`；`Minimum necessary file authorization` | 实际候选的 owned 清理/隔离、异常状态、拒绝/失效/断卷、已知 watch 降级与手动恢复，保留 08 的权限/网络证据归属，不新增自动恢复、权限绕过或 Chrome 清扫。 |
| 7.3 / macos-menu-bar-app、macos-finder-service | `Opt-in launch at login`（`First use without login startup`、`Change the login-start preference`）；`System-language localization`（`Use Simplified Chinese`、`Use English or an unsupported language`）；`First-use service guidance`（`Read first-use instructions`、`Revisit guidance for a disabled service`） | 最终候选消费 10/11 的完整功能和资源，实际 native/Finder 观察纳入证据汇总，不以资源存在、App-only Locale 或 SDK 调用返回替代。 |
| 7.1–7.3 / [macos-app-packaging](../specs/macos-app-packaging/spec.md) | `Universal self-contained application`（无开发工具、同一候选在支持架构运行）；`Candidate DMG and Services-only Finder integration`（可安装候选、新 Services 入口）；`Functional acceptance exercises actual user paths`（真实 Finder-to-browser、复用/停止、异常宿主退出）；`Explicitly deferred formal distribution`（候选评审、完成本阶段） | 消费 12 的实际产物，完成候选级真实路径和支持环境证明。静态架构/部署/签名检查只作其本身证据；不降低或提前完成正式分发标准。 |

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
| 面板、浏览器与明确控制 | 同名不同路径可区分，实际阶段/不可访问/已知降级/原因可查看；Copy URL 定位原目标，主动 Open 复用服务；关闭 popover/浏览器不停止，单停/全部停止有效。浏览器系统拒绝时服务/URL保留，可主动重开或复制。 | Native 可见动作必须对应真实内容/资源。正常/未知访问或 watcher 状态可沿用已接受的安静表面，不以新正向标签/措辞判失败。真实系统拒绝与受控辅助明确分开。 |
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
| 删除、移动、专用卷不可访问 | 对实际已打开的专用目标/测试卷执行获准删除/移动/断卷，再访问原URL；明确原路径 unavailable，保留真实查看/停止入口，既有服务不跟踪新位置、不自动重启/重连、不误为空目录。 | 只操作可丢弃目标/自己的测试卷，不卸载用户工作卷；实际检测/请求触发与状态到达分别记录，不强加不存在的即时发现保证。恢复可访问路径后可按指导主动刷新/停止/重开，不承诺 watch 自动恢复。 |
| 外接与已挂载网络卷、已知 watch 降级 | 真实卷的目录/文件在允许范围内正确预览，真实不可访问/共享拒绝反馈明确。实际预算/注册失败/覆盖不完整使 watcher degraded 时，可查看真实状态/原因、预览仍可用、浏览器手动刷新反映新内容。 | 不用普通目录/别名/命名模拟外接或网络卷；未出现降级不虚构已测降级，网络卷无可靠事件不当普通 watcher。保留无轮询/自动重连策略，08的网络缺口须补真实证据并独立审阅，本票不自动关闭08/3.3。 |

对不易触发的浏览器系统拒绝或特定 watcher 故障，优先现有消费者行为回归与可安全执行的实际场景；需要临时受控辅助时先明确批准、标注所控制边界及产物差异，不能把 mock/受控输出冒充真实系统拒绝、权限主体、断卷或候选所有权证明。辅助观察不改变最终交付候选，结束后还原/清理并从12最终产物复核正常路径，结果归属不同候选时不得混记。

### 7.3 — Supported environments and evidence summary

使用同一候选，记录每次运行的实际硬件、OS版本、执行架构/是否翻译、App/Go来源、native/browser结果与清理。最低系统和两种硬件运行是独立覆盖维度，可以由同一真实环境同时满足，**不新增每个OS小版本或四格全组合运行保证**。

| 覆盖维度 | 所需证据 | 不能替代 |
|---|---|---|
| Apple Silicon / macOS 13+ | 在实际 Apple Silicon 环境运行候选宿主和包内 Go，实际入口到正确内容、面板操作和 owned 清理符合规格。 | arm64 Mach-O 存在、构建成功或另一App/工具运行不够。 |
| Intel / macOS 13+ | 在实际 Intel 环境运行同一候选的宿主/工具，记录对应架构的实际预览、操作和收尾。 | Apple Silicon 的 Rosetta 运行可另记辅助结果，但不冒充 Intel 硬件验证；x86_64 slice/交叉编译不够。 |
| 最低 macOS 13 | 在真实 macOS 13 环境执行候选路径，记录宿主/工具及原生集成的实际结果，可与上面一个硬件维度重叠。 | `MACOSX_DEPLOYMENT_TARGET=13.0`、Mach-O最低版本或较新系统成功不证明最低系统运行。 |
| 当前及其他已测环境 | 分别记录实际版本/架构、已执行场景及结果，与静态架构/部署/签名检查和未测环境区分。 | 当前机器的一轮不能覆盖所有支持环境；缺任一必需维度/路径时对应验收与7.3仍未完成，13保持open。 |

执行相关现有 Go/Swift 行为套件，记录实际命令、结果、skip及与12候选源版本的对应；不引用旧 tests 数量代替当前检查，不为确认已知错误重跑无关lint，不把失败或skip改名为通过。组件行为套件与候选实际运行结果须一致，分歧回到所属组件处理。本票不新建永久集成测试/框架、源码/文案/资源存在/拷贝/YAML/wiring/mock echo断言或重复内部矩阵。

证据写入本票实际执行/关闭记录及现有相关架构/交接说明，区分**本轮真正观察、引用前置记录、静态检查、受控辅助、未执行/缺环境**；记录候选标识、场景/操作、期望、实际结果、截图/输出出处、owned资源收尾及未完成原因。必要使用说明修正回到其所属票，不以本票汇总接收12未完成文档。不凭假设生成PASS，不能以“未见失败”关闭缺环境的场景。

## Native/environment approval and failure routing

- 安装/替换候选前核对已有App、运行宿主、owned会话及同bundle ID副本，保存明确可用回退副本/状态。退出/强杀仅限本次确切宿主/owned PID，保留独立CLI；不删除用户App/数据或批量扫杀。
- 保护目标、外接/网络/测试卷、权限拒绝/允许、删除/移动/断卷仅用获准专用环境，保存并按计划还原。网络共享与所需平台须真实可用；本轮不尝试枚举/修改用户环境或提前声称已具备。
- 若三种系统/Finder语言、服务启用开关、登录注册或首次/最近键场景需要状态变更，事先逐项许可、记录原状态、按约定恢复；不清整个UserDefaults域、不删除旧历史、不重置全局TCC、不默认强制重启Finder/pbs。前置票许可和以前的环境还原记录不是本轮操作许可。
- 暂缺环境/许可/可观察窗口时完成其他获准检查并明确阻塞哪项；保留整票open及相应checkbox未勾，不缩为普通目录/本机编译。发现可信缺陷报告现象、候选身份、来源requirement和所属01–12票；修正按单票审批/必要scoped TDD/独立审阅执行，候选重新交付后复验，不在13增加实现或防御抽象。
- 本票不修改发现/预算规则、不轮询、不追踪移动、不自动恢复服务/卷/watch、不增加设置中心/登录helper，不管理任意Chrome后代；不重建CI/打包系统、不取消CLI release或改变CLI网络政策。Developer ID、公证、正式干净安装和公开发布另行安排，后续标准保留。

## Acceptance

- [ ] 12的6.1–6.4完整交付并独立Standards/Spec确认，最终候选和来源/签名/架构/部署/使用说明可核对；13范围、现有套件、真实观察、受控辅助及具体安装/数据/系统操作另获批准，未把发布当执行许可。
- [ ] 7.1从最终交付包取出的同一候选实际完成Finder冷启动到默认浏览器正确内容，观察确切宿主/包内Go/目标URL，无源码、开发Go/PATH、旧扩展、Dock或终端使用前提，不以声明/进程/HTTP/测试代替真实表面。
- [ ] 目录递归、中文/空格单文件、主目录外允许目标、真正空目录及新增Markdown按实际watch状态正确；别名/并发/跨入口复用和父/子/文件独立有真实identity/PID/URL/内容证据，无固定端口fallback或文件转父目录。
- [ ] 5/6边界、6路径归一为5、取消无启动/浏览器副作用、混合有效/不支持/真实失败仍保留成功且一次汇总、全成功不提示实际观察通过；共享非模态提示期间实际Services请求不丢失，FIFO不扩张为全局报告顺序保证。
- [ ] 面板同名路径/实际状态/失败/已知降级可辨，Copy/Open/单停/Stop All/正常Quit作用于真实目标；关闭UI/浏览器不停止，浏览器失败不抹掉running/URL及主动恢复入口，正常或未知正向状态不被新增措辞标准误判。
- [ ] 最近至多20项/别名去重/recency、实际最近重开、跨启动仅恢复记录及清空不停止/跨启动为空与组件回归一致；陈旧路径有目标/原因/适用指导，用户原键按许可保存还原，旧历史未迁移或删除。
- [ ] 7.2正常退出、运行中多owned强杀宿主、ready前已有真实Go child时终止宿主、一个child意外退出及实际启动失败有配对进程/端口/状态证据；owned资源释放、其他会话/CLI按对应规则隔离，无假running、自动重启/孤儿服务或Chrome后代清扫，未将正常退出或无child的starting当强杀证明。
- [ ] 候选实际系统授权主体、受保护允许/拒绝、专用目标删除/移动/断卷、真实外接/网络卷与已知watch降级/手动刷新有对应证据，失败不误为空目录，不绕过权限/统一要求FDA，不自动追踪/重连/轮询；缺环境保留阻塞，08/3.3缺口不被代签或自动关闭。
- [ ] 最终候选指导/Help、服务关闭检查、简中/英文/不支持语言的完整native及真实Finder名称、真实登录状态与用户开启/关闭行为已核对；无登录启动仍冷启动有效，底层诊断保留、不假成功或擅自取消已有注册；S-A3仅记录非阻塞观察，不新增焦点保证或自动修改10。
- [ ] 7.3同一候选有实际Apple Silicon、Intel及最低macOS13运行覆盖，记录真实OS/硬件/执行架构/对应场景/收尾；静态或Rosetta辅助与实机结果分开，缺必需环境保持7.3未完成及13open，不新增全小版本/全组合保证。
- [ ] 相关现有Go/Swift行为套件有实际命令/结果/skip并与候选路径一致，无新永久wiring/文字/资源/拷贝/mockecho或重复内部矩阵测试；受控辅助标注、还原、清理并与最终候选证据分开，环境/用户数据按批准计划恢复。
- [ ] 全部本票必要场景、环境及证据汇总完成并经独立Standards/Spec审阅、无未决范围后才关闭13并勾选7.1–7.3；08网络/3.3仍独立按批准条件关闭，票据全部发布不等于整变更完成，未宣称正式签名/公证/干净安装/公开发布或同步/归档已获许可。

## Whole-change coverage checkpoint

[完整覆盖地图](../coverage-plan.md)逐项保留**26个OpenSpec checkbox、30个requirement、68个scenario**与全部贡献者。本轮新增最终规划票 **13-candidate-functional-acceptance**，完整映射 **7.1–7.3**，状态open、待审阅；至此既定 **01–13均已发布**，没有剩余未发布切片或未分配范围。

当前01–07、09、10 done；08 open/网络共享待补证；11、12、13 open，13的12实施前置尚未满足。账本**17/26**，3.3、5.3、6.1–6.4、7.1–7.3保持未勾。本轮只刷新地图中的当前事实，不重写既有票的历史发布快照、不修改批准artifacts或任务进度，也不授予实施、环境操作或正式分发许可。
