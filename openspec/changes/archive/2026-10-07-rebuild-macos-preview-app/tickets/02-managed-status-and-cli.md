# 02 — Go managed 状态反馈与独立 CLI 共存

Status: done
Blocked by: 01-managed-preview-lifecycle
Covers OpenSpec tasks: 1.3, 1.4
Behavior source: ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

从真实 managed Go 进程获得目标访问与热重载覆盖的结构化初始状态及后续变化：可访问空目录不等于不可访问，已知监视降级不等于服务退出；可访问内容在降级时仍可通过 HTTP 和浏览器手动刷新获取。同时实际验证独立 CLI 的目录/文件、端口、JSON 与所有权隔离契约，交付对应行为回归、真实文件事件/HTTP/CLI smoke，以及使用与架构说明。

本票完成 Go 侧状态与共存切片，供后续原生进程适配器消费；不宣称已经实现 App 状态显示、Finder、TCC 或候选包验收。发布仅供审阅，不授权编写测试或实现；实施前须另行确认交付边界、TDD seams、真实 smoke 方法并获得明确批准。

`tasks.md` 是唯一变更进度账本；下面的 checkbox 仅为本票验收证据。只有整票交付、验证与独立 Standards/Spec 审阅均完成后，才能关闭本票并据完整证据勾选 1.3、1.4；不检查其他任务。

## Source mapping

行为依据为 [macos-preview-sessions](../specs/macos-preview-sessions/spec.md)，不是现有实现、测试或日志文字。本票只贡献下列 requirement 的 Go 部分；其余贡献见完整覆盖地图。

| Requirement / scenarios | 本票贡献与完整能力边界 |
|---|---|
| Directory and single-file preview modes — `Preview nested documents`、`Preview a selected file`、`Keep an empty directory session`、`Add Markdown to a watched empty directory` | 保留首票的目录/文件/空状态基础；证明有效监视的空根新增 Markdown 后，现有热重载链路可使同一服务显示新文档，不需重建。原生打开与候选浏览器链路仍属 05、11。 |
| Unavailable paths are not empty directories — `A target is moved or deleted`、`A mounted volume becomes unavailable` | 从实际访问原根目录/所选文件或相关文件事件发现的失败反馈 unavailable 与原因，不跟踪新位置、不误报空目录。实际挂载卷/native 权限和面板入口由 06、07、11 验证，协调器状态由 04 消费；本票不以本地文件失败冒充断卷/TCC 证明。 |
| Visible hot-reload degradation — `Watch coverage is incomplete`、`Use a network-volume preview` | 初始及运行期已知监视失败/缺失覆盖的结构化状态和原因；降级保留可访问内容及手动刷新，不改变监视政策、不拒绝网络卷身份、不承诺捕捉静默丢事件。真实网络卷访问与原生显示仍需 06、07、11。 |
| Independent CLI operation remains available — `Launch the renderer without the application` | 完整验证独立目录/文件预览、默认端口回退、显式端口严格策略、`--json`、不消费 ownership stdin，以及停止一个 managed 服务不影响另一 managed 和独立 CLI。生产 Foundation/协调器/App 操作隔离仍需 03、04、11。 |

