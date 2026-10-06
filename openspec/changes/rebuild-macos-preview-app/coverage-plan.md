# Whole-change coverage plan

本文件是整个变更的覆盖地图，不是实施票，也不是行为规格；任务进度仅由 [tasks.md](tasks.md) 记录。本轮在用户确认12发布审阅周期后，仅发布 [13](tickets/13-candidate-functional-acceptance.md) 并刷新当前事实：10已按关闭记录完成并独立审阅，5.1/5.2已勾；11/12仍open，13直接实施前置为12完整最终候选。08→09顺序例外和网络缺口归属按批准 [design.md](design.md) Migration Plan及既有票保留。不修改proposal、四份spec、design、tasks或既有票，不新增行为、勾选或实施授权；至此既定01–13全部发布，发布完成不代表实施/验收或正式分发完成。

已发布 [01](tickets/01-managed-preview-lifecycle.md)（1.1、1.2）、[02](tickets/02-managed-status-and-cli.md)（1.3、1.4）、[03](tickets/03-owned-process-launch.md)（2.1、2.5；2.4进程层）、[04](tickets/04-target-session-control.md)（2.2；结合03完成2.4）、[05](tickets/05-native-finder-preview.md)、[06](tickets/06-native-batch-preview.md)、[07](tickets/07-finder-services-preview.md)，均done，05–07完整贡献已完成3.1/3.2/3.4，07共享非模态提示修正独立确认。[08](tickets/08-native-target-access.md)仍open，已交付部分独立确认、网络共享待补证。[09](tickets/09-recent-target-panel.md) done，4.1–4.3已勾。[10](tickets/10-localized-service-guidance.md)现为done，5.1/5.2完整交付并独立Standards/Spec确认、两项已勾；开发构建/三种语言证据不代候选证明，S-A3仅保留非阻塞观察。[11](tickets/11-opt-in-login-start.md)覆盖5.3、open。[12](tickets/12-universal-candidate.md)覆盖6.1–6.4、open；10前置已满足，11仍待完整交付。[13](tickets/13-candidate-functional-acceptance.md)本轮发布open、待审阅，覆盖7.1–7.3；12完整最终候选前置尚未满足。

**01–07、09、10已发布done；08 open/网络共享待补证；11/12 open；13本轮新发布open、待审阅、12前置待满足；既定01–13全部发布，无未发布切片。** 08已交付的真实授权主体、受保护允许/拒绝、外接卷及失败指导经独立确认，无未决路线裁决；完整权限/网络gate未通过，环境具备后回08补证并独立审阅。若真实证据否定内置工具路线，停止相关集成并重评，不增helper/detach/FDA兜底。09关闭及07非模态共享修正已确认，FIFO按presenter入槽顺序理解，不扩张为全局报告顺序。10关闭记录取代其发布时进行中快照；本轮只记录，不重跑或冒认其开发native/语言结果。11/12/13发布不许可代码、测试、候选执行、安装或偏好操作，不覆盖其他会话工作。12消费完整功能/资源，13只验证最终候选与实际平台/卷，不接收组件遗漏，不代签08，也不新增S-A3焦点保证；App候选artifact与CLI原release隔离。本轮核对账本 **17/26**，5.1/5.2已勾，3.3、5.3、6.1–6.4、7.1–7.3仍未勾；以实际CLI/tasks与各票证据为准，其他票历史快照保留。

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
| [08-native-target-access](tickets/08-native-target-access.md) | native→Go 实际授权主体、最小权限/拒绝反馈与本地/受保护/外接/网络卷访问 gate | 3.3 | 07：已接通并验证 Finder/手动完整原生入口（已满足） | 已发布，open；已交付部分通过本会话独立确认，网络共享证据待补、整票未完成；本轮不变更其票面 |
| [09-recent-target-panel](tickets/09-recent-target-panel.md) | 最近20项、完整状态面板、统一主动重开与清历史不停止 | 4.1、4.2、4.3 | 08 已交付贡献经独立审阅通过且无未决路线裁决（已满足）；批准的阶段性顺序例外保留网络共享缺口，不替代08完整验收 | 已发布，done；完整交付/独立确认，三项已勾 |
| [10-localized-service-guidance](tickets/10-localized-service-guidance.md) | 首次/再次Services引导、菜单故障/权限指导、完整native简中/英文/英文回退和真实Finder名称 | 5.1、5.2 | 09完整面板/最近行为交付独立审阅（已满足） | 已发布，done；按当前关闭记录完整交付/独立确认，两项已勾；S-A3仅保留13非阻塞观察 |
| [11-opt-in-login-start](tickets/11-opt-in-login-start.md) | 用户可选登录项、真实系统状态与失败/批准反馈，无登录启动仍能 Services 冷启动 | 5.3 | 09：实际面板/宿主（已满足）；不以10整票完成为硬前置，新增语言表面对接同一资源方案 | 已发布，open；发布审阅周期获用户确认，不代表实施/验收或系统注册修改许可 |
| [12-universal-candidate](tickets/12-universal-candidate.md) | 自包含universal archive/App/DMG、候选CI artifact和实际使用/切换/回退文档 | 6.1、6.2、6.3、6.4 | 10完整交付/审阅（已满足）；11完整功能/资源及审阅（尚未满足）；开发包来自05，不代08或13证明 | 已发布，open；用户确认继续发布13不代表12实施/验收完成 |
| [13-candidate-functional-acceptance](tickets/13-candidate-functional-acceptance.md) | 同一最终候选的真实Finder/浏览器/身份/批次/面板/失败/所有权与实际支持环境证据 | 7.1、7.2、7.3 | 12完整最终可运行候选和独立审阅（尚未满足）；实际平台/卷和必要操作许可 | 本轮已发布，open、待审阅；验证-only，不接收组件遗漏；缺必需环境/证据不能关闭 |

