package models

import (
	"time"

	"gorm.io/gorm"
)

type Sale struct {
	gorm.Model
	TotalSalesAmount float64
	ProductCount     int
	PaymentStatus    string
	SaleDate         time.Time
	OrderID          uint
}
