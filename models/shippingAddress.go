package models

import (
	"gorm.io/gorm"
)

type ShippingAddress struct {
	gorm.Model
	UserID    uint     `gorm:"not null;index"`
	OrderID   uint     `gorm:"not null;index"`
	FirstName string   `gorm:"size:100"`
	LastName  string   `gorm:"size:100"`
	Mobile    string   `gorm:"size:15"`
	Address   string   `gorm:"size:255"`
	Landmark  string   `gorm:"size:255"`
	Country   string   `gorm:"size:100"`
	State     string   `gorm:"size:100"`
	City      string   `gorm:"size:100"`
	PinCode   string   `gorm:"not null"`
	UserAuth  UserAuth `gorm:"foreignKey:UserID;references:ID"`
}
 