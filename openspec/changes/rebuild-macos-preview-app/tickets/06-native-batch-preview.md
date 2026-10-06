# 06 — 统一多选批次与原生结果汇总

Status: done
Blocked by: 05-native-finder-preview
Covers OpenSpec tasks: 2.3, 3.2 (manual-multiselection contribution only), 3.4 (batch-browser/error-report contribution only)
Behavior source: ../specs/macos-finder-service/spec.md, ../specs/macos-preview-sessions/spec.md, ../specs/macos-menu-bar-app/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

在 05 的真实菜单栏开发 App 中，通过 NSOpenPanel 一次选择多个目录/Markdown 文件，全部目标先后台准备、按规范化身份去重，再决定是否确认。5 个不同有效目标直接处理，6 个先显示数量确认；取消不启动或重开本批目标、不停止既有会话。获准后逐项异步使用唯一生产协调器、内置 managed Go 和已验证实际 URL 打开默认浏览器；部分失败继续并保留成功会话，结束时一次原生提示汇总目标及原因，全部成功不额外提示。

本票交付可运行的“实际多选 → 数量确认/取消 → 真实服务及浏览器内容 → 一次失败报告/原生恢复”纵向切片，不是只设置 allowsMultipleSelection、循环单项弹窗或 mock 转发。生产批次入口接收全部所选 URL，单项手动选择也进入同一路径；07 再接入 Finder Services，本票不提前创建 provider/注册占位或宣称 Finder 已可用。

用户已确认修订后的 05 范围，故本轮发布这张独立 06 文件；**05 仍 open、未实施，是本票尚未满足的真实前置**。本票发布仅供审阅，不授权实施、编写测试或直接补做 05；实施前须另行确认本票范围、scoped TDD seams、实际检查及 smoke，并取得实施许可。

`tasks.md` 是唯一进度账本，下面 checkbox 仅为验收细节。完整本票交付及独立 Standards/Spec 审阅通过后可勾选 2.3；3.2、3.4 只完成多选/批次贡献，仍须 05 和 07 的其余入口贡献及完整验证，不能提前勾选。确认票面或发布文件不等于任务完成，不改 3.1 或其他 checkbox。

## Source mapping

批准 specs 是行为权威，不以旧管理器、现有测试或票面代替。下表保留 requirement/scenario 原名，右列限定本票贡献；整行场景中的 Finder、完整面板/历史、授权主体和候选部分仍由相应后票证明。

