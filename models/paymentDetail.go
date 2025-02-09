package models

import (
	"gorm.io/gorm"
)

type PaymentDetail struct {
	gorm.Model
	UserID        uint     `gorm:"not null"`
	OrderID       uint     `gorm:"not null"`
	PaymentID     string   `json:"paymentid"`
	Receipt       string   `gorm:"size:255"`
	PaymentStatus string   `gorm:"size:50"`
	PaymentAmount float64  `gorm:"type:numeric(10,2)"`
	TransactionID string   `gorm:"size:100"`
	PaymentMethod string   `gorm:"size:50"`
	OrderDetail   Order    `gorm:"foreignKey:OrderID"`
	UserAuth      UserAuth `gorm:"foreignKey:UserID"`
}
