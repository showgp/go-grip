# Proposal

## Why

现有 `macos-candidate` 产物链只生成一个 universal 候选包(实测 zip 15M / DMG 16M),用户无论什么机器都要下整包。实测单架构产物(arm64 DMG 7.8M、x86_64 DMG 8.5M)证明:改为按架构分发可把任一用户的下载量减半,且两包之和(16.3M)与现有单个 universal 包基本持平,产物层无冗余增长。

## What Changes

- **BREAKING**:候选产物从单个 universal App/DMG 改为两个单架构候选包:`GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`,不再产出 universal DMG 与候选 App zip(后者原为公证前置物,公证属非目标)
- `macos-archive` 改为按 `ARCHS=arm64` / `ARCHS=x86_64` 分别 archive 一次;`macos/Scripts/build-go-grip.sh` 感知目标架构,只编译对应 Go 架构切片(不再无条件交叉编双架构再 lipo)
- `macos-dmg` 一次产出两个 DMG(含 `/Applications` 链接,命名与现配方一致)
- `macos-candidate-check` 对每个包分别校验:codesign 严格校验、bundle/工具标识符、`verify_arch` 各自单架构、`Signature=adhoc`
- `.github/workflows/release.yml` 上传两个单架构 DMG 作为 workflow artifact(候选产物仅 artifact、正式发布另行审批的边界不变)
- 新增一条简短 ADR:候选分发由 universal 包调整为双单架构包及实测依据(ADR-0002 描述的 Developer ID 正式公证分发形态由后续 change 处理)
- README 打包一节同步两个候选包的命令与产物名

## Capabilities

### Modified Capabilities
- `macos-app-packaging`:universal 候选要求调整为按架构分发的候选包;自包含(无 Go 工具链/CLI/PATH 依赖)、macOS 13+ 双架构覆盖、实际用户路径验收等要求保持不变;"实事求是标注未验证"的延期约束(Intel/macOS 13 实机)在新产物上同样生效

### New Capabilities
(无)

## Non-goals

- Developer ID 签名与公证(维持 ad-hoc)
- chroma 词法器裁剪、嵌入 JS 去留等二进制内容优化(与产物形态正交)
- 正式发布流程(GitHub Release 页发布、用户选包引导页)——仅允许 README 打包一节小改
- `make macos` / `macos-run` 日常开发构建语义(维持现状)
