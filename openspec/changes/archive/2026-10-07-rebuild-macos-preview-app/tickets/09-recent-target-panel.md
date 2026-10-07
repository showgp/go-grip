# 09 — 最近 20 个目标与完整原生面板闭环

Status: done
Blocked by: None
Covers OpenSpec tasks: 4.1, 4.2, 4.3
Behavior source: ../specs/macos-menu-bar-app/spec.md, ../specs/macos-preview-sessions/spec.md, ../specs/macos-finder-service/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

在采用 08 已交付且经独立审阅确认的 native→Go 授权/访问反馈、适用已批准阶段性顺序例外的真实开发 App 上，提供“运行会话在前、最近目标在后”的紧凑原生菜单栏面板：用户能区分同名目标、查看真实阶段及访问/热重载/失败状态，明确重开浏览器、复制实际 URL、停止单个/全部、手动打开和退出；最近至多 20 个规范化目标跨启动保留，通过同一生产打开路径主动重开。清空最近记录在本次及下次启动都生效，但不停止预览或丢失运行会话的管理入口。重启只恢复历史显示，不自动访问目标、恢复会话或打开浏览器。

本票完整交付 **4.1、4.2、4.3**，不是只做存储类或把旧 50 项上限改为 20。复用 01–07 的 managed Go、唯一协调器、批次、原生报告和生命周期，以及 08 已交付且经审阅确认的访问反馈/适用指导；不创建另一份运行事实或打开管线。网络共享补证仍归 08，候选包与支持环境验收由 12/13 完成，本轮真实开发 App 面板证据不替代它们。

01–07 已完成，07 已在本会话通过独立确认；08 已交付部分也已通过本会话独立 Standards/Spec 审阅，未发现代码/文档缺陷或未决路线裁决。**08 仍 open、3.3 未完成，实际已挂载网络共享验证仍是已知缺口。** 用户已批准 design Migration Plan 的 **08 → 09 阶段性顺序例外**，并已对齐 `tasks.md` 第 4 组：上述已交付部分的真实授权主体、受保护目录允许/拒绝、外接卷访问及失败指导经独立审阅通过后，网络共享环境缺口不再作为 09 实施的硬前置；不是完整权限/卷 gate 已通过，也不删除网络卷要求。本轮仅对齐票据前置，不授权编码、写测试、安装或改变用户持久数据；实施前仍须另行批准本票范围、scoped TDD seams、native smoke/具体数据与环境操作及实施。

`tasks.md` 是唯一进度账本。只有本票全部交付与实际表面验证、独立 Standards/Spec 审阅通过且无未决范围裁决，才关闭 09 并勾选完整 **4.1、4.2、4.3**；不凭存储测试、截图、发布或单个子部分完成勾选，不改 08/其他票和 checkbox，不自动发布或实施 10。

## Source mapping

批准 specs 是行为权威。下表保留 requirement/scenario 原名，只限定本票历史/面板贡献；已有机制与后续候选环境的整行保证不能仅凭 09 完成。

