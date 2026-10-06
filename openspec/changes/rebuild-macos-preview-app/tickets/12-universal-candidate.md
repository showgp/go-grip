# 12 — 自包含 Universal 候选 App/DMG 与候选 CI 交付

Status: open
Blocked by: 10-localized-service-guidance, 11-opt-in-login-start
Covers OpenSpec tasks: 6.1, 6.2, 6.3, 6.4
Behavior source: ../specs/macos-app-packaging/spec.md, ../specs/macos-finder-service/spec.md, ../specs/macos-menu-bar-app/spec.md, ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

从唯一的 `macos/GoGrip.xcodeproj` 及现有 Make/CI 入口，实际生成**宿主与内置 Go 都包含 arm64/x86_64、支持 macOS 13+ 的自包含候选 App**，统一显式 archive、ad-hoc 签名和由该 archive App 生成的候选 DMG。DMG 提供 Applications 安装入口；候选 App 脱离源码与开发 PATH 后仍通过已落地的 native 入口打开目标到默认浏览器并能停止真实 owned 服务。相关 macOS CI 交付 App/DMG 的候选 artifact，而不是自动上传正式 Release；随候选提供可实际执行的构建、使用、升级与回退说明。

本票完整覆盖 **6.1、6.2、6.3、6.4**。不承担 10 的引导/本地化实现、11 的登录项实现、08 的网络共享补证，也不提前完成 13 的 **7.1–7.3** 候选级跨组件/平台矩阵。行为权威为本 change 的批准 specs；任务进度仅由 [tasks.md](../tasks.md) 记录，票内清单不是第二本账本。

**发布核对时状态**：01–07、09 已发布 done；09 的完整面板/最近行为及 07 重开的非模态共享提示修正已独立确认。08 整票仍 open、网络共享待补证、3.3 未勾。10（5.1、5.2）和 11（5.3）均为 open，三项未勾；用户已确认可以在本轮发布 12，这不是 10/11 实施完成或验收通过的证据。当前账本为 **15/26**，6.1–6.4 全部未勾，13 未发布。

**真正实施前置**：依据 tasks 第 6 组“依赖第 5 组”，10 与 11 的完整交付和独立 Standards/Spec 审阅须先完成，资源与宿主接口以实施当时的最终版本为准。05 的开发包、当前代码中存在的资源引用、11 的发布确认均不代替这些前置。08 的既有网络缺口保持原归属和关闭条件，本票不把打包通过解释为完整权限 gate 已通过；若实际候选证据否定内置工具路线，按批准设计停止相关集成并与用户重评，不自行增加 helper、detach 或 FDA 兜底。

**发布不是实施许可**。本轮仅新增本票并刷新 [完整覆盖地图](../coverage-plan.md)，不修改既有票、proposal/specs/design/tasks、应用/测试/构建代码或用户环境。实施前另行批准单票范围、相关现有行为套件、实际候选 smoke 及具体安装/数据/系统操作；不由本轮发布推导推送、打 tag、触发远端工作流、发布 Release、同步主规格或归档的许可。

## Source mapping

- [proposal.md](../proposal.md)：内置工具的 macOS 13+ 双架构宿主；完整功能及可用候选 App/DMG；Services-only 替换旧示例；候选与正式签名/公证/发布分阶段。
- [design.md](../design.md)：Decision 10 的单一 Xcode 工程、显式架构/部署与 `CGO_ENABLED=0`、确定工具路径、先工具后宿主 ad-hoc 签名、显式 archive/DMG、候选 CI artifact；Decision 9 的资源、真实登录状态与不持久化服务；Verification Strategy 的实际候选运行及静态检查边界；Migration Plan 的 CLI release 隔离、升级/回退和 08 缺口。
- [tasks.md](../tasks.md)：以物理 checkbox **6.1、6.2、6.3、6.4** 及其完整验收描述为准，不使用 apply JSON 顺序 ID。第 7 组只验证已落地组件，不接收本票遗漏的构建、测试或使用文档。

