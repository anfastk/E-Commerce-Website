package models

import "gorm.io/gorm"

type UserProfile struct {
	gorm.Model
	UserID     uint   `gorm:"unique"`
	Mobile     string `gorm:"unique;size:15" json:"number"`
	Country    string `gorm:"size:100" json:"user_country"`
	State      string `gorm:"size:100" json:"user_state"`
	Pincode    string `gorm:"size:10"  json:"pincode"`
	UserAuth   UserAuth
}
