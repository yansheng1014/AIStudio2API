# 构建与调试笔记

## 构建环境

- Go 1.27+（源码 `go.mod` 要求 1.27）
- Node.js 22.18+/24.11+ 与 npm，用于构建 `internal/webui/dist`（go:embed）
- gcc / C++ 编译器
- SpiderMonkey 静态库（Mag1cFall/firefox Releases，如 v152.0.4-aistudio2api.1）
- libwebp 静态库（Mag1cFall/libwebp Releases，v0.3.1 起必需）

```bash
export CGO_ENABLED=1
export CGO_CPPFLAGS="-I$SPIDERMONKEY_DIR/include -I$LIBWEBP_DIR/include"
export CGO_CXXFLAGS="$CGO_CPPFLAGS"
export CGO_LDFLAGS="-L$SPIDERMONKEY_DIR/lib -L$LIBWEBP_DIR/lib"
npm ci --prefix web && npm run build --prefix web
go build -trimpath -ldflags "-X github.com/Mag1cFall/AIStudio2API/internal/app.version=v0.3.15-uncut.1" -o aistudio2api ./cmd/aistudio2api
```

国内网络：

```bash
export GOPROXY="https://goproxy.io,https://mirrors.aliyun.com/goproxy/,direct"
export GOSUMDB=off
```

常见报错：

- `webp/encode.h: No such file or directory`：缺少 libwebp 头文件/路径
- `proxy.golang.org ... i/o timeout`：换 GOPROXY 镜像
- 版本显示 `(devel)`：构建时漏了 `-X .../internal/app.version=...`

## 上游通道

- 上游 Build 通道已停用生成，请求会返回 `PROXY_GENERATION_DISABLED: generateContent is not supported`
- 建议 `.env`：`UPSTREAM_CHANNELS=playground`，`BUILD_NATIVE_NONSTREAM=false`
- 只有明确需要 File Search 等 Build 特性时才启用 build，并接受该通道可能不可用

## 思考档位

- 映射：`low=1`、`medium=2`、`high=3`、`minimal/none=4`
- 实测：低档（low/minimal）上游通常不返回思考摘要；部分模型（如 gemini-4-argon）默认 medium 也不返回，只有 high 返回
- Gemini 原生协议需 `generationConfig.thinkingConfig.includeThoughts=true`
- Anthropic：`thinking.type=enabled` 且 budget 映射到 medium/high 才常见思考块（budget 1024 → low 无摘要；16384 → high）
- OpenAI Chat 默认返回 `reasoning_content`（取决于模型/档位）

## API Key

- v0.3.1 支持多个 key（逗号分隔）与每 key 绑定账户（`API_KEY_ACCOUNTS`）
- 修改密钥后需要**重启进程**才生效，管理页会提示
- key 内不能含逗号；留空时只有本机页面可调用

## 抗截断调试

- 看日志 `uncut 包装完成` 的 `tool_chars/plain_chars/fallback/finish_reason`
- `fallback=true` 说明模型没有调用工具，正文仍走普通通道
- 长历史、敏感内容会降低模型服从度；可提高思考档、压缩历史、换模型
- 小 `max_tokens` 可能被 thinking 吃光，导致空正文

## 内存清理调试

- 启动日志 `智能清理已启用 | 间隔=... | 高水位=...`
- 动作日志 `智能清理：内存超过高水位，关闭空闲 Worker 释放内存 | 清理前=... | 清理后=...`
- 想验证可通过临时调低 `CLEANUP_MEMORY_HIGH_MB` 与 `CLEANUP_IDLE_SECONDS` 观察，再恢复

## 反代注意

- 反向代理需要转发 `Host`、`X-Real-IP`、`X-Forwarded-For`、`X-Forwarded-Proto`
- WebSocket 路径需要 `Upgrade`/`Connection` 头，否则管理页 403 或握手失败
- 流式响应建议关闭代理缓冲（如 `X-Accel-Buffering: no`）
--- debug note ---
v0.3.15 构建依赖：Go 1.27.3、SpiderMonkey v158.0.1-aistudio2api.2（gcc 用 linux-amd64-gnu）、libwebp v1.6.0-aistudio2api.1。
