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
	// DIUBAH: Ambil merchant_id dari query param
	merchantIDStr := c.Query("merchant_id")
	amountStr := c.Query("amount")

	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid amount",
		})
		return
	}

	// BARU: Validasi merchant_id tidak boleh kosong
	if merchantIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "merchant_id is required",
		})
		return
	}

	// BARU: Query merchant dari database berdasarkan ID
	var merchant models.Merchant
	if err := database.DB.Where("id = ? AND is_active = ?", merchantIDStr, true).First(&merchant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "merchant not found",
		})
		return
	}

	// DIUBAH: Generate QRIS dengan data merchant dari DB
	qr := GenerateQRISWithMerchant(amount, merchant.MerchantName)

	c.JSON(http.StatusOK, gin.H{
		"qris_payload": qr,
		"merchant_id":  merchant.ID,
		"amount":       amount,
	})
}

// BARU: Function untuk generate QRIS dengan merchant name dari database
func GenerateQRISWithMerchant(amount int, merchantName string) string {
	payload := ""

	// payload format
	payload += tlv("00", "01")

	// dynamic QR
	payload += tlv("01", "12")

	// merchant info (dari database sekarang)
	merchant := ""
	merchant += tlv("00", "ID.CO.QRIS.WWW")
	merchant += tlv("01", "1234567890")
	payload += tlv("26", merchant)

	// MCC
	payload += tlv("52", "5411")

	// currency IDR
	payload += tlv("53", "360")

	// amount
	payload += tlv("54", strconv.Itoa(amount))

	// country
	payload += tlv("58", "ID")

	// merchant name (DIUBAH: Dari database)
	payload += tlv("59", merchantName)

	// city (bisa hardcoded atau dari database)
	payload += tlv("60", "INDONESIA")

	// CRC placeholder
	payload += "6304"

	crc := crc16(payload)
	payload += crc

	return payload
}