package models

import (
	"gorm.io/gorm"
)

type Rating struct {
	gorm.Model
	UserID        uint `json:"rateing_user"`
	ProductID     uint `json:"rateing_product"`
	Value         float64 `json:"rateing"`
	UserAuth      UserAuth
	ProductDetail ProductDetail
}
