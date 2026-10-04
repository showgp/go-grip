# 03 — Foundation owned 进程启动与真实所有权验证

Status: done
Blocked by: 01-managed-preview-lifecycle, 02-managed-status-and-cli
Covers OpenSpec tasks: 2.1, 2.5; 2.4 (process-layer contribution only)
Behavior source: ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

通过生产 Foundation 进程适配器直接启动真实 managed Go，连续消费机器反馈，在启动期限内确认实际回环 URL 可用，并向调用方提供已验证的启动结果、结构化事件和真实退出结果；失败启动及主动关闭所有权写端均收尾确切归属的子进程。使用同一生产 launch 路径实际验证宿主在 ready 前及运行中被 SIGKILL 后 Go 退出、多会话 writer 不继承和独立 CLI 隔离。

本票交付可运行的 Foundation→真实 Go→HTTP→真实退出切片、对应行为回归、所有权集成证据及架构说明，不是只创建 Process 或返回一个猜测端口。临时 Foundation 拥有者可以调用生产适配器证明这一切片；它不冒充已接通 Finder、AppKit 退出协调或候选 App。

发布仅供审阅，不授权编写测试或实施。票据审阅后仍须明确确认交付边界、TDD seams、实际检查和 smoke，再获得实施批准。`tasks.md` 是唯一变更进度账本；下面的 checkbox 只记录本票验收。完整交付及独立 Standards/Spec 审阅通过后才可勾选 2.1、2.5；本票不能勾选整个 2.4。

## Source mapping

行为依据为 [macos-preview-sessions](../specs/macos-preview-sessions/spec.md)，不是旧 Swift 管理器、测试或 Go 日志。以下是本票对既有全量覆盖地图的贡献；完整原生能力仍需其他票。

| Requirement / scenarios | 本票贡献与完整能力边界 |
|---|---|
| Verified startup and actual preview URL — `A nondefault port is used`、`Open a file with spaces and non-ASCII characters`、`Startup ends without valid readiness information` | 消费真实完整 URL，校验 v1/代次/必要字段、HTTP 回环及一次 HEAD；无效、缺失、EOF、超时或早退不产生成功启动结果，清理本次子进程。浏览器请求及原生失败显示仍由 05、11 证明。 |
| Explicit stopping confirms service termination — `Stop one of several sessions`、`Stop all application sessions` | 提供进程层的关闭专用 writer、等待真实退出及 4 秒确切 owned SIGKILL 兜底；验证停止 A 不影响 B/独立 CLI。协调器的单个/全部停止、启动中停止、重复停止及状态转换由 04 完成，原生动作由 05、11 验证。 |
| Services end with their owning application — `Quit the application normally`、`Force termination of the host`、`Host exits during startup` | 建立并检查 Foundation descriptor 所有权，真实 Go 验证 ready 前/运行中宿主 SIGKILL 以及主动释放 writer。AppKit 延后退出、禁止新启动及 starting 全部清理仍由 04、05、11 完成，不能以本票临时拥有者退出冒充完整 App quit。 |
| Independent CLI operation remains available — `Launch the renderer without the application` | Foundation 操作只作用于本次确切 owned 子进程；停止单个、失败清理和宿主 SIGKILL 不终止独立 CLI。沿用 01、02 的独立参数/JSON/网络政策；完整协调器及候选隔离由 04、11 继续验证。 |

