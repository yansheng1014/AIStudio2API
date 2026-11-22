# AIStudio2API —— 个人分支（抗截断 + 智能清理 + 调试修复）

本仓库是 [Mag1cFall/AIStudio2API](https://github.com/Mag1cFall/AIStudio2API) 的个人分支，基线为 **v0.3.1**（该版本在我们的部署上长期验证通过）。

这里不包含任何账号、Cookie、API Key、服务器地址、域名、代理节点或密钥；所有配置请在你的本机 `.env` 中自行填写。

## 本分支新增/修改

### 1. `-uncut` 抗截断（防外审截断）

把对话改写成「单工具调用链」，让模型把完整回答写进 `emit_complete_response(display_stream)` 的参数，再由服务端还原成普通正文。

- 触发：模型名后缀 `-uncut`，兼容 pgbridge 风格前缀 `防外审截断-`，可与思考档位组合：
  - `gemini-3.8-flash-uncut`
  - `gemini-3.8-flash-uncut(low|medium|high)`
- 四套协议自动生效：OpenAI Chat Completions、Responses、Anthropic Messages、Gemini generateContent/streamGenerateContent
- 工具：包装模式下客户端 tools 会被自动隐藏（日志 WARN `uncut 模式忽略客户端工具`），避免与合成工具冲突
- 诊断：每次包装结束记录 `uncut 包装完成`，字段含 `tool_chars / plain_chars / fallback / finish_reason / output_tokens / thinking_tokens`
  - `fallback=true` 表示模型没走工具、正文走普通文本路径，仍可能被上游截断
  - `finish_reason=length` 是 token 上限，不是审查
- 限制：仅纯文本轮次；图片/音视频、客户端工具、Build 通道不支持；上游硬策略拦截（`prohibited_content`）无法绕过

实现文件：`internal/aistudio/anticensor.go`、`anticensor_test.go`、`internal/api/anticensor.go`，以及 `generate.go/service.go/types.go/app/runtime.go/api/*.go` 中的挂接。

### 2. 智能清理（内存治理）

长期运行的 Camoufox/Firefox 会持续涨内存；新增内存守卫：

- Go 堆超过阈值：`runtime.GC + debug.FreeOSMemory`（毫秒级）
- 服务 cgroup 内存超过高水位且 Worker 空闲：关闭空闲 Worker，释放整个浏览器内存（下次请求冷启动，实测约 10–13 秒）
- 超过紧急水位：同样动作，日志 WARN
- 保护响应速度：只处理「空闲 ≥ 15 分钟且无活动请求」的 Worker，两次动作最小间隔 10 分钟

配置（`.env`，改完重启生效）：

```ini
CLEANUP_ENABLED=true
CLEANUP_INTERVAL_SECONDS=60
CLEANUP_MEMORY_HIGH_MB=1536
CLEANUP_MEMORY_CRITICAL_MB=2560
CLEANUP_IDLE_SECONDS=900
CLEANUP_ACTION_GAP_SECONDS=600
CLEANUP_GO_HEAP_MB=128
```

实现文件：`internal/app/memory_cleanup.go`、`internal/app/worker_manager.go`、`internal/app/runtime.go`、`internal/config/config.go`、管理接口字段。

### 3. 调试与运维修复

- 上游 Build 通道已停用生成（`PROXY_GENERATION_DISABLED`）：建议 `UPSTREAM_CHANNELS=playground`，并保持 `BUILD_NATIVE_NONSTREAM=false`
- 多 API Key / 每 Key 绑定账户：修改后需要重启进程才生效（管理页有提示）
- 思考链：`low/minimal` 档上游不返回思考摘要；部分模型（如 gemini-4-argon）只有 `high` 档才给；Gemini 原生协议需 `thinkingConfig.includeThoughts=true`
- 构建新增 libwebp 静态库依赖（v0.3.1 起）；国内网络可用 `GOPROXY=https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct`

## 构建

依赖：Go 1.27+、Node.js 22+、gcc、SpiderMonkey 与 libwebp 静态库（见上游 Releases）。

```bash
export SPIDERMONKEY_DIR=/path/to/spidermonkey LIBWEBP_DIR=/path/to/libwebp
export CGO_CPPFLAGS="-I$SPIDERMONKEY_DIR/include -I$LIBWEBP_DIR/include"
export CGO_CXXFLAGS="$CGO_CPPFLAGS"
export CGO_LDFLAGS="-L$SPIDERMONKEY_DIR/lib -L$LIBWEBP_DIR/lib"
npm ci --prefix web && npm run build --prefix web
go build -trimpath -ldflags "-X github.com/Mag1cFall/AIStudio2API/internal/app.version=v0.3.1-uncut.5" -o aistudio2api ./cmd/aistudio2api
```

或使用 `scripts/build-linux.sh`（需先设置 `SPIDERMONKEY_DIR`、`LIBWEBP_DIR`）。

## 免责声明

- 本分支与上游作者无关，仅用于个人学习、调试与研究
- 「抗截断」本质是绕过上游内容审查，可能违反上游服务条款并提高账号风控/封禁概率；请自行评估风险，不要用于违法用途或对外提供服务
- 上游更新后本分支可能无法直接合并，请自行解决冲突
