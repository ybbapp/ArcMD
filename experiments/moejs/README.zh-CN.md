# ArcMD MoeJS 主程序验证版

这是独立实验 PR 的 `cmd/arcmd` 主程序，macOS 13+ Apple Silicon。**运行时无需安装 Node，也不依赖终端 PATH，从 Finder 直接打开。** 应用内嵌 moejs、可信 MDX 编译器及 React 浏览器运行时代码，没有暗藏 Node、npm、esbuild 可执行文件或本地服务。

## 安装与验收

1. 解压后将 `ArcMD MoeJS Test.app` 拖入应用程序。它与 main 版的名称、bundle ID、配置和默认空间目录分离，不要替换 main 应用。它不会自动导入原版的设置/空间，可自行打开测试文件或文件夹。
2. 整包做 ad-hoc 签名；无 Developer ID、未公证。首次只对这份已校验的应用 Control 单击→打开，或在系统设置“隐私与安全性”允许。必要时只对这一份应用执行 `xattr -d com.apple.quarantine "/Applications/ArcMD MoeJS Test.app"`；不要关闭 Gatekeeper/SIP。
3. 直接从 Finder 打开应用，再打开同包的 `demo.mdx`。应看到 “JSX · moejs 内置编译器（无需 Node）”，独立计数器从 5 开始，Callout 内计数器从 9 开始。点击后切换浅/深主题，计数不重置，高度跟随内容。
4. 复制 `demo.mdx` 到新文件，修改数字/文字并重新打开；应渲染新内容。仅 `.mdx` 启用编译，不是所有 `.md` 都执行 JSX。
5. `unknown-component.mdx` 故意含未知组件，应在组件区域显示 `JSX rendering failed` 和原始源码；`invalid.mdx` 故意语法错误，应显示原生诊断和源码回退，而不是无声失败。主窗口干净文档可点“重新载入并重试 JSX”。
6. 版本为 `0.2.0-moejs-test`。这不是完整 MDX：不支持 import/完整模块系统、行内 JSX、跨岛屿共享状态，未改只读标题折叠。编辑后重开/重载才重新编译；不承诺计数状态跨保存保留。

## 可检查的本地诊断

不需要 Node，可在终端运行真实应用二进制：

```sh
"/Applications/ArcMD MoeJS Test.app/Contents/MacOS/ArcMD" --jsx-check "/path/to/file.mdx" > /tmp/arcmd-moejs-manifest.json
```

成功产出当前文件的 manifest/HTML；语法或编译失败返回非零并在 stderr 输出错误。未知组件是浏览器渲染期错误，需在页面查看；编译成功不等于组件运行成功。输出包含文件源码，只保存在你指定的位置。

## 安全与依赖边界

- Node/npm/esbuild 只用于开发构建时生成固定编译器/React bundle；安装包运行时不调用它们。MDX 的 compileSync 已能直接生成 JavaScript，因此这次无需 esbuild Go API。
- moejs 只运行受信任的编译器，用户 `.mdx` 作为字符串数据解析。用户表达式不在 moejs/Go 中执行；编译出的表达式只在 Untrusted WebView 运行。
- moejs 不安装主机函数、importer、Node/文件/网络 API，关闭 eval/Function；每次编译独立 runtime，5 秒截止、256 MiB JS 分配限制、512 KiB 源文件限制、16 MiB 返回值限制、最多 32 岛屿。这不是整个 OS 进程总内存的硬限制。
- WebView 无 Go bridge，保持 hash CSP、禁网络/导航；主题/测高继续用表达式 IIFE，不靠重载更新主题。渲染不会为投影改写磁盘源文件，保存会还原命名空间围栏。
- 编译、逐岛屿编译、WebView 创建/加载失败在主程序显示；组件错误由 React boundary 在岛屿中显示。不把源码回退当成功。

## 尚未验证

已在 Linux 做真实可执行文件的无 Node/PATH 测试和 JSDOM 交互测试，交叉编译 Mac 包、校验 ZIP/签名结构；**没有真实 Mac 的 Finder 启动、WKWebView、签名信任验收**。请以本次实际 Mac 测试为准。moejs 是年轻项目，本版仅预发布验证，不合并 main。