技术依据为 [design Decisions 4–7](../design.md#decisions)：适配器报告带代次的事件、不另建会话字典；参数数组、v1 NDJSON 与 64 KiB 缓冲；15 秒及一次轻量 readiness；专用 Pipe、CLOEXEC 临界区和确切 owned 停止。Decision 8 是消费真实 reload/target 反馈的上下文，不在本票新增扫描或恢复策略。Decision 10 规定唯一内置工具位置；完整开发宿主接通及 universal 候选流程仍属 05、10。

## Scope and implementation handoff

路径相对于仓库根；以下是发布时观察到的接入面，不固定新的内部类型名或增加另一套框架。

| 当前表面 | 本票需要的交付与切换边界 |
|---|---|
| `macos/GoGrip/Services/ProcessManager.swift`：原始路径字典、独立 `--json`、丢弃 stderr、EOF/超时回退 6419、仅 terminate/interrupt | 新生产进程适配器不复用这些成功判定或所有权逻辑；单次 launch 接收目标及其目录/文件类别和 generation，拥有该 Process/管道并报告事件，不负责身份去重、批次、历史或会话列表。旧宿主/管理器整体切换在 05；不为新适配器添加旧 port API、转发别名或兼容解析。 |
| `internal/managed.go`、`cmd/root.go`：01、02 已完成的 managed argv、v1 NDJSON、真实 URL、HEAD 与 owner loss | 直接使用 `Process.executableURL` 和 arguments，传 `--managed <generation>`、目录的 `-r` 和目标前的 `--`，不经 shell、不传独立 `--json`/端口/导出组合、不让 Go 请求浏览器；不重新定义 Go wire shape 或引入测试专用 CLI flags。 |
| Foundation `Process`/`Pipe` launch 与退出 | 每 child 独立 ownership stdin Pipe；串行建立 descriptor 和 launch，在 launch 前设置并检查非 stdio 端 `FD_CLOEXEC`，尤其 writer；将 Pipe 本身交给 `standardInput`，不复制 writer、不传给其他进程。进程/管道 I/O 与 HEAD 不阻塞主线程；launch 失败和后续启动失败都释放本次通道。 |
| stdout 与 stderr 消费 | 连续解析 UTF-8 NDJSON 的分片/多行，完整行及时消费，跨读取残余和诊断留存各有 64 KiB 上限；ready 后仍消费 reload/target/fatal/退出，不因停止读取而阻塞 Go。机器反馈校验 version/generation/必要字段；stderr 保留可获得诊断但不参与机器状态判断。 |
| 启动结果及退出回调 | 成功创建子进程后 15 秒覆盖反馈和一次 HEAD；只有匹配的 ready、HTTP 回环 URL、same-origin HEAD 204/匹配 `X-GoGrip-Generation` 及进程仍存活才成功。失败不 fallback；所有反馈/超时/退出结果携带本次代次，协调器的旧代次隔离及会话阶段留给 04。 |
| `macos/GoGripTests/PortReaderTests.swift`、`ProcessManagerTests.swift`、`IntegrationTests.swift` | 旧 EOF/timeout→6419、混日志找 port、mock shell echo 或仅进程存在不能成为新契约或 smoke 证据。移除与本票切换冲突的旧断言，不重钉；复用适用行为覆盖，更新真正受影响的调用。其余旧宿主测试随各自切换处理，不以本票清理无关代码。 |
| `macos/GoGrip.xcodeproj/project.pbxproj`：当前仅宿主与 FinderSync target；`GoGripTests/` 文件未配置为测试 target | 为本票源码编译及批准 seams 的行为回归落实最小可执行构建/测试入口，使用已提交工程作为工程来源，不假称现有 `xcodebuild test` 已能运行这些文件、不增加第二份 generator 配置。FinderSync/旧宿主整体去除及工程切换仍属 05。 |
| `macos/Scripts/build-go-grip.sh` 当前输出 Resources；旧管理器会 runtime chmod；`AppDelegate.swift` 用 port 拼 URL | 本票生产适配器只接受确定的 `Contents/MacOS/go-grip` 路线，不增加 Resources/PATH/源码回退或 runtime chmod；真实集成可将本次构建 Go 放入隔离临时 bundle 并调用同一生产适配器。正式开发宿主嵌入、浏览器与 AppDelegate 切换由 05 完成，不要求先实施 05 或 10 才能验证本票。 |
| `docs/ARCHITECTURE.md` | 随实际证据更新宿主协议、缓冲/期限、descriptor 所有权、停止及已验证/未验证边界；不将临时 Foundation 宿主视为 Finder/TCC/候选证明。 |

### Current machine contract

沿用 [02 的用户确认及关闭记录](02-managed-status-and-cli.md#closure-record) 和 [01 的范围裁决](01-managed-preview-lifecycle.md#closure-record)：每条消息含 `version: 1`、`event`、本次 `generation`；成功启动时 ready 是首帧且只发布一次，之前 reload 变化折叠进快照。Go 致命启动失败可能以 fatal 首帧结束，不能当作 ready。

- `ready`：完整 `url`，嵌套 `reload.state` 和可获得 `reload.reason`。
- `reload-status`：嵌套 `reload`，不是顶层 state/reason；状态为 pending/active/degraded，保留已裁决的 `--no-reload` disabled 快照语义，不因此增加 App 参数界面。
- `target-status`：嵌套 `target.state`（available/unavailable）及可获得 `target.reason`。
- `fatal`：可分类 `code`、实际 `message`；保留可获得 stderr 诊断，不从日志猜状态。

验证使用实际 ready URL 的同一 origin 的 `/__gogrip/ready` 做一次 HEAD，检查 204 和代次，不 GET 正文作为 readiness、不接受跳转到其他 origin、不加周期健康轮询。真实内容 GET 仅用于本票 smoke 证明目标身份，不能替代生产 HEAD 路径。

### Stop and ownership boundary

主动停止先关闭本 child writer，再等待真实退出；仍存活且尚未确认退出的确切 owned child 在 4 秒后使用 SIGKILL 并继续等待退出结果。不能只发信号、清一行或设置布尔值就宣称已停止，不按进程名/未知 PID 清扫；Go 的独立 2 秒善后 watchdog 保持不变。

宿主 SIGKILL 不运行正常退出回调，必须由内核关闭独占 writer 使真实 Go EOF/读错误退出。多会话必须实际证明没有其他进程继承 writer；读取源码或只检查 CLOEXEC 标志不替代此结果。不泛化为任意 Chrome 后代管理，不对不可中断内核 I/O 承诺精确硬截止。

## Prerequisites and non-goals

真正前置为 [01](01-managed-preview-lifecycle.md) 和 [02](02-managed-status-and-cli.md)，发布时均为 done；Go 回环/实际 URL、v1 ready/reload/target/fatal、HEAD 和 ownership/watchdog 已可复用。目前没有未满足的前置票。原生 UI、04 的协调器和 05 的完整开发包不是本票 blocker；为证明本票所需的真实 Foundation 构建/测试及临时 bundle 属于本票可运行切片，不是后续候选包。

03 与 04 共同贡献 2.4：本票只交付进程层停止及真实退出，04 仍须完成统一协调器的单个/全部、starting 停止、重复停止、正常/意外退出区分及代次行为。03 完成时 2.4 保持未勾选；不因为所有已发布贡献完成就忽略未发布的 04。

明确不做：规范化身份/同目标单飞、批次/数量确认、Finder Services/手动面板/浏览器动作、AppKit 退出协调、最近目标/本地化/登录项、TCC/外接或网络卷访问宣称、watcher 恢复/轮询/重连/自动重启、helper/XPC/launchd、渲染器重构、universal/archive/DMG/CI/正式分发、提交/推送、主规格同步或归档。批准 artifacts 出现缺口或范围争议时先请求独立裁决，不自行增加防御保证或永久测试。

## Acceptance

- [x] 在另行确认的 seams 上按 scoped TDD 验证分片 NDJSON、无效/缺失反馈、EOF、错误版本/代次、ready 前退出及失败清理等消费者可见行为，记录 red/green；缓冲、一次 HEAD 与有界停止按已批准 design 验证，不新增字段复制、纯转发、mock echo、资源存在、源文本或内部默认值测试，不重复 Go 已证明的内部矩阵。
- [x] 生产适配器以 executableURL/argv 启动真实 Go；每 child 有独立 ownership stdin Pipe，launch 临界区设置并检查非 stdio descriptor 的 CLOEXEC；无 shell、writer 复制、runtime chmod、Resources/PATH/源码或默认端口回退。构建/测试入口实际可执行，不以仅存在 Swift 测试文件当作测试通过。
- [x] 连续解码跨读取和跨 UTF-8 字符分片、同次读取多帧；v1/代次/必要字段按契约校验，采用 02 确认的嵌套 reload/target 形状和 ready 快照。ready 后继续消费结构化变化、fatal 和退出；反馈残余/诊断缓冲各有 64 KiB 上限，诊断不驱动机器状态。
- [x] 成功创建 child 后 15 秒覆盖机器反馈及一次同 origin HEAD；204、匹配代次、HTTP 回环 URL 和进程仍存活同时成立才提供成功启动结果。无效 URL/反馈、EOF、超时、早退或 HEAD 失败均报告失败并收尾本 child，不猜 6419、不误给出成功或遗留服务；不增加周期探测或跨 origin 跳转。
- [x] 从调用生产适配器的真实 Foundation 拥有者启动隔离临时 bundle 中的当前 Go，目录递归、所选中文/空格单文件及可访问空目录得到实际 URL；GET 内容/空状态与目标匹配，HEAD 路径与真实 listener/端口有证据。不能用 fake port、仅 isRunning、创建 Process 或旧 CLI JSON 代替此集成结果。
- [x] 真实 Go 启动访问失败及适配器判断的启动失败都有可获得原因；观察对应子进程实际结束、本次 ownership/反馈/诊断通道释放和已创建端口不再可用。受控坏反馈仅用于验证解码/失败行为，不能替代真实 managed Go 所有权证明；旧 EOF/timeout→6419 等冲突断言删除而非改成新的固定期望。
- [x] 对真实 Foundation 拥有者在 ready 前发 SIGKILL，记录已经创建的 Go PID、尚未有效 ready 的时序及其随后退出；对验证成功的拥有者再次执行运行期 SIGKILL，记录对应 Go PID 和实际监听消失。两种结果均不依赖正常退出回调、下次启动、`/bin/cat` 或另一套测试 launch 实现；不以 UI 尚未显示 running 冒充 ready 前。
- [x] 同一 Foundation 拥有者启动至少两个真实 Go 会话，主动关闭 A writer 后 A 真实退出且端口释放，B 继续提供正确内容；对拥有者 SIGKILL 后其余 owned Go 也结束，证明兄弟进程没有继承相应 writer。启动失败释放的 writer 不被随后会话持有；保留可复查的 PID/端口及实际内容结果，不只检查 descriptor 标志。
- [x] 进程层停止关闭专用 writer 并等待真实退出，4 秒后仍存活的确切 owned child 使用 SIGKILL 且确认退出结果；以受控的本次 owned 进程验证兜底，不对用户独立进程发信号、不仅删除记录或把已发信号当作已停止。协调器的 starting/重复/单个/全部语义仍由 04 验证。
- [x] 与真实独立 CLI 同时运行时，Foundation 单会话停止、失败清理及宿主 SIGKILL 均不终止独立 CLI；实际请求继续取得其原目标内容。沿用独立 CLI 的 JSON/端口/网络政策，不把 App-only 回环限制或 ownership stdin 套到独立模式。
- [x] 本票 Swift 行为回归、实际 Foundation→Go/HTTP/owner smoke 与必要构建检查通过；运行现有 Go 行为套件确认契约仍可用，记录已知既存检查失败，不以关闭警告/删除有效测试冒充通过。随证据更新 `docs/ARCHITECTURE.md`，明确内核阻塞、Chrome 后代、Finder/UI/TCC/候选及 Intel/macOS 13 未验证边界，临时 smoke 文件和本次 owned 资源清理完毕。
- [x] 完整验收及独立只读 Standards/Spec 审阅通过、无未解决范围裁决后才关闭本票，并按完整证据勾选 2.1、2.5；2.4 保持未勾选，04–11 不发布或实施。不自动提交、推送、同步主规格、归档或正式发布。

## Change-wide coverage reference

全部 26 个任务、30 个 requirement、68 个 scenario 的贡献与发布状态见 [coverage-plan.md](../coverage-plan.md)。本轮仅新增 03：01、02 已发布且 done，03 已发布且 done（证据见本票 Closure record）；04–11 均为未发布规划。多票共同贡献的 requirement 不能因 03 完成而直接视为整体完成。

## Closure record

- 关闭：2026-10-04，状态 `done`；OpenSpec 任务 2.1、2.5 按实际证据勾选；2.4 保持未勾选（协调器的单个/全部、启动中停止、重复停止与正常/意外退出区分仍属未发布的 04）。
- 变更面：`macos/GoGrip/Services/ManagedProtocol.swift`（v1 增量解码与字段/嵌套形状/64 KiB 上限校验）、`macos/GoGrip/Services/ManagedProcess.swift`（生产 Foundation 适配器：参数数组、每 child 独立 Pipe 与 FD_CLOEXEC 串行临界区、15s 期限 + 一次 same-origin HEAD、失败清理、停止与 4s 确切 owned SIGKILL 兜底、带代次事件）、`macos/GoGripTests/ManagedProtocolTests.swift`、`ManagedProcessTests.swift`、`ManagedProcessGoIntegrationTests.swift`；工程入口：`macos/GoGrip.xcodeproj/project.pbxproj`（适配器源码进入 app target + 新 GoGripTests logic test target）、`macos/GoGrip.xcodeproj/xcshareddata/xcschemes/GoGrip.xcscheme`（Build/Archive 保持原行为，TestAction 指向 GoGripTests）、`Makefile`（`macos-test`：构建 Go 并复制到系统卷临时 bundle 后运行 xcodebuild test）、`.gitignore`（`macos/.build`）；`docs/ARCHITECTURE.md` 第七节新增 Swift 宿主进程适配器契约与边界；删除 `macos/GoGripTests/PortReaderTests.swift`（旧 EOF/超时→6419 与混日志找 port 断言删除而非重钉）。旧宿主/管理器、FinderSync、`project.yml` 与旧测试文件未动（整体切换属 05）。
- Scoped TDD 证据：S1 解码器以缺失 API 的编译失败为 red，实现后通过；S2/S3 真实 Go 成功路径与受控坏反馈（坏版本/代次、非法 URL、HEAD 失败/跳转、超时、EOF、早退、诊断与残余上限）先红后绿；S4/S5 停止、4s SIGKILL 兜底、双会话 writer 不继承与独立 CLI 隔离先红后绿；审阅修正轮补充成功提交原子化、运行期违规终止与真实 redirect 回归后 41/41。
- 自动检查：`make macos-test` 41 tests / 0 failures（17 ManagedProtocolTests、17 ManagedProcessTests、7 ManagedProcessGoIntegrationTests；含生产默认 4s termination grace 实测 4.2s 后才 SIGKILL 的兜底用例）；fresh `xcodebuild build`（Debug、CODE_SIGNING_ALLOWED=NO）BUILD SUCCEEDED，新文件无警告（仅既有 AppDelegate/PopoverView 的 `run` unused 警告）；`gofmt -l .` 干净、`go vet ./...` 干净、`go test ./... -count=1` 全部通过（Go 侧零改动）。已知既存 `internal/server_test.go:339-340` errcheck 未触碰、未复跑 lint、不声称 lint 通过。
- 真实 Go 集成（生产 launch 路径，工具置于隔离临时 bundle）：目录递归与嵌套文档 GET 200；`中文 说明.md` 单文件 URL 精确定位且同目录其他 `.md` 返回 404；空目录返回 `No Markdown files` 空状态；不存在目标得到真实 `fatal target-unavailable` 且子进程已结束；两个真实会话停 A → A 真实退出（可捕获兄弟继承 writer 的回归）而 B 继续 200；失败启动后新会话正常；独立 CLI（`--json --browser=false --no-reload`）在适配器会话停止后继续 200 且进程存活。
- 真实 smoke（临时 Foundation 拥有者由同一适配器源码 + swiftc 编译，真实 Go；不依赖正常退出回调、`/bin/cat`、mock port 或仅 isRunning）：
  1) ready 前：拥有者在 spawned 回调内自 SIGKILL（确定性先于 readiness 校验，exit 137）；子 PID 已创建（argv `--managed smoke-prekill -r -- <target>` 由 `ps` 复核，PPID=1），0.27s 后子进程已不存在，`lsof` 无监听，拥有者输出无 READY。
  2) 运行中：`lsof` 记录 `TCP 127.0.0.1:49593 (LISTEN)`、GET 200；宿主 SIGKILL 后子进程退出、端口不再服务。
  3) 双会话：A/B 各自 LISTEN+200，宿主 SIGKILL 后两个 owned child 均退出。
  4) 启动失败释放：真实 fatal 的失败 child 已结束，随后会话启动并在拥有者退出后结束。
  5) 独立 CLI：owner SIGKILL 前后 GET 200、CLI 进程存活。
  临时 smoke/诊断文件与本次 owned 资源已清理。
