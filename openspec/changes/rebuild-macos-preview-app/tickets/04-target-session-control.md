# 04 — 规范化目标会话协调与真实停止

Status: done
Blocked by: 03-owned-process-launch
Covers OpenSpec tasks: 2.2, 2.4 (coordinator contribution; completes 2.4 together with 03)
Behavior source: ../specs/macos-preview-sessions/spec.md, ../specs/macos-finder-service/spec.md, ../specs/macos-menu-bar-app/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

由唯一主 actor 协调器接收目标，后台准备规范化身份与目录/Markdown 文件类别，通过 03 的生产 Foundation 适配器启动真实 Go；相同目标并发启动单飞、运行中复用实际 URL、停止期间重开等待旧进程真实结束，不合并父目录、子目录与文件。协调器维护带启动代次的会话事实及可获得失败原因，显式停止单个/全部包括 starting 的 owned 会话并确认真实退出，不自动重启、不影响独立 CLI。

本票交付可运行的协调器→生产 ManagedProcess→真实 Go/HTTP→真实退出切片、批准 seams 的行为回归及架构说明。临时 Foundation 拥有者可以调用生产协调器证明该切片；不以只新增会话字典、mock 启动次数或状态枚举冒充实际交付，不宣称已经接通 Finder、AppKit 正常退出、默认浏览器、完整面板或候选 App。

发布仅供审阅，不授权编写测试或实施。实施前仍须明确确认本票范围、scoped TDD seams、实际检查与 smoke 方法，并另行取得实施批准。`tasks.md` 是唯一变更进度账本；本票 checkbox 仅为验收细节。整票验证及独立 Standards/Spec 审阅通过后才可勾选 2.2，并结合已完成 03 的进程层证据勾选完整 2.4；不重新勾选或重开 2.1、2.5，不检查 2.3 或任务 3–7。

## Source mapping

行为以以下批准 specs 为准，不以旧管理器、票据、测试或日志代替。每行只发布本票贡献；完整原生能力与候选证明仍需后续票。

