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
	r.POST("/api/transactions/:id/confirm", customer.ConfirmPayment)

	// check health endpoint
	r.GET("/api/ping", service.Ping)
}
