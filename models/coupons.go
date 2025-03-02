package models

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	CouponCode     string    `gorm:"unique" json:"code"`
	DiscountAmount int       `json:"discound_amount"`
	MinOrderPrice  float64   `gorm:"type:numeric(10,2)" json:"min_order_price"`
	UsersUsedCount int       `json:"used_count"`
	MaxUseCount    int       `json:"max_usecount"`
	ValidFrom      time.Time `json:"validfrom"`
	ExpirationDate time.Time `json:"validto"`
	IsActive       bool      `gorm:"default:true"`
	Status         string
}