| 行为源 / Requirement | 对应 scenarios | 本票贡献与不越界的部分 |
|---|---|---|
| [macos-app-packaging](../specs/macos-app-packaging/spec.md) / `Universal self-contained application` | `Use the application without development tools`；`Use the same candidate on supported architectures` | 6.1/6.2 生成完整双架构 App，检查宿主与工具各 slice 的架构/部署信息，并实际运行本机候选。Intel 与 macOS 13 的实际运行证据仍由 13/7.3 补齐；交叉编译或静态检查不完成整个 scenario。 |
| 同上 / `Candidate DMG and Services-only Finder integration` | `Produce an installable candidate`；`Use the new Finder integration` | 6.2 从显式 archive App 生成含 Applications 入口的实际 DMG；消费 10/11 的完整功能及资源，不包含旧 Finder Sync/转发层；6.4 的安装/使用指导和本票候选 smoke 不替代 13 的完整 Finder 矩阵。 |
| 同上 / `Functional acceptance exercises actual user paths` | `Verify a real Finder-to-browser launch`；`Verify reuse and explicit stopping`；`Verify abnormal host termination` | 本票为 13 提供同一可运行候选，并按 6.1/6.2/6.4 实际验证包内工具、正确页面及停止。整个 requirement、完整复用和异常退出矩阵归 13；旧开发包、构建或自动测试不能替代候选证据。 |
| 同上 / `Explicitly deferred formal distribution` | `Present an unsigned candidate for functional review`；`Complete this implementation stage` | 6.3/6.4 标明候选身份、实际信任限制及尚未完成的正式签名、公证、干净安装与公开分发。ad-hoc 不等于 Developer ID 签名；本票完成不关闭 08 或 13，不授权发布。 |
| [macos-finder-service](../specs/macos-finder-service/spec.md) / `Finder Services entry`、`First-use service guidance` | `Open a selected folder from Finder`；`Invoke the service while the host is not running`；`Read first-use instructions`；`Revisit guidance for a disabled service` | 将已落地 provider、首次/再次查看的指导与 Services 本地化放入实际候选，按文档检查系统启用状态；不重做注册入口、不自动启用服务，不以声明/重扫或资源存在冒充真实 Finder 进入正确候选。 |
| [macos-menu-bar-app](../specs/macos-menu-bar-app/spec.md) / `Persistent native menu bar host`、`Opt-in launch at login`、`System-language localization` | `Open and dismiss the panel`；`First use without login startup`；`Change the login-start preference`；`Use Simplified Chinese`；`Use English or an unsupported language` | 候选消费 09–11 的完整原生功能和统一资源，保持 LSUIElement、关闭 UI 不停止、真实系统登录状态、简中/英文/英文回退。打包不增加自动注册或语言设置；本票不重复实现或代签 10/11 的验收。 |
| [macos-preview-sessions](../specs/macos-preview-sessions/spec.md) / `Verified startup and actual preview URL`、`Explicit stopping confirms service termination`、`Independent CLI operation remains available` | `A nondefault port is used`；`Open a file with spaces and non-ASCII characters`；`Startup ends without valid readiness information`；`Stop one of several sessions`；`Stop all application sessions`；`Launch the renderer without the application` | 实际候选使用包内真实 Go 和已批准的启动/停止契约，不猜 URL 或回退固定端口；正常停止/退出观察 owned 资源释放。构建和 App CI 切换不改独立 CLI 的模式、网络政策或现有 release 资产流程；完整会话矩阵仍归 13。 |

## Implementation handoff and boundaries

### 1. Current build surface and genuine blockers

实施前重读 CLI status/apply 的全部当前上下文及 01–11，确认 10/11 已完成且无未决 scope decision；对其他会话的代码变更重新盘点，不按本票发布快照覆盖工作。

发布时直接核对的现状如下；这些是实现入口，不是完成证明：

