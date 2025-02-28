package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID           uint      `gorm:"not null"`
	CouponCode       string    `gorm:"size:255"`
	CouponID         uint      `gorm:"default:NULL"`
	OrderTotalAmount float64   `gorm:"type:numeric(10,2)"`
	ShippingCharge   float64   `gorm:"type:numeric(10,2)"`
	Tax              float64   `gorm:"not null"`
	OrderDate        time.Time `gorm:"not null"`
	UserAuth         UserAuth  `gorm:"foreignKey:UserID;references:ID"`
	CouponDiscount   float64   `gorm:"type:numeric(10,2)"`
	CouponDetail     Coupon    `gorm:"foreignKey:CouponID;references:ID"`
	OrderItem        OrderItem `gorm:"foreignKey:OrderID"`
}
