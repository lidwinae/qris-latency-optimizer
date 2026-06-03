package handler

import (
	"net/http"
	"qris-latency-optimizer/delivery/middleware"
	"qris-latency-optimizer/domain/entity"
	"qris-latency-optimizer/usecase"
	"time"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	usecase usecase.TransactionUsecase
}

func NewTransactionHandler(u usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{usecase: u}
}

func (h *TransactionHandler) ScanQR(c *gin.Context) {
	var req entity.ScanQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.usecase.ScanQR(req)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "merchant not found" {
			status = http.StatusNotFound
		} else if err.Error() == "failed to create transaction" {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data":    resp,
		"message": "transaction created successfully",
	})
	middleware.RecordTransactionCreated()
}

func (h *TransactionHandler) ConfirmPaymentAsync(c *gin.Context) {
	start := time.Now()
	transactionID := c.Param("id")
	err := h.usecase.ConfirmPaymentAsync(transactionID)
	if err != nil {
		middleware.RecordPaymentConfirmation("async", "error", time.Since(start).Seconds())
		status := http.StatusBadRequest
		if err.Error() != "invalid transaction id" {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"transaction_id": transactionID,
			"status":         "PROCESSING",
		},
		"message": "payment accepted and is being processed in background",
	})
	middleware.RecordPaymentConfirmation("async", "queued", time.Since(start).Seconds())
}

func (h *TransactionHandler) ConfirmPaymentSync(c *gin.Context) {
	start := time.Now()
	transactionID := c.Param("id")
	resp, err := h.usecase.ConfirmPaymentSync(transactionID)
	if err != nil {
		middleware.RecordPaymentConfirmation("sync", "error", time.Since(start).Seconds())
		status := http.StatusBadRequest
		if err.Error() == "transaction not found" {
			status = http.StatusNotFound
		} else if err.Error() == "failed to confirm payment" {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    resp,
		"message": "payment confirmed successfully (sync)",
	})
	middleware.RecordPaymentConfirmation("sync", "success", time.Since(start).Seconds())
}

func (h *TransactionHandler) GetTransactionStatus(c *gin.Context) {
	transactionID := c.Param("id")
	resp, err := h.usecase.GetTransactionStatus(transactionID)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "transaction not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	msg := "transaction data (from database)"
	if resp.CachedFrom {
		msg = "transaction data (from cache)"
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    resp,
		"message": msg,
	})
}
