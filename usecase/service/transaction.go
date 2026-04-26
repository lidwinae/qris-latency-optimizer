package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"
	"qris-latency-optimizer/repository/redis"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ScanQRRequest - payload dari client saat scan QR
type ScanQRRequest struct {
	QRPayload  string  `json:"qr_payload" binding:"required"`
	MerchantID string  `json:"merchant_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

// TransactionResponse - response untuk client
type TransactionResponse struct {
	TransactionID string    `json:"transaction_id"`
	MerchantID    string    `json:"merchant_id"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	CachedFrom    bool      `json:"cached_from,omitempty"`
}

// ScanQR - endpoint untuk scan QR dari customer
func ScanQR(c *gin.Context) {
	var req ScanQRRequest

	// Parse request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request: " + err.Error(),
		})
		return
	}

	// Generate transaction ID
	transactionID := uuid.New().String()
	cacheKey := fmt.Sprintf("transaction:%s", transactionID)

	// Buat transaction model
	transaction := models.Transaction{
		ID:         uuid.MustParse(transactionID),
		MerchantID: uuid.MustParse(req.MerchantID),
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

	// Simpan ke Redis dengan TTL 10 menit
	transactionJSON, _ := json.Marshal(transaction)
	redis.Set(cacheKey, string(transactionJSON), 10*time.Minute)

	response := TransactionResponse{
		TransactionID: transactionID,
		MerchantID:    req.MerchantID,
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
	
	// Debug: Print untuk lihat status cache
	if err != nil {
		fmt.Printf("Redis Get Error: %v\n", err)
	}
	fmt.Printf("Cache Key: %s | Data Exists: %v | Error: %v\n", cacheKey, cachedData != "", err)
	
	if err == nil && cachedData != "" {
		fmt.Println("✓ Cache HIT - returning cached data")
		// Cache hit!
		var transaction models.Transaction
		json.Unmarshal([]byte(cachedData), &transaction)

		response := TransactionResponse{
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

	fmt.Println("✗ Cache MISS - querying database")
	
	// Cache miss - query database
	var transaction models.Transaction
	if err := database.DB.First(&transaction, "id = ?", transactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "transaction not found",
		})
		return
	}

	// Simpan ke Redis untuk next request
	transactionJSON, _ := json.Marshal(transaction)
	saveErr := redis.Set(cacheKey, string(transactionJSON), 10*time.Minute)
	if saveErr != nil {
		fmt.Printf("Failed to save to Redis: %v\n", saveErr)
	} else {
		fmt.Println("✓ Data saved to Redis cache")
	}

	response := TransactionResponse{
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
	if err := database.DB.Model(&models.Transaction{}).
		Where("id = ?", transactionID).
		Update("status", "SUCCESS").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to confirm payment: " + err.Error(),
		})
		return
	}

	// Hapus dari cache (invalidate)
	redis.Delete(cacheKey)
	fmt.Println("✓ Cache invalidated after payment confirmation")

	// Ambil data transaksi yang sudah updated
	var transaction models.Transaction
	database.DB.First(&transaction, "id = ?", transactionID)

	response := TransactionResponse{
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