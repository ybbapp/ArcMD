# ArcMD JSX Preview · macOS Apple Silicon 实验预览版

**这是独立实验，不是完整 ArcMD 产品，不是完整 MDX 阅读器。**
此包在 Linux 上交叉构建，尚未在真实 macOS 上启动或进行 GUI 验收。
请只用于体验随包附带并经过审阅的内置示例。

## 环境与启动

- Apple Silicon（M1/M2/M3/M4 等 arm64）Mac，macOS 12 或更高版本。
- 使用系统自带的 WKWebView、AppKit 和 Metal；无需安装 Go、Node、React、
  WebKitGTK 或额外服务，也不需要联网下载组件。
- 解压 ZIP 后，双击 **ArcMD JSX Preview.app**。也可以将这个 .app 单独移到
  Applications 或其他目录。不要只拷贝 Contents/MacOS 下的二进制。
- 默认仅打开内置示例，不会自动打开、执行或写回你电脑上的其他文档。
- 如需从终端观察错误，进入解压目录后运行：

```sh
"./ArcMD JSX Preview.app/Contents/MacOS/ArcMD-JSX-Preview"
```

路径可包含空格；启动不依赖当前工作目录。内置示例由主程序中固定的 SHA-256
校验，缺失或被修改时会拒绝执行，不能把任意文件放到同名位置冒充内置示例。

## macOS 安全提示与签名状态

本包**没有 Apple Developer ID 发行签名，没有经过 Apple 公证**，也没有公证票据。
Go 链接器为 arm64 主程序生成的 ad-hoc 代码签名不等于开发者身份认证，整个
.app 也没有经过 macOS codesign 的资源封装签名。Gatekeeper 可能阻止首次打开，
不能保证系统会允许在每台 Mac 上运行。

请先核对下载 ZIP 的 SHA-256，并确认你信任来源。如 macOS 提示开发者未经验证，
可依据你当前系统版本提供的“隐私与安全性 → 仍要打开”或 Finder 对该应用的
“打开”机制，仅为这一份应用明确授权。系统不提供此选项时不要强行绕过，
请等待经过 Developer ID 签名和公证的版本。**不要关闭 Gatekeeper、SIP 或全局安全设置，
也不要使用移除整个目录隔离属性的命令。** 本包未向你索取证书、密码或管理员权限。

## 应当看到什么

1. 顶部“切换源码”和“原生浮层”按钮；没有专用的源码切换快捷键。
2. 原生标题和正文之后，第一个计数器初始为 5，点击 Increment 应递增。
3. 后面的 Callout 含真正的粗体 Markdown 和初始为 10 的第二个独立计数器。
4. UnknownComponent 是故意加入的错误示例，应显示错误和源码，不应让整个窗口空白。
5. 再往下有 ArcMD 原生 HTML 语义两栏及用于验证滚动的长篇正文。

请检查：在组件上滚轮是否滚动外层正文；边缘裁剪和缩放是否对齐；滚远再回来、
切换源码两次后计数器状态是否保留；原生浮层是否覆盖网页并拦截点击；焦点是否正常。
这些仍是待验项目，不应把“构建成功”当作已通过 macOS 交互验收。

## 明确限制

- 仍是单独的实验窗口，未整合普通 ArcMD 的 Space、侧栏和文件打开工作流。
- 普通正文原生绘制，仅选定的块级组件嵌入 WebView，不是整页网页。
- 只预置 Counter、Callout；组件块固定为 240 DIP，没有自动测高。
- 不支持任意 MDX 文件开箱即用、任意 npm/local import、文件级共享变量、
  行内 JSX、跨块共享 React context/state；没有文件监听热刷新和 JSX 源码编辑。
- 不支持将网页组件打印或导出为 PDF。不写回原始 .mdx，也不把构建产物当作新文档格式。

## 安全边界

WebView 使用 MyGo Untrusted 模式，不开放 Go 方法或本机 Shell/文件桥。
编译产物使用脚本/样式哈希 CSP，禁用网络、frame、表单提交和非空白导航；
运行内置示例不需要本地 HTTP 服务或云服务。浏览器本身仍有渲染引擎攻击面；
没有 CPU/内存配额，复杂或恶意 JSX 可能卡住页面。**这不是恶意代码安全沙箱。**

## 新文档仅供开发者的流程

新 .mdx 必须先在源码仓库内，通过已安装依赖的编译工具构建成独立 manifest，
审阅源文件与产物后，再显式传入 -trust。普通用户体验内置示例不需要此流程。
不要给未审阅的下载文档授权，也不要期待直接双击任意 .mdx 能运行。

```sh
# 在源码仓库中；构建依赖需要 Go/Node/npm，运行随包示例则不需要。
npm ci --ignore-scripts --prefix experiments/jsx-webview
node experiments/jsx-webview/compile.mjs /absolute/path/reviewed.mdx /tmp/new-reviewed.manifest.json
# 使用解压后的实验程序；输出 manifest 必须使用尚不存在的新路径。
"/path/to/ArcMD JSX Preview.app/Contents/MacOS/ArcMD-JSX-Preview" -trust /tmp/new-reviewed.manifest.json
```

打包脚本：`python3 experiments/jsx-webview/package_macos.py --out /workspace/releases`
它使用固定版本依赖和公开 Go 工具链，不使用私有签名服务；构建要求干净的已提交源码。
示例命名空间随机生成，因此重新构建的 ZIP 不保证逐字节相同。

## 包内容与验证边界

.app 内含完整主程序、Info.plist、图标、预编译 React/MDX 示例、原始 demo.mdx、
构建信息、项目 LICENSE 和第三方许可证。所有资源均以应用包相对路径定位。
根目录另附示例源文件、说明和文件校验表；源码提交及构建记录见 BUILD-INFO.json。

自动检查覆盖：Mach-O arm64 架构与签名结构、系统动态库引用、.app 元数据、
资源完整性、ZIP 可解压性及执行权限、含空格路径的示例加载单测、显式信任和篡改拒绝。
React 点击行为使用 jsdom 验证，原生排版使用无窗口 Tester 验证。
**未做真实 macOS、WKWebView、Finder、Gatekeeper、公证或屏幕交互测试。**
此前完整 Go 测试有 6 项应用测试在未修改基线也失败，具体记录见随包
Developer-notes.md；不能宣称全量测试全绿。