| Requirement / scenarios | 06 贡献及剩余边界 |
|---|---|
| macos-finder-service — Supported open targets：`Receive a Markdown file`、`Receive a directory outside the home folder`、`Receive an unsupported file` | 共享批次对全部输入运行时分类：目录/常规大小写不敏感 .md；普通主目录外位置不被排除，不支持文件不改为父目录，失败带目标及原因。Finder 接收由 07，真实 TCC/卷 gate 由 08，候选由 13。 |
| macos-finder-service — Batch deduplication and quantity confirmation：`Open exactly five distinct targets`、`Confirm six distinct targets`、`Cancel a large batch`、`Count aliases only once` | 全批准备/去重在任何启动或浏览器动作之前；按不同有效身份计数，5/6 边界、6 路径归一为 5、确认与取消均有行为回归及真实批次证据。Finder 多选由 07，候选由 13。 |
| macos-finder-service — Partial batch failures：`Some targets fail`、`All targets succeed` | 已获准批次继续其余项目、不回滚成功；不支持、准备/启动及浏览器错误在一次原生报告中显示目标/可获得原因，面板可查看，全部成功无额外提示。Services 接入由 07，候选由 13。 |
| macos-finder-service — Minimum necessary file authorization：`Access is denied`、`Open an ordinary accessible target` | 普通可访问目标直接处理；已观察访问失败纳入批次，不显示为空目录/假 running，不绕过权限或统一要求 FDA/辅助功能/自动化。实际授权检查指导及受保护/挂载卷验证由 08，候选由 13。 |
| macos-preview-sessions — One session per normalized target：`Reopen a target through a symbolic link`、`Receive concurrent requests for the same target` | 批内按 04 身份去重，已有 starting/running 仍计入有效目标且沿用同一协调器单飞/复用；不建立另一份运行字典。Finder 跨入口由 07，候选由 13。 |
| macos-preview-sessions — Containment does not merge target identities：`Open a child directory and a file under an active parent` | 批次不按包含关系合并，父/子/文件分别计数及打开；复用 04 事实和 05 原生基础，不重复其内部状态矩阵。候选由 13。 |
| macos-preview-sessions — Directory and single-file preview modes：`Preview nested documents`、`Preview a selected file`、`Keep an empty directory session`、`Add Markdown to a watched empty directory` | 混合多选沿用真实目录递归/文件直接定位/空目录会话及有效监视，不改变 Go 发现政策；实际批次验证相应内容。Finder 由 07，候选完整文件事件矩阵由 13。 |
| macos-preview-sessions — Verified startup and actual preview URL：`A nondefault port is used`、`Open a file with spaces and non-ASCII characters`、`Startup ends without valid readiness information` | 每项采用 03/04 已验证完整 URL；中文/空格文件对应原目标，失败不猜端口/假 running/遗留本次服务，且不阻断后项。Finder 由 07，候选由 13。 |
| macos-preview-sessions — Browser opening is separate from service lifetime：`Browser opening fails after service startup`、`Close the browser preview` | 将 05 浏览器结果处理用于批次但不逐项弹窗；拒绝后保留真实 running/URL及重开/复制，其他项目继续，关闭浏览器不回收服务。完整面板由 09，候选由 13。 |
| macos-preview-sessions — Services end with their owning application：`Quit the application normally`、`Force termination of the host`、`Host exits during startup` | 新增批次确认/异步等待必须遵守 05 退出准入；确认恢复或前项完成不能绕过退出创建后项服务/打开浏览器。复用既有真实收尾/所有权，不重做 03/05 异常退出矩阵。Services 请求由 07，候选由 13。 |
| macos-menu-bar-app — Manual opening uses the same target contract：`Manually reopen a Finder-started target`、`Choose a new target manually` | NSOpenPanel 目录/文件多选，把全部 URL 交统一批次；单选、别名和活动目标用相同身份/复用/确认/错误规则。Finder/手动跨入口由 07，最近入口由 09，候选由 13。 |
| macos-menu-bar-app — Action failures remain understandable and recoverable：`Open a stale recent target`、`Recover from browser opening failure`、`A service exits unexpectedly` | 一份当前操作/批次报告与会话真实最后原因分开；一次原生汇总且可查看，不依赖通知授权，不将浏览器失败变成服务退出。Services 由 07、历史及完整面板由 09、候选由 13。 |

