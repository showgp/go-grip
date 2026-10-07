# 08 — 原生目标访问、最小权限与真实授权验证

Status: done
Blocked by: 07-finder-services-preview
Covers OpenSpec tasks: 3.3
Behavior source: ../specs/macos-finder-service/spec.md, ../specs/macos-preview-sessions/spec.md, ../specs/macos-menu-bar-app/spec.md, ../specs/macos-app-packaging/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

通过 07 完成后的真实 Finder Services 和手动入口，在同一开发 App → 唯一批次/协调器 → 内置 managed Go 路径上验证普通目标、受保护目录及外接卷的实际访问；已挂载网络卷的实际访问与降级证据经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），不作为本票完成条件。允许访问时默认浏览器显示对应目标的正确内容；访问或授权被拒绝时，原生反馈保留目标及实际可获得原因，并给出适用的文件/卷或系统授权检查指导，不能显示为空目录、假 running 或用权限提升隐藏失败。记录系统实际归属的授权主体，不能从父子关系、签名或 Terminal 下 CLI 可读推断 native 路径已获准。

本票是批准的早期权限 gate，不是只写一段 FDA 提示，也不是预建权限管理框架。复用 01–07 的真实工具、身份、批次、访问状态、默认浏览器及 owned 收尾；只有实际观察发现 3.3 的行为缺口时才作最小实现修正。若证据否定直接内置工具路线，保留集成阻塞并请求设计裁决，不擅自切换 helper、detach 或强制 FDA。完整面板、帮助/本地化及候选包由 09–13 交付。

01–06 已完成并通过对应审阅；用户已确认 07 并报告正在实施，尚未取得本轮完整验收/关闭证据。07 是本票真实直接前置。**本轮只发布 08 供审阅，不授权实施、编写测试或改变权限/挂载/用户安装环境。** 实施前须另行批准本票范围、scoped verification/TDD seams、实际目标与环境变更以及实施许可；07 完成并通过独立审阅后才能开始。

`tasks.md` 是唯一进度账本。只有本票按调整后范围全部交付（已挂载网络卷证据延期至后续 change，不作为关闭条件，见 [tasks.md](../tasks.md) 延期清单）、必需环境证据及独立 Standards/Spec 审阅通过、没有未解决路线裁决后，才可关闭 08 并按调整后范围勾选 **3.3**；关闭按本票自身开发构建标准核对既有贡献，不因 13 使用新候选要求重做本票验收。不凭发布或普通目录成功勾选，不改 3.1/3.2/3.4、任务 4–7 或既有票状态；不自动启动 09。

## Source mapping

下表保留批准 requirement/scenario 原名，只限定本票贡献。票面不新增授权 API、恢复机制或平台保证；真实候选及完整面板场景仍有后续贡献者。

