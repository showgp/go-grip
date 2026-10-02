# 01 — Go managed 预览启动与所有权生命周期

Status: open
Blocked by: None
Covers OpenSpec tasks: 1.1, 1.2
Behavior source: ../specs/macos-preview-sessions/spec.md
Design source: ../design.md
Task source: ../tasks.md
Proposal source: ../proposal.md

## Deliverable

从持有专用 stdin 写端的拥有者程序直接启动真实 Go 工具，能够取得并访问目标的实际回环预览 URL；正常关闭该写端或强制终止拥有者时，Go 服务结束并释放其监听与监视资源，包括仍在启动的服务。独立 CLI 不进入该所有权路径。

本票交付可运行的 Go managed 纵向切片、对应行为回归、真实子进程/HTTP smoke 及架构文档，不是只声明参数或协议的占位实现。它为新原生宿主提供基础，不宣称已完成 Finder、Swift 会话管理或候选 App 验收。

本轮只发布本票供审阅；尚未批准实施。`tasks.md` 是唯一变更进度账本；下列 acceptance checkbox 是本票的验收细节，不代替 OpenSpec 任务进度。票据保持 open，直至整张票验证完成并通过独立 Standards/Spec 审阅。

## Source mapping

行为依据是 [macos-preview-sessions](../specs/macos-preview-sessions/spec.md)，不是当前实现或旧测试：

| Requirement | 本票贡献及完整能力边界 |
|---|---|
| Directory and single-file preview modes | 实际递归目录、单文件与可访问空目录的 Go 服务基础；原生入口、浏览器动作及空目录新增文档的完整链路仍由后票覆盖。 |
| Verified startup and actual preview URL | 提供可用服务的真实 URL、启动确认与失败退出；Swift 对反馈/期限/进程存活的验证由后票覆盖。 |
| Loopback-only application previews | managed 模式实际绑定回环，不只改变展示 Host。 |
| Explicit stopping confirms service termination | 关闭所有权写端使真实服务退出并释放资源；App 停止一个/全部及状态协调由后票覆盖。 |
| Services end with their owning application | 实现 Go 所有权监视与独立退出 watchdog；生产 Foundation launch 的 descriptor 继承验证和 App 正常/异常退出仍需后票证明。 |
| Independent CLI operation remains available | 不将 ownership stdin 或 managed 网络策略应用到独立 CLI；完整 CLI 回归由后票补齐。 |