| Requirement / scenarios | 本票贡献与完整能力边界 |
|---|---|
| macos-preview-sessions — One session per normalized target：`Reopen a target through a symbolic link`、`Receive concurrent requests for the same target` | 标准化绝对路径并解析符号链接，保留所选路径用于展示；同身份 starting 请求等待同一结果，running 请求复用实际 URL，不启动重复服务。再次请求默认浏览器及跨入口效果由 05、11 验证。 |
| macos-preview-sessions — Containment does not merge target identities：`Open a child directory and a file under an active parent` | 三个不同规范化路径分别拥有真实会话，子目录有自身浏览根，文件保持单文件模式；不能用父目录已递归覆盖作为复用依据。候选操作由 11 验证。 |
| macos-preview-sessions — Explicit stopping confirms service termination：`Stop one of several sessions`、`Stop all application sessions` | 协调器显式停止单个/全部，仅管理自己记录的 owned 实例；包括 starting、重复停止和停止期间重开，真实退出前不删记录或宣称 terminated；复用 03 writer close、4 秒 exact-owned SIGKILL 与真实失败结果。原生动作由 05、11 验证。 |
| macos-preview-sessions — Services end with their owning application：`Quit the application normally`、`Force termination of the host`、`Host exits during startup` | 提供可供退出路径调用的全部会话收尾，包括 starting 的本票贡献；沿用 03 的异常 owner loss 机制，不另建所有权方案。AppKit 延后退出、退出时禁止新启动及完整 App quit 接线由 05 完成，候选正常/异常退出由 11 验证，不能把本票临时拥有者退出当作完整 App quit。 |
| macos-preview-sessions — No automatic session restoration or restart：`Relaunch with recent targets`、`A preview process exits unexpectedly` | 协调器不持久化或恢复运行状态；意外退出结束 running、保留可获得原因，不自动启动新代次；用户显式停止不被标为意外退出。最近目标持久化/主动重开入口由 07、11 验证。 |
| macos-preview-sessions — Unavailable paths are not empty directories：`A target is moved or deleted`、`A mounted volume becomes unavailable` | 使用建立时确定的路径，消费真实 target-status，目标失效不变成空目录、不删除仍存活服务的管理入口；保留原因与停止能力，不移动跟踪或重连。实际卷/TCC 与原生显示由 06、07、11 验证。 |
| macos-preview-sessions — Visible hot-reload degradation：`Watch coverage is incomplete`、`Use a network-volume preview` | 将启动结果及运行期 reload 状态保留在唯一事实源，降级不等于服务退出，也不影响可用 URL 与停止；不解析日志、不增加监视恢复/轮询。真实网络卷、面板降级/手动刷新及候选效果仍由 06、07、11 验证。 |
| macos-preview-sessions — Independent CLI operation remains available：`Launch the renderer without the application` | 协调器单个/全部停止及失败收尾只作用于自己拥有的 managed 实例；真实独立 CLI 继续提供其原目标内容，参数/JSON/网络政策不改。候选隔离由 11 验证。 |
| macos-finder-service — Supported open targets：`Receive a Markdown file`、`Receive a directory outside the home folder`、`Receive an unsupported file` | 提供入口共用的目标准备：目录和大小写不敏感 `.md` 文件、拒绝其他文件而不改为父目录，不因位置在主目录外而拒绝。Finder 接收/原生报告属 05，实际外接/网络卷与权限 gate 属 06，完整结果属 11。 |
| macos-finder-service — Batch deduplication and quantity confirmation：`Open exactly five distinct targets`、`Confirm six distinct targets`、`Cancel a large batch`、`Count aliases only once` | 仅交付可供 05 使用的一致规范化身份；不实施批次执行、5/6 确认、取消或汇总。完整批次规则由 05、11 验证，本票不能勾选 2.3。 |
| macos-menu-bar-app — Visible sessions and explicit controls：`Distinguish targets with the same name`、`Copy an active preview address`、`Exit through the panel` | 提供目标展示路径、实际 URL/阶段/失败原因及真实单个/全部停止事实，不为 UI 另建会话字典。基础原生控制/退出属 05，完整面板/剪贴板与候选操作属 07、11。 |
| macos-menu-bar-app — Manual opening uses the same target contract：`Manually reopen a Finder-started target`、`Choose a new target manually` | 提供入口无关的目标身份及会话复用能力；不实现 NSOpenPanel/Finder/最近入口。05、07、11 继续验证实际跨入口行为。 |
| macos-menu-bar-app — Action failures remain understandable and recoverable：`Open a stale recent target`、`Recover from browser opening failure`、`A service exits unexpectedly` | 本票提供启动/停止/真实退出失败事实及可获得原因；不把失效或已退出服务标为 running。原生提示、浏览器错误及历史恢复由 05、07、11 完成；不提前实现浏览器或历史功能。 |

