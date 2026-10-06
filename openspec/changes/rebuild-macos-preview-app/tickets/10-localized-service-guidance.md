# 10 — Finder/Services 首次引导与完整原生双语体验

Status: done
Blocked by: 09-recent-target-panel

## Deliverable

在 09 的完整原生面板上交付**首次启动及可再次查看的 Finder/Services 简短引导**，并让整个 native 使用路径——面板、数量确认、错误与操作指导、帮助和实际 Finder Services 命令名称——按系统语言使用简体中文或英文，其他语言回退英文。不以添加 strings、翻译现有不完整 popover 或修改 plist 代替完整可用体验。

本票完整覆盖 OpenSpec **5.1、5.2**，不覆盖 **5.3**。5.1 必须保留首次/再次查看、服务未出现时的系统设置检查、权限指导、不自动启用及不将重扫当作菜单可见性证明；5.2 必须保留三种语言环境、完整 native 表面、原始诊断和真实 Finder 菜单验证。行为权威是本 change 的批准 spec/design，任务进度仅由 [tasks.md](../tasks.md) 记录，票内清单不是另一本进度账本。

**发布时状态**：01–07 done；08 已交付的真实授权主体、受保护目录允许/拒绝、外接卷及失败指导已通过独立 Standards/Spec 确认，但整票仍 open，真实网络共享待补证，3.3 未勾。用户报告 09 正在独立会话实施；本会话不判定其具体授权，不宣称其整票完成，当前 4.1–4.3 未勾。10 的真正前置是 09 完成并经独立审阅的完整面板及最近目标行为；发布可以先进行，不能在其进行中代码上提前实施或把旧简化面板当交付。

08→09 阶段性顺序例外保持不变；本票不替代网络共享证明、不将 08/3.3 判定完成、不另加权限路线兜底。若后续真实证据否定直接内置工具路线，按已批准设计停止相关集成并重评。11 登录项、12 候选打包、13 候选级真实验收另票交付，不构成 10 的反向前置。

**发布不是实施许可**。本轮只发布本票及更新覆盖地图；不修改其他票、已批准 artifacts、checkbox、应用代码或用户环境。实施仍须单票审阅、明确批准范围、必要的 scoped TDD seams、native smoke 和具体数据/环境操作；不从用户另会话实施 09 推导 10 的编码或系统设置授权。

## Source mapping

- [proposal.md](../proposal.md)：原生菜单栏宿主、Finder Services、简单双语使用体验；不恢复旧 Finder Sync/转发路径。
- [tasks.md](../tasks.md)：**5.1** 首次及可再次查看的 Services 引导、菜单故障/权限说明及真实操作验收；**5.2** 完整系统语言本地化与 ServicesMenu、真实三种语言/native/Finder 验收。完整物理单行描述均须实现，不缩为标题摘要；5.3 留给 11。
- [design.md](../design.md)：Decision 1 原生根/注册时序与 macOS 13 基线；Decision 2 单一 Services、默认精确键/ServicesMenu 与系统偏好边界；Decision 6 真实失败报告；Decision 8 最小权限和实际授权证据；Decision 9 根级错误、语言与旧数据边界；Verification、Risks、Migration Plan 的真实 native 证明及 08→09 例外。
- [09](09-recent-target-panel.md)：10 消费其完整运行/最近目标面板、明确 controls、target/reload/失败原因及统一主动重开；不重新实现最近存储或改变其成功提交、清历史和生命周期语义。实施前读取实际完成后的源/API，不锁定本票发布时的中间接口。

