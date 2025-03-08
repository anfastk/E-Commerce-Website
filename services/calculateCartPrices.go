package services

import (
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
)

func CalculateCartPrices(cartItems []models.CartItem) (float64, float64, float64, float64, float64, int) {
	var (
		regularPrice    float64
		salePrice       float64
		productDiscount float64
		tax             float64
		totalDiscount   float64
	)
	shippingCharge := 100

	for _, item := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(item.ProductID, item.ProductVariant.CategoryID, item.ProductVariant.RegularPrice, item.ProductVariant.SalePrice)
		regularPrice += item.ProductVariant.RegularPrice * float64(item.Quantity)
		salePrice += (item.ProductVariant.SalePrice - discountAmount) * float64(item.Quantity)
	}

	tax = (salePrice * 18) / 100
	productDiscount = regularPrice - salePrice

	if salePrice < 1000 {
		totalDiscount = productDiscount
	} else {
		totalDiscount = productDiscount + float64(shippingCharge)
		shippingCharge = 0
	}

	return regularPrice, salePrice, tax, productDiscount, totalDiscount, shippingCharge
}
