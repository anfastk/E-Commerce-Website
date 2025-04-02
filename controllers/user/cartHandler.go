package controllers

import (
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ShowCart(c *gin.Context) {
	logger.Log.Info("Requested to show cart")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	type CartItemDetails struct {
		CartItem      models.CartItem
		ProductImage  string
		ProductDetail models.ProductVariantDetails
		DiscountPrice float64
		Status        string
	}

	helper.CreateCart(c, userID)

	helper.CreateWallet(c, userID)
	_, cartItemDetail, err := services.FetchCartItems(userID)
	if err != nil {
		logger.Log.Error("Failed to fetch cart items", zap.Uint("userID", userID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, err.Error(), "Something Went Wrong", "")
		return
	}

	var cartItemResponceDetails []CartItemDetails
	for _, item := range cartItemDetail {
		status := ""
		if item.ProductDetails.StockQuantity == 0 {
			status = "Out Of Stock"
		}
		if item.ProductDetails.IsDeleted {
			status = "Unavailable"
		}
		cartItemResponceDetails = append(cartItemResponceDetails, CartItemDetails{
			CartItem:      item.CartItem,
			ProductImage:  item.ProductImage,
			ProductDetail: item.ProductDetails,
			DiscountPrice: item.DiscountPrice,
			Status:        status,
		})
	}

	for i := range cartItemResponceDetails {
		if cartItemResponceDetails[i].ProductDetail.StockQuantity < 3 && cartItemResponceDetails[i].Status == "" {
			if cartItemResponceDetails[i].ProductDetail.StockQuantity < cartItemResponceDetails[i].CartItem.Quantity {
				cartItemResponceDetails[i].CartItem.Quantity = cartItemResponceDetails[i].ProductDetail.StockQuantity
				if createErr := config.DB.Model(&cartItemResponceDetails[i].CartItem).Updates(map[string]interface{}{
					"quantity": cartItemResponceDetails[i].CartItem.Quantity,
				}).Error; createErr != nil {
					logger.Log.Error("Failed to update cart item quantity",
						zap.Uint("cartItemID", cartItemResponceDetails[i].CartItem.ID),
						zap.Error(createErr))
					helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Something Went Wrong", "")
					return
				}
				logger.Log.Info("Adjusted cart item quantity due to low stock",
					zap.Uint("cartItemID", cartItemResponceDetails[i].CartItem.ID),
					zap.Int("newQuantity", int(cartItemResponceDetails[i].CartItem.Quantity)))
			}
		}
	}

	var productIDs []uint
	for _, cartItem := range cartItemDetail {
		productIDs = append(productIDs, cartItem.CartItem.ProductID)
	}

	type suggestion struct {
		ID              uint    `json:"id"`
		ProductName     string  `json:"product_name"`
		RegularPrice    float64 `json:"regular_price"`
		SalePrice       float64 `json:"sale_price"`
		OfferPersentage int     `json:"offer_persentage"`
		Images          string  `json:"images"`
		CategoryName    string  `json:"category_name"`
		IsInWishlist    bool    `json:"is_in_wishlist"`
	}

	var suggestionProduct []models.ProductVariantDetails
	if len(productIDs) == 0 {
		result := config.DB.Limit(4).
			Preload("VariantsImages", "is_deleted = ? ", false).
			Where("is_deleted = ? AND stock_quantity > ?", false, 0).
			Find(&suggestionProduct)
		if result.Error != nil {
			logger.Log.Error("Failed to fetch suggestion products (empty cart)", zap.Error(result.Error))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch ", "product variants", "")
			return
		}
	} else {
		result := config.DB.Limit(4).Preload("VariantsImages").Preload("Category").
			Where("is_deleted = ? AND id NOT IN ? AND stock_quantity > ?", false, productIDs, 0).
			Find(&suggestionProduct)
		if result.Error != nil {
			logger.Log.Error("Failed to fetch suggestion products", zap.Error(result.Error))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product variants", "Failed to fetch product variants", "")
			return
		}
	}

	var suggest []suggestion
	for _, variant := range suggestionProduct {
		var wishlistItems models.WishlistItem
		IsInWishlist := true
		if err := config.DB.Where("wishlist_id = (SELECT id FROM wishlists WHERE user_id = ?) AND product_variant_id = ?", userID, variant.ID).First(&wishlistItems).Error; err != nil {
			IsInWishlist = false
		}
		DiscountAmount, TotalPercentage, _ := helper.DiscountCalculation(variant.ProductID, variant.CategoryID, variant.RegularPrice, variant.SalePrice)

		suggest = append(suggest, suggestion{
			ID:              variant.ID,
			ProductName:     variant.ProductName,
			RegularPrice:    variant.RegularPrice,
			SalePrice:       variant.SalePrice - DiscountAmount,
			OfferPersentage: int(TotalPercentage),
			Images:          variant.VariantsImages[0].ProductVariantsImages,
			CategoryName:    variant.Category.Name,
			IsInWishlist:    IsInWishlist,
		})
	}

	logger.Log.Info("Cart page loaded successfully",
		zap.Uint("userID", userID),
		zap.Int("cartItemCount", len(cartItemResponceDetails)),
		zap.Int("suggestionCount", len(suggest)))
	c.HTML(http.StatusOK, "cart.html", gin.H{
		"Suggestion": suggest,
		"CartItem":   cartItemResponceDetails,
	})
}

func AddToCart(c *gin.Context) {
	logger.Log.Info("Requested to add item to cart")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var cart models.Cart
	tx := config.DB.Begin()

	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		cart = models.Cart{
			UserID: userID,
		}
		if createErr := tx.Create(&cart).Error; createErr != nil {
			logger.Log.Error("Failed to create cart", zap.Uint("userID", userID), zap.Error(createErr))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Cart creation Failed", "Cart creation Failed", "")
			return
		}
		logger.Log.Info("Created new cart", zap.Uint("cartID", cart.ID))
	}

	productID, iderr := strconv.Atoi(c.Param("id"))
	if iderr != nil {
		logger.Log.Error("Invalid product ID", zap.String("productID", c.Param("id")), zap.Error(iderr))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Product Id Not Found", "Product Id Not Found", "")
		return
	}

	var product models.ProductVariantDetails
	if err := tx.First(&product, productID).Error; err != nil {
		logger.Log.Error("Product not found", zap.Int("productID", productID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found ", "Product Not Found ", "")
		return
	}

	if product.StockQuantity == 0 {
		logger.Log.Warn("Product out of stock", zap.Int("productID", productID))
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
			logger.Log.Error("Failed to create cart item",
				zap.Int("productID", productID),
				zap.Error(createErr))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
			return
		}
	} else if cartItems.Quantity < 3 {
		qty := cartItems.Quantity
		cartItems.Quantity = qty + 1
		if updateErr := tx.Model(&cartItems).Updates(cartItems).Error; updateErr != nil {
			logger.Log.Error("Failed to update cart item quantity",
				zap.Uint("cartItemID", cartItems.ID),
				zap.Error(updateErr))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
			return
		}
	}

	tx.Commit()
	logger.Log.Info("Item added to cart successfully",
		zap.Uint("userID", userID),
		zap.Int("productID", productID),
		zap.Int("quantity", int(cartItems.Quantity)))
	helper.RespondWithError(c, http.StatusOK, "Add to Cart Success", "Add to Cart Success", "")
}

func CartItemUpdate(c *gin.Context) {
	logger.Log.Info("Requested to update cart item")

	itemID := c.Param("id")
	var cartItems models.CartItem
	if err := config.DB.First(&cartItems, itemID).Error; err != nil {
		logger.Log.Error("Cart item not found", zap.String("itemID", itemID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Product Not Found", "")
		return
	}

	var requestBody struct {
		Action string `json:"action"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		logger.Log.Error("Invalid request body", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request body", "Invalid request body", "")
		return
	}

	qty := cartItems.Quantity
	if requestBody.Action == "increase" {
		var product models.ProductVariantDetails
		if err := config.DB.First(&product, cartItems.ProductVariantID).Error; err != nil {
			logger.Log.Error("Product not found for cart item",
				zap.Uint("productVariantID", cartItems.ProductVariantID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Product not found", "")
			return
		}

		if cartItems.Quantity < product.StockQuantity {
			if cartItems.Quantity < 3 {
				cartItems.Quantity = qty + 1
				if createErr := config.DB.Model(&cartItems).Updates(cartItems).Error; createErr != nil {
					logger.Log.Error("Failed to increase cart item quantity",
						zap.Uint("cartItemID", cartItems.ID),
						zap.Error(createErr))
					helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
					return
				}
			} else {
				logger.Log.Warn("Maximum quantity exceeded",
					zap.Uint("cartItemID", cartItems.ID))
				helper.RespondWithError(c, http.StatusBadRequest, "Maximum quantity exceeded", "Maximum quantity exceeded", "")
				return
			}
		} else {
			logger.Log.Warn("Not enough stock available",
				zap.Uint("cartItemID", cartItems.ID),
				zap.Int("stockQuantity", int(product.StockQuantity)))
			helper.RespondWithError(c, http.StatusBadRequest, "Not enough stock available", "Not enough stock available", "")
			return
		}
	} else {
		cartItems.Quantity = qty - 1
		if cartItems.Quantity <= 0 {
			if err := config.DB.Unscoped().Delete(&cartItems).Error; err != nil {
				logger.Log.Error("Failed to delete cart item",
					zap.Uint("cartItemID", cartItems.ID),
					zap.Error(err))
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
				return
			}
			logger.Log.Info("Cart item deleted due to zero quantity",
				zap.Uint("cartItemID", cartItems.ID))
		} else {
			if createErr := config.DB.Model(&cartItems).Updates(cartItems).Error; createErr != nil {
				logger.Log.Error("Failed to decrease cart item quantity",
					zap.Uint("cartItemID", cartItems.ID),
					zap.Error(createErr))
				helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
				return
			}
		}
	}

	logger.Log.Info("Cart item quantity updated successfully",
		zap.Uint("cartItemID", cartItems.ID),
		zap.Int("newQuantity", int(cartItems.Quantity)))
	c.JSON(http.StatusOK, gin.H{
		"status":   "OK",
		"message":  "Quantity updated",
		"code":     http.StatusOK,
		"quantity": cartItems.Quantity,
	})
}

func ShowCartTotal(c *gin.Context) {
	logger.Log.Info("Requested cart total")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var subTotal float64
	var total float64
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found", zap.Uint("userID", userID), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return
	}

	var cartItems []models.CartItem
	if err := config.DB.
		Preload("ProductVariant").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch cart items for total",
			zap.Uint("cartID", cart.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return
	}

	for _, items := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(items.ProductID, items.ProductVariant.CategoryID, items.ProductVariant.RegularPrice, items.ProductVariant.SalePrice)
		subTotal += items.ProductVariant.RegularPrice * float64(items.Quantity)
		total += (items.ProductVariant.SalePrice - discountAmount) * float64(items.Quantity)
	}
	cartDiscountAmount := subTotal - total

	count := CartCount(c)
	logger.Log.Info("Cart total calculated successfully",
		zap.Uint("userID", userID),
		zap.Float64("subTotal", subTotal),
		zap.Float64("total", total),
		zap.Int("itemCount", count))
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
	logger.Log.Info("Requested cart count")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var count int
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found for count", zap.Uint("userID", userID), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return 0
	}

	var cartItems []models.CartItem
	if err := config.DB.Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch cart items for count",
			zap.Uint("cartID", cart.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return 0
	}

	var productIDs []uint
	for _, cartItem := range cartItems {
		productIDs = append(productIDs, cartItem.ProductID)
	}

	var product []models.ProductVariantDetails
	if err := config.DB.Unscoped().Find(&product, "id IN ?", productIDs).Error; err != nil {
		logger.Log.Error("Failed to fetch products for count",
			zap.Any("productIDs", productIDs),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
		return 0
	}
	count = len(product)

	logger.Log.Info("Cart count retrieved successfully",
		zap.Uint("userID", userID),
		zap.Int("count", count))
	return count
}

func DeleteCartItems(c *gin.Context) {
	logger.Log.Info("Requested to delete cart item")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found", zap.Uint("userID", userID), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Cart Not Found", "Something Went Wrong", "")
		return
	}

	itemID := c.Param("id")
	var cartItem models.CartItem
	if err := config.DB.First(&cartItem, "id = ? AND cart_id = ?", itemID, cart.ID).Error; err != nil {
		logger.Log.Error("Cart item not found",
			zap.String("itemID", itemID),
			zap.Uint("cartID", cart.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Inavlid request", "Inavlid request", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&cartItem).Error; err != nil {
		logger.Log.Error("Failed to delete cart item",
			zap.Uint("cartItemID", cartItem.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
		return
	}

	logger.Log.Info("Cart item deleted successfully",
		zap.Uint("userID", userID),
		zap.Uint("cartItemID", cartItem.ID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Item delete successfully",
		"code":    http.StatusOK,
	})
}
