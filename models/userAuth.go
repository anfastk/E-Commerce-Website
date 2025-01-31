package models

import "gorm.io/gorm"

type UserAuth struct {
	gorm.Model
	FullName   string `gorm:"type:varchar(255);not null" json:"name"`
	Email      string `gorm:"type:varchar(255);unique;not null" json:"email"`
	Password   string `gorm:"type:varchar(255);not null" json:"password"`
	GoogleID   string `gorm:"type:varchar(100)" json:"google_id"`
	ProfilePic string `gorm:"type:varchar(255)" json:"profile_pic"`
	Status     string `gorm:"CHECK IN ('Active','Blocked','Deleted');not null;default:'Active'" json:"status"`
	IsDeleted  bool   `gorm:"default:false" json:"is_deleted"`
	IsBlocked  bool   `gorm:"default:false" json:"is_blocked"`
	IsVerified bool   `gorm:"default:false" json:"is_verified"`
}