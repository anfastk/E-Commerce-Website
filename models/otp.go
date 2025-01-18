package models

import (
	"gorm.io/gorm"
	"time"
)

type Otp struct {
	gorm.Model
	Email      string `gorm:"size:255" json:"otp_email"`
	OTP        string `gorm:"size:10"  json:"otp"`
	ExpireTime time.Time
}
