# Whole-change coverage plan

本文件是整个变更的覆盖地图，不是实施票或行为规格；任务进度仅由 [tasks.md](tasks.md) 记录。2026-10-06 本轮按用户“另立/修订票据处理，先提 ticket”请求，发布 [14-partial-tree-preview](tickets/14-partial-tree-preview.md)（open）：补正 13 候选验收发现的“可读根含不可读子树即整会话启动失败”，不新增规格行为。本轮只修改新票与本地图，不修改 proposal、四份 spec、design、tasks、既有票、代码或用户环境。（同日后续：14 已实现、独立 Standards/Spec 两轴审阅通过并关闭 `done`；发布时文本保留，状态以下方为准。）

**事实快照**（2026-10-06 关闭 14 后，并含同日 12 重新交付；历史，保留不回写）：01–07、09–12、14 均 done；08 open/网络共享待补证；13 open，已有旧候选执行记录，组件缺陷修正（14）已完成并通过独立两轴审阅，缺环境/未执行项仍阻塞关闭。CLI 当前 **22/26**，未勾为 3.3、7.1–7.3。用户裁决账本处理：1.1/1.3 勾选保持不动，14 关闭记录按 01/02 + 14 全部贡献核对整项标准（历史 done 证据与新补正贡献不冲突地记录）；**12 的候选重新交付（含 14 Go 修复）已于 2026-10-06 按批准完成**（新 archive/App zip/DMG、包内 renderer managed smoke 与取出 App 的 native/浏览器 smoke 证据见 [12 Redelivery record](tickets/12-universal-candidate.md)；不安装 `/Applications`）；13 的新候选复验已于 2026-10-06 按批准执行（安装新候选后完成 Finder 冷启动→浏览器、递归/单文件/空目录热重载、符号链接/并发/父子独立、5/6/取消/别名/部分失败、面板 Copy/Stop/Stop All/Recent/Clear、正常退出/强杀/ready 前/子进程异常/删除目标/启动失败/watch 降级、CLI 隔离与 TCC 允许等可执行场景；见票 13「新候选复验」记录），仍缺网络/断卷/浏览器拒绝/TCC 拒绝/语言/登录项及 Intel/macOS 13，故 13 保持 open、7.1–7.3 不勾。

**当前事实**（2026-10-07 范围对齐后）：01–07、09–12、14 done；08、13 保持 open，按 2026-10-07 批准的调整后范围收尾。范围调整仅调整验收时点与声明边界、不删除行为要求：Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷验收延期至后续 change（未验证、不阻塞本阶段关闭，见 [tasks.md](tasks.md) 延期清单与下方「当前范围维度映射」）；浏览器失败分支按已批准 A 以既有消费者行为回归验收（候选级真实系统拒绝未触发如实标注，不作为阻塞或通过；不替代正常浏览器路径的实际观察）。本机剩余验收（仍须完成）：TCC 真实拒绝及授权主体、目标移动、服务关闭状态下再次查看 Help、简体中文界面与 Finder 中文服务名（含不支持语言回退）、登录项真实开启/关闭、最终候选外接卷只读访问。CLI 账本 **22/26**（3.3、7.1–7.3 未勾）；spec 计数 **32 requirement / 72 scenario**（含 2026-10-07 新增 2 个验收边界 requirement、4 个 scenario）。08/13 关闭与独立审阅、主规格同步/归档、正式分发仍须另行批准。

**依赖与后续**：14 的前置是已完成的 01/02，不以 13 关闭或网络/Intel/macOS 13 环境为前置。14 已交付 Go 修复、真实 Go smoke 与独立审阅 → **12 已于 2026-10-06 重新交付含 14 修复的新候选**（归档/双架构/DMG 与包内 renderer、取出 App 的 partial-tree smoke 见 12 Redelivery record）→ **13 已于同日在新候选上执行可执行复验**（见票 13「新候选复验」记录）。12/13 后续工作是已发布票的后续贡献，不是提前发布的新票；其他未发布切片无。08→09 顺序例外与正式分发边界不变。2026-10-07 范围对齐后：延期项（Intel 与 macOS 13 实机运行、网络卷、候选级断卷）由后续 change 承接（环境、证据标准与完成标准见 [tasks.md](tasks.md) 延期清单），不再作为本阶段阻塞；本机剩余验收与 08/13 按调整后范围的关闭流程仍待执行；主规格同步/归档仍须另行批准。

## 当前范围维度映射（2026-10-07 范围对齐）

