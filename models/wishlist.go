package models

import "gorm.io/gorm"

type Wishlist struct {
	gorm.Model
	UserID   uint     `gorm:"not null;index"`
	UserAuth UserAuth `gorm:"foreignkey:UserID"`
}
