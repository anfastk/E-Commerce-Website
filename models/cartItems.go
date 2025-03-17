package models

import (
	"gorm.io/gorm"
)

type CartItem struct {
	gorm.Model
	CartID           uint                  `gorm:"not null;index"`
	ProductID        uint                  `gorm:"not null;index"`
	ProductVariantID uint                  `gorm:"not null;index"`
	Quantity         int                   `gorm:"not null"`
	Cart             Cart                  `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
	ProductDetail    ProductDetail         `gorm:"foreignKey:ProductID"`
	ProductVariant   ProductVariantDetails `gorm:"foreignKey:ProductVariantID"`
}
