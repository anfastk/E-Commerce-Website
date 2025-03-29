package models

import (
	"time"

	"gorm.io/gorm"
)

type ReservedStock struct {
	gorm.Model
	UserID           uint                  `gorm:"not null;index"`
	ProductVariantID uint                  `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Quantity         int                   `gorm:"not null;check:quantity > 0"`
	ReservedAt       time.Time             `gorm:"index;default:now()"`
	ReserveTill      time.Time             `gorm:"index;default:CURRENT_TIMESTAMP + INTERVAL '15 minutes'"`
	IsConfirmed      bool                  `gorm:"index;default:false"`
	ReservedCouponID uint                  `gorm:"index"`
	ProductVariant   ProductVariantDetails `gorm:"foreignKey:ProductVariantID"`
}
 