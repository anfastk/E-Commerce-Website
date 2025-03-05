package models

import "gorm.io/gorm"

type Categories struct {
	gorm.Model
	Name            string            `gorm:"size:255" form:"name" json:"categoryname"`
	Description     string            `gorm:"not null"`
	Status          string            `gorm:"not null;default:Active;check:status in ('Active','Blocked')" form:"status" json:"status"`
	IsDeleted       bool              `gorm:"default:false"`
	OfferByCategory []OfferByCategory `gorm:"foreignKey:CategoryID"`
}
