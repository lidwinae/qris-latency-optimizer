package redis

import (
	"encoding/json"
	"fmt"
	"qris-latency-optimizer/delivery/middleware"
	"qris-latency-optimizer/domain/entity"
)

func transactionCacheKey(id string) string {
	return fmt.Sprintf("transaction:%s", id)
}

type TransactionCache struct{}

func NewTransactionCache() TransactionCache {
	return TransactionCache{}
}

func (TransactionCache) GetTransaction(id string) (*entity.Transaction, bool) {
	return GetTransaction(id)
}

func (TransactionCache) CacheTransaction(tx entity.Transaction) {
	CacheTransaction(tx)
}

func (TransactionCache) DeleteTransaction(id string) {
	DeleteTransaction(id)
}

func GetTransaction(id string) (*entity.Transaction, bool) {
	if !RedisAvailable || id == "" {
		middleware.RecordCacheLookup("transaction", "error")
		return nil, false
	}
	cachedData, err := Get(transactionCacheKey(id))
	if err != nil || cachedData == "" {
		middleware.RecordCacheLookup("transaction", "miss")
		return nil, false
	}
	var tx entity.Transaction
	if err := json.Unmarshal([]byte(cachedData), &tx); err != nil {
		_ = Delete(transactionCacheKey(id))
		middleware.RecordCacheLookup("transaction", "error")
		return nil, false
	}
	middleware.RecordCacheLookup("transaction", "hit")
	return &tx, true
}

func CacheTransaction(tx entity.Transaction) {
	if !RedisAvailable || tx.ID.String() == "" {
		middleware.RecordCacheWrite("transaction", "error")
		return
	}
	data, err := json.Marshal(tx)
	if err != nil {
		middleware.RecordCacheWrite("transaction", "error")
		return
	}
	if err := Set(transactionCacheKey(tx.ID.String()), string(data), TTLTransaction); err != nil {
		middleware.RecordCacheWrite("transaction", "error")
		return
	}
	middleware.RecordCacheWrite("transaction", "success")
}

func DeleteTransaction(id string) {
	_ = Delete(transactionCacheKey(id))
}