- [Makefile](../../../../Makefile) 的 `macos`/`macos-run` 使用同一工程和 DerivedData 开发 build，`macos-test` 驱动已有真实 Go 的 Swift 行为套件；当前尚无统一候选 archive/DMG Make 入口。保留开发/测试用途，让本地与 CI 复用本票确定的候选构建/归档流程，不引入第二套工程定义。
- [project.pbxproj](../../../../macos/GoGrip.xcodeproj/project.pbxproj) 保留宿主与测试目标，App target 先执行 Build go-grip，再编译 Sources/Resources；宿主 bundle ID 为 `com.showgp.GoGrip`，已有 ad-hoc identity 和 13.0 deployment 设置。当前未显式声明 `ARCHS = arm64 x86_64`，Debug 项目层 `ONLY_ACTIVE_ARCH=YES`；不能据默认值或旧构建记录宣称本次候选两部分均 universal。当前工程已有 `Localizable.strings`/`ServicesMenu.strings` 的资源引用，但 10 仍 open；引用存在不是三种真实语言环境已验收的证据。
- [build-go-grip.sh](../../../../macos/Scripts/build-go-grip.sh) 当前构建 darwin arm64/amd64，经 lipo 合并到 `Contents/MacOS/go-grip`，再以独立 identifier `com.showgp.GoGrip.go-grip` ad-hoc 签名。现有两个 `go build` 未显式设置 `CGO_ENABLED=0`，本票按批准要求补齐；不把开发构建时的 PATH 配置移成运行期查找逻辑。
- [GoGrip.xcscheme](../../../../macos/GoGrip.xcodeproj/xcshareddata/xcschemes/GoGrip.xcscheme) 的 ArchiveAction 使用 Release，TestAction 使用 Debug。候选 archive 必须明确输出，不以 `make macos` 成功、测试临时 App 或旧 DerivedData App 代替。
- [build.yml](../../../../.github/workflows/build.yml) 的 PR macOS job 当前只 build 且覆盖 `CODE_SIGNING_ALLOWED=NO`；[release.yml](../../../../.github/workflows/release.yml) 的 App job 当前禁用签名 archive、用 hdiutil 生成 DMG 并经 `softprops/action-gh-release` 自动上传。**本票必须切断 App/DMG 的自动正式 Release 路径**，将相关 macOS 路径更新为候选构建/行为检查及 artifact 交付；CLI release job、其资产和既有 tag 流程保持原样，不把整个 CLI release 流改成候选流程。

### 2. One archive, two universal executables, candidate-only delivery

- 宿主显式 `arm64 x86_64`、`ONLY_ACTIVE_ARCH=NO`、macOS deployment 13.0；Go 两个 darwin 构建显式 `CGO_ENABLED=0`，再合并。核查宿主和工具每个 slice 的实际架构及最低系统信息，确认与 macOS 13 声明兼容；不要求把 Go 工具元数据伪装成宿主的同一个值，也不以宿主的 deployment setting 推断 Go 兼容性。
- 保持唯一运行路径 `Bundle.main.bundleURL/Contents/MacOS/go-grip`。工具先完成双架构合并和独立 identifier 的 ad-hoc 签名，再由 Xcode 签宿主；候选路径不能保留禁用 App 签名的 CI 覆盖，也不能用宿主强制深度重签替代独立工具签名。测试 bundle 的既有无签名设置不是需要改成候选 App 签名的目标。
- 明确 archive 输出路径，并从该次 archive 的 `Products/Applications/GoGrip.app` 取得候选 App；只由该 App 生成含 `/Applications` 安装入口的 DMG。具体 Make 命令、路径和 artifact 名称由实施时的现有约定确定并写入文档，不假称它们已存在，不从旧 `macos/build/Release` 或另一个成功 build 猜测产物。
- 候选包含完整宿主、工具、Services 声明及最终本地化/引导/登录 UI；不嵌入旧扩展，不恢复 URL 转发、Resources/PATH/源码 fallback 或运行期 chmod。需要修改前置组件才能达到这些行为时，回到其所属票处理，不把缺功能的包交给 13 补实现。
- CI 上传可取出并运行的候选 App 和 DMG；App 可用保持 bundle 结构、符号链接和执行权限的归档容器交付，DMG 保持实际安装内容。不要求消费者下载后 chmod 修包，不新增第二套正式分发格式。按实际归档/解包和运行证明交付内容，不用文件存在/拷贝测试证明。
- 本地和相关 macOS CI 使用同等构建、archive、签名检查、打包与已有行为套件命令，工具链沿用当前仓库约定；不引入 XcodeGen、Developer ID/公证 secret 或凭据依赖。远端未运行时只报告本地等价验证，不宣称 artifact 上传已观察；不代用户推送或触发发布。

