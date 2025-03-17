package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model
	OrderID               uint                  `gorm:"not null;index"`
	UserID                uint                  `gorm:"not null;index"`
	ProductVariantID      uint                  `gorm:"not null;index"`
	Quantity              int                   `gorm:"not null;index"`
	ProductImage          string                `gorm:"not null"`
	ProductName           string                `gorm:"not null;index"`
	ProductSummary        string                `gorm:"not null"`
	ProductCategory       string                `gorm:"not null"`
	ProductRegularPrice   float64               `gorm:"index;type:numeric(10,2);not null"`
	ProductSalePrice      float64               `gorm:"index;type:numeric(10,2);not null"`
	SubTotal              float64               `gorm:"index;type:numeric(10,2)" json:"subtotal"`
	Tax                   float64               `gorm:"index;type:numeric(10,2)" json:"tax"`
	Total                 float64               `gorm:"index;type:numeric(10,2)" json:"total"`
	OrderStatus           string                `gorm:"index;type:varchar(255);index;default:'Pending'" json:"status"`
	IsDelivered           bool                  `gorm:"index;default:false"`
	ExpectedDeliveryDate  time.Time             `gorm:"index;not null"`
	ReturnableStatus      bool                  `gorm:"default:true"`
	OrderUID              string                `gorm:"unique,not null" json:"orderuid"`
	Reason				  string
	ReturnDate            time.Time             
	DeliveryDate          time.Time             
	ShippedDate           time.Time             
	OutOfDeliveryDate     time.Time 
	CancelDate            time.Time
	UserAuth              UserAuth              `gorm:"foreignKey:UserID;references:ID"`
	ProductVariantDetails ProductVariantDetails `gorm:"foreignKey:ProductVariantID;references:ID"`
}
