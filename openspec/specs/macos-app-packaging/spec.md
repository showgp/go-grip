# macOS App Packaging Specification

## Purpose

定义完整 macOS 功能在本阶段的交付载体与验收边界：生成支持 macOS 13+ 两种架构、内置渲染工具的候选 App 和 DMG，实际验证用户操作路径。本阶段不替代之后的正式签名、公证及安装包验证，也不能把可构建或自动测试通过等同于正式发布就绪。

## Requirements

### Requirement: Architecture-split self-contained application
打包流程 SHALL 按 Apple Silicon(arm64)和 Intel(x86_64)架构分别生成候选 App,每个候选 App 内置同一版本、能在对应架构运行的 Go 渲染工具,并支持 macOS 13 及以上。用户使用某一架构的候选 App SHALL 不需要安装 Go、单独安装 CLI、配置 PATH 或保留源码目录;宿主与内置工具 SHALL 能在对应架构上共同提供完整预览功能。macOS 13+ 双架构覆盖 SHALL 由两个架构候选包共同满足,任一架构候选 App 只保证在对应架构运行。

#### Scenario: Use the application without development tools
- **WHEN** 用户运行对应架构打包生成的候选 App,环境中没有 Go 开发工具或额外安装的 go-grip CLI,也未配置其 PATH
- **THEN** App 使用随包提供的工具完成 Finder 或手动打开的目标预览,不要求从源码目录寻找二进制文件

#### Scenario: A candidate runs on its own architecture
- **WHEN** 用户在 Apple Silicon 上运行 arm64 候选 App,或在 Intel Mac 上运行 x86_64 候选 App
- **THEN** 宿主及随包工具在对应架构运行并提供完整预览功能

#### Scenario: Candidates together cover both architectures
- **WHEN** 检查两个候选包
- **THEN** 每个包仅含其架构的宿主与工具切片,并集覆盖 macOS 13+ 的 Apple Silicon 和 Intel

### Requirement: Candidate DMG and Services-only Finder integration

本阶段 SHALL 提供可生成上述候选 App 及可用于安装的候选 DMG 的打包流程:每个架构生成一个候选 DMG,文件名为 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`,各包含对应架构完整 App 与 `/Applications` 快捷方式;打包流程 SHALL NOT 再产出 universal 候选 DMG 或候选 App zip。候选 App SHALL 包含服务注册和所需本地化内容,不包含遗留 Finder Sync 扩展或其专用转发兼容层;使用 Finder 入口 SHALL 不要求用户启用旧扩展。

#### Scenario: Produce an installable candidate
- **WHEN** 执行本阶段的打包流程
- **THEN** 产生 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`,各包含对应架构完整 App 与 `/Applications` 快捷方式,安装后能提供本规格定义的 Services、菜单栏及预览功能

#### Scenario: Use the new Finder integration
- **WHEN** 用户安装候选 App 并按引导检查系统 Services 启用状态
- **THEN** Finder 打开能力由 Services 提供，不依赖随包发布或单独启用遗留 Finder Sync 扩展

### Requirement: Functional acceptance exercises actual user paths

本阶段的功能验收 SHALL 对每个架构的候选 App 实际运行验证:Finder 服务进入真实浏览器预览、目标复用、停止和异常退出清理,不能仅以编译、自动测试或进程成功创建代替。验收 SHALL 明确区分实际观察到的行为与未验证环境,不把源码中存在对应逻辑当作运行证据。此功能验收 SHALL 不被描述为已完成后续正式安装包的签名、公证或干净安装验证。

#### Scenario: Verify a real Finder-to-browser launch
- **WHEN** 对候选 App 进行 Finder 入口验收
- **THEN** 实际选择目标并调用服务，观察宿主处理目标和浏览器显示对应内容，而不是只检查服务声明或构建结果

#### Scenario: Verify reuse and explicit stopping
- **WHEN** 对候选 App 进行会话复用和停止验收
- **THEN** 实际重复打开同一目标并观察没有重复服务，再执行停止并观察对应服务退出及资源释放

#### Scenario: Verify abnormal host termination
- **WHEN** 对候选 App 进行异常退出清理验收
- **THEN** 在有预览服务时实际强制终止宿主，观察其服务也结束，不仅验证正常退出回调

### Requirement: Explicitly deferred formal distribution

本阶段 SHALL 交付完整功能及打包流程，但将 Developer ID 签名、公证、正式安装包验证及公开发布另行安排。候选产物和交付说明 SHALL 明确本阶段不代表正式发布就绪；未签名候选包的系统信任提示 SHALL 不被当作最终产品安装体验。正式分发的签名、公证与干净安装标准 SHALL 不因本阶段暂缓而被取消。

#### Scenario: Present an unsigned candidate for functional review
- **WHEN** 本阶段提供尚未完成正式签名或公证的候选 App 或 DMG
- **THEN** 交付说明标明候选阶段及未完成的正式分发验证，不宣称普通用户安装信任流程已经验收