### 3. Candidate use, cutover and rollback documentation

更新现有 [README.md](../../../../README.md) 与 [docs/ARCHITECTURE.md](../../../../docs/ARCHITECTURE.md)，并仅更新相关旧记录 [macos-app-plan.md](../../../../docs/macos-app-plan.md)、[macos-app-impl.md](../../../../docs/macos-app-impl.md)、[HANDOVER.md](../../../../docs/HANDOVER.md) 的当前状态/替代路径说明，不把历史示例再变成工程来源。

文档至少包括：实际构建和 archive/DMG 命令及产物路径；候选 App/DMG 与 CI artifact 的取得/解包方式；安装和 Finder Services/手动入口、默认浏览器、实际 URL、停止/退出；沿用 10 的服务菜单/权限指导和 11 的真实登录状态说明；候选信任提示与 ad-hoc 的身份限制；已经观察的架构/系统和仍缺的环境。CLI 正式资产说明与 App 候选说明分开，旧计划中的“tag 自动正式 DMG”步骤明确为被替代的历史路径，不用实际推送证明文案。

升级先停止旧宿主的会话并退出旧宿主，避免 LaunchServices 选中另一个相同 bundle ID 安装副本；不按进程名扫杀独立 CLI。新最近 schema 不自动迁移、读取或删除旧 `go-grip-history` 数据，也不恢复旧 PID/端口/服务。回退先停止新宿主的 owned 服务并退出，再由用户选择明确旧版本或独立 CLI；不保留兼容 shim，不把回退解释为旧示例可靠。正式 Developer ID 签名、公证、干净安装标准和公开发布保持后续单独审批，既有正式标准不降低。

### 4. Non-goals

不重做 Go 渲染器、独立 CLI 网络/端口政策或 release 格式；不增加安装器、自动更新、正式发布自动化、签名凭据管理、helper、权限绕过、进程名清扫或资源 fallback；不扩张浏览器/Chrome 后代清理；不添加新的并发构建、重试或通用打包框架保证。08/3.3 和 13/7.1–7.3 的范围保留，不因本票已发布/完成而勾选；不主规格同步、不归档、不提交或推送。

## Scoped verification and actual candidate smoke

本票主要是构建/交付 cutover，没有批准新增运行期产品行为。**默认不新增永久测试**；执行现有相关 Go/Swift 行为套件并修复本票改动破坏的已批准契约。若需要新增行为或 seam，先确认对应批准 requirement 与实施 scope，再按 scoped TDD；不得新增 ARCHS/路径/字符串/资源存在、YAML 源文本、拷贝或 mock 转发测试。旧 incidental/wiring 断言不重新固化。临时检查/脚本只辅助实际产物验证，完成后移除。

