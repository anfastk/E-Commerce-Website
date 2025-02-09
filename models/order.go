package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID          uint            `gorm:"not null"`
	AddressID       uint            `gorm:"not null"`
	CouponCode      string          `gorm:"size:255"`
	CouponID        uint            `gorm:"default:NULL"` 
	OrderAmount     float64         `gorm:"type:numeric(10,2)"`
	ShippingCharge  float64         `gorm:"type:numeric(10,2)"`
	Tax             float64         `gorm:"not null"`
	OrderDate       time.Time       `gorm:"not null"`
	OrderStatus     string          `gorm:"default:'Pending'"`
	UserAuth        UserAuth        `gorm:"foreignKey:UserID;references:ID"`
	ShippingAddress ShippingAddress `gorm:"foreignKey:AddressID;references:ID"`
	CouponDetail    Coupon          `gorm:"foreignKey:CouponID;references:ID"`
	OrderItem       OrderItem       `gorm:"foreignKey:OrderID"`
}
