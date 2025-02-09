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
	var (
		regularPrice    float64
		salePrice       float64
		productDiscount float64
		totalDiscount   float64
		tax             float64
		total           float64
	)
	shippingCharge := 100
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "", "Something Went Wrong", "/cart")
		return
	}
	var cartItems []models.CartItem
	if err := config.DB.Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "", "Product Not Found", "/cart")
		return
	}

	if len(cartItems) == 0 {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/cart")
		return
	}

	for _, item := range cartItems {
		if item.ProductVariant.StockQuantity <item.Quantity {
			helper.RespondWithError(c, http.StatusConflict, "Stock unavailable", "One or more items in your cart are out of stock. Please update your cart.", "/cart")
			return
		}
	}
	
	for _, items := range cartItems {
		regularPrice += items.ProductVariant.RegularPrice * float64(items.Quantity)
		salePrice += items.ProductVariant.SalePrice * float64(items.Quantity)
	}
	tax = (salePrice * 18) / 100
	productDiscount = regularPrice - salePrice

	if salePrice < 1000 {
		totalDiscount = productDiscount
	} else {
		totalDiscount = productDiscount + float64(shippingCharge)
		shippingCharge = 0
	}
	total = salePrice + tax

	var address []models.UserAddress
	if err := config.DB.Order("updated_at DESC").Find(&address, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	c.HTML(http.StatusOK, "checkOut.html", gin.H{
		"status":          "OK",
		"message":         "Checkout fetch success",
		"Address":         address,
		"CartItem":        cartItems,
		"SubTotal":        regularPrice,
		"Shipping":        shippingCharge,
		"Tax":             tax,
		"ProductDiscount": productDiscount,
		"TotalDiscount":   totalDiscount,
		"Total":           total,
		"code":            http.StatusOK,
	})
}
