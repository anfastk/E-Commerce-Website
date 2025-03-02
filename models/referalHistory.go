package models

import (
	"gorm.io/gorm"
)

type ReferalHistory struct {
	gorm.Model
	UserId       uint     `gorm:"not null"`
	JoinedUserId uint     `gorm:"not null"`
	Status       string   `gorm:"type:varchar(10);default:'Pending'"`
	Reward       float64  `gorm:"not null"`
	User         UserAuth `gorm:"foreignKey:UserId;references:ID"`
	JoinedUser   UserAuth `gorm:"foreignKey:JoinedUserId;references:ID"`
}