## 原 05 范围保留与编号对照

| 原 05 范围 | 拆分后贡献 | 完整任务勾选条件 |
|---|---|---|
| 2.3：准备、去重、5/6 确认、取消、部分失败/汇总及真实批次验证 | 06 | 06 整票验收/审阅完成；05 不贡献此任务 |
| 3.1：AppKit 根、内置工具/签名、基础运行控制、退出及旧管理器/扩展清理；Services provider 与注册时序 | 05（前半）、07（provider/注册） | 05、07 相关贡献共同验证完整描述；不能在 05 关闭时提前勾选 |
| 3.2：手动选择及多选、全部 file URL、Finder Services/冷启动和跨入口复用 | 05（单目标手动基础）、06（多选）、07（Services/跨入口） | 05–07 相关贡献共同验证完整描述 |
| 3.4：单项默认浏览器结果/恢复、批次一次原生汇总、运行期失败可查看及全部入口集成 | 05（单项/运行期事实）、06（批次）、07（Services） | 05–07 相关贡献共同验证完整描述 |

05保留稳定ID/文件名 `05-native-finder-preview`，交付是单目标手动预览；05–07已共同完成原生入口任务，08已交付部分独立确认/网络待补，09与10已完成。本轮仅新增13并刷新当前地图，11/12保持open；不另建旧整票或修改已有票。08→09阶段性例外不代表完整权限gate完成。原大票范围保持在05–07，07共享提示修正同票闭合，不以拆分删除真实证明。当前依赖按批准artifacts/票据和本地图，进度按实际账本/验收；历史发布快照保留。

