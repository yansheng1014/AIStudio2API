# AIStudio2API —— 个人分支（v0.3.15 + 抗截断 + 智能清理）

本仓库是 [Mag1cFall/AIStudio2API](https://github.com/Mag1cFall/AIStudio2API) 的个人分支，当前基线 **v0.3.15**，并保留此前验证过的自定义特性。

不包含任何账号、Cookie、API Key、服务器地址、域名、代理节点或密钥。

## 特性

### 1. `-uncut` 抗截断（防外审截断）

把对话改写成单工具调用链，模型把完整回答写进 `emit_complete_response(display_stream)`，服务端再还原正文。

- 触发：模型名后缀 `-uncut`（兼容 `防外审截断-` 前缀），可叠加思考档位，如 `gemini-3.8-flash-uncut(high)`
- 四协议自动生效：OpenAI Chat/Responses、Anthropic Messages、Gemini generateContent/streamGenerateContent
- 包装模式下客户端 tools 自动隐藏、上游 reply tool 关闭；日志 `uncut 包装完成` 含 `tool_chars/plain_chars/fallback/finish_reason`
- 限制：纯文本轮次；图片/工具/Build 通道不支持；上游硬策略拦截无法绕过

实现：`internal/aistudio/anticensor.go`、`anticensor_test.go`、`internal/api/anticensor.go`，及 generate/service/types/runtime/api 挂接。

### 2. 智能清理（内存治理）

- Go 堆超阈值：`runtime.GC + debug.FreeOSMemory`
- 服务 cgroup 内存超高水位且 Worker 空闲：关闭空闲 Worker 释放浏览器内存（下次请求冷启动）
- 超过紧急水位：同样动作，WARN
- 只有空闲 ≥ `CLEANUP_IDLE_SECONDS` 且无活动请求才处理；两次动作最小间隔 `CLEANUP_ACTION_GAP_SECONDS`

配置从进程环境或工作目录 `.env` 读取：

```ini
CLEANUP_ENABLED=true
CLEANUP_INTERVAL_SECONDS=60
CLEANUP_MEMORY_HIGH_MB=1536
CLEANUP_MEMORY_CRITICAL_MB=2560
CLEANUP_IDLE_SECONDS=900
CLEANUP_ACTION_GAP_SECONDS=600
CLEANUP_GO_HEAP_MB=128
```

### 3. v0.3.15 上游更新（已合入）

- 换号间隔 / 额度换号上限、模型可用账户等新设置；大量设置改为下一次请求生效
- gemini-4 账户迁移到 Interactions 后不再 400；长回复经 reply tool 分段交付
- Firefox 内核浏览器导入、Camoufox 空闲重载修复、版本检查走 Release/atom
- 依赖更新：Go 1.27.3、SpiderMonkey v158.0.1-aistudio2api.2、libwebp v1.6.0-aistudio2api.1

## 构建

```bash
export SPIDERMONKEY_DIR=/path/to/spidermonkey-v158.0.1-aistudio2api.2
export LIBWEBP_DIR=/path/to/libwebp-v1.6.0-aistudio2api.1
export CGO_CPPFLAGS="-I$SPIDERMONKEY_DIR/include -I$LIBWEBP_DIR/include"
export CGO_CXXFLAGS="$CGO_CPPFLAGS"
export CGO_LDFLAGS="-L$SPIDERMONKEY_DIR/lib -L$LIBWEBP_DIR/lib"
npm ci --prefix web && npm run build --prefix web
go build -trimpath -ldflags "-X github.com/Mag1cFall/AIStudio2API/internal/app.version=v0.3.15-uncut.1" -o aistudio2api ./cmd/aistudio2api
```

Linux 用 gcc 时请选 SpiderMonkey 的 `linux-amd64-gnu` 包；Zig/clang 构建选通用包。

## 免责声明

与上游作者无关，仅供学习研究；抗截断会绕过上游内容审查，可能违反服务条款并提高账号风险，请自行评估，勿用于违法用途或对外服务。