跨 spec 混合场景按维度拆分（不再整条记为未验证）；「本机待验」=本阶段仍须完成，「延期且未验证」=后续 change 承接（见 [tasks.md](tasks.md) 延期清单）。本表分别列出本场景直接证据、相关成功路径证据及旧候选/开发构建对照：相关证据不等于该场景已验证，也不替代「本机待验」列中的最终候选观察；各行证据身份与边界保留，不把对照记录提升为最终候选证明；本表不改变各 requirement 的行为要求。

| requirement / scenario | 已有证据及其适用边界 | 本机待验维度 | 延期且未验证维度 |
|---|---|---|---|
| macos-app-packaging `Universal self-contained application`：`Use the same candidate on supported architectures`、`Use the application without development tools` | 当前 Apple Silicon＋当前 macOS 的候选安装/运行与自包含（新候选直接证据） | — | Intel 实机维度、macOS 13 实机维度 |
| macos-finder-service `Supported open targets`：`Receive a directory outside the home folder` | 主目录外本地目标（新候选直接证据：`/tmp`） | 最终候选外接卷只读访问 | 已挂载网络卷维度 |
| macos-finder-service `First-use service guidance`：`Revisit guidance for a disabled service` | 首启引导与 Help 实际出现/可打开（新候选直接证据） | 服务关闭状态下再次查看 Help | — |
| macos-finder-service `Minimum necessary file authorization`：`Access is denied` | 新候选 POSIX 拒绝（chmod 000）与授权指导、TCC 允许（直接证据）；旧候选/开发构建 TCC 拒绝与主体（对照，不作为本场景最终候选证明） | 最终候选 TCC 真实拒绝及授权主体 | — |
| macos-preview-sessions `Browser opening is separate from service lifetime`：`Browser opening fails after service startup` | 新候选正常打开路径（直接证据：自动打开成功）；关闭标签不停止（既有证据）；失败分支按 A：既有消费者行为回归（受控辅助如使用需另行批准，非该分支必需证据） | — | —（候选级真实系统拒绝未触发：如实标注，不作为阻塞或通过） |
| macos-preview-sessions `Unavailable paths are not empty directories`：`A target is moved or deleted`、`A mounted volume becomes unavailable` | 新候选删除目标（直接证据：unavailable、404 不重连） | 目标移动（最终候选） | 候选级可卸载卷断卷 |
| macos-preview-sessions `Visible hot-reload degradation`：`Watch coverage is incomplete`、`Use a network-volume preview` | 新候选本地卷预算耗尽/walk error 降级、可读内容与手动刷新（直接证据） | — | 网络卷维度（`Use a network-volume preview`） |
| macos-menu-bar-app `Opt-in launch at login`：`Change the login-start preference` | 新候选无登录项冷启动（直接证据）；旧候选真实开关（对照） | 最终候选登录项真实开启/关闭 | — |
| macos-menu-bar-app `System-language localization`：`Use Simplified Chinese`、`Use English or an unsupported language` | 新候选英文界面（直接证据）；旧候选简中/Finder 中文名/回退（对照） | 简中界面与 Finder 中文名、不支持语言回退（最终候选） | — |
| macos-menu-bar-app `Action failures remain understandable and recoverable`：`Recover from browser opening failure` | 同 A：消费者行为回归（失败分支证据） | — | —（真实系统拒绝未触发如实标注） |

## Planned vertical slices and genuine blockers