| 检查 / 场景 | 必须观察的结果与证据 | 不能替代 / 边界 |
|---|---|---|
| 实际候选构建和 archive | 按最终本地/CI 同等命令成功生成明确 archive；记录工程、配置、构建命令、工具链、产物路径及本次候选标识；相关现有 Go/Swift 行为套件通过。 | 不以过去 tests 数量、开发 App、仅 Swift 编译或流水线文本代替；不是本轮发布要执行的构建。 |
| 两个实际 Mach-O 的架构和部署 | 用 `lipo`/Mach-O 信息检查宿主与包内 Go 均含 arm64/x86_64，逐 slice 记录最低系统信息并与 macOS 13 声明核对；实际 Go 构建记录确认 `CGO_ENABLED=0`。 | 只有工具 universal、只看 settings、只在本机启动成功均不够。静态结果不是 Intel/macOS 13 实际运行证据。 |
| 独立工具和宿主签名 | 对 archive App 中工具/宿主实际检查 ad-hoc、各自 identifier 及 bundle 签名结构；签名验证通过，打包/取出后仍有效。 | 不靠完全禁签或运行时重签/chmod补救；`codesign` 成功不证明 TCC、Gatekeeper、Developer ID 或公证通过。 |
| 实际 DMG 和候选 App artifact | 由该 archive App 生成 DMG，挂载后有完整 App 和 Applications 安装入口；候选 App 归档交付时实际解包，保留可执行内容和签名。记录来源/候选标识，运行验证使用从交付包取出的 App。 | 不运行测试临时 App 或另一轮 DerivedData 输出；不只是断言文件、symlink 或资源存在；不伪称远端上传/下载已发生。 |
| 脱离源码和开发 PATH 的真实预览 | 将取出的 App 放到源码/构建目录之外，使用不依赖开发 Go 或额外 CLI 的受控运行环境；经候选面板手动打开目录及中文/空格 Markdown 单文件，观察包内真实 Go executable、owned PID、实际回环 URL、默认浏览器的对应内容。 | 不需要卸载用户 Go 或移动/删除源码来制造环境；环境中仍有开发 Go 时，不宣称已观察“机器未安装 Go”。HTTP/进程结果不能替代默认浏览器实际页面。 |
| 按使用文档的候选 Finder 路径 | 在获准的安装/入口准备后，确认系统实际启动的是该候选路径而非旧副本；宿主未运行且不以登录启动为前提，从 Finder Services 打开专用目标到正确浏览器内容。候选面板、首次/再次查看的指导及语言/登录状态表面消费最终 10/11 版本。 | 不用 provider 调用或声明/重扫代替 Finder；不自动修改服务偏好、登录项或系统语言。10/11 的完整原生检查和 13 的候选矩阵仍按各自范围验收。 |
| 文档中的停止与正常退出 | 对上述真实预览按文档停止，观察 owned PID/监听消失；关闭 UI 不停止，正常退出完成 owned 收尾，独立 CLI 不受影响。按批准准备保留明确可用回退版本/CLI，核实升级/回退步骤不要求删除旧数据或自动恢复服务。 | 不仅删除面板行，不按名称扫杀；本票正常收尾 smoke 不替代 13 的 ready 前/运行中异常退出矩阵，不擅自安装/启动旧示例。 |
| macOS CI 候选交付与 CLI 隔离 | 实际执行同等 archive/打包/行为检查命令并检查候选内容；App/DMG 输出交给 candidate artifact 路径，旧 App 正式 Release 上传路径已切断，CLI 现有 release job/资产流程未改。记录实际远端 run 或本地等价验证的区别。 | 不用 YAML 测试、简单转发或复制断言代替构建；不要求或执行 tag/push/Release，不虚构 Actions 上传证据。 |
| 文档/证据边界 | 按 README 的实际候选命令和使用步骤完成打开/停止；列明当前机器的真实运行与每个静态检查结果，保留 08 网络卷与 13 支持环境/完整集成缺口。 | 不能宣称旧示例可靠、所有 CLI 已限本机、Intel/macOS 13 已运行、TCC 跨构建稳定或候选正式发布就绪。 |

具体 native 环境操作另行批准：替换 `/Applications` 或其他已有安装前，核对路径、运行宿主和 owned 会话，保存已验证回退副本/状态；不删除用户 App/数据，不强制重启 Finder，不改全局语言、TCC 或 Services 偏好，不操作用户登录注册。挂载本次候选 DMG、复制安装、处理信任提示及任何用户设置操作均在批准的 smoke 计划内说明；只卸载本次自己的挂载，不触碰用户卷或其他应用。ad-hoc 可能重新触发授权，按真实系统反馈记录，不自动继承旧构建的授权成功结论。

12 的关闭要求是本票全部产物、行为套件、真实候选 smoke、使用文档和独立 Standards/Spec 审阅完成；无法观察的本票必要场景保留验收缺口，不能把它转交 13 后关闭。**实际 Apple Silicon/Intel、macOS 13+ 支持矩阵属于 13/7.3**：缺所需环境时 13 保持阻塞，不以交叉编译冒充完成；这也不把 12 的静态架构检查和当前机器 smoke 改成整个变更已验收。

