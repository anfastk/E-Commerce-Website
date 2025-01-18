package models

import (
	"gorm.io/gorm"
)

type CartItem struct {
	gorm.Model
	CartID           uint
	ProductID        uint
	ProductVariantID uint
	Quantity         int
	CartDetail       Cart
	ProductDetail    ProductDetail
	ProductVariant   ProductVariantDetails
}