| Requirement / scenarios | 09 贡献及剩余边界 |
|---|---|
| macos-menu-bar-app — Persistent native menu bar host：`Open and dismiss the panel` | 在原有 AppKit/status item/popover 内完善面板，不建独立管理主窗口；打开/关闭、历史空/有记录时仍无 Dock/终端，关闭面板不停止。05 已提供宿主，13 保留候选证明。 |
| macos-menu-bar-app — Visible sessions and explicit controls：`Distinguish targets with the same name`、`Copy an active preview address`、`Exit through the panel` | 完整面板以足够路径区分同名目标，运行区域先于最近区域；明确重开/复制/单停/全停/退出，显示实际 phase、访问、reload 及失败信息，不用状态点隐式动作。机制复用 04/05，候选由 13。 |
| macos-menu-bar-app — Manual opening uses the same target contract：`Manually reopen a Finder-started target`、`Choose a new target manually` | 最近重开和现有手动多选均消费同一生产批次/身份/复用/退出准入；Finder、手动、最近和运行行的主动重开进入同一成功目标 recency 语义。类型、5/6、失败汇总及权限路径复用 04–08，不另建历史启动器；13 保留候选证明。 |
| macos-menu-bar-app — Bounded persistent recent targets：`Open more than twenty distinct targets`、`Reopen an older recent target`、`Restart after using previews` | 独立版本化存储最多 20 个规范化身份，实际成功建立会话/主动重开按最近使用更新；别名不重复，用户选择路径可显示。重新创建存储及重启 App 只恢复位置/必要显示信息，没有正文/进程事实或自动启动。13 保留候选重启证明。 |
| macos-menu-bar-app — Clearing history does not stop sessions：`Clear history with an active preview`、`Restart after clearing history` | 清空只作用于新最近记录及其独立键，当前/下次启动为空；运行事实、可用 URL、管理动作与 owned 资源不变。不清除旧数据或整域用户偏好，13 保留候选验证。 |
| macos-menu-bar-app — Action failures remain understandable and recoverable：`Open a stale recent target`、`Recover from browser opening failure`、`A service exits unexpectedly` | 陈旧最近路径主动重开显示目标/实际可得原因及适用指导，不产生假 running；浏览器失败仍有重开/复制与真实服务；退出原因可看并可从最近主动重开。沿用 07 根级一次提示和同一可查看报告、08 指导，完整双语由 10，候选由 13。 |
| macos-preview-sessions — One session per normalized target：`Reopen a target through a symbolic link`、`Receive concurrent requests for the same target`；Containment does not merge target identities：`Open a child directory and a file under an active parent` | 新历史采用同一身份，最近重开复用活动会话，父/子/文件记录不因包含关系合并；已有单飞/代次机制由 04，不在存储里另做会话协调或重跑整套竞态矩阵，候选由 13。 |
| macos-preview-sessions — Verified startup and actual preview URL：`A nondefault port is used`、`Open a file with spaces and non-ASCII characters`、`Startup ends without valid readiness information`；Browser opening is separate from service lifetime：`Browser opening fails after service startup`、`Close the browser preview` | 已验证成功会话才成为成功历史；请求浏览器失败不撤销已建立目标记录、不抹掉 running/实际 URL。完整面板消费已验证 URL，复制/重开不猜地址，关闭标签页/面板不收尾。01/03–07 已提供机制，13 保留候选证明。 |
| macos-preview-sessions — No automatic session restoration or restart：`Relaunch with recent targets`、`A preview process exits unexpectedly` | 恢复历史不访问目标、不恢复 PID/端口/会话或打开浏览器；意外退出不再显示 running，原因及主动重开入口保留；显式停止不误报意外退出。机制沿用 04/05，13 保留候选异常/重启验收。 |
| macos-preview-sessions — Unavailable paths are not empty directories：`A target is moved or deleted`、`A mounted volume becomes unavailable`；Visible hot-reload degradation：`Watch coverage is incomplete`、`Use a network-volume preview` | 完整呈现唯一事实源已有 unavailable/degraded 及实际原因，与 phase 分开；保留查看/停止和浏览器手动刷新，不把未知/不可访问显示成空目录或完整监视，不追踪移动/重连/轮询/自动恢复。02/04 提供结构化机制，08 提供已验证的访问/卷证据并保留网络共享补证缺口，13 保留完整候选故障矩阵；09 不以其他卷或受控状态冒充网络卷验证。 |
| macos-preview-sessions — Explicit stopping confirms service termination：`Stop one of several sessions`、`Stop all application sessions`；Services end with their owning application：`Quit the application normally`、`Force termination of the host`、`Host exits during startup`；Independent CLI operation remains available：`Launch the renderer without the application` | 新面板仍用共享单停/全停/退出，本轮 UI 操作观察 owned PID/监听释放与其他目标/独立 CLI 隔离；不把清历史/删行当停止。复用 01–07 所有权证明，不重复整套 SIGKILL/writer 矩阵；候选由 13。 |
| macos-finder-service — Supported open targets：`Receive a Markdown file`、`Receive a directory outside the home folder`、`Receive an unsupported file` | 最近消费者保持既有目录/.md 类型和失败规则，不能转换父目录或另设位置限制；实际访问范围依赖 08，不重做目标机制，候选由 13。 |
| macos-finder-service — Batch deduplication and quantity confirmation：`Open exactly five distinct targets`、`Confirm six distinct targets`、`Cancel a large batch`、`Count aliases only once` | 保留现有手动多选及共享批次；新历史按成功目标更新，不能使取消/失败提前写入，不新增最近多选界面或第二批次框架。机制复用 06/07，候选由 13。 |
| macos-finder-service — Partial batch failures：`Some targets fail`、`All targets succeed` | 新最近入口与面板不破坏继续处理/成功保留和根级一次报告，全成功不加提示；不复制 06/07 内部矩阵，真实访问/指导由 08，候选由 13。 |
| macos-finder-service — Minimum necessary file authorization：`Access is denied`、`Open an ordinary accessible target` | 最近入口沿用 08 已交付且经审阅确认的最小权限/失败指导，不以 FDA/辅助功能/自动化作为前置或用历史记录绕过实际访问判定；本票不重验或重建 TCC gate，08 网络共享补证不由 09 替代，候选由 13。 |