| Spec requirement | 全部相关 scenario 原名 | 本票贡献与保持的既有契约 |
|---|---|---|
| [macos-finder-service](../specs/macos-finder-service/spec.md) / First-use service guidance | `Read first-use instructions`；`Revisit guidance for a disabled service` | 首次帮助、可再次查看、Finder 入口/系统 Services 设置排查和权限说明；不改服务偏好或假称菜单已出现。 |
| macos-finder-service / Finder Services entry | `Open a selected folder from Finder`；`Invoke the service while the host is not running` | 实际服务名称本地化和帮助说明；保持 07 的唯一 provider、冷启动、同一批次及系统菜单/用户偏好控制。 |
| macos-finder-service / Batch deduplication and quantity confirmation | `Open exactly five distinct targets`；`Confirm six distinct targets`；`Cancel a large batch`；`Count aliases only once` | 已有原生数量确认的标题、说明和操作本地化；数量、去重/取消语义仍归已完成 04/06/07，不重做批次矩阵。 |
| macos-finder-service / Partial batch failures | `Some targets fail`；`All targets succeed` | 既有一次根级失败汇总/可查看报告的本地化；实际目标/底层原因保留，全部成功不增加成功提示。 |
| macos-finder-service / Minimum necessary file authorization | `Access is denied`；`Open an ordinary accessible target` | 08 已有条件式访问检查/系统设置指导本地化，普通可访问目标不误加权限提示；不推导 FDA、Accessibility 或 Automation 要求。 |
| [macos-menu-bar-app](../specs/macos-menu-bar-app/spec.md) / System-language localization | `Use Simplified Chinese`；`Use English or an unsupported language` | 本票主要完整语言契约：完整面板、错误/行动指导、帮助和实际 Finder Services 名称；简中/英文跟随系统，其他语言英文回退，无应用内语言设置。 |
| macos-menu-bar-app / Persistent native menu bar host | `Open and dismiss the panel` | 引导/再次查看仍是原生正常对话框或面板入口，不另建管理主窗口，不把关闭 UI 变成停止服务。 |
| macos-menu-bar-app / Visible sessions and explicit controls | `Distinguish targets with the same name`；`Copy an active preview address`；`Exit through the panel` | 完整 09 面板的运行/最近区域、阶段/可用性/reload 状态、原因、空状态、操作和报告本地化；路径/真实 URL 与 controls 作用对象不变。 |
| macos-menu-bar-app / Manual opening uses the same target contract | `Manually reopen a Finder-started target`；`Choose a new target manually` | 手动入口的应用提供文本纳入语言覆盖；仍进入已有统一打开/授权路径，不另建本地化专用打开流程。 |
| macos-menu-bar-app / Bounded persistent recent targets；Clearing history does not stop sessions | `Open more than twenty distinct targets`；`Reopen an older recent target`；`Restart after using previews`；`Clear history with an active preview`；`Restart after clearing history` | 本地化最近区域/动作而不重置、迁移、重排或清除实际记录；帮助或语言选择不改变活动服务，旧 50 项数据不碰。核心行为由 09 交付。 |
| macos-menu-bar-app / Action failures remain understandable and recoverable | `Open a stale recent target`；`Recover from browser opening failure`；`A service exits unexpectedly` | 完整原生错误标签与可执行指导本地化，保留真实目标、阶段、原因与恢复入口；不新增第二份状态/报告或自动恢复。 |
| [macos-preview-sessions](../specs/macos-preview-sessions/spec.md) / Verified startup and actual preview URL；Browser opening is separate from service lifetime | `A nondefault port is used`；`Open a file with spaces and non-ASCII characters`；`Startup ends without valid readiness information`；`Browser opening fails after service startup`；`Close the browser preview` | 本地化 native 启动/浏览器失败说明及操作，不改 ready/URL/诊断协议、不把浏览器失败当服务退出。 |
| macos-preview-sessions / Unavailable paths are not empty directories；Visible hot-reload degradation | `A target is moved or deleted`；`A mounted volume becomes unavailable`；`Watch coverage is incomplete`；`Use a network-volume preview` | 本地化 native 状态标签/说明，保留实际 target/reload reason 和 degraded/unavailable 区别；不承诺 watch、不伪造网络卷证据。 |
| [macos-app-packaging](../specs/macos-app-packaging/spec.md) / Candidate DMG and Services-only Finder integration | `Produce an installable candidate`；`Use the new Finder integration` | 完成 native 双语/ServicesMenu/引导资源，供 12 候选集成；实际开发 App 语言证明不替代 13 候选、架构或最低系统验收。 |

跨票完整覆盖仍见 [coverage-plan.md](../coverage-plan.md)，含全部 26 个 checkbox、30 个 requirement、68 个 scenario；上述消费者映射只说明本票贡献，不扩大原规格或提前完成其他任务。

## Implementation handoff and boundaries

### First-use and revisitable guidance

