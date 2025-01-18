package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductOffer struct {
	gorm.Model
	OfferName       string  `gorm:"size:255" json:"offername"`
	OfferDetails    string  `gorm:"size:255" json:"offer"`
	OfferAmount     float64 `gorm:"type:numeric(10,2)" json:"discount"`
	OfferPercentage float64 `json:"discountpercentage"`
	StartDate       time.Time
	EndDate         time.Time
	ProductID       uint          `gorm:"unique;not null" json:"productid"`
	IsValid         bool          `gorm:"default:true"`
	IsDeleted       bool          `gorm:"default:false"`
	Product         ProductDetail `gorm:"foreignKey:ProductID"`
}
