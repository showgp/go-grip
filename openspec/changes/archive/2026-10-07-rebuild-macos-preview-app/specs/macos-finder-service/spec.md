# macOS Finder Service Spec Delta

## Purpose

让用户从 Finder 的已选目录或 Markdown 文件直接进入 GoGrip 本地预览，不依赖终端或登录启动。此能力定义 Services 入口、目标接收、批量操作、首次引导及权限边界；系统控制的菜单位置与启用状态不属于 App 可自行保证的范围。

## ADDED Requirements

### Requirement: Finder Services entry

应用 SHALL 提供“用 GoGrip 打开”的 macOS Services 命令，让用户从 Finder 对已选目标执行打开操作。宿主未运行时，服务请求 SHALL 能启动宿主并处理所选目标，不以登录启动已开启为前提。应用 SHALL 接受系统决定的服务菜单位置和用户启用状态，不要求右键菜单第一层或目录空白处入口。

#### Scenario: Open a selected folder from Finder
- **WHEN** 用户已安装应用并在系统允许显示该服务的情况下，对 Finder 中选中的可访问目录调用“用 GoGrip 打开”
- **THEN** 应用接收该目录并按预览会话规格打开预览，不要求用户执行终端命令

#### Scenario: Invoke the service while the host is not running
- **WHEN** 登录启动未开启、宿主未运行，用户对有效目标调用已启用的服务
- **THEN** 系统启动宿主，应用处理本次目标并请求浏览器打开预览，无需用户先手动启动宿主

### Requirement: Supported open targets

应用 SHALL 接受可访问目录和扩展名为 `.md`（大小写不敏感）的 Markdown 文件。目标范围 SHALL 包括主目录之外的本地目录、外接磁盘和已挂载网络卷，但仍受文件系统及 macOS 授权限制。应用 SHALL 拒绝不支持的文件类型，不把它们隐式转换为所在目录；目标的目录或单文件预览行为由 `macos-preview-sessions` 定义。

#### Scenario: Receive a Markdown file
- **WHEN** 服务收到可访问的 `说明.MD` 文件
- **THEN** 应用将该文件作为单文件目标处理，而不是打开其父目录

#### Scenario: Receive a directory outside the home folder
- **WHEN** 服务收到主目录之外、本地磁盘、外接磁盘或已挂载网络卷上的可访问目录
- **THEN** 应用接受该目录，不因其不在用户主目录内而拒绝

#### Scenario: Receive an unsupported file
- **WHEN** 服务收到普通图片等非 Markdown 文件
- **THEN** 应用不为该文件启动预览，也不改为预览父目录，并将不支持的目标纳入本次操作的错误反馈

### Requirement: Batch deduplication and quantity confirmation

应用 SHALL 支持一次请求中的多个目标，使用 `macos-preview-sessions` 的规范化身份去重。当去重后的不同有效目标数量大于 5 时，应用 SHALL 在打开任何本次目标前请求用户确认并显示目标数量；数量不超过 5 时 SHALL 不显示数量确认。取消确认 SHALL 不启动或重新打开本批目标，也不停止已有会话。

#### Scenario: Open exactly five distinct targets
- **WHEN** 一次请求包含 5 个不同有效目标
- **THEN** 应用直接处理这些目标，不显示数量确认

#### Scenario: Confirm six distinct targets
- **WHEN** 一次请求包含 6 个不同有效目标
- **THEN** 应用先显示包含数量 6 的确认，只有用户确认后才打开本批目标

#### Scenario: Cancel a large batch
- **WHEN** 用户取消超过 5 个不同有效目标的数量确认
- **THEN** 应用不为本批目标启动服务或请求打开浏览器，并保持原有会话不变

#### Scenario: Count aliases only once
- **WHEN** 一次请求包含 6 个路径，但规范化并解析符号链接后只对应 5 个不同有效目标
- **THEN** 应用按 5 个目标处理，不显示数量确认，也不为重复目标重复启动服务

### Requirement: Partial batch failures

应用 SHALL 对已获准处理的批次继续打开可用目标，不因单个目标不支持、不可访问或启动失败而放弃其余目标，不回滚已成功的会话。本次失败及不支持的目标 SHALL 汇总为一次原生提示，包含对应目标和原因，并在面板中保留可查看的失败信息；全部成功时 SHALL 不额外显示成功提示。

#### Scenario: Some targets fail
- **WHEN** 已获准的批次同时包含可用目标、不支持的文件和无法打开的目标
- **THEN** 应用处理可用目标并保留成功会话，用一次提示汇总失败及不支持的项目，而不是逐项弹出错误

#### Scenario: All targets succeed
- **WHEN** 已获准批次的所有目标都成功打开
- **THEN** 应用打开对应预览，不额外弹出整批成功提示

### Requirement: First-use service guidance

应用 SHALL 在首次启动提供简短引导，说明选中目标后的服务操作以及菜单未出现时如何检查系统 Services 设置。该帮助 SHALL 可以在之后再次查看。应用 SHALL 不自动修改用户的服务启用配置，也不得仅凭宿主成功启动宣称 Finder 菜单已可用。

#### Scenario: Read first-use instructions
- **WHEN** 用户首次启动应用
- **THEN** 用户能看到 Finder 操作说明及服务未出现时的系统设置检查步骤

#### Scenario: Revisit guidance for a disabled service
- **WHEN** 用户之后需要检查被系统或用户关闭的服务
- **THEN** 用户能再次查看帮助并按说明检查设置，应用不自行启用该服务

### Requirement: Minimum necessary file authorization

应用 SHALL 按实际目标使用系统允许的文件访问授权，在需要授权或访问被拒绝时说明目标、原因及适用的授权检查路径。应用 SHALL 不绕过系统权限，不默认要求完全磁盘访问、辅助功能或自动化权限；“任意目录” SHALL 限定为用户及系统允许访问的目录。

#### Scenario: Access is denied
- **WHEN** 目标受文件权限或 macOS 隐私限制而无法访问
- **THEN** 应用明确报告访问失败及适用的授权检查方式，不把目标报告为空目录或成功运行，也不尝试绕过权限

#### Scenario: Open an ordinary accessible target
- **WHEN** 用户打开不需要额外授权的可访问目标
- **THEN** 应用直接处理目标，不以授予完全磁盘访问、辅助功能或自动化权限作为统一使用前提
