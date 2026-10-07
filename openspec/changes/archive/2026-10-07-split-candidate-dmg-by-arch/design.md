# Design

## Context

当前产物链是单条 universal 管线:一次 `xcodebuild archive`(默认双架构),`macos/Scripts/build-go-grip.sh` 无条件交叉编 Go 双架构并 `lipo` 合并,产出 1 个 DMG + 1 个候选 zip。`macos-candidate-check` 按 universal 校验(要求两个 `verify_arch` 都通过)。实测数据(strip 后):universal zip 15M / DMG 16M;单架构 zip:arm 6.6M;单架构 DMG:arm 7.8M / x64 8.5M。

## Goals / Non-Goals

**Goals:**

- 一条 `make macos-candidate` 命令产出并校验 `GoGrip-arm64.dmg` + `GoGrip-x86_64.dmg`
- 每个包只含本架构宿主与 Go 工具切片,校验能证明这一点
- dev 构建(`make macos` / `macos-run` / `macos-test`)语义不变

**Non-Goals:**

- 同 proposal(公证、chroma 裁剪、正式发布流程、dev 构建语义);`macos/Scripts/build-go-grip.sh` 当前的 Go 构建参数 `-ldflags "-s -w"`(剥离符号表与 DWARF)属于既有构建状态,本 change 不改动二进制内容本身

## Decisions

1. **两次 archive 替代一次 universal archive**:`macos-archive` 循环 `ARCHS=arm64` / `ARCHS=x86_64` 各跑一次 `xcodebuild archive`,产物分别落在 `macos/.build/GoGrip-<arch>.xcarchive`。理由:per-arch App 是双包的最小完整单元,后续 DMG/校验都从各自 archive 取,互相不掺;缺点是编译时间 ×2(~30s→~60s,可接受)。不用 `lipo -thin` 从 universal 归档事后拆——thin 后需要重签名,还留下未删除的另一架构元数据,校验语义更脏。
2. **`build-go-grip.sh` 按 `ARCHS` 环境变量适配**:Xcode 会把目标架构传给 script phase。脚本解析 `$ARCHS`:`arm64`→`GOARCH=arm64`、`x86_64`→`GOARCH=amd64`;单架构时直接拷贝,双架构时保持现有 `lipo -create` 路径。这样 dev 构建(不设 ARCHS,Xcode 默认双架构)零改动,候选构建自然单片。
3. **`macos-dmg` 产出两个 DMG,先清空 `candidate/`**:配方沿用(ditto + `/Applications` 符号链接 + `hdiutil UDZO`),文件名按架构;开头 `rm -rf $(CANDIDATE_DIR)` 防止旧 `GoGrip.app.zip` / `GoGrip.dmg` 陈留。
4. **架构互斥校验的实现**:对包内两个二进制分别跑 `lipo -verify_arch <own>`(必须成功)与 `lipo -verify_arch <other>`(必须失败),加上现有 codesign/标识符/adhoc 检查,全部按 arm/x64 两包循环执行。
5. **`macos-candidate` 移除 zip 步骤**:zip 原是公证前置物,公证属非目标;`macos-candidate` 语义收敛为 `macos-dmg + macos-candidate-check`。
6. **CI 保持单一 artifact 容器**:release.yml 的 `build-macos-app` job 的 upload-artifact 名字不变(`GoGrip-macos-candidate-<ref>`),`path:` 改为两个 DMG 文件;`build.yml` 的 PR 候选上传步骤同步改为两个 DMG 文件(切流迁移,消除对已移除 zip/universal 产物的引用)。理由:改动最小,两个包仍属同一候选批次。
7. **ADR**:`docs/adr/0004-split-candidate-packages-by-arch.md` 记录"候选由 universal 改双架构包"及实测依据;ADR-0002 描述的 Developer ID 正式公证分发形态由后续 change 处理,本 change 不改写该 ADR。

## Risks / Trade-offs

- 验收环境为未安装 Rosetta 的 Apple Silicon Mac,无法执行 x86_64 候选包;Intel 与 macOS 13 实机运行均按 `Explicitly deferred environment-dependent acceptance` 标注未验证并由后续 change 承接。新增校验能保证切片正确性与签名,不冒充运行证据。
- `ARCHS` 传入 script phase 依赖 Xcode 行为;如果未来 Xcode 改为不传变量,脚本回退到双架构(候选校验会失败并暴露,不静默错发)。
- 编译时间 ×2,换取的是产物瘦身与校验可证明性。
