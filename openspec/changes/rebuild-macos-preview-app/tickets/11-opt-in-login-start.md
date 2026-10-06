# 11 — 用户可选登录启动与真实系统状态

Status: done
Blocked by: 09-recent-target-panel
Covers OpenSpec tasks: 5.3
Behavior source: ../specs/macos-menu-bar-app/spec.md, ../specs/macos-finder-service/spec.md, ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

在已完成 09 的原生菜单栏宿主中提供由用户明确开启或关闭的登录启动选项。通过 `SMAppService.mainApp` 读取及修改当前 App 的系统注册状态；未注册时默认关闭，已由用户注册的状态不被启动、服务调用或退出流程擅自注销。需要批准或操作失败时显示真实状态和可获得原因，不把用户期望值或 API 调用返回当作已启用事实。登录启动关闭、宿主未运行时，已有 Finder Services 冷启动仍可打开目标到默认浏览器。

本票完整覆盖 **5.3**，不重复交付 10 的 **5.1、5.2**，也不新增登录 helper、LaunchAgent 或独立设置中心。新增登录 UI 的人类可读标签、状态和行动指导须接入批准的简中/英文/英文回退方案，保留系统原始诊断；不能以本票不承担 5.2 为由漏译新增表面。

**发布核对时状态**：01–07、09 已发布 done；09 的 4.1–4.3 已完成并独立审阅，满足本票实际面板/宿主前置。07 重开的共享提示修正已独立确认无阻塞，根报告与数量确认保持非模态单槽；FIFO 按用户确认的 presenter 入槽顺序理解。08 已交付的授权/访问贡献经独立确认，整票仍 open、网络共享待补证，3.3 未勾。10 已发布 open，用户计划在新会话处理；本会话不推断其具体实施许可、进行中改动或完成。当前账本为 15/26，5.3 未勾。

**发布不是实施许可**。范围、必要 scoped TDD seam、检查、native smoke 及具体安装/登录项/用户数据操作须在实施前另行批准。仅发布本票和刷新覆盖地图；不写应用或测试代码、不修改其他票或批准 artifacts、不变更任务勾选、不操作用户登录项，不发布 12。

## Source mapping

批准 specs 是行为权威，任务进度仅由 [tasks.md](../tasks.md) 记录，票内清单不是另一本进度账本。

| Requirement / scenarios | 11 的贡献与保持边界 |
|---|---|
| [macos-menu-bar-app](../specs/macos-menu-bar-app/spec.md) / Opt-in launch at login：`First use without login startup`、`Change the login-start preference` | 本票主交付：未注册默认关闭、仅用户动作开启/关闭、系统状态与 UI 一致、拒绝/需批准不假成功；真实开启/关闭与无登录启动的 Services 冷启动。 |
| macos-menu-bar-app / Persistent native menu bar host：`Open and dismiss the panel` | 登录选项在既有原生宿主可读可达，不另建管理主窗口、恢复空 Settings scene 或显示 Dock/终端；关闭 UI 不停止服务。 |
| macos-menu-bar-app / System-language localization：`Use Simplified Chinese`、`Use English or an unsupported language` | 新增登录标签、真实状态、失败包装及批准指导复用 10 的统一资源方案；观察新增表面的简中、英文和不支持语言回退，不替代 10 的完整 native/Finder 菜单验收。 |
| [macos-finder-service](../specs/macos-finder-service/spec.md) / Finder Services entry：`Open a selected folder from Finder`、`Invoke the service while the host is not running` | 登录项不是服务前置；沿用 07 唯一 provider、冷启动和统一批次，实际关闭登录启动后从 Finder 得到真实预览。不重做 Services 声明或索引。 |
| [macos-preview-sessions](../specs/macos-preview-sessions/spec.md) / Services end with their owning application：`Quit the application normally`、`Force termination of the host`、`Host exits during startup` | 保持已交付所有权/退出机制；登录项操作不充当停止服务，退出不擅自修改登录偏好。仅复查受影响的 native 路径，不重演强杀矩阵。 |
| macos-preview-sessions / No automatic session restoration or restart：`Relaunch with recent targets`、`A preview process exits unexpectedly` | 登录注册不持久化 PID/端口或恢复服务；App 启动仍只恢复最近列表，不自动启动预览、打开浏览器或重启退出的服务。 |

