package models

import (
	"gorm.io/gorm"
)

type PaymentDetail struct {
	gorm.Model
	UserID        uint      `gorm:"not null"`
	OrderItemID   uint      `gorm:"not null"`
	PaymentID     string    `json:"paymentid"`
	Receipt       string    `gorm:"size:255"`
	PaymentStatus string    `gorm:"type:varchar(255);default:'Pending'" json:"status"`
	PaymentAmount float64   `gorm:"type:numeric(10,2)"`
	TransactionID string    `gorm:"size:100"`
	PaymentMethod string    `gorm:"size:50"`
	OrderItem     OrderItem `gorm:"foreignKey:OrderItemID"`
	UserAuth      UserAuth  `gorm:"foreignKey:UserID"`
}
