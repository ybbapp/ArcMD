# ArcMD JSX 岛屿：调研与实验记录

本文记录把块级 JSX 接到 ArcMD 原生编辑器的目标、调研结论、实验过程和第一期范围。实现以用户验收通过的最后一次实验预览为基线，不承诺第一期就“完美兼容”全部 MDX。

## 目标

希望形成这样的工作流：**agent 用开放的文件格式写文档，人用舒服的界面阅读和交互，再通过自然语言让 agent 修改文件。**

在 ArcMD 里，这意味着：

- 正文仍由原生编辑器绘制，而不是把整个窗口换成网页。
- 只有需要执行 JSX 的块级组件嵌入系统 WebView。
- 文档以本地 `.mdx` 保存，人不必亲自编辑编译产物。
- 长期目标是更完整的 JSX / MDX 兼容；第一期只做已经跑通的能力。

## 调研结论

### MDX 不管主题

[`@mdx-js/mdx`](https://mdxjs.com/docs/what-is-mdx/) 只把 Markdown 和 JSX 编成组件树。颜色、字体、明暗模式发生在渲染阶段。官方做法是由宿主提供带样式的组件，或让组件读取页面上的 CSS 变量，而不是让解析器读主题。

因此 ArcMD 在运行时把原生 `ui.Theme`（纸张 `Background`、正文 `Text`，以及 Surface / Border / Accent / Danger 等控件色）写成 `--bg`、`--text` 等 CSS 变量，并设置 `color-scheme`。内置 Counter、Callout 只用这些变量，不写死黑白。主题变化时用 `Page.Eval` 更新已打开的页面，不 `LoadHTML`、不重编译，以免丢失组件状态。

### MyGo 已支持嵌入文章的 WebView

ArcMD 使用的 MyGo v0.7.4 已能在原生布局里嵌入 WebView，并提供：

- `PageOptions.Untrusted`：不向页面开放 Go 方法桥。
- `ScrollToContainer`：滚轮交给外层原生正文。
- 随原生滚动容器移动、裁剪，离开视口后隐藏。
- `Page.Eval`：**不受页面 CSP 限制**，适合注入主题和测量高度。

WebView 必须由宿主按文档创建一次，在关闭或替换文档时销毁，不能每帧 `NewWebView`。高度应等于页面内容高度，而不是写死一块空白。

### 安全边界

用户在自己的空间打开本地 `.mdx`，视为允许编译和渲染，但仍遵守：

- Untrusted WebView，不绑定 Go 方法，不提供本机 Shell 或文件桥。
- 编译产物使用 hash-only 的 `script-src` / `style-src` CSP，禁止网络、frame、表单和非空白导航。
- 不要为了主题或测高去加 `style-src 'unsafe-inline'`；宿主通过 `Page.Eval` 写 CSS 变量和读取高度。
- 不要求用户关闭 Gatekeeper、SIP 或任何全局安全设置。
- 第一期不把完整 Node 运行时打进安装包，也不做隐蔽的本地 HTTP 服务。

页面仍有渲染引擎攻击面，这不是恶意代码沙箱。

## 实验流程与验收基线

实验在组织仓库 [ybbapp/ArcMD](https://github.com/ybbapp/ArcMD) 的 `experiment/jsx-webview` 上进行，不改动上游 [Ciao1019/ArcMD](https://github.com/Ciao1019/ArcMD)。独立入口是 `cmd/arcmd-jsx`，用于验证“原生正文 + 局部 WebView”，尚未要求侧栏整合。下面按时间列出关键预览。

| 提交 / 预发布 | 内容 | 验收结果 |
| --- | --- | --- |
| `d106b8b` | 只读 JSX 岛屿原型：原生正文中嵌入 WebView。 | 本地实现，Linux 缺少 WebKit，未做真实界面验收。 |
| `c54f189` / `jsx-preview-c54f189` | 第一个 macOS Apple Silicon 实验预览包。 | 可下载，但未公证，系统会拦截；聊天里直接打开 zip 也会失败。折叠箭头在只读模式下无效（当时未改，本轮也未改）。 |
| `8ebbbe8` / [jsx-preview-8ebbbe8](https://github.com/ybbapp/ArcMD/releases/tag/jsx-preview-8ebbbe8) | JSX 组件跟随窗口浅色 / 深色主题，整包 ad-hoc 签名。 | 主题跟随通过。随后发现每个岛屿固定预留 240 DIP，小计数器下面空出一大块。 |
| `326edc7` / [jsx-preview-326edc7](https://github.com/ybbapp/ArcMD/releases/tag/jsx-preview-326edc7) | 按内容测高。 | **失败。** 两个 JSX 区块都变成红色 `JSX rendering failed: Script error.`，组件被卸掉。 |
| **`59c688b` / [jsx-preview-59c688b](https://github.com/ybbapp/ArcMD/releases/tag/jsx-preview-59c688b)** | 测高脚本改为合法表达式 IIFE，并忽略脱敏的 `Script error.`。 | **用户认为没问题的最后一次实验预览。** 组件重新出现，高度紧贴内容，主题和计数状态保持。本接入以该提交为基线。 |

完整提交：[`59c688b22173362bd0770ed41844ce626914e835`](https://github.com/ybbapp/ArcMD/commit/59c688b22173362bd0770ed41844ce626914e835)。

### 中间踩坑

**固定高度造成空白。** `compile.mjs` 把每个岛屿高度写死为 240 DIP。小 Counter 大约只有 80–100 像素内容，看起来像文档中断了一大截。Callout 更高，所以不那么显眼。不能只改 demo 里的数字，而要按实际内容测高。

**不要用 `document.scrollHeight` 作为唯一依据。** WebView 视口若仍是 240，`document.documentElement.scrollHeight` 和 `body.scrollHeight` 常常是 `max(内容, 视口)`，测不出缩短。应测量 `#root`（或其子元素）的 `getBoundingClientRect()` / `offsetHeight`，再加上 `body` 的上下 padding。`html` / `body` 不要设 `height: 100%`。

**`326edc7` 的 Script error。** MyGo 的 `Page.Eval` 会先把代码当作表达式执行：

```text
try{return JSON.stringify({ok:true,v:await(\n%s\n)})}catch...
```

只有 completion 报错且错误字符串以 `SyntaxError` 开头时，才改用函数体（这时顶层 `return` 才合法）。主题脚本本来就是 `(function(t){...})(payload)`，第一次就能成功，所以 `8ebbbe8` 正常。`326edc7` 的测高脚本是多句加末尾 `return {height, ready}`，在 WKWebView `callAsyncJavaScript` 里是编译期 SyntaxError。Darwin 会把 NSError 交回 Go，但这次语法错误同时触发了页面 `window` 的 `error` 事件；因为页面是 `LoadHTML(..., "about:blank")`，事件常被脱敏成 `message === "Script error."` 且没有 Error 对象。island 运行时的 `window.addEventListener('error', ...)` 把 `#root` 换成了失败源码，两个组件都被毁掉。

**修复方式：** 测高脚本改成和主题脚本一样的单一 IIFE 表达式，第一次就走 `evalExpr`，不要依赖“先 SyntaxError 再当函数体”。观察者回调自行吞掉异常，不得抛到 `window`。页面错误处理忽略没有 Error 对象、且消息为 `Script error.` 或文件名为空的脱敏事件；真正的组件异常仍走 Error Boundary 或带 error 对象的事件。

## 第一期接入范围

已把上述能力接到 `cmd/arcmd` 的真正文档生命周期，而不再只靠独立预览窗口：

- 从空间、侧栏或系统打开本地 `.mdx`（走现有 Markdown 打开路径；`.mdx` 本就在 `vault.Extensions` 中）时，识别块级 JSX 岛屿并嵌入 WebView。正文仍走原生编辑器。
- 打开时若本机有 Node，并且能找到 `experiments/jsx-webview/compile.mjs` 及其 `node_modules`，则调用该编译器生成岛屿 HTML。编译产物可按源文件 SHA-256 缓存在旁路目录，**不会改写用户的 `.mdx`**。保存时把投影里的岛屿围栏还原为原来的 JSX。
- 没有 Node、找不到编译器或编译失败时，回退为普通 Markdown 源码显示，不崩溃。
- 主题跟随与按内容测高沿用 `59c688b` 的表达式 IIFE，不绑定 Go，不放宽 CSP。
- 主题切换不重新加载 WebView，计数器状态应保持。编辑后的重新编译发生在重新打开或从磁盘重新载入时；**不承诺计数器状态跨保存保留**。
- 独立的 `cmd/arcmd-jsx` 仍保留为实验入口。

### 运行时依赖

第一期**不把完整 Node 运行时打进安装包**。要渲染 JSX 岛屿，需要：

1. 本机 `PATH` 上有 `node`（可用环境变量 `ARCMD_JSX_NODE` 指定）。
2. 能找到编译脚本 `experiments/jsx-webview/compile.mjs`（可用 `ARCMD_JSX_COMPILER` 指定），并且该目录已安装依赖（`npm ci --ignore-scripts --prefix experiments/jsx-webview`）。
3. 可选：`ARCMD_JSX_CACHE` 指向缓存目录。

开发时从仓库运行即可满足以上条件。被打包分发的应用若未附带 Node 和编译器，打开 `.mdx` 会显示源码而不是交互组件。

### 明确不做的事

- 文件级 `import` / 完整 MDX 模块系统、任意 npm 包、行内 JSX、跨岛屿共享 React 状态。
- 只读标题折叠（实验预览里点箭头不会收起；用户未选择该项）。
- 向下游 / 上游 Ciao1019/ArcMD 发 PR 或推送。
- 要求用户关闭 Gatekeeper、SIP 或全局安全设置。
- 新的 macOS 实验预览包（接入不必重打预览）。
- 把基线 `d8f9ee7` 上已存在的 6 项 `internal/app` 测试失败借机“修复”。

## 相关代码

- 编译器：`experiments/jsx-webview/compile.mjs`
- 共享主题、测高、编译调用：`internal/jsx`
- 主程序接入：`internal/app/jsx.go`、`internal/app/document.go`
- 实验入口：`cmd/arcmd-jsx`

## 2026-11-17：同步上游、合并接入与主程序发布

- 用户授权每小时同步上游：无新上游 Release 时只同步 `main`；有新 Release 才解决冲突并构建发布。本次用户另行授权立即合并 PR 和发布主程序。
- 本次检查上游 `Ciao1019/ArcMD` 默认分支 `main`，SHA `d8f9ee7767b2104246e1207d869765f850efffd9`；最新 Release `v0.2.0`，ID `447020319`。上游未比 fork 起点增加提交，因此无需冲突处理。
- [PR #1](https://github.com/ybbapp/ArcMD/pull/1) 通过普通 merge 合入 fork `main`，merge commit `1737ba4e2aaec83308141c6e92508b99483468ee`。没有 force push，没有写入上游。
- `package_main_macos.py` 新增主程序专用打包入口：构建 `./cmd/arcmd`，而非 `./cmd/arcmd-jsx`；随包携带编译器、锁定版本的生产 JS 依赖和 **darwin-arm64** esbuild，资源位置能被现有 `FindCompiler` 找到；不捆绑 Node。嵌套 esbuild 与整个应用均做 ad-hoc 签名。
- JSX 运行条件：用户自行安装 Node.js 22+，启动环境能找到 Node（或设 `ARCMD_JSX_NODE`）。附带显式启动脚本处理 Homebrew PATH，缺 Node 时脚本报错，不把源码 fallback 说成渲染成功。独立安装说明见 `main-release.zh-CN.md`。
- 打包时用随包 JS 依赖在 Linux Node 上编译示例（仅此烟测覆盖为 Linux esbuild）；校验目标二进制 Mach-O arm64、ad-hoc 签名结构、ZIP CRC/可执行权限。真实 macOS 的原生启动、系统信任、目标 esbuild 与 WebView 仍需用户验收。
- 无上游自动更新配置或上游更新签名；fork 的包从本 fork Release 手动下载。没有关闭 Gatekeeper/SIP，也没有本地服务。
- 维护锁、状态字段、可复用命令见 [MAINTENANCE.zh-CN.md](MAINTENANCE.zh-CN.md)。Release 附 `BUILD-INFO.json` 记录最终源码提交与上游信息，SHA-256 文件独立附带。
