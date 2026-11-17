# ArcMD 0.2.3 · ybbapp 主线版（MDX 3.1.1 第一阶段）

这是 ybbapp/ArcMD:main 的正式 fork Release，真实 cmd/arcmd 主程序，macOS 13+ / Apple Silicon（arm64）。运行不需要 Node/npm/esbuild，没有内置 Node 或本地服务。

## 0.2.3：明确版本下的阶段兼容

编译器精确锁定 @mdx-js/mdx 3.1.1，宿主 profile 为 arcmd-mdx3-phase1，不宣称完整 MDX 3 兼容。

- 原生 Markdown 正文，块级 JSX/表达式使用隔离 WebView。新增独立表达式、数组/map JSX、局部 IIFE、小写 JSX 容器、expression/spread props 和事件。
- 顶层段落内行内 JSX/表达式部分支持：整个受影响段落变成一个岛屿，不是原生 inline run 或整页网页；普通段落不受影响。
- 新增独立实现的 Tabs/TabItem/Accordion，保留 Counter/Callout。随包 demo.mdx 包含行内42、数组1/2/3、局部41、Counter初值5和嵌套Tabs/Accordion。
- 主题跟随且不重载组件，内容测高、源码保存还原、错误诊断/重试保留；正常没有技术横幅。

## 验收来源

用户反馈 [PR #3](https://github.com/ybbapp/ArcMD/pull/3) 的 [2283afd 测试包](https://github.com/ybbapp/ArcMD/releases/tag/mdx3-phase1-2283afd) 效果可接受，明确要求合并并新增 Release。该测试源码为 2283afdbfe3321a5630c48e2f085fec4d5a69310；本次以普通 merge 合入，再做版本/打包说明与构建元数据修正。**新生成的0.2.3主线ZIP未再次在真实Mac/Finder/WKWebView验收**，不能把先前测试包验收当作本包已实机验收。

## 安装与使用

1. 保存文档并完全退出所有旧 ArcMD 进程，将本包 ArcMD.app 明确替换到 /Applications/ArcMD.app，从这个路径启动。稳定名称 ArcMD.app、bundle ID org.ybbapp.arcmd、正常配置与单实例均不变；不新增测试产品身份，也不会自动删除文档、配置或其他应用副本。建议先备份重要文档/配置。
2. 全包 ad-hoc 临时签名，无 Developer ID，未 Apple 公证。若系统拦截，请按 macOS 针对单个应用的安全提示决定是否允许运行；不要关闭 Gatekeeper/SIP。
3. 可从 Finder 启动，打开本地 .mdx（空间/侧栏/系统打开均可）。先试同包 demo.mdx，再让agent改一个新文件验证动态编译。普通正文仍原生。
4. JSX/受提升段落请让agent或外部编辑器改原文件，再回应用重载；源码视图仍是只读导出预览。先保存/解决原生编辑，再重载。保存还原岛屿原文，不写回HTML；岛屿外Markdown仍可能被原生序列化规范化。
5. 切主题不重载组件；重新打开/重载可重置状态。非活动Tab面板卸载，不承诺状态跨切换保留。编译/WebView/未知组件失败应显示诊断和源码，可重试；源码fallback不是成功。故意错误文件 intentional-error-unknown-component.mdx 用于检查渲染诊断。
6. 不同目录副本和macOS旧登记可能造成“打开方式”重复条目，不能仅凭菜单确定本机副本数量。本包不主动清理用户Mac。需要恢复旧正式版可退出后重装 https://github.com/ybbapp/ArcMD/releases/tag/arcmd-v0.2.2 ，不需要删除数据。

## 范围与安全

不支持 import/exports 模块解析、文档共享作用域、用户组件模块、跨岛屿状态、dynamic import/import.meta/await，以及岛屿外标题/列表/引用中的行内MDX。空表达式仍有最小留白，完整MDX插件/引用上下文/资源解析没有实现；不是完整MDX支持。不改只读标题折叠。

运行时通过内嵌 moejs 动态编译受支持的新文件，不是预编译示例冒充支持。开发者重新生成编译器/React资源和运行npm测试需要Node；普通Go构建用提交的资源，应用运行不需要Node。

可信编译器在moejs中只把文档作为文本解析，不执行文档表达式；无Go/OS/文件/网络API、无模块加载器或动态eval。保持5秒求值截止、256MiB JS分配、512KiB源文件、32岛屿、16MiB/200000节点返回限制；固定编译器初始解析在求值期限之前，这些不是进程总RSS限制。

文档表达式只在Untrusted WebView执行，无Go bridge，hash CSP、禁网络/导航保持。编译预算不是浏览器不可信表达式的CPU/内存沙箱承诺。现有Linux WebView设置不放宽。本fork不配置上游自动替换更新，小时维护不是用户电脑上的自动安装器。

## 诊断与测试边界

终端只读编译诊断（无需Node）：
/Applications/ArcMD.app/Contents/MacOS/ArcMD --jsx-check /绝对路径/文件.mdx

输出包含compiler精确版本/profile/build SHA、源SHA、岛屿与诊断。编译/部分不支持返回非零；未知组件等浏览器错误仍要在界面检查。输出含原文，请只保存到自己的本地文件。

canonical编译器/浏览器测试44/44（含预期拒绝项，不是44项能力全支持），旧套件15/15和3/3、相关Go/vet、无Node新文件动态编译及交叉构建验证。完整Go测试保留6个已知Linux app baseline失败，不借机修复。JSDOM和交叉构建/ZIP/签名结构核查不替代真实WKWebView与macOS安全验证。详见Release测试记录和docs/mdx-compatibility.md。
