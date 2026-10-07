# macOS Preview Sessions Specification

## Purpose

定义 macOS 宿主拥有的本地预览会话，使用户能够可靠地启动、复用、打开和停止目标，而不是仅凭进程已创建或猜测的端口判断成功。会话契约覆盖目标身份、实际预览地址、仅本机监听、正常与异常退出，以及访问和热重载降级的可见反馈。

## Requirements

### Requirement: One session per normalized target

应用 SHALL 规范化目标路径并解析符号链接，对所有打开入口使用一致的目标身份。相同目标 SHALL 最多拥有一个运行中的预览会话；重复或同时到达的请求 SHALL 不为同一目标启动重复服务。应用 SHALL 保留用户选择的路径供显示，不把身份归一误用为丢失用户选择信息。

#### Scenario: Reopen a target through a symbolic link
- **WHEN** 一个目录已有运行会话，用户通过指向该目录的符号链接再次打开
- **THEN** 应用复用该会话并请求浏览器打开已有预览，不启动第二个服务

#### Scenario: Receive concurrent requests for the same target
- **WHEN** 同一目标的多个打开请求在首次启动尚未完成时到达
- **THEN** 应用协调这些请求，同一目标不产生多个预览服务；成功后这些请求使用同一会话

### Requirement: Containment does not merge target identities

应用 SHALL 将不同的父目录、子目录及文件视为独立目标，不因父目录递归预览已经覆盖子目标而吞并打开请求。只有规范化后相同的目标 SHALL 复用会话。

#### Scenario: Open a child directory and a file under an active parent
- **WHEN** 父目录已有会话，用户再打开其中的子目录和 Markdown 文件
- **THEN** 三个不同目标分别拥有会话，子目录以自身为浏览根，文件以单文件模式预览

### Requirement: Directory and single-file preview modes

应用启动的目录会话 SHALL 默认递归浏览 Markdown 并启用热重载；文件会话 SHALL 直接预览所选 Markdown，不改为父目录浏览。可访问但没有 Markdown 的目录 SHALL 成功建立会话并显示明确的空状态。上述行为 SHALL 沿用现有渲染体验，不要求重设计浏览器界面；热重载覆盖限制由降级要求定义。

#### Scenario: Preview nested documents
- **WHEN** 用户打开包含子目录及嵌套 Markdown 的可访问目录
- **THEN** 浏览器进入该目录的递归预览，嵌套 Markdown 可按现有目录浏览方式访问，热重载默认启用

#### Scenario: Preview a selected file
- **WHEN** 用户打开一个可访问 Markdown 文件
- **THEN** 浏览器直接进入该文件的单文件预览，不以父目录的其他文章作为打开结果

#### Scenario: Keep an empty directory session
- **WHEN** 用户打开可访问但没有 Markdown 的目录
- **THEN** 应用保留运行会话并打开浏览器的空状态页面，而不是把本次操作视为启动失败

#### Scenario: Add Markdown to a watched empty directory
- **WHEN** 一个空目录会话仍运行，且目录受有效热重载监视覆盖，用户在其中新增 Markdown
- **THEN** 浏览器热重载后能够显示该文档，用户无需先停止并重建会话

### Requirement: Verified startup and actual preview URL

应用 SHALL 仅在成功建立可用服务并取得其实际预览地址后，将会话显示为运行中并请求默认浏览器打开。地址 SHALL 对应实际服务端口及所选目标，不能只拼接默认端口或丢失单文件路径。启动反馈缺失、无效或服务在成功启动前退出时，应用 SHALL 报告启动失败，不猜测端口，不显示虚假的运行会话，也不遗留本次失败启动的服务。

#### Scenario: A nondefault port is used
- **WHEN** 默认端口不可用，而服务成功使用其他端口提供目标预览
- **THEN** 应用记录并打开实际预览 URL，不打开默认端口上的其他服务

#### Scenario: Open a file with spaces and non-ASCII characters
- **WHEN** 用户打开路径包含空格及中文的有效 Markdown 文件
- **THEN** 应用使用能够正确定位该文件的实际 URL，浏览器显示该文件而不是其他目标或错误页面

#### Scenario: Startup ends without valid readiness information
- **WHEN** 服务未成功提供可用预览地址就退出，或应用判断启动反馈缺失或无效
- **THEN** 应用报告启动失败，不以固定端口继续打开浏览器，不留下假的运行状态或本次未完成启动的服务

### Requirement: Loopback-only application previews

App 拥有的预览服务 SHALL 实际仅监听本机回环地址，不能仅在展示 URL 中使用 `localhost` 而绑定所有网络接口。应用 SHALL 不提供这些会话的局域网共享入口。

#### Scenario: Reach a preview from the same machine
- **WHEN** App 成功启动一个预览会话
- **THEN** 本机浏览器能通过该会话的回环 URL 访问预览，服务不监听通配或非回环地址

#### Scenario: Attempt to use a nonloopback interface
- **WHEN** 其他设备尝试通过宿主的局域网地址连接 App 预览端口
- **THEN** 该会话不通过该网络接口提供服务，不能因页面展示 localhost 就把通配监听视为符合要求

### Requirement: Browser opening is separate from service lifetime

成功建立服务后，应用 SHALL 请求默认浏览器打开该会话的实际 URL。重复打开运行目标 SHALL 复用服务并再次请求浏览器打开，不跟踪或承诺聚焦同一标签页。若系统未能打开浏览器，应用 SHALL 明确报错但保留可用服务，允许用户再次打开或复制 URL；关闭浏览器或没有页面访问 SHALL 不自动停止服务。

