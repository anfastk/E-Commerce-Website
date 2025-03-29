package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderUID             string          `gorm:"index;not null"`
	UserID               uint            `gorm:"not null;index"`
	SubTotal             float64         `gorm:"index;type:numeric(10,2)"`
	TotalProductDiscount float64         `gorm:"type:numeric(10,2)"`
	TotalDiscount        float64         `gorm:"type:numeric(10,2)"`
	TotalAmount          float64         `gorm:"index;type:numeric(10,2)"`
	ShippingCharge       float64         `gorm:"type:numeric(10,2)"`
	Tax                  float64         `gorm:"index;not null"`
	OrderDate            time.Time       `gorm:"not null"`
	UserAuth             UserAuth        `gorm:"foreignKey:UserID;references:ID"`
	IsCouponApplied      bool            `gorm:"index;default:false"`
	CouponCode           string          `gorm:"size:255"`
	CouponDiscountAmount float64         `gorm:"index;type:numeric(10,2)"`
	CouponDiscription    string          `gorm:"size:255"`
	ShippingAddress      ShippingAddress `gorm:"foreignKey:OrderID;references:ID"`
	OrderItem            []OrderItem     `gorm:"foreignKey:OrderID;references:ID"`
}
 