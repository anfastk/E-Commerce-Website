package models

import (
	"gorm.io/gorm"
)

type WishlistItem struct {
	gorm.Model
	WishlistID            uint                  `gorm:"not null;index"`
	ProductVariantID      uint                  `gorm:"not null;index"`
	ProductID             uint                  `gorm:"not null;index"`
	ProductVariantDetails ProductVariantDetails `gorm:"foreignkey:ProductVariantID"`
	ProductDetail         ProductDetail         `gorm:"foreignkey:ProductID"`
}
 