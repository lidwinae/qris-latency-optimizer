package service

import (
	"fmt"
	"net/http"
	"time"

	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"

	"github.com/gin-gonic/gin"
)

// GetTransactionStatusLegacy - get transaction tanpa redis (pure DB)
func GetTransactionStatusLegacy(c *gin.Context) {
	transactionID := c.Param("id")

	// Measure query time untuk benchmark
	startTime := time.Now()

	var transaction models.Transaction
	if err := database.DB.First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "transaction not found",
		})
		return
	}

	queryTime := time.Since(startTime).Milliseconds()
	fmt.Printf("✓ Database Query Time: %dms (LEGACY - NO CACHE)\n", queryTime)

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
		"message": "transaction data (from database - LEGACY)",
	})
}