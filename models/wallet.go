package models

import "gorm.io/gorm"

type Wallet struct {
	gorm.Model
	UserID            uint `gorm:"not null"`
	Balance           float64
	UserAuth          UserAuth          `gorm:"foreignKey:UserID;references:ID"`
	WalletTransaction []WalletTransaction `gorm:"foreignKey:WalletID;references:ID"`
}
