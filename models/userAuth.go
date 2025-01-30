package models

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	FullName   string `gorm:"not null,size:255" json:"name"`
	Email      string `gorm:"unique,not null,size:255" json:"email"`
	Password   string `gorm:"not null,size:255" json:"password"`
	GoogleID   string `gorm:"type:varchar(100)"`
	ProfilePic string
	Status     string `gorm:"not null;default:Active;check:status IN ('Active','Blocked')"`
	IsDeleted  bool   `gorm:"default:false"`
	IsBlocked  bool   `gorm:"default:false"`
	IsVerified bool
}