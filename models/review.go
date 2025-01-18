package models

import (
	"gorm.io/gorm"
)

type Review struct {
	gorm.Model
	UserID        uint `json:"review_user"`
	ProductID     uint `json:"review_product"`
	Review        string `gorm:"size:1000" json:"review"`
	UserAuth      UserAuth
	ProductDetail ProductDetail
}
