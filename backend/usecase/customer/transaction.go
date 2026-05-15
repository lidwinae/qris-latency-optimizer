package customer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"
	"qris-latency-optimizer/repository/redis"
	"qris-latency-optimizer/usecase/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ScanQR - endpoint untuk scan QR dari customer
func ScanQR(c *gin.Context) {
	var req models.ScanQRRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: " + err.Error(),
		})
		return
	}

	var merchant models.Merchant
	merchantID, err := uuid.Parse(req.MerchantID)
	if err == nil {
		if err := database.DB.Where("id = ? AND is_active = ?", merchantID, true).First(&merchant).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "merchant not found",
			})
			return
		}
	} else {
		if cachedMerchant, ok := redis.GetMerchant(req.MerchantID); ok {
			merchant = *cachedMerchant
		} else if err := database.DB.Where("qr_id = ? AND is_active = ?", req.MerchantID, true).First(&merchant).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "merchant not found",
			})
			return
		} else {
			redis.CacheMerchant(merchant)
		}
	}
	merchantID = merchant.ID

	qrMerchantID, qrAmount, err := service.ParsePayload(req.QRPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if qrMerchantID != merchant.QRID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "qr payload merchant does not match merchant id",
		})
		return
	}
	if float64(qrAmount) != req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "qr payload amount does not match amount",
		})
		return
	}

	transactionID := uuid.New()
	cacheKey := fmt.Sprintf("transaction:%s", transactionID.String())

	// Buat transaction model
	transaction := models.Transaction{
		ID:         transactionID,
		MerchantID: merchantID,
		Amount:     req.Amount,
		Status:     "PENDING",
		CreatedAt:  time.Now(),
	}

	// Simpan ke database
	if err := database.DB.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create transaction: " + err.Error(),
		})
		return
	}

	// Simpan ke Redis dengan TTL transaksi
	if transactionJSON, err := json.Marshal(transaction); err == nil {
		_ = redis.Set(cacheKey, string(transactionJSON), redis.TTLTransaction)
	}

	response := models.TransactionResponse{
		TransactionID: transactionID.String(),
		MerchantID:    merchantID.String(),
		Amount:        req.Amount,
		Status:        "PENDING",
		CreatedAt:     transaction.CreatedAt,
		CachedFrom:    false,
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    response,
		"message": "transaction created successfully",
	})
}

// ConfirmPayment - endpoint untuk confirm pembayaran
func ConfirmPayment(c *gin.Context) {
	transactionID := c.Param("id")

	// Validasi UUID
	if _, err := uuid.Parse(transactionID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid transaction id",
		})
		return
	}

	cacheKey := fmt.Sprintf("transaction:%s", transactionID)

	// Update di database
	result := database.DB.Model(&models.Transaction{}).
		Where("id = ?", transactionID).
		Update("status", "SUCCESS")
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to confirm payment: " + result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "transaction not found",
		})
		return
	}

	// Hapus dari cache (invalidate)
	redis.Delete(cacheKey)

	// Ambil data transaksi yang sudah updated
	var transaction models.Transaction
	if err := database.DB.First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "transaction not found",
		})
		return
	}

	response := models.TransactionResponse{
		TransactionID: transactionID,
		MerchantID:    transaction.MerchantID.String(),
		Amount:        transaction.Amount,
		Status:        transaction.Status,
		CreatedAt:     transaction.CreatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    response,
		"message": "payment confirmed successfully",
	})
}
