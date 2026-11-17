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
