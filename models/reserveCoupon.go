package models

import (
	"gorm.io/gorm"
)

type ReservedCoupon struct {
	gorm.Model
	CouponCode           string    `gorm:"not null" json:"code"`
	Discription          string    `gorm:"not null" json:"description"`
	CouponDiscountAmount float64   `json:"couponDiscountAmount"`
	CouponID             uint      `gorm:"not null;index"`
}
 