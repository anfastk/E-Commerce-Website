package controllers

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ShowCheckoutPage(c *gin.Context) {
	userID := c.MustGet("userid")
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound,"","Something Went Wrong","/cart")
	}
	var cartItems []models.CartItem
	if err := config.DB.Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound,"","Product Not Found","/cart")
		return
	}
	var activeCartItems []models.CartItem
	for _, item := range cartItems {
		if item.ProductVariant.ID != 0 && item.ProductVariant.StockQuantity != 0 {
			activeCartItems = append(activeCartItems, item)
		}
	}
	for i := range activeCartItems {
		if activeCartItems[i].ProductVariant.StockQuantity < 3 {
			if activeCartItems[i].ProductVariant.StockQuantity < activeCartItems[i].Quantity {
				helper.RespondWithError(c,http.StatusConflict,"Stock unavailable","One or more items in your cart are out of stock. Please update your cart.","/cart")
			}
		}
	}
}
