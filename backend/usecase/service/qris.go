package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"
	"qris-latency-optimizer/repository/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Generate QRIS Dynamic
func GenerateDynamic(c *gin.Context) {
	merchantIDStr := c.Query("merchant_id")
	amountStr := c.Query("amount")

	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid amount",
		})
		return
	}

	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "merchant_id is required",
		})
		return
	}

	var merchant models.Merchant
	if err := database.DB.Where("id = ? AND is_active = ?", merchantIDStr, true).First(&merchant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "merchant not found",
		})
		return
	}
	redis.CacheMerchant(merchant)
	go redis.PrefetchRelatedMerchants(merchant.QRID)

	qr, err := GeneratePayload(amount, merchant.MerchantName, merchant.QRID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"qris_payload": qr,
		"merchant_id":  merchant.ID,
		"amount":       amount,
	})
}

// GetTransactionStatus - endpoint untuk check status transaksi
func GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("id")

	// Validasi UUID
	if _, err := uuid.Parse(transactionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid transaction id",
		})
		return
	}

	cacheKey := fmt.Sprintf("transaction:%s", transactionID)

	// Cek di Redis dulu (cache)
	cachedData, err := redis.Get(cacheKey)

	if err == nil && cachedData != "" {
		// Cache hit!
		var transaction models.Transaction
		if err := json.Unmarshal([]byte(cachedData), &transaction); err == nil {
			response := models.TransactionResponse{
				TransactionID: transactionID,
				MerchantID:    transaction.MerchantID.String(),
				Amount:        transaction.Amount,
				Status:        transaction.Status,
				CreatedAt:     transaction.CreatedAt,
				CachedFrom:    true,
			}

			c.JSON(http.StatusOK, gin.H{
				"data":    response,
				"message": "transaction data (from cache)",
			})
			return
		}

		_ = redis.Delete(cacheKey)
	}

	// Cache miss - query database
	var transaction models.Transaction
	if err := database.DB.First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "transaction not found",
		})
		return
	}

	// Simpan ke Redis untuk next request
	if transactionJSON, err := json.Marshal(transaction); err == nil {
		_ = redis.Set(cacheKey, string(transactionJSON), redis.TTLTransaction)
	}

	response := models.TransactionResponse{
		TransactionID: transactionID,
		MerchantID:    transaction.MerchantID.String(),
		Amount:        transaction.Amount,
		Status:        transaction.Status,
		CreatedAt:     transaction.CreatedAt,
		CachedFrom:    false,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "transaction data (from database)",
	})
}
