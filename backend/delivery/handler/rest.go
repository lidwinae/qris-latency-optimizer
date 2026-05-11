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
	
	r.GET("/api/legacy/qris", service.GenerateDynamicLegacy)
	r.GET("/api/legacy/merchants", service.GetMerchantsLegacy)
	r.POST("/api/legacy/transactions/scan", customer.ScanQRLegacy)
	r.GET("/api/legacy/transactions/:id", service.GetTransactionStatusLegacy)
	r.POST("/api/legacy/transactions/:id/confirm", customer.ConfirmPaymentLegacy)
}