- 用简短原生引导说明：Finder 选择目录/Markdown → Services 命令 → 默认浏览器预览；未出现时在系统的键盘快捷键/Services 设置检查服务是否启用，并说明可用的手动打开入口。服务可用不要求登录启动，App 启动、provider 注册或源码声明存在都不证明系统菜单可见。
- 实际首次启动提供引导；宿主运行后有明确可再次查看的入口，包括用户关闭该 Services 的情况。保持紧凑面板，不增加独立管理主窗口、设置中心、强制 onboarding 完成门槛或语言/高级参数选择器。
- 如选择提供 `NSUpdateDynamicServices()`，仅请求系统重扫声明；它是批准设计允许的手段，不是必做的新按钮或菜单出现保证。不得调用服务启用偏好 API、写系统服务偏好、清索引或自动启用用户关闭的服务。
- 权限说明沿用 08 的实际条件式指导：普通目标不泛化提示；失败保留原目标/原因，可检查 Files and Folders 等适用系统授权，但不把所有访问错误诊断成 TCC，不要求扩大权限。网络共享访问/监视仍有 08 未证边界。
- 保持 App 根依赖就绪后、面板展示前注册唯一 provider；引导不能成为 Services 请求的前置确认或改走新路径。冷启动已有处理能力、共享错误报告、关闭面板不停止服务等约束必须保留；不借本票增加通用模态队列或重试框架。
- 若首次判断确需持久状态，只保存必要的首次使用状态，使用与最近记录/旧历史独立的最小边界；不预定 key、重置/升级策略或额外行为保证，不清整个 UserDefaults 域，不修改新近记录或旧 `go-grip-history`。真实首次状态验证使用获准的专用测试状态，不重置用户数据。

### Complete native language surface

- 工程按批准设计提供 `en` / `zh-Hans` 资源和英文 development region；资源必须进入实际 App 的构建路径。当前盘点无现成帮助/完整本地化，运行时插值错误也不是 SwiftUI 静态字面量自动覆盖；实施前重查完成后的 09 与唯一 Xcode 工程，复用现有结构，不引入新的语言服务框架。
- 遍历并覆盖完整原生消费者：运行与最近目标、所有 phase/target/reload 的人类可读标签/说明、路径区分、空状态、打开/重开/复制/停止/全部停止/清历史/退出/帮助；原生数量确认、应用提供的手动选择提示、provider 输入失败、目标准备/启动/浏览器/终止失败、根级汇总与可查看报告、条件式权限指导以及首次/重看引导。11 将来新增的登录 UI 应复用该语言方案，10 不提前交付登录功能。
- 语言分界是**原生标签和行动指导**与**真实技术诊断**：Go fatal code/message、target/reload state/reason、协议/HTTP/进程底层 detail、系统 `localizedDescription`、实际路径/URL 仍保留其原始信息；用对应语言组织说明和操作，不机械翻译机器枚举/诊断，不丢失原因，不替换成泛化成功/监视状态。原生英文硬编码操作指导不允许作为“原始诊断”漏译。
- Finder 是独立的系统消费表面：保持唯一 `NSServices`、`NSMessage=openWithGoGrip`、`NSPortName=GoGrip` 与 `NSMenuItem.default` 精确值 **`Open with GoGrip`**；在 **`en.lproj/ServicesMenu.strings`** 和 **`zh-Hans.lproj/ServicesMenu.strings`** 以该精确默认键提供名称。`Localizable.strings`、plist 拷贝/字符串存在或 provider 单测都不能证明 Finder 实际本地化。
- macOS 13+ AppKit/SwiftUI 兼容和菜单栏-only 宿主边界不变。语言改变只影响展示，不改变 normalized identity、原始路径、真实 URL、最近成功/recency、运行状态或服务所有权；中文/英文及真实长路径下操作/指导须可读可达，不新增像素尺寸、layout snapshot 或新的布局规格。

### Non-goals

不做 09 的存储/面板功能重写、旧数据迁移或全域清理；不新增语言/端口/认证/watch 等设置、通知权限、TCC 管理层、书签/FDA/helper/detach/自动恢复/重试；不翻译 Go CLI、renderer/browser 页面或机器协议；不重复已完成的批次/owner/协议矩阵；不处理 11 登录项、12 universal/DMG/CI 或 13 候选及支持平台证明；不关闭 08/09、不勾选其他任务、不提交、同步、归档或正式分发。

## Scoped TDD and real verification

实施前另行批准最小行为 seam 与具体 native 操作，不因本票发布自动开始红绿循环、编译、安装、语言/Services/TCC 修改或索引重扫。

