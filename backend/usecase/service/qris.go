package service

import (
	"net/http"
	"strconv"

	"qris-latency-optimizer/models"
	"qris-latency-optimizer/repository/database"

	"github.com/gin-gonic/gin"
)

// GenerateDynamic - generate QRIS dengan merchant_id dan amount
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

	// DIUBAH: Pass UUID merchant ke function
	qr := GenerateQRISWithMerchant(amount, merchant.MerchantName, merchant.ID.String())

	c.JSON(http.StatusOK, gin.H{
		"qris_payload": qr,
		"merchant_id":  merchant.ID,
		"amount":       amount,
	})
}

// DIUBAH: Function sekarang terima merchantID sebagai parameter
func GenerateQRISWithMerchant(amount int, merchantName string, merchantID string) string {
	payload := ""

	// payload format
	payload += tlv("00", "01")

	// dynamic QR
	payload += tlv("01", "12")

	// merchant info dengan UUID yang benar
	merchant := ""
	merchant += tlv("00", "ID.CO.QRIS.WWW")
	merchant += tlv("01", merchantID) // ← DIUBAH: Pakai UUID merchant dari DB
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

	// city (bisa dari DB juga kalau mau)
	payload += tlv("60", "INDONESIA")

	// CRC placeholder
	payload += "6304"

	crc := crc16(payload)
	payload += crc

	return payload
}