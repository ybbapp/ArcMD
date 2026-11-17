# ArcMD 0.2.1 · ybbapp 主线版（moejs）

这是 `ybbapp/ArcMD:main` 的正式 fork Release，真正的 `cmd/arcmd` 主程序，不是独立实验窗口。运行时无需 Node/npm/esbuild，也不需要之前的 `Open-with-JSX.command`。适用 macOS 13+、Apple Silicon（arm64）。

## 验收来源与本包区别

用户已确认 [moejs-test-8dc66bb](https://github.com/ybbapp/ArcMD/releases/tag/moejs-test-8dc66bb) 测试通过，并要求合入及发布；[PR #2](https://github.com/ybbapp/ArcMD/pull/2) 以普通 merge 合入 main。本包沿用验收过的 moejs 编译/渲染实现，并恢复正式 ArcMD 名称、版本与主线配置目录，更新主线打包入口。**新生成的主线包尚未在真实 Mac 再次测试**，不能把实验包验收当成本 ZIP 已实机验收。

## 安装与使用

1. 解压后将 `ArcMD.app` 放入“应用程序”；退出正在运行的旧 ArcMD 后再替换。主线沿用原先 `ArcMD` 配置与默认空间目录，不再使用实验版的 `ArcMD MoeJS Test` 独立配置；实验版设置不会自动迁移。建议先备份重要文件再更新。
2. 整个应用已做 ad-hoc 签名，仍无 Developer ID、未公证。首次只信任这份应用：Control 单击→打开，或在“系统设置→隐私与安全性”允许。必要时仅对这一份已校验的应用执行 `xattr -d com.apple.quarantine "/Applications/ArcMD.app"`，不要关闭 Gatekeeper/SIP。
3. 可直接从 Finder 启动，不必安装 Node、不必在终端设置 PATH。打开本地 `.mdx`；随包 `demo.mdx` 有 Counter 和嵌套 Callout，可复制、改数字/文字后重新打开，确认针对新文件编译。普通正文仍由原生编辑器排版。
4. 点击计数器再切换浅/深主题，主题更新不重载组件；高度按内容测量。修改后重新打开/重载才重新编译；不保证交互状态跨保存保留。
5. 编译、逐岛屿编译及 WebView 创建/加载失败会显示诊断；未知组件会在岛屿中显示 `JSX rendering failed` 和源码。源码 fallback 不代表渲染成功。干净文档可重新载入并重试。

## 本地诊断（无需 Node）

```sh
/Applications/ArcMD.app/Contents/MacOS/ArcMD --jsx-check "/path/to/file.mdx" > /tmp/arcmd-manifest.json
```

输出当前文件的编译 manifest/HTML；编译错误返回非零并写 stderr。未知组件是浏览器渲染期错误，仍要在界面查看。输出含原文，只写入你指定的本地文件。

## 范围与安全限制

- 只支持当前块级 JSX 岛屿，不是完整 MDX；不支持文件 import/完整模块系统、行内 JSX、跨岛屿共享状态，未改只读标题折叠。仅 `.mdx` 启用该编译流程。
- 构建开发时可用 Node 生成固定 compiler/React bundle；安装包运行时通过内嵌 moejs 编译任意受支持的新文件，不执行外部 Node，也不是预编译 demo 冒充动态支持。
- moejs 仅运行可信编译器，将用户输入作为字符串解析，不执行文档表达式。无 Go/文件/网络绑定，禁动态 eval；5 秒编译截止、256 MiB JS 分配预算、512 KiB 源文件、32 岛屿及 16 MiB 返回上限。这不是 OS 进程总内存硬隔离。
- 文档表达式只在 Untrusted WebView 运行；继续无 Go bridge、hash CSP、禁网络/导航。渲染投影不写回用户源文件，保存时还原命名空间围栏。
- 本 fork 包不配置上游自动替换更新；手动从 ybbapp/ArcMD Release 下载。小时任务更新仓库/发布包，不是用户机器上的自动安装器。

## 测试边界

Linux 上验证真实二进制在无 Node/PATH 环境动态编译新文件，相关 Go 测试/go vet、JSDOM 交互、资源预算/中断、源 round-trip 与可见诊断；交叉编译 macOS arm64、整包签名并校验 ZIP/权限/资源哈希。`internal/app` 保留原有 6 个基线失败，详见 Release 记录。新主线 ZIP 的 Finder/WKWebView/Gatekeeper 行为仍需真实 Mac 验证。