- 仅当本票新增确实可测的首次/再次查看状态转移时，针对用户可见的引导可达行为及与会话/最近记录分离的必要边界做 scoped TDD；使用隔离可销毁域/专用状态与确定性事件，不为测试发明存储抽象或升级规则。纯资源/文案/UI 路径以真实 native proof 为主，不新增源码文案、字符串存在、翻译字典/默认值、plist/资源拷贝、wiring、mock echo 或布局快照测试。
- 保留既有有价值的行为回归；受本地化影响的纯措辞/实现测试删除，不重新固定中文/英文文案。不要用第二套完整启动、批次、访问原因或退出矩阵换取覆盖；已有模型/协调器行为 seam 只能测消费者可见事实，不测文本转发。
- 运行受影响的现有 Swift/Go 检查及实际 App 构建，但单测/编译不替代下表。证据记录具体开发 App、bundle/内置 Go、系统与 Finder 的实际语言和观察表面；受控故障只补难触发 UI，必须标注且还原，不冒充真实系统权限、网络卷或候选证明。

| Actual native scenario | 必须观察的证据 |
|---|---|
| 首次启动 → Finder/Services → 浏览器 | 获准的首次状态下有简短可操作说明；provider 不依赖先打开 popover/完成引导；实际 Finder 命令仍进入同一内置 Go pipeline，记录真实 URL/内容和 owned 会话，不启动 Terminal、不要求登录项。 |
| 关闭引导后再次查看；用户关闭服务 | 运行中明确入口可再次看到说明；服务关闭/不可见时有可操作的系统设置检查与权限说明，App 不自动启用或改变偏好；用户按说明自行检查/恢复后可重新找到实际入口。获准的 Services 偏好变更记录并恢复；如请求重扫，不把请求返回当菜单已出现。 |
| 简体中文、英文、不支持语言 → 完整面板/帮助 | 三种实际环境逐一观察 09 完整运行/最近面板、空状态与明确动作、首次/再次查看帮助、数量确认及失败/指导；简中/英文随系统、不支持语言英文回退，路径/URL/诊断仍真实可见，长路径不遮蔽操作。 |
| 同三种环境 → Finder 服务实际名称与调用 | 在 Finder 的真实 Services 菜单观察对应名称及实际调用；App 启动参数或 per-app 语言仅证明 App，不能代替 Finder 语言/菜单证明。需要系统/Finder 语言、安装、偏好或重扫操作时先获具体批准，记录原状态并恢复；没有该环境则语言/菜单 acceptance 保持未完成。 |
| 错误和可恢复动作 | 实际访问/陈旧目标失败、批次失败汇总、浏览器失败及运行期退出中观察本地化标签/行动指导和保留的原始原因；只有一次既有根报告，可重看；成功无额外提示，浏览器失败不撤销真实运行，unavailable/degraded 不伪装为空目录/正常 watch。既有行为证明可复用，新增表面须真正展示，不能只改 reason 字符串。 |
| 活动服务中查看/关闭帮助与语言展示；真实退出 | 同一实际服务/URL/内容仍可管理，帮助不停止或篡改历史；既有复制/重开/停止/退出仍作用于正确会话，实际 owned PID/监听收尾且独立 CLI 不被停止。只复查本票可能影响的原生路径，不重演完整已关闭协议/强杀矩阵。 |

缺乏真实 Finder/系统语言/首次/用户关闭服务环境时，完成所有可达工作并明确缺口；不得降格为 strings、selector 调用或静态检查而宣称 5.1/5.2 完成。08 网络共享缺口继续单独记账；开发 App 证明不提升为候选或跨架构验收。

## Acceptance

