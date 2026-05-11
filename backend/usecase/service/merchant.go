package service

import (
	"net/http"
	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"

	"github.com/gin-gonic/gin"
)

// GetMerchantsLegacy - fetch merchant tanpa caching
func GetMerchantsLegacy(c *gin.Context) {
	var merchants []models.Merchant

	// Pure database query - no cache
	if err := database.DB.Find(&merchants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch merchants",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"merchants": merchants,
	})
}