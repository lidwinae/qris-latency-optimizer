package handler

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// ==========================================
// Stress Simulation Middleware
// Simulates real-world latency under load
// ==========================================

var (
	// activeRequests tracks concurrent in-flight requests
	activeRequests int64

	// stressEnabled controls whether stress simulation is active
	stressEnabled int32 = 0
)

// EnableStress turns on the stress simulation
func EnableStress() {
	atomic.StoreInt32(&stressEnabled, 1)
	fmt.Println("⚡ STRESS MODE: ENABLED — Requests will experience simulated latency")
}

// DisableStress turns off the stress simulation
func DisableStress() {
	atomic.StoreInt32(&stressEnabled, 0)
	fmt.Println("✅ STRESS MODE: DISABLED — Normal operation resumed")
}

// StressMiddleware simulates realistic database contention and network latency
// under high concurrency. The more concurrent requests, the slower each one gets.
func StressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if atomic.LoadInt32(&stressEnabled) == 0 {
			c.Next()
			return
		}

		// Track concurrent requests
		current := atomic.AddInt64(&activeRequests, 1)
		defer atomic.AddInt64(&activeRequests, -1)

		// ──────────────────────────────────────────
		// Simulate latency that scales with load
		// More concurrent requests = more contention = more delay
		// ──────────────────────────────────────────

		var baseDelay time.Duration

		switch {
		case current > 100:
			// Extreme load: 2-8 seconds (DB connection pool exhausted)
			baseDelay = time.Duration(2000+rand.Intn(6000)) * time.Millisecond
		case current > 50:
			// Heavy load: 1-4 seconds (significant contention)
			baseDelay = time.Duration(1000+rand.Intn(3000)) * time.Millisecond
		case current > 20:
			// Moderate load: 500ms-2s (noticeable lag)
			baseDelay = time.Duration(500+rand.Intn(1500)) * time.Millisecond
		case current > 10:
			// Light load: 200ms-800ms (slight delay)
			baseDelay = time.Duration(200+rand.Intn(600)) * time.Millisecond
		default:
			// Normal: 50-200ms (baseline network latency)
			baseDelay = time.Duration(50+rand.Intn(150)) * time.Millisecond
		}

		// Add random jitter (±30%) for realism
		jitter := float64(baseDelay) * (0.7 + rand.Float64()*0.6)
		delay := time.Duration(jitter)

		// Simulate occasional timeouts under extreme load (5% chance when >80 concurrent)
		if current > 80 && rand.Intn(100) < 5 {
			delay = 10 * time.Second
		}

		if current%10 == 0 {
			fmt.Printf("🔥 Stress: %d concurrent | delay: %s\n", current, delay)
		}

		time.Sleep(delay)

		c.Next()
	}
}