- [ ] 重新读取当前 approved artifacts、实际完成的 09 和既有 01–08 贡献；09 的 4.1–4.3 完整交付/独立审阅前置已满足，另获本票范围、必要 TDD seams、native 操作及具体数据/环境许可；不覆盖另一会话的进行中工作，不将发布等同实施授权。
- [ ] 实际首次启动提供简短 Finder/Services 使用说明，宿主内有明确可再次查看入口；按说明可以定位 Services 命令和菜单不出现时的系统设置检查，包含适用权限指导，不增加管理主窗口、强制 onboarding、语言或高级设置。
- [ ] 实际用户关闭服务/菜单不可见场景中说明仍可查看且可操作，按说明自行恢复后找到服务；App 不自动启用或修改系统服务偏好，不把 App 启动、声明存在、provider 注册或可选重扫当作菜单可见性保证，测试环境变更已获准并还原。
- [ ] 在实际简中、英文及不支持语言环境观察完整 09 native 面板、状态/原因/全部操作、数量确认、应用提供的手动选择文本、首次/重看帮助及错误行动指导；简中/英文跟随系统、其他语言英文回退，不只翻译旧简化 popover，不提供应用内语言选择。
- [ ] `en`/`zh-Hans` 资源与英文 development region 进入实际构建；Finder Services 两份 `ServicesMenu.strings` 使用精确默认键 `Open with GoGrip`，保持唯一声明/selector/port；上述三种 Finder 实际语言环境中观察服务名称和真实调用，而非以 App locale、plist/strings/资源拷贝测试替代。
- [ ] 所有原生人类可读标签、错误包装和可执行指导按对应语言呈现，含条件式访问指导；真实 path/URL、Go code/message、target/reload reason 和系统/协议/进程底层诊断保留，不把英文应用操作说明误当原始诊断漏译，不扩张权限或伪造 running/watch。
- [ ] 保持已批准根级一次失败汇总、可查看报告与既有失败横幅决策，成功不添成功提示；浏览器失败保留运行事实/恢复入口，陈旧最近目标及异常退出的实际原因/操作可理解，不另建报告或状态事实源、不自动恢复。
- [ ] 首次/重看帮助、语言资源和展示不改变当前真实服务、URL、normalized identity、新最近记录或 recency；必要首次状态独立且最小，不读/迁移/写/删旧 50 项、不清整个 UserDefaults 域，不把清历史变成停服务。
- [ ] 仅为获准且真正新增的消费行为/状态边界保留必要 scoped TDD；永久测试确定性、隔离、全套安全；删除受影响的纯措辞/实现断言而不重钉译文，不新增字符串存在、源码、默认值、plist 拷贝、wiring、mock echo、尺寸/布局或重复矩阵测试。
- [ ] 完成上表真实 native/Finder/browser 观察并记录具体 App/bundle、系统/Finder 语言、URL/内容、会话与 owned 收尾证据；长路径/中文操作可达，首次帮助不阻断冷启动 provider，受控故障明确标注并还原，测试/编译不代替实际表面。
- [ ] 受影响的现有 Swift/Go 行为检查及实际 App 构建通过，工程/调用点使用同一资源方案，无本票切换产生的旧文案路径或多余框架；更新 README 首次使用、菜单故障/权限步骤与语言支持，架构/交接说明准确记录实现和未证边界，清理本票临时验证资源。
- [ ] 保留 08 open/3.3 与真实网络共享补证边界，不宣称候选、Intel/macOS 13 支持矩阵、登录项或正式分发已完成；12/13 可消费本票资源与记录，但各自实际集成证明不被提前完成。
- [ ] 全部本票验收及独立 Standards/Spec 审阅通过、无未决范围后，才在唯一账本勾选 5.1、5.2；不更新其他任务、关闭其他票、发布 11、提交、同步或归档。

## Whole-change coverage checkpoint

[coverage-plan.md](../coverage-plan.md) 保留全部 26 个 checkbox、30 个 requirement、68 个 scenario 及完整未来切片/真正 blocker。本轮仅新增本票并刷新该地图：09 按用户告知标为另会话实施中、不宣称完成；10 已发布 open，仍 blocked by 09；11–13 未发布。源 artifacts、已有票据及任务账本不因发布发生变更。

## Closure record

