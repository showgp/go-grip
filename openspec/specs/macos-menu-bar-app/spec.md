# macOS Menu Bar App Specification

## Purpose

提供日常可用的原生菜单栏宿主，让用户在没有终端和 Dock 窗口的情况下查看与管理预览、重新打开最近目标，并理解失败原因。此能力定义紧凑面板、手动打开、最近记录、可选登录启动与中英文界面，不扩展为高级配置或独立管理窗口。

## Requirements

### Requirement: Persistent native menu bar host

应用 SHALL 作为常驻菜单栏宿主运行，不显示 Dock 图标，也不为预览打开终端窗口。主界面 SHALL 是紧凑的原生菜单栏弹出面板，不另设独立管理主窗口；关闭面板 SHALL 不退出应用或停止预览。正常对话框及手动选择目标的界面不属于独立管理主窗口。

#### Scenario: Open and dismiss the panel
- **WHEN** 用户从菜单栏打开面板，查看后关闭面板
- **THEN** 菜单栏宿主保持运行，已有预览继续运行，不出现 Dock 图标或终端窗口

### Requirement: Visible sessions and explicit controls

面板 SHALL 展示运行会话及其目标路径和实际状态，并提供明确的重新打开浏览器、复制实际预览 URL、停止单个会话、停止全部及退出入口。运行会话 SHALL 位于最近目标区域之前；动作 SHALL 不依赖用户猜测状态点等隐式交互。所有动作 SHALL 遵循 `macos-preview-sessions` 的复用、停止和宿主退出规则，不提供高级启动参数配置界面。

#### Scenario: Distinguish targets with the same name
- **WHEN** 不同目录下的同名目标分别拥有会话
- **THEN** 面板显示足以区分目标的路径及各自状态，用户能明确选择对应目标的打开或停止操作

#### Scenario: Copy an active preview address
- **WHEN** 用户对运行会话执行复制 URL
- **THEN** 剪贴板获得该会话实际可用的预览地址，而不是默认端口拼接或其他目标的地址

#### Scenario: Exit through the panel
- **WHEN** 用户执行面板的退出操作
- **THEN** 应用按会话生命周期要求停止其预览服务并退出，不把仅关闭面板当作退出

### Requirement: Manual opening uses the same target contract

面板 SHALL 提供手动选择目录或 Markdown 文件的入口。手动打开及最近目标重开 SHALL 使用与 Finder 服务相同的目标类型、规范化身份、会话复用和批量规则，不形成只有某个入口才复用会话的行为差异。

#### Scenario: Manually reopen a Finder-started target
- **WHEN** Finder 服务已为目标启动会话，用户通过手动选择或最近记录再次打开该目标
- **THEN** 应用复用已有会话并请求浏览器打开，不重复启动服务

#### Scenario: Choose a new target manually
- **WHEN** 用户通过面板选择有效目录或 Markdown 文件
- **THEN** 应用按对应的目录或单文件契约打开预览，无需用户执行命令或填写启动参数

### Requirement: Bounded persistent recent targets

应用 SHALL 跨启动保留最近打开的至多 20 个不同目标，按最近使用次序提供重开入口。记录 SHALL 使用与会话一致的规范化目标身份，重复打开同一目标不产生重复历史项目，并保留用户选择的路径供显示。持久记录 SHALL 只包含目标位置及必要显示信息，不保存 Markdown 正文；恢复历史 SHALL 不恢复运行状态或自动启动预览。

#### Scenario: Open more than twenty distinct targets
- **WHEN** 用户依次打开 21 个不同目标
- **THEN** 最近记录保留最近使用的 20 个目标，最早且未再使用的目标不在列表中

#### Scenario: Reopen an older recent target
- **WHEN** 用户重新打开一个已经存在于最近记录中的目标，包括通过其符号链接打开
- **THEN** 该目标更新为最近使用，不新增同一规范化目标的重复历史条目

#### Scenario: Restart after using previews
- **WHEN** 用户退出并重新启动 App
- **THEN** 最近目标仍可查看和手动重开，记录不含正文，也不把上次运行的目标显示为当前运行会话或自动打开它们

### Requirement: Clearing history does not stop sessions

面板 SHALL 提供清空最近记录的操作。清空 SHALL 同时作用于当前及之后启动时的最近记录，但不停止服务、不清除运行会话的管理入口，也不影响用户仍在使用的预览。

#### Scenario: Clear history with an active preview
- **WHEN** 用户在有运行会话时清空最近记录
- **THEN** 最近列表被清空，运行会话仍可打开、复制 URL 或停止，预览继续可用

#### Scenario: Restart after clearing history
- **WHEN** 用户清空记录后没有再打开目标，然后重新启动 App
- **THEN** 最近列表保持为空，不从旧的持久记录重新出现已清除项目

### Requirement: Opt-in launch at login

应用 SHALL 默认关闭登录启动，并提供由用户主动开启或关闭的选项，不在安装、首次启动或服务调用时自行开启。应用 SHALL 不把登录启动作为 Finder 服务可用的前提。

#### Scenario: First use without login startup
- **WHEN** 用户首次使用应用且从未主动开启登录启动
- **THEN** 登录启动保持关闭，用户仍可通过 Finder 服务启动宿主及预览

#### Scenario: Change the login-start preference
- **WHEN** 用户主动开启或关闭登录启动
- **THEN** 应用按该选择设置登录启动；若系统拒绝操作，则明确报告失败，不把未生效的设置显示为成功

### Requirement: System-language localization

应用 SHALL 支持简体中文与英文，跟随系统选择的语言；不支持的语言 SHALL 回退英文。本地化 SHALL 覆盖菜单栏面板、错误、首次引导及 Services 命令名称，不只覆盖主界面；界面 SHALL 不提供本次范围之外的高级配置。

#### Scenario: Use Simplified Chinese
- **WHEN** 系统为应用选择简体中文
- **THEN** 面板、错误、引导与服务名称均使用对应的简体中文界面内容

#### Scenario: Use English or an unsupported language
- **WHEN** 系统为应用选择英文或不受支持的语言
- **THEN** 相应界面内容使用英文，包括服务名称、错误及首次引导，而不是部分内容硬编码为中文

### Requirement: Action failures remain understandable and recoverable

用户打开目标、启动预览或重新打开浏览器失败时，应用 SHALL 显示明确的原生错误提示，包含目标及可获得的原因，并在面板保留可查看的失败信息；批量错误 SHALL 遵循 Finder 服务规格的一次汇总规则，不依赖系统通知授权才让用户知道失败。失败信息 SHALL 与真实会话状态一致，不将浏览器打开失败误报为服务退出，也不将服务退出显示为仍运行。

#### Scenario: Open a stale recent target
- **WHEN** 用户从最近记录重开一个已不存在或不可访问的目标
- **THEN** 应用显示该目标的打开错误，面板保留可查看信息，不将该操作标记为成功运行

#### Scenario: Recover from browser opening failure
- **WHEN** 一个可用会话未能由系统打开浏览器
- **THEN** 应用显示原生错误并保留真实运行状态及失败详情，用户仍能重新打开浏览器或复制 URL

#### Scenario: A service exits unexpectedly
- **WHEN** 运行中的服务意外退出
- **THEN** 面板不再显示其为运行中，用户能查看退出原因并从最近目标主动重开，不产生自动重启
