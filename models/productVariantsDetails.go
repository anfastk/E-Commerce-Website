package models

import (
	"gorm.io/gorm"
)

type ProductVariantDetails struct {
	gorm.Model
	ProductName    string                 `gorm:"size:100" json:"productname"`
	ProductID      uint                   `gorm:"not null;index"`
	Size           string                 `gorm:"index;size:100" json:"size"`
	Colour         string                 `gorm:"index;size:100" json:"color"`
	Ram            string                 `gorm:"index;size:100" json:"ram"`
	Storage        string                 `gorm:"index;size:100" json:"storage"`
	StockQuantity  int                    `gorm:"index;not null;index" json:"stock"`
	RegularPrice   float64                `gorm:"type:numeric(10,2);index" json:"regular"`
	SalePrice      float64                `gorm:"type:numeric(10,2);index" json:"saleprice"`
	SKU            string                 `gorm:"index;unique" json:"sku"`
	ProductSummary string                 `gorm:"size:255" json:"summery"`
	IsDeleted      bool                   `gorm:"default:false"`
	Product        ProductDetail          `gorm:"foreignKey:ProductID"`
	CategoryID     uint                   `gorm:"not null;index" json:"category_id"`
	Category       Categories             `gorm:"foreignKey:CategoryID;references:ID"`
	VariantsImages []ProductVariantsImage `gorm:"foreignKey:ProductVariantID"`
	Specification  []ProductSpecification `gorm:"foreignKey:ProductVariantID"`
}
