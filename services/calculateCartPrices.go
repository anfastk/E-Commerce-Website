package services

import (
	"github.com/anfastk/E-Commerce-Website/utils/helper"
)

func CalculateCartPrices(cartItems []CartItemDetailWithDiscount) (float64, float64, float64, float64, float64, int) {
	var (
		regularPrice    float64
		salePrice       float64
		productDiscount float64
		tax             float64
		totalDiscount   float64
	)
	shippingCharge := 100

	for _, item := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(item.CartItem.ProductID, item.ProductDetails.CategoryID, item.ProductDetails.RegularPrice, item.ProductDetails.SalePrice)
		regularPrice += item.ProductDetails.RegularPrice * float64(item.CartItem.Quantity)
		salePrice += (item.ProductDetails.SalePrice - discountAmount) * float64(item.CartItem.Quantity)
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
