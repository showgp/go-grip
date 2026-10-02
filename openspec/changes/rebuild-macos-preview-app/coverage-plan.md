# Whole-change coverage plan

本文件是整个变更的覆盖地图，不是实施票，也不是行为规格；任务进度仅由 [tasks.md](tasks.md) 记录。本次拆分不修改已批准的 [proposal.md](proposal.md)、四份 spec 或 [design.md](design.md)，不新增行为或实施授权。

当前唯一已发布票是 [01 — Go managed 预览启动与所有权生命周期](tickets/01-managed-preview-lifecycle.md)，仅对应任务 **1.1、1.2**，状态为 **open，待审阅、未实施**。以下后续切片只用于规划，不代表票据已经创建或获准实施。

以下是本轮的完整覆盖规划，不是批准或发布后续票。**只有 01 已发布，状态仍为 open；02–11 均为未发布规划。** 后续每次出版重读批准 artifacts、现有票与实际 checkbox 标签，再检查范围及依赖。规格/任务可能需要多票共同贡献；所有当前已发布票完成也不等于其余任务或整项 requirement 已完成。

## Planned vertical slices and genuine blockers

| Slice | 可验证的交付 | OpenSpec tasks | 真正前置 | 发布状态 |
|---|---|---|---|---|
| 01-managed-preview-lifecycle | 真实 Go 预览启动、回环 URL 与 owner loss 后退出 | 1.1、1.2 | None | 已发布，open |
| 02-managed-status-and-cli | 实际目标/监视状态变化及降级仍可访问，独立 CLI 完整回归 | 1.3、1.4 | 01：managed 事件与资源生命周期 | 未发布规划 |
| 03-owned-process-launch | Foundation 从真实 Go 反馈到可用 URL，失败清理与所有权/SIGKILL 集成 | 2.1、2.5；2.4 的进程层贡献 | 01、02：完整 Go 机器契约 | 未发布规划 |
| 04-target-session-control | 规范化目标单飞/复用、代次隔离与真实单个/全部停止 | 2.2、2.4 的协调器及完整行为 | 03：真实进程适配 | 未发布规划 |
| 05-native-finder-preview | 新宿主冷启动接收 Finder/手动批次，实际浏览器和运行控制，切换旧链路 | 2.3、3.1、3.2、3.4 | 04：会话事实源与停止 | 未发布规划 |
| 06-native-target-access | native→Go 实际授权主体、拒绝与本地/外接/网络卷访问 gate | 3.3 | 05：可运行原生入口 | 未发布规划 |
| 07-recent-target-panel | 最近 20 项、完整状态面板、重开与清历史不停止 | 4.1、4.2、4.3 | 06：批准的早期权限 gate；其入口来自 05 | 未发布规划 |
| 08-localized-service-guidance | 可再次查看的 Services 帮助与完整中英文/英文回退体验 | 5.1、5.2 | 07：待本地化的完整面板 | 未发布规划 |
| 09-opt-in-login-start | 用户可选登录项、真实系统状态与失败反馈 | 5.3 | 07：实际面板和宿主；不依赖 08 | 未发布规划 |
| 10-universal-candidate | 自包含 universal archive/App/DMG、候选 CI artifact 和真实使用文档 | 6.1、6.2、6.3、6.4 | 08、09：完整功能与资源；开发运行包来自 05 | 未发布规划 |
| 11-candidate-functional-acceptance | 候选跨组件真实 Finder/浏览器/所有权/架构验收，缺环境明确阻塞 | 7.1、7.2、7.3 | 10：同一可运行候选产物；实际平台/卷环境 | 未发布规划，仅集成验证 |

## Every OpenSpec checkbox

| 实际 checkbox 标签 | 贡献 slice | 发布状态 |
|---|---|---|
| 1.1 | 01 | 已发布，未实施 |
| 1.2 | 01 | 已发布，未实施 |
| 1.3 | 02 | 未发布规划 |
| 1.4 | 02 | 未发布规划 |
| 2.1 | 03 | 未发布规划 |
| 2.2 | 04 | 未发布规划 |
| 2.3 | 05 | 未发布规划 |
| 2.4 | 03（进程层）、04（完整协调器行为） | 未发布规划；两项贡献均需验证 |
| 2.5 | 03 | 未发布规划 |
| 3.1 | 05 | 未发布规划 |
| 3.2 | 05 | 未发布规划 |
| 3.3 | 06 | 未发布规划 |
| 3.4 | 05 | 未发布规划 |
| 4.1 | 07 | 未发布规划 |
| 4.2 | 07 | 未发布规划 |
| 4.3 | 07 | 未发布规划 |
| 5.1 | 08 | 未发布规划 |
| 5.2 | 08 | 未发布规划 |
| 5.3 | 09 | 未发布规划 |
| 6.1 | 10 | 未发布规划 |
| 6.2 | 10 | 未发布规划 |
| 6.3 | 10 | 未发布规划 |
| 6.4 | 10 | 未发布规划 |
| 7.1 | 11 | 未发布规划 |
| 7.2 | 11 | 未发布规划 |
| 7.3 | 11 | 未发布规划；实际支持环境缺失不能结案 |

## Every requirement and scenario

下表逐一列出四份 spec 的全部 requirement 和 scenario 原名；每行列出的贡献者共同完成该行覆盖，01 的发布不代表整个 App requirement 已完成。未写明 01 的行全部仍是未发布规划；出现 01 的行仅其 Go 部分已发布，其余仍未发布。最后的 11 是候选级集成证明，不替代组件票的 tests/smoke/docs。

### macos-finder-service

