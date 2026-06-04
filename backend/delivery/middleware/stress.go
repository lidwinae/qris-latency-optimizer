package middleware

import (
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// StressConfig holds the configuration for the stress simulation middleware.
// When enabled, it injects artificial delays to simulate database contention,
// slow queries, and lock waits that occur under high-concurrency scenarios.
type StressConfig struct {
	Enabled  bool          `json:"enabled"`
	MinDelay time.Duration `json:"min_delay_ms"`
	MaxDelay time.Duration `json:"max_delay_ms"`
}

var (
	stressConfig = StressConfig{
		Enabled:  false,
		MinDelay: 0,
		MaxDelay: 0,
	}
	stressMu sync.RWMutex
)

// GetStressConfig returns the current stress configuration (thread-safe).
func GetStressConfig() StressConfig {
	stressMu.RLock()
	defer stressMu.RUnlock()
	return stressConfig
}

// SetStressConfig updates the stress configuration (thread-safe).
func SetStressConfig(cfg StressConfig) {
	stressMu.Lock()
	defer stressMu.Unlock()
	stressConfig = cfg
	if cfg.Enabled {
		log.Printf("[Stress] ENABLED — delay range: %v – %v", cfg.MinDelay, cfg.MaxDelay)
	} else {
		log.Println("[Stress] DISABLED — no artificial delays")
	}
}

// StressMiddleware injects artificial latency into every request when stress
// mode is active. The delay is randomised between MinDelay and MaxDelay to
// simulate realistic jitter caused by DB contention or lock waits.
func StressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := GetStressConfig()
		if cfg.Enabled && cfg.MaxDelay > 0 {
			spread := cfg.MaxDelay - cfg.MinDelay
			if spread <= 0 {
				spread = 1
			}
			delay := cfg.MinDelay + time.Duration(rand.Int63n(int64(spread)))
			time.Sleep(delay)
		}
		c.Next()
	}
}

// StressStatusHandler returns the current stress mode configuration.
func StressStatusHandler(c *gin.Context) {
	cfg := GetStressConfig()
	c.JSON(http.StatusOK, gin.H{
		"enabled":      cfg.Enabled,
		"min_delay_ms": cfg.MinDelay.Milliseconds(),
		"max_delay_ms": cfg.MaxDelay.Milliseconds(),
	})
}

// StressEnableRequest is the payload accepted by the enable endpoint.
type StressEnableRequest struct {
	MinDelayMs int64 `json:"min_delay_ms" binding:"required"`
	MaxDelayMs int64 `json:"max_delay_ms" binding:"required"`
}

// StressEnableHandler activates stress mode with the given delay range.
func StressEnableHandler(c *gin.Context) {
	var req StressEnableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}

	SetStressConfig(StressConfig{
		Enabled:  true,
		MinDelay: time.Duration(req.MinDelayMs) * time.Millisecond,
		MaxDelay: time.Duration(req.MaxDelayMs) * time.Millisecond,
	})

	c.JSON(http.StatusOK, gin.H{
		"status":       "stress mode enabled",
		"min_delay_ms": req.MinDelayMs,
		"max_delay_ms": req.MaxDelayMs,
	})
}

// StressDisableHandler deactivates stress mode, returning to normal operation.
func StressDisableHandler(c *gin.Context) {
	SetStressConfig(StressConfig{Enabled: false})
	c.JSON(http.StatusOK, gin.H{"status": "stress mode disabled"})
}
