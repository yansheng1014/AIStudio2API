# Fork 变更记录

## v0.3.15-uncut.1（2026-11-28）

- 基线升级：上游 v0.3.1 → v0.3.15
- 保留并适配：`-uncut` 抗截断、智能清理
- 适配点：
  - Client.Generate 新增 interaction/task 路由检查，包装请求拒绝之
  - 包装模式下关闭上游 reply tool（`useReplyTool=false`），避免与合成工具冲突
  - 智能清理配置改为从进程环境或 `.env` 读取，不依赖新版配置热更新结构
  - Worker 空闲判定使用账户池活动状态（常驻 Worker 自带 runtimeLease）
- 新增：`docs/uncut.md`、`docs/smart-cleanup.md`、`docs/debugging.md`、`scripts/build-linux.sh`
- 构建依赖：Go 1.27.3、SpiderMonkey v158.0.1-aistudio2api.2（gcc 用 linux-amd64-gnu）、libwebp v1.6.0-aistudio2api.1

## v0.3.1-uncut.5

- 初版抗截断与智能清理（基于 v0.3.1）