原未发布规划编号与当前稳定编号对照如下；目前既定01–13均已发布，无剩余未发布切片，既有票面及历史记录保持原样。旧票的未来编号按此表理解，当前范围/依赖以本地图及相应已发布票为准，不将历史编号或全部发布当作实施授权。

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
| 1.1 | 01 | 已发布，done |
| 1.2 | 01 | 已发布，done |
| 1.3 | 02 | 已发布，done |
| 1.4 | 02 | 已发布，done |
| 2.1 | 03 | 已发布，done |
| 2.2 | 04 | 已发布，done |
| 2.3 | 06 | 已发布，done；完整批次验收/审阅通过，checkbox 已勾选 |
| 2.4 | 03（进程层）、04（完整协调器行为） | 已发布，done；04 结合 03 证据完成整个任务，checkbox 已勾选 |
| 2.5 | 03 | 已发布，done |
| 3.1 | 05（App/工具/控制/退出/清理）、07（provider/注册） | 05/07 已发布 done；完整贡献及整项验证已完成，checkbox 已勾选 |
| 3.2 | 05（单目标手动）、06（手动多选）、07（Finder/跨入口） | 05–07 已发布 done；完整贡献及整项验证已完成，checkbox 已勾选 |
| 3.3 | 08 | 已发布 open；07 前置已满足，已交付部分经独立确认，网络共享待补证；整项未完成/未勾选 |
| 3.4 | 05（单项浏览器/错误）、06（批次汇总）、07（Services 集成） | 05–07 已发布 done；完整贡献及整项验证已完成，checkbox 已勾选 |
| 4.1 | 09 | 已发布 done；完整交付并独立确认，checkbox已勾 |
| 4.2 | 09 | 已发布 done；完整交付并独立确认，checkbox已勾 |
| 4.3 | 09 | 已发布 done；完整交付并独立确认，checkbox已勾 |
| 5.1 | 10 | 已发布 done；首次/再次查看、关闭服务及指导实际native验收与独立审阅完成，checkbox已勾；开发证明不代13候选 |
| 5.2 | 10 | 已发布 done；完整native/Finder名称实际简中/英文/不支持语言验收与独立审阅完成，checkbox已勾；候选语言实际观察归13 |
| 5.3 | 11 | 已发布 open、09前置已满足；发布确认不是实施完成，真实注册/关闭和失败/批准反馈须验收，checkbox未勾 |
| 6.1 | 12 | 已发布 open；宿主/Go架构/部署、实际App/Go运行及相关套件须验收；10前置已满足、11待满足，checkbox未勾 |
| 6.2 | 12 | 已发布 open；显式archive、工具/宿主签名、DMG与自包含包内预览须验收；10前置已满足、11待满足，checkbox未勾 |
| 6.3 | 12 | 已发布 open；实际同等CI命令/候选artifact、切断App正式Release且CLI流程不变须验收；10前置已满足、11待满足，checkbox未勾 |
| 6.4 | 12 | 已发布 open；候选实际使用/切换/回退说明须验收，旧数据保留与正式边界不变；10前置已满足、11待满足，checkbox未勾 |
| 7.1 | 13 | 本轮已发布 open；同一最终候选真实Finder/browser、目标/批次/面板/历史路径须验收；12前置待满足，checkbox未勾 |
| 7.2 | 13 | 本轮已发布 open；最终候选正常/异常退出、真实权限/失效/卷/watch与owned/CLI隔离须验收；08网络缺口保留，checkbox未勾 |
| 7.3 | 13 | 本轮已发布 open；同一候选实际Apple Silicon/Intel、最低macOS13与证据汇总须验收；缺必需环境/路径保持阻塞，checkbox未勾 |

## Every requirement and scenario

下表逐一保留四份spec全部requirement/scenario原名及完整共同贡献者。01–07、09、10 done；08已发布open/已交付部分独立确认/网络待补；11/12 open，12的10前置已满足而11待满足；13本轮发布open、12前置待满足。所有贡献slice均已发布，无未分配范围；组件完成、全部发布或12短smoke不等于整个App requirement完成。13只验证最终候选，静态/真实/受控/缺环境分开，不接收组件遗漏、代签08或正式分发；S-A3仅非阻塞观察。

### macos-finder-service

