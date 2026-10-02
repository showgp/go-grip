# macOS App Packaging Spec Delta

## Purpose

定义完整 macOS 功能在本阶段的交付载体与验收边界：生成支持 macOS 13+ 两种架构、内置渲染工具的候选 App 和 DMG，实际验证用户操作路径。本阶段不替代之后的正式签名、公证及安装包验证，也不能把可构建或自动测试通过等同于正式发布就绪。

## ADDED Requirements

### Requirement: Universal self-contained application

打包流程 SHALL 生成支持 macOS 13 及以上系统、Apple Silicon 和 Intel 的 universal App，并随 App 提供能在对应架构运行的 Go 渲染工具。用户使用候选 App SHALL 不需要安装 Go、单独安装 CLI、配置 PATH 或保留源码目录；宿主与内置工具 SHALL 能在所声明的系统和架构上共同提供完整预览功能。

#### Scenario: Use the application without development tools
- **WHEN** 用户运行本阶段打包生成的 App，环境中没有 Go 开发工具或额外安装的 go-grip CLI，也未配置其 PATH
- **THEN** App 使用随包提供的工具完成 Finder 或手动打开的目标预览，不要求从源码目录寻找二进制文件

#### Scenario: Use the same candidate on supported architectures
- **WHEN** 同一候选 App 分别在 macOS 13+ 的 Apple Silicon 和 Intel 环境运行
- **THEN** 宿主及随包提供的工具能够在对应架构运行并提供预览，而不是只有宿主包含两种架构、渲染工具仅支持一种

### Requirement: Candidate DMG and Services-only Finder integration

本阶段 SHALL 提供可生成上述完整 App 及可用于安装该 App 的候选 DMG 的打包流程，不仅提供源码或占位构建输出。候选 App SHALL 包含服务注册和所需本地化内容，不包含遗留 Finder Sync 扩展或其专用转发兼容层；使用 Finder 入口 SHALL 不要求用户启用旧扩展。

#### Scenario: Produce an installable candidate
- **WHEN** 执行本阶段的打包流程
- **THEN** 产生包含完整 universal App 的候选 DMG，安装后的 App 能提供本次规格定义的 Services、菜单栏及预览功能

#### Scenario: Use the new Finder integration
- **WHEN** 用户安装候选 App 并按引导检查系统 Services 启用状态
- **THEN** Finder 打开能力由 Services 提供，不依赖随包发布或单独启用遗留 Finder Sync 扩展

### Requirement: Functional acceptance exercises actual user paths

本阶段的功能验收 SHALL 实际运行候选 App，验证 Finder 服务进入真实浏览器预览、目标复用、停止和异常退出清理，不能仅以编译、自动测试或进程成功创建代替。验收 SHALL 明确区分实际观察到的行为与未验证环境，不把源码中存在对应逻辑当作运行证据。此功能验收 SHALL 不被描述为已完成后续正式安装包的签名、公证或干净安装验证。

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
