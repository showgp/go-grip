# 01 Per-arch build script and dual candidate pipeline

Status: done
Blocked by: None
Covers OpenSpec tasks: 1.1, 1.2, 2.1, 2.2, 2.3, 2.4
Behavior source: ../specs/macos-app-packaging/spec.md
Design source: ../design.md
Task source: ../tasks.md

## Deliverable

`make macos-candidate` 产出并校验两个单架构候选包(`GoGrip-arm64.dmg`、`GoGrip-x86_64.dmg`),dev 构建(`make macos` / `macos-run` / `macos-test`)语义保持不变。包含 `macos/Scripts/build-go-grip.sh` 的 `$ARCHS` 适配与 Makefile 双包流水线(clear candidate、per-arch archive、双 DMG、互斥校验、去 zip)。

## Spec and design mapping

以下场景名称均属于 Behavior source 指向的 delta spec;任务进度只记入 `../tasks.md`,本表用于界定静态校验与后续实机验收的归属。

| Requirement | Scenario | 本工单贡献与验收归属 |
|---|---|---|
| Per-architecture candidate verification | Verify a candidate before acceptance / Reject wrong architecture assignment | 任务 2.3:两包签名、标识符及架构互斥校验,并证明负向冒烟命中架构拒绝分支 |
| Architecture-split self-contained application | Candidates together cover both architectures | 任务 1.2、2.1:两包宿主与随包工具的对应单架构静态检查;不替代 Intel/macOS 13 实机证据 |
| Architecture-split self-contained application | Use the application without development tools / A candidate runs on its own architecture | 任务 1.1、1.2、2.1 提供随包工具与两架构构建;arm64 自包含运行由工单 03 的任务 4.2 验收,Intel/macOS 13 实机维度仍延期 |
| Candidate DMG and Services-only Finder integration | Produce an installable candidate | 任务 2.2、2.3:双 DMG 与去 zip;实际挂载布局及 Finder/菜单栏/预览由工单 03 的任务 4.1、4.2 验收 |

相关设计决策:1(双 archive)、2(`$ARCHS` 适配)、3(双 DMG 与清空候选目录)、4(架构互斥校验)、5(去 zip)。任务 2.4 承接 dev 构建语义不变的设计目标,包含实际启动 dev App。

## Acceptance

