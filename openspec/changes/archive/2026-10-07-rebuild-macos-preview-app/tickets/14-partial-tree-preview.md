# 14 — 不可读子树不阻断可访问目录预览

Status: done
Blocked by: 01-managed-preview-lifecycle, 02-managed-status-and-cli
Covers OpenSpec tasks: 1.1 (initial preview/startup corrective contribution), 1.3 (partial-tree access and reload degradation corrective contribution), 1.4 (affected standalone CLI compatibility regression only)
Behavior source: ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

修正 13 实际候选验收发现、用户已裁决为缺陷的 Go 侧行为：**所选根目录可访问、但包含权限拒绝的子树时，目录会话仍能建立真实可用的递归预览，浏览和手动刷新可读内容，同时通过既有结构化状态报告实际已知的热重载覆盖降级**。不可读子树不能在初始文章发现或后续页面目录发现中使整个可读根失败，也不能将整根误标为 unavailable。所选根目录或所选单文件本身不可访问仍按既有失败契约处理，不绕过权限。

这是对已批准规格与已有 01/02 交付的缺陷补正，不新增产品保证、不修改行为规格、不重设计发现顺序或监视政策。14 交付 Go 修复、对应消费者行为回归、真实 Go/HTTP/状态/手动刷新 smoke 及 README/架构说明；候选重新交付沿用 12，最终候选受影响路径复验沿用 13，均须另行批准，不在本次发布中执行。

## Evidence and authority