| Slice | 可验证的交付 | OpenSpec tasks | 真正前置 | 发布状态 |
|---|---|---|---|---|
| 01-managed-preview-lifecycle | 真实 Go 预览启动、回环 URL 与 owner loss 后退出 | 1.1、1.2 | None | 已发布，done |
| 02-managed-status-and-cli | 实际目标/监视状态变化及降级仍可访问，独立 CLI 完整回归 | 1.3、1.4 | 01：managed 事件与资源生命周期（已满足） | 已发布，done |
| 03-owned-process-launch | Foundation 从真实 Go 反馈到可用 URL，失败清理与所有权/SIGKILL 集成 | 2.1、2.5；2.4 的进程层贡献 | 01、02：完整 Go 机器契约（已满足） | 已发布，done |
| 04-target-session-control | 规范化目标单飞/复用、代次隔离与真实单个/全部停止 | 2.2、2.4 的协调器及完整行为 | 03：真实进程适配（已满足） | 已发布，done |
| 05-native-finder-preview（保留已发布 ID） | 单目标手动选择→实际 App/内置 Go→默认浏览器；基础控制与真实退出、旧链路清理 | 3.1 主体、3.2 手动单目标、3.4 单项浏览器/错误；均仅部分贡献 | 04：会话事实源与停止（已满足） | 已发布，done（2026-10-04）；与 06/07 的整合已完成 3.1/3.2/3.4 勾选 |
| [06-native-batch-preview](tickets/06-native-batch-preview.md) | 实际手动多选：全批准备/去重、5/6 确认、取消无副作用、部分失败继续且一次汇总 | 2.3；3.2 手动多选、3.4 批次报告 | 05：真实 App/手动目标/浏览器与退出准入（已满足） | 已发布，done（2026-10-05）；2.3 已勾选，3.2/3.4 已由 05–07 完整整合验证 |
| [07-finder-services-preview](tickets/07-finder-services-preview.md) | Finder 实际 Services 冷启动→统一批次→浏览器；跨入口复用，完整 Services 接通 | 3.1 provider/注册剩余、3.2 Finder 剩余、3.4 Services 入口集成 | 06：完整生产批次（已满足）；间接包含 05 | 已发布，done（2026-10-05）；独立确认通过，3.1/3.2/3.4 已按整项勾选 |
| [08-native-target-access](tickets/08-native-target-access.md) | native→Go 实际授权主体、最小权限/拒绝反馈与本地/受保护/外接/网络卷访问 gate | 3.3 | 07：已接通并验证 Finder/手动完整原生入口（已满足） | 已发布，open；已交付部分（受保护允许/拒绝、真实授权主体、外接卷、指导）经独立确认；2026-10-07 范围对齐：网络卷维度延期（未验证、不阻塞），票面范围/Acceptance/关闭条件已更新；待调整后关闭记录与独立审阅后关闭 |
| [09-recent-target-panel](tickets/09-recent-target-panel.md) | 最近20项、完整状态面板、统一主动重开与清历史不停止 | 4.1、4.2、4.3 | 08 已交付贡献经独立审阅通过且无未决路线裁决（已满足）；批准的阶段性顺序例外与 2026-10-07 调整：网络卷维度延期；其余交付经独立确认，不自动替代 08 关闭流程 | 已发布，done；完整交付/独立确认，三项已勾 |
| [10-localized-service-guidance](tickets/10-localized-service-guidance.md) | 首次/再次Services引导、菜单故障/权限指导、完整native简中/英文/英文回退和真实Finder名称 | 5.1、5.2 | 09完整面板/最近行为交付独立审阅（已满足） | 已发布，done；按当前关闭记录完整交付/独立确认，两项已勾；S-A3仅保留13非阻塞观察 |
| [11-opt-in-login-start](tickets/11-opt-in-login-start.md) | 用户可选登录项、真实系统状态与失败/批准反馈，无登录启动仍能 Services 冷启动 | 5.3 | 09：实际面板/宿主；新增语言表面对接同一资源方案 | 已发布 done；独立审阅与 5.3 勾选见关闭记录 |
| [12-universal-candidate](tickets/12-universal-candidate.md) | 自包含 universal archive/App/DMG、候选 CI artifact 和使用/切换/回退文档 | 6.1、6.2、6.3、6.4 | 10/11 完整交付/审阅已满足；14 修复后重新交付已获批准并执行 | 已发布 done（旧候选 `f5c821df…`/`91af28d6…`）；**2026-10-06 已按 14 修复重新交付新候选**（zip `5062025…`、dmg `5dd330e…`、包内工具 `61040c82…`；构建/静态/包内 renderer 与取出 App smoke 见 12 Redelivery record）；13 新候选复验已于 2026-10-06 执行（见票 13）；延期项见上「当前范围维度映射」 |
| [13-candidate-functional-acceptance](tickets/13-candidate-functional-acceptance.md) | 同一最终候选的真实 Finder/浏览器/身份/批次/面板/失败/所有权与支持环境证据 | 7.1、7.2、7.3 | 12 已重新交付含 14 修复的新候选（2026-10-06）；2026-10-07 范围对齐覆盖延期分类与 A 标准 | 已发布 open；旧候选证据仅作对照；**2026-10-06 新候选复验已执行可执行场景**（记录见票 13）；2026-10-07 范围对齐：延期项与 A 标准已入票面，本机剩余验收待执行，未关闭 |
| [14-partial-tree-preview](tickets/14-partial-tree-preview.md) | 可读根含权限拒绝子树仍提供可读递归内容与手动刷新，结构化 watch 降级；保持根失败和 CLI 边界 | 1.1、1.3 补正；1.4 兼容复验 | 01/02 基础能力（已满足）；不依赖 13 关闭或网络/平台补证 | 已发布 done（2026-10-06 实现 + 独立两轴审阅通过；账本按用户裁决保持 1.1/1.3 已勾并在关闭记录核对；候选重新交付已于同日完成，13 对 partial-tree 路径的候选复验已于同日执行） |

