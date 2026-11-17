# Fork 同步与主程序打包约定

仅写入 `ybbapp/ArcMD`，上游 `Ciao1019/ArcMD` 只读。每小时流程由外部 routine 调用；本文件和打包脚本本身不安装 cron。

## 并发与状态

- 所有 git fetch/merge/checkout/push、PR 合并、构建、发布及状态写入必须先持有同一把锁：`/workspace/arcmd-maintenance/repo.lock`（`flock -n`，锁忙则本小时跳过）。持锁覆盖整个事务，不可只锁 fetch。
- 本机 checkout：`/workspace/arcmd-jsx-experiment`。
- 状态：`/workspace/arcmd-maintenance/state.json`，写入同目录临时文件，fsync 后 `os.replace` 原子替换。`schema_version=1`。
- `last_successful_sync_sha`：成功同步并 push 的 fork main SHA；`last_successful_sync_upstream_sha`：当次上游 main SHA；`last_successful_sync_at`：UTC+8 ISO 时间。
- `last_packaged_upstream_release`：仅成功发布并读回资产后写入，含 `id/tag_name/html_url`。检测该 ID/tag 是否变化避免同一上游 Release 重复发布。首次手动发布也记录该基线。
- `fork_release`：最新本 fork 发布的 `tag/url/id/source_commit/assets/sha256`。
- `build`：打包脚本和命令；`lock_path`：同一锁的绝对路径。
- 同步成功而构建失败时可只更新同步字段，绝不能推进打包标记。发布中断时先查询确定性 tag 是否已存在，核对 source SHA 与资产；缺资产或上传失败不冒充成功，报告并等待处理，不覆盖已发布资产。

## 每小时策略

1. `flock -n /workspace/arcmd-maintenance/repo.lock ...`，检查工作树干净；不干净则停止并报告，不 reset/stash 用户内容。
2. 读取上游默认分支及最新 Release（只读 API），`git fetch origin && git fetch upstream`。`main` 正常 fast-forward 到 origin；再 `git merge --no-edit upstream/main`，保留 JSX 实现。普通 merge 解决明确冲突，不 rebase、force push 或重写已推历史；不确定/语义冲突停止报告。
3. push 仅 `origin main`，原子更新同步状态。除非另获授权，不自动合并用户新建的 PR（本次 #1 已明确授权并合并）。
4. 上游 Release ID/tag 未变化：只更新 main 与同步状态，不构建、不发布。
5. 有新上游 Release：运行相关测试，区分既有 6 个 app 基线失败；新回归停止。使用下面脚本构建主程序，发布到 fork，再读回 Release 与 assets。成功后原子推进打包标记。

## 可复用构建命令（持锁时执行）

```sh
gh api repos/Ciao1019/ArcMD/releases/latest \
  --jq '{id,tag_name,target_commitish,published_at,html_url}' \
  > /workspace/arcmd-maintenance/upstream-release.json
npm ci --prefix experiments/jsx-webview --no-audit --no-fund
npm test --prefix experiments/jsx-webview
go test ./internal/jsx ./cmd/arcmd-jsx ./internal/editor ./cmd/arcmd
go vet ./internal/jsx ./internal/app ./internal/editor ./cmd/arcmd ./cmd/arcmd-jsx
# internal/app 全量还须记录那 6 项已知基线失败；不能把它们算为新增回归。
python3 experiments/jsx-webview/package_main_macos.py \
  --out /workspace/arcmd-maintenance/releases \
  --upstream-sha "$(git rev-parse upstream/main)" \
  --upstream-release-json /workspace/arcmd-maintenance/upstream-release.json
```

仅从干净、已提交且已经 push 的源码构建。脚本拒绝覆盖同名 ZIP；依赖 Node/npm、Go、rcodesign；只产物不发布。手动 `gh release create ... --repo ybbapp/ArcMD --target <source SHA>`；标题/说明中文，包含基线、上游/源码提交、SHA256、测试与真实 Mac 尚未验收、Node 依赖和应用签名限制。非 `v*` tag 避免误触上游签名更新 workflow；此 fork 包不提供上游签名更新清单。

本轮主程序包仍属预发布质量，Release 应标记 prerelease。只有用户真实 Mac 验收后才可声称可用性经过 Mac 验证。
