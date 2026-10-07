# 03 End-to-end candidate verification

Status: done
Blocked by: 01, 02
Covers OpenSpec tasks: 4.1, 4.2
Behavior source: ../specs/macos-app-packaging/spec.md
Design source: ../design.md
Task source: ../tasks.md

## Deliverable

对完整候选产物链做一次端到端验证并留档:双 DMG 存在、大小符合预期、校验通过、安装布局正确;arm64 包实际运行走真实用户路径(Finder→浏览器预览、会话复用/停止、异常退出清理),x86_64 按既有延期约束如实标注。本工单为 verification-only(verification-only integration 工单),不改产品代码;若验证暴露缺陷,缺陷修复按发现单独报告,不在此工单夹带。

## Spec and design mapping

| Requirement | Scenario | 本工单贡献与验收归属 |
|---|---|---|
| Per-architecture candidate verification | Verify a candidate before acceptance / Reject wrong architecture assignment | 任务 4.1 复核最终双包的全链路正向校验;负向架构拒绝证据由工单 01 的任务 2.3 提供,不重复构造样本 |
| Architecture-split self-contained application | Use the application without development tools / A candidate runs on its own architecture | 任务 4.2:在排除开发依赖的环境中实际运行 arm64 包,观察随包工具完成预览;Intel/macOS 13 实机维度延期且如实留档 |
| Architecture-split self-contained application | Candidates together cover both architectures | 任务 4.1 复核工单 01 的双包架构静态校验,不表述为 Intel/macOS 13 实机运行证明 |
| Candidate DMG and Services-only Finder integration | Produce an installable candidate / Use the new Finder integration | 任务 4.1、4.2:双 DMG 挂载布局与 arm64 候选的 Services、菜单栏及预览实际路径 |
| Functional acceptance exercises actual user paths | Verify a real Finder-to-browser launch / Verify reuse and explicit stopping / Verify abnormal host termination | 任务 4.2:观察真实浏览器预览、复用、显式停止及异常宿主终止后的服务清理 |
| Explicitly deferred environment-dependent acceptance | Close this implementation stage with deferred environments / Claim support for a deferred environment | 任务 4.2:候选说明保留 Intel/macOS 13、网络卷及候选级可卸载卷断卷的延期维度,不宣称已验证 |

相关设计决策:1、3、4、5 的最终双包链路复核;沿用 Risks / Trade-offs 的环境限制。保留[主包装规格](../../../../specs/macos-app-packaging/spec.md) `Explicitly deferred formal distribution` 的 `Present an unsigned candidate for functional review` / `Complete this implementation stage` 边界:本工单不验收正式签名、公证、干净安装或授予公开发布许可。

## Acceptance

- [x] `make macos-candidate` 全链路一次,`candidate/` 下两个 DMG 存在并记录大小(预期 ~8M 量级),`macos-candidate-check` 通过;分别挂载确认各含 `GoGrip.app` + `Applications` 链接(任务 4.1)
- [x] 在隔离 Go 工具链与额外 go-grip CLI、未配置其 PATH 且不依赖源码目录的运行环境中实际运行 arm64 候选 App,观察菜单栏出现及随包工具完成预览:Finder 服务进入真实浏览器预览、重复打开同一目标无重复服务、显式停止后服务退出且资源释放、有预览服务时强制终止宿主其服务亦结束;环境条件与这些路径的实际观察结果写入候选说明,不以源码存在性代替运行证据(任务 4.2)
- [x] 同一候选说明中标注 Intel 与 macOS 13 实机运行未验证并由后续 change 承接,保留网络卷访问/热重载降级及候选级可卸载卷断卷的既有延期标注,不以静态检查或其他环境结果替代延期维度的实际运行证据;arm64 验收不表述为已完成正式签名、公证或干净安装验证(任务 4.2)

## Closure record