## 原 05 范围保留与编号对照

| 原 05 范围 | 拆分后贡献 | 完整任务勾选条件 |
|---|---|---|
| 2.3：准备、去重、5/6 确认、取消、部分失败/汇总及真实批次验证 | 06 | 06 整票验收/审阅完成；05 不贡献此任务 |
| 3.1：AppKit 根、内置工具/签名、基础运行控制、退出及旧管理器/扩展清理；Services provider 与注册时序 | 05（前半）、07（provider/注册） | 05、07 相关贡献共同验证完整描述；不能在 05 关闭时提前勾选 |
| 3.2：手动选择及多选、全部 file URL、Finder Services/冷启动和跨入口复用 | 05（单目标手动基础）、06（多选）、07（Services/跨入口） | 05–07 相关贡献共同验证完整描述 |
| 3.4：单项默认浏览器结果/恢复、批次一次原生汇总、运行期失败可查看及全部入口集成 | 05（单项/运行期事实）、06（批次）、07（Services） | 05–07 相关贡献共同验证完整描述 |

原 05 范围及稳定 ID 保留。05–07 已共同完成原生入口任务；08 的已交付部分（除网络卷维度外）经独立确认，按调整后范围的关闭流程待执行；09–12、14 已完成并记录各自证据。14 只补正既有 Go 契约，不修改 05–13 的票面，不扩大 08→09 阶段性例外或完整权限 gate。历史发布记录与未来编号按下表解释，当前事实以顶部及实际票/tasks/CLI 为准。

既定 01–14 全部已发布（14 已 done）；**12 重新交付与 13 的新候选复验均已于 2026-10-06 执行**；13 按调整后范围保留本机剩余验收，延期项由后续 change 承接，无新增未发布票。旧票的未来编号不作为新工作授权。

| 原未发布编号 / 切片 | 当前编号 / 切片 |
|---|---|
| 06-native-target-access | 08-native-target-access |
| 07-recent-target-panel | 09-recent-target-panel |
| 08-localized-service-guidance | 10-localized-service-guidance |
| 09-opt-in-login-start | 11-opt-in-login-start |
| 10-universal-candidate | 12-universal-candidate |
| 11-candidate-functional-acceptance | 13-candidate-functional-acceptance |

## Every OpenSpec checkbox

