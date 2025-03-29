package services

import (
	"errors"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
) 

type CartItemDetailWithDiscount struct {
	CartItem       models.CartItem
	ProductImage   string
	ProductDetails models.ProductVariantDetails
	DiscountPrice  float64
}

func FetchCartItems(userID uint) (models.Cart, []CartItemDetailWithDiscount, error) {
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		return cart, []CartItemDetailWithDiscount{}, errors.New("Cart not found")
	}

	var cartItems []models.CartItem
	if err := config.DB.Order("created_at DESC").Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		return cart, []CartItemDetailWithDiscount{}, errors.New("Error fetching cart items")
	}

	var cartItemsDetails []CartItemDetailWithDiscount
	for _, item := range cartItems {
		var productDetail models.ProductVariantDetails
		if err := config.DB.Unscoped().Preload("Category").
			First(&productDetail, item.ProductVariantID).Error; err != nil {
			return cart, []CartItemDetailWithDiscount{}, errors.New("Error fetching cart items")
		}
		var productImage []models.ProductVariantsImage
		if err := config.DB.Select("product_variants_images").
			Where("product_variant_id = ?", item.ProductVariantID).
			Find(&productImage).Error; err != nil {
			return cart, []CartItemDetailWithDiscount{}, errors.New("Product Image Not Found")
		}
		discountAmount, _, _ := helper.DiscountCalculation(productDetail.ID, productDetail.CategoryID, productDetail.RegularPrice, productDetail.SalePrice)
		cartItemsDetails = append(cartItemsDetails, CartItemDetailWithDiscount{
			CartItem:       item,
			ProductDetails: productDetail,
			ProductImage:   productImage[0].ProductVariantsImages,
			DiscountPrice:  productDetail.SalePrice - discountAmount,
		})
	}

	return cart, cartItemsDetails, nil
}