#### Scenario: Complete this implementation stage
- **WHEN** 完整功能及候选打包流程达到本阶段验收要求
- **THEN** 本阶段可以交付功能验收结果，但不因此自动获得公开发布许可，也不把正式签名、公证和安装包验证标为已完成

### Requirement: Explicitly deferred environment-dependent acceptance

本阶段 SHALL 将 Intel 与 macOS 13 实机运行、已挂载网络卷的实际访问与热重载降级、候选级可卸载卷断卷的验收延期至后续 change：上述场景 SHALL NOT 作为本阶段完成条件，也 SHALL NOT 被描述为验证通过；universal 或双单架构候选构建、架构/部署静态检查或另一环境的结果 SHALL NOT 替代其实际运行证据。相关行为要求及场景 SHALL 保持有效，由后续 change 承接实际运行验收及所需环境、证据与完成标准；延期维度按以下场景逐项记录。

#### Scenario: Close this implementation stage with deferred environments
- **WHEN** 本阶段按调整后范围完成并汇总候选证据
- **THEN** 仅将以下延期维度记录为未验证并注明由后续 change 承接：`Architecture-split self-contained application` 的 `A candidate runs on its own architecture` 与 `Use the application without development tools` 中的 Intel/macOS 13 实机维度；`macos-finder-service` 的 `Supported open targets` / `Receive a directory outside the home folder` 中的网络卷维度；`macos-preview-sessions` 的 `Visible hot-reload degradation` / `Use a network-volume preview` 中的网络卷验收；以及 `Unavailable paths are not empty directories` / `A mounted volume becomes unavailable` 中的候选级可卸载卷断卷验收。这些延期维度不作为本阶段完成条件，不被勾选或宣称已验证；同一场景中已取得证据的其他维度保留其实际结果

#### Scenario: Claim support for a deferred environment
- **WHEN** 交付说明、验收记录或候选材料提及 Intel、macOS 13、已挂载网络卷或可卸载卷断开支持
- **THEN** 材料标注尚未取得实际运行证据并由后续 change 承接，不以静态检查、构建或替代/受控结果顶替

### Requirement: Acceptance evidence for the browser-open failure branch

候选级“系统报告浏览器打开失败”分支（`macos-preview-sessions` 的 `Browser opening fails after service startup`、`macos-menu-bar-app` 的 `Recover from browser opening failure`）SHALL 以现有消费者行为回归的确定性结果作为完成证据；受控辅助观察 SHALL NOT 作为该分支的必需证据，仅在既有回归不足以覆盖且另行批准时使用，并 SHALL 标注其控制边界与产物差异、结束后还原并从最终交付候选复核正常路径。验收记录 SHALL 如实区分消费者行为回归结果与候选级真实系统拒绝的实际观察；本轮候选级真实系统拒绝未触发，SHALL NOT 将消费者回归结果表述为已触发系统级拒绝或已通过系统级拒绝验证。本证据标准仅适用于该失败分支，不替代正常浏览器打开路径的实际观察。

#### Scenario: Accept the browser-open failure branch
- **WHEN** 本阶段判定候选级浏览器打开失败分支的完整证据
- **THEN** 以现有消费者行为回归的确定性结果验收，并如实记录候选级真实系统拒绝是否触发；本轮未触发，不将消费者回归结果表述为已触发系统级拒绝或已通过系统级拒绝验证

#### Scenario: Use a controlled auxiliary observation
- **WHEN** 既有回归不足以覆盖该分支且用户另行批准受控辅助
- **THEN** 受控观察标注所控制边界与产物差异，结束后还原并从最终交付候选复核正常路径，不冒充真实系统拒绝

### Requirement: Per-architecture candidate verification
打包流程 SHALL 对每个候选 DMG 及其中的 App 执行确定性校验:codesign 严格校验通过、Bundle ID 为 `com.showgp.GoGrip`、内置工具标识符为 `com.showgp.GoGrip.go-grip`、签名均为 ad-hoc;包内每个二进制 SHALL 通过其对应架构的 `verify_arch` 校验,SHALL NOT 通过另一架构的 `verify_arch` 校验。

#### Scenario: Verify a candidate before acceptance
- **WHEN** 打包流程完成对两个候选包的校验
- **THEN** arm64 包内宿主与工具均为 arm64 单架构,签名与标识符校验通过;x86_64 包同理

#### Scenario: Reject wrong architecture assignment
- **WHEN** 某候选包内的二进制包含错误架构切片,或同时包含 arm64 与 x86_64
- **THEN** 校验失败,该包不作为候选交付

### Requirement: CI uploads split candidate artifacts
CI release 工作流 SHALL 在候选打包时上传 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg` 作为 workflow artifact,SHALL NOT 将其表述为正式发布或已完成公证的产物(沿用既有"候选仅 artifact、正式发布另行审批"边界)。

#### Scenario: Tag build produces candidate artifacts
- **WHEN** CI 运行 release 工作流
- **THEN** workflow artifact 包含 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`
