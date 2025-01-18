package models

import "gorm.io/gorm"

type ProductSpecification struct {
	gorm.Model
	ProductVariantID   uint                  `gorm:"not null"`
	SpecificationKey   string                `gorm:"not null;size:255" json:"specificationkey"`
	SpecificationValue string                `gorm:"not null;size:255" json:"specificationvalue"`
	IsDeleted          bool                  `gorm:"default:false"`
	ProductVariant     ProductVariantDetails `gorm:"foreignKey:ProductVariantID"`
}
