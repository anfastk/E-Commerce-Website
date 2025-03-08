package services

import (
	"errors"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
)

func FetchCartItems(userID uint) (models.Cart, []models.CartItem, error) {
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		return cart, nil, errors.New("Cart not found")
	}

	var cartItems []models.CartItem
	if err := config.DB.Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		return cart, nil, errors.New("Error fetching cart items")
	}

	if len(cartItems) == 0 {
		return cart, nil, errors.New("Cart is empty")
	}
	return cart, cartItems, nil
}
