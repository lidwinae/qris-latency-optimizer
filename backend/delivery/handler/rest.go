package handler

import (
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
	
	// DIUBAH: Endpoint QRIS dengan merchant_id support
	r.GET("/api/qris", service.GenerateDynamic)
	
	// BARU: Endpoint untuk fetch daftar merchant
	r.GET("/api/merchants", service.GetMerchants)

	r.GET("/ping", service.Ping)

	// New transaction endpoints
	r.POST("/api/transactions/scan", customer.ScanQR)
	r.GET("/api/transactions/:id", service.GetTransactionStatus)
	r.POST("/api/transactions/:id/confirm", customer.ConfirmPayment)
}