- 关闭：2026-10-06，状态 `done`；13 项验收全部按实际证据通过。实施批准：用户显式批准本票范围（5.1/5.2 一体）、S1 单 scoped TDD seam、native smoke 与具体环境操作（/Applications 替换并备份、Services 启用开关切换、首次引导键目标性切换、Finder/pbs 重启、全局 AppleLanguages 轻量切换及必要时升级）。
- 变更面：新增 `macos/GoGrip/Utilities/FirstUseGuidanceStore.swift`、`macos/GoGripTests/FirstUseGuidanceStoreTests.swift`、`macos/GoGrip/en.lproj/{Localizable,ServicesMenu}.strings`、`macos/GoGrip/zh-Hans.lproj/{Localizable,ServicesMenu}.strings`；修改 `macos/GoGrip/AppDelegate.swift`（首启展示、Help 闭包、setup helper 主 actor 标注）、`Services/AppAlertPresenter.swift`（guidance 单槽 + 提示文案本地化）、`Views/PopoverView.swift`（Help 入口、phase/fallback 本地化）、`PreviewAppModel.swift`、`Services/PreviewSessionCoordinator.swift`、`Services/TargetPreparation.swift`、`FinderServiceProvider.swift`（包装句本地化）、`macos/GoGrip.xcodeproj/project.pbxproj`（Resources phase、两个 PBXVariantGroup、zh-Hans knownRegion、新源/测试文件）、`README.md`、`docs/ARCHITECTURE.md`、`docs/HANDOVER.md`。
- Scoped TDD 证据（批准 seam S1）：red 为 `cannot find 'FirstUseGuidanceStore' in scope` 编译失败；实现后 2 项隔离域回归（首次展示一次且重建实例保持；与最近键/旧 `go-grip-history` 哨兵分离）转绿；未新增字符串存在、资源拷贝、默认值、plist、wiring、mock echo、尺寸/布局或重复矩阵测试；既有文本断言在无资源 test bundle 中确定性解析英文键且为行为断言，未重钉译文。
- 自动检查（最终树）：`make macos-test` **92 tests / 0 failures**（87 → 92）；`make macos` Release `BUILD SUCCEEDED` + `codesign --verify --deep --strict` 通过；产物含 `en.lproj`/`zh-Hans.lproj` × `Localizable.strings`/`ServicesMenu.strings` 四文件、`CFBundleDevelopmentRegion=en`、`knownRegions` 含 `zh-Hans`；53 个键两表对齐、`plutil -lint` 通过。Go 零改动，未重跑 Go 测试/errcheck。
- 真实 smoke（2026-10-06，本机 macOS 27.0.1，最终开发构建 `3949aca2…` 安装于 `/Applications`；系统/Finder 语言经全局 AppleLanguages 在 en-US→zh-Hans→ja 间切换并已还原 en-US；Services 启用开关、首次键与最近键均记录并还原，旧 `go-grip-history` 未动；详见 `docs/ARCHITECTURE.md` 10 记录）：en 首次冷启动（宿主由 launchd 启动，PPID=1）自动显示引导，同时 Finder 服务已建立 owned child/回环端口/HTTP 内容——引导不阻断 provider/批次；第二次启动不自动显示，面板 Help 可重看且运行会话 PID/URL 不变；zh-Hans 面板/6 目标数量确认/批次汇总/陈旧目标/浏览器失败/运行期退出原因均为中文包装并保留原始 path/URL/reason；真实 Finder Services 菜单在 zh 显示“用 GoGrip 打开”、不支持语言（ja）回退 “Open with GoGrip”，两者均经真实调用进入同一内置 Go pipeline；用户经 System Settings 关闭服务后引导仍可查看、App 未自动启用，按说明恢复后菜单重现；受控浏览器拒绝构建（标注、源码 diff 还原一致、最终重装）观察到 zh 浏览器失败提示且会话/URL 保持；面板 Quit 释放 owned PID/端口，`Stop All` 不影响独立 CLI（6419 继续 HTTP 200）。
- 独立只读审阅：Standards 首轮 1 阻塞（引导未说明“预览在默认浏览器打开”）→ 修复（en/zh + README）并在实际 en/zh/Help 表面复核 → 增量复审 0/0；Spec 首轮 0 阻塞、3 advisory（S-A1 证据时序、S-A2 设置路径缺 Keyboard 父级、S-A3 自动引导激活）+ 1 scope（S-D1 自动展示触发范围）；S-A1/S-A2 已修并复审 CLOSED，S-D1 维持解释 A（规格 scenario 字面满足，无需修订 artifact），S-A3 记为任务 13 候选 smoke 观察项（若观察到浏览器焦点被抢，自动实例改为不激活展示，Help 不变）。最终两轴均无未决阻塞或范围裁决。
- 保留的 advisory/限制：S-A3（自动引导激活，任务 13 观察项）；登录项（5.3）、Intel/macOS 13、已挂载网络卷（08 保持 open、3.3 未勾选）与候选包（12/13）未验证；开发构建证明不提升为候选或跨环境验收；ad-hoc 身份跨构建可能需重新授权（既有实测限制）。
- tasks.md：本票完整交付并验证 **5.1、5.2**，两项已勾选；其他任务、票据与 checkbox 未变。
- 未覆盖：未提交/推送、未同步主规格、未归档、未发布/实施 11；08 保持 open；覆盖映射的发布时快照留待下一次发布轮更新（其自身说明与 05–07 先例）。
- 环境与用户数据还原：`/Applications` 为本票最终开发构建 `3949aca2…`；替换前构建 `7a2c78fa…` 备份于 `~/Desktop/GoGrip-backups/GoGrip-07-backup.app`；全局 `AppleLanguages=en-US`；Services 无 GoGrip 覆盖项；首次/最近键删除（会话前即为缺失）；临时目标与验证资源已清理。
