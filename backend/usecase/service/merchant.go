package service

import (
	"net/http"
	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"

	"github.com/gin-gonic/gin"
)

// GetMerchants - endpoint untuk fetch semua merchant dari DB
func GetMerchants(c *gin.Context) {
	var merchants []models.Merchant

	// Query semua merchant yang aktif dari database
	if err := database.DB.Where("is_active = ?", true).Find(&merchants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch merchants",
		})
		return
	}

	// Return response dengan format yang sesuai frontend
	c.JSON(http.StatusOK, gin.H{
		"merchants": merchants,
	})
}