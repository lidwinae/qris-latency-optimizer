package models

import (
	"time"
)

type Transaction struct {
	ID         int       `gorm:"primaryKey" json:"id"`
	MerchantID int       `gorm:"column:merchant_id" json:"merchant_id"` // Loose FK
	Amount     int64     `gorm:"column:amount;type:bigint;not null" json:"amount"`
	Status     string    `gorm:"column:status;type:text;default:'PENDING'" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// ScanQRRequest - payload dari client saat scan QR
type ScanQRRequest struct {
	QRPayload  string `json:"qr_payload" binding:"required"`
	MerchantID int    `json:"merchant_id" binding:"required"`
	Amount     int64  `json:"amount" binding:"required,gt=0"`
}

// TransactionResponse - response untuk client
type TransactionResponse struct {
	TransactionID int       `json:"transaction_id"`
	MerchantID    int       `json:"merchant_id"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	QueryTime     int64     `json:"query_time_ms,omitempty"` // Untuk benchmark
}