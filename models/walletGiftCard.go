package models

import (
	"time"

	"gorm.io/gorm"
)

type WalletGiftCard struct {
	gorm.Model
	GiftCardCode    string `gorm:"unique"`
	GiftCardValue   float64
	ExpDate         time.Time
	UserID          uint   `gorm:"not null;index"`
	WalletID        uint   `gorm:"not null;index"`
	ReceiverName    string `gorm:"size:255"`
	ReceiverEmail   string `gorm:"size:255"`
	PaymentStatus   bool   `gorm:"default:false"`
	IsValid         bool   `gorm:"default:true"`
	Status          string `gorm:"default:'Active'"`
	RedeemedUserID  uint
	RedeemedAt      time.Time
	TransactionType string   `gorm:"size:50"`
	TransactionID   string   `gorm:"size:100"`
	Sender          UserAuth `gorm:"foreignKey:UserID"`
	Reciver         UserAuth `gorm:"foreignKey:RedeemedUserID"`
	WalletDetail    Wallet   `gorm:"foreignKey:WalletID"`
}