技术依据为 [design Decisions 2–4、6–7、9](../design.md#decisions)：全部 URL 输入、后台准备与执行分离、批内逐项异步、单一会话事实源、实际 URL/浏览器结果、退出准入、一次 NSAlert 及可查看报告。Decision 2 仅贡献手动多选/同一处理入口，不实施该节 Services 声明和 selector。

## Scope and implementation handoff

当前可用的是 04 的 `TargetPreparation.swift`、`PreviewSessionCoordinator.swift` 和 `ManagedProcessLaunching.swift`；旧 `Views/PopoverView.swift` 仍是单选/旧管理器路线，不能作为新批次基础。05 尚未实施，下面是交付接口边界而非假称已存在的 App 根/浏览器 API。实施本票前重新读取并复用 **实际完成的 05**，不锁定未来内部类型名、不另建调度框架。

- **完整输入与准备**：05 手动入口改为允许目录/文件多选，读取 `panel.urls` 的全部值，不取 first。运行时仍使用 04 已裁决的目标规则，保留所选展示路径，规范化/符号链接解析/实际类型和访问错误在后台完成；不按原始字符串、inode/bookmark、路径整体小写或包含关系合并。全批准备去重结束前不建立会话/启动 Go/请求浏览器，先收集无效项目，数量只计不同有效身份，不只计尚未运行的目标。
- **准备/执行接点**：当前 `PreviewSessionCoordinator.open(URL)` 自行准备再执行。为复用全批准备结果，必要时最小提取其已有单项执行路径，让已准备目标进入同一记录/代次/单飞/停止机制；不再完整重复元数据准备，不复制 launcher 或创建第二份身份/运行模型。预检查不保证稍后的实际访问成功，Go 始终是实际访问的最终判定者，准备后删除/访问变化仍按该项失败处理。
- **确认与执行**：超过 5 个不同有效目标时用原生确认显示实际数量，在获准前无本批启动或浏览器请求，包括已有服务的重开；不超过 5 不显示数量确认。取消终止本批处理，既有会话及其管理入口保持原样，不以 stopAll 回滚。获准后逐项异步处理，失败继续，成功保留；不同请求可在 await 时交错，同目标仍由唯一协调器单飞，不新增批内并发、全局任务队列或互斥式调度框架。
- **浏览器与汇总**：复用 05 的真实会话/浏览器结果，但本批单项不逐项呈现操作 alert；执行结束将所有不支持、准备/启动及浏览器失败汇总成一次 NSAlert，包含对应目标与可获得原因，并在现有原生面板保留可查看信息。全部成功不显示成功提示。报告只保存当前操作结果/原因，不存正文、持久运行状态或第二份 PID/URL/阶段事实；运行期失效继续按会话事实显示，不改为无条件逐项弹窗。浏览器 false 保留服务，true 不证明页面渲染，重开/复制仍使用实际 URL。
- **退出准入**：沿用 05 的单一退出准入，确认回复、后台准备/单项启动结果及批内后项恢复不能在退出准入关闭后创建新 child 或发起浏览器打开；已有本批 starting/running 仍由 05/04 真实收尾。不要仅 cancel Task 或把 stopAll 的已有记录快照当作未来批次不会启动的证明，不另建准入状态或承诺阻塞内核 I/O 的硬截止。
- **调用与文档切换**：迁移 05 实际手动单项打开调用到统一批次入口，移除被取代的逐项错误提示/单选专用入口，不保留互不一致的旧新两条 UI 路径或 shims。基础会话控制/重开及浏览器结果处理继续复用，不能因去重或汇总丢掉成功会话管理。生产源码/必要行为回归进入既有唯一工程/test target；随实际证据更新 `README.md` 多选/数量确认/取消/部分失败操作与 `docs/ARCHITECTURE.md` 批次先后/报告及退出边界，明确手动多选已可用但 Finder 尚未接通。新历史及旧 50 项数据切换仍留给 09，不读写/删除用户旧数据。

### Proposed scoped TDD seams

发布不授权先写测试。实施前批准以下三组消费者行为 seams；每项永久测试必须对应上述已批准 requirement/task，不逐层复制 03–05 的矩阵，不新增字段拷贝、默认值、源文本/文案、UI wiring 或 mock echo 断言。

1. **全批准备、数量与取消**：结合真实隔离目标及最小受控确认结果，证明 5 个直接打开、6 个确认前无启动/浏览器副作用、6 路径归一为 5、不支持/准备失败不计入有效数量；取消保留原有会话/URL及管理能力，包含本批要重开的活动目标。断言实际确认数量与目标结果，不固定按钮文字，不只验证转发参数/启动次数，不重复 04 单项目标分类测试。
2. **继续处理与一次可恢复报告**：构造有效目录、中文/空格 .MD、空目录、不支持/准备失败、启动失败与浏览器拒绝的独立消费者回归，验证失败后仍处理后项、具体成功会话/实际 URL保持、报告含具体失败目标及原因且只汇总一次、全部成功无额外提示；浏览器失败会话仍可恢复。控制结果仅用于确定性故障语义，不把 mock 返回值原样拷贝或单独测试报告非空当作证明。
3. **退出与确认恢复**：在数量确认/逐项异步等待期间受控关闭 05 的退出准入，再恢复结果，证明未处理目标不产生新服务/浏览器动作，已创建 owned 会话由真实收尾管理；取消数量确认本身不停止原会话。只覆盖本票引入的批次等待窗口，不重做 03 descriptor/readiness/4 秒或 05 正常/异常 App 退出全矩阵，不添加构造边界的未批准恢复/重试保证。

### Real verification

- 使用实际构建/本地签名的 05 开发 App，真实 NSOpenPanel 选择目录和文件多选。分别观察 5 个、6 个确认并接受、6 个取消、6 路径归一为 5；确认表面显示实际数量，在确认前/取消后记录原会话 PID/generation/URL及 HTTP 内容与本批无新 child/浏览器请求证据，随后获准的批次在默认浏览器显示各目标正确内容，真实页面不能由 NSWorkspace true 替代。
- 混合成功/失败批次必须走同一生产批次入口→协调器→生产适配器→当前内置 managed Go，观察失败后项继续、目录递归/中文空格单文件/空目录实际内容、已有目标复用及成功 owned 资源仍可管理；一次原生报告包含具体失败目标/原因，重开/复制实际 URL可恢复。若真实文件选择器过滤不支持文件，使用调用 **同一生产批次入口** 的临时真实 Go smoke 验证混合输入，不拿 `.png` mock echo 或另一套批次实现替代，不宣称这已证明未来 Finder 混选。
- 浏览器拒绝可用最小受控结果证明并观察实际原生错误表面，标明与真实系统返回的区别；普通成功浏览器及多选/确认/取消表面必须实际观察。证明实际 Go 访问失败的混合路径，不以 mock failure 代替所有失败证据。确认恢复与退出使用上述 seam，实际批次完成后执行 App 退出，记录 owned PID/端口释放及独立 CLI 内容仍可用，不复用旧 04/05 PID或结果作为本票证据。
- 复用 `make macos-test`、现有 scheme/test target及 05 实际开发构建入口，记录对应 Swift 回归/真实 Go 集成/构建命令、结果和 skip；Go 无改动不为确认已知既存 errcheck 重跑 lint，不宣称 lint 修复。缺少可操作原生/浏览器环境时保留相应验收未完成并说明缺少的观察，不能用测试通过/仅进程创建冒充 smoke，不自行授予 FDA/辅助功能/自动化或改走 helper。

## Prerequisites and non-goals

唯一直接 blocker 为 [05-native-finder-preview](05-native-finder-preview.md)：真实开发 App、单项目标/浏览器结果、基础控制及退出准入必须先实际完成；**范围确认不解除此实施前置**。01–04 已 done 间接提供 Go/身份/状态/所有权/真实停止，旧编号对应见 [覆盖地图](../coverage-plan.md#原-05-范围保留与编号对照)。本票不能借补发文件提前实施 05 或修改其已确认范围；05 完成后的具体 API/证据须重新核对。

不做 Services/provider/selector/plist/真实 Finder 冷启动/跨入口完整验收（07）；不做完整 TCC/受保护目录/外接或网络卷 gate（08）；不做新历史/完整面板/引导/本地化/登录项/universal/archive/DMG/CI/候选支持环境（09–13）。不新增 Go 机器契约、扫描政策/监视恢复、轮询/重启/恢复/重连/移动跟踪、拖拽/高级参数/标签页跟踪、helper/XPC/launchd、未知 PID/进程名或 Chrome 后代清扫。真实普通目标若否定已选工具/授权路线，保留阻塞并请求设计裁决，不把失败藏到 08 或绕过权限。提交/推送、主规格同步、归档和正式分发仍未授权。

## Acceptance

- [ ] 实施前确认 05 已真实完成并通过独立审阅，另行批准本票范围、三组 scoped TDD seams/检查/实际 smoke及实施；记录 red/green，每项新永久测试对应批准的独立消费者回归，复用03–05而不新增文案/默认值/字段拷贝/wiring/mock echo或仅isRunning证明。
- [ ] 真实 NSOpenPanel允许目录/文件多选，全部所选URL进入唯一生产批次入口，单项手动选择也同路；后台全批准备/去重先于任何服务或浏览器副作用，采用04已裁决规范化/符号链接/实际类型，保留展示路径，不支持文件不转父目录，不重复完整元数据准备或另建运行/launch模型。
- [ ] 按去重后的不同有效身份计数，已有starting/running目标仍计数，父/子/文件不合并，不支持/准备失败不计数；行为回归和实际批次证明5直接、6先显示数量6、6路径归一为5无需确认，确认前包括重开既有服务在内无本批启动/浏览器副作用。
- [ ] 数量确认取消不启动或重开本批目标、不停止既有会话，真实原有PID/generation/URL与HTTP内容及管理入口保持，不能以stopAll回滚；确认获准后才逐项异步执行，不建立批内并发或额外全局调度框架。
- [ ] 混合批次中的不支持/准备失败/真实Go启动访问失败不阻断后项、不回滚成功，目录递归/中文空格.MD单文件/空目录有正确实际URL及浏览器/HTTP内容；复用活动目标及别名不产生重复owned服务，不猜端口/假running/遗留本次失败服务，全部成功无额外成功提示。
- [ ] 本批单项错误不逐项弹窗，获准批次结束时一次NSAlert汇总所有失败及不支持目标和可获得原因，面板可查看；报告不取代真实会话事实/管理入口、不依赖通知授权、不存正文或持久运行状态，运行期失效沿用事实显示且无自动恢复。
- [ ] 浏览器拒绝也纳入本批一次汇总且其余项目继续，已可用服务保持实际running/URL，可通过真实面板重开/复制恢复；标明受控拒绝与真实系统证据，正常默认浏览器页面已观察，关闭浏览器不回收服务，不将true当渲染或标签页聚焦证明。
- [ ] 复用05单一退出准入，数量确认及批内async恢复不能在退出准入关闭后创建后项child或打开浏览器；受控回归证明新增等待窗口，已有starting/running由同一协调器真实收尾，实际批次后App退出确认owned PID/端口释放且独立CLI仍返回原内容，不仅cancel任务/删行或清扫未知进程。
- [ ] 迁移05手动入口到统一批次并移除被替代的单选专用/逐项提示路径，复用现有唯一工程/test target、真实managed Go及确定内置工具/开发构建入口；运行受影响Swift回归、实际Go集成及开发App构建，记录命令/结果/skip，不以临时Foundation smoke替代原生多选/确认/取消或未来Finder/候选证明，不重跑旧errcheck确认已知失败。
- [ ] 随实际证据更新README多选/确认/取消/一次错误及恢复说明与架构批次/准入/报告边界，清理本次临时文件和确切owned资源；不接入或删除用户旧50项数据，不改变Go/独立CLI/发现政策，明确Finder、TCC/卷、完整面板、候选与最低系统/Intel仍未验证边界。
- [ ] 全部本票验收及独立只读Standards/Spec审阅通过且无未解决范围裁决后才关闭06并据完整证据勾选2.3；3.2/3.4仅本票贡献，仍待05/07完整验证，3.1及其他任务不变。不得自动发布/实施07、提交/推送、同步主规格、归档或正式发布。

## Change-wide coverage reference

完整26任务、30 requirement、68 scenario及各票当前发布状态见 [coverage-plan.md](../coverage-plan.md)。01–04 done，05 open且范围已确认但未实施；本轮仅新增06，open待审阅、实施被05阻塞；07–13仍未发布。其他票面中的“本轮/未来编号/未发布”是各自发布时的记录，当前计划以覆盖地图为准，不重写已关闭历史或05已确认票面。

## Closure record

- 关闭：2026-10-05，状态 `done`；10 项验收全部按实际证据通过。`tasks.md`：本票唯一完整任务 **2.3** 已勾选；3.2/3.4 仅本票贡献，保持未勾选（仍须 05+07 / 05+06+07 完整验证）；3.1、3.3 未触碰。05 已于 2026-10-04 done 并留下可复用 API/证据。
- 变更面：新增 `macos/GoGrip/Services/NativeBatchQuantityConfirmation.swift`（原生数量确认，显示实际数量；`nonisolated init()` 以便作为默认参数）；改写 `macos/GoGrip/PreviewAppModel.swift`（`BatchQuantityConfirming` seam、`directOpenLimit = 5`、`lastOperationFailures: [OperationFailure]` 取代单条报告、唯一批次入口 `openBatch(_:)`、私有 `requestBrowser(_:)`、移除被取代的单目标 `open(_:)`）、`macos/GoGrip/Services/PreviewSessionCoordinator.swift`（新增 `BatchPreparation`、`prepareBatch(_:)`，把原 `open(_ URL)` 主体最小提取为 `open(prepared:)`，`open(_ URL)` 改为“准备后委托”）、`macos/GoGrip/Views/PopoverView.swift`（`allowsMultipleSelection = true` + `panel.urls` → `openBatch`；一次汇总提示；会话列表下方报告清单）、`macos/GoGripTests/PreviewAppModelTests.swift`（05 的宿主回归迁移到批次入口 + 9 项批次回归 + 受控 doubles：`ControlledConfirmation`/`SequentialLaunchScript`/`ReportRecorder`/逐 URL 拒绝/`ControlledProcess.target`）、`macos/GoGrip.xcodeproj/project.pbxproj`（新文件在两个 target 各注册一次）；文档 `README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md` 状态行。
- 净零改动（已还原/删除，无残留）：`macos/GoGripTests/ManagedProcessGoIntegrationTests.swift` 的临时同入口真实 Go smoke（记录证据后删除）；`macos/GoGrip/Services/WorkspaceBrowserOpening.swift` 的带标注临时受控拒绝构建（还原后 `grep TEMPORARY` 为空）。
- Scoped TDD 证据：seam 1 以缺少 `openBatch`/`lastOperationFailures`/`BatchQuantityConfirming` 的编译失败为 red 并解析到多选/数量/取消行为，实现后 74/0；seam 2、3 的行为随 seam 1 的入口契约一并落地而首轮即绿，故对两者做变异检验——临时改为逐项发布报告并删除批次循环的两处准入 `break` 后实测 3 项红（`testMixedBatchContinuesPastFailuresAndReportsThemOnce` 报告发布数 3≠1、`testQuitAdmissionDuringABatchLaunchStopsTheRemainingTargets` 在准入关闭后仍发起浏览器请求、`testTerminationAdmissionRefusesLatePreparationAndLateStartup` 出现迟到目标报告），还原后 78/0，证明相应回归并非空断言。
- 自动检查（最终树）：`make macos-test` **78 tests / 0 failures / 0 skipped / 0 expected failures**（xcresult：ManagedProtocolTests 17、ManagedProcessTests 16、ManagedDiagnosticsTailTests 2、ManagedProcessGoIntegrationTests 7、PreviewSessionCoordinatorTests 14、TargetPreparationTests 5、PreviewAppModelTests 17 = 原有 8 + 本票新增 9；基线 69）；`make macos` Release **BUILD SUCCEEDED**、`codesign --verify --deep --strict` 通过、工具仅在 `Contents/MacOS/go-grip`、`LSUIElement=true`。Go 零改动，未重跑已知 errcheck，未宣称 lint 修复。
- 真实 Go 同入口 smoke（临时用例，已删除）：输入 [目录、`chmod 000` 的 `.md`、中文+空格 `.md`、空目录、`.png`、缺失路径] → 4 个有效目标不发数量确认；选择顺序第 2 位不可读 `.md` 得真实 Go 失败 `target-unavailable: access target: … permission denied`，其后中文单文件（`http://127.0.0.1:59661/说明%20文章.md`，正文含“你好”）与空目录（`59663/`）仍各自启动且真实 HTTP 200；结束时一份报告含 3 个失败目标与可获得原因；`stopAll` 释放全部 owned PID、端口不再服务。
- 原生开发 App smoke（用户操作 + 终端侧 PID/端口/HTTP 记录）：多选 5 个目标直接打开；6 个先显示数量 6 的确认、接受后打开；6 个取消既不新建也不重开且既有会话 PID/generation/URL 不变；6 条路径（`alias-d1`→`d1`）归一为 5 个目标且不发确认、d1 只有一个会话；`Open in Browser`/`Copy URL`/`Stop` 符合预期；配对观察一轮 2 目标批次 owned child 38360/38361 在 60579/60582 提供 HTTP 200（正文 `target6 root`/`target7 root`），面板 Quit 后同一采样点 app/child/端口全部释放；独立 CLI（6419）在批次与退出前后均 HTTP 200 且不出现在面板。带标注受控拒绝构建（已还原）观察到**一次**原生提示同时列出两个目标及原因、不是逐项弹窗，逐行 `Open in Browser` 恢复只清除该行；用户确认面板过滤 `photo.png`（`allowedContentTypes = [.folder, markdown]`），混合不支持输入按本票约定改由上述同入口真实 Go smoke 证明。
- 独立只读审阅：Standards 两轮（首轮 2 项 advisory：README 恢复措辞、混合批次原因断言过弱 → 均修复；复检 0 阻塞/0 advisory，worst severity none，并确认原因断言非空断言化）与 Spec 一轮（0 阻塞/0 advisory/0 范围裁决/0 未请求行为，逐行引用实现确认本票声明的 contribution）。审阅未改文件、未重跑检查。
- 保留的 advisory/限制：批次汇总的原生提示已观察，但面板清单**版面**未单独复核（由同一数组渲染）；受控拒绝不是真实系统拒绝（真实拒绝表面见 05 记录）；真实 Go smoke 覆盖 4 个有效目标 + 3 类失败，未覆盖嵌套文档导航/活动空目录新增文件等完整监视矩阵（13）；Finder Services/跨入口复用（07）、TCC/受保护/外接或网络卷（08）、Intel/macOS 13、完整面板/历史/登录项/本地化与候选包（09–13）仍未验证；不对准备后目标被替换的文件系统竞态提供保证，Go 仍是实际访问的最终判定；不管理 Chrome 后代。
- 未覆盖/边界：未提交/推送、未同步主规格、未归档、未正式发布；未改动 Go/独立 CLI/发现政策；未读写或删除用户旧 50 项数据；`coverage-plan.md` 的发布时状态快照仍记 06 未实施，按 05 先例留待下一次发布轮更新。