## Acceptance

- [ ] 10 的 5.1/5.2 与 11 的 5.3 完整交付并独立 Standards/Spec 确认，无未决 scope decision；12 的范围、现有套件、候选 smoke 和具体环境操作已获另行批准，最终资源/接口重新盘点且未覆盖其他会话工作。
- [ ] 唯一 Xcode 工程实际构建 universal 宿主，显式 arm64/x86_64、ONLY_ACTIVE_ARCH=NO、macOS deployment 13.0；darwin Go 双架构构建显式 CGO_ENABLED=0，并合并在唯一 Contents/MacOS/go-grip 路径。
- [ ] 对 archive 中的宿主和 Go 每个 slice 记录实际 Mach-O 架构/最低系统信息，与 macOS 13 声明兼容；本机真实运行通过，未以静态检查或交叉编译宣称 Intel/macOS 13 实际运行已验收。
- [ ] 先独立 identifier ad-hoc 签工具、再 Xcode 签宿主；archive 与交付包取出的 App 签名结构检查通过，不依赖 Developer ID/公证凭据，不在运行时 chmod/重签。
- [ ] 明确且统一 archive 输出，由该 archive App 生成实际含完整 App 与 Applications 安装入口的候选 DMG；候选 App artifact 的归档/取出保留包结构、执行权限与签名，不猜测旧 build/Release 或运行另一轮开发 App。
- [ ] 取出的候选 App 脱离源码/构建目录及开发 PATH，经真实 native 手动入口打开目录、中文/空格 Markdown 文件，包内实际 Go、正确回环 URL 与默认浏览器对应内容得到配对证据，无 CLI/Resources/源码 fallback。
- [ ] 按文档在获准环境验证该候选的 Finder Services 冷启动到正确页面，不把旧副本、声明或重扫当证明；候选消费最终引导、完整语言资源与真实登录状态，无旧扩展/专用转发，不自动修改系统服务或登录偏好。
- [ ] 从候选按文档实际停止/正常退出，owned PID 与端口释放；关闭 UI 不停止，其他 owned/独立 CLI 按既有契约隔离，无进程名清扫、用户数据删除或自动服务恢复，回退入口明确且受批准边界约束。
- [ ] 相关 macOS CI 与本地同等构建/归档/签名/打包/已有行为检查命令实际执行通过，App/DMG 仅走候选 artifact，App 自动正式 Release 上传已切断；CLI release job/资产/tag 流程不变，本地等价验证与实际远端上传证据明确区分，不代用户推送或发布。
- [ ] README/架构与相关旧计划状态说明随候选更新，实际按候选构建/使用步骤完成打开和停止；升级/回退、旧数据保留、新历史不迁移、重复 bundle 安装风险、信任提示与 ad-hoc 限制准确，候选与 CLI 正式资产说明分开。
- [ ] 相关现有 Go/Swift 行为套件通过；不新增文字/默认值/文件存在/资源拷贝/YAML/纯转发测试，不重新固化旧 incidental 断言；临时检查或 smoke 脚本移除，最终候选不含受控 fake 或开发替代工具。
- [ ] 全部本票必要验收、真实证据及独立 Standards/Spec 审阅完成后才关闭 12 并更新 6.1–6.4；08/3.3 网络共享和 13/7.1–7.3 候选矩阵缺口保留，不宣称整变更完成、正式分发就绪或已获同步/归档/发布许可。

## Whole-change coverage checkpoint

[完整覆盖地图](../coverage-plan.md) 逐项保留全部 **26 个 OpenSpec checkbox、30 个 requirement、68 个 scenario**；每项区分已发布贡献和未来规划，不因当前已发布票结束而提前勾选仍有其他贡献者的任务。

本轮只新增 **12-universal-candidate**，完整映射 **6.1–6.4**，状态 open、待审阅，10/11 实施前置尚未满足。01–07、09 done；08 open/网络共享待补证；10/11 open，发布确认不代表实现确认；13-candidate-functional-acceptance 仍为未发布规划。当前账本 **15/26** 不变，批准 artifacts、既有票及全部任务勾选不在本轮变更范围内。
