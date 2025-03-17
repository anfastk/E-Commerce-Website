package models

import (
	"time"

	"gorm.io/gorm"
)

type Coupon struct {
	gorm.Model
	CouponCode       string    `gorm:"unique;index" json:"code"`
	Discription      string    `gorm:"not null"`
	DiscountValue    float64   `gorm:"not null;index" json:"discount_value"`
	MaxDiscountValue float64   `gorm:"not null;index" json:"max_value"`
	MinOrderValue    float64   `gorm:"type:numeric(10,2)" json:"min_productvalue"`
	UsersUsedCount   int       `gorm:"not null;index" json:"used_count"`
	MaxUseCount      int       `gorm:"not null;index" json:"max_use_count"`
	ApplicableFor    string    `gorm:"not null;index" json:"applicable"`
	ValidFrom        time.Time `gorm:"not null;index" json:"validfrom"`
	ExpirationDate   time.Time `gorm:"not null;index" json:"validto"`
	IsFixedCoupon    bool      `gorm:"not null;index" gorm:"default:false"`
	CouponType       string    `gorm:"not null;index" gorm:"default:'Percentage'"`
	Status           string    `gorm:"default:'Active'"`
}
