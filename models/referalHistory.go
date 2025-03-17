package models

import (
	"gorm.io/gorm"
)

type ReferalHistory struct {
	gorm.Model
	ReferralID      uint            `gorm:"not null;index"`
	JoinedUserId    uint            `gorm:"not null;index"`
	Status          string          `gorm:"type:varchar(10);default:'Pending'"`
	Reward          float64         `gorm:"not null"`
	JoinedUser      UserAuth        `gorm:"foreignKey:JoinedUserId;references:ID"`
	ReferralAccount ReferralAccount `gorm:"foreignKey:ReferralID;references:ID"`
}
