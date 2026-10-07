# Tasks

## 1. Go 构建脚本按架构适配

- [x] 1.1 重构 `macos/Scripts/build-go-grip.sh`:解析 `$ARCHS` 映射 Go 目标(arm64→GOARCH=arm64,x86_64→GOARCH=amd64),单架构时仅编译该架构并直接输出,双架构时保留现有 `lipo -create` 合并路径,保留 `-ldflags "-s -w"`;验证:`make macos`(dev 构建走双架构路径)后 `lipo -info` 显示宿主与 go-grip 均为 arm64+x86_64,`make macos-test` 通过
- [x] 1.2 用模拟环境变量单独执行脚本(设 `SRCROOT`/`BUILT_PRODUCTS_DIR`/`CONTENTS_FOLDER_PATH`/`ARCHS=arm64`),验证产出单片二进制;验证:`lipo -info` 仅显示 arm64,重复一次 `ARCHS=x86_64` 仅显示 x86_64

## 2. Makefile 双包流水线

- [x] 2.1 `macos-archive` 改为按 `ARCHS=arm64` / `ARCHS=x86_64` 循环 archive 到 `macos/.build/GoGrip-<arch>.xcarchive`;验证:两个 archive 各含 `Products/Applications/GoGrip.app`,`lipo -info` 确认宿主与 go-grip 均为对应单架构
- [x] 2.2 `macos-dmg` 先 `rm -rf` candidate 目录,再按 arch 循环以现有配方(ditto + `/Applications` 链接 + `hdiutil UDZO`)产出 `GoGrip-arm64.dmg` 与 `GoGrip-x86_64.dmg`;验证:`make macos-dmg` 后 `candidate/` 下只有这两个 DMG 且不再有 `GoGrip.app.zip` / 旧 universal `GoGrip.dmg`
- [x] 2.3 `macos-candidate` 移除 zip 打包步骤;`macos-candidate-check` 改为按包循环校验:codesign 严格校验、`Identifier=com.showgp.GoGrip(.go-grip)`、`Signature=adhoc`、本架构 `verify_arch` 必须通过且另一架构必须失败;验证:`make macos-candidate-check` 两包全部通过,再用临时冒烟脚本把 universal 二进制塞入 arm 包的临时副本并按既有标识符重做 ad-hoc 签名,先确认严格签名、标识符与 ad-hoc 前置检查通过,再确认因架构互斥检查失败非零退出而非签名失败,结束后清理临时样本与脚本
- [x] 2.4 确认 `make macos` / `macos-run` / `macos-test` 语义未变;验证:三个目标各跑一次,`macos-test` 全绿,`macos-run` 能启动 dev App

## 3. CI 与文档

- [x] 3.1 `.github/workflows/release.yml` 的 `Upload candidate App and DMG` 步骤(artifact 名保持 `GoGrip-macos-candidate-<ref>`)与 `.github/workflows/build.yml` 的 PR 候选上传步骤 `path:` 均改为两个 DMG 文件(切流迁移,消除对已移除 zip/universal 产物的引用);验证:`actionlint`(如可用)通过,两处 path 文件名与 2.2 产物逐字一致
- [x] 3.2 新增 `docs/adr/0004-split-candidate-packages-by-arch.md`:记录候选由 universal 改双架构包、实测大小依据、与 ADR-0002 的关系(不改写 ADR-0002);验证:文件存在,引用 ADR-0002 并注明其正式公证分发形态由后续 change 处理
- [x] 3.3 README 打包一节与 `docs/ARCHITECTURE.md` 构建与发布小节更新为 `make macos-candidate` 产出两个单架构 DMG;验证:文档中命令与产物名与 Makefile 产物逐字一致

## 4. 端到端候选验证

- [x] 4.1 `make macos-candidate` 全链路一次;验证:两个 DMG 存在且记录大小(预期 ~8M 量级),`macos-candidate-check` 通过,分别挂载确认 `GoGrip.app` + `Applications` 链接的安装布局
- [x] 4.2 在隔离 Go 工具链与额外 go-grip CLI、未配置其 PATH 且不依赖源码目录的运行环境中实际运行 arm64 候选 App,验证菜单栏出现、Finder 服务进入真实浏览器预览、会话复用/停止及异常宿主终止后的服务清理;验证:这些路径及随包工具的实际观察结果写入候选说明,同一说明标注 Intel 与 macOS 13 实机运行未验证并由后续 change 承接,保留网络卷访问/热重载降级及候选级可卸载卷断卷的既有延期标注,不以静态检查或其他环境结果替代延期维度的实际运行证据,不宣称正式签名、公证或干净安装验证已完成