技术路线依据 [design Decision 9](../design.md#9-最近记录错误及登录项不混入进程持久化)，宿主和退出边界依据 Decisions 1、7；任务完整描述见 5.3。Apple 的 [SMAppService](https://developer.apple.com/documentation/servicemanagement/smappservice)、[mainApp](https://developer.apple.com/documentation/servicemanagement/smappservice/mainapp)、[Status](https://developer.apple.com/documentation/servicemanagement/smappservice/status-swift.enum) 与 [unregister](https://developer.apple.com/documentation/servicemanagement/smappservice/unregister()) 文档用于核对 macOS 13 API 和系统状态含义，不替代规格或实际观察。

## Implementation handoff and boundaries

### Current surface and genuine prerequisites

发布时只读盘点：`AppDelegate.swift` 强持有唯一 `PreviewAppModel`、协调器、Services provider、popover 与共享 `AppAlertPresenter`；`Views/PopoverView.swift` 已有运行/最近区域、失败清单及显式 Open/Stop All/Quit 控制。当前 `macos/GoGrip`、行为测试和唯一工程没有 `SMAppService`/`ServiceManagement` 登录实现。它们是新增能力，不把旧示例、SDK 类型存在或规划描述当成已完成 API。

- 真正宿主前置为 [09](09-recent-target-panel.md)，已满足。5.3 与 5.1/5.2 同属 tasks 第 5 组，不因相邻编号把 10 整票完成变成系统注册行为的硬前置；10 也不依赖 11。新增文本仍必须对接同一标准资源约定，实施前重读 10 当时的实际资源/调用点，不预定另一会话尚未创建的接口。具体共享文件需要协调时先明确边界，不擅自覆盖 10 的工作，也不另造翻译框架规避集成。
- 08 网络共享缺口不作为本票登录选项的硬前置；08/3.3 保持未完成，本票不替代真实卷访问、完整候选或跨平台证明。若后续证据否定已批准内置工具路线，仍按原设计停止相关集成并请求裁决。
- [coverage-plan.md](../coverage-plan.md) 中 12/13 保持未来规划：12 消费完整功能及资源生成候选，13 验证同一候选。发布 11 不授权发布/实施它们，也不预先完成第 6/7 组。

### One truthful system-backed option

- 使用部署下限支持的 `SMAppService.mainApp` 操作当前主 App；不建立 helper bundle、launchd/LaunchAgent plist、旧登录 API 兼容链路、权限提升或新的系统管理框架。
- 真实 `status` 是注册/授权事实源，不增加持久化的期望 bool 并把它当作系统事实。首次尚未注册时显示关闭；已有用户注册显示其实际状态，不为满足“默认关闭”在启动、首次帮助、Services 调用或 App 退出时自动注销。生产系统依赖与隔离测试分开，既有会话/批次测试不能默认触发真实注册或注销。
- `register`/`unregister` 仅由用户明确操作发起。操作返回或抛错后以可读取的实际状态呈现结果，保留可获得的系统原因；不先把请求值当成最终成功，也不把抛错后强制反转 toggle 当成正确恢复。使用所选 SDK 的真实完成边界，不预定同步/异步包装类、通用调度队列、重试或额外竞态保证。
- 区分 SDK 的 `notRegistered`、`enabled`、`requiresApproval` 和 `notFound`，不把所有非 enabled 状态静默折叠成一次成功关闭。`requiresApproval` 表示已注册但需用户在系统设置操作，不是已经允许登录启动；从未注册的主 App 在当前 macOS 实测返回 `notFound`，且该状态下 `register()` 可成功，因此 `notFound` 与 `notRegistered` 同样按“未注册”呈现并提供用户开启，真实结果以 `register()` 返回或抛错后重读的系统状态为准，失败保留可得原因，不编造授权拒绝（2026-10-06 用户裁决，实测证据见本票关闭记录）。界面形式可选择能准确表达状态的最小原生选项，不强制仅用二值 toggle 掩盖待批准状态。
- 需批准时显示真实状态及可执行系统设置检查；若提供打开系统设置入口，也须由用户主动触发，不自动批准、反复注册或更改其他应用的登录项。实际操作失败沿用现有根级一次原生报告和可查看失败路径，不另建通知/报告事实源；待批准说明不必伪装成 SDK 抛错。
- 新增标签/状态/行动指导使用批准的 `en`/`zh-Hans`、英文回退方案，真实系统错误/路径保留。登录选项不改变最近列表、recency、目标身份、运行阶段/URL、浏览器行为或 owned 资源；保持 provider 就绪时序及现有非模态呈现，不引入强制 onboarding 或登录前置。

### Non-goals

不新增独立管理窗口/设置中心、语言开关或高级参数；不重写历史/批次/会话、迁移或清理旧 `go-grip-history`、清全域 UserDefaults、自动恢复/重试/重连；不更改 Services 启用偏好/TCC、不要求 FDA/辅助功能/自动化；不注册其他应用或按进程名清扫。不重做 10 完整语言/引导、08 网络共享、12 universal/archive/DMG/CI、13 候选/支持环境矩阵，不新增 API/资源存在或纯调用转发测试。不提交/推送、发布 Release、同步主规格、归档或正式签名公证。

## Scoped TDD and actual verification

实施前另行批准最小行为 seam 和具体系统操作。本票发布不授权写测试、构建、安装、注册/注销或修改系统设置。

- 5.3 要求的消费者回归仅围绕**用户选择没有变成假成功**：注册/注销被拒绝时，界面结果仍对应实际系统状态并有失败原因；注册后需批准时不被显示为已启用，并有可用指导。使用最小受控系统操作 seam 和确定性完成事件，断言消费状态/失败结果，不仅断言 SDK 方法调用、enum/参数复制、mock echo、文案、默认值或 UI wiring。具体测试边界在实施前批准，不为构造边界新增通用状态/存储框架。
- 既有身份/批次/owner/历史回归复用，不重复矩阵。隔离测试不得改变用户实际登录项或默认存储；受新增 UI 文本影响的纯措辞/实现测试删除，不重钉翻译文字。
- 运行受影响的现有 Swift 行为检查及真实 App 构建，Go 无改动不重跑已知 errcheck；声明/编译/status 模拟不能代替下表 native 操作。证据记录确切 App 路径、代码身份、系统原状态、用户动作、实际 status 和恢复结果，不把同 bundle ID 的其他副本或独立 Swift 脚本的 `mainApp` 当作 GoGrip。

| Actual native scenario | 必须观察的证据 |
|---|---|
| 未注册或已有注册 → 打开面板 | 在确切 App 上读取实际状态；未注册显示关闭且不自行注册，已有用户注册不因启动/打开面板被注销。若原状态不是未注册，不擅自清除以制造首次状态。 |
| 用户明确开启/关闭 | 从真实面板操作，观察系统注册/授权状态与 UI 结果对应；需要批准时有真实说明、不假成功，获准后状态可确认；关闭登录启动后当前宿主/活动预览仍运行。登录项修改须先批准并恢复原状态。 |
| SDK 操作失败或待批准 | 必要消费者回归证明失败/待批准不产生假成功；实际环境能触发时观察原生表面与原因，难触发可用标注并还原的临时受控 UI 补证，不冒充真实系统拒绝或审批成功。 |
| 登录启动关闭、宿主未运行 → Finder Services | 真正退出当前宿主，从 Finder 调用已启用的服务；系统冷启动本轮确切 App，统一 pipeline 得到目标实际 URL、HTTP/默认浏览器正文及 owned 会话，证明登录项不是前提。非 Finder 调用不能替代。 |
| 新增登录 UI 的简中/英文/英文回退 | 观察新增选项、状态与失败/批准指导的真实 native 表面，原始系统原因仍可见；可复用 10 获准的语言环境，不重演全部 Finder 名称矩阵，不以资源存在代替。 |
| 活动服务中登录项操作及关闭 UI/退出 | 活动目标的身份、URL/内容、最近记录及管理入口保持；关闭 UI 不停止，退出仍收尾本 App owned PID/端口，独立 CLI 不受影响；退出不自行注销用户登录项。只复查受影响路径，不重复已关闭强杀矩阵。 |

实际注销/注册可能改变用户已选择的登录偏好。实施前记录当前确切 App/安装副本、注册状态和所需修改，获得具体批准；安装替换保留并核对可回退副本，完成后恢复获准修改的系统/测试状态，不以全局重置代替。真实登出/重登或重启不预定为本票必做操作，更不由发布授权；若另获批准并观察可补充记录，没有观察就不宣称实际登录会话自动启动已验证。缺乏真实注册/关闭、Finder 或新增语言表面的环境时保持相应验收未完成，不能以 fake/status/声明完成 5.3。

## Acceptance

- [x] 重读 current CLI 返回的全部 approved artifacts、09 及 07 修正和 10 当时的实际资源；确认真实宿主前置及共享文件边界，另获 11 范围、必要 TDD seam、检查/native smoke、安装/登录项及数据操作许可，不覆盖另一会话工作。
- [x] 既有菜单栏宿主有明确可达的登录启动选项，生产路径使用 `SMAppService.mainApp` 的实际状态；不恢复空 Settings scene、另建管理主窗口、helper/LaunchAgent 或持久化假系统事实。
- [x] 首次尚未注册时默认关闭，启动/首次帮助/Services 调用/退出均不自行注册或注销；已由用户注册的登录项保持其实际状态，不为“默认关闭”擅自取消。
- [x] 用户主动开启/关闭经过实际 SDK 操作后按真实状态呈现；注册/注销失败保留可得原因并沿用一次原生错误及可查看信息，不保持假成功 toggle，也不以强制反转代替读取事实。
- [x] 需批准与已启用/未注册有真实可理解区别，批准指导可操作；`requiresApproval` 不被当作 enabled，失败或系统状态不被编造为另一原因，不自动批准、重试或更改其他登录项。
- [x] 登录启动关闭、宿主未运行时，实际 Finder 服务仍冷启动本轮 App，通过既有内置 Go/统一批次打开正确浏览器内容与实际 URL；不以登录注册、终端或直接 selector 调用替代。
- [x] 新增登录标签、状态和失败/行动指导接入统一简中/英文/英文回退资源并实际展示，保留系统原始诊断；不重做 10 整票、不另建语言框架或用译文/资源存在断言替代 native。
- [x] 登录项操作/显示不改变运行会话、身份/URL、最近记录/recency或 owned 管理；关闭 UI 服务继续，正常退出确切 owned PID/端口释放且独立 CLI 保持，不擅自注销登录项，不恢复旧服务。
- [x] 获准 seam 的行为回归验证失败/待批准与消费者事实一致，永久测试确定性、隔离及全套安全，不触及用户实际注册/默认域；不新增 SDK wiring、mock echo、enum/文案复制、默认值或重复协议/强杀矩阵测试。
- [x] 受影响 Swift 行为检查与实际 App 构建通过，真实注册/关闭、冷启动 Finder、新增语言表面、活动服务与收尾有本轮证据；受控结果明确标注且还原，未观察实际登录自动启动或支持环境时不宣称已验证。
- [x] 更新 README 登录项操作、系统批准/失败及无需登录启动的服务说明，更新架构/交接的真实状态与已证边界；获准环境修改已恢复、临时资源清理，旧数据/Services偏好/TCC及其他应用不变。
- [x] 全部本票验收及独立 Standards/Spec 审阅通过、无未决范围后才在唯一 tasks 账本勾选 5.3；08/3.3、10/5.1/5.2、12/13及其他票/任务不代勾、不关闭，不发布后续票、不提交/同步/归档。

## Whole-change coverage checkpoint

[coverage-plan.md](../coverage-plan.md) 完整保留 26 checkbox、30 requirement、68 scenario 及每项已发布/未来切片。本轮仅新增 11 并刷新该地图：09 已完成/独立确认，10 open、待由用户新会话处理，11 open、待审阅，其 09 前置已满足；08 open/网络共享缺口保留，12/13仍未发布。planning artifacts、既有票据与 checkbox 不因发布变化，发布核对进度为 15/26；后续以实际 CLI/tasks 和各票验收为准。（关闭时状态见下方 Closure record：10 已按其后关闭记录完成，当前账本 18/26。）

## Closure record

- 关闭：2026-10-06，状态 `done`；12 项验收全部按实际证据通过。实施批准：用户显式批准本票范围（5.3 一体）、S1 单 scoped TDD seam、native smoke 与具体环境操作（`/Applications` 替换并备份、真实注册/注销、System Settings 登录项条目核对/移除、全局 `AppleLanguages` 轻量切换并还原、真实 Finder Services 调用、必要时带标注受控构建）。
- 范围裁决（用户批准，2026-10-06）：实测本机从未注册的主 App 返回 `SMAppService.Status.notFound`（rawValue 3）而非 `.notRegistered`，且该状态下 `register()` 成功、注销后状态变 `.notRegistered`。据此 `.notFound` 与 `.notRegistered` 同作“未注册”呈现（`Not set` + `Turn On`），真实结果以 `register()` 返回/抛错后重读的系统状态为准；票面“One truthful system-backed option”该句与验收第 5 项已按裁决同步修订，实现与该修订一致。
- 变更面：新增 `macos/GoGrip/Services/LoginItemRegistering.swift`（协议 + `MainAppLoginItem`）；修改 `macos/GoGrip/PreviewAppModel.swift`（注入 `loginItem`、`loginItemStatus` 于 init 读取、`refreshLoginItemStatus`、`setLoginItemEnabled`（真实 SDK 调用后重读；抛错发布一次既有根级报告）、`openLoginItemsSettings`）、`macos/GoGrip/Views/PopoverView.swift`（登录行：状态文本、`Turn On`/`Turn Off`、`requiresApproval` 指导与 `Open Login Items…`）、`macos/GoGrip/AppDelegate.swift`（注入 `MainAppLoginItem`、面板打开时刷新、`togglePopover` 主 actor 标注）、`macos/GoGrip/en.lproj/Localizable.strings` 与 `zh-Hans.lproj/Localizable.strings`（+10 键，两表各 63 键对齐）、`macos/GoGrip.xcodeproj/project.pbxproj`（新源文件进入 app/test 两个 target 与 Services 组）、`macos/GoGripTests/PreviewAppModelTests.swift`（+3 回归、`ControlledLoginItem`、23 处 `PreviewAppModel(...)` 调用点注入）、`macos/GoGripTests/FinderServiceProviderTests.swift`（注入）、`README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`、本票（批准修订）。
- Scoped TDD 证据（批准 seam S1）：red 为 `cannot find type 'LoginItemRegistering' in scope` 与 `extra argument 'loginItem' in call`；实现后 3 项转绿——被拒 register、被拒 unregister 各保持系统真实状态并恰好发布一次含系统原因的失败；register 成功但需批准时呈 `requiresApproval`、非 enabled 且不发布失败。变异检验：把“重读真实状态”改为“假定请求生效”后 3 项全部失败，源码已还原。未新增 SDK 调用、enum/文案复制、默认值、资源存在或 wiring 断言。
- 自动检查（最终树）：`make macos-test` **95 tests / 0 failures**（92 → 95）；`make macos` Release `BUILD SUCCEEDED` + `codesign --verify --deep --strict` 通过；产物为 universal、链接 ServiceManagement、含 en/zh-Hans 四份资源；`plutil -lint` 与两表键对齐通过。Go 零改动，未重跑已知 errcheck。
- 真实 smoke（2026-10-06，本机 macOS 27.0.1；最终开发构建 CDHash `d22aee2e…` 安装于 `/Applications`，10 构建 `3949aca2…` 备份于 `~/Desktop/GoGrip-backups/GoGrip-10-backup.app`）：首次未注册状态面板显示 `Not set` + `Turn On`，启动/打开面板不自行注册（BTM 无 GoGrip 条目）；真实 `Turn On` → 状态重读 `Enabled`，`sfltool dumpbtm` 条目 `GoGrip / Type: app / Disposition: [enabled, allowed, notified] / Identifier 2.com.showgp.GoGrip / URL /Applications/GoGrip.app`，`Turn Off` → `Not set`；用户在 System Settings 登录项中实际移除条目后冷启动仍为 `Not set`（状态由系统事实驱动，不假成功）；**登录项关闭、宿主未运行**的真实 Finder Services 调用由 launchd 冷启动宿主（PID 86023，PPID=1）+ owned child 86035（`--managed … -r -- /tmp/gogrip-11-smoke`，127.0.0.1:64631，正文 `cold-start-marker-11 unique`）并打开默认浏览器；活动会话期间 `Turn On` 不改变 PID/端口/URL 与最近记录，关闭面板继续服务，面板 Quit 释放宿主与 owned PID/端口且 BTM 仍保留用户注册（退出不注销），独立 CLI（6419）在 App 退出前后 HTTP 200 与自身内容不变。
- 受控补证（明确标注并还原）：临时受控构建 `667481e7…`（适配器仅报告 `.requiresApproval` 并抛 `NSError(domain: "SMAppServiceErrorDomain", code: 1)`）观察到 `Waiting for approval` + 指导 + `Open Login Items…`/`Turn Off`；按 `Turn Off` 得到一次根提示（`Launch at login` / `Could not change the login startup setting: The operation couldn’t be completed. (SMAppServiceErrorDomain error 1.)`）且面板保持真实状态与可查看失败行；`Open Login Items…` 实际打开 System Settings 的 Login Items 面板。源码还原后重建 CDHash 复现 `d22aee2e…`；不冒充真实系统拒绝或审批成功。
- 语言表面：zh-Hans 真实面板 `登录时启动 / 已开启 / 未开启`（关闭按钮可用），受控构建的批准指导与失败包装为中文且系统原因原文保留；不支持语言（ja）回退英文；全局 `AppleLanguages` 已还原 `en-US`。
- 环境与数据还原：登录项未注册（用户在 System Settings 确认无 GoGrip 条目）、System Settings 窗口关闭；临时目标、AX/点击脚本、截图与源码备份清理；App 自有 `go-grip-first-use-guidance-shown` 与 `go-grip-recent-targets-v1` 恢复为会话前缺失状态；旧 `go-grip-history`、Services 偏好、TCC 与其他应用登录项未动。
- 独立只读审阅：Standards **0 阻塞 / 0 advisory / 0 scope**（3 项新测试各自对应不同被拒/批准结果并断言保留的系统状态与失败报告；未发现未加范围或实现耦合断言）；Spec **0 阻塞 / 0 advisory / 0 scope**（mapped scenarios 由实现与证据覆盖，受控与缺环境边界如实保留）。两轴 worst severity 均为 none，无未决范围裁决。
- 保留的 advisory/限制：`requiresApproval` 与操作失败仅为受控补证（本轮无真实系统拒绝/审批成功）；未执行真实注销/重启登录会话的自动启动观察，不宣称已验证；候选包/Intel/macOS 13（12/13）与 08 网络卷缺口不属本票。
- tasks.md：本票完整交付并验证 **5.3** 并勾选；其他任务（含 08/3.3、12/13 及 6.x/7.x）与本票无关，未代勾、未关闭其他票。
- 未覆盖：未提交/推送、未同步主规格、未归档、未发布或实施 12/13。