| Requirement / scenarios | 08 贡献及剩余边界 |
|---|---|
| macos-finder-service — Supported open targets：`Receive a Markdown file`、`Receive a directory outside the home folder`、`Receive an unsupported file` | 真实普通/受保护位置及外接卷的目录与常规大小写不敏感 .md 访问验证，沿用已有运行时分类与拒绝规则；不把权限允许范围误写为所有位置都保证可读。已挂载网络卷维度经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单）。类型/批次由 04–07，候选由 13。 |
| macos-finder-service — Minimum necessary file authorization：`Access is denied`、`Open an ordinary accessible target` | 本票主要行为：观察 native→Go 实际授权主体、允许/拒绝结果和适用检查指导；普通可访问目标不先要求 FDA/辅助功能/自动化，不绕过拒绝。真实系统权限证据不能由 chmod 或 mock 替代；候选由 13。 |
| macos-finder-service — Partial batch failures：`Some targets fail`、`All targets succeed` | 已观察到的访问失败仍走 06 的一次原生汇总，保留其他成功会话；只补实际访问/指导缺口，不重复 06/07 批次状态矩阵。候选由 13。 |
| macos-preview-sessions — Directory and single-file preview modes：`Preview nested documents`、`Preview a selected file`、`Keep an empty directory session`、`Add Markdown to a watched empty directory` | 用可辨认的实际内容和真实空目录对照证明访问成功与失败不混淆；沿用现有目录递归/单文件/监视政策。完整文件事件矩阵由 13，不因网络卷新增扫描或恢复。 |
| macos-preview-sessions — Verified startup and actual preview URL：`A nondefault port is used`、`Open a file with spaces and non-ASCII characters`、`Startup ends without valid readiness information` | 访问成功使用生产验证的实际 URL；启动访问失败无假运行/默认端口 fallback/本次遗留 child。复用 03–07 readiness，不新建探测或重试；候选由 13。 |
| macos-preview-sessions — Unavailable paths are not empty directories：`A target is moved or deleted`、`A mounted volume becomes unavailable` | 对专用测试目标/卷观察实际不可访问反馈与既有结构化状态，原路径不迁移、不后台重连、不误报空目录；保留真实会话管理和停止入口。完整面板状态呈现由 09、候选矩阵由 13，既有 Go 状态由 02/04。 |
| macos-preview-sessions — Visible hot-reload degradation：`Watch coverage is incomplete`、`Use a network-volume preview` | 不因卷类型拒绝访问的行为要求保留，但已挂载网络卷维度（`Use a network-volume preview`）经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），证据由后续 change 承接；已知监视失败/不完整保留真实 degraded 原因及手动刷新能力，不声称网络卷总能自动刷新。既有状态机制由 02/04，完整面板由 09，完整故障矩阵由 13；不强行制造所有 watch 错误。 |
| macos-preview-sessions — Explicit stopping confirms service termination：`Stop one of several sessions`、`Stop all application sessions`；Independent CLI operation remains available：`Launch the renderer without the application` | 本票实际测试目标/卷上的 owned 服务可正常管理、停止及释放 PID/监听；独立 CLI 不受影响。复用已有停止和所有权，不重做 03–07 内部矩阵，候选由 13。 |
| macos-menu-bar-app — Manual opening uses the same target contract：`Manually reopen a Finder-started target`、`Choose a new target manually`；Action failures remain understandable and recoverable | Finder/手动沿用 07 共享入口及同一授权/访问失败反馈，原生提示和可查看报告含实际目标、原因与适用指导，不把访问失败隐藏在浏览器成功请求之后。不实施最近记录；完整面板/最近场景由 09，本地化由 10。 |
| macos-app-packaging — Universal self-contained application：`Use the application without development tools`；Functional acceptance exercises actual user paths | 本轮只验证确切开发 App 的内置工具/native 访问路线；不以 Terminal/源码/PATH 工具访问或签名检查充当授权证明，不宣称候选 App/DMG、Intel/macOS 13 或正式分发验收。12/13 保留完整要求。 |

