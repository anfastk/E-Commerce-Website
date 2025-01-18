package models

import (
	"gorm.io/gorm"
)

type PaymentDetail struct {
	gorm.Model
	UserID        uint
	OrderID       uint
	PaymentID     string `gorm:"size:100" json:"paymentid"`
	Receipt       string `gorm:"size:255"`
	PaymentStatus string `gorm:"size:50"`
	PaymentAmount float64
	TransactionID string `gorm:"size:100"`
	PaymentMethod string `gorm:"size:50"`
	OrderDetail   Order
	UserAuth      UserAuth
}
