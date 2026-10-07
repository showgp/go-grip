# 02 CI upload and packaging docs

Status: done
Blocked by: 01
Covers OpenSpec tasks: 3.1, 3.2, 3.3
Behavior source: ../specs/macos-app-packaging/spec.md
Design source: ../design.md
Task source: ../tasks.md

## Deliverable

CI 的 release 与 PR macOS 流程均上传两个单架构候选 DMG 作为 workflow artifact(沿用"候选仅 artifact、正式发布另行审批"边界),仓库文档记录新的候选分发形态:新增 ADR-0004(不改写 ADR-0002)、README 打包一节与 `docs/ARCHITECTURE.md` 构建与发布小节同步双包命令与产物名。

## Spec and design mapping

| Requirement | Scenario | 本工单贡献与验收归属 |
|---|---|---|
| CI uploads split candidate artifacts | Tag build produces candidate artifacts | 任务 3.1:release 与 PR 候选上传均改为工单 01 产出的双 DMG,保持 workflow artifact 定位,不表述为正式发布或已完成公证 |
| Candidate DMG and Services-only Finder integration | Produce an installable candidate | 任务 3.2、3.3:ADR、README 与 ARCHITECTURE.md 准确描述双包形态及命令;产物构建由工单 01、挂载与实际运行由工单 03 验收 |

相关设计决策:6(单一 artifact 容器与 release/PR 候选上传的双 DMG 路径)、7(ADR-0004 与 ADR-0002 的关系);README 同步决策 3、5 的双包及去 zip 结果。保留[主包装规格](../../../../specs/macos-app-packaging/spec.md) `Explicitly deferred formal distribution` 的 `Present an unsigned candidate for functional review` / `Complete this implementation stage` 边界:文档与 CI 候选材料不授予公开发布许可,不宣称正式签名、公证或安装信任流程已验收。

## Acceptance

- [x] `.github/workflows/release.yml` 的 upload 步骤 `path:` 为 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg` 两个文件,artifact 名保持 `GoGrip-macos-candidate-<ref>`;`.github/workflows/build.yml` 的 PR 候选上传步骤 `path:` 同步为两个 DMG(切流迁移,不再引用已移除的 zip/universal 产物);`actionlint`(如可用)通过,两处文件名与 01 产物逐字一致(任务 3.1)
- [x] `docs/adr/0004-split-candidate-packages-by-arch.md` 存在,记录双架构候选决策与实测大小依据,引用 ADR-0002 并注明其 Developer ID 正式公证分发形态由后续 change 处理(任务 3.2)
- [x] README 打包一节与 `docs/ARCHITECTURE.md` 构建与发布小节为 `make macos-candidate` 产出两个单架构 DMG,命令与产物名和 Makefile 逐字一致(任务 3.3)

## Closure record

- 关闭:2026-10-07,状态 `done`。用户批准:本票范围、验证缝与开工批准;ADR 编号修订(0003→0004,因 `docs/adr/0003-one-preview-session-per-open-target.md` 已存在且被 `.scratch/macos-platform-v1/spec.md` 引用);范围扩展(纳入 build.yml PR 候选上传切流迁移与 `docs/ARCHITECTURE.md` 构建与发布小节同步)。发布时账本 6/11,无已勾任务冲突。
- 变更面(本票):`.github/workflows/release.yml`(`Upload candidate App and DMG` 步骤 `path:` 改为 `macos/.build/candidate/GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`,artifact 名 `GoGrip-macos-candidate-${{ github.ref_name }}` 保持);`.github/workflows/build.yml`(PR 候选上传 `path:` 同步两个 DMG,artifact 名 `GoGrip-macos-candidate` 保持);`docs/adr/0004-split-candidate-packages-by-arch.md`(新增);`README.md`(打包一节:命令注释、per-arch archive/DMG 管线、按架构安装指引,去掉 zip 与 `ditto -x -k`);`docs/ARCHITECTURE.md`(§十六:make 目标注释、构建脚本/archive 配置/产物/CI 条目同步双包与去 zip);获批工件修订:`design.md` 决策 6、7、`tasks.md` 3.1、3.3、本工单交付物/映射/验收。无新增永久测试(仓库打包链按 tasks.md 以命令级验证为准);临时一致性脚本与临时 universal 测量镜像已删除。`Makefile` 与 `macos/Scripts/build-go-grip.sh` 的改动属工单 01,未计入本票。
- 实测证据(本次亲见,Apple Silicon):
  - 任务 3.1:`actionlint` v1.7.12(经 `go run`,不在 PATH)对 `release.yml`(默认与 `-shellcheck=` 各一次)及 `release.yml`+`build.yml` 均 exit 0;YAML 解析断言两处 upload `path` 等于 Makefile 展开推导的产物路径,`make macos-candidate` 均先于 upload,artifact 名逐字保持,两工作流无旧产物名残留。
  - 任务 3.2:ADR-0004 存在;含 `GoGrip-arm64.dmg`/`GoGrip-x86_64.dmg`、ADR-0002 引用与后续 change 承接;quoted 实测大小与盘上产物逐字节一致(8,085,364 / 8,875,546 B);universal 对比 16,659,223 B 为本次以同一 DMG 配方(残留 universal xcarchive App + ditto + /Applications 链接 + hdiutil UDZO)实测;合计 +1.81% 结论按 Spec 审阅修复为 "about 1.8% more"。
  - 任务 3.3:README 打包一节与 ARCHITECTURE.md §十六 的 `make macos-candidate` 命令与产物名与 Makefile 逐字一致;工作流/README/ARCHITECTURE.md 三处均无 `GoGrip.app.zip`/`GoGrip.dmg`/`GoGrip.xcarchive`/`ditto -x -k` 残留。
  - 一次性 ruby 一致性脚本 14/14 PASS(`candidate/` 恰含两个 DMG);两轴独立只读审阅首轮:Spec 1 blocking(ADR 合计尺寸结论)+ Standards 2 advisory(README CI 声称受 build.yml 失真、ADR 结论);修复与获批扩展后增量复审两轴 0 blocking / 0 advisory / 0 未决 scope decision;审阅未复跑实现者检查。审阅覆盖限制:未实际运行 GitHub Actions;DMG 挂载与实机运行不在本票。
- 实施偏差与解释:ADR 编号按批准由 0003 改为 0004;`build.yml` 与 `ARCHITECTURE.md` 为工单 01 切流后的遗留消费者/陈旧文档,经批准以切流迁移纳入本票并同步修订工件;两处 upload 步骤显示名按设计决策 6"改动最小"保持未改(仅 `path:` 与行为变更)。
- 账本核对:3.1、3.2、3.3 唯一贡献者为本票,整项验证标准已在集成状态观察;`tasks.md` 三项已勾;4.1–4.2(票 03,原阻塞于本票)保持未勾。`openspec instructions apply` 复核 11 项 / 9 完成,与文件一致。
- 未覆盖/限制:未实际运行 GitHub Actions(本地等价校验:actionlint + YAML 解析 + 路径/产物逐字比对);DMG 挂载布局与真实用户路径由票 03 验收;Intel/macOS 13 实机、网络卷与候选级断卷维度按 `Explicitly deferred environment-dependent acceptance` 未验证;README 候选运行状态小节("both slices statically checked" 等运行/延期证据描述)由票 03 任务 4.2 随候选说明更新,本票未改;不声称正式签名、公证或干净安装。