行为源：[spec.md](specs/macos-finder-service/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Finder Services entry | `Open a selected folder from Finder`；`Invoke the service while the host is not running` | 07（真实 Services/冷启动）已发布 done；10（服务名称本地化及使用说明，不重做 provider）已发布done；11（保持无登录启动的真实Services冷启动）已发布open；12（从交付包取出的候选按文档实际Finder冷启动，不替代完整矩阵）已发布open；13 本轮已发布open；05 不交付 Finder |
| Supported open targets | `Receive a Markdown file`；`Receive a directory outside the home folder`；`Receive an unsupported file` | 04（目标）、05（手动目标基础）、06（批次分类）、07（Finder 入口）已发布 done；08（真实目标/卷访问）已发布 open、已交付部分经独立确认/网络共享待补证；09（最近消费者保持既有目标契约）已发布 done；13 本轮已发布open |
| Batch deduplication and quantity confirmation | `Open exactly five distinct targets`；`Confirm six distinct targets`；`Cancel a large batch`；`Count aliases only once` | 04（身份）、06（完整批次/手动多选）、07（Finder 多选）已发布 done；09（保留共享批次、成功历史不提前提交）已发布 done；10（原生数量确认本地化，不改批次语义）已发布done；13 本轮已发布open；05 不贡献批次 |
| Partial batch failures | `Some targets fail`；`All targets succeed` | 06（继续处理/一次汇总）、07（Finder 接入）已发布 done；08（真实访问失败反馈）已发布 open、已交付部分经独立确认/网络共享待补证；09（最近/完整面板保持成功与一次报告语义）已发布 done；10（根报告/操作指导本地化，保留原始原因）已发布done；13 本轮已发布open |
| First-use service guidance | `Read first-use instructions`；`Revisit guidance for a disabled service` | 10（首次/再看及菜单不可见/关闭服务说明，不改系统服务偏好）已发布done；12（候选消费完整引导并提供实际使用文档，不重做10）已发布open；13（最终候选实际native/Finder/系统状态观察，不重复组件实现）本轮已发布open |
| Minimum necessary file authorization | `Access is denied`；`Open an ordinary accessible target` | 05（单项原生错误/普通目标）、06（批次错误）、07（Services 错误）已发布 done；08（真实权限 gate/适用指导）已发布 open、已交付部分经独立确认/网络共享待补证；09（最近入口沿用访问判定/指导）已发布 done；10（既有条件式权限指导本地化，不扩大权限）已发布done；13 本轮已发布open |

### macos-preview-sessions

行为源：[spec.md](specs/macos-preview-sessions/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| One session per normalized target | `Reopen a target through a symbolic link`；`Receive concurrent requests for the same target` | 04（身份）、05（手动复用）、06（批次别名）、07（跨入口）已发布 done；09（最近身份/主动重开消费）已发布 done；13 本轮已发布open |
| Containment does not merge target identities | `Open a child directory and a file under an active parent` | 04、05（手动目标边界）、06（批次独立计数/打开）已发布 done；09（最近记录不按包含关系合并）已发布 done；13 本轮已发布open |
| Directory and single-file preview modes | `Preview nested documents`；`Preview a selected file`；`Keep an empty directory session`；`Add Markdown to a watched empty directory` | 01（启动/内容）、02（新增文档/watch）、05（手动原生打开）、06（多选打开）、07（Finder 打开）已发布 done；08（访问与真实空目录对照）已发布 open、已交付部分经独立确认/网络共享待补证；13 本轮已发布open |
| Verified startup and actual preview URL | `A nondefault port is used`；`Open a file with spaces and non-ASCII characters`；`Startup ends without valid readiness information` | 01（服务/URL）、03（消费/失败清理）、05（手动原生结果）、06（批次失败）、07（Finder 结果）已发布 done；08（真实访问成功/失败）已发布 open、已交付部分经独立确认/网络共享待补证；09（成功历史时机/完整面板实际 URL 消费）已发布 done；10（native 启动错误说明本地化，保持真实 URL/诊断）已发布done；12（交付包内真实Go与中文/空格文件实际URL/浏览器内容smoke）已发布open；13 本轮已发布open |
| Loopback-only application previews | `Reach a preview from the same machine`；`Attempt to use a nonloopback interface` | 01（真实回环 bind）、05（App 启动路径）已发布 done；13（候选实际接口）本轮已发布open |
| Browser opening is separate from service lifetime | `Browser opening fails after service startup`；`Close the browser preview` | 05（单项默认浏览器/恢复）、06（批次结果）、07（Finder 接入）已发布 done；09（完整面板恢复入口/浏览器失败不撤销成功历史）已发布 done；10（native 浏览器失败说明/操作本地化，不改生命周期）已发布done；13 本轮已发布open |
| Explicit stopping confirms service termination | `Stop one of several sessions`；`Stop all application sessions` | 01（Go 退出）、03（适配器）、04（单个/全部）、05（原生动作）已发布 done；08（本轮访问目标/卷 owned 收尾验证）已发布 open、已交付部分经独立确认/网络共享待补证；09（完整面板动作/清历史不停止）已发布 done；12（候选按文档实际停止并观察owned资源释放，不替代完整矩阵）已发布open；13 本轮已发布open |
| Services end with their owning application | `Quit the application normally`；`Force termination of the host`；`Host exits during startup` | 01（Go owner loss）、03（Foundation 所有权）、04（starting 停止）、05（App quit/准备准入）、06（确认恢复准入）、07（Services 请求接入）已发布 done；09（完整面板保持共享退出路径）已发布 done；11（登录项操作不停止当前宿主/服务，退出仍沿用owned收尾）已发布open；12（候选正常停止/退出使用smoke，不代异常退出矩阵）已发布open；13 本轮已发布open |
| No automatic session restoration or restart | `Relaunch with recent targets`；`A preview process exits unexpectedly` | 04（退出/不重启）、05（原生根/退出显示）已发布 done；09（历史恢复不访问/启动、完整退出原因与主动重开）已发布 done；11（注册与宿主重启不恢复旧服务）已发布open；12（升级/回退不恢复旧服务或迁移旧数据的说明，不增恢复逻辑）已发布open；13 本轮已发布open |
| Unavailable paths are not empty directories | `A target is moved or deleted`；`A mounted volume becomes unavailable` | 02（Go 状态）、04（状态事实）已发布 done；08（实际访问/专用卷验证）已发布 open、已交付部分经独立确认/网络共享待补证；09（完整面板/陈旧最近路径反馈）已发布 done；10（native 状态/指导本地化，保留原始原因）已发布done；13 本轮已发布open |
| Visible hot-reload degradation | `Watch coverage is incomplete`；`Use a network-volume preview` | 02（状态/可用内容）、04（状态事实）已发布 done；08（实际网络卷与已知降级）已发布 open、该网络卷访问/降级证据待补；09（完整状态/原因显示与浏览器手动刷新）已发布 done；10（native degraded/原因说明本地化，不补网络证据）已发布done；13 本轮已发布open |
| Independent CLI operation remains available | `Launch the renderer without the application` | 01（所有权隔离）、02（完整 CLI）、03（owned 操作）、04（协调器）、05（App 操作/退出隔离）已发布 done；08（本轮收尾隔离验证）已发布 open、已交付部分经独立确认/网络共享待补证；09（完整面板/清除及停止隔离）已发布 done；12（App构建/候选CI与CLI原release流程隔离、使用/回退说明）已发布open；13 本轮已发布open |

### macos-menu-bar-app

行为源：[spec.md](specs/macos-menu-bar-app/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Persistent native menu bar host | `Open and dismiss the panel` | 05（原生宿主）已发布 done；09（完整面板/历史切换保留显示与生命周期分离）已发布 done；10（首次/重看帮助仍保持原生宿主边界）已发布done；11（既有宿主的登录选项、关闭UI不停止）已发布open；12（候选消费完整宿主并实际使用，不另建UI）已发布open；13 本轮已发布open |
| Visible sessions and explicit controls | `Distinguish targets with the same name`；`Copy an active preview address`；`Exit through the panel` | 04（状态/停止）、05（基础控制/退出）已发布 done；09（完整面板/明确状态与动作）已发布 done；10（完整状态标签/操作本地化，不改事实源）已发布done；13 本轮已发布open |
| Manual opening uses the same target contract | `Manually reopen a Finder-started target`；`Choose a new target manually` | 04（统一身份）、05（单目标手动）、06（多选）、07（Finder/手动跨入口）已发布 done；08（共享 native 访问/反馈）已发布 open、已交付部分经独立确认/网络共享待补证；09（最近重开/统一成功历史）已发布 done；10（native 手动入口应用文本本地化，不另建 pipeline）已发布done；12（候选取出后实际手动目录/单文件预览，不重做入口）已发布open；13 本轮已发布open |
| Bounded persistent recent targets | `Open more than twenty distinct targets`；`Reopen an older recent target`；`Restart after using previews` | 09（版本化至多 20 项/规范化去重/recency/跨启动只恢复记录）已发布 done；10（最近区域/操作本地化，不改变存储和顺序）已发布done；13 本轮已发布open |
| Clearing history does not stop sessions | `Clear history with an active preview`；`Restart after clearing history` | 09（仅清新记录/跨启动为空/活动管理与服务不变）已发布 done；10（清历史操作本地化，不改清除语义）已发布done；13 本轮已发布open |
| Opt-in launch at login | `First use without login startup`；`Change the login-start preference` | 07（无登录项Finder冷启动）已发布done；11（用户明确注册/注销、真实状态和失败/批准反馈、无登录启动的Services冷启动）已发布open、09前置已满足；12（候选消费最终登录UI/真实状态并保持非前提，不重复11实现）已发布open；05不提供登录设置；13（最终候选实际native/Finder/系统状态观察，不重复组件实现）本轮已发布open |
| System-language localization | `Use Simplified Chinese`；`Use English or an unsupported language` | 10（完整native/错误/帮助及实际Finder服务名称，简中/英文/不支持语言回退）已发布done；11（新增登录UI的标签/真实状态/行动指导复用同一资源方案并观察新增表面）已发布open，不替代10完整5.2或交付其引导；12（候选消费最终完整资源，不以拷贝存在代运行或补10缺口）已发布open；13（最终候选实际native/Finder/系统状态观察，不重复组件实现）本轮已发布open |
| Action failures remain understandable and recoverable | `Open a stale recent target`；`Recover from browser opening failure`；`A service exits unexpectedly` | 04（真实退出）、05（单项原生报告/恢复）、06（批次汇总）、07（Services 接入）已发布 done；08（访问失败及适用检查指导）已发布 open、已交付部分经独立确认/网络共享待补证；09（历史/完整面板/陈旧错误与主动恢复）已发布 done；10（完整 native 错误标签/行动指导本地化，保留原始诊断）已发布done；13 本轮已发布open |

### macos-app-packaging

行为源：[spec.md](specs/macos-app-packaging/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Universal self-contained application | `Use the application without development tools`；`Use the same candidate on supported architectures` | 05（确定内置工具/开发签名）已发布 done；08（开发 App 内置工具/native 访问路线）已发布 open、已交付部分经独立确认/网络共享待补证；12（宿主/Go双架构与部署检查、自包含候选、实际本机运行）已发布open；13（实际支持平台运行矩阵）本轮已发布open |
| Candidate DMG and Services-only Finder integration | `Produce an installable candidate`；`Use the new Finder integration` | 05（去旧扩展/转发）、07（新 Services 接通）已发布 done；10（完整 native 资源/ServicesMenu/引导供候选消费）已发布done；12（显式archive App/DMG、签名、候选artifact和实际使用文档/smoke）已发布open；13（完整候选Finder矩阵）本轮已发布open |
| Functional acceptance exercises actual user paths | `Verify a real Finder-to-browser launch`；`Verify reuse and explicit stopping`；`Verify abnormal host termination` | 13（完整跨组件/异常退出/平台矩阵）本轮已发布open；12（实际交付候选并做构建/使用/正常收尾smoke，为13提供同一产物）已发布open；组件01–07 done、08已发布open且已交付部分经独立确认/网络共享待补证、09已发布done、10已发布done/5.1与5.2已勾并独立审阅、11已发布open，各自须真实smoke；开发App或12短smoke不替代整个requirement及跨支持环境证明 |
| Explicitly deferred formal distribution | `Present an unsigned candidate for functional review`；`Complete this implementation stage` | 12（候选标识/CI artifact/信任限制、CLI原release隔离、切换/回退说明）已发布open；13（完整阶段验收与交付边界）本轮已发布open；不安排正式分发，不降低后续Developer ID/公证/干净安装标准 |

本轮完整保留26个OpenSpec checkbox、30个requirement、68个scenario，全部范围已分配至既定01–13，至此全部票据已发布。仅13本轮新增open、覆盖7.1–7.3、待审阅且12实施前置待满足。10当前done/5.1与5.2已勾据关闭记录刷新；11/12 open，08网络缺口保留。当前核对账本17/26，3.3、5.3、6.1–6.4、7.1–7.3未勾；发布不修改既有票/批准artifacts/进度，不授予实施、环境操作、关闭、同步/归档或正式发布权。
