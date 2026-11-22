package app

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// 智能清理：在“响应速度”与“内存占用”之间取平衡。
// - Go 堆超过阈值时只做 runtime.GC + FreeOSMemory（开销毫秒级，不影响请求）
// - 服务内存超过高水位时，只对“空闲一段时间且没有租约”的常驻 Worker 重新载入页面（保留浏览器进程，比冷启动快）
// - 超过紧急水位时才关闭空闲 Worker（下次请求需要冷启动）
// - 任意两次 Worker 动作之间有最小间隔，避免抖动
const (
	cleanupDefaultInterval  = 60 * time.Second
	cleanupDefaultHigh      = uint64(1536) << 20
	cleanupDefaultCritical  = uint64(2560) << 20
	cleanupDefaultIdle      = 15 * time.Minute
	cleanupDefaultActionGap = 10 * time.Minute
	cleanupDefaultGoHeap    = uint64(128) << 20
)

type cleanupSettings struct {
	enabled   bool
	interval  time.Duration
	idle      time.Duration
	actionGap time.Duration
	high      uint64
	critical  uint64
	goHeap    uint64
}

func (settings cleanupSettings) withDefaults() cleanupSettings {
	if settings.interval <= 0 {
		settings.interval = cleanupDefaultInterval
	}
	if settings.idle <= 0 {
		settings.idle = cleanupDefaultIdle
	}
	if settings.actionGap <= 0 {
		settings.actionGap = cleanupDefaultActionGap
	}
	if settings.high == 0 {
		settings.high = cleanupDefaultHigh
	}
	if settings.critical == 0 {
		settings.critical = cleanupDefaultCritical
	}
	if settings.critical < settings.high {
		settings.critical = settings.high
	}
	if settings.goHeap == 0 {
		settings.goHeap = cleanupDefaultGoHeap
	}
	return settings
}

// startMemoryCleanup 随生成服务生命周期运行，定期检查内存并按水位做最小代价的清理
func (service *trackedService) startMemoryCleanup(ctx context.Context) {
	config := service.cleanup.withDefaults()
	if !config.enabled {
		service.requests.log("service", "INFO", "智能清理已关闭")
		return
	}
	service.requests.log("service", "INFO", fmt.Sprintf(
		"智能清理已启用 | 间隔=%s | 高水位=%dMB | 紧急水位=%dMB | 空闲=%s | 动作间隔=%s | Go堆阈值=%dMB",
		config.interval, config.high>>20, config.critical>>20, config.idle, config.actionGap, config.goHeap>>20,
	))
	var lastAction time.Time
	ticker := time.NewTicker(config.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			service.runMemoryCleanup(ctx, config, &lastAction)
		}
	}
}

// runMemoryCleanup 一次清理：先回收 Go 堆，再按内存水位回收空闲 Worker
func (service *trackedService) runMemoryCleanup(ctx context.Context, config cleanupSettings, lastAction *time.Time) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	if stats.HeapAlloc > config.goHeap {
		before := stats.HeapAlloc
		runtime.GC()
		debug.FreeOSMemory()
		runtime.ReadMemStats(&stats)
		service.requests.log("service", "INFO", fmt.Sprintf(
			"智能清理：Go 堆回收 | 回收前=%dMB | 回收后=%dMB", before>>20, stats.HeapAlloc>>20))
	}
	usage, _ := serviceMemoryUsage()
	if usage == 0 {
		return
	}
	level := ""
	switch {
	case usage >= config.critical:
		level = "critical"
	case usage >= config.high:
		level = "high"
	default:
		return
	}
	if time.Since(*lastAction) < config.actionGap {
		return
	}
	before := time.Now().Add(-config.idle)
	accountID := service.workers.IdleWarmAccount(before)
	if accountID == "" {
		return
	}
	startedAt := time.Now()
	if level == "critical" {
		evicted, err := service.workers.evictIdleWorker(ctx, accountID)
		if err != nil {
			service.requests.log("service", "WARN", fmt.Sprintf(
				"智能清理：关闭空闲 Worker 失败 | 账户=%s | 错误=%v", accountID, err))
			return
		}
		if !evicted {
			return
		}
		debug.FreeOSMemory()
		*lastAction = time.Now()
		service.requests.log("service", "WARN", fmt.Sprintf(
			"智能清理：内存超过紧急水位，关闭空闲 Worker | 账户=%s | 内存=%dMB | 耗时=%s",
			accountID, usage>>20, time.Since(startedAt).Round(time.Millisecond)))
		return
	}
	evicted, err := service.workers.evictIdleWorker(ctx, accountID)
	if err != nil {
		service.requests.log("service", "WARN", fmt.Sprintf(
			"智能清理：关闭空闲 Worker 失败 | 账户=%s | 错误=%v", accountID, err))
		return
	}
	if !evicted {
		return
	}
	debug.FreeOSMemory()
	*lastAction = time.Now()
	after, _ := serviceMemoryUsage()
	service.requests.log("service", "INFO", fmt.Sprintf(
		"智能清理：内存超过高水位，关闭空闲 Worker 释放内存 | 账户=%s | 清理前=%dMB | 清理后=%dMB | 耗时=%s",
		accountID, usage>>20, after>>20, time.Since(startedAt).Round(time.Millisecond)))
}

// serviceMemoryUsage 读取服务 cgroup v2 的内存用量与上限；读不到时回退到进程 RSS
func serviceMemoryUsage() (usage uint64, limit uint64) {
	if content, err := os.ReadFile("/proc/self/cgroup"); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			parts := strings.SplitN(line, ":", 3)
			if len(parts) != 3 || parts[0] != "0" || parts[1] != "" {
				continue
			}
			base := "/sys/fs/cgroup" + parts[2]
			if value := readUintFile(base + "/memory.current"); value > 0 {
				usage = value
				limit = readUintFile(base + "/memory.max")
			}
			break
		}
	}
	if usage == 0 {
		if content, err := os.ReadFile("/proc/self/statm"); err == nil {
			fields := strings.Fields(string(content))
			if len(fields) > 1 {
				if pages, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					usage = pages * uint64(os.Getpagesize())
				}
			}
		}
	}
	return usage, limit
}

func readUintFile(path string) uint64 {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	value := strings.TrimSpace(string(content))
	if value == "" || value == "max" {
		return 0
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
