package models

import "gorm.io/gorm"

type ReturnRequest struct {
	gorm.Model
	RequestUID            string                `gorm:"not null;index"`
	OrderItemID           uint                  `gorm:"not null;index"`
	ProductVariantID      uint                  `gorm:"not null;index"`
	UserID                uint                  `gorm:"not null;index"`
	Reason                string                `gorm:"not null"`
	AdditionalDetails     string                `gorm:"not null"`
	AdminNotes            string              
	Status                string                `gorm:"default:'pending'"`
	UserDetails           UserAuth              `gorm:"foreignKey:UserID;references:ID"`
	ProductVariantDetails ProductVariantDetails `gorm:"foreignKey:ProductVariantID;references:ID"`
	OrderItem             OrderItem             `gorm:"foreignKey:OrderItemID;references:ID"`
}
 