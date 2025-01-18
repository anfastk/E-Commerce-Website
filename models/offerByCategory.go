package models

import (
	"time"

	"gorm.io/gorm"
)

type OfferByCategory struct {
	gorm.Model
	CategoryOfferName       string  `gorm:"size:255"`
	CategoryOfferPercentage float64 `json:"off_percentage"`
	OfferDescription        string  `gorm:"size:255" json:"offer_description"`
	CategoryID              uint
	OfferStatus             bool      `gorm:"default:false"`
	StartDate               time.Time `json:"validfrom"`
	EndDate                 time.Time `json:"validto"`
	IsOfferDeleted          bool      `gorm:"default:false"`
	Category                Categories
}
