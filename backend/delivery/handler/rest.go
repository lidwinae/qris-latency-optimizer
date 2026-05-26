package handler

import (
	"qris-latency-optimizer/usecase/customer"
	"qris-latency-optimizer/usecase/service"
	"qris-latency-optimizer/internal/websocket" // TAMBAHKAN

	"github.com/gin-gonic/gin"
)

// DIUBAH: add wsHub parameter
func Rest(r *gin.Engine, wsHub *websocket.Hub) {
	// Create QR code endpoint
	r.GET("/api/qris", service.GenerateDynamic)

	// Backend endpoints
	r.GET("/api/merchants", service.GetMerchants)
	r.GET("/api/transactions/:id", service.GetTransactionStatus)

	// Customer endpoints
	r.POST("/api/transactions/scan", customer.ScanQR)
	r.POST("/api/transactions/:id/confirm", customer.ConfirmPayment)

	// check health endpoint
	r.GET("/api/ping", service.Ping)

	// WebSocket endpoint
	r.GET("/ws", func(c *gin.Context) {
		wsHub.HandleWebSocket(c)
	})
}