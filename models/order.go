package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	UserID         uint
	CartID         uint
	AddressID      uint
	CouponCode     string `gorm:"size:255"`
	CouponID       uint
	OrderAmount    float64
	ShippingCharge float64
	Tax            float64
	OrderDate      time.Time
	OrderStatus    string `gorm:"default:'Pending'"`
	CartDetail     Cart
	UserAuth       UserAuth
	UserAddress    UserAddress
	CouponDetail   Coupon
}
