# ArcMD 0.2.14 · ybbapp 主线版（上游官方 v0.5.0 + MDX 3.1.1 第一阶段）

这是 ybbapp/ArcMD:main 的正式 fork Release，真实 cmd/arcmd 主程序，macOS 13+ / Apple Silicon（arm64）。运行不需要 Node/npm/esbuild，没有内置 Node 或本地服务。

本包对应上游官方 Release [v0.5.0](https://github.com/Ciao1019/ArcMD/releases/tag/v0.5.0)（ID 459523665，源码 `8e62d8dcb591c114a126291a40b05df80ad42810`）。本 fork 已发布过 0.2.1–0.2.13，因此本包版本为 0.2.14，避免版本号冲突。

## 上游 v0.5.0 能力（已纳入本包）

- MDZip（`.mdz`）v1.2.0 单文件打包导出、原地只读阅读与一键解包（Ciao1019/ArcMD#79、#85，感谢 `@tajiaoyezi`）：纯 Go（零 cgo、零磁盘临时解压）编解码 MDZip v1.2.0。可通过「⋯ › 导出 › 导出 MDZip (.mdz)」或命令栏，把当前文档连同封面、本地图片、内嵌白板/PDF/书籍及本地非 Markdown 附件打成单个 `.mdz`；也可直接只读打开、在多页 project 模式中浏览包内文件，并一键解包另存为文档。已接入系统文件关联、拖放导入、悬停预览、空间检索、空间问答与本地 MCP/CLI `doc_read`。本 fork 中 MDZip 只读归档与既有 JSX 岛屿、HTML/EPUB Untrusted WebView、数据库 Linked View 生命周期并存。`mygo.json` 保留上游新增的 `.mdz` 关联，identifier 仍为 org.ybbapp.arcmd。
- 原生可编辑 Markdown 源码模式（#37）：将原只读导出预览升级为可编辑源码模式（⇧⌘M /「⋯」菜单 / 命令栏），支持语法高亮、页内查找替换、Tab 缩进/反缩进，以及可拖拽比例、可左右对调的「源码 + 富文本实时预览」分栏，并可将原始 Markdown 文本直接保存（修改不兼容语法时不被富文本重排改写）。本 fork 进入源码时在 ExportMarkdown 之后经 `RestoreSource` 还原 JSX/MDX 原文；从源码写回前会 `closeJSX` 并 `projectJSX`，再 `ReplaceDocument`。导出与 `text()` 在源码缓冲已初始化时使用源码预览，否则在数据库文件之外仍经 `RestoreSource` 还原原文。源码 fallback 不能当作 JSX 运行通过。
- YAML Frontmatter 可视化卡片（#83）：普通 Markdown 文档顶部可折叠、带 YAML 高亮的元数据卡片，支持复制、内联编辑、增删并纳入撤销栈；数据库关联记录页自动隐藏，只读模式与 `.mdz` 归档保持只读查看与复制。本 fork 中该卡片与 JSX 叠加层一并绘制。
- 资料库点击切换空间与跨空间拖拽整理（#81），以及 LaTeX 兼容扩展与大纲标题公式渲染（#78）。
- 修复超 800 块长文档目录跳转闪退（#76，多轮收敛排版）、浮层菜单悬停穿透编辑区（#77），以及 macOS 视图菜单抢占缩放/重载快捷键和若干 MDZ 边界问题。本 fork 在编辑区绘制时先画 frontmatter 控件，再画嵌入的 JSX 叠加层。
- 源码已于 2026-11-26 22:24 UTC+8 合入本 fork main（合并提交 `dbe2ae1`）；官方 GitHub Release 已于 22:29 UTC+8 发布（ID 459523665，标签 `8e62d8dcb591c114a126291a40b05df80ad42810`），本次据此打包。未把上游 `mygo.json` 中指向 Ciao1019/ArcMD 的应用内更新源写入本 fork。官方多平台安装包与签名自动更新仅适用于上游发行；本 fork Release 仍只提供 macOS Apple Silicon（arm64）。

## 上游 v0.4.0 及更早能力（仍纳入本包）

- 本地编码 Agent 协同（Ciao1019/ArcMD#75）：内置零额外依赖的 `arcmd mcp` stdio MCP JSON-RPC 2.0 服务与 `arcmd space / doc / board / db / skill` CLI 控制子命令；桌面窗口打开时经本机 127.0.0.1 鉴权 IPC 在内存模型上实时更新并保留完整 ⌘Z 撤销，窗口未运行时回退无头模式安全读写本地文件。「设置 › 本地 Agent 接入」可自动探测并一键安装或更新 Claude Code、Codex、Gemini CLI、Antigravity、Cursor 与 OpenCode 的 SKILL.md 与富格式规范。本 fork 的 `--jsx-check` 仍在 CLI 分发之前处理。
- 整页等比缩放与比例药丸菜单（#12、#74，感谢 `@tajiaoyezi`）：支持 ⌘=/⌘-/⌘0（Windows / Linux 为 Ctrl）、滚轮或触控板捏合，在 50%–200% 间按 10% 步进缩放 Markdown 与纯 Go 书籍视图；非 100% 时显示比例药丸；PDF 打印导出恒定 100%。本 fork 中内联 JSX 岛屿不视为 HasZoomModal，非 100% 页缩放下其 WebView 尺寸仍跟随测得内容高度，不会被额外缩放。整页缩放此前已于 2026-11-25 20:13 UTC+8 随 #74 合入本 fork main，本包随官方 v0.4.0 一并发布。
- 一键清除样式（#69、#72，感谢 `@tajiaoyezi`）：可通过选中文本浮动工具栏、正文右键、命令栏或 ⇧⌘\（Windows / Linux Ctrl+Shift+\）清除行内样式与链接，完整保留标题、列表与表格等块结构。
- 侧边栏三档行高密度（紧凑 / 标准 / 舒适）、安静单色文件类型图标、看板 120 Hz 弹簧拖拽重构，以及 EPUB / 纯 Go 书籍阅读器的键盘翻章与滚动。
- 稳定性与沙箱：为数据库封面建立缩略图缓存并减少帧循环磁盘阻塞；筛选态拖拽排序保护；exFAT / FAT32 / SMB 上原子写回退；重命名或移动前释放驻留 PDF 句柄；为 wazero WASM 插件补齐内存上限、上下文终止、日志限流与 ZIP 解压上限。上游还增加了 .excalidraw / .pdf / .epub / .html 系统文件关联元数据；本 fork 安装包仍使用既有经过验证的打包流程与 bundle ID org.ybbapp.arcmd。
- 源码已于 2026-11-25 21:14 UTC+8 合入本 fork main（合并提交 `d3df507`）；官方 GitHub Release 已于 20:53 UTC+8 发布（ID 458111160，标签 `1e7a6f9a6676db0c8c3c9b33b4bb0f104f6327fa`），本次据此打包。未把上游 `mygo.json` 中指向 Ciao1019/ArcMD 的应用内更新源写入本 fork。官方多平台安装包与签名自动更新仅适用于上游发行；本 fork Release 仍只提供 macOS Apple Silicon（arm64）。

## 上游 v0.3.0 及更早能力（仍纳入本包）

- 纯 Go 原生两栖多维数据库引擎（Ciao1019/ArcMD#71）：记录源支持文件夹与内联行；18 种属性类型（含 Relation、Rollup、Formula 2.0）；多层筛选/排序/分组、列统计、条件着色、记录模板、撤销重做，以及 CSV / Notion ZIP 导入与 CSV / Markdown 导出。数据库文件为 `.db.md`，记录为带 YAML Frontmatter 的标准 `.md`，并自带可离线阅读的 GFM 表格快照。
- 9 种纯 Go 原生视图与 120 Hz 微交互（`internal/dbview`）：表格、看板、日历、甘特时间轴、列表、画廊、图表、动态流与表单；遵守系统 `ReduceMotion`。
- 文档内嵌 Linked View 两栖块与三模式记录预览：可通过 `/` 菜单或命令栏嵌入已有数据库视图并调节高度，支持侧边预览抽屉 / 居中弹窗 / 整页打开。本 fork 中该块与既有 JSX 岛屿、HTML/EPUB Untrusted WebView 生命周期并存；测高跳过嵌入的 JSX 岛屿与数据库块；导出与 `text()` 在数据库文件之外仍经 `RestoreSource` 还原 JSX/MDX 原文。
- 统一新建向导与空间生态打通：侧边栏与命令栏可一站式创建文档、数据库、白板、文件夹、套用模板或导入 CSV / Notion；Schema 与记录 Frontmatter 接入全文检索、空间问答与侧边栏重命名/移动。
- 修复排序后表格自动滚到底部，以及整页打开数据库记录时侧边栏伴随文件夹无法收起的问题；补全中英文翻译与第三方归属。
- 源码已于 2026-11-24 22:19 UTC+8 同步进本 fork main（合并提交 `9b61fc3`）；当时官方 GitHub Release 尚未发布。现已发布（ID 456903018，published 22:47 UTC+8），本次据此打包。未把上游 `mygo.json` 中指向 Ciao1019/ArcMD 的应用内更新源写入本 fork。官方多平台安装包与签名自动更新仅适用于上游发行；本 fork Release 仍只提供 macOS Apple Silicon（arm64）。

## 上游 v0.2.8 及更早能力（仍纳入本包）

- HTML 与 EPUB 原生阅读、文档内嵌预览与空间检索（Ciao1019/ArcMD#64）：可在空间侧边栏、外部打开与拖放导入中直接阅读 `.epub`、`.html` 与 `.htm`。EPUB 由纯 Go 解析目录、封面与内嵌图片；HTML 以沙箱化 Untrusted WebView 渲染并限制同目录本地资源。支持将 HTML/EPUB 内嵌到 Markdown 就地浏览，并接入空间检索、悬停预览、PDF 导出与问答引用跳转。本 fork 中该书档 WebView 与既有 JSX 岛屿生命周期并存，互不替代。
- 飞书风格有序列表编号值编辑（#62）：文档模型升级至 v10，支持 1–999999 起始编号，可从序号浮层或块手柄菜单继续编号、开始新列表或修改编号值，并在 Markdown/HTML/思维导图导入导出中保持一致。
- 系统代理自动识别与插件市场失败重试：优先遵循 `HTTP_PROXY` / `HTTPS_PROXY`，否则回退操作系统代理；插件市场请求超时放宽至 30 秒，失败时可重试。本 fork 仍不配置指向上游的应用内自动替换更新。
- Arch Linux 原生 `.pkg.tar.zst` 与 CI 自动构建（#63，感谢 `@Lyrecoul`）：上游 linux-amd64 CI 可构建 Arch 包。该元数据已纳入本 fork 源码，但本 Release 安装包仍只提供 macOS Apple Silicon（arm64）。
- 彻底加固输入 `/` 后粘贴文本及修饰键下的浮层方向键问题（#64）；重写上下方向键跨块纵向导航；统一内嵌白板、PDF、HTML/EPUB 与异常图片块的选中删除体验；思维导图默认快捷键改为 `⌃⌘M`（Windows / Linux `Ctrl+Alt+M`），并修复工具栏角标对齐。
- 源码已于 2026-11-23 19:13 UTC+8 同步进本 fork main（合并提交 `1ca4ce4`）；当时官方 GitHub Release 尚未发布。现已发布（ID 455320913，published 19:19 UTC+8），本次据此打包。

## 上游 v0.2.7 及更早能力（仍纳入本包）

- 修复 macOS 空间内含损坏或异常 PDF 时启动闪退，以及纯 Go PDF 解析器的并发竞态（Ciao1019/ArcMD#57、#60）：`render_darwin.go` 在打开失败或 0 页 PDF 时不再返回带类型的 nil 接口，避免后台搜索索引和打开 PDF 时的 SIGSEGV；并为 rawPDF 解析缓存加锁，加固内容流/CMap 与交叉引用表边界。
- 输入 `/` 后粘贴文本立即关闭斜杠菜单；当 `/`、`[[`、`:`、`@` 过滤无匹配结果时放行方向键与回车，单列过滤列表仍可用左右键移动查询光标（#58）。
- 飞书 / Notion 风格全宽页面模式与右上角页面宽度切换（#41、#59）：可在文档右上角工具栏、专注模式顶栏、「⋯」更多菜单、「设置 › 外观」、命令栏与快捷键 ⌥⌘P（Windows / Linux 为 Ctrl+Alt+P）在窄（680）/ 标准（820）/ 较宽（1040）/ 全宽之间切换；全宽保留左侧块手柄安全留白并避让大纲按钮，再次按下可恢复进入全宽前的栏宽。
- Linux KDE Wayland 将最小化、最大化/还原与关闭按钮融合进主窗口和独立文档窗口顶部工具栏（#56）；不影响 macOS、Windows 与其他 Linux 会话。本 fork 的 macOS Apple Silicon 包走非 Linux 窗口路径，构造上等价于原先的 `mygo.NewWindow`。
- 以上 #56、#58、#59、#60 及版本提交 #61 均随本次同步首次进入本 fork main。

## 上游 v0.2.6 及更早能力（仍纳入本包）

- 原生零 CGO PDF 阅读器、文档内嵌 PDF 块、思维导图编辑与右键菜单，以及外部 `.md`/`.pdf` 拖拽导入（Ciao1019/ArcMD#36、#40、#44）：基于系统原生能力（macOS PDFKit / Linux Poppler / Windows WinRT）做零 CGO 渲染与大纲/文本/全文检索；思维导图支持节点右键菜单、分支聚焦、结构化增删移动并导出 PNG/SVG；可从访达/文件管理器拖入 `.md` 与 `.pdf` 到窗口或侧栏文件夹并支持撤销。
- 打开 PDF 时，若平台后端已打开文档但页面几何不完整或大纲为空，则回退到纯 Go 解析结果填充页尺寸与大纲（#55）。此修复包含在官方 v0.2.6 标签提交中，但不在 #54 撰写的官方 Release 说明正文里。
- 独立远程插件市场与官方「图床上传」插件（#45）：内置插件已拆到 [`Ciao1019/arcmd-plugins`](https://github.com/Ciao1019/arcmd-plugins)，可按需从 `catalog` Release 下载（SHA-256 校验）；图床插件支持粘贴/拖入图片自动上传到 S3 兼容存储或自定义 HTTP 图床并改写外链，失败回退本地并可单步撤销。
- 长代码块与长图片可配置紧凑最大高度及块内滚动（#43、#46），阅读时也可点右上角展开/收起，不改写原始 `.md`，也不影响 PDF 打印。
- HTML 网页导出补齐代码语法高亮，并将 Mermaid 渲染为跟随明暗主题的内联矢量 SVG（#51、#52）；复制到剪贴板的语义 HTML 仍保持纯文本代码与 Mermaid 源码。解析失败的 Mermaid 保留源码。
- 体验修复：去掉侧栏文件缝隙悬停「+ 新建文档」按钮（#47）；正文左侧 `+` 按钮不再被裁切或拦截点击（#48）；白板自由画笔可无框连续绘制（#42、#49）；思维导图有序列表继承全文编号与缩进层级（#50）。
- 其中 #44–#52 此前已同步进本 fork 的 main，并包含在本官方 Release 中。

## 上游 v0.2.5 及更早能力（仍纳入本包）

- 侧栏复制、粘贴与创建副本（Ciao1019/ArcMD#10、#39）：可在侧栏对文档、文件夹、多选和收藏磁贴使用右键或 ⌘C/⌘V（Windows/Linux 为 Ctrl），支持跨文件夹和跨空间粘贴，并保留或重定位附件/封面；副本进入行内重命名并打开到正文，撤销走回收站。
- 更清晰的公式绘制、按内容宽度扩展的表格、窄窗口侧栏展开迟滞，以及设置/命令栏中的可选自动保存（#32）：关闭自动保存后，离开或关闭未保存文档会提示「保存 / 不保存 / 取消」。
- 从表格、行内代码和正文复制时保留字面纯文本，不再带 Markdown 转义或对齐用 HTML 包裹；单词内部的下划线也不再被转义。
- 表格工具栏在行索引超出列数时不再崩溃（#34、#35）。
- 空间问答的流式 Responses 与 Chat Completions 协议选择（#23）已在此前的 main 同步中合入，并包含在官方 v0.2.5 说明中。
- 本 fork 在文档退场路径上同时销毁 JSX WebView（`document.close()` 调用幂等 `closeJSX()`；丢弃未保存编辑时 `replaceReferences` 也会先关闭旧文档的 JSX），避免只回收位图而留下岛屿运行时。
- 修复反复切换 Markdown 文档、侧栏悬停预览与空间时内存持续增长的问题（#33），以及已在 0.2.4 / 0.2.5 / 0.2.6 / 0.2.7 / 0.2.8 / 0.2.9 中打包的上游 v0.2.1–v0.2.4 能力（纯 Go 矢量 PDF、原生 Mermaid、手绘白板、`\boldsymbol`、拼音搜索、Obsidian 图片嵌入，macOS 未安装楷体时避免 CoreText MobileAsset 挂起，`wazero` 插件沙箱与云端同步、三栏合并等）。

## 0.2.3 起已有的阶段兼容（仍适用于 0.2.14）

编译器精确锁定 @mdx-js/mdx 3.1.1，宿主 profile 为 arcmd-mdx3-phase1，不宣称完整 MDX 3 兼容。

- 原生 Markdown 正文，块级 JSX/表达式使用隔离 WebView。新增独立表达式、数组/map JSX、局部 IIFE、小写 JSX 容器、expression/spread props 和事件。
- 顶层段落内行内 JSX/表达式部分支持：整个受影响段落变成一个岛屿，不是原生 inline run 或整页网页；普通段落不受影响。
- 新增独立实现的 Tabs/TabItem/Accordion，保留 Counter/Callout。随包 demo.mdx 包含行内42、数组1/2/3、局部41、Counter初值5和嵌套Tabs/Accordion。
- 主题跟随且不重载组件，内容测高、源码保存还原、错误诊断/重试保留；正常没有技术横幅。

## 验收来源

用户反馈 [PR #3](https://github.com/ybbapp/ArcMD/pull/3) 的 [2283afd 测试包](https://github.com/ybbapp/ArcMD/releases/tag/mdx3-phase1-2283afd) 效果可接受，并已合入 0.2.3。该测试源码为 2283afdbfe3321a5630c48e2f085fec4d5a69310。**新生成的 0.2.14 主线 ZIP 未在真实 Mac/Finder/WKWebView 验收**，不能把先前测试包验收或 0.2.3 / 0.2.4 / 0.2.5 / 0.2.6 / 0.2.7 / 0.2.8 / 0.2.9 / 0.2.10 / 0.2.11 / 0.2.12 / 0.2.13 包当作本包已实机验收。

## 安装与使用

1. 保存文档并完全退出所有旧 ArcMD 进程，将本包 ArcMD.app 明确替换到 /Applications/ArcMD.app，从这个路径启动。稳定名称 ArcMD.app、bundle ID org.ybbapp.arcmd、正常配置与单实例均不变；不新增测试产品身份，也不会自动删除文档、配置或其他应用副本。建议先备份重要文档/配置。
2. 全包 ad-hoc 临时签名，无 Developer ID，未 Apple 公证。若系统拦截，请按 macOS 针对单个应用的安全提示决定是否允许运行；不要关闭 Gatekeeper/SIP。
3. 可从 Finder 启动，打开本地 .mdx（空间/侧栏/系统打开均可）。先试同包 demo.mdx，再让agent改一个新文件验证动态编译。普通正文仍原生。
4. JSX/受提升段落仍建议让 agent 或外部编辑器改原文件，再回应用重载，以便走内嵌编译器动态编译。本包已纳入上游可编辑 Markdown 源码模式（⇧⌘M /「⋯」菜单 / 命令栏）：进入源码时在 ExportMarkdown 之后经 RestoreSource 还原 JSX/MDX 原文再编辑；从源码写回前会 closeJSX 并 projectJSX，再 ReplaceDocument。源码模式支持语法高亮、查找替换、Tab 缩进与可拖拽分栏实时预览，并可将原始 Markdown 文本直接保存（避免不兼容语法被富文本重排改写）。保存还原岛屿原文，不写回 HTML；岛屿外 Markdown 在非源码保存路径上仍可能被原生序列化规范化。源码 fallback 不是 JSX 运行通过。
5. 切主题不重载组件；重新打开/重载可重置状态。非活动Tab面板卸载，不承诺状态跨切换保留。编译/WebView/未知组件失败应显示诊断和源码，可重试；源码fallback不是成功。故意错误文件 intentional-error-unknown-component.mdx 用于检查渲染诊断。
6. 不同目录副本和macOS旧登记可能造成“打开方式”重复条目，不能仅凭菜单确定本机副本数量。本包不主动清理用户Mac。需要恢复上一正式版可退出后重装 https://github.com/ybbapp/ArcMD/releases/tag/arcmd-v0.2.13 ，不需要删除数据。

## 范围与安全

不支持 import/exports 模块解析、文档共享作用域、用户组件模块、跨岛屿状态、dynamic import/import.meta/await，以及岛屿外标题/列表/引用中的行内MDX。空表达式仍有最小留白，完整MDX插件/引用上下文/资源解析没有实现；不是完整MDX支持。不改只读标题折叠。

JSX 编译器定位：`@mdx-js/mdx` 3.1.1 / `arcmd-mdx3-phase1` 由 `internal/jsx/assets/mdx-compiler.js` 经 moejs 编译后 **go:embed 进 ArcMD 主程序**。运行时通过内嵌 moejs 动态编译受支持的新文件，不是预编译示例冒充支持。开发者重新生成编译器/React资源和运行npm测试需要Node；普通Go构建用提交的资源，应用运行不需要Node，也不把源码 fallback 当作 JSX 运行通过。

可信编译器在moejs中只把文档作为文本解析，不执行文档表达式；无Go/OS/文件/网络API、无模块加载器或动态eval。保持5秒求值截止、256MiB JS分配、512KiB源文件、32岛屿、16MiB/200000节点返回限制；固定编译器初始解析在求值期限之前，这些不是进程总RSS限制。

文档表达式只在Untrusted WebView执行，无Go bridge，hash CSP、禁网络/导航保持。编译预算不是浏览器不可信表达式的CPU/内存沙箱承诺。现有Linux WebView设置不放宽。本fork不配置上游自动替换更新，小时维护不是用户电脑上的自动安装器。

## 诊断与测试边界

终端只读编译诊断（无需Node）：
/Applications/ArcMD.app/Contents/MacOS/ArcMD --jsx-check /绝对路径/文件.mdx

输出包含compiler精确版本/profile/build SHA、源SHA、岛屿与诊断。编译/部分不支持返回非零；未知组件等浏览器错误仍要在界面检查。输出含原文，请只保存到自己的本地文件。

canonical编译器/浏览器测试与相关Go/vet、无Node新文件动态编译及交叉构建将随本次 Release 记录。完整Go测试若仅保留既有六个已知Linux app baseline失败及既有 Issue 19 Linux 环境失败，不借机修复。JSDOM和交叉构建/ZIP/签名结构核查不替代真实WKWebView与macOS安全验证。详见Release测试记录和docs/mdx-compatibility.md。
