package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model
	OrderID               uint                  `gorm:"not null"`
	UserID                uint                  `gorm:"not null"`
	ProductVariantID      uint                  `gorm:"not null"`
	Quantity              int                   `gorm:"not null"`
	Subtotal              float64               `gorm:"type:numeric(10,2)" json:"subtotal"`
	OrderStatus           string                `gorm:"default:'Pending'" json:"status"`
	IsDelivered           bool                  `gorm:"default:false"`
	DeliveryDate          time.Time             `gorm:"not null"`
	ReturnableStatus      bool                  `gorm:"default:true"`
	ReturnDate            time.Time             
	UserAuth              UserAuth              `gorm:"foreignKey:UserID;references:ID"`
	ProductVariantDetails ProductVariantDetails `gorm:"foreignKey:ProductVariantID;references:ID"`
}