上述 Finder requirement 是最近消费者必须保留的共享入口约束，本票不新增其机制或重复 06–08 矩阵；首次引导由 10。包装仅作为验收边界参照 [packaging spec](../specs/macos-app-packaging/spec.md)，09 不贡献 universal/DMG/正式分发完成声明。

技术依据为 [design Decisions 1–4、6、8–9、Verification Strategy、Risks / Trade-offs 与 Migration Plan](../design.md#decisions)：AppKit 生命周期/SwiftUI 呈现、后台准备/规范化身份、ready 后最近更新与浏览器请求分离、结构化访问/reload 状态、独立版本化 UserDefaults 和早期权限 gate，以及已批准的 08 → 09 阶段性顺序例外。本次仅将票据前置对齐该裁决；实施不改批准 artifacts，需扩展行为或再次改变 gate 时仍先请求独立审批。

## Scope and implementation handoff

发布时相关实际表面为 `macos/GoGrip/AppDelegate.swift`、`PreviewAppModel.swift`、`Services/PreviewSessionCoordinator.swift`、`Services/TargetPreparation.swift`、`Services/ManagedProtocol.swift` 与 `Views/PopoverView.swift`。目前没有新 RecentTarget 存储；Popover 只有基础会话/报告/控制，尚未完整渲染 targetStatus/reloadStatus。以下是现状定位，不锁定内部 API：实施前重新读取 08 已交付部分及网络共享已知缺口、实际源/工程和验证记录，不覆盖另一会话工作。

- **唯一进程事实与紧凑面板**：由现有强持有应用根/模型拥有最近存储并向同一 popover 呈现；运行会话始终来自协调器，不把 RecentTarget 当 RunningInstance。macOS 13 的 ObservableObject/@Published 模式保持一致，原生面板内显示足够路径及实际阶段、可访问性、热重载状态/原因、浏览器/退出/操作失败。pending 或未知访问不能被擅自显示为完整监视/可访问空目录；running 与 unavailable/degraded/浏览器失败可同时存在。布局应让控制与完整失败原因实际可达，不固定现有 420×320 或新增独立管理窗口；不做设计系统/高级设置。
- **最近记录提交时机**：遵循 design Decisions 4/9，在已验证会话成功建立或用户主动重开运行目标时，按规范化身份更新最近使用，然后请求浏览器。Finder、手动、最近和运行行明确重开保持相同语义；浏览器 false 不撤销成功目标的历史。被拒绝/取消的批次、类型/准备/启动失败不能提前作为成功目标写入。仅复制 URL、停止、清历史或加载历史不是新打开。不监听视图出现、遍历当前行或观察启动恢复来反推“最近使用”，不通过日志推断成功。
- **小型版本化持久模型**：依照 design 的 RecentTarget 保存规范化位置、用户选择的展示路径、目标类型和最近使用时间，最多 20 个不同身份；更新旧目标的次序，父/子/文件仍独立。使用新独立版本化 UserDefaults 键，只持久化位置与必要展示信息，不存 Markdown 正文、PID、端口、运行阶段或运行错误。加载仅恢复这些记录，不探测/访问每个历史路径，不建服务。键名/内部 API 由实施按既有模式确定，不预设额外兼容迁移、权限存储、恢复、泛化存储或损坏数据保证。
- **最近重开定位与失效反馈**：定位已记录的规范化目标，显示信息保留用户选择路径，不借 bookmark/inode 追踪移动或重新定位符号链接指向。主动打开仍进入同一生产准备/批次/复用路径；已运行目标 PID/generation/验证 URL 保持，陈旧/不可访问路径保留具体错误及 08 适用指导，不改为父目录、固定端口或后台重连。意外退出后仅主动重开可开始新代次；沿用原生根提示与可查看报告，不再加面板本地第二弹窗。
- **清除边界及旧数据**：清空只清新最近键和内存列表，不能触碰协调器/owned writer 或停止入口；后续重启仍为空，除非用户之后又明确打开目标。保留 05 的历史失败报告策略，成功浏览器重试只清对应行错误，不把清最近当作清会话/整域偏好。旧 `go-grip-history` 50 项用户数据不读取、迁移、写入或删除；新 schema 初次升级不宣称旧记录已迁移。
- **干净切换旧源与测试**：05 已将 `Utilities/Storage.swift`、`Models/HistoryEntry.swift`、`GoGripTests/StorageTests.swift` 的最终切换留给 09。实施时移除被新存储取代的旧源、工程引用与旧 50 项/原始路径/默认域测试，不保留 aliases/re-exports/shims，也不把 maxEntries=50 的断言重钉为 20。旧测试直接使用 standard 并清 `go-grip-history`，不能运行它们来证明新历史或清理用户数据；新持久化测试使用独立可销毁域。更新所有本票替换调用与唯一 Xcode 工程/test target，不清理无关代码。
- **文档切换**：随真实证据更新 `README.md` 最近/状态/重开/清除/失败恢复操作，`docs/ARCHITECTURE.md` 成功历史时机、独立新 schema、事实与状态呈现及不自动恢复；更新受影响的旧示例切换说明（现有 `docs/HANDOVER.md`、`docs/macos-app-impl.md`、`docs/macos-app-plan.md` 中本票过期部分）。说明旧数据保留、新历史初次为空、浏览器失败不丢成功目标、清历史不停止和跨启动只恢复目标列表。08 的访问指导保持，双语/引导/登录及候选支持范围不提前声称完成。

## Scoped verification and proposed TDD seams

发布不授权写测试。实施前确认下列消费者 seams，永久测试只 trace 到 4.1/4.3、上述批准 requirements 或明确仓库质量规则；真实面板/渲染以实际 native 操作验证，不做源文本或布局 snapshot 测试。

1. **持久最近边界**：隔离 UserDefaults 域和可销毁目标，打开 21 个不同身份后断言具体保留的 20 个目标及被淘汰目标；重开较旧目标/实际别名按身份去重并更新具体次序；重新创建 store 后仍恢复相同位置/次序，不接入旧 50 项。不能只断言 length grew/非空或复制字段。
2. **成功提交与失败分离**：在已有生产打开/进程 seam 上证明未验证或失败目标不会抢占成功历史，成功建立服务后即使浏览器拒绝仍保留该目标和运行管理；主动重开复用同一事实并更新 recency。复用 03–07 的 controlled process/browser 与 ready/失败回归，不复制整套协议/数量/退出矩阵，也不以 fake forwarding 参数或 browser mock echo 作为历史集成证明。
3. **清历史与存活会话分离**：在活动预览时清新记录，随后重新创建 store 为空，真实会话身份/阶段/URL及重开/复制/停止能力未丢，独立域中其他/旧数据未被整域清除。复用既有会话 seam；最终存活/停止还需下面真实 Go/native 证据，不只检查“未调用 stop”或 isRunning。

不得新增默认 key/schema 文案、字段转发、JSON 源文本、URL 拼接、按钮 wiring、布局尺寸、仅不抛异常、非空 reason 或重复同路径用例。确定性同步用已有显式进程/事件边界，不用单次 Task.yield 假定完成；独立域/临时目标须清理，不改真实用户默认域。新 red/green 以实施前批准的最小测试集为准，已有不合规格的旧测试删除而非重钉。

## Real verification

下表是实施后必需观察，当前均未执行。使用 08 已交付且经独立审阅确认的 native→Go 路线、确切开发 App/bundle 内置 Go、可辨认内容与明确 owned 资源；实际面板、默认浏览器、PID/generation/URL和资源观察一起记录，不能用 store 单测、plist、直接 selector 或临时 Foundation 拥有者代替 native 闭环。阶段性顺序例外不缩减本表或本票功能验收，也不以本轮状态/卷观察替代 08 的网络共享补证。

| 场景 | 必须观察的结果与证据边界 |
|---|---|
| Finder/手动成功目标 → 最近列表 | 普通目录、中文空格 `.MD`、真实空目录经现有 native 入口打开对应浏览器内容；成功目标可在最近区域主动重开。运行区域先于最近区域，关闭/重新打开面板不收尾，也不从视图出现重复写历史或弹旧报告。复用 07/08 入口证据，不重跑完整5/6矩阵。 |
| 同名不同路径与明确动作 | 至少两个不同父目录下同名目标，路径足以区分；各行复制/重开/单停对应真实目标及已验证 URL，单停只释放该 owned PID/监听，其他目标及独立 CLI继续；全停/退出真实收尾，不仅删行。手动选择继续可用，无隐式状态点或高级配置。 |
| 最近记录重排、别名与活动复用 | 从最近记录主动重开活动目标以及通过现有入口重开其别名，只有同一规范化身份的一项历史与一个服务，recency 更新，PID/generation/URL不变，实际浏览器显示正确内容。父/子/文件不被最近存储按包含关系吞并；21→20的具体淘汰由隔离行为回归证明，不要求为视觉验收同时运行21个服务。 |
| 重启仅恢复最近列表 | 经真实面板退出后 owned PID/端口释放；重新启动同一构建恢复最近记录而没有 running/旧端口、目标访问/新 child或浏览器请求；随后用户主动重开才建立对应服务。不得从 JSON字段或 store重新创建单独推断 App没有自动启动。 |
| 活动会话中清空与清后重启 | 清新最近列表后同一真实服务/URL继续HTTP提供原内容，运行行和复制/重开/停止可用；在没有再次打开目标前退出/重启，最近保持为空、无自动会话。实际用户旧键/偏好不读写；需要数据边界测试用隔离域哨兵，不用用户数据作 fixture。 |
| 陈旧目标与真实状态呈现 | 对专用最近目标删除/移动后主动重开，观察原路径/实际原因/08指导及可查看报告，无假running或新位置跟踪；显示已有 unavailable 与已知 reload degraded 的原因并保留管理/浏览器手动刷新，不能把它们等同进程退出或普通空目录。复用 08 实际访问/卷和 02/04 状态机制，记录本轮面板实际呈现，不重做完整系统权限/Go故障矩阵。 |
| 浏览器失败与进程意外退出 | 浏览器拒绝时真实服务、URL和成功历史仍在，恢复可重新打开/复制；既有失败横幅按已批准策略保留。意外退出时 phase/原因正确、无自动重启、最近入口仍可主动重开；明确停止不误报意外。难触发表面可在另行批准后用带标注临时受控构建补充，记录哪些是受控结果并还原，不冒充真实系统拒绝/TCC/watch故障，不替代必需的真实Go/native证据。 |

运行受影响 Swift 回归、已有真实 Go 集成与开发 App 构建，记录实际命令/结果/skip；复用唯一工程和 `make macos-test`/`make macos`，不 pin 07 的旧 80 项数量为永久要求。必要安装替换、专用目标权限/删除/移动、系统/卷或持久域操作先取得具体许可，08 环境尚在使用时不擅自改变。Go 无改动不为确认已知 errcheck 重跑 lint；构建、签名或交叉编译不能替代面板/浏览器、TCC或 Intel/macOS13观察。

若缺少 required native/浏览器/状态或重启环境，记录具体缺失并保持对应验收与4.1/4.2/4.3未完成，不缩为纯存储或静态UI后关闭。清理本轮临时测试域/目标/受控补丁及确切owned资源，恢复获准环境，不以进程名清扫独立CLI或用户数据。

## Prerequisites and non-goals

前置仍依赖 [08-native-target-access](08-native-target-access.md) 已交付的真实授权主体、受保护目录允许/拒绝、外接卷访问及失败指导通过独立 Standards/Spec 审阅，且无未决路线裁决；这些已在本会话确认。依据已批准 design Migration Plan 与 `tasks.md` 第 4 组的阶段性顺序例外，不再要求 08 整票关闭或网络共享环境证据先齐备才能实施 09，因此票面无未完成的票据 blocker；本票审阅及范围、scoped TDD、native smoke、具体数据/环境操作和实施许可仍待另行批准。08 保持 open、3.3 未勾选，网络卷要求不删减，环境具备后回到 08 补证并独立审阅；若后续证据否定直接内置工具路线，停止相关集成并与用户重评设计，不自行加入 helper、detach 或 FDA 兜底。05–07 已完成，提供真实 native 根、全部入口、生产批次、根报告与退出；01–04 提供 Go/状态/身份/停止。

不做首次引导/完整简中英文/ServicesMenu（10）、SMAppService登录选项（11）、universal/archive/DMG/候选CI/候选跨架构及最低系统完整矩阵（12/13）。不改变08权限路线/扩充权限管理，不做 bookmark/inode移动追踪、旧数据迁移/兼容层、恢复运行状态、自动启动/重启/重连/健康轮询/watch恢复、新Go发现或预算政策、浏览器标签页管理、拖拽/搜索/单项删除历史/高级配置/独立管理窗口；不建泛化存储、第二份会话或批次、helper/XPC/launchd/权限提升、未知PID/进程名/Chrome后代清扫。提交/推送、主规格同步、归档及正式分发仍未授权。

## Acceptance

- [x] 实施前核对08已交付的真实授权主体、受保护目录允许/拒绝、外接卷访问及失败指导已通过独立Standards/Spec审阅且无未决路线裁决，并适用已批准design/tasks的阶段性顺序例外；网络共享仅列08已知待补证缺口，08仍open、3.3未勾选，不据此宣称完整权限/卷gate通过。另行批准09范围、scoped TDD seams、native smoke及具体数据/环境操作和实施；不覆盖另会话08工作，不凭发布或源码解除前置。
- [x] 新独立版本化最近存储最多20个不同规范化身份，成功建立会话/主动重开按最近使用更新，实际别名不重复、父子文件不按包含关系合并；保留用户选择显示信息，不借bookmark/inode跨移动跟踪。
- [x] 成功历史提交在验证ready/真实会话成立之后且与浏览器结果分离，所有批准打开/主动重开入口语义一致；准备/类型/启动失败及取消批次不提前写成功历史，浏览器拒绝不抹掉已成功目标/服务，不从视图或日志推断最近使用。
- [x] 持久化仅位置/必要显示信息，恢复无正文/PID/端口/phase或自动访问/启动/浏览器；隔离域回归证明21项的具体20项与淘汰、旧目标/别名重排、重新创建store恢复，真实App重启只见最近且主动重开才启动。
- [x] 紧凑原生popover运行会话在最近之前，同名目标有足够路径和明确打开/复制/单停/全停/手动/退出；无独立管理主窗口、Dock/终端或隐式状态点，关闭面板不停止，完整原因和动作在实际面板可达，不新增布局/文案/wiring永久测试。
- [x] 面板直接消费协调器实际phase、target/reload及可得失败，running与unavailable/degraded/浏览器失败分开，未知不假充available/完整监视；启动失败无假running、意外退出原因可看且不自动重启、明确停止不误报意外，管理/手动刷新入口保持。
- [x] 最近重开共享Finder/手动的准备/批次/身份/复用/权限/退出准入，活动目标PID/generation/实际URL不变且浏览器内容正确；陈旧/不可访问目标有原路径/真实原因/08适用指导及一次根提示/可查看报告，不转父目录、追踪移动或后台重连。
- [x] 活动预览中清新最近列表不改变会话/URL/资源/管理，实际HTTP与复制/重开/停止继续；没有后续打开时重建store及真实App重启仍为空；只清新键，不整域清除/读写迁移旧go-grip-history50项或其他用户偏好。
- [x] 删除被新schema取代的旧Storage/HistoryEntry/StorageTests及工程引用/冲突测试，迁移全部受影响调用，无旧运行兼容层；新永久测试按获准消费者seam确定性且隔离域/目标，不把50改20重钉或运行旧standard清理测试，不复制03–08机制矩阵。
- [x] 本轮真实native面板/默认浏览器及Go状态/内容有观察，覆盖同名目标动作、最近/别名活动复用、重启不恢复、清历史不停止/持久为空、陈旧错误、degraded/unavailable及意外退出/浏览器恢复表面；受控结果单独标注还原，ownedPID/端口真实收尾且独立CLI隔离，不以单测/静态UI或旧PID代替。
- [x] 受影响Swift行为回归、已有真实Go集成及开发App构建有实际命令/结果/skip，新行为按批准seam做red/green；实际README/架构/受影响切换说明同步，临时域/目标/补丁/owned资源清理、获准环境恢复，未验证候选/双语/登录/支持平台边界明确，不重跑已知errcheck。
- [x] 全部09验收及独立只读Standards/Spec审阅通过且无未决范围裁决后才关闭09并勾选完整4.1/4.2/4.3；其他票和checkbox不变，不自动发布/实施10、提交/推送、同步主规格、归档或正式分发。

## Change-wide coverage reference

[coverage-plan.md](../coverage-plan.md)逐行展示完整26任务、30 requirements、68 scenarios及发布时覆盖映射，其中08/09依赖状态仍是旧快照，尚待另行批准对齐；其旧“08未验收/09被08阻塞”文字不能覆盖本次已批准的design/tasks阶段性裁决。当前01–07已发布done，08已交付部分经独立确认但整票open、网络共享待补证，09仍open待审阅、未获准实施且不再以该网络共享缺口为硬前置，10–13仍是未发布规划。账本当前12/26；3.3及4.1/4.2/4.3未勾选。本轮仅修订09票的依赖/前置，不修改proposal/specs/design/tasks、08或其他票，也不修改覆盖映射、关闭票据或授权编码。

## Closure record

- 关闭：2026-10-05，状态 `done`；12 项验收全部按实际证据通过。实施批准：用户显式批准本票范围（4.1/4.2/4.3 一体）、S1–S3 scoped TDD seams、native smoke 与具体数据/环境操作（临时专用目标与权限；退出/重启与 kill 专用 owned child；替换 `/Applications` 安装；带标注临时受控浏览器拒绝构建）。审阅裁决 1 项已由用户裁决并实施：最近行重开以记录规范化身份定位、保留用户选择展示路径（`openBatch` 展示覆盖），改指/删除符号链接不重定向到新目标。
- 变更面：新增 `macos/GoGrip/Models/RecentTarget.swift`、`macos/GoGrip/Utilities/RecentTargetsStore.swift`、`macos/GoGripTests/RecentTargetsTests.swift`；修改 `macos/GoGrip/PreviewAppModel.swift`（store 注入、`recentTargets`、验证成功与显式重开时更新 recency、`reopenRecent` 身份定位 + 展示覆盖、`clearRecentTargets`、`openBrowser` 在退出准入关闭后不写历史）、`macos/GoGrip/Views/PopoverView.swift`（运行区在前含 phase/target unavailable/reload pending/degraded/off 与失败原因；最近区逐行 Open 与 Clear；420×460、内部滚动）、`macos/GoGrip/AppDelegate.swift`（生产 store 注入与 popover 尺寸）、`macos/GoGrip/Services/ManagedProcess.swift`（`TargetMode: String, Codable`）、`macos/GoGripTests/PreviewAppModelTests.swift`（隔离域 store 注入 + 3 项回归，含改指链接断言）、`macos/GoGripTests/FinderServiceProviderTests.swift`（隔离域注入 + 批处理等待条件补第 3 次浏览器请求）、`macos/GoGrip.xcodeproj/project.pbxproj`；删除 `macos/GoGrip/Models/HistoryEntry.swift`、`macos/GoGrip/Utilities/Storage.swift`、`macos/GoGripTests/StorageTests.swift`；更新 `README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`、`docs/macos-app-impl.md`、`docs/macos-app-plan.md`。
- Scoped TDD 证据（批准 seam）：red 为 `cannot find 'RecentTargetsStore' in scope` 与 `extra argument 'recentStore' in call`（聚焦运行观察的编译失败）；实现后 S1 3 项隔离域 store 回归（21→20 具体淘汰与次序；较旧身份前移不重复且展示路径取最新；重建 store 恢复次序；clear 只清新键并保留同域哨兵与旧键名哨兵）与 S2/S3 3 项模型回归（仅验证 running 入历史且浏览器拒绝不撤销；最近/真实别名重开复用同一 PID/generation/URL、一条历史并更新次序；清空后运行会话与真实停止保持）全绿；批处理测试补等待第 3 次浏览器请求以去除时序竞争。无布局/文案/wiring/50→20 重钉测试，旧 StorageTests 删除而非改写，隔离域/临时目标清理。
- 自动检查（最终树）：基线 `make macos-test` 81/0 → **87 tests / 0 failures**（+3 store、+3 model）；裁决修正与去竞争后连续 3 次全量 `make macos-test` 均 87/0。`make macos` Release `BUILD SUCCEEDED` + `codesign --verify --deep --strict` 通过；开发构建 CDHash 依次 `2195a0bc…`（正式）、`8474d42f…`（受控拒绝，已还原）、`9c4ddf9b…`（最终）。Go 零改动，未重跑已知 errcheck。
- 真实 smoke（面板文本/动作经 AX 读取/触发，进程/端口/HTTP 终端配对，2026-10-05，详见 `docs/ARCHITECTURE.md` 任务 4 记录）：5 目标（含同名不同父路径、中文+空格 `.MD`、真实空目录）各自独立回环端口与可辨认内容；Copy URL 得到实际 URL；Open in Browser 不重启；单停只释放对应 child/端口且独立 CLI 继续；Clear 有活动会话时不停止、跨重启为空；Quit 与重启只恢复列表、主动重开才启动；陈旧最近目标原路径 + 真实原因 + 08 指导、无假 running；unavailable 与 reload degraded 原因显示且可恢复/手动刷新；kill -9 意外退出原因可看、不自动重启、显式 Stop 不误报；受控拒绝一次提示、服务/URL/历史保留、第三次真实打开并清除。最终构建另行复核改指链接场景：仍打开记录目标且展示路径保留。
- 独立只读审阅：首轮 Standards **1 P2 阻塞**（退出准入关闭后 `openBrowser` 仍写 recency；已修）+ 3 P3 advisory（去未使用 key 配置、store 重排断言最新展示路径、去他域断言；均已落实）+ 1 scope decision（重开路径语义；用户裁决并实施）；增量复审 Standards **0 阻塞**、1 P3（架构测试清单陈旧；已修正）并通过。首轮 Spec **0 阻塞、0 advisory** + 同一 scope decision + 覆盖限制；增量复审 Spec **0 阻塞/0 advisory/0 裁决**并确认实现语义符合票面。
- 保留的 advisory/限制：根失败提示为 `NSAlert.runModal`，其显示期间 Services 请求粘贴板交换超时丢失、提示叠加无法关闭且需重启实例（07 已记录限制的首次实际触发；本任务按批准策略沿用以示并已记录）；受控拒绝为受控结果而非真实系统拒绝；网络卷访问仍为 08 待补证缺口且 3.3 未勾选；登录项/本地化/候选包与 Intel/macOS 13 未验证；ad-hoc 身份跨构建可能需重新授权；清理阶段 08 构建唯一副本随备份被移除、`/Applications` 现为本票最终开发构建 `9c4ddf9b…`（既有身份限制，如实披露）。
- tasks.md：本票完整交付并验证 **4.1、4.2、4.3**，三项已勾选；其他任务、票据与 checkbox 未变。
- 未覆盖：未提交/推送、未同步主规格、未归档、未实施 10；覆盖映射的发布时快照按 05–07 先例留待下一次发布轮随新票更新。
- 交叉引用（2026-10-06，07 票重开的共享提示呈现修正关闭）：本记录所记“根失败提示为 `NSAlert.runModal`，其显示期间 Services 请求在 pasteboard 交换超时丢失、提示叠加无法关闭且需强制重启”的限制已由该修正消除（非模态单槽、提示显示期间请求正常处理、OK/Return 可关闭、失败报告不叠加且面板保留）；09 已交付的最近目标/完整面板/清历史功能未改动。详见 07 票“Reopened 2026-10-06”与 `docs/ARCHITECTURE.md` 的同日修正记录。
