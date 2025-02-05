package controllers

import (
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ShowCart(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		cart = models.Cart{
			UserID: userID,
		}
		if createErr := config.DB.Create(&cart).Error; createErr != nil {
			config.DB.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Cart creation Failed")
		}
	}
	var cartItems []models.CartItem
	if err := config.DB.Preload("ProductDetail").
		Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Specification").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Product Not Found")
		return
	}
	for i := range cartItems {
		if cartItems[i].ProductVariant.StockQuantity < 3 {
			if cartItems[i].Quantity >= 3 {
				cartItems[i].Quantity = cartItems[i].ProductVariant.StockQuantity
				if createErr := config.DB.Model(&cartItems[i]).Updates(cartItems[i]).Error; createErr != nil {
					helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed")
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
		ID           uint     `json:"id"`
		ProductName  string   `json:"product_name"`
		RegularPrice float64  `json:"regular_price"`
		SalePrice    float64  `json:"sale_price"`
		Images       []string `json:"images"`
	}
	var suggestionProduct []models.ProductVariantDetails
	if len(productIDs) == 0 {
		result := config.DB.Limit(4).
			Preload("VariantsImages", "is_deleted = ?", false).
			Where("is_deleted = ?", false).
			Find(&suggestionProduct)
		if result.Error != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product variants")
			return
		}
	} else {
		result := config.DB.Limit(4).Preload("VariantsImages", "is_deleted = ?", false).
			Where("is_deleted = ? AND id NOT IN ?", false, productIDs).
			Find(&suggestionProduct)
		if result.Error != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product variants")
			return
		}
	}
	var suggest []suggestion
	for _, variant := range suggestionProduct {
		var images []string
		for _, img := range variant.VariantsImages {
			images = append(images, img.ProductVariantsImages)
		}

		suggest = append(suggest, suggestion{
			ID:           variant.ID,
			ProductName:  variant.ProductName,
			RegularPrice: variant.RegularPrice,
			SalePrice:    variant.SalePrice,
			Images:       images,
		})
	}

	c.HTML(http.StatusOK, "cart.html", gin.H{
		"Suggestion": suggest,
		"CartItem":   cartItems,
	})
}

func AddToCart(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var cart models.Cart
	tx := config.DB.Begin()

	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		cart = models.Cart{
			UserID: userID,
		}
		if createErr := tx.Create(&cart).Error; createErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Cart creation Failed")
		}
	}
	productID, iderr := strconv.Atoi(c.Param("id"))
	if iderr != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Product Id Not Found")
		return
	}
	var product models.ProductVariantDetails
	if err := tx.First(&product, productID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Product Id Not Found")
		return
	}

	var cartItems models.CartItem
	if err := tx.First(&cartItems, "product_id = ? AND cart_id = ?", productID, cart.ID).Error; err != nil {
		cartItems = models.CartItem{
			CartID:           cart.ID,
			ProductID:        product.ProductID,
			ProductVariantID: uint(productID),
			Quantity:         1,
		}
		if createErr := tx.Create(&cartItems).Error; createErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed")
			return
		}
	} else {
		qty := cartItems.Quantity
		cartItems.Quantity = qty + 1
		if updateErr := tx.Model(&cartItems).Updates(cartItems).Error; updateErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed")
			return
		}
	}
	tx.Commit()
	helper.RespondWithError(c, http.StatusOK, "Add to Cart Success")
}

func CartItemUpdate(c *gin.Context) {
	itemID := c.Param("id")
	var cartItems models.CartItem
	if err := config.DB.First(&cartItems, itemID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Product Not Found")
	}

	var requestBody struct {
		Action string `json:"action"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	qty := cartItems.Quantity
	if requestBody.Action == "increase" {
		var product models.ProductVariantDetails
		if err := config.DB.First(&product, cartItems.ProductVariantID).Error; err != nil {
			helper.RespondWithError(c, http.StatusBadRequest, "Product not found")
			return
		}

		if product.StockQuantity >= 3 {

			if cartItems.Quantity < 3 {
				cartItems.Quantity = qty + 1
				if createErr := config.DB.Model(&cartItems).Updates(cartItems).Error; createErr != nil {
					helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed")
					return
				}
			} else {
				helper.RespondWithError(c, http.StatusBadRequest, "Maximum quantity exceeded")
				return
			}

		} else {
			helper.RespondWithError(c, http.StatusBadRequest, "Not enough stock available")
			return
		}

	} else {
		cartItems.Quantity = qty - 1
		if cartItems.Quantity == 0 {
			if err := config.DB.Unscoped().Delete(&cartItems).Error; err != nil {
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item")
				return
			}
		} else {
			if createErr := config.DB.Model(&cartItems).Updates(cartItems).Error; createErr != nil {
				helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed")
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
	userID := c.MustGet("userid").(uint)
	var total float64
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something went wrong")
		return
	}
	var cartItems []models.CartItem
	if err := config.DB.
		Preload("ProductVariant").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Something went wrong")
		return
	}
	for _, items := range cartItems {
		total += items.ProductVariant.SalePrice * float64(items.Quantity)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Total fetch success",
		"Count":   len(cartItems),
		"Total":   total,
		"code":    http.StatusOK,
	})
}

func DeleteCartItems(c *gin.Context) {
	itemID := c.Param("id")

	var cartItem models.CartItem
	if err := config.DB.First(&cartItem, itemID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Inavlid request")
	}

	if err := config.DB.Unscoped().Delete(&cartItem).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item")
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Item delete successfully",
		"code":    http.StatusOK,
	})
}
