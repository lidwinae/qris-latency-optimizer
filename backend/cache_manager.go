package cache

import (
	"context"
	"time"

	"qris-latency-optimizer/config"
)

// TTL untuk setiap jenis data
const (
	TTLMerchant = 10 * time.Minute // data merchant jarang berubah
	TTLInquiry  = 2 * time.Minute  // hasil inquiry QRIS
)

// Get ambil data dari Redis
// Return: (value, true) = HIT | ("", false) = MISS atau Redis tidak aktif
func Get(ctx context.Context, key string) (string, bool) {
	if !config.RedisAvailable {
		return "", false
	}
	val, err := config.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// Set simpan data ke Redis dengan TTL
func Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if !config.RedisAvailable {
		return nil // silent skip kalau Redis mati
	}
	return config.RedisClient.Set(ctx, key, value, ttl).Err()
}

// Delete hapus cache (dipanggil saat data berubah)
func Delete(ctx context.Context, key string) error {
	if !config.RedisAvailable {
		return nil
	}
	return config.RedisClient.Del(ctx, key).Err()
}

// Exists cek apakah key ada
func Exists(ctx context.Context, key string) bool {
	if !config.RedisAvailable {
		return false
	}
	count, err := config.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false
	}
	return count > 0
}