- 缺陷记录：[13 Execution record](13-candidate-functional-acceptance.md)，定位 `Execution record — 2026-10-06` 的“未执行/阻塞”段：可访问目录只要含一个 `chmod 000` 子目录，即以 `target-unavailable: resolve initial preview path: … permission denied` 使整个会话失败；记录同时指出 TCC 受限子树路径。用户已裁决为缺陷，并要求回到组件修正、12 重新交付后复验。该实测是本票起点，不要求为发布重复制造系统授权或故障。
- [macos-preview-sessions / Directory and single-file preview modes](../specs/macos-preview-sessions/spec.md#requirement-directory-and-single-file-preview-modes)：目录默认递归浏览可访问 Markdown，单文件仍直达所选文件；`Preview nested documents` 是本票的可用内容要求，不能只补降级标签却让整个会话继续失败。
- [Visible hot-reload degradation](../specs/macos-preview-sessions/spec.md#requirement-visible-hot-reload-degradation) / `Watch coverage is incomplete`：已知监视错误或不完整覆盖必须可见，服务与可读内容的手动刷新仍可用。不是“忽略所有错误”，不是完整自动刷新承诺。
- [Verified startup and actual preview URL](../specs/macos-preview-sessions/spec.md#requirement-verified-startup-and-actual-preview-url)：修复后 ready、实际 URL 与 HTTP 服务必须一致；根不可访问时不产生假 ready、固定端口回退或遗留服务。
- [Unavailable paths are not empty directories](../specs/macos-preview-sessions/spec.md#requirement-unavailable-paths-are-not-empty-directories)：区分所选根目标访问失败、子树访问失败和真正空目录；不能把根访问失败改写为成功空状态。
- [Independent CLI operation remains available](../specs/macos-preview-sessions/spec.md#requirement-independent-cli-operation-remains-available)：共享发现代码的修复不能破坏独立目录/单文件预览、实际 URL/JSON、端口政策或所有权隔离。
- [design Decision 8](../design.md#8-结构化访问与热重载状态不增加扫描政策)：所选根/单文件与子页面的错误作用域不同；watcher walk error 属已批准的降级触发；保留既有忽略目录、预算与 Markdown 发现规则，不加轮询。
- [01 Closure record](01-managed-preview-lifecycle.md#closure-record) 与 [02 Source mapping](02-managed-status-and-cli.md#source-mapping) 的既有裁决：目录按目录级可访问性判定，单个文章不可读不能一律污染整根状态。不得借本票重开该裁决。

## Source mapping and progress boundary

| 物理任务标签 | 本票贡献 | 完整完成条件与后续贡献 |
|---|---|---|
| 1.1 | 修正可访问目录的初始文章/URL 发现被不可读子树阻断；证明真实 ready 和可用 HTTP | 原 01 的启动/所有权保证保持；本票不是整个 1.1 的重新实现或完整矩阵替代 |
| 1.3 | 子树权限错误不污染根可访问状态，已知 watch 覆盖不足仍结构化 degraded，页面发现/手动刷新可访问内容 | 原 02 的状态/事件/生命周期保证保持；14 修复与独立审阅完成前不能把这一缺陷贡献当作已完成 |
| 1.4 | 共享 Go 修改影响的独立 CLI 目录/单文件/JSON 与 ownership 隔离回归 | 不改变独立 CLI 的网络政策、端口回退、输出格式或发布资产；无受影响路径时复用现有回归，不新增重复测试 |

**发布时账本冲突明确保留**：CLI 当前为 **22/26**，1.1、1.3、1.4 及 6.1–6.4 已勾；01/02/12 历史票为 done；08、13 为 open。新缺陷否定 1.1/1.3 的上述部分，并增加 14 这一未完成贡献者。发布不自动撤销勾选、不重写历史验收或改动票 12/13。实施前须就 1.1/1.3 的重开和后续候选重新交付的账本处理获得明确批准；1.4 是兼容复验而非已发现 CLI 缺陷。若需修订 tasks 或既有票，逐文件另行批准；不能带着冲突声称所有对应任务已完成。

覆盖地图另列 14 → 沿用 12 重新交付 → 沿用 13 复验的顺序。14 不勾 6.1/6.2 或 7.1–7.3，不关闭 08/3.3，不以 Go smoke 冒充候选原生/平台证据。14 关闭不等于新候选已经交付或 13 可以关闭。

## Implementation handoff

发布时已读取的代码接入面（路径相对于仓库根）：

| 接入面 | 已观察现状与修复约束 |
|---|---|
| `internal/articles.go` / `discoverArticlesInDir` | `os.ReadDir` 失败返回 error；递归子目录的 error 直接向根传播。区分所选根与权限拒绝的子树，使可读兄弟内容保留；不全局吞错，不改变可读文档排序、Markdown 类型或递归开关 |
| `internal/managed.go` / `managedRuntime.serve` | `initialPathForTarget` 失败关闭 listener/reloader 并发 `target-unavailable` fatal。修复真正的初始内容发现，而不是屏蔽 fatal 或随意改成根 URL 让后续 HTTP 仍失败 |
| `internal/server.go`、`internal/target.go` | HTTP 与初始 URL 消费同一目录发现语义；实施时检查所有受影响消费者，保证实际页面、导航和手动刷新能使用可读内容，根失败/单文件失败边界不被共享修复抹掉 |
| `internal/hotreload/hotreload.go` | 已有初始/运行期 walk error 降级和 reporter。先复用已有机制；只有真实证据显示必要反馈缺失才在批准范围补齐，不新增轮询、预算策略或第二种状态通道 |
| `internal/managed_test.go`、`internal/server_test.go`、`internal/articles_test.go`、`internal/hotreload/hotreload_test.go`、`cmd/managed_test.go`、`cmd/cli_test.go` | 优先现有真实进程/HTTP 消费者缝；watch walk error 已有测试，不能重复添加 reporter/mock echo 测试代替“服务实际可用” |
| `README.md`、`docs/ARCHITECTURE.md` | 明确可访问根中的不可读子树与根不可访问的差异、覆盖降级/手动刷新边界；不宣称 TCC 被绕过、所有目录完全可读或完整自动刷新 |

实施时重读最终文件与所有受影响调用。以上是已观察入口，不预先锁死内部 API、返回值形状或测试注入设计。以最小修复解决真实权限拒绝的子树；其他错误类别若需改变现有行为，先说明实际触发、影响与规格依据并请求 scope decision，不借机定义通用扫描容错框架。

## Scoped TDD and real smoke proposal

**以下是拟议 seams，发布不是测试/实施批准。** 开工前确认本票范围、seams、根/子树分类边界、具体 native/数据操作以及进度账本调整。

1. 第一条 red：在现有真实 managed 进程/HTTP 缝，专用可访问根同时含可读 Markdown（含嵌套内容）与权限拒绝子树。断言得到实际 ready/回环 URL、GET 返回正确可读正文；既有实现应失败，最小修复后通过。这是 1.1/1.3 的真实缺陷回归，不以内部调用次数或忽略错误的返回值作为验收。
2. 结构化状态：复用现有 watch 错误覆盖；在同一真实权限 smoke 中观察 ready 快照或后续 `reload-status` 报实际 degraded 与原因，根仍可访问，改写可读文档后同一服务手动 GET/刷新得到新内容。若 watcher 的初始扫描尚未结束，允许既有 pending → degraded 时序；不强钉 NDJSON 帧次数或诊断文字。
3. 页面发现与根边界：只有第一条未覆盖的消费者风险才增加独立回归——后续正文/导航请求不能再次被同一子树阻断；所选根或所选单文件本身权限拒绝仍失败，无 ready/假空目录/残留进程。优先已有测试，不在多层重复相同 fixture/assertion。
4. CLI 兼容：运行受影响现有目录/单文件/JSON/端口及 ownership 隔离回归，真实 CLI 与 managed 同时运行时只停止确切 owned 服务，CLI 仍能提供原内容；不新增 CLI 网络政策或针对不存在缺陷的重复边界用例。

真实权限 fixture 仅使用本轮自建可丢弃目录；清理前还原权限，删除自己创建的文件，不操作用户目录。若运行身份可绕过 `chmod 000`，不得把成功访问冒充拒绝权限：采用已批准的确定性访问错误 seam 做永久回归，另在有效非特权环境观察真实拒绝。TCC 是系统实测，不由 POSIX 权限测试代签；需新 TCC 操作时单独批准。

验证命令候选：相关 focused Go 行为回归、`go test ./... -count=1`、适用 `go test -race`/`go vet`/gofmt/lint；实际构建并运行真实 Go binary 的 HTTP/NDJSON/owner-stop smoke。确切命令按实施时工具与现有约定确认，仅报告已执行结果；不得把既有 lint 缺口隐藏为通过。

## Prerequisites, redelivery and non-goals

- 真正前置：01/02 的基础 managed/结构化状态与 CLI 已交付，当前两票均 done；无未完成前置票。13 的缺陷记录是输入，不要求 13 先关闭，避免依赖环。
- 14 单独交付 Go 修复与真实 Go 可用性，不在本票修改 Swift 宿主、Services、登录项、本地化、打包或 CI。若必要修复超出 Go 侧已批准契约，先请求 owning artifact/ticket 修订，不静默扩大范围。
- 后续候选：14 完整验收及两轴独立审阅后，另行批准沿用 12 的显式 archive/签名/App zip/DMG 流程重新交付，记录新候选身份与旧候选区别；不得替换包内工具后手工重签冒充统一 archive，也不复用旧候选的授权成功结论。
- 后续复验：13 对新候选重跑受影响的原生 Finder/手动预览、降级显示、浏览器可读内容/手动刷新与 owned 停止隔离；其他原候选证据须标明版本，不自动套用。13 的网络卷/断卷/系统拒绝浏览器/Help/Intel/macOS 13 等未完成项仍按原批准范围处理。
- 不增加权限提升/FDA/helper/detach、后台重连/自动恢复/轮询、监视预算/忽略规则/发现排序变化、重试框架、通用错误吞并、浏览器渲染改造或无关 CLI 容错。不会为了保留可读内容把真正根不可访问误报为成功。
- 不在发布时生成/安装候选、操作 TCC/Services/登录项/系统语言、重启 Finder、修改用户数据；不提交/推送/正式发布/主规格同步/归档。

## Acceptance

- [x] 实施前另行批准本票范围、TDD seams、真实 smoke/数据操作及 1.1/1.3 与候选重新交付的账本处理；明确既有 done 票历史证据与新缺陷贡献不冲突地记录，未覆盖用户改动。
- [x] 可访问根含权限拒绝子树时，真实 managed 启动得到可用实际 URL/ready，根和可读嵌套 Markdown 正确服务；不再以该子树错误导致整个会话 fatal，URL 不是固定端口或随意兜底地址。
- [x] 后续正文/导航发现与手动刷新保持可读内容可访问，同一 child 无需重启即可返回改写后的正确内容；不是只让启动通过而页面继续报整根错误。
- [x] watcher 已知 walk/覆盖不足经既有 v1 状态通道报告 degraded 与可获得原因，初始或后续时序均不丢失；子树失败不将可访问根误报 unavailable，服务仍可用，不伪称完整热重载或自动恢复。
- [x] 所选根或所选单文件本身不可访问仍按真实失败契约处理，无假 ready/空状态/运行会话/残留进程；真正可访问空目录与缺失根的现有行为不回退。
- [x] 受影响共享发现调用全部一致迁移，既有递归/单文件、可读文档排序、忽略规则与 watch 预算保持；只处理有依据的子树访问拒绝，不吞并未批准的错误类别或绕过权限。
- [x] scoped TDD 留存第一条缺陷回归的失败前/通过后证据，相关现有 Go 回归与适用格式/静态检查通过或准确列明既有缺口；永久测试不重复 reporter/wiring/文字/字段复制或同路径多层断言。
- [x] 真实 binary/HTTP/NDJSON smoke 配对记录目标、实际内容/URL、降级与手动刷新结果、PID/端口；owner 停止释放本次 owned 服务，另一服务与独立 CLI 保持可用，临时 fixture/脚本清理。
- [x] README 与架构说明随修复更新，明确根/子树边界和降级限制，不把 POSIX 权限拒绝观察当作候选 TCC、网络卷或完整原生矩阵证明。
- [x] 独立 Standards/Spec 审阅无未决阻塞及 scope decision 后才关闭 14；任务勾选按所有贡献者及整项标准实际核对；12 重新交付与 13 新候选复验另行批准，08/3.3 和 7.1–7.3 未完成项不因此关闭。

## Whole-change coverage checkpoint

完整 26 个 checkbox、30 个 requirement、68 个 scenario 及全部已发布/后续贡献见 [coverage-plan.md](../coverage-plan.md)。本轮只新增 14（open）并刷新覆盖地图；01–07、09–12 的历史 done 状态不改；08/13 保持 open，22/26 勾选保持且明确 1.1/1.3 的缺陷冲突待实施前裁决。不新增规范场景、不修改 proposal/spec/design/tasks、不把 14 发布当作开始实施许可。

## Closure record

- 关闭：2026-10-06，状态 `done`。用户批准：范围与 TDD seams（首红 cmd 真实进程、根边界 in-process）、容忍边界（仅 `fs.ErrPermission`）、真实 smoke/数据操作（仅自建可丢弃 fixture）、账本处理（**1.1/1.3 勾选保持不动，关闭时按完整贡献核对并记录**）。未改动 `tasks.md` 勾选；未重开历史验收。
- 变更面：`internal/articles.go`（递归子目录读取错误属权限类时跳过该子树并保留可读兄弟；根自身读取错误与其余错误类别保持原传播）、`cmd/managed_test.go`（新增真实 binary 回归 `TestManagedAccessibleRootWithUnreadableSubtreeServesReadableContent`）、`internal/managed_test.go`（新增根边界回归 `TestManagedServeRejectsUnreadableDirectoryRoot`）、`README.md`（2 处边界说明）、`docs/ARCHITECTURE.md`（§5.4 语义、managed 两处说明、验证记录）。未触碰 Swift/协议/构建/CI；发布轮未提交的 openspec 文件为用户改动，保留。
- Scoped TDD 证据：真实进程回归先红——`first event = {Event:fatal Code:target-unavailable Message:resolve initial preview path: …/denied: permission denied}`（stderr 同步 `HotReload: walk error at …/denied: … permission denied`）；修复后同一用例通过。根边界回归在修复前后均通过（钉住未被吞掉的根读取错误）。
- 自动检查：`go test ./... -count=1` 全部包通过；`go test -race ./internal/... ./cmd/... -count=1` 3 轮通过；`go vet ./...`、`gofmt -l .` 干净；本机无 `golangci-lint`，已装 staticcheck 二进制（go1.24 构建）无法加载 go1.26 模块，均如实列明而非报为通过；既有 `internal/server_test.go:339-340` errcheck 缺口不在本票范围。
- 真实 smoke（darwin，真实二进制与真实 HTTP/NDJSON，本次亲见）：可访问根含 `chmod 000` 空子树 → managed PID 37458 `http://127.0.0.1:54531/…` ready（reload 初始 `pending`），根与嵌套正文 HTTP 200，随后 `reload-status degraded`（reason 为 denied 路径的 walk error），全程 0 个 `target-status unavailable`；改写 README 后同一服务 GET 得到新内容；关闭 ownership writer → exit 0 且端口关闭；独立 CLI PID 37463（`*:54538`，既有通配政策）在 managed 停止后继续服务，既有 `TestManagedStopLeavesOtherSessionsServing` 覆盖双 managed + CLI 隔离；fixture 权限还原后删除，无残留进程。
- 独立审阅（只读子代理，未复跑实现者检查）：Spec 轴首轮 0 阻塞 + 1 P3 文档 advisory，Standards 轴 0 阻塞 + 2 P2 文档 advisory（同一条“not served/listed”过度表述 + “Only”句过度收窄容忍边界）；两条 README 表述按最小修正在范围内改窄（discovery 排除 ≠ 直接请求禁止；非权限类子树错误保持原失败行为），两位审阅者定向复核均确认 resolved；两轴最终 0 阻塞 / 0 advisory / 0 scope decision。
- 账本核对（按用户裁决保持勾选）：1.1 = 01 + 14、1.3 = 02 + 14、1.4 = 02 + 14（兼容复验，未发现 CLI 缺陷）的全部贡献者均 done，整项标准在集成状态可观察（scoped TDD、真实 Go smoke、文档更新）；CLI 账本 **22/26** 不变（未勾 3.3、7.1–7.3），08/13 保持 open；`openspec instructions apply` 复核 26 项 / 22 完成与文件一致。
- 未覆盖/限制：本票不提供候选 App/DMG、TCC、网络/断卷、Intel/macOS 13 或原生面板显示证据（POSIX `chmod` 观察不代 TCC）；chmod 类回归在权限不被强制（如 root 运行身份）时显式 skip，不冒充拒绝；12 重新交付与 13 新候选复验、主规格同步/归档须另行批准。