技术依据为 [design Decisions 5–8](../design.md#decisions)：v1 序列化机器事件、真实 ready 快照、一次轻量 readiness、保持所有权生命周期、结构化访问与监视状态，不扩展扫描政策。

前票已确认的范围裁决继续有效，见 [01 Closure record](01-managed-preview-lifecycle.md#closure-record)：`--no-reload` 的快照为 `disabled`；目录按目录级可访问性判定，单个文章不可读按请求级处理，不能一律使整个根目标 unavailable；显式非激活标志按值判断。本票不重开这些裁决或新增高级配置。

## Scope and implementation handoff

路径相对于仓库根；以下是发布时已观察到的接入面，不预设新内部类型或额外框架。

| 当前表面 | 本票需要的交付 |
|---|---|
| `internal/managed.go`：`managedWriter` 序列化 ready/fatal，ready 仅有 `reload.state` 初始快照 | 复用同一 v1 NDJSON 输出路径，接入 `reload-status` 与 `target-status` 及原因；每条仍带 version/generation，不混日志、独立 JSON 或并发拼接帧。初始已知降级及原因不因扫描/ready 时序丢失；不改实际 URL 或所有权契约。 |
| `internal/hotreload/hotreload.go`：`New` 立即启动扫描，`State` 只提供初始覆盖，运行期部分失败仍仅写日志 | 在初始扫描前接入可选状态出口，覆盖 watcher 创建、walk、预算跳过、注册和运行期 watcher 错误；提供真实快照，后续只报告实际状态或相关原因变化。没有宿主接收方时保留 CLI 日志，不为每个文档请求新增状态通道。 |
| 同文件的已有 root watch、Create/Write/Rename/Remove、目录注册与预算路径 | 保持预算计算、忽略目录、扫描/发现规则和已有有效监视链路；验证空根新增文档及初始/运行期覆盖不足的状态。已知不完整不能假 active，也不能用永久 pending 掩盖完成扫描。 |
| `internal/server.go`、`internal/target.go`、`internal/articles.go`：目标解析、根目录发现、所选文件读取与现有 HTTP 错误/空状态 | 从实际访问结果区分 available/unavailable 与原因；只有确认可访问且没有 Markdown 才显示空状态。普通图片/不存在的子页面 404、非目标文章失败不一律污染根目标；单文件失效不能降级为父目录预览。保留现有错误页和文章路由，不重构渲染器。 |
| `cmd/root.go`、`internal/listener.go`、独立 `Server.Serve` | 保留独立参数与输出：未显式指定端口时默认 6419 并按现有策略回退；显式 `--port` 冲突即失败；独立 `--json` 使用实际端口/完整 URL，不成为 managed 事件流。独立 stdin EOF 不触发 owner loss，既有网络暴露政策不变。 |
| `internal/hotreload/hotreload_test.go`、`internal/managed_test.go`、`internal/server_test.go`、`internal/listener_test.go`、`cmd/managed_test.go` | 优先扩展现有行为/真实子进程契约；确定性预算/注册/错误 seam 用于证明降级语义，真实文件事件与 HTTP 验证内容和隔离。不得按层重复同一消费者行为或新增字段复制、mock echo、纯转发、源文本测试。 |
| `README.md`、`docs/ARCHITECTURE.md` | 更新结构化状态、空目录/失效目标/降级差异、手动刷新与重新打开的操作边界，以及目录/文件 CLI、端口与 `--json` 用法；明确 App-only 回环不代表所有 CLI 都仅限本机。 |

`reload-status` 按设计使用 `pending` / `active` / `degraded` 及可获得原因；保留已裁决的 reload 关闭语义。`target-status` 使用 `available` / `unavailable` 及实际访问失败原因，不从 stderr 推断状态。ready 仍只成功发布一次，必须反映当时已经知道的初始化降级；本票不重新定义 Swift 的解码字段校验、64 KiB 缓冲或 15 秒期限。

发现原路径失效后保留服务及可查看、可停止的机器状态，不追踪移动或后台重连。实际再次访问原路径成功可以反馈 available，但不能据此宣称已损失的 watcher 自动恢复；不增加周期访问、监视恢复或重启。只对实际观察到的访问与文件事件作声明，不承诺静默失效的检测期限。

## Prerequisites and boundaries

唯一实现前置为 [01-managed-preview-lifecycle](01-managed-preview-lifecycle.md)，发布时其状态为 done，任务 1.1、1.2 已完成；Go 的回环监听、事件 writer、readiness 与 ownership 生命周期可复用。当前无未满足的前置票，不依赖后续 Swift/UI 或打包。

本票必须保持 owner EOF/读错误、有界关闭与独立 watchdog，不将状态反馈变成另一套所有权机制。真实 smoke 只停止本次创建的确切 owned 进程，不能占用端口后清扫用户服务或按进程名杀进程。

明确不做：Swift/Foundation/AppKit/Finder 修改、原生授权/TCC 宣称、面板/最近目标/登录项、本地化、扫描预算或忽略规则调整、轮询/重连/监视恢复/自动重启、helper/XPC/launchd、渲染界面重设计、打包/CI、正式签名公证、提交/推送/发布、主规格同步或归档。范围争议或批准 artifacts 的缺口须先请求独立裁决，不以本票补写行为保证。

## Acceptance

- [x] 在明确批准的 seams 上按 scoped TDD 验证任务 1.3、1.4 的消费者可见回归，记录失败前/通过后证据；复用既有有效覆盖，不重复首票完整矩阵，不新增 wiring、字段复制、mock echo、日志文字、资源存在或内部默认值测试。
- [x] 用初始扫描前已安装的状态出口覆盖 watcher 创建/walk/预算/注册失败，ready 快照保留当时已知降级及可获得原因；扫描完成后能反馈真实 active 或 degraded，后续运行期已知监视错误/新增目录覆盖不足产生结构化变化，不能丢初始化失败、假 active 或永久 pending。
- [x] 所有 managed 状态反馈是可连续解析的 UTF-8 v1 NDJSON，每条 version/generation 正确，ready 同代次仅一次；并发真实请求与文件事件仍经过同一序列化输出路径，stdout 无普通日志/独立 JSON，stderr 不被当作状态输入。
- [x] 对可访问空根运行真实 managed 进程，确认初始空状态及有效监视后建立现有 WebSocket 热重载连接；新增 Markdown 观察实际 reload 信号，再请求同一服务验证新增文档内容，无需停止/重建。此 Go/WebSocket smoke 不冒充后续候选浏览器或原生界面验收。
- [x] 运行中移动/删除原根目录及所选单文件后，通过实际访问或相关文件事件观察 target unavailable 和原因，HTTP 不误显示空目录或其他目标；不改为新路径/父目录、不自动重连或重启，服务仍可通过原所有权通道停止。初始化访问失败沿用 fatal 退出且不发布成功 ready；实际原路径再次可访问时允许反馈 available，但不能假称 watcher 恢复。
- [x] 可访问根中的普通缺失图片/不存在子页面请求保持其原 HTTP 错误且不使目标 unavailable；目录内单个文章访问失败不一律污染根状态。真实空目录与根不可访问有不同结果，不以无关 404 或空文章列表推断目标失效。
- [x] 在可控预算/注册失败及运行期监视错误下证明结构化 degraded 与可获得原因；真实 HTTP smoke 验证降级服务仍能读取可访问文档，修改内容后手动刷新取得新内容。不更改既有预算/忽略/发现规则、不加轮询，不宣称网络卷自动刷新可靠或本地 smoke 已证明实际断卷/TCC。
- [x] 实际独立 CLI 在无 App 的情况下预览目录（含递归内容）和所选单文件；`--json` 仍输出实际端口/完整目标 URL 而非 managed 帧，HTTP 正文与目标匹配，关闭独立 stdin 不终止服务；观察实际 listener 并保持既有网络政策，不能仅凭 URL 的 localhost 宣称全 CLI 回环。
- [x] 在隔离且由本次 smoke 持有的端口占用条件下，实际无显式端口的 CLI 从默认 6419 按原策略回退并提供正确内容，JSON/URL 与实际端口一致；实际显式 `--port` 冲突失败且不另开端口。若默认端口已有用户服务，不停止它，记录占用并按原策略验证；完成后释放本次创建的占位 listener。
- [x] 同时运行两个 managed 服务和独立 CLI，关闭一个 managed writer 后观察其真实 PID/端口释放，另一 managed 与独立 CLI 继续提供各自内容；独立 stdin EOF 不触发 managed 生命周期。本票结果不冒充生产 Foundation/App 退出隔离。
- [x] 运行对应 Go 行为回归及 `go test ./...`，记录真实 binary/HTTP/文件事件/端口占用/owned 停止证据；同步 `README.md` 和 `docs/ARCHITECTURE.md` 的降级、手动刷新与 CLI 用法/网络边界，明确未验证环境，通过独立 Standards/Spec 审阅后才关闭本票并按完整证据勾选 1.3、1.4。

## Change-wide coverage reference

完整的 26 个任务、30 个 requirement、68 个 scenario 的贡献与发布状态见 [coverage-plan.md](../coverage-plan.md)。本轮仅发布 02：01 已发布且 done，02 已发布且 done（证据见本票 Closure record），03–11 仍是未发布规划；后续票不会因本票发布或完成而自动获得授权。

本票仅覆盖 **1.3、1.4**。coverage 地图不是第二份进度账本，本票 checklist 也不能替代 `tasks.md`；多票贡献的 requirement 仍须全部组件贡献及候选实际验收完成。

## Closure record

- 关闭：2026-10-03，状态 `done`；OpenSpec 任务 1.3、1.4 按完整证据勾选，其他任务未勾选。
- 变更面：`internal/hotreload/hotreload.go`（`StateReporter`/`NewWithReporter`、`Status()` 带原因、初始与运行期降级出口、`watchIO`/`openSource` 确定性 seam）、`internal/managed.go`（`reload-status`/`target-status` 帧与 `reason`、ready 快照含已知降级且为首帧、发布快照去重、reload/target 出口接线、目录级 `verifyTargetAccess`）、`internal/server.go`（`TargetStatusReporter` 与真实访问分类）、`README.md`、`docs/ARCHITECTURE.md`；测试：`internal/hotreload/hotreload_test.go`、`internal/managed_test.go`、`cmd/managed_test.go`、新增 `cmd/cli_test.go`。
- Scoped TDD 证据：reporter/Status/seam 与 managed 事件字段先以缺失 API 的编译失败为 red，再实现通过；按审阅修正补 `TestManagedReloadStatusSuppressesSnapshotDuplicate`（ready 快照去重）并把运行期覆盖不足测试改为经真实 watch 循环 Create 分发驱动。
- 自动检查：`gofmt -l .`、`go vet ./...` 干净；`go test ./... -count=1` 全部通过；`go test -race ./internal/ ./cmd/ -count=1` 连续 5 轮通过；`golangci-lint run ./...` 仅剩既有 `internal/server_test.go:339-340` errcheck（本票未触碰该文件）。
- 真实 smoke（darwin，真实二进制，本次亲见）：目录 ready → GET 200 → 删除根 → 打开中的文章 URL GET 404 且 stdout `target-status unavailable`（reason 为原根路径）→ 重建 → GET 200 + `target-status available` → owner stdin 关闭 exit 0 且端口关闭；`ulimit -n 128`（kqueue 预算截断）下 ready 快照或 `reload-status` 报 `degraded` + 预算原因（两种时序均观察到），GET 200 v1 → 改写 → GET 200 v2（手动刷新），exit 0；空根 + WebSocket：ready/空状态 → 新建 Markdown → 收到 `reload:first.md` → 同一服务 GET 新文档（cmd 真实二进制测试）；CLI：`--json` 实际端口/完整 URL、递归与单文件内容 200、关闭 stdin 仍 200、lsof 观察 `TCP *:<port> (LISTEN)`；默认 6419 被本次占位时回退 6420 且内容正确；显式 `--port` 冲突 exit 1、stderr 指明端口、stdout 为空；两 managed + 独立 CLI 共存，停止一个 managed 后其端口释放、其余继续服务。
- 独立审阅：Standards 与 Spec 两轴最终均无未解决 blocking finding。首轮 findings 均在范围内修复并 fresh re-review 通过（0 剩余）：target 重检与状态迁移串行化；直接文章请求发现根丢失时报告 unavailable；ready 快照重复上报抑制；运行期覆盖不足测试改走真实分发。
- 范围裁决（用户确认）：线协议形状（ready `reload.reason`、`reload-status`/`target-status` 嵌套对象）、ready 首帧与发布前变化折叠进快照、目录级 vs 单文件目标判定、`openSource` 测试 seam。
- 未覆盖/限制：网络卷与断卷、TCC/原生授权、Finder/面板显示、候选包验收属 05/06/07/11；预算降级真实 smoke 依赖 kqueue 平台（`ulimit` 生效），非 darwin 由 unit seam 覆盖、cmd 测试显式跳过；未声称独立 CLI 全为回环（实测为通配绑定）、未声称 watcher 恢复、静默事件丢失检测或网络卷自动刷新可靠；未做 Intel/macOS 13 运行。
