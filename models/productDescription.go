package models

import (
	"gorm.io/gorm"
)

type ProductDescription struct {
	gorm.Model
	ProductID     uint          `gorm:"index"`/* `gorm:"not null;index"` */
	Heading       string        `gorm:"size:255" json:"descriptionheading"`
	Description   string        `gorm:"size:1000" json:"description"`
	IsDeleted     bool          `gorm:"default:false"`
	ProductDetail ProductDetail `gorm:"foreignKey:ProductID"`
}