- 独立审阅：Standards 与 Spec 两轴最终均无未解决 blocking/advisory/scope decision（修正两轮：Standards 首轮 P1 原子提交 + 4×P2、二轮 2×P2；Spec 首轮 3×P2 + 2 advisory；全部在范围内修复并 fresh re-review 通过 0 剩余）。修正包括：成功提交与失败/退出状态在同一把锁内原子判定（含真实 `process.isRunning`）；完整帧与未完成残余的 64 KiB 上限均在追加前检查；ready 后解码/校验违规以 `protocolViolation` 事件报告一次并终止该 child、持续排空 stdout；fatal 要求非空 message；redirect 回归改为真实 302 并以控制请求证明目标 origin 可达、断言未被跟随；删除重复与越权的永久测试。
- 范围裁决与修正（最小合规选项）：删除无 child 停止的独立永久测试（保留防止挂起的 guard；协调器语义属 04）；架构文档不再把重复停止/预启动停止列为本票验收；失败时 stderr 诊断尾部标注为尽力而为；SIGKILL 兜底仅针对该 owned pid（理论 PID 复用窗口无演示可达证据）。
- 未覆盖/限制：AppKit/Finder/TCC/浏览器动作与候选包属 05–11；未在 Intel/macOS 13 实际运行；不对不可中断内核 I/O 承诺硬截止；不管理 Chrome 后代；本仓库位于外置卷时 xctest 无法打开其文件（dyld/open 阻塞）且 xcodebuild 不转发自定义环境变量，故 `make macos-test` 先将工具复制到系统卷临时 bundle 再由测试使用，生产宿主的确定内置路径（`Contents/MacOS/go-grip`）仍属 05/10。

