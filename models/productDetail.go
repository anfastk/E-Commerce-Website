package models

import (
	"gorm.io/gorm"
)

type ProductDetail struct {
	gorm.Model
	ProductName    string                  `gorm:"size:255" json:"productname"`
	CategoryID     uint                    `gorm:"not null;index"`
	BrandName      string                  `gorm:"size:100" json:"brand"`
	IsCODAvailable bool                    `gorm:"default:true"`
	IsReturnable   bool                    `gorm:"default:true"`
	IsDeleted      bool                    `gorm:"default:false"`
	Category       Categories              `gorm:"foreignKey:CategoryID"`
	Descriptions   []ProductDescription    `gorm:"foreignKey:ProductID"`
	Variants       []ProductVariantDetails `gorm:"foreignKey:ProductID"`
	Images         []ProductImage          `gorm:"foreignKey:ProductID"`
	Offers         []ProductOffer          `gorm:"foreignKey:ProductID"`
}
