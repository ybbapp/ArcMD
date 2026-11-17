# 为什么可能同时看到多个 ArcMD

本记录只核查本仓库打包脚本和 box 上留存 ZIP 的 Info.plist，没有读取用户 Mac、没有枚举或删除用户本机应用。下面说明曾发布的应用身份，不表示已确认用户实际安装的数量/路径。

## 已核对的包

| 发布类型 | ZIP 内应用名称 | CFBundleIdentifier | 配置与单实例 |
| --- | --- | --- | --- |
| JSX 预览 c54f189 / 8ebbbe8 / 326edc7 / 59c688b | `ArcMD JSX Preview.app` | `org.ybbapp.arcmd.jsx-preview` | 独立只读实验入口，不走 ArcMD 主程序的空间配置；没有显式调用主程序的单实例锁。MyGo 默认按应用名保存窗口等状态。 |
| moejs-test-8dc66bb | `ArcMD MoeJS Test.app` | `org.ybbapp.arcmd.moejs-test` | 默认 `~/Library/Application Support/ArcMD MoeJS Test`，默认空间 `~/Documents/ArcMD MoeJS Test`；与主线独立的单实例身份。 |
| 早期主程序 bc5ad65（0.2.0）、正式 63fbf16（0.2.1）、本次补丁 0.2.2 | `ArcMD.app` | `org.ybbapp.arcmd` | 正常主线配置 `~/Library/Application Support/ArcMD`，默认空间 `~/Documents/ArcMD`；主线保持同一个应用名和单实例身份。 |

macOS 路径是默认值；显式设置 `ARCMD_HOME` 时会使用该设置，不应因此删除配置。

预览版与 Test 版为了实验隔离，确实采用了不同名称和 bundle ID，所以把正式版 `ArcMD.app` 覆盖到 `/Applications` 并不会自动卸载这两类应用。不是每次补丁都新建身份：0.2.1 与 0.2.2 保持相同主线名称、ID、配置路径。

此外，上游原版配置中使用 `com.cizai.arcmd`，本 fork 正式包使用 `org.ybbapp.arcmd`；只有用户确实装过两者时，这个区别才与其本机情况有关。

## 单实例与“仍打开旧版本”

`internal/app/app.go` 调用 MyGo 的 `RequestSingleInstanceLock`。MyGo v0.7.4 的 `singleInstanceKey` 使用应用名及打包应用的 bundle ID，不是单纯按文件名/安装路径锁定；名称/ID 不同的测试身份不互斥。同一正式身份已经运行时，再启动新进程会把参数转给原进程并退出，所以更新后应先彻底退出旧版再打开新版。

同一个名称/ID 的应用仍可能在不同解压目录或 Downloads、Applications 等不同路径各留有一份；ZIP 解压不是自动卸载器。macOS 的 LaunchServices /“打开方式”旧记录也可能显示额外条目。**未读取用户 Mac 清单，不能把这些可能性当成已经确认的本机事实，也不能认定覆盖安装必然失败。**

## 稳妥更新方式

1. 保存工作，正常退出仍在运行的 ArcMD / 预览 / Test 进程。
2. 解压新包，把其中的 `ArcMD.app` 明确拖到 `/Applications`，替换同一路径已有的正式 `ArcMD.app`，不要选择保留两个；之后从这个路径启动。
3. 在“关于/设置”确认版本 0.2.2，不要再从旧下载解压目录或旧 Dock 快捷方式误开其他副本。
4. 实验/Test 配置和数据不会自动迁移到主线；需要时从主线打开对应文件/文件夹。不要为了清理版本删除文档、空间或配置目录。
5. 若仍显示多个，先区分是真实 `.app` 文件、正在运行的进程，还是“打开方式”菜单条目，再按完整路径核对；本次没有自动访问、迁移或清理用户 Mac。

证据入口：`experiments/jsx-webview/package_macos.py`、实验提交 `8dc66bb` 的 `internal/app/app.go`、`experiments/moejs/package_macos.py`、当前 `experiments/moejs/package_main_macos.py`。留存 ZIP 的 plist 已逐一只读核对；新补丁包另外再次断言主线 name/ID。
