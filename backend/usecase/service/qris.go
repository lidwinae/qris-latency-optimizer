package service

import (
	"net/http"
	"strconv"

	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"

	"github.com/gin-gonic/gin"
)

// GenerateDynamicLegacy - generate QRIS untuk legacy (tanpa caching)
func GenerateDynamicLegacy(c *gin.Context) {
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

	// LEGACY: Query merchant dengan INT ID
	var merchant models.Merchant
	merchantID := 0
	if _, err := strconv.Atoi(merchantIDStr); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid merchant_id format",
		})
		return
	}
	merchantID, _ = strconv.Atoi(merchantIDStr)

	if err := database.DB.Where("id = ?", merchantID).First(&merchant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "merchant not found",
		})
		return
	}

	// Generate QRIS dengan merchant ID (bukan UUID)
	qr := GenerateQRISLegacy(amount, merchant.MerchantName, strconv.Itoa(merchant.ID))

	c.JSON(http.StatusOK, gin.H{
		"qris_payload": qr,
		"merchant_id":  merchant.ID,
		"amount":       amount,
	})
}

// GenerateQRISLegacy - generate QRIS dengan merchant ID integer
func GenerateQRISLegacy(amount int, merchantName string, merchantID string) string {
	payload := ""

	// payload format
	payload += tlv("00", "01")

	// dynamic QR
	payload += tlv("01", "12")

	// merchant info
	merchant := ""
	merchant += tlv("00", "ID.CO.QRIS.WWW")
	merchant += tlv("01", merchantID) // Bisa pakai integer ID langsung
	payload += tlv("26", merchant)

	// MCC
	payload += tlv("52", "5411")

	// currency IDR
	payload += tlv("53", "360")

	// amount
	payload += tlv("54", strconv.Itoa(amount))

	// country
	payload += tlv("58", "ID")

	// merchant name
	payload += tlv("59", merchantName)

	// city
	payload += tlv("60", "INDONESIA")

	// CRC placeholder
	payload += "6304"

	crc := crc16(payload)
	payload += crc

	return payload
}