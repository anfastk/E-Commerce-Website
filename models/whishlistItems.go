package models

import (
	"gorm.io/gorm"
)

type WishlistItem struct {
	gorm.Model
	WishlistID            uint
	ProductVariantID      uint
	ProductID             uint
	ProductVariantDetails ProductVariantDetails `gorm:"foreignkey:ProductVariantID"`
	ProductDetail         ProductDetail         `gorm:"foreignkey:ProductID"`
}
