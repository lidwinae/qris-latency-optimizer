package customer

import (
	"fmt"
	"net/http"
	"time"

	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"

	"github.com/gin-gonic/gin"
)

// ScanQRLegacy - create transaction tanpa redis
func ScanQRLegacy(c *gin.Context) {
	var req models.ScanQRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: " + err.Error(),
		})
		return
	}

	// Buat transaction model (auto-increment ID)
	transaction := models.Transaction{
		MerchantID: req.MerchantID,
		Amount:     req.Amount,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}

	// Simpan ke database saja (no redis)
	startTime := time.Now()
	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create transaction: " + err.Error(),
		})
		return
	}
	queryTime := time.Since(startTime).Milliseconds()

	fmt.Printf("✓ Transaction Created: ID=%d | Time: %dms (LEGACY - NO CACHE)\n", transaction.ID, queryTime)

	response := models.TransactionResponse{
		TransactionID: transaction.ID,
		MerchantID:    req.MerchantID,
		Amount:        req.Amount,
		Status:        "PENDING",
		CreatedAt:     transaction.CreatedAt,
		QueryTime:     queryTime,
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    response,
		"message": "transaction created successfully (LEGACY)",
	})
}

// ConfirmPaymentLegacy - confirm payment tanpa redis
func ConfirmPaymentLegacy(c *gin.Context) {
	transactionID := c.Param("id")

	startTime := time.Now()

	// Update database saja (no cache invalidation)
	if err := database.DB.Model(&models.Transaction{}).
		Where("id = ?", transactionID).
		Update("status", "SUCCESS").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to confirm payment: " + err.Error(),
		})
		return
	}

	queryTime := time.Since(startTime).Milliseconds()
	fmt.Printf("✓ Payment Confirmed: ID=%s | Time: %dms (LEGACY - NO CACHE)\n", transactionID, queryTime)

	// Ambil data yang sudah updated
	var transaction models.Transaction
	database.DB.First(&transaction, "id = ?", transactionID)

	response := models.TransactionResponse{
		TransactionID: transaction.ID,
		MerchantID:    transaction.MerchantID,
		Amount:        transaction.Amount,
		Status:        transaction.Status,
		CreatedAt:     transaction.CreatedAt,
		QueryTime:     queryTime,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "payment confirmed successfully (LEGACY)",
	})
}