- [x] `make macos` 后 `lipo -info` 显示宿主与 go-grip 均为 arm64+x86_64,`make macos-test` 通过;运行 `make macos-run` 并实际观察 dev App 启动(任务 1.1、2.4)
- [x] 模拟 Xcode 环境单独执行脚本:`ARCHS=arm64` 产出仅 arm64 的 go-grip,`ARCHS=x86_64` 仅 x86_64(任务 1.2)
- [x] `make macos-archive` 后 `macos/.build/GoGrip-arm64.xcarchive` 与 `GoGrip-x86_64.xcarchive` 各含 `Products/Applications/GoGrip.app`,`lipo -info` 均为对应单架构(任务 2.1)
- [x] `make macos-dmg` 后 `candidate/` 只有 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`,无 `GoGrip.app.zip` / 旧 universal `GoGrip.dmg`(任务 2.2)
- [x] `make macos-candidate-check` 两包通过;临时冒烟脚本把 universal 二进制塞入 arm 包的临时副本,按既有标识符重做 ad-hoc 签名并确认严格签名、标识符与 ad-hoc 前置检查通过,再确认因架构互斥检查失败非零退出而非签名失败;结束后清理临时样本与脚本(任务 2.3)

## Closure record

- 关闭:2026-10-07,状态 `done`。用户批准:本票范围、实施计划与验收缝(`build-go-grip.sh` 按 `$ARCHS` 输出 + Makefile 双包流水线;负向冒烟以 `CANDIDATE_BUILD` 覆盖在临时副本上运行真实 `macos-candidate-check`)。发布时账本 0/11,无已勾任务冲突。
- 变更面:`macos/Scripts/build-go-grip.sh`(`$ARCHS` 精确 `arm64`→GOARCH=arm64、`x86_64`→GOARCH=amd64,单架构编译到 /tmp 后 `cp` 进 `Contents/MacOS/go-grip`;其余值保留原双架构 `lipo -create` 路径;保留 `-ldflags "-s -w"`(票前用户未提交改动)与 ad-hoc 标识符);`Makefile`(`CANDIDATE_BUILD`/`CANDIDATE_ARCHS`/`CANDIDATE_DIR`;`macos-archive` 按 ARCHS 循环出 `GoGrip-<arch>.xcarchive`;`macos-dmg` 先清空并重建 `candidate/`、按 arch 暂存 + `ditto` + `/Applications` 链接 + `hdiutil UDZO`;`macos-candidate` 去 zip;`macos-candidate-check` 按包循环:严格签名、`Identifier`、`Signature=adhoc`、本架构 `verify_arch` 通过且另一架构必须失败)。无新增永久测试或仓库文件(本仓库打包链按 tasks.md 以命令级验证为准);临时冒烟脚本与样本已删除。
- 实测证据(Apple Silicon / Xcode 27.0 / Go 1.26.3,本次亲见,均为本机集成状态):
  - 任务 1.1:`make macos` BUILD SUCCEEDED,宿主与 go-grip `lipo -info` 均 `x86_64 arm64`;`make macos-test` TEST SUCCEEDED,95 tests / 0 failures。
  - 任务 1.2:模拟环境单独执行脚本(`SRCROOT`/`BUILT_PRODUCTS_DIR`/`CONTENTS_FOLDER_PATH`):`ARCHS=arm64` 产 thin arm64(负向 x86_64 verify exit 1),`ARCHS=x86_64` 产 thin x86_64(负向 arm64 verify exit 1);并验证单架构运行可覆盖此前双架构 fat 工具。
  - 任务 2.1:`make macos-archive` 两个 archive 各含 `Products/Applications/GoGrip.app`,宿主与 go-grip 均对应单架构。
  - 任务 2.2:`make macos-dmg`(及 `make macos-candidate` 全链路)后 `candidate/` 仅 `GoGrip-arm64.dmg`(8,085,360 B)与 `GoGrip-x86_64.dmg`(8,875,534 B),无 `GoGrip.app.zip`、无旧 universal `GoGrip.dmg`。
  - 任务 2.3:`macos-candidate-check` 两包 `candidate checks passed`。负向冒烟:arm 包临时副本塞入 `lipo -create` universal 宿主、按 `com.showgp.GoGrip` 重签,前置严格签名/标识符/adhoc 全通过,`make macos-candidate-check CANDIDATE_BUILD=<tmp>` 非零退出且信息为 `unexpected x86_64 slice in …GoGrip (candidate must contain arm64 only)`(架构互斥分支,非签名失败);样本与脚本已清理。
  - 任务 2.4:`make macos-run` Debug 构建成功并实际启动(pid 71976,System Events 进程存在),随后正常退出;`macos-test` 全绿;`macos` 保持双架构(见 1.1)。
- 实施偏差与解释:Go 1.26.3 的 `go build -o` 拒绝覆盖已存在的非 object 文件(universal/fat Mach-O 不在 `objectMagic` 表内),故单架构路径按设计决策 2 编译到 /tmp 后 `cp`(对外行为等价);`macos-dmg` 在清空后显式 `mkdir -p candidate`(首次实测暴露的缺口:`rm -rf` 后 hdiutil 因父目录缺失失败,已修复并复测通过)。`hdiutil create` 弃用警告为既有命令行为,不在本票范围。
- 独立审阅(只读子代理,未复跑实现者检查):Standards 轴 0 阻塞 / 0 advisory,YAGNI 逐项追溯至任务 1.1/1.2、2.1–2.3 与设计决策 1–5;Spec 轴 0 阻塞 / 0 advisory / 0 未决 scope decision。审阅覆盖限制:挂载布局、候选实际运行路径、Intel/macOS 13 实机维度不在本票 pass 范围。
- 账本核对:1.1、1.2、2.1、2.2、2.3、2.4 的唯一贡献者为本票,整项验证标准已在集成状态观察(如上证据),`tasks.md` 六项已勾;3.1–3.3(票 02)与 4.1–4.2(票 03)保持未勾,由后续票承接。`openspec instructions apply` 复核 11 项 / 6 完成,与文件一致。
- 未覆盖/限制:本次校验面向 per-arch archive 内的 App(设计决策 1/4 与任务 2.3 约定的缝),DMG 挂载布局与真实用户路径由票 03 验收;Intel 与 macOS 13 实机、网络卷及候选级断卷维度按 `Explicitly deferred environment-dependent acceptance` 未验证;不声称正式签名、公证或干净安装;`macos/.build` 内旧 universal `GoGrip.xcarchive` 为历史构建残留在本票未删除(不在验收范围)。
