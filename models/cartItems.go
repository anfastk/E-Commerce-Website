package models

import (
	"gorm.io/gorm"
)

type CartItem struct {
	gorm.Model
	CartID           uint                  `gorm:"not null"`
	ProductID        uint                  `gorm:"not null"`
	ProductVariantID uint                  `gorm:"not null"`
	Quantity         int                   `gorm:"not null"`
	Cart             Cart                  `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
	ProductDetail    ProductDetail         `gorm:"foreignKey:ProductID"`
	ProductVariant   ProductVariantDetails `gorm:"foreignKey:ProductVariantID"`
}
