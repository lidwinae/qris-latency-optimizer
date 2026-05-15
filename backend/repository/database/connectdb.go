package database

import (
	"qris-latency-optimizer/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(LoadDatabaseConfig()), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	if err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto`).Error; err != nil {
		panic(err)
	}

	var c models.Merchant
	var d models.Transaction

	if err := DB.AutoMigrate(&c, &d); err != nil {
		panic(err)
	}

	seedMerchants()
}

func seedMerchants() {
	merchants := []models.Merchant{
		{QRID: "TEST001", MerchantName: "Kantin FILKOM UB", IsActive: true},
		{QRID: "TEST002", MerchantName: "TESTING STORE", IsActive: true},
	}

	for _, merchant := range merchants {
		if err := DB.Where("qr_id = ?", merchant.QRID).FirstOrCreate(&merchant).Error; err != nil {
			panic(err)
		}
	}
}
