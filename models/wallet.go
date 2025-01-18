package models

import "gorm.io/gorm"

type Wallet struct {
	gorm.Model
	UserID          uint
	Balance         float64
	PaymentStatus   bool `gorm:"default:false"`
	AddedAmount     float64 `gorm:"type:numeric(10,2)"`
	TransactionType string `gorm:"size:50"`
	TransactionID   string `gorm:"size:100"`
	UserAuth        UserAuth
}
