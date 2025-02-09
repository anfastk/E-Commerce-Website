package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductVariantDetails struct {
	gorm.Model
	ProductName          string                 `gorm:"size:100" json:"productname"`
	ProductID            uint                   `gorm:"not null"`
	Size                 string                 `gorm:"size:100" json:"size"`
	Colour               string                 `gorm:"size:100" json:"color"`
	Ram                  string                 `gorm:"size:100" json:"ram"`
	Storage              string                 `gorm:"size:100" json:"storage"`
	StockQuantity        int                    `json:"stock"`
	RegularPrice         float64                `gorm:"type:numeric(10,2)" json:"regular"`
	SalePrice            float64                `gorm:"type:numeric(10,2)" json:"saleprice"`
	SKU                  string                 `gorm:"unique" json:"sku"`
	ProductSummary       string                 `gorm:"size:255" json:"summery"`
	IsDeleted            bool                   `gorm:"default:false"`
	Product              ProductDetail          `gorm:"foreignKey:ProductID"`
	CategoryID           uint                   `json:"category_id"`
	Category             Categories             `gorm:"foreignKey:CategoryID;references:ID"`
	VariantsImages       []ProductVariantsImage `gorm:"foreignKey:ProductVariantID"`
	Specification        []ProductSpecification `gorm:"foreignKey:ProductVariantID"`
	ReservedStock        uint                   `gorm:"default:0"`
	ReservationExpiresAt time.Time
}