| 实际 checkbox 标签 | 贡献 slice | 发布状态 |
|---|---|---|
| 1.1 | 01（历史完成）、14（初始发现/启动补正） | 01 done、14 done；勾选按用户裁决保持，整项标准已在 14 关闭记录核对 |
| 1.2 | 01 | 已发布 done，已勾；14 保持所有权/清理保证，不重实现 |
| 1.3 | 02（历史完成）、14（子树访问/降级补正） | 02 done、14 done；勾选按用户裁决保持，整项标准已在 14 关闭记录核对 |
| 1.4 | 02（历史完成）、14（受影响 CLI 兼容复验） | 02 done、14 done（兼容复验通过，未发现 CLI 缺陷）；勾选保持 |
| 2.1 | 03 | 已发布，done |
| 2.2 | 04 | 已发布，done |
| 2.3 | 06 | 已发布，done；完整批次验收/审阅通过，checkbox 已勾选 |
| 2.4 | 03（进程层）、04（完整协调器行为） | 已发布，done；04 结合 03 证据完成整个任务，checkbox 已勾选 |
| 2.5 | 03 | 已发布，done |
| 3.1 | 05（App/工具/控制/退出/清理）、07（provider/注册） | 05/07 已发布 done；完整贡献及整项验证已完成，checkbox 已勾选 |
| 3.2 | 05（单目标手动）、06（手动多选）、07（Finder/跨入口） | 05–07 已发布 done；完整贡献及整项验证已完成，checkbox 已勾选 |
| 3.3 | 08 | 已发布 open；07 前置已满足，已交付部分经独立确认；2026-10-07 调整后范围：网络卷维度延期（未验证、不阻塞），其余为关闭条件；待关闭记录与独立审阅后关闭；未勾选 |
| 3.4 | 05（单项浏览器/错误）、06（批次汇总）、07（Services 集成） | 05–07 已发布 done；完整贡献及整项验证已完成，checkbox 已勾选 |
| 4.1 | 09 | 已发布 done；完整交付并独立确认，checkbox已勾 |
| 4.2 | 09 | 已发布 done；完整交付并独立确认，checkbox已勾 |
| 4.3 | 09 | 已发布 done；完整交付并独立确认，checkbox已勾 |
| 5.1 | 10 | 已发布 done；首次/再次查看、关闭服务及指导实际native验收与独立审阅完成，checkbox已勾；开发证明不代13候选 |
| 5.2 | 10 | 已发布 done；完整native/Finder名称实际简中/英文/不支持语言验收与独立审阅完成，checkbox已勾；候选语言实际观察归13 |
| 5.3 | 11 | 已发布 done，已勾；真实系统状态与注册/关闭证据见关闭记录 |
| 6.1 | 12；14 后沿用 12 重新交付 | 12 done、现已勾；2026-10-06 已按 14 修复重建新候选并完成双架构/minos/签名/套件与包内 renderer 实际运行观察（12 Redelivery record） |
| 6.2 | 12；14 后沿用 12 重新交付 | 12 done、现已勾；新 archive/App zip/DMG 已按新身份生成并核对签名、架构、DMG 挂载与取出包结构/运行（12 Redelivery record）；旧候选证据未沿用 |
| 6.3 | 12 | 已发布 done、已勾；CI 配置及本地等价证据见关闭记录，远端上传未观察；14 不修改流水线 |
| 6.4 | 12 | 已发布 done、已勾；14 的说明更新与 2026-10-06 重新交付沿用既有使用/回退说明（本轮未安装 `/Applications`）；13 复验时按实际文档复核 |
| 7.1 | 13；14 修复后新候选受影响路径复验 | 已发布 open、未勾；新候选（2026-10-06）已完成 Finder 冷启动/浏览器、递归/单文件/空目录热重载、符号链接/并发/父子独立、5/6/取消/别名/部分失败、面板 Copy/Stop/Stop All/Recent/Clear 与 TCC 允许；本机待验：最终候选外接卷只读访问；浏览器失败分支按 A（回归证据；真实拒绝未触发如实标注，不作为阻塞）；延期：网络卷（未验证）；未勾 |
| 7.2 | 13；14 修复后新候选降级/失败/隔离复验 | 已发布 open、未勾；新候选已完成正常退出、运行中与 ready 前强杀、子进程异常退出、删除目标、启动失败、watch 降级与独立 CLI 隔离；本机待验：TCC 真实拒绝及授权主体、目标移动；延期：候选级断卷与网络卷降级证据（未验证、不阻塞）；未勾 |
| 7.3 | 13 | 已发布 open、未勾；新候选仅 arm64/当前 macOS 实机运行（见 13 记录）；本机待验：服务关闭时 Help、简中/Finder 中文名与不支持语言回退、登录项真实开关；延期：Intel 与 macOS 13 实机运行（未验证、不阻塞；静态/交叉构建不能代替）；未勾 |

## Every requirement and scenario

下表保留四份 spec 的全部 **32 个 requirement / 72 个 scenario** 原名与共同贡献者（含 2026-10-07 新增的 2 个验收边界 requirement、4 个 scenario）。01–07、09–12、14 done；08/13 open、按调整后范围收尾。14 的补正已交付并独立审阅；12 已重新交付含 14 修复的候选（2026-10-06），13 已在其上执行可执行复验。**维度规则**：各行中的「08 open/网络待补」「卷收尾待完整」等表述按 2026-10-07 调整理解——网络卷与候选级断卷维度延期（未验证、不阻塞）；本机待验项与延期项逐条见上「当前范围维度映射」及票 08/13 调整后 Acceptance；组件 done、全部发布或旧候选通过部分路径，不等于整个 requirement 或变更完成。

### macos-finder-service