技术依据为 [design Decisions 1–4、6–8、10、Risks / Trade-offs 与 Migration Plan](../design.md#decisions)：后台准备、实际 Go 访问最终判定、非沙盒正常父子 launch、真实 TCC 主体验证、已有状态与错误报告、最小权限和早期路线 gate。系统权限背景见 design 已引用的 [Apple 文件访问授权说明](https://support.apple.com/guide/security/controlling-app-access-to-files-secddd1d86a6/web)；它不证明当前构建/目标已获授权。

## Scope and implementation handoff

发布时 05/06 已有 `PreviewAppModel.openBatch`、`PreviewSessionCoordinator.prepareBatch/open(prepared:)`、`TargetPreparation`、`FoundationManagedProcessFactory` 与原生报告/会话控制。07 正在将 Finder provider 注册在生产根依赖就绪之后；目前存在相关源码不代表冷启动/权限验收已完成。实施 08 前重新读 **实际完成的 07**、其确切运行包及验证记录，不覆盖正在修改的 07、不锁定未来内部 API。

- **唯一实际路径**：使用已完成 07 的 Finder Services 和现有 NSOpenPanel，进入同一 `openBatch`/协调器与 `Contents/MacOS/go-grip`。保留 bundle ID、正常父子启动及非沙盒路线，不用 Terminal、独立 CLI、临时 Foundation 拥有者或直接 selector 调用替代 native 授权证明；这些仅可作为明确标注的辅助诊断。
- **最小访问与适用指导**：普通允许访问的目录/文件直接打开；失败仍保存所选目标、实际可获得原因及适用检查方式。区分可观察的文件/目录权限、卷挂载/共享权限和系统 Files and Folders 等授权情况；无法确知原因时保留真实诊断并给条件式检查指导，不能把所有 EPERM/EACCES、缺失路径或卷故障都断言为 TCC，也不能把所有失败统一导向 FDA。不新增自动设置权限、申请整盘访问、权限诊断数据库或配置界面；简短原生提示/既有报告及 README 指导足以满足时不另建服务。
- **两阶段访问事实**：后台元数据准备可能先拒绝，也可能允许准备后由 Go 在实际读取时失败；二者都纳入同一失败报告，不用元数据成功保证后续可读。允许目标的实际 HTTP/浏览器内容才是成功证据，NSWorkspace true、进程已创建或空正文不是证明。真实可访问空目录成功与不可访问目录失败须有对照。
- **保留既有状态与恢复边界**：启动失败不保留假 running/URL 或本次未收尾 child；运行中的 Go/目标不可访问仍按真实 phase、target/reload 状态及既有错误表面处理，不仅删行、不自动重启/重连/移动跟踪。用户改变访问许可后可主动重开原路径；仍在运行的服务按既有复用规则处理，不另建授权重试循环，不承诺原 watcher 自动恢复。完整面板布局/状态展示留给 09，不能据此省掉本票访问错误和指导。
- **实际授权主体**：记录确切 App bundle/宿主与内置工具路径、构建/代码身份、实际入口、所选/规范化目标以及系统提示、系统设置或可获得系统诊断中实际可观察的授权对象。仅有签名 identifier、父 PID 或 HTTP 成功不足以单独证明 TCC 归属；未能观察的部分明确为未知并保持对应验收未完成，不把已允许的测试位置冒充受保护拒绝。ad-hoc 重建可能改变身份/重新请求许可，不保证跨构建授权稳定。
- **环境与用户数据安全**：优先使用专用可丢弃目标及明确外接卷（已挂载网络卷环境不属本阶段，其证据经批准延期至后续 change，见 [tasks.md](../tasks.md) 延期清单），记录卷类型/挂载位置与当前访问状态。创建/挂载/卸载测试卷、调整专用目标权限、改变系统授权或安装位置前取得具体许可；不改用户已有文档权限、不操作共享工作卷、不重置全局 TCC/索引、不授予 FDA/辅助功能/自动化、不自动更改服务或登录项。不要求用户把开发工具的授权当作 App 的许可。
- **路线被否定时**：若真实普通/获准目标仍因直接内置 Go 路线无法访问，保留原始 native/系统/Go 证据与失败目标，停止对应集成并请求重评 design。不得自行添加 helper/XPC/launchd、detach、sudo 或 FDA 兜底；也不得把实际路线问题藏到 12/13 后再勾选 3.3。
- **文档及调用切换**：只作 3.3 所需最小修正，复用现有报告、唯一工程/test target和开发构建入口；同步更新 `README.md` 的访问/授权/卷检查及用户主动恢复说明、`docs/ARCHITECTURE.md` 的实际授权主体/观测与路线边界。不改 Go 发现/预算/忽略政策、独立 CLI 网络政策或用户旧 50 项数据。

## Scoped verification and proposed TDD seam

发布不授权写测试。本票的核心证明在真实 macOS/native/文件系统 seam；不能预先凭猜测给它加权限管理抽象或一组 mock TCC 测试。

复用 03–07 已批准的目标分类、批次部分失败、浏览器结果、无假 running、退出准入及 owned 收尾回归。若实际观察需要新增行为代码，实施前单独确认唯一最小消费者 seam：**真实访问失败进入共享报告时，目标/可获得原因及适用检查动作不丢失，不改变已成功会话事实和管理能力**；只验证本次新增分类/可执行指导行为，能由既有回归覆盖则不重复新增。指导若只是说明文字，用实际原生表面和文档操作验证，不新增文案/源文本测试，不为测试新建指导模型或泛化权限服务。

新增永久测试必须 trace 到 3.3 和上述 approved requirements，经实施前批准后再做 red/green；不重复 03–07 的状态矩阵，不测 plist、字符串存在、签名拷贝、默认值、字段转发、mock errno echo、非空 reason 或仅 isRunning。临时 POSIX 权限 smoke可补确定性访问失败，但不是系统授权主体/TCC 拒绝或外接/网络卷证明。

## Real verification

实施前确认本轮可操作环境和专用目标；下表是必需实际观察，不是已完成的记录。每次记录入口、确切 App/工具、展示及规范化目标、当前访问许可/系统观测、实际 URL/内容或失败报告，以及 owned 资源的最终收尾。

| 场景 | 必须观察的结果与证据边界 |
|---|---|
| 普通可访问本地目录、中文/空格 `.MD`、真实空目录 | 经 Finder/手动共享 native 路径直接处理，无统一权限提升前置；对应默认浏览器/实际 HTTP 内容、完整文件 URL和空状态正确，空目录不是拒绝访问的替代样本。复用已有身份/批次，不重做其内部矩阵。 |
| 受保护目录的允许与拒绝 | 在明确记录授权状态的专用目标上走真实 native→内置 Go，记录系统实际授权主体及可获得允许/拒绝证据；允许后有正确页面，拒绝后原生报告含目标/真实原因/适用指导且无假运行/假空目录。仅 chmod、mock failure、当前 App 已获准而直接成功或 Terminal 可读都不替代这组证明。授权变更必须先许可，不重置全局状态。 |
| 外接卷上的目录/Markdown | 使用实际外接卷/确切挂载路径和普通可读内容，观察 native→Go 正确页面及已观察拒绝/不可访问原因的指导；不以系统卷临时目录、符号链接或目录名模拟外接环境。专用测试权限变化不得影响其他数据。 |
| 已挂载网络卷上的目录/Markdown（**经批准延期**） | 该维度经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），仍为未验证的延期验收项，但不阻塞本票及 3.3 按调整后范围完成；延期不代表验证通过，不得以开发构建或本地卷结果替代其实际运行证据。 |
| 专用目标/测试卷变为不可访问 | 在已获准操作的专用环境，触发实际请求并观察原路径访问失败/结构化 unavailable 而非空目录，既有服务仍有真实管理入口；仅在确实安全且获准的测试卷做断卷，不卸载用户工作卷。全套删除/移动/watch 故障矩阵由 13，完整面板状态呈现由 09。 |
| 错误指导、主动恢复与隔离收尾 | 在 native 错误表面与可查看报告确认指导可操作、原因不过度归类；经用户主动许可/修正访问后重开对应目标，页面正确且沿用共享身份/复用规则。停止/退出后本轮 owned PID/端口释放，其他目标按单停规则继续，独立 CLI仍提供自己的内容；不复用旧 05–07 PID 充当本轮证明。 |

运行受影响 Swift 回归、现有真实 Go 集成和实际开发 App 构建，记录命令/结果/skip；优先复用 `make macos-test` 与已完成 07 的开发构建/签名入口。Go 无改动不为确认已知 errcheck 重跑 lint，不宣称已修复。签名检查只证明构建身份/结构，不证明 TCC 访问；交叉编译不证明最低系统/Intel 运行。

缺少实际受保护拒绝、授权主体、外接卷、可操作 native/浏览器或必要许可时，明确缺少哪一项证据并保持相应验收及 3.3 未完成（已挂载网络卷证据仍未验证，但不阻塞本票及 3.3 按调整后范围完成），不能缩成只测普通 `/tmp`、chmod 或文档后关闭。辅助证据与系统实际观察分别标明；不能因尚无环境而预埋替代架构。

## Prerequisites and non-goals

唯一直接 blocker 为 [07-finder-services-preview](07-finder-services-preview.md)：实际完整 Finder/手动原生入口与确切运行 App 必须先完成、验证并通过独立审阅。05/06 已完成间接提供 App/批次/浏览器/退出；01–04 提供 Go/状态/身份/所有权。不复制或提前补做 07，不把其进行中的源码当作已通过的前置。实际专用保护目标/卷和必要用户许可是验收条件，不凭发布提前声称可用。

不做新历史/完整面板（09）、首次引导/完整中英文（10）、登录项（11）、universal/archive/DMG/CI/候选及支持环境矩阵（12/13）；不扩展为所有 macOS 权限类别管理、自动权限申请/重设、移动跟踪、卷重连、watch 恢复/轮询、后台重试、进程名/未知 PID/Chrome 后代清扫、helper/XPC/launchd/sudo 或签名公证路线。提交/推送、主规格同步、归档及正式分发均未授权。

## Acceptance

- [x] 实施前核对07整票完成/实际native验证及独立审阅，另行批准08范围、最小scoped verification/TDD seam、专用目标/权限与卷操作许可及实施；本轮发布不提前编码/测试或解除真实前置。
- [x] Finder及手动入口沿用07同一生产批次/协调器/正常父子launch与Contents/MacOS/go-grip，普通允许目标无需FDA/辅助功能/自动化，元数据与真实读取不阻塞主线程，不新增第二套身份/权限运行模型或PATH/源码/Terminal替代路径。
- [x] 实际native→Go观察受保护目录允许与拒绝，记录确切App/工具/目标/许可状态及系统实际授权主体；允许后正确浏览器内容，拒绝后明确目标、可获得原因与适用检查指导；签名、父子关系、chmod或mock不能替代TCC/native证据，未知主体不记为通过。
- [x] 真实外接卷的目录/Markdown在系统允许范围内可从native路径读取正确内容，不因主目录之外或卷类型统一拒绝；记录实际挂载权限与观察到的不可访问指导，不用系统卷临时目录或别名冒充实际卷环境。已挂载网络卷的实际访问/降级证据经批准延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），不属本票关闭条件。
- [x] 原生访问失败报告与实际事实一致，包含所选目标/可获得原因/适用指导，不把所有失败猜成TCC或统一要求FDA；准备阶段与Go真实读取失败均走共享反馈，其他成功会话继续可管理，真正空目录成功与拒绝/不可访问失败有实际对照，无假running/默认端口fallback/本次遗留child。
- [x] 仅在获准专用目标/卷上观察原路径实际不可访问与既有unavailable/degraded反馈，保留管理/停止及手动刷新，不迁移路径、自动重启/重连或承诺网络卷完整热重载；完整面板由09、完整候选故障矩阵由13，不据此省略08访问错误/指导。
- [x] 按实际指导经用户主动允许/修正访问后重开得到对应正确页面，遵循既有身份/复用规则；本轮停止/退出确认owned PID/端口释放且独立CLI保持自身内容，不仅删行/发信号，不引用旧05–07资源结果冒充08。
- [x] 环境/授权/卷或用户安装变更逐项取得许可，仅使用专用可丢弃目标，不动用户现有权限/数据或工作卷、不重置全局TCC/服务索引、不自动授予FDA/辅助功能/自动化；若证据否定内置工具路线则保留集成阻塞并请求设计裁决，无detach/helper/权限提升掩盖。
- [x] 受影响Swift回归、真实Go集成及开发App构建有实际命令/结果/skip，必要新行为只按获准消费者seam做red/green并复用既有回归，无字符串/wiring/mock echo等永久测试；缺实际主体/保护拒绝/外接卷或native环境时明确未完成（已挂载网络卷证据仍未验证，但不阻塞本票及 3.3 按调整后范围完成），不以静态检查或只测普通目录关闭3.3，不重跑已知errcheck。
- [x] 随本轮证据更新README访问/授权/卷检查及主动恢复操作、架构中的实际授权主体/身份限制/路线gate和未验证边界，清理本次临时目标/确切owned资源并按获准计划恢复专用测试环境；不改旧50项用户数据、Go发现或独立CLI网络政策。
- [x] 全部08验收（按调整后范围：已挂载网络卷证据延期至后续 change，不作为关闭条件）及独立只读Standards/Spec审阅通过、必需环境证据齐备且无未解决路线/范围裁决后才关闭08并按调整后范围勾选3.3；其他票/checkbox不变，不自动实施09、提交/推送、同步主规格、归档或正式分发。

## Change-wide coverage reference

[coverage-plan.md](../coverage-plan.md)逐行覆盖完整26任务、30 requirements、68 scenarios，标明01–06已发布done、07已确认且正在实施（未验收）、本轮08已发布open待审阅且被07阻塞、09–13仍为未发布规划。既有票中的发布时状态保留，不据其历史描述否认已经完成的05/06，也不由进行中的07源码推断其完成。本轮不修改批准的proposal/specs/design/tasks或既有tickets。

注（2026-10-07 范围调整，新增）：已挂载网络卷证据延期至后续 change（见 [tasks.md](../tasks.md) 延期清单），仍为未验证的延期验收项；本注仅适用于此次范围对齐，不改写上方发布时快照与历史执行记录；本票当前关闭条件以 Acceptance 调整后文本为准。

## Execution record — 2026-10-05（开发构建轮次 A–F；2026-10-07 证据持久化与复审注记）

**范围与身份**：本票按自身**开发构建标准**执行（07 final ad-hoc 构建：宿主 CDHash `5db6f61d…`、包内工具 `0767a836…`；macOS 27.0.1 arm64）；候选级验收由 13 承担，本票不要求。已挂载网络卷证据经批准延期至后续 change（**未验证**，不属本票完成条件）。

**执行与结果（Rounds A–F，逐轮 PASS）**：
- **A 普通目标冷启动**：Finder 冷启动宿主 48759（PPID=1/Finder，UIElement），一个 Finder 调用产生三个 owned child——真实空目录（56360，“No Markdown” 空状态）、递归目录（56369，含 ordinary markers）、中文/空格 `.MD` 单文件（56374，单文件模式）；全部仅 `127.0.0.1:<随机>`，无 6419/默认端口 fallback。
- **B 受保护目录拒绝→允许→重开**：tccd 授权主体＝app bundle `com.showgp.GoGrip`（pid 48759）；拒绝 Create 记录 17:25:09（Bundle ID）；包内工具访问经 responsible/app 主体评估（accessing `com.showgp.GoGrip.go-grip` pid 49036）；拒绝后无遗留 child/监听；System Settings 允许后（Modify 17:26:38）新 child 49155 → 56629 正确内容。
- **C 外接卷**：提示主体＝app bundle；允许后（Create 17:50:56）child 49744 → 57392，真实外接 USB/APFS 卷内容正确，不因卷类型统一拒绝。
- **D 不可访问**：改名/删除 + 一次性磁盘镜像（`/Volumes/GoGrip08Test`）断卷；三个受影响会话显示访问错误而非空目录，真实空目录仍为空状态，面板保留会话与 Stop；无自动迁移/重启/重连。
- **E 停止/退出隔离**：单停仅释放对应 child 50782/端口 57959；Quit 释放宿主与全部五个 owned child（端口 56360/56369/56374/56629/57392 全关）；独立 CLI（pid 51060，`*:6419`）全程 200，不受面板/停止范围影响。
- **F 新构建复验**：ad-hoc 重建改变 CDHash（`02d35053…`→`5db6f61d…`）并重新请求 Documents/RemovableVolumes 授权（身份限制如实记录）；逐字拒绝报告含目标＋可获得原因＋条件式授权检查路径（System Settings → Privacy & Security → Files and Folders）；允许后重开成功；外接卷类同。
- **指导修复（经批准的最小 seam）**：red 81/2（两个断言失败）→ `PreviewSessionCoordinator.swift` 追加条件式检查句 → green 81/0；`make macos` Release `BUILD SUCCEEDED`；`codesign --verify --deep --strict` OK。基线 80/0。

**独立审阅**：2026-10-05 Standards/Spec 各 0 阻塞（各 1 P2 建议 → README 修正后增量复核 0/0）；2026-10-07 独立两轴复审（`ReviewStandards08R2`/`ReviewSpec08R2`）：各 **0 阻塞、0 建议、0 待裁决**；覆盖限制如实披露（无 08 单独 git 固定点、未重跑实施者检查、网络卷延期不作为阻塞）。未发现未溯源行为或测试。

**证据持久化（2026-10-07）**：`/tmp/gogrip-08/EVIDENCE-LOG.md`（2026-10-05 18:53；sha256 `b1f743750e32379db15ac33b369bec70b26799a423659786b9a91c723ee3703d`）已**逐字复制**至 `~/GoGrip-acceptance-13/state-08/EVIDENCE-LOG.md`（sha256 一致）；来源/时间/边界见 `state-08/PROVENANCE.txt`。原始内容、当时构建身份与当时范围保留，未改写；原日志文本中“网络卷未验证、3.3 不能按该证据关闭”为当时记录，按 2026-10-06 批准的范围调整，网络卷延期后不阻塞本票与 3.3 关闭。

**关闭边界**：已挂载网络卷维度保持**未验证**（延期至后续 change）；本记录不表述候选级/正式分发已验收。

## Closure record — 2026-10-07

按调整后范围关闭本票：开发构建轮次 A–F 证据已持久化（`~/GoGrip-acceptance-13/state-08/EVIDENCE-LOG.md` 逐字副本＋`PROVENANCE.txt`，sha256 一致）；独立两轴审阅通过（2026-10-05 首轮 + 2026-10-07 复审 `ReviewStandards08R2`/`ReviewSpec08R2`：各 0 阻塞/0 建议/0 待裁决）。已挂载网络卷维度保持**未验证**（延期至后续 change）；审阅限制（无本票单独 git 固定点、未重跑实施者检查、网络卷未观察）如实保留于执行记录与两轴报告。本票 `Status: done`；Acceptance 按调整后文本全部勾选；tasks **3.3** 已勾。候选级验收由 13 承担；本关闭不表示正式分发（Developer ID/公证/干净安装/公开发布）或主规格同步/归档已获许可。