技术依据为 [design Decisions 5–7](../design.md#decisions)：managed 机器契约、readiness、所有权管道及有界善后。Decision 8 仅在本票所需的真实初始 reload 快照上适用；持续 target/reload 状态报告属于任务 1.3，不以永久 pending 或伪造 active 代替。

## Scope and implementation handoff

代码路径均相对于仓库根；下列现状只用于定位，不取代 specs/design，也不预设新内部类型名。

| 当前表面 | 本票需要的切换 |
|---|---|
| `cmd/root.go`：Cobra 参数、导出分支和服务器构造 | 在 managed 预览路径建立所有权监视，再进行目标 I/O；隐藏 `--managed <generation>`，拒绝不适用的导出组合，禁止 Go 自行请求浏览器，保留独立 CLI 参数行为。 |
| `internal/server.go`：`Serve` 在 `http.Serve` 前输出独立 JSON；`Stop` 仅关 listener | managed 路径建立可用 HTTP 服务后一次发布 ready，返回完整 URL；保留独立 `--json`，建立取消与有界资源关闭路径。 |
| `internal/listener.go`：当前绑定 `:port` 并实现默认回退/严格端口 | managed 显式绑定 `127.0.0.1:0` 并使用真实端口；不把回环绑定或 ephemeral 策略全局套到独立 CLI。 |
| `internal/target.go`、现有初始文章 URL 路径 | 复用目标/文章规则和 URL escaping，managed 启动确认实际目标可访问，不丢单文件路径，不重构渲染或目录发现。 |
| `internal/hotreload/hotreload.go`：当前 Stop 只发结束信号 | 为服务取消提供 watcher 结束等待及 WebSocket 释放，保留已有预算/忽略/发现政策；取得真实初始 reload 快照，不冒充已完成任务 1.3。 |
| `internal/server.go` 的惰性 PDF 资源与 `internal/pdf.go` 的 Close | 正常停止仅释放已经初始化的资源，不能因停止新建 generator；不宣称 EOF 能可靠清扫任意 Chrome 后代。 |
| `docs/ARCHITECTURE.md` | 随验证结果更新 Go 启动、managed 协议、独立 CLI 区别与所有权退出限制。 |

协议遵循已批准 design：managed stdout 仅含序列化的 UTF-8 v1 NDJSON，每条带 generation，ready 含完整 url 和真实 reload 快照，致命失败发可分类 fatal 后退出，日志走 stderr。`/__gogrip/ready` 仅在 managed 路径提供 204 与 `X-GoGrip-Generation`，不访问文件系统或重新渲染正文。本票不实现 Swift 的 64 KiB 解码/诊断缓冲和 15 秒启动期限。

所有权监视必须先于目标解析、卷读取、watcher 和 listener；owner EOF/读错误取消运行，并由独立 watchdog 在 2 秒善后期限后强制退出，不能等阻塞启动返回。正常关闭含有界 HTTP 关闭、watcher 结束等待、WebSocket 关闭和已初始化 PDF 资源释放。允许强制退出绕过 defer；不对不可中断内核 I/O 承诺精确硬截止。Swift 的 CLOEXEC 临界区、4 秒 SIGKILL 兜底及 App 退出协调属于后票。

明确不做：Swift/AppKit/Finder 修改、批次或历史功能、持续目标/watch 故障状态、轮询/重连/自动重启、PID 或进程名清扫、helper/XPC/launchd、渲染器重构、打包/CI、正式签名公证或发布。需要修订批准 artifacts 时先停下请求独立审批，不以本票补写新行为保证。

## Acceptance

- [ ] 按 scoped TDD 先复现 startup 失败无有效 ready、含中文/空格单文件定位、真实回环监听及 owner loss 的消费者可见行为，再实现并记录失败前/通过后的对应检查；永久测试使用隔离目标与实际结果，不固定内部类名、字段复制、日志/源文本、转发或资源搬运。
- [ ] 实际构建 Go 工具，以 argv 和专用 stdin Pipe 启动 `--managed <generation>`，目录传 `-r`、目标前传 `--`；递归目录、所选单文件及可访问空目录均能取得实际 URL，GET 验证对应内容/空状态，中文/空格文件不会定位到父目录或另一文章。
- [ ] 从真实 listener/socket 观察 managed 绑定 `127.0.0.1` 和 OS 分配端口；反馈 URL 与该端口对应，HEAD 启动确认得到 204、匹配代次及无正文，不把 localhost 文本或仅创建进程当作监听/可用性证明。
- [ ] 目标访问或服务启动失败时有可获得的 fatal 原因并退出，不发布成功 ready 或留下 listener/watcher；不适用的导出组合不读取/导出目标。managed stdout 不混入独立 JSON/普通日志，初始 reload 是实际状态，不用永久 pending、假 active 或固定端口掩盖未实现能力。
- [ ] ready 前 owner loss 和运行中关闭 writer 均导致对应 Go 子进程退出；回归包含 ownership 读错误及受控阻塞启动工作，证明取消/2 秒 watchdog 独立于主启动完成。真实 Go smoke 记录 PID、实际 URL/端口与释放结果，不能由 fake shell port、`isRunning` 或正常退出回调代替。
- [ ] 实际强制终止临时拥有者，观察它启动的 managed Go 结束且端口不再提供内容，不依赖再次启动拥有者清理；该 smoke 不冒充尚未实现的生产 Foundation launch、App SIGKILL 或 TCC 证据，也不新增测试专用 CLI flags。
- [ ] 正常关闭能够结束并等待 watcher、关闭 WebSocket 和已经初始化的 PDF 资源，未初始化 PDF 时不创建 generator；阻塞善后由独立 watchdog 收尾，不把信号发出或仅关闭 listener 当作完整停止。
- [ ] 同时启动另一 managed 服务及独立 CLI，关闭一个 managed writer 后其余服务仍可提供各自内容；独立 CLI 的 stdin 关闭或无 App 不会触发 managed 退出，既有 `--json` 和端口策略不被本票改坏，不借此宣称任务 1.4 已全量完成。
- [ ] 本票对应行为回归和 `go test ./...` 通过，真实 binary/HTTP/owner smoke 有可复查结果，`docs/ARCHITECTURE.md` 与已观察事实同步，未验证的原生/平台环境明确标出；通过独立 Standards/Spec 审阅后才允许关闭票及勾选已完整满足的 1.1、1.2，不检查其他任务。

## Change-wide coverage reference

完整任务与规格覆盖地图见 [coverage-plan.md](../coverage-plan.md)。该文件区分已发布票与未发布的未来切片，不是本票的实施或验收清单。

本票实施范围仅为 **1.1、1.2**，以上范围及验收条件不变；其他任务须逐票发布、审阅及获准实施。
