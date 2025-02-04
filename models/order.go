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
	OrderAmount    float64 `gorm:"type:numeric(10,2)"`
	ShippingCharge float64 `gorm:"type:numeric(10,2)"`
	Tax            float64
	OrderDate      time.Time
	OrderStatus    string      `gorm:"default:'Pending'"`
	CartDetail     Cart        `gorm:"foreignKey:CartID;references:ID"`
	UserAuth       UserAuth    `gorm:"foreignKey:UserID;references:ID"`
	UserAddress    UserAddress `gorm:"foreignKey:AddressID;references:ID"`
	CouponDetail   Coupon      `gorm:"foreignKey:CouponID;references:ID"`
}
