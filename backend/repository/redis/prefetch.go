package redis

import (
	"context"
	"encoding/json"
	"log"

	"qris-latency-optimizer/backend/config"
	"qris-latency-optimizer/backend/models"
)

// PrefetchMerchant ambil 1 merchant dari DB dan simpan ke Redis
// Dipanggil secara async: go PrefetchMerchant(qrID)
func PrefetchMerchant(qrID string) {
	if !config.RedisAvailable {
		return
	}

	ctx := context.Background()
	cacheKey := "merchant:" + qrID

	// Skip kalau sudah ada di cache
	if Exists(ctx, cacheKey) {
		return
	}

	log.Printf("[PREFETCH] Memuat merchant %s ke cache...", qrID)

	var merchant models.Merchant
	result := config.DB.Where("qr_id = ? AND is_active = true", qrID).First(&merchant)
	if result.Error != nil {
		log.Printf("[PREFETCH] Merchant %s tidak ditemukan: %v", qrID, result.Error)
		return
	}

	data, err := json.Marshal(merchant)
	if err != nil {
		log.Printf("[PREFETCH] Gagal marshal: %v", err)
		return
	}

	Set(ctx, cacheKey, string(data), TTLMerchant)
	log.Printf("[PREFETCH] ✅ Merchant %s berhasil di-prefetch", qrID)
}

// PrefetchRelatedMerchants prefetch merchant lain yang mungkin diakses berikutnya
// Strategi: setelah user inquiry merchant A, load merchant lain ke cache secara spekulatif
func PrefetchRelatedMerchants(currentQrID string) {
	if !config.RedisAvailable {
		return
	}

	ctx := context.Background()
	log.Printf("[PREFETCH] Memuat merchant terkait untuk %s...", currentQrID)

	var merchants []models.Merchant
	config.DB.
		Where("is_active = true AND qr_id != ?", currentQrID).
		Limit(5).
		Find(&merchants)

	count := 0
	for _, merchant := range merchants {
		cacheKey := "merchant:" + merchant.QRID
		if Exists(ctx, cacheKey) {
			continue
		}
		data, err := json.Marshal(merchant)
		if err != nil {
			continue
		}
		// TTL setengah dari normal karena ini prefetch spekulatif
		Set(ctx, cacheKey, string(data), TTLMerchant/2)
		count++
	}

	log.Printf("[PREFETCH] ✅ %d merchant terkait di-prefetch", count)
}

// WarmUpCache isi Redis dengan semua merchant aktif saat server pertama start
// Tujuan: hindari cold-start (semua request pertama harus ke DB)
func WarmUpCache() {
	if !config.RedisAvailable {
		log.Println("[WARMUP] Redis tidak aktif, skip warm-up")
		return
	}

	ctx := context.Background()
	log.Println("[WARMUP] Memulai cache warm-up...")

	var merchants []models.Merchant
	config.DB.Where("is_active = true").Find(&merchants)

	count := 0
	for _, merchant := range merchants {
		cacheKey := "merchant:" + merchant.QRID
		data, err := json.Marshal(merchant)
		if err != nil {
			continue
		}
		Set(ctx, cacheKey, string(data), TTLMerchant)
		count++
	}

	log.Printf("[WARMUP] ✅ %d merchant berhasil di-load ke cache", count)
}
