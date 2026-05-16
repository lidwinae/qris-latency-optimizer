package handler

import (
	"qris-latency-optimizer/usecase/customer"
	"qris-latency-optimizer/usecase/service"

	"github.com/gin-gonic/gin"
)

func Rest(r *gin.Engine) {
	// Create QR code endpoint
	r.GET("/api/qris", service.GenerateDynamic)

	// Backend endpoints
	r.GET("/api/merchants", service.GetMerchants)
	r.GET("/api/transactions/:id", service.GetTransactionStatus)

	// Customer endpoints
	r.POST("/api/transactions/scan", customer.ScanQR)

	// ENDPOINT 1: OPTIMIZED (RabbitMQ / Cepat) -> Kotak Hijau Grafana
	r.POST("/api/transactions/:id/confirm", customer.ConfirmPayment)

	// ENDPOINT 2: NON-OPTIMIZED (Sync DB / Lambat) -> Kotak Merah Grafana
	r.POST("/api/transactions/:id/confirm-sync", customer.ConfirmPaymentSync)

	// check health endpoint
	r.GET("/api/ping", service.Ping)

	// System monitoring endpoint (CPU, memory, services)
	r.GET("/api/monitor/system", service.GetSystemMonitor)

	// K6 load test monitoring endpoints
	r.POST("/api/monitor/k6/data", service.PostK6Data)
	r.POST("/api/monitor/k6/summary", service.PostK6Summary)
	r.GET("/api/monitor/k6", service.GetK6Dashboard)
	r.DELETE("/api/monitor/k6", service.ClearK6Data)

	// Live latency tracking endpoint
	r.GET("/api/monitor/live", service.GetLiveLatency)

	// Serve real-time monitoring dashboards
	r.StaticFile("/monitor", "./monitoring/index.html")
	r.StaticFile("/latency", "./monitoring/latency.html")
}