#### Scenario: Browser opening fails after service startup
- **WHEN** 服务已成功运行，但系统报告浏览器打开失败
- **THEN** 会话仍显示实际运行状态，同时显示错误，用户可以再次打开浏览器或复制同一会话 URL，而不是重复启动服务

#### Scenario: Close the browser preview
- **WHEN** 用户关闭该会话的浏览器标签页或浏览器，且未执行停止或退出 App
- **THEN** 预览服务继续运行，不因标签页关闭或闲置而回收

### Requirement: Explicit stopping confirms service termination

应用 SHALL 支持停止单个会话和停止全部归属本 App 的会话。停止操作完成时，相应服务 SHALL 已退出并释放监听及监视资源；应用 SHALL 不仅删除界面条目就将仍运行的服务视为已停止。停止单个会话 SHALL 不影响其他目标；与 App 无关的独立 CLI 服务 SHALL 不属于这些停止操作的范围。

#### Scenario: Stop one of several sessions
- **WHEN** 多个会话运行，用户停止其中一个
- **THEN** 对应服务退出并释放资源，界面不再将其显示为运行中，其余会话继续工作

#### Scenario: Stop all application sessions
- **WHEN** 用户执行停止全部
- **THEN** 本 App 拥有的全部预览服务退出，独立启动且不归属本 App 的 CLI 服务不受影响

### Requirement: Services end with their owning application

应用 SHALL 为其预览服务建立有效归属，使宿主正常退出、崩溃或被强制退出后，这些服务也结束，不继续成为无人管理的后台预览服务。该要求 SHALL 覆盖退出时仍处于启动过程的服务，不以已经显示在运行列表中为清理前提。

#### Scenario: Quit the application normally
- **WHEN** 有运行会话，用户退出 App
- **THEN** 宿主的预览服务全部退出，原预览端口不再由这些服务提供内容

#### Scenario: Force termination of the host
- **WHEN** 宿主崩溃或被强制退出，而其预览服务仍运行
- **THEN** 这些服务也结束并释放资源，不等待用户再次启动 App 才清理

#### Scenario: Host exits during startup
- **WHEN** 宿主退出或被强制终止时，某个归属它的预览服务尚未完成启动
- **THEN** 该服务不因尚未进入运行列表而遗留在后台

### Requirement: No automatic session restoration or restart

应用 SHALL 不在自身重新启动时自动恢复预览服务或打开旧预览。预览服务意外退出时，应用 SHALL 更新状态并让退出原因可查看，不把退出的服务继续显示为运行中，也不自动重启；用户可以从最近目标主动重新打开。用户明确执行停止 SHALL 不被误报为意外退出。

#### Scenario: Relaunch with recent targets
- **WHEN** 用户重新启动 App，历史中包含曾运行的目标
- **THEN** App 只恢复历史显示，不自动启动服务、访问这些目标或打开浏览器

#### Scenario: A preview process exits unexpectedly
- **WHEN** 一个运行中的预览服务意外退出
- **THEN** App 不再显示其为运行中，退出原因可查看，不自动重启；用户仍能主动重开目标

### Requirement: Unavailable paths are not empty directories

会话 SHALL 继续以建立时确定的路径为目标，不自动追踪目标移动或后台重连。发现目标因删除、移动、权限变化或卷断开而不可访问时，应用 SHALL 明确报告访问失败，并保留查看及停止现有会话的入口；只有确认可访问且未发现 Markdown 时，才 SHALL 显示无 Markdown 的空状态。

#### Scenario: A target is moved or deleted
- **WHEN** 会话访问原目标路径时发现目标已移动或删除
- **THEN** 应用报告原路径不可访问，不自动改为新位置，不将该情况显示为普通空目录，用户仍能查看及停止会话

#### Scenario: A mounted volume becomes unavailable
- **WHEN** 会话发现目标所在卷已经断开或不可访问
- **THEN** 应用明确显示访问失败，不后台重连或承诺自动恢复监视，用户可以停止后按需重新打开

### Requirement: Visible hot-reload degradation

当已知监视注册失败、预算不足或覆盖不完整时，应用 SHALL 在会话中显示热重载降级信息，保留可用的预览服务及手动刷新能力，不将仅部分受监视的会话伪装为完整自动刷新。应用 SHALL 不增加轮询兜底，也不承诺网络卷上的文件事件与自动刷新始终可靠。

#### Scenario: Watch coverage is incomplete
- **WHEN** 服务可提供预览，但已知资源预算或监视错误使部分目录无法受监视
- **THEN** App 显示该会话的热重载降级状态，用户仍能打开预览并通过浏览器手动刷新查看可访问内容

#### Scenario: Use a network-volume preview
- **WHEN** 用户打开可访问网络卷上的目录
- **THEN** 预览不因网络卷身份而被拒绝，但不会宣称自动刷新始终可靠；已知监视失败仍按降级规则显示，不自动启用轮询

### Requirement: Independent CLI operation remains available

为宿主完善集成后，Go 工具 SHALL 仍支持不依赖 App 的独立目录及单文件预览。App 的会话归属和退出行为 SHALL 不被应用于并未由 App 拥有的独立 CLI 运行。

#### Scenario: Launch the renderer without the application
- **WHEN** 用户未运行 App，独立使用 CLI 打开目录或 Markdown 文件
- **THEN** CLI 仍能提供对应预览，不要求菜单栏宿主存在，也不因宿主随后退出而被终止