技术依据为 [design Decisions 1、3、4、7–8](../design.md#decisions)：macOS 13 可用的主 actor/ObservableObject 事实源、后台元数据准备、规范化身份、明确阶段和代次隔离、真实停止及结构化状态。Decisions 5–6 的协议/readiness 已由 03 实现，本票复用，不再建立解码器、期限、健康探测或日志状态解析。

## Scope and implementation handoff

路径相对于仓库根；新内部类型名和最小测试 seam 由实施时结合现有代码选择，不要求额外框架或第二份状态模型。

| 当前表面 | 本票交付与切换边界 |
|---|---|
| `macos/GoGrip/Services/ManagedProcess.swift`：一次 launch/一个 owned child；`start()`、`stop()`、带 generation 的 `ManagedEvent` 与真实结果 | 作为唯一生产进程 seam；协调器每次实际新启动分配 generation，拥有适配器并消费结果/事件。单飞、同目标复用和会话字典在协调器，不下放到适配器、不另写 Process launch。 |
| `macos/GoGrip/Services/ManagedProtocol.swift`：结构化 reload/target 快照 | 沿用 02/03 已确认形状与状态，包括已有 disabled 快照；不以 stderr/空内容推断目标可访问性或热重载覆盖，不重新要求宿主校验 ready 前 runtime 事件顺序。 |
| 规范化和目标准备，目前无新协调器 | 在后台标准化绝对 file URL、解析符号链接并分类，保留所选展示路径；身份使用解析后的路径，不使用 inode/bookmark 跟踪移动，不笼统转换路径大小写。元数据失败返回目标及可获得原因，不创建假 running；Go 实际访问仍是最终判定，预检查不承诺后续访问成功。 |
| 唯一主 actor 会话记录及异步回调 | 保留 target/mode、展示路径、generation、执行阶段、实际 URL、目标可访问性、reload 状态及可获得失败原因所需事实；使用 macOS 13 可用的 ObservableObject/Published。阻塞元数据、管道与 HTTP 不在主线程执行；入口/UI 不再建立自己的运行事实源。 |
| 同目标 starting/running/stopping | starting 等待同次启动；running 返回已有已验证 URL；stopping 的重开等待旧 owned child 真实结束后才允许新 generation。不同父/子/文件目标独立；失败或停止不能使尚存活的旧服务与新代次重叠。 |
| 启动、状态、退出与停止结果 | 所有结果应用前按目标身份与 generation 判定是否仍属当前记录；旧代次启动完成/状态/退出回调不覆盖新记录。启动成功不能把已经 stopping/terminated 的同代次恢复为 running；无有效 URL/已退出不能成为成功会话。 |
| `ManagedProcess.stop()` 的 no-child 结果及真正 owned 退出 | 协调器保留 starting 的停止意图：可阻止尚未 launch 的启动，或在 child 已创建后通过同一适配器收尾；不能把 no-child 结果误认为清理完成而随后遗留服务。重复停止等待同一真实收尾，不重复创建服务；真实停止失败保留实际失败和管理事实，不以删行伪装退出。 |
| target/reload 与执行阶段 | unavailable/degraded 与服务阶段分开；运行服务仍可查看、取得已有 URL和停止。意外退出结束 running 并保留可获得原因；显式停止不报意外退出，不自动重启、恢复 watcher、迁移路径或重连。 |
| `ProcessManager.swift`、`AppDelegate.swift`、旧 Views/Storage/HistoryEntry | 本票不把旧 port/path API 包装成新协调器，不新增 aliases/shims；旧宿主调用方及旧管理器整体切换、移除在 05，旧历史切换在 07。04 临时宿主直接调用新生产协调器，不因旧 UI 未接通而要求先实施 05。 |
| `macos/GoGrip.xcodeproj/project.pbxproj`、现有 GoGripTests 和 `Makefile` 的 `macos-test` | 将本票生产源码与批准 seam 的行为回归接入既有唯一工程/test target；复用真实 Go 测试工具入口及系统卷临时 bundle，不创建第二套工程、生产 PATH/Resources 回退或测试专用 Go flags。 |
| `docs/ARCHITECTURE.md` | 按实际证据更新规范化身份、事实源、代次/阶段、停止与正常/意外退出语义；说明临时协调器拥有者证明和仍未接通的原生/权限/候选边界。 |

### Proposed scoped TDD seams and real verification

以下只提出本票待实施前确认的 seams，不授权先写测试；测试只捕获批准的独立消费者回归，不逐字段验证复制、mock echo、转发、UI 文案、默认值或资源存在。

- 目标/身份 seam：隔离真实目录、所选 `.MD` 文件、符号链接与不支持文件，证明别名归一且保留展示路径、不同包含关系不合并、错误不转成父目录预览；不为每个路径拼写重复同一回归。
- 协调器生命周期 seam：在最小受控启动/退出 seam 上确定性覆盖并发同目标共享结果、旧 generation 回调不覆盖新会话、停止期间重开等待真实收尾；断言可观察会话/URL及是否产生新的服务，而非固定内部集合或方法转发。
- 停止/故障 seam：starting 停止、重复停止、实际停止失败与正常/意外退出区分，验证停止结束前的事实及结束后不自动重启；复用 03 的进程层覆盖，不复制其完整解码/readiness/4 秒矩阵。
- 真实验证：临时 Foundation 拥有者调用生产协调器→生产 ManagedProcess→当前 Go，记录规范化目标、generation、PID、实际 URL/内容与退出；真实同目标/符号链接复用、父/子/文件独立、单个/全部停止、starting 停止和独立 CLI 隔离不能仅由受控 seam 或旧 03 记录代替。

## Prerequisites and non-goals

唯一直接前置为 [03-owned-process-launch](03-owned-process-launch.md)，发布时已 done，并已完成最后一轮零剩余 Standards/Spec 复核及本地提交；其依赖 01/02 均已满足。03 的生产启动、结构化事件、真实退出、4 秒兜底、FD 释放与 owned/CLI 隔离可复用；当前没有未满足的前置票。

04 是 2.4 的剩余完整协调器贡献：结合 03 已完成进程层证据后，本票须真正验证单个/全部、starting/重复停止、停止期间重开及正常/意外退出，才能勾选完整 2.4。05 仍负责 AppKit 正常退出时禁止新启动、调用全部收尾并延后回复退出；不因 04 完成而宣称完整 App quit/Finder/浏览器或 requirement 已全部完成。

明确不做：2.3 批次/数量确认/取消/汇总；AppKit 应用根、Services、NSOpenPanel、默认浏览器/剪贴板、完整面板及旧链路切换；最近存储/迁移、登录项、本地化/引导；真实 TCC/外接或网络卷验收；新 Go 行为、扫描政策、日志解析、健康轮询、自动重启/重连/移动跟踪；helper/XPC/launchd、进程名/未知 PID 清扫、Chrome 后代管理、渲染器重构；universal/archive/DMG/CI、正式分发、提交/推送、主规格同步或归档。批准 artifacts 有缺口或新范围争议时先请求独立裁决，不以构造 edge case 或新测试补写行为保证。

## Acceptance

- [x] 实施前另行确认本票范围、上述 scoped TDD seams、实际检查与 smoke；在批准 seam 记录 red/green，复用现有有效覆盖，每项新永久测试有独立批准回归，不新增字段复制、wiring、mock echo、non-empty、源文本或内部默认值测试。
- [x] 生产协调器是唯一主 actor 会话事实源，后台完成元数据/身份准备；目录与大小写不敏感 `.md` 文件按实际类型处理，符号链接归一并保留展示路径，不支持/准备失败不改为父目录、不创建假 running，不使用 inode/bookmark 或路径小写化跟踪身份。
- [x] 多个同目标请求在 starting 等待同一次启动结果，running 复用同一实际 URL与 owned 服务；真实 Go 同时/重复打开原路径及其符号链接，记录同一有效会话/PID/URL和正确 HTTP 内容，不以 mock 次数或仅 isRunning 证明复用。
- [x] 真实父目录、子目录及其 Markdown 文件分别有独立会话/实际 URL；GET 验证父目录递归内容、子目录自身浏览根、文件仅所选内容，不因包含关系吞并目标。
- [x] 每次实际新启动有独立 generation；旧代次启动完成/状态/退出回调不能覆盖新代次，启动中停止后迟到成功不能恢复 running。确定性行为回归覆盖旧代次退出及停止期间重开，真实 Go 重开只在旧 owned 进程实际结束后建立新服务。
- [x] 启动成功只采用 03 已验证的实际 URL及结果，失败/早退保存目标与可获得原因，无默认端口猜测或假 running；消费当前代次 target/reload/fatal/违规/退出事实，目标 unavailable 或 reload degraded 不删除仍运行服务的管理入口、不伪造空目录或自动恢复。
- [x] 显式停止单个会话从实际阶段进入收尾，经同一适配器关闭 writer、等待真实退出，仍存活 owned child 沿用 4 秒 SIGKILL；只有真实结束才能完成停止，实际失败不以删行/信号已发出伪装成功。真实 A/B 会话停 A 后确认 A PID/端口释放，B 继续返回其正确内容。
- [x] 停止全部覆盖协调器接收并记录的 starting 与 running 会话；受控回归验证 child 尚未建立/ready 未完成时的停止意图和重复停止，真实 Go 在观察到 spawned 且协调器仍 starting 时执行停止，确认该 child 已退出、不被迟到启动结果恢复。其余全部 owned 会话停止后真实 PID/端口不再服务。
- [x] 与真实独立 CLI 共存时，协调器单个/全部停止和失败收尾均不影响 CLI，其原目标仍返回正确内容；只收尾本次确切拥有的进程，不按进程名或未知 PID 清扫，不把独立 stdin/JSON/端口/网络政策改成 managed。
- [x] 正常用户停止与运行期意外退出有不同事实；真实 owned Go 意外退出后不再 running、原因可查看且不自动重启，用户仍可主动新开目标；停止失败保留可查看原因及实际管理状态，不持久化/恢复 PID、端口或运行会话。
- [x] 本票源码和行为回归接入既有唯一工程及 test target；运行受影响 Swift 回归、必要宿主构建与当前真实 Go 集成，记录实际命令/结果/skip，未运行的检查不报通过。若 Go 无改动，不为确认已知既存 errcheck 重跑 lint或宣称其修复；契约变更须先获得批准并验证对应 Go 行为。
- [x] 临时 Foundation 拥有者真实 smoke 证明协调器复用、独立目标、单个/全部与 starting 停止、正常/意外退出及 CLI 隔离；随证据更新 `docs/ARCHITECTURE.md` 的身份/阶段/停止/故障语义，清理临时文件和本次 owned 资源，明确内核阻塞、Chrome 后代、AppKit/Finder/浏览器/TCC/候选及 Intel/macOS 13 未验证边界，不复用旧 PID/端口。
- [x] 全部本票验收及独立只读 Standards/Spec 审阅通过、无未解决范围裁决后才关闭；仅按完整证据勾选 2.2，并结合已完成 03 的贡献勾选整个 2.4，保持其他任务不变。不自动提交/推送或发布实施 05，不同步主规格、归档或正式发布。

## Change-wide coverage reference

全部 26 个任务、30 个 requirement、68 个 scenario 的贡献与发布状态见 [coverage-plan.md](../coverage-plan.md)。01–03 已发布且 done；04 已发布且 done（证据见下方 Closure record）；05–11 仍为未发布规划。每个多票 requirement 仍须后续组件贡献和候选实际验收，不能把发布或完成 04 当作整体能力完成。

## Closure record

- 关闭：2026-10-04，状态 `done`；OpenSpec 任务 2.2 按实际证据勾选；2.4 结合已完成的 03 进程层证据勾选（03 交付 writer 关闭、真实退出与 4 秒确切 owned SIGKILL 兜底；本票交付协调器单个/全部、starting/重复停止、停止期间重开及正常/意外退出区分）。2.1、2.5 由 03 勾选，本票未重开或重勾；2.3 与任务 3–7 未触碰。
- 变更面：新增 `macos/GoGrip/Services/PreviewSessionCoordinator.swift`（唯一主 actor 会话事实源：背景目标准备接入、单飞/复用、generation 隔离、单个/全部与 starting 停止、真实退出确认、故障与目标/reload 状态）、`TargetPreparation.swift`（标准化绝对路径、解析符号链接、按解析后实际类型分类与失败原因）、`ManagedProcessLaunching.swift`（生产适配器 seam 与 `FoundationManagedProcessFactory`）、`macos/GoGripTests/PreviewSessionCoordinatorTests.swift`（14 个受控生命周期回归）、`TargetPreparationTests.swift`（5 个身份回归）；`macos/GoGrip.xcodeproj/project.pbxproj`（3 个生产源进入 app + test target、2 个测试文件进入 GoGripTests）；`docs/ARCHITECTURE.md` 第七节新增“Swift 会话协调器（macOS 端）”并按证据更新日期与适配器停止条目。Go 侧零改动；`Makefile`、scheme、`.gitignore` 复用 03 的 `macos-test` 入口，未改。
- Scoped TDD 证据：S1 以缺少 `PreviewTarget`/`TargetPreparation` 的编译失败为 red，实现后 5 绿；S2 以缺少 `PreviewSession`/`OpenOutcome`/`ManagedProcessLaunching`/`ManagedProcessFactory` 的编译失败为 red，实现后 14 绿。绿过程由测试发现并修正两处实现缺陷：新记录默认 `.starting` 但无 launch 造成空转；已提交的成功启动未标记 child 已建立，使停止等待不会到来的 `.spawned` 事件。复用 03 的适配器行为/真实 Go 覆盖，未复制其解码/readiness/4 秒矩阵。
- 自动检查：`make macos-test` **61 tests / 0 failures**（ManagedProtocolTests 17、ManagedProcessTests 16、ManagedDiagnosticsTailTests 2、ManagedProcessGoIntegrationTests 7、PreviewSessionCoordinatorTests 14、TargetPreparationTests 5）；unsigned Debug `xcodebuild build` **BUILD SUCCEEDED**，新文件无警告；Go 零改动，未重跑已知既存 `internal/server_test.go` errcheck、未宣称 lint 通过。
- 真实 smoke（临时 Foundation 拥有者以生产源 + swiftc 编译，调用生产协调器 → 生产 `ManagedProcess` → 当前 Go；**54 项检查全过，exit 0**）：
  1) 父目录/子目录/`中文 说明.md` 三个独立会话（PID 3584/3585/3586，各自实际回环 URL 与 generation）；父会话 GET `/nested/inner.md` 200、子会话 GET `/child.md` 200 且 `/nested/inner.md` 404、文件会话仅所选文件 200 且同目录其他 `.md` 404。
  2) 原路径与符号链接并发重开复用同一会话/PID 3584/generation，无重复服务。
  3) SIGKILL 真实 owned child 3586 → 会话 terminated、URL 清除、原因 `The preview service exited unexpectedly (status 9)`、无自动重启；用户主动重开得到新 PID。
  4) 移动目录后会话保持 running/URL 并真实报告 `unavailable`（reason 含 `no such file or directory`），仍可停止。
  5) Markdown 命名 FIFO 在准备阶段按实际类型拒绝（`Only directories and Markdown files can be previewed`），未启动服务；`chmod 000` 的常规 Markdown 目标得到真实 Go fatal `target-unavailable: ... permission denied`。
  6) 停 A（PID 3584）后该 PID 退出且其地址不再服务，其他会话继续返回正确内容；真实大目录目标在观察到 spawned child 3595 且协调器仍 starting 时执行 `stopAll`，child 真实退出、迟到启动结果不恢复 running、全部 recorded 会话 terminated。
  7) 独立 CLI（`--json --browser=false --no-reload --port 65117`）在 stop all 前后继续返回其内容，且从未成为协调器会话。
  临时 smoke/诊断文件与本次 owned 资源已清理。
