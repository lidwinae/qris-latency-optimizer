package handler

import (
	"math"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// ==========================================
// Real-time Metrics Collection System
// ==========================================

type latencyEntry struct {
	Timestamp  time.Time
	LatencyMs  float64
	StatusCode int
	ClientIP   string
}

type requestLog struct {
	Timestamp  string  `json:"timestamp"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	StatusCode int     `json:"status_code"`
	LatencyMs  float64 `json:"latency_ms"`
	ClientIP   string  `json:"client_ip"`
}

type throughputBucket struct {
	Count      int64
	ErrorCount int64
	LatencySum float64
}

type metricsCollector struct {
	mu             sync.RWMutex
	startTime      time.Time
	activeRequests int64

	endpointLatencies map[string][]latencyEntry
	throughputBuckets map[int64]*throughputBucket
	recentLog         []requestLog

	cpuMu        sync.RWMutex
	lastCPUTotal uint64
	lastCPUIdle  uint64
	cpuPercent   float64
}

var collector *metricsCollector

func init() {
	collector = &metricsCollector{
		startTime:         time.Now(),
		endpointLatencies: make(map[string][]latencyEntry),
		throughputBuckets:  make(map[int64]*throughputBucket),
		recentLog:          make([]requestLog, 0, 100),
	}
	go collector.sampleCPU()
}

func (m *metricsCollector) sampleCPU() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		if len(lines) == 0 {
			continue
		}
		fields := strings.Fields(lines[0])
		if len(fields) < 8 || fields[0] != "cpu" {
			continue
		}
		var total, idle uint64
		for i := 1; i < len(fields); i++ {
			val, _ := strconv.ParseUint(fields[i], 10, 64)
			total += val
			if i == 4 {
				idle = val
			}
		}
		m.cpuMu.Lock()
		if m.lastCPUTotal > 0 {
			totalDelta := total - m.lastCPUTotal
			idleDelta := idle - m.lastCPUIdle
			if totalDelta > 0 {
				m.cpuPercent = float64(totalDelta-idleDelta) / float64(totalDelta) * 100
			}
		}
		m.lastCPUTotal = total
		m.lastCPUIdle = idle
		m.cpuMu.Unlock()
	}
}

func (m *metricsCollector) record(method, path, clientIP string, statusCode int, latencyMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	key := method + " " + path

	m.endpointLatencies[key] = append(m.endpointLatencies[key], latencyEntry{
		Timestamp: now, LatencyMs: latencyMs, StatusCode: statusCode, ClientIP: clientIP,
	})

	sec := now.Unix()
	bucket, ok := m.throughputBuckets[sec]
	if !ok {
		bucket = &throughputBucket{}
		m.throughputBuckets[sec] = bucket
	}
	bucket.Count++
	bucket.LatencySum += latencyMs
	if statusCode >= 400 {
		bucket.ErrorCount++
	}

	entry := requestLog{
		Timestamp: now.Format(time.RFC3339), Method: method, Path: path,
		StatusCode: statusCode, LatencyMs: math.Round(latencyMs*100) / 100, ClientIP: clientIP,
	}
	if len(m.recentLog) >= 100 {
		m.recentLog = m.recentLog[1:]
	}
	m.recentLog = append(m.recentLog, entry)

	// Cleanup: keep last 2 minutes
	cutoff := now.Add(-2 * time.Minute)
	cutoffSec := cutoff.Unix()
	for k, entries := range m.endpointLatencies {
		start := 0
		for i, e := range entries {
			if e.Timestamp.After(cutoff) {
				start = i
				break
			}
		}
		if start > 0 {
			m.endpointLatencies[k] = entries[start:]
		}
	}
	for ts := range m.throughputBuckets {
		if ts < cutoffSec {
			delete(m.throughputBuckets, ts)
		}
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

// MetricsMiddleware collects per-request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/stress") || c.Request.URL.Path == "/api/metrics" || c.Request.URL.Path == "/monitor" {
			c.Next()
			return
		}

		atomic.AddInt64(&collector.activeRequests, 1)
		start := time.Now()

		c.Next()

		latencyMs := float64(time.Since(start).Microseconds()) / 1000.0
		atomic.AddInt64(&collector.activeRequests, -1)

		collector.record(c.Request.Method, c.FullPath(), c.ClientIP(), c.Writer.Status(), latencyMs)
	}
}

// GetMetrics returns JSON metrics for the dashboard
func GetMetrics(c *gin.Context) {
	collector.mu.RLock()
	defer collector.mu.RUnlock()

	now := time.Now()
	uptime := now.Sub(collector.startTime).Seconds()

	// System stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	collector.cpuMu.RLock()
	cpuPct := math.Round(collector.cpuPercent*10) / 10
	collector.cpuMu.RUnlock()

	// Aggregate endpoint stats
	type epStat struct {
		Method     string  `json:"method"`
		Path       string  `json:"path"`
		TotalCount int     `json:"total_count"`
		ErrorCount int     `json:"error_count"`
		AvgMs      float64 `json:"avg_ms"`
		MinMs      float64 `json:"min_ms"`
		MaxMs      float64 `json:"max_ms"`
		P50Ms      float64 `json:"p50_ms"`
		P90Ms      float64 `json:"p90_ms"`
		P95Ms      float64 `json:"p95_ms"`
		P99Ms      float64 `json:"p99_ms"`
	}

	var endpoints []epStat
	var totalReqs, totalErrors int

	for key, entries := range collector.endpointLatencies {
		parts := strings.SplitN(key, " ", 2)
		if len(parts) != 2 || len(entries) == 0 {
			continue
		}

		latencies := make([]float64, 0, len(entries))
		errCount := 0
		for _, e := range entries {
			latencies = append(latencies, e.LatencyMs)
			if e.StatusCode >= 400 {
				errCount++
			}
		}
		sort.Float64s(latencies)

		sum := 0.0
		for _, l := range latencies {
			sum += l
		}

		ep := epStat{
			Method: parts[0], Path: parts[1],
			TotalCount: len(entries), ErrorCount: errCount,
			AvgMs: math.Round(sum/float64(len(latencies))*100) / 100,
			MinMs: math.Round(latencies[0]*100) / 100,
			MaxMs: math.Round(latencies[len(latencies)-1]*100) / 100,
			P50Ms: math.Round(percentile(latencies, 0.50)*100) / 100,
			P90Ms: math.Round(percentile(latencies, 0.90)*100) / 100,
			P95Ms: math.Round(percentile(latencies, 0.95)*100) / 100,
			P99Ms: math.Round(percentile(latencies, 0.99)*100) / 100,
		}
		endpoints = append(endpoints, ep)
		totalReqs += len(entries)
		totalErrors += errCount
	}

	// Throughput & latency history (last 60 seconds)
	type tsPoint struct {
		Ts    string  `json:"ts"`
		Value float64 `json:"value"`
	}
	var throughputHistory, latencyHistory []tsPoint
	for i := 59; i >= 0; i-- {
		sec := now.Add(-time.Duration(i) * time.Second).Unix()
		ts := time.Unix(sec, 0).Format("15:04:05")
		bucket, ok := collector.throughputBuckets[sec]
		if ok && bucket.Count > 0 {
			throughputHistory = append(throughputHistory, tsPoint{Ts: ts, Value: float64(bucket.Count)})
			latencyHistory = append(latencyHistory, tsPoint{Ts: ts, Value: math.Round(bucket.LatencySum/float64(bucket.Count)*100) / 100})
		} else {
			throughputHistory = append(throughputHistory, tsPoint{Ts: ts, Value: 0})
			latencyHistory = append(latencyHistory, tsPoint{Ts: ts, Value: 0})
		}
	}

	// Overall avg latency
	allLatencies := make([]float64, 0)
	for _, entries := range collector.endpointLatencies {
		for _, e := range entries {
			allLatencies = append(allLatencies, e.LatencyMs)
		}
	}
	avgLatency := 0.0
	if len(allLatencies) > 0 {
		sum := 0.0
		for _, l := range allLatencies {
			sum += l
		}
		avgLatency = math.Round(sum/float64(len(allLatencies))*100) / 100
	}

	// RPS (last 5 seconds average)
	rpsSum := int64(0)
	rpsCount := 0
	for i := 1; i <= 5; i++ {
		sec := now.Add(-time.Duration(i) * time.Second).Unix()
		if b, ok := collector.throughputBuckets[sec]; ok {
			rpsSum += b.Count
			rpsCount++
		}
	}
	rps := 0.0
	if rpsCount > 0 {
		rps = math.Round(float64(rpsSum)/float64(rpsCount)*10) / 10
	}

	errorRate := 0.0
	if totalReqs > 0 {
		errorRate = math.Round(float64(totalErrors)/float64(totalReqs)*10000) / 100
	}

	// Reverse recent log for newest-first
	recentReversed := make([]requestLog, len(collector.recentLog))
	for i, r := range collector.recentLog {
		recentReversed[len(collector.recentLog)-1-i] = r
	}

	c.JSON(http.StatusOK, gin.H{
		"uptime_seconds": math.Round(uptime),
		"stress_mode":    atomic.LoadInt32(&stressEnabled) == 1,
		"system": gin.H{
			"cpu_percent":    cpuPct,
			"memory_used_mb": math.Round(float64(memStats.Alloc)/1024/1024*10) / 10,
			"memory_sys_mb":  math.Round(float64(memStats.Sys)/1024/1024*10) / 10,
			"goroutines":     runtime.NumGoroutine(),
			"gc_cycles":      memStats.NumGC,
		},
		"overview": gin.H{
			"total_requests":     totalReqs,
			"total_errors":       totalErrors,
			"active_connections": atomic.LoadInt64(&collector.activeRequests),
			"avg_latency_ms":     avgLatency,
			"requests_per_sec":   rps,
			"error_rate_pct":     errorRate,
		},
		"endpoints":          endpoints,
		"throughput_history":  throughputHistory,
		"latency_history":    latencyHistory,
		"recent_requests":    recentReversed,
	})
}
