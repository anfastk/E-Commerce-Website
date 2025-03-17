package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductOffer struct {
	gorm.Model
	OfferName       string        `gorm:"size:255" json:"offername"`
	OfferDetails    string        `gorm:"size:255" json:"offer"`
	OfferPercentage float64       `json:"discountpercentage"`
	StartDate       time.Time     `gorm:"not null"`
	EndDate         time.Time     `gorm:"not null"`
	ProductID       uint          `gorm:"unique;index;not null" json:"productid"`
	Status          string        `gorm:"not null"`
	Product         ProductDetail `gorm:"foreignKey:ProductID"`
}
