package models

import "gorm.io/gorm"

type AdminModel struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"unique,not null" json:"email"`
	Password string `gorm:"not null" json:"password"`
}
 