package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderItem struct {
	gorm.Model
	OrderID               uint                  `gorm:"not null"`
	UserID                uint                  `gorm:"not null"`
	ProductVariantID      uint                  `gorm:"not null"`
	Quantity              int                   `gorm:"not null"`
	ProductImage          string                `gorm:"not null"`
	ProductName           string                `gorm:"not null"`
	ProductSummary        string                `gorm:"not null"`
	ProductCategory       string                `gorm:"not null"`
	ProductRegularPrice   float64               `gorm:"type:numeric(10,2);not null"`
	ProductSalePrice      float64               `gorm:"type:numeric(10,2);not null"`
	CouponDiscount        float64 				`gorm:"type:numeric(10,2)"`
	SubTotal              float64               `gorm:"type:numeric(10,2)" json:"subtotal"`
	ShippingCharge        float64               `gorm:"type:numeric(10,2)" json:"shippingcharge"`
	Tax                   float64               `gorm:"type:numeric(10,2)" json:"tax"`
	Total                 float64               `gorm:"type:numeric(10,2)" json:"total"`
	OrderStatus           string                `gorm:"type:varchar(255);default:'Pending'" json:"status"`
	IsDelivered           bool                  `gorm:"default:false"`
	ExpectedDeliveryDate  time.Time             `gorm:"not null"`
	ReturnableStatus      bool                  `gorm:"default:true"`
	OrderUID              string                `gorm:"unique,not null" json:"orderuid"`
	ReturnDate            time.Time             
	DeliveryDate          time.Time             
	ShippedDate           time.Time             
	OutOfDeliveryDate     time.Time             
	UserAuth              UserAuth              `gorm:"foreignKey:UserID;references:ID"`
	ProductVariantDetails ProductVariantDetails `gorm:"foreignKey:ProductVariantID;references:ID"`
}
