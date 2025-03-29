package models

import "gorm.io/gorm"

type ProductImage struct {
	gorm.Model
	ProductImages string        `gorm:"not null"`
	ProductID     uint          `gorm:"not null;index"`
	IsDeleted     bool          `gorm:"default:false"`
	Product       ProductDetail `gorm:"foreignKey:ProductID"`
}

 