package models

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	CouponCode         string    `gorm:"unique" json:"code"`
	FixedDiscount      int       `json:"fixdiscound"`
	DiscountPercentage float64   `json:"discount_percentage"`
	MaxDiscount        int       `json:"max_value"`
	MinProductPrice    float64   `gorm:"type:numeric(10,2)" json:"min_productvalue"`
	UsersUsedCount     int       `json:"used_count"`
	MaxUseCount        int       `json:"max_usecount"`
	ValidFrom          time.Time `json:"validfrom"`
	ExpirationDate     time.Time `json:"validto"`
	IsFixedCoupon      bool      `gorm:"default:false"`
	IsActive           bool      `gorm:"default:true"`
	CouponType         string    `gorm:"default:'percentage'"`
	Status             string
}