行为源：[spec.md](specs/macos-finder-service/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Finder Services entry | `Open a selected folder from Finder`；`Invoke the service while the host is not running` | 07（Services/冷启动）、10（名称/引导）、11（无登录项前提）、12（旧候选文档/冷启动）done；13 open：2026-10-06 新候选实际 Finder 冷启动→浏览器成功（见票 13「新候选复验」）；语言/登录项场景为本机待验（最终候选）；05 不交付 Finder |
| Supported open targets | `Receive a Markdown file`；`Receive a directory outside the home folder`；`Receive an unsupported file` | 04/05/06/07（类型/手动/批次/Finder）、09（最近消费者）done；08 open（网络卷维度延期）；13 open：新候选已验证 CJK/空格单文件、主目录外 `/tmp` 目标、不支持 PNG 拒绝与汇总；TCC 允许的 `~/Documents` 目标归授权场景（见票 13）；14 不新增目标类型或卷权限策略 |
| Batch deduplication and quantity confirmation | `Open exactly five distinct targets`；`Confirm six distinct targets`；`Cancel a large batch`；`Count aliases only once` | 04（身份）、06（完整批次/手动）、07（Finder 多选）、09（共享批次/成功历史）、10（数量确认本地化）done；13 open：新候选实测 5 直接 / 6 确认与取消无副作用 / 6 路径归一 5 不确认（见票 13）；05 不贡献批次 |
| Partial batch failures | `Some targets fail`；`All targets succeed` | 06/07（继续处理/一次汇总）、09（成功会话/报告）、10（本地化）done；08 open（网络卷维度延期；其余经独立确认）；13 open：新候选混合批次保留成功并一次汇总（含权限拒绝与不支持项，见票 13）；14 不改变批次失败规则 |
| First-use service guidance | `Read first-use instructions`；`Revisit guidance for a disabled service` | 10（首次/再看/系统检查）、12（旧候选资源/使用说明）done；13 open：新候选首启引导实际出现；服务关闭状态 Help 为本机待验；14 不重做指导或修改服务偏好 |
| Minimum necessary file authorization | `Access is denied`；`Open an ordinary accessible target` | 05/06/07（原生错误）、09（最近访问判定）、10（权限指导本地化）done；08 open（网络卷维度延期）；13 open：新候选实测 chmod 000 根拒绝与授权指导、TCC Documents 允许（拒绝路径未执行）（见票 13）；14 保持根权限失败边界，不新增授权或绕过权限 |

### macos-preview-sessions

行为源：[spec.md](specs/macos-preview-sessions/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| One session per normalized target | `Reopen a target through a symbolic link`；`Receive concurrent requests for the same target` | 04/05/06/07（身份/入口）、09（最近身份）done；13 open：新候选实测符号链接复用同一 child 与并发同目标单飞（见票 13）；14 保持身份规则 |
| Containment does not merge target identities | `Open a child directory and a file under an active parent` | 04/05/06（目标边界）、09（最近身份）done；13 open：新候选实测父/子/文件三会话独立且子会话以自身为根（见票 13）；14 不改变父/子/文件独立性 |
| Directory and single-file preview modes | `Preview nested documents`；`Preview a selected file`；`Keep an empty directory session`；`Add Markdown to a watched empty directory` | 01/02/05/06/07 历史 done；08 open（网络卷维度延期）；14 done，补正可读根含不可读子树仍服务可读递归内容，并保持单文件/空目录边界；13 open：新候选实测递归、CJK 单文件（Finder 与面板两条路径）、空目录空状态与新增文档热重载（见票 13）；12（新候选包内 renderer 实测）done |
| Verified startup and actual preview URL | `A nondefault port is used`；`Open a file with spaces and non-ASCII characters`；`Startup ends without valid readiness information` | 01/03/05/06/07、09、10、12（旧候选；2026-10-06 新候选包内 renderer ready/真实回环 URL/中文空格单文件 URL 实证）done；08 open；14 done，补正初始发现/真实 URL，根失败仍无假 ready；13 open：新候选实测实际 URL 与页面配对、启动失败无假 running（见票 13） |
| Loopback-only application previews | `Reach a preview from the same machine`；`Attempt to use a nonloopback interface` | 01（实际 bind）、05（App 启动）done；13 open：新候选 owned child 经 `lsof` 实测仅 `127.0.0.1:<port>` 监听（见票 13）；14 保持 App-only 回环契约 |
| Browser opening is separate from service lifetime | `Browser opening fails after service startup`；`Close the browser preview` | 05/06/07、09、10 done；13 open：新候选自动打开成功、关闭标签不停止按既有证据；失败分支按 A（消费者行为回归验收；候选级真实系统拒绝未触发如实标注，不作为阻塞或通过）；14 不改变浏览器契约 |
| Explicit stopping confirms service termination | `Stop one of several sessions`；`Stop all application sessions` | 01/03/04/05、09、12（旧候选；2026-10-06 新候选面板 Stop 实测释放 owned child/端口）done；08 open（网络卷维度延期；外接卷访问、owned 收尾及独立 CLI 隔离仍按调整后 Acceptance 核对）；14 done，已对受影响 Go 服务补做 owned/CLI 隔离 smoke，不代完整 App 证明；13 open：新候选单停/Stop All 与独立 CLI 隔离实测（见票 13） |
| Services end with their owning application | `Quit the application normally`；`Force termination of the host`；`Host exits during startup` | 01/03/04/05/06/07、09、11、12（旧候选正常收尾；2026-10-06 新候选宿主退出释放 owned child 与监听）done；13 open：新候选正常 Quit、运行中 SIGKILL（多 child）与 ready 前强杀均实测释放（见票 13）；14 保持所有权机制 |
| No automatic session restoration or restart | `Relaunch with recent targets`；`A preview process exits unexpectedly` | 04/05、09、11、12（升级/回退说明）done；13 open：新候选实测意外退出无自动重启、Recent 主动重开可行、清史后重启不恢复（见票 13）；14 不新增恢复、重连或重启 |
| Unavailable paths are not empty directories | `A target is moved or deleted`；`A mounted volume becomes unavailable` | 02/04、09、10 历史 done；08 open（候选级可卸载卷断卷验收延期；其他本阶段访问及不可用反馈要求保留）；14 done，保持根/所选单文件访问失败，不因子树权限拒绝污染可读根；12（2026-10-06 新候选：可读根含不可读子树时 0 个 unavailable 报告）done；13 open：新候选删除目标实测面板 unavailable + 保留 Stop、HTTP 404 不重连（见票 13）；移动为本机待验（最终候选） |
| Visible hot-reload degradation | `Watch coverage is incomplete`；`Use a network-volume preview` | 02/04、09、10 历史 done；08 open（网络卷维度延期）；14 done，复用 walk error 状态并保留可读服务/手动刷新，不只补标签；12（2026-10-06 新候选面板显示 `Hot reload degraded: walk error …`、包内 renderer `reload-status degraded` 带原因）done；13 open：新候选面板实测降级显示且可读内容继续可用（见票 13）；网络卷维度（`Use a network-volume preview`）延期（未验证、不阻塞） |
| Independent CLI operation remains available | `Launch the renderer without the application` | 01/02/03/04/05、09、12（CLI release/回退隔离）done；08 open（网络卷维度延期；外接卷访问、owned 收尾及独立 CLI 隔离仍按调整后 Acceptance 核对）；14 done，受影响共享发现的 CLI 兼容与隔离复验；13 open：新候选验收期间独立 CLI（`/Users/ray/go/bin/go-grip`，7801）在单停/Stop All/宿主强杀/Quit 前后持续 200（见票 13） |

### macos-menu-bar-app

行为源：[spec.md](specs/macos-menu-bar-app/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Persistent native menu bar host | `Open and dismiss the panel` | 05/09/10/11/12 done；13 open：新候选安装后实测面板打开/关闭不停止会话、无 Dock 图标（见票 13）；14 不重做宿主 |
| Visible sessions and explicit controls | `Distinguish targets with the same name`；`Copy an active preview address`；`Exit through the panel` | 04/05、09、10 done；13 open：新候选实测同名不同路径区分、Copy URL 实际值、面板 Quit 收尾（见票 13）；14 复用既有状态消费者 |
| Manual opening uses the same target contract | `Manually reopen a Finder-started target`；`Choose a new target manually` | 04/05/06/07、09、10、12（旧候选；2026-10-06 新候选面板手动打开目录与 CJK 单文件实测）done；08 open（网络卷维度延期）；13 open：新候选 Recent 重开与手动路径按同一契约（见票 13）；14 不另建入口 |
| Bounded persistent recent targets | `Open more than twenty distinct targets`；`Reopen an older recent target`；`Restart after using previews` | 09/10 done；13 open：新候选实测 Recent 重开已停止目标新建会话、非空历史重启仅恢复列表且不自动打开/无会话（见票 13）；14 不改变 schema/recency/恢复边界 |
| Clearing history does not stop sessions | `Clear history with an active preview`；`Restart after clearing history` | 09/10 done；13 open：新候选实测有会话清空后会话继续、重启后仍为空（见票 13）；14 不改变清空或停止语义 |
| Opt-in launch at login | `First use without login startup`；`Change the login-start preference` | 07（无登录项冷启动）、11（真实注册/注销/反馈）、12（旧候选消费）done；13 open：新候选无登录项冷启动已实测；登录项真实开关为本机待验（最终候选）；05 不提供设置；14 不操作登录项 |
| System-language localization | `Use Simplified Chinese`；`Use English or an unsupported language` | 10（全表面/Finder 名称）、11（新增登录表面）、12（旧候选资源/消费）done；13 open：新候选英文环境首启/面板/错误实际观察；简中/Finder 中文名与不支持语言回退为本机待验（最终候选）；14 不改资源/新增语言保证 |
| Action failures remain understandable and recoverable | `Open a stale recent target`；`Recover from browser opening failure`；`A service exits unexpectedly` | 04/05/06/07、09、10 done；08 open（网络卷维度延期）；13 open：新候选实测子进程意外退出原因可见、意外退出不自动重启、删除目标显式失败（见票 13）；浏览器失败恢复分支按 A（回归证据；真实拒绝未触发如实标注）；14 保持真实根失败/状态分类 |

### macos-app-packaging

行为源：[spec.md](specs/macos-app-packaging/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Universal self-contained application | `Use the application without development tools`；`Use the same candidate on supported architectures` | 05 done；08 open（网络卷维度延期）；12 done（旧候选；2026-10-06 已按 14 修复重建并重新交付，双架构/minos/签名/包内 renderer 与取出 App 运行见 12 Redelivery record）；13 open：新候选安装后 arm64/当前 macOS 实机运行（无源码/开发 PATH）；Intel 与 macOS 13 实机维度延期（未验证、不阻塞） |
| Candidate DMG and Services-only Finder integration | `Produce an installable candidate`；`Use the new Finder integration` | 05/07/10 done；12 done（旧 archive/App/DMG；2026-10-06 新 archive/zip/DMG 复核：含 Applications 入口、挂载/解包一致、签名与执行位保留）；13 open：新候选经 DMG 安装后 Finder Services 冷启动实测（见票 13）；不新增正式发布流程 |
| Functional acceptance exercises actual user paths | `Verify a real Finder-to-browser launch`；`Verify reuse and explicit stopping`；`Verify abnormal host termination` | 13 open，旧候选已有部分证据且发现组件缺陷；组件 01–07/09–12 历史 done、08 open；14 done（缺陷补正完成）→ 12 已重新交付（2026-10-06）→ 13 已于同日在新候选执行受影响路径与可用矩阵（见票 13），本机待验与延期项见上「当前范围维度映射」（延期不阻塞） |
| Explicitly deferred formal distribution | `Present an unsigned candidate for functional review`；`Complete this implementation stage` | 12 done（旧候选身份/CI/信任说明；2026-10-06 重新交付仍为 ad-hoc 候选、未安装分发，不改变正式分发边界）；13 open，整个阶段未完成；正式 Developer ID/公证/干净安装/公开发布仍另行审批 |
| Explicitly deferred environment-dependent acceptance | `Close this implementation stage with deferred environments`；`Claim support for a deferred environment` | 2026-10-07 范围调整新增的验收边界 requirement（packaging spec）；延期映射见上「当前范围维度映射」与 [tasks.md](tasks.md) 延期清单；非待运行行为，延期维度不得记为已验证或已勾选 |
| Acceptance evidence for the browser-open failure branch | `Accept the browser-open failure branch`；`Use a controlled auxiliary observation` | 2026-10-07 新增（A 裁决）；关联 preview-sessions/menu-bar-app 失败分支与票 13/7.1；候选级真实系统拒绝未触发如实标注，不作为阻塞或通过 |

本轮完整保留 26 个 checkbox、**32 个 requirement、72 个 scenario**（含 2026-10-07 新增 2 个验收边界 requirement、4 个 scenario）；全部行为范围分配至已发布 01–14 及 12 的重新交付与 13 的新候选复验，无未分配范围或未发布票。14（done）是既有 1.1/1.3 的缺陷补正及 1.4 兼容复验，已按批准范围实施并通过独立 Standards/Spec 两轴审阅；1.1/1.3/1.4 勾选按用户裁决保持不变，整项标准已按 01/02 + 14 全部贡献在 14 关闭记录核对。12 的候选重新交付已于 2026-10-06 完成并记录；13 已于同日在新候选上执行并记录获准且可执行的复验路径，本机剩余验收仍待完成。当前 CLI **22/26** 保持，3.3、7.1–7.3 未勾，08/13 保持 open；2026-10-07 调整后：延期项记录为未验证、不阻塞关闭，本机剩余验收与关闭流程待执行；验收时点调整不将任何待验行为或新增验收政策记为通过；主规格同步/归档仍须另行批准。
