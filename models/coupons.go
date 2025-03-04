package models

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	CouponCode       string    `gorm:"unique" json:"code"`
	Discription      string    `gorm:"not null"`
	DiscountValue    float64   `json:"discount_value"`
	MaxDiscountValue float64   `json:"max_value"`
	MinOrdervalue    float64   `gorm:"type:numeric(10,2)" json:"min_productvalue"`
	UsersUsedCount   int       `json:"used_count"`
	MaxUseCount      int       `json:"max_use_count"`
	ApplicableFor    string    `json:"applicable"`
	ValidFrom        time.Time `json:"validfrom"`
	ExpirationDate   time.Time `json:"validto"`
	IsFixedCoupon    bool      `gorm:"default:false"`
	CouponType       string    `gorm:"default:'Percentage'"`
	Status           string    `gorm:"default:'Active'"`
}