### 修正记录（关闭后第二轮独立复核）

关闭后被复核发现 1 项 P2 阻断（Standards）与 3 项建议，均已按最小合规方案修复并重新验证；修复只补足通道释放与测试确定性，不改变已批准行为，Spec 轴当轮 0 阻断。

- **P2 阻断——退出后未释放 stdout/stderr 读端**：两个读取循环在 EOF 结束后未关闭各自的读 FileHandle，适配器又持有 `Process`/`Pipe`，失败或已结束实例会依赖销毁时机，无法确认验收 6“失败启动释放反馈／诊断通道”。修复：两个读取循环在 EOF 结束处显式 `close()` 读端（退出前排空行为保留）；child 真实退出时一并释放 ownership writer（覆盖未 arm termination 的早退路径）；未产生 child 的启动失败（`Process.run()` 抛错或 CLOEXEC 检查失败）显式释放本次三个 Pipe 的两端。
- **建议 1——诊断尾部测试调度竞态**：原测试在退出时立即断言含 `DIAG-END`，与“诊断尽力而为”矛盾（快照可能在 stderr reader 排空前完成）。修复：cap 与“保留最新尾部”语义移入确定性的 `ManagedDiagnosticsTail` 直接测试（精确输入的边界用例）；适配器级测试改为只断言调度无法改变的不变量——100 KiB 超过任何管道容量，child 只能在本 reader 已排空部分 stderr 后退出（尾部非空），且不得超过 64 KiB cap。
- **建议 2——双会话测试缺少 A 清理**：`testStopOneOfTwoSessionsExitsOnlyThatOne` 在 A 启动后立即登记 teardown；B 启动失败或后续 GET 抛错时 A 也会被停止。
- **建议 3——“失败后通道释放”未检测实际释放**：原测试只证明“随后会话能启动”。改为 `testStartupFailureReleasesPipeDescriptors`：保留全部失败适配器实例，统计本进程 `/dev/fd` 中 FIFO 描述符——预热失败的释放（±2）与 6 次真实 Go `target-unavailable` 失败后计数回到基线（±2）；随后真实会话仍能启动、返回 200 并正常退出。
- **重新验证**：`make macos-test` **43/43**（ManagedProtocolTests 17、ManagedProcessTests 17、ManagedDiagnosticsTailTests 2、ManagedProcessGoIntegrationTests 7；含默认 4s grace 兜底用例）；两个调度敏感测试（`testDiagnosticsFromNoisyChildStayBounded`、`testStartupFailureReleasesPipeDescriptors`）连跑 5 次全部通过；敏感性核验：临时移除两处读端 `close()` 后新的 descriptor 测试立即失败（`retained: 6`），恢复后再次全绿，证明该测试确实检测原泄漏而非空过。`docs/ARCHITECTURE.md` 增加“通道释放”条目（writer 于停止/退出关闭、读端于 EOF 关闭、启动失败显式释放，真实 Go 失败路径以保留实例的 FIFO 计数验证）。
- 本轮无 Go 侧改动，未提交、未推送、未发布或推进 04。

### 修正记录（第三轮只读复核 P3 建议）

第三轮只读复核结论：Standards 0 阻断 / 1 建议（P3），Spec 0 发现。建议指向 `testDiagnosticsFromNoisyChildStayBounded` 替代后的弱断言：非空依赖“100 KiB 超过管道容量”的细微波时序推理，上限与最新尾部语义则已由确定性 `ManagedDiagnosticsTailTests` 固定，早退与真实退出另有覆盖。

- 处理：删除该适配器级用例（连同 `DIAG-BEGIN/DIAG-END` 噪声脚本与“非空/≤ 上限”断言），套件由 43 降为 **42**；不新增“失败结果必须携带完整诊断”的生产保证——适配器级 stderr 内容仍按已记录语义为尽力而为，通道打开与释放由 `testStartupFailureReleasesPipeDescriptors` 覆盖。`ManagedProcessTests` 的节标题同步去掉 `diagnostics`。
- 本轮未改生产代码、票据正文、账本与规格；未提交、未推送、未发布或推进 04。