- 独立只读审阅：Standards 与 Spec 两轴三轮。首轮 Standards 3×P2 + 1 advisory、Spec 1×P2 + 1 scope decision；修复后第二轮 Standards 1×P2（测试确定性）、Spec 0 阻断；第三轮 Standards 0 阻断（2×P3 advisory 记录于下）、Spec 0 发现/0 未决裁决。修复：匹配 generation 的 `.exited` 清理由启动竞态遗留的 active launch（敏感性核验：还原该行使新回归失败）；停止未确认时重开返回失败而非无界重试（敏感性核验：还原旧循环使测试在 2s 期限失败而非挂起）；事件在主 actor 发射时按顺序内联应用，使旧代次回调回归确定（敏感性核验：移除 generation 校验使测试在 emit 后立即失败）；工厂/协调器初始化收敛为 `executableURL`；`.md` 只接受常规文件。范围裁决：符号链接 `.md` 判定按解析后实际文件（人工裁决，`docs/ARCHITECTURE.md` 记录），无其他未决项。
- 保留的 advisory/限制：内联事件分支假定“主线程上的同步发射来自主 actor”；当前生产适配器一律从自身队列发射（launch、stdout、stderr、termination、超时），无已知触发路径；05/07 引入新的进程内发射点需保持该前提。停止失败回归的期限观察在回归路径上会留下挂起的测试任务（测试实例局部、无 OS 进程，不影响其他测试）。元数据准备尚未完成、还未成为会话的请求不在 `stopAll` 取消范围内——App 退出的“禁止新启动”由 05 负责。
- 未覆盖/限制：AppKit/Finder/默认浏览器/完整面板/最近记录/登录项/TCC 与候选包属 05–11；不管理 Chrome 后代；未在 Intel/macOS 13 实际运行；不对不可中断内核 I/O 承诺硬截止；未提交/推送、未同步主规格、未归档。
