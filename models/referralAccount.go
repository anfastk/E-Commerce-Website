package models

import "gorm.io/gorm"

type ReferralAccount struct {
	gorm.Model
	UserID         uint `gorm:"not null"`
	Count          uint
	Balance        float64
	UserAuth       UserAuth         `gorm:"foreignKey:UserID;references:ID"`
	ReferalHistory []ReferalHistory `gorm:"foreignKey:ReferralID;references:ID"`
}
