package models

import "gorm.io/gorm"

type WalletTransaction struct {
	gorm.Model
	UserID        uint     `gorm:"not null"`
	WalletID      uint     `gorm:"not null"`
	Amount        float64  `gorm:"not null"`
	Description   string   `gorm:"size:150"`
	Type          string   `gorm:"size:50"`
	Receipt       string   `gorm:"size:255"`
	OrderId       string   `gorm:"size:100"`
	TransactionID string   `gorm:"size:100"`
	PaymentMethod string   `gorm:"size:50"`
	UserAuth      UserAuth `gorm:"foreignKey:UserID;references:ID"`
}
