package models

import (
	"time"
)

type Merchant struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	QRID         string    `gorm:"column:qr_id;type:text;not null" json:"qr_id"`
	MerchantName string    `gorm:"column:merchant_name;type:text;not null" json:"merchant_name"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Transactions []Transaction `gorm:"foreignKey:MerchantID" json:"transactions,omitempty"`
}
