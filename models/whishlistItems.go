package models

import (
	"gorm.io/gorm"
)

type WishlistItem struct {
	gorm.Model
	WishlistID    uint
	ProductID     uint
	ProductDetail ProductDetail `gorm:"foreignkey:ProductID"`
}
