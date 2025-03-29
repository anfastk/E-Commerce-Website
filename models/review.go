package models

import (
	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	UserID        uint          `gorm:"not null;index" json:"review_user"`
	ProductID     uint          `gorm:"not null;index" json:"review_product"`
	Review        string        `gorm:"size:1000" json:"review"`
	UserAuth      UserAuth      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	ProductDetail ProductDetail `gorm:"foreignKey:ProductID;references:ID"`
}
 