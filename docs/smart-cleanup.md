# 智能清理（内存治理）

## 背景

服务长期运行（Camoufox/Firefox 常驻页面）时，浏览器进程内存会缓慢增长；Go 进程本身通常只有几十 MB。

## 策略

1. Go 堆超过 `CLEANUP_GO_HEAP_MB`：`runtime.GC + debug.FreeOSMemory`（毫秒级）
2. 服务 cgroup 内存 ≥ `CLEANUP_MEMORY_HIGH_MB`：关闭「空闲且无活动请求」的常驻 Worker，释放整个浏览器内存
3. 达到 `CLEANUP_MEMORY_CRITICAL_MB`：同样动作，日志 WARN

安全约束：

- 只有 Worker 空闲 ≥ `CLEANUP_IDLE_SECONDS`、账户池无活动请求/鉴权/独占租约时才处理
- 两次动作最小间隔 `CLEANUP_ACTION_GAP_SECONDS`
- 不处理进行中的请求；不清理模型目录等小对象

内存读数：`/sys/fs/cgroup/.../memory.current`（包含浏览器子进程），无 cgroup 时回退进程 RSS。

## 配置

```ini
CLEANUP_ENABLED=true
CLEANUP_INTERVAL_SECONDS=60
CLEANUP_MEMORY_HIGH_MB=1536
CLEANUP_MEMORY_CRITICAL_MB=2560
CLEANUP_IDLE_SECONDS=900
CLEANUP_ACTION_GAP_SECONDS=600
CLEANUP_GO_HEAP_MB=128
```

## 实测（8GB 机器）

- 低阈值验证：403MB → 40MB，动作耗时 74ms；清理后首次请求冷启动约 12.7s（Worker 启动 9.2s）
- 默认值下常驻约 0.4–1.1GB 不触发；只有超过 1.5GB 且 15 分钟无请求才回收

## 调参

- 4GB：HIGH=1024、CRITICAL=1792、IDLE=600
- 8GB：默认
- 16GB+：HIGH=3072、CRITICAL=4096
- 对冷启动敏感：增大 IDLE 或 HIGH；不需要清理：CLEANUP_ENABLED=false
