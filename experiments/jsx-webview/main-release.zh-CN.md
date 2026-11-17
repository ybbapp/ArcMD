# ArcMD 主程序 · macOS Apple Silicon 接入版

这是 `cmd/arcmd` 主程序，不是之前的独立 JSX 演示窗口。适用 macOS 13+、Apple Silicon（arm64）。含原生编辑器与基于验收预览 `59c688b` 的块级 JSX 岛屿；不是完整 MDX 实现。尚未在真实 Mac 验收。

## 安装与打开

1. 解压后，把 `ArcMD.app` 拖入“应用程序”，不要覆盖正在运行的实例。
2. 整个应用及内含 esbuild 使用 ad-hoc 签名；没有 Apple Developer ID，也未公证。首次 Control 单击这一个应用→打开，或在“系统设置→隐私与安全性”中允许这一份应用。必要时仅对这份已校验的应用执行 `xattr -d com.apple.quarantine "/Applications/ArcMD.app"`。不要关闭 Gatekeeper/SIP。
3. 无须 Node 即可使用普通 Markdown 编辑功能。**要渲染 JSX，必须自行安装本机 Apple Silicon Node.js 22 或更高版本；本包没有捆绑 Node。** 编译器、生产 JS 依赖、darwin-arm64 esbuild 已在 `ArcMD.app/Contents/Resources/jsx-webview`，主程序能自动定位，无须仓库、npm install 或本地服务。
4. Finder 启动可能找不到 Homebrew 的 Node。可从终端运行：

   ```sh
   ARCMD_JSX_NODE="$(command -v node)" /Applications/ArcMD.app/Contents/MacOS/ArcMD
   ```

   也可在完成应用信任、安装 Node 后，运行同一解压目录里的 `Open-with-JSX.command`（支持 `/opt/homebrew/bin` 与 `/usr/local/bin`）。该脚本只设置这次启动的环境，不改全局设置。先退出已运行的 ArcMD，再使用这条命令。
5. 在自己的空间或系统文件打开中选择 `.mdx`；附带的 `demo.mdx` 可用于验收。普通正文仍是原生编辑器，Counter/Callout 为块级 WebView；点击 Counter 后切换主题，状态不应重置，卡片应按内容测高。

## 实际依赖与限制

- 没有 Node、Node 不在启动环境中、编译失败时仅显示源码，**这不是 JSX 渲染成功**。先用终端检查 `node --version`，并按上面的显式启动命令重试。
- 无 Go bridge；Untrusted WebView、hash CSP、禁网络。编译期只编译文件，不执行文件里的用户 JS。编译依赖已随包携带，不会运行时下载。
- 不支持 import、完整 MDX 模块系统、行内 JSX；不改只读标题折叠。修改后重新打开/重载触发编译；不保证计数器状态跨保存保留。
- 此包不含上游自动更新签名/配置，需手动下载本 fork 的后续 Release，避免被上游原版覆盖。
- Linux 上完成 Go 交叉编译、Mach-O arm64/可执行权限/ZIP CRC/资源签名结构检查；打包 JS 依赖在 Linux Node 下通过编译烟测（仅烟测覆盖使用 Linux esbuild）。不能代替真实 Mac 的应用启动、系统信任、原生 esbuild 执行与 WebView 验收。
- 测试：JSX npm 15/15，相关 Go 测试及 go vet 通过；`internal/app` 仍有未改基线已有的 6 个失败，详见 `RESEARCH.zh-CN.md`。
