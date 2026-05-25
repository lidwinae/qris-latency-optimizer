package handler

import (
	"net/http"

	"qris-latency-optimizer/usecase/customer"
	"qris-latency-optimizer/usecase/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Rest(r *gin.Engine) {
	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"}, 
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

	// Metrics collection middleware (captures real latency including stress delays)
	r.Use(MetricsMiddleware())

	// Stress simulation middleware — affects all /api/legacy routes
	r.Use(StressMiddleware())

	// Monitoring endpoints
	r.GET("/api/metrics", GetMetrics)
	r.GET("/monitor", func(c *gin.Context) {
		c.File("../monitoring/index.html")
	})

	// Stress toggle endpoints (not affected by stress middleware itself)
	r.POST("/stress/on", func(c *gin.Context) {
		EnableStress()
		c.JSON(http.StatusOK, gin.H{"stress_mode": "ENABLED", "message": "All API requests will now experience simulated latency"})
	})
	r.POST("/stress/off", func(c *gin.Context) {
		DisableStress()
		c.JSON(http.StatusOK, gin.H{"stress_mode": "DISABLED", "message": "Normal operation resumed"})
	})
	
	r.GET("/api/legacy/qris", service.GenerateDynamicLegacy)
	r.GET("/api/legacy/merchants", service.GetMerchantsLegacy)
	r.POST("/api/legacy/transactions/scan", customer.ScanQRLegacy)
	r.GET("/api/legacy/transactions/:id", service.GetTransactionStatusLegacy)
	r.POST("/api/legacy/transactions/:id/confirm", customer.ConfirmPaymentLegacy)
}
