# Fork 变更记录（基于上游 v0.3.1）

## v0.3.1-uncut.5

### 新增
- `internal/aistudio/anticensor.go`：`-uncut` / `防外审截断-` 模型解析、请求改写、事件还原、输入校验
- `internal/aistudio/anticensor_test.go`：请求链/补位/校验/过滤器单测
- `internal/api/anticensor.go`：模型目录派生 `-uncut` 别名
- `internal/app/memory_cleanup.go`：智能清理（Go 堆回收 + 空闲 Worker 内存回收）

### 修改
- `internal/aistudio/generate.go`：解析后缀、校验、构建包装请求、事件过滤、丢弃客户端工具
- `internal/aistudio/service.go`：账户选择/冷却使用基础模型，`-uncut` 强制 Playground
- `internal/aistudio/types.go`：`GenerateRequest.AntiCensor` 内部标记
- `internal/app/runtime.go`：剥离后缀并设置标记、清理配置注入、启动清理循环
- `internal/app/worker_manager.go`：空闲 Worker 识别与页面重载/关闭辅助
- `internal/config/config.go`、`internal/api/admin.go`、`internal/app/admin.go`：清理配置项
- `internal/api/openai.go`、`internal/api/gemini.go`：模型列表返回 `-uncut` 别名

### 文档
- `FORK.md`、`docs/uncut.md`、`docs/smart-cleanup.md`、`docs/debugging.md`
- `scripts/build-linux.sh`

## 上游基线

- 上游仓库：https://github.com/Mag1cFall/AIStudio2API
- 基线 tag：v0.3.1