- 关闭:2026-10-07,状态 `done`。用户批准:本票范围(verification-only)与验收缝(链构建 + DMG 挂载布局、净化环境 arm64 真实运行路径、README 候选说明),以及执行许可——替换 `/Applications/GoGrip.app` 不备份、立即退出运行中的旧宿主、以 `env -i` 直接执行作隔离取证、文档仅改 README 候选说明。发布时账本 9/11(4.1、4.2 未勾),无已勾任务冲突。
- 变更面(本票):`README.md` 候选运行状态条目("Not available yet" bullet)重写(观察结果 + 延期标注,去掉 universal 期措辞 "both slices statically checked"/"ticket 13");本工单文件与 `tasks.md` 4.1/4.2 勾选。无产品代码、测试、工作流或其他文档改动;证据留存 `~/GoGrip-split03-evidence/`(索引 `state/EVIDENCE-SUMMARY.md`)。
- 实测证据(Apple Silicon / macOS 27.0.1 / Xcode 27.0;本次亲见,均为集成状态):
  - 任务 4.1:`make macos-candidate` 全链路 EXIT 0(两次 `ARCHIVE SUCCEEDED`、两 DMG created、`macos-candidate-check` 两包 `candidate checks passed`);`GoGrip-arm64.dmg` 8,085,368 B、`GoGrip-x86_64.dmg` 8,875,538 B;分别挂载确认各含 `GoGrip.app` + `Applications → /Applications` 链接,宿主/工具标识符(`com.showgp.GoGrip`/`.go-grip`)、`Signature=adhoc`、严格签名与对应 thin 架构校验通过;两镜像均已 detach。负向架构拒绝证据沿用票 01 任务 2.3(未重复构造)。
  - 任务 4.2:arm64 候选经挂载镜像装入 `/Applications`(宿主 `906fcacf…`、工具 `08176ed2…`;替换旧 universal 候选,用户批准不备份);以 `env -i HOME=/Users/ray` 在源码目录之外直接启动宿主(pid 73603),`sysctl KERN_PROCARGS2` 读取其环境为 HOME + dyld 注入项、无 `PATH`;AX 观察菜单栏项出现;Finder Services 实际调用(真实 Finder 选择 + Services 菜单项 "Open with GoGrip")由该宿主处理:包内工具子进程 73712(argv 为 `/Applications/GoGrip.app/Contents/MacOS/go-grip --managed … -r -- /tmp/gogrip03-split/main`,环境同样无 `PATH`)监听 127.0.0.1:63022,HTTP 200 且正文含 fixture marker,默认浏览器(Chrome)真实打开该 URL 并以 AX 读到渲染文本;重复打开同一目标(再次 Services)后子进程数/PID/UUID/端口/URL 与宿主数均不变(无重复服务);面板显式 **Stop** → 子进程与监听消失、宿主存活、行 `Stopped`;有活动预览时 `kill -9 73603` → 2 s 内宿主、子进程与监听全部消失。收尾:本轮 3 个浏览器标签、Finder 窗口、`/tmp/gogrip03-split` fixture 均已清理,无残留进程/监听/挂载。
  - 任务 4.2(候选说明):README 条目记录上述环境条件与观察;Intel 与 macOS 13 实机运行标注未验证(本机无 Rosetta,x86_64 证据仅静态签名/架构检查)并由后续 change 承接;保留网络卷访问/热重载降级与候选级可卸载卷断卷的延期标注;不宣称正式签名、公证或干净安装。
- 实施偏差与披露:旧宿主退出后其浏览器标签(`127.0.0.1:59450`,已失效)按最小干预保留未关;`com.showgp.GoGrip` 最近目标默认值现含已删除的 `/tmp` fixture(应用正常写入,未回改、未清理);替换前发现并 detach 了先前会话残留的 `/Volumes/GoGrip`(旧 universal 候选镜像)挂载。
- 独立审阅(并行只读子代理 `ReviewStandards03`/`ReviewSpec03`,未复跑实现者检查):Standards 轴 0 阻塞 / 0 advisory(2 处台账注记:证据摘要中 child env 文件指引与项数措辞失准,已按原始输出修正,不影响 README 结论);Spec 轴 0 阻塞 / 0 advisory / 0 未决 scope decision。审阅覆盖限制:未重跑构建/运行、未独立见证执行;延期维度与正式签名/公证/干净安装不在本票范围。
- 账本核对:4.1、4.2 唯一贡献者为本票,整项验证标准已在集成状态观察;`tasks.md` 两项已勾;1.1–2.4(票 01)、3.1–3.3(票 02)保持已勾;`openspec instructions apply` 复核 11 项 / 11 完成,与文件一致。未执行:提交、主规格同步、归档、下一票;延期维度(Intel/macOS 13 实机、已挂载网络卷访问与热重载降级、候选级可卸载卷断卷)保持未验证,由后续 change 承接。