行为源：[spec.md](specs/macos-finder-service/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Finder Services entry | `Open a selected folder from Finder`；`Invoke the service while the host is not running` | 05、11：规划 |
| Supported open targets | `Receive a Markdown file`；`Receive a directory outside the home folder`；`Receive an unsupported file` | 04（目标）、05（入口）、06（真实卷访问）、11：规划 |
| Batch deduplication and quantity confirmation | `Open exactly five distinct targets`；`Confirm six distinct targets`；`Cancel a large batch`；`Count aliases only once` | 04（身份）、05（批次）、11：规划 |
| Partial batch failures | `Some targets fail`；`All targets succeed` | 05、11：规划 |
| First-use service guidance | `Read first-use instructions`；`Revisit guidance for a disabled service` | 08：规划 |
| Minimum necessary file authorization | `Access is denied`；`Open an ordinary accessible target` | 05（原生错误）、06（权限 gate）、11：规划 |

### macos-preview-sessions

行为源：[spec.md](specs/macos-preview-sessions/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| One session per normalized target | `Reopen a target through a symbolic link`；`Receive concurrent requests for the same target` | 04、05（跨入口）、11：规划 |
| Containment does not merge target identities | `Open a child directory and a file under an active parent` | 04、11：规划 |
| Directory and single-file preview modes | `Preview nested documents`；`Preview a selected file`；`Keep an empty directory session`；`Add Markdown to a watched empty directory` | 01（启动/内容）已发布；02（新增文档/watch）、05（原生打开）、11 规划 |
| Verified startup and actual preview URL | `A nondefault port is used`；`Open a file with spaces and non-ASCII characters`；`Startup ends without valid readiness information` | 01（服务/URL）已发布；03（消费/失败清理）、05（原生结果）、11 规划 |
| Loopback-only application previews | `Reach a preview from the same machine`；`Attempt to use a nonloopback interface` | 01（真实回环 bind）已发布；05（App 启动路径）、11（候选实际接口）规划 |
| Browser opening is separate from service lifetime | `Browser opening fails after service startup`；`Close the browser preview` | 05、07（恢复入口）、11：规划 |
| Explicit stopping confirms service termination | `Stop one of several sessions`；`Stop all application sessions` | 01（Go 退出）已发布；03（适配器）、04（单个/全部）、05（原生动作）、11 规划 |
| Services end with their owning application | `Quit the application normally`；`Force termination of the host`；`Host exits during startup` | 01（Go owner loss）已发布；03（Foundation 所有权）、04（starting 停止）、05（App quit）、11 规划 |
| No automatic session restoration or restart | `Relaunch with recent targets`；`A preview process exits unexpectedly` | 04（退出/不重启）、07（历史/重启）、11：规划 |
| Unavailable paths are not empty directories | `A target is moved or deleted`；`A mounted volume becomes unavailable` | 02（Go 状态）、04（状态事实）、06（实际卷）、07（面板）、11：规划 |
| Visible hot-reload degradation | `Watch coverage is incomplete`；`Use a network-volume preview` | 02（状态/可用内容）、06（网络卷）、07（显示/刷新）、11：规划 |
| Independent CLI operation remains available | `Launch the renderer without the application` | 01（所有权隔离）已发布；02（完整 CLI）、03/04（owned 操作隔离）、11 规划 |

### macos-menu-bar-app

行为源：[spec.md](specs/macos-menu-bar-app/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Persistent native menu bar host | `Open and dismiss the panel` | 05、11：规划 |
| Visible sessions and explicit controls | `Distinguish targets with the same name`；`Copy an active preview address`；`Exit through the panel` | 04（状态/停止）、05（基础控制/退出）、07（完整面板）、11：规划 |
| Manual opening uses the same target contract | `Manually reopen a Finder-started target`；`Choose a new target manually` | 04（统一身份）、05（手动入口）、07（最近重开）、11：规划 |
| Bounded persistent recent targets | `Open more than twenty distinct targets`；`Reopen an older recent target`；`Restart after using previews` | 07、11：规划 |
| Clearing history does not stop sessions | `Clear history with an active preview`；`Restart after clearing history` | 07、11：规划 |
| Opt-in launch at login | `First use without login startup`；`Change the login-start preference` | 05（无登录项冷启动）、09（系统登录项）：规划 |
| System-language localization | `Use Simplified Chinese`；`Use English or an unsupported language` | 08：规划 |
| Action failures remain understandable and recoverable | `Open a stale recent target`；`Recover from browser opening failure`；`A service exits unexpectedly` | 04（真实退出）、05（原生报告）、07（历史/保留失败/恢复）、11：规划 |

### macos-app-packaging

行为源：[spec.md](specs/macos-app-packaging/spec.md)。

| Requirement | 全部 scenarios | 贡献 slice 与发布状态 |
|---|---|---|
| Universal self-contained application | `Use the application without development tools`；`Use the same candidate on supported architectures` | 05（确定内置工具路径）、10（universal 候选）、11（实际运行矩阵）：规划 |
| Candidate DMG and Services-only Finder integration | `Produce an installable candidate`；`Use the new Finder integration` | 05（去旧扩展）、08（资源/引导）、10（DMG）、11：规划 |
| Functional acceptance exercises actual user paths | `Verify a real Finder-to-browser launch`；`Verify reuse and explicit stopping`；`Verify abnormal host termination` | 11：规划；此前各组件票仍各自完成实际 smoke |
| Explicitly deferred formal distribution | `Present an unsigned candidate for functional review`；`Complete this implementation stage` | 10（候选标识/说明）、11（交付边界）：规划；不安排正式分发工作 |

本轮覆盖规划含全部 26 个 OpenSpec checkbox、30 个 requirement、68 个 scenario；没有未分配的范围。发布后续票或实施前仍须重查实际 artifacts 与此前票据，不能把这份未来规划当作额外授权。
