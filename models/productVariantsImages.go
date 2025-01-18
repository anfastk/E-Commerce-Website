package models

import "gorm.io/gorm"

type ProductVariantsImage struct {
	gorm.Model
	ProductVariantsImages string                `gorm:"not null"`
	ProductVariantID      uint                  `gorm:"not null"`
	IsDeleted             bool                  `gorm:"default:false"`
	ProductVariant        ProductVariantDetails `gorm:"foreignKey:ProductVariantID"`
}
