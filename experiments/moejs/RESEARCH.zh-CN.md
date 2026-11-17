# 用 moejs 取代运行时 Node：验证记录

## 用户目标与边界

用户在 main 的 bc5ad65 主程序包看到 Counter/Callout 退成源码；随后明确要求新开 PR 验证 moejs 并提供测试包。本实验从 bc5ad65 独立 worktree/分支开始，不合并 main，不写上游 Ciao1019/ArcMD，不改变小时维护任务的 main 发布标记。保留“agent 写文件、人阅读、人让 agent 修改”的原生正文＋块级 WebView 方案，不转成整页网页。

## 官方依据（2026-11-17，UTC+8）

- 官方项目：https://github.com/MoeForge/moejs
- 固定版本：[v0.1.1](https://github.com/MoeForge/moejs/releases/tag/v0.1.1)，官方 release target `b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070`，Go 模块完整性由 go.sum 固定。
- 官方 README / guide 说明：纯 Go、可 CGO_ENABLED=0 交叉编译；`Compile/Module.Hook`、`NewRuntime/Load/Call/UnmarshalWith`；`Options.MemoryLimit`、`DisableDynamicCode`、跨 goroutine `Interrupt`；默认无 host API，Importer 由宿主决定。本实验直接查阅该版本的 Go API 与源码，不把“支持现代 JS”当作 MDX 兼容性证明。
- MDX 官方编译器 `@mdx-js/mdx@3.1.1` 的 `compileSync` 可以输出 `function-body` 格式的普通 JavaScript。因此无需在用户电脑安装 Node，也无需运行时 esbuild：可信编译器与固定 React runtime 在构建期打成 bundle，最终 `go:embed` 进入应用。

## 最小实验与踩坑

1. 构建期用 Node/npm 和锁定版本 esbuild，把真实 MDX 编译器打包为独立 ESM；没有用手写 JSX 正则替代 MDX。
2. 首次 `moejs.Load` 失败，实际错误 `ReferenceError: document is not defined`。定位为 browser condition 选中了 `decode-named-character-reference/index.dom.js`；改用同一个包的纯 JS `index.js`，没有给 moejs 安装 DOM 或任意 OS API。
3. 最小 Go probe 用 moejs 真正编译新输入 `<Counter initial={2+3} />`，成功返回普通 JS 函数体、随机命名空间投影，且输入表达式未执行。保留可重跑的 `TestMoeProbe`，不是只跑 `1+1`。
4. 接入 `internal/jsx.Compile`，主程序打开 `.mdx` 直接走内置引擎；不调用 node，不查 PATH，不退回旧 Node 编译器。每次文件内容动态编译，不是只打包预编译 demo。固定 React runtime 与编译结果合成 hash-CSP HTML，由原来的 Untrusted WebView 渲染。
5. 新增 `--jsx-check file.mdx`，直接调用真实主程序的编译链。Linux 执行实际二进制时使用 `env -i PATH=/no-node ARCMD_JSX_NODE=/missing ARCMD_JSX_COMPILER=/missing`，新文件含 Counter、嵌套 Callout、Unicode 均产出 HTML。
6. 新的 JSDOM 测试运行实际无 Node 子进程的输出：嵌套 Counter 从 `17+24=41` 点到 42，修改主题 token 后仍为 42；Markdown 子节点变粗体；未知组件明确显示错误与源码；混合大小写 `</SCRIPT>` 不穿透脚本，CSP 哈希保持正确。

## 信任与资源限制

moejs 仅运行固定的可信编译器。用户表达式是 parser 的字符串输入，只生成输出 JS，不在 moejs 或 Go 中求值。测试含一个如果执行就抛错的表达式，编译仍成功，证明这一步不执行文档表达式。

- 不调用 SetGlobal 安装任何主机函数；Importer 为 nil；不安装 node、webapi、eventloop。process/require/fetch/document 都是 undefined。
- 禁止 eval/Function，5 秒 context 截止并使用 Interrupt；每文档新 runtime，256 MiB JS 资源限制，512 KiB 源码、32 岛屿、16 MiB/200000 节点返回上限。测试验证死循环可中断、分配越界返回 ErrMemoryLimit、过长源文件拒绝。这是 JS 分配预算，不是操作系统总内存硬隔离；可信固定 bundle 的一次 Go 解析不处理文档输入。
- 用户表达式最终仅在 Untrusted WebView 执行，无 Go bind，继续 hash CSP、禁网络/导航。主题更新/测高继续是 Page.Eval IIFE，不重载。React boundary 对未知组件给可见错误。
- 投影不写回用户文件；保存通过命名空间还原源码。打开挂载、未变脏、源文件未被改写和 round-trip 继续有测试。

## 失败可见性修复

主程序消费 `jsxErr`、manifest.Diagnostics、逐岛屿编译错误，以及 WebView 创建/加载错误，显示在原生文档区域；干净的主窗口文档可重新载入并重试。不会把源码 fallback 说成渲染成功。未知组件在 WebView 内显示错误。没有加入 Go bridge 来回传组件事件。

## 构建与交付

```sh
npm ci --prefix experiments/jsx-webview --no-audit --no-fund
node experiments/moejs/build.mjs  # 固定 bundle 已提交，正常 go build 无需 Node
# 构建时许可证快照在 third_party/moejs-mdx，moejs 的 THIRD_PARTY_NOTICES 随包携带。
go test ./internal/jsx ./cmd/arcmd ./cmd/arcmd-jsx ./internal/editor
go test ./internal/app
go vet ./internal/jsx ./internal/app ./cmd/arcmd ./cmd/arcmd-jsx ./internal/editor
npm test --prefix experiments/jsx-webview
node --test experiments/moejs/runtime.test.mjs
PYTHONDONTWRITEBYTECODE=1 python3 experiments/moejs/package_macos.py --out /workspace/arcmd-maintenance/moejs-releases
```

测试包是真实 `cmd/arcmd`，名称 `ArcMD MoeJS Test.app`、版本 `0.2.0-moejs-test`，独立 bundle ID 与默认配置/空间目录。无 Node/npm/esbuild 可执行文件或本地服务；不配置上游自动更新。整包 ad-hoc 签名，未公证。源提交、编译器哈希、构建信息和 ZIP SHA256 随 Release 发布。

未真实 Mac 测试；Finder、WKWebView、系统信任仍需用户验收。第一期仍不支持文件 import/完整模块系统、行内 JSX、只读标题折叠。只在重新打开/重载时编译，不保证计数状态跨保存保留。此验证不自动合并 main。

## 测试结果

- 相关 Go 测试、go vet 通过；旧 Node 实验 npm 15/15 通过。
- 新 moejs 实际应用输出的浏览器测试 3/3 通过；无 Node/PATH 用例、资源预算、输入不执行、可见错误测试通过。
- `internal/app` 全量结果需和 main 的 6 个已知基线失败区分，最终结果见本 PR 正文与 Release 说明；不借机修改基线失败。
- 本轮全量 `go test ./internal/app` 已完成：仍恰好是基线原有 6 项失败（TestBottomBlockHandleMenuAndIndentSubmenu、TestRebindShortcuts、TestLibrarySpaces、TestLibraryCardDrag、TestQACitationPreviewDoesNotOccludeAnswerText、TestMoreMenuAndSettings），没有新增失败。

## 2026-11-17（UTC+8）：用户验收后合入与主线发布

用户确认实验包 `moejs-test-8dc66bb` 测试通过并授权合入、发布。PR #2 普通 merge 提交为 `0f8ecd239f94291fe530f6fb55ceaedfc64f3220`，不改写实验历史，保留原预发布。随后仅做主线发布支持修正：产品名/默认配置恢复 ArcMD、版本 0.2.1；新增 moejs 主线打包入口，旧 `experiments/jsx-webview/package_main_macos.py` 转交它，杜绝小时构建退回 Node 方案。

主线包源码与上游基线、无 Node 烟测结果、签名/哈希随 BUILD-INFO 发布。用户验收的是先前实验包；新主线 ZIP 未再次真实 Mac 测试，发布说明明确区分。事务全程持有维护锁；成功发布并核验资产后才原子更新主线维护状态，上游已打包 Release 基线不凭空推进。详见 [主线安装说明](main-release.zh-CN.md)。

## 2026-11-17（UTC+8）：0.2.2 正常界面清理与身份核查

用户指出正式版仍显示 `JSX · moejs 内置编译器（无需 Node）`，要求处理后重新发布。移除这个无条件正常态技术横幅；无诊断时函数直接返回，不生成状态节点、padding 或占位；失败时仍显示编译/逐岛屿/WebView 诊断与主窗口重试，CLI 诊断保留。新增正常态文本/布局零占位、真实文档无技术横幅、逐岛屿错误和重试保留的测试。相关生产路径检查未发现其他同类常驻实验状态 UI；没有扩大到其他编辑器功能。

同时只读检查此前各类包的 Info.plist、配置及 MyGo 单实例键，确认预览、Test、正式主线是可并存的不同身份；主线 0.2.1/0.2.2 名称、bundle ID、配置均保持一致。详见 [APP-IDENTITY.zh-CN.md](APP-IDENTITY.zh-CN.md)。未访问用户 Mac，不能确定用户副本实际数量、位置或 LaunchServices 状态，也没有自动删除用户应用/文档/配置。

### 0.2.2 验证结果（Linux box）

- `go test ./internal/app -run 'Test(JSX|OpenMDX|OpenMoe)' -count=1`：通过，覆盖成功无横幅/无布局占位、失败可见与重试、无 Node 嵌套文档、源码完整性。
- `go test ./internal/jsx ./cmd/arcmd ./cmd/arcmd-jsx ./internal/editor`：通过。
- `go vet ./internal/jsx ./internal/app ./cmd/arcmd ./cmd/arcmd-jsx ./internal/editor`：通过；`go run ./tools/ci version` 输出 `0.2.2`。
- `npm test --prefix experiments/jsx-webview`：15/15；`node --test experiments/moejs/runtime.test.mjs`：3/3（实际主程序无 Node 编译、交互、未知组件可见失败、CSP）。Node 只用于开发测试，不作为应用运行时依赖。
- `go test ./internal/app -count=1`：只有既知 6 项基线失败：`TestBottomBlockHandleMenuAndIndentSubmenu`、`TestRebindShortcuts`、`TestLibrarySpaces`、`TestLibraryCardDrag`、`TestQACitationPreviewDoesNotOccludeAnswerText`、`TestMoreMenuAndSettings`。未顺手修改这些无关功能，不将本次全包测试描述为全绿。
- 构建真实 `cmd/arcmd` 的 macOS arm64 包，复用整包 ad-hoc 签名及无 Node 主线打包入口。包结构/签名资源/公开下载 SHA256 在发布流程另行核验，结果写 Release；不宣称新包已在真实 Mac 验收。

## 2026-11-17（UTC+8）：MDX 3.1.1 分阶段兼容 / phase 1

用户授权只在 fork 新分支继续迭代并更新根 README；不改已提交上游分支、不合并 main、不发 Release。核查当前直接依赖本来已经精确 pin `@mdx-js/mdx` 3.1.1，因此不追 latest、不做无必要版本升级。官方 3.1.1 package manifest：<https://github.com/mdx-js/mdx/blob/3.1.1/packages/mdx/package.json>；语法定义 <https://mdxjs.com/docs/what-is-mdx/>；MDX 3 的 await/ES2024 等能力见 <https://mdxjs.com/blog/v3/>。编译器版本、宿主模块/组件解析、浏览器执行必须分别验证；使用版本号绝不等于完整兼容。

先读真实 native block 模型和 MDX AST，确认 `EmbeddedBlock` 只能替换整块而没有原生行内 runs。选择有边界的增量：独立 flow expression、IIFE 内局部变量、array/map JSX、含表达式/组件的小写 JSX；顶层段落含 active MDX 时只提升该段落为岛屿，其他普通正文保持原生。不是整页 WebView，也不把“按段提升”伪称原生行内控件。标题/列表/引用嵌套仍明确报边界错误。document ESM/共享作用域、async/await、import.meta 与模块资源仍未支持；把 exports 复制到每个 WebView 会错误模拟共享状态，因此没有这样做。

新增 Tabs/TabItem/Accordion 为本仓库独立实现，未搬终端 mdx 仓库源码；它们是宿主组件，不当作版本兼容的替代指标。用户文档表达式仍只在 WebView 执行，可信 compiler 在 moejs 中只处理输入字符串；不绑定 Go/文件/网络，不改 CSP/预算。正常无横幅；已有 theme/height IIFE 和观察者处理保持。

建立 `tools/mdx/compatibility.json` 可执行分层矩阵：开发测试先用官方精确版本验证 fixture 语法，再调用真实 Go 主程序（PATH 无 Node）检查 host 接受/拒绝与生成 HTML，最后 JSDOM 验证输出/事件/组件错误。预期拒绝项通过不算功能支持；JSDOM 不算 Mac WKWebView 验收。新增 deterministic build-info，记录 exact pins/profile/锁文件与生成资源 SHA256；Go 实际校验内嵌资产而非只加版本标签。旧 `experiments/moejs/build.mjs` 转交 canonical 工具，避免旧 source 重新生成覆盖新编译器。

根中英文 README 去掉过期“独立只读实验、主程序未改”和全局“零 WebView”的矛盾表述，保留上游归属，明确本阶段无新 release。源视图实际是只读导出预览，修改 JSX 仍由 agent/外部编辑器完成后重载。新增 Go 测试对含 Unicode 的 inline/flow expression 执行真实保存、预览还原、外部改源后重载及 source SHA/一次创建视图的验证。版本、用法、限制与下一阶段详见 `docs/mdx-compatibility.md`。

### 本轮实际验证（MDX 3 第一阶段）

- `npm run build --prefix tools/mdx` 重复生成的 compiler/runtime/metadata 三项字节一致；metadata SHA256 为 `fd15b48bc5fb3ae6f4e55c02775140dfa7e9674d13f2e3c3bb0310613f92114d`。
- canonical `npm test --prefix tools/mdx` **44/44**：16 个回归、26 个分层矩阵 fixture、元数据和宿主组件集成测试。26 项状态分布 supported 12 / partial 5 / unsupported 9；预期拒绝也是通过，绝非 44 项语言能力全部支持。涵盖真实 Go 无 Node 编译、JSDOM 输出/用户事件/Tabs 键盘及 Accordion/主题状态。
- 额外新建从未预编译的含 inline/flow 的文件，真实 `cmd/arcmd --jsx-check` 在仅设置 `HOME` 与 `PATH=/no-node` 的空环境运行成功，输出 exact compiler/profile、2 个 HTML 岛屿及匹配源 SHA，原文件未改动。
- `go test ./internal/jsx ./cmd/arcmd`、相关 app JSX/MDX/open/save/reload 测试、`go vet ./...`、`go mod tidy -diff` 通过；Darwin arm64 / Linux amd64 / Windows amd64 的 `CGO_ENABLED=0 go build ./...` 通过。`go run ./tools/ci version` 仍为 0.2.2，未改应用标识或发布配置。
- 全量 `go test ./... -count=1` 仅已有六项 app baseline 失败：`TestBottomBlockHandleMenuAndIndentSubmenu`、`TestRebindShortcuts`、`TestLibrarySpaces`、`TestLibraryCardDrag`、`TestQACitationPreviewDoesNotOccludeAnswerText`、`TestMoreMenuAndSettings`。未修复或算作本次回归；其余包通过。
- 原生 `Snapshot` 测试通过，检查 light/dark 阅读图无异常；这些是原生正文截图，不是实际 WKWebView 验收。旧 jsx-webview npm 测试 15/15、旧 moejs runtime 测试 3/3 通过，亦不替代本轮新能力验证。
- 本轮没有真实 Mac / Finder / WKWebView 验收，没有新 Release，也不改变 main、维护 state 或已有上游贡献分支。


## 0.2.3 主线发布

用户确认 PR #3 / `2283afd` 测试包效果可接受，授权普通 merge 和新正式 Release。保留精确 MDX 3.1.1 / phase1 编译和渲染实现，只更新主线 0.2.3 版本、正式 README/包内说明及打包元数据/示例。稳定 ArcMD.app / org.ybbapp.arcmd / 正常配置不变。主线打包继续内嵌 moejs，无运行时 Node，纳入 canonical 新功能示例和精确 compiler build metadata；不是预编译示例方案。新 main ZIP 未再次真实 Mac 验收，用户验收来源仍单独标注为上述分支测试包。保留所有实验记录与 Release，既有上游贡献分支不变。实际测试/源码 SHA/产物 SHA256 随正式 Release 记录；维护状态仅在公开下载校验成功后原子更新，沿用 repo.lock 整事务互斥，上游 Release 基线保持 v0.2.0 / 447020319。
