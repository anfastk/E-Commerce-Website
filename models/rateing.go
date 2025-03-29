package models

import (
	"gorm.io/gorm"
)

type Rating struct {
	gorm.Model
	UserID        uint           `gorm:"not null;index" json:"rateing_user"`
	ProductID     uint           `gorm:"not null;index" json:"rateing_product"`
	Value         float64       `json:"rateing"`
	UserAuth      UserAuth      `gorm:"foreignKey:UserID;references:ID"`
	ProductDetail ProductDetail `gorm:"foreignKey:ProductID;references:ID"`
}
 