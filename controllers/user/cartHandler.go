package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ShowCart(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

	type CartItemWithDiscount struct {
		Item          models.CartItem
		DiscountPrice float64
	}
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		cart = models.Cart{
			UserID: userID,
		}
		if createErr := config.DB.Create(&cart).Error; createErr != nil {
			config.DB.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Cart creation Failed", "Cart creation Failed", "")
		}
	}
	CreateWallet(c, userID)
	var cartItems []models.CartItem
	if err := config.DB.Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Product Not Found", "")
		return
	}
	var activeCartItems []CartItemWithDiscount
	for _, item := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(item.ProductVariantID, item.ProductVariant.CategoryID, item.ProductVariant.RegularPrice, item.ProductVariant.SalePrice)
		if item.ProductVariant.ID != 0 {
			activeCartItems = append(activeCartItems, CartItemWithDiscount{
				Item:          item,
				DiscountPrice: item.ProductVariant.SalePrice - discountAmount,
			})
		}
	}

	for i := range activeCartItems {
		if activeCartItems[i].Item.ProductVariant.StockQuantity < 3 {
			if activeCartItems[i].Item.ProductVariant.StockQuantity < activeCartItems[i].Item.Quantity {
				activeCartItems[i].Item.Quantity = activeCartItems[i].Item.ProductVariant.StockQuantity
				if createErr := config.DB.Model(&activeCartItems[i].Item).Updates(map[string]interface{}{
					"quantity": activeCartItems[i].Item.Quantity,
				}).Error; createErr != nil {
					helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Something Went Wrong", "")
					return
				}
			}
		}
	}

	var productIDs []uint
	for _, cartItem := range cartItems {
		productIDs = append(productIDs, cartItem.ProductID)
	}
	type suggestion struct {
		ID              uint     `json:"id"`
		ProductName     string   `json:"product_name"`
		RegularPrice    float64  `json:"regular_price"`
		SalePrice       float64  `json:"sale_price"`
		OfferPersentage int      `jso:"offer_persentage"`
		Images          []string `json:"images"`
	}
	var suggestionProduct []models.ProductVariantDetails
	if len(productIDs) == 0 {
		result := config.DB.Limit(4).
			Preload("VariantsImages", "is_deleted = ? ", false).
			Where("is_deleted = ? AND stock_quantity != ?", false, 0).
			Find(&suggestionProduct)
		if result.Error != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch ", "product variants", "")
			return
		}
	} else {
		result := config.DB.Limit(4).Preload("VariantsImages", "is_deleted = ?", false).
			Where("is_deleted = ? AND id NOT IN ? AND stock_quantity != ?", false, productIDs, 0).
			Find(&suggestionProduct)
		if result.Error != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product variants", "Failed to fetch product variants", "")
			return
		}
	}
	var suggest []suggestion
	for _, variant := range suggestionProduct {
		var images []string
		for _, img := range variant.VariantsImages {
			images = append(images, img.ProductVariantsImages)
		}
		DiscountAmount, TotalPercentage, _ := helper.DiscountCalculation(variant.ID, variant.CategoryID, variant.RegularPrice, variant.SalePrice)

		suggest = append(suggest, suggestion{
			ID:              variant.ID,
			ProductName:     variant.ProductName,
			RegularPrice:    variant.RegularPrice,
			SalePrice:       variant.SalePrice - DiscountAmount,
			OfferPersentage: int(TotalPercentage),
			Images:          images,
		})
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"Suggestion": suggest,
		"CartItem":   activeCartItems,
	})
}

func AddToCart(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

	var cart models.Cart
	tx := config.DB.Begin()

	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		cart = models.Cart{
			UserID: userID,
		}
		if createErr := tx.Create(&cart).Error; createErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Cart creation Failed", "Cart creation Failed", "")
		}
	}

	productID, iderr := strconv.Atoi(c.Param("id"))
	if iderr != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Product Id Not Found", "Product Id Not Found", "")
		return
	}
	var product models.ProductVariantDetails
	if err := tx.First(&product, productID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found ", "Product Not Found ", "")
		return
	}
	if product.StockQuantity == 0 {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusConflict, "Product Out Of Stock", "Product Out Of Stock", "")
		return
	}

	var cartItems models.CartItem
	if err := tx.First(&cartItems, "product_id = ? AND product_variant_id = ? AND cart_id = ?", product.ProductID, product.ID, cart.ID).Error; err != nil {
		cartItems = models.CartItem{
			CartID:           cart.ID,
			ProductID:        product.ProductID,
			ProductVariantID: uint(productID),
			Quantity:         1,
		}
		if createErr := tx.Create(&cartItems).Error; createErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
			return
		}
	} else {
		qty := cartItems.Quantity
		cartItems.Quantity = qty + 1
		if updateErr := tx.Model(&cartItems).Updates(cartItems).Error; updateErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
			return
		}
	}
	tx.Commit()
	helper.RespondWithError(c, http.StatusOK, "Add to Cart Success", "Add to Cart Success", "")
}

func CartItemUpdate(c *gin.Context) {
	itemID := c.Param("id")
	var cartItems models.CartItem
	if err := config.DB.First(&cartItems, itemID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Product Not Found", "")
	}

	var requestBody struct {
		Action string `json:"action"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request body", "Invalid request body", "")
		return
	}

	qty := cartItems.Quantity
	if requestBody.Action == "increase" {
		var product models.ProductVariantDetails
		if err := config.DB.First(&product, cartItems.ProductVariantID).Error; err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Product not found", "")
			return
		}

		if cartItems.Quantity < product.StockQuantity {

			if cartItems.Quantity < 3 {
				cartItems.Quantity = qty + 1
				if createErr := config.DB.Model(&cartItems).Updates(cartItems).Error; createErr != nil {
					helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
					return
				}
			} else {
				helper.RespondWithError(c, http.StatusBadRequest, "Maximum quantity exceeded", "Maximum quantity exceeded", "")
				return
			}

		} else {
			helper.RespondWithError(c, http.StatusBadRequest, "Not enough stock available", "Not enough stock available", "")
			return
		}

	} else {
		cartItems.Quantity = qty - 1
		if cartItems.Quantity <= 0 {
			if err := config.DB.Unscoped().Delete(&cartItems).Error; err != nil {
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
				return
			}
		} else {
			if createErr := config.DB.Model(&cartItems).Updates(cartItems).Error; createErr != nil {
				helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
				return
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "OK",
		"message":  "Quantity updated",
		"code":     http.StatusOK,
		"quantity": cartItems.Quantity,
	})
}

func ShowCartTotal(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

	var subTotal float64
	var total float64
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return
	}
	var cartItems []models.CartItem
	if err := config.DB.
		Preload("ProductVariant").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return
	}
	for _, items := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(items.ProductID, items.ProductVariant.CategoryID, items.ProductVariant.RegularPrice, items.ProductVariant.SalePrice)
		subTotal += items.ProductVariant.RegularPrice * float64(items.Quantity)
		total += (items.ProductVariant.SalePrice - discountAmount) * float64(items.Quantity)
		fmt.Println(subTotal, total, int(subTotal-total))

	}
	cartDiscountAmount := subTotal - total

	count := CartCount(c)
	c.JSON(http.StatusOK, gin.H{
		"status":         "OK",
		"message":        "Total fetch success",
		"Count":          count,
		"SubTotal":       subTotal,
		"DiscountAmount": cartDiscountAmount,
		"Total":          total,
		"code":           http.StatusOK,
	})
}
func CartCount(c *gin.Context) int {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return 0
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return 0
	}

	var count int
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return 0
	}
	var cartItems []models.CartItem
	if err := config.DB.Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return 0
	}
	var productIDs []uint
	for _, cartItem := range cartItems {
		productIDs = append(productIDs, cartItem.ProductID)
	}
	var product []models.ProductVariantDetails
	if err := config.DB.Find(&product, "id IN ?", productIDs).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return 0
	}
	count = len(product)

	return count
}

func DeleteCartItems(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Cart Not Found", "Something Went Wrong", "")
		return
	}

	itemID := c.Param("id")

	var cartItem models.CartItem
	if err := config.DB.First(&cartItem, "id = ? AND cart_id = ?", itemID, cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Inavlid request", "Inavlid request", "")
	}

	if err := config.DB.Unscoped().Delete(&cartItem).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Item delete successfully",
		"code":    http.StatusOK,
	})
}
