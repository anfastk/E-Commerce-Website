package models

import (
	"time"

	"gorm.io/gorm"
)

type WalletGiftCard struct {
	gorm.Model
	GiftCardCode   string    `gorm:"unique;not null"`
	GiftCardValue  float64   `gorm:"not null;index"`
	ExpDate        time.Time `gorm:"not null;index"`
	UserID         uint      `gorm:"not null;index"`
	RecipientName  string    `gorm:"size:255"` 
	RecipientEmail string    `gorm:"size:255"`
	Message        string    `gorm:"size:1000"`
	Status         string    `gorm:"default:Active"`
	PaymentMethod  string    `gorm:"size:50"`
	TransactionID  string    `gorm:"size:100"`
	RedeemedUserID *uint     `gorm:"index"`
	RedeemedAt     *time.Time
	Sender         UserAuth  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	RedeemedUser   *UserAuth `gorm:"foreignKey:RedeemedUserID;constraint:OnDelete:SET NULL"`
}
