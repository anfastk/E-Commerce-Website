package models

import "gorm.io/gorm"

type Cart struct {
	gorm.Model
	UserID    uint       `gorm:"unique;not null"`
	UserAuth  UserAuth   `gorm:"foreignKey:UserID"`
	CartItems []CartItem `gorm:"foreignKey:CartID"`
}
