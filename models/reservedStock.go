package models

import (
	"time"

	"gorm.io/gorm"
)

type ReservedStock struct {
	gorm.Model
	UserID           uint                  `gorm:"not null"`
	ProductVariantID uint                  `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Quantity         int                   `gorm:"not null;check:quantity > 0"`
	ReservedAt       time.Time             `gorm:"default:now()"`
	ReserveTill      time.Time             `gorm:"default:CURRENT_TIMESTAMP + INTERVAL '15 minutes'"`
	IsConfirmed      bool                  `gorm:"default:false"`
	ProductVariant   ProductVariantDetails `gorm:"foreignKey:ProductVariantID"`
}
