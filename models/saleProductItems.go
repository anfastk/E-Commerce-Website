package models

import "gorm.io/gorm"

type SalesProductItem struct {
	gorm.Model
	ProductVariantsID uint
	Quantity          int
	Price             float64 `gorm:"type:numeric(10,2)"`
	ProductID         uint
	SaleID            uint
}
