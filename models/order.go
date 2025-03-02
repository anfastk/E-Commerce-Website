package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderUID              string      `gorm:"not null"`
	UserID               uint        `gorm:"not null"`
	CouponCode           string      `gorm:"size:255"`
	CouponID             uint        `gorm:"default:NULL"`
	SubTotal             float64     `gorm:"type:numeric(10,2)"`
	TotalProductDiscount float64     `gorm:"type:numeric(10,2)"`
	TotalDiscount        float64     `gorm:"type:numeric(10,2)"`
	TotalAmount          float64     `gorm:"type:numeric(10,2)"`
	ShippingCharge       float64     `gorm:"type:numeric(10,2)"`
	Tax                  float64     `gorm:"not null"`
	OrderDate            time.Time   `gorm:"not null"`
	UserAuth             UserAuth    `gorm:"foreignKey:UserID;references:ID"`
	CouponDiscountAmount float64     `gorm:"type:numeric(10,2)"`
	IsCouponApplied      bool        `gorm:"default:false"`
	CouponDetail         Coupon      `gorm:"foreignKey:CouponID;references:ID"`
	OrderItem            []OrderItem `gorm:"foreignKey:OrderID;references:ID"`
}
