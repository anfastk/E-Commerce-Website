package controllers

import (
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ShowWishlist(c *gin.Context) {
	logger.Log.Info("Showing wishlist")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	type response struct {
		WishListID          uint
		ProductID           uint
		ProductName         string
		ProductImage        string
		ProductRegularPrice float64
		ProductSalePrice    float64
		DiscountPercentage  int
		IsInCart            bool
		IsStockAvailable    bool
	}

	var wishlist models.Wishlist
	if err := config.DB.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		wishlist = models.Wishlist{UserID: userID}
		if createErr := config.DB.Create(&wishlist).Error; createErr != nil {
			logger.Log.Error("Failed to create wishlist",
				zap.Uint("userID", userID),
				zap.Error(createErr))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create wishlist", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Wishlist created",
			zap.Uint("userID", userID),
			zap.Uint("wishlistID", wishlist.ID))
	}

	var wishlistItems []models.WishlistItem
	if err := config.DB.Preload("ProductVariantDetails").
		Preload("ProductVariantDetails.VariantsImages").
		Find(&wishlistItems, "wishlist_id = ?", wishlist.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch wishlist items",
			zap.Uint("wishlistID", wishlist.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
		return
	}

	var wishlistResponse []response
	for _, item := range wishlistItems {
		var cart models.Cart
		if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
			logger.Log.Error("Cart not found",
				zap.Uint("userID", userID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Cart Not Found", "Something Went Wrong", "")
			return
		}

		isInCart := true
		var cartItem models.CartItem
		if err := config.DB.First(&cartItem, "cart_id = ? AND product_id = ? AND product_variant_id = ?", cart.ID, item.ProductID, item.ProductVariantID).Error; err != nil {
			isInCart = false
		}

		discountAmount, discountPercentage, _ := helper.DiscountCalculation(item.ProductVariantID, item.ProductVariantDetails.CategoryID, item.ProductVariantDetails.RegularPrice, item.ProductVariantDetails.SalePrice)
		items := response{
			WishListID:          item.ID,
			ProductID:           item.ProductVariantDetails.ID,
			ProductName:         item.ProductVariantDetails.ProductName,
			ProductImage:        item.ProductVariantDetails.VariantsImages[0].ProductVariantsImages,
			ProductRegularPrice: item.ProductVariantDetails.RegularPrice,
			ProductSalePrice:    item.ProductVariantDetails.SalePrice - discountAmount,
			DiscountPercentage:  int(discountPercentage),
			IsStockAvailable:    item.ProductVariantDetails.StockQuantity > 0,
			IsInCart:            isInCart,
		}
		wishlistResponse = append(wishlistResponse, items)
	}

	count := len(wishlistResponse)
	logger.Log.Info("Wishlist loaded",
		zap.Uint("userID", userID),
		zap.Uint("wishlistID", wishlist.ID),
		zap.Int("itemCount", count))
	c.HTML(http.StatusOK, "wishlist.html", gin.H{
		"status":  "Success",
		"Data":    wishlistResponse,
		"Count":   count,
		"message": "Added to Wishlist Successfully",
		"code":    http.StatusOK,
	})
}

func AddToWishlist(c *gin.Context) {
	logger.Log.Info("Adding to wishlist")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	tx := config.DB.Begin()

	var wishlist models.Wishlist
	if err := tx.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		wishlist = models.Wishlist{UserID: userID}
		if createErr := tx.Create(&wishlist).Error; createErr != nil {
			logger.Log.Error("Failed to create wishlist",
				zap.Uint("userID", userID),
				zap.Error(createErr))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create wishlist", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Wishlist created",
			zap.Uint("userID", userID),
			zap.Uint("wishlistID", wishlist.ID))
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logger.Log.Error("Invalid product ID",
			zap.String("productID", c.Param("id")),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Product ID", "Product ID Not Found", "")
		return
	}

	var product models.ProductVariantDetails
	if err := tx.First(&product, "id = ?", productID).Error; err != nil {
		logger.Log.Error("Product not found",
			zap.Int("productID", productID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Product Not Found", "")
		return
	}

	var wishlistItem models.WishlistItem
	if err := tx.First(&wishlistItem, "product_id = ? AND product_variant_id = ? AND wishlist_id = ?", product.ProductID, product.ID, wishlist.ID).Error; err != nil {
		wishlistItem = models.WishlistItem{
			WishlistID:       wishlist.ID,
			ProductVariantID: product.ID,
			ProductID:        product.ProductID,
		}
		if createErr := tx.Create(&wishlistItem).Error; createErr != nil {
			logger.Log.Error("Failed to add item to wishlist",
				zap.Uint("wishlistID", wishlist.ID),
				zap.Uint("productID", product.ID),
				zap.Error(createErr))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to add to wishlist", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Item added to wishlist",
			zap.Uint("wishlistID", wishlist.ID),
			zap.Uint("wishlistItemID", wishlistItem.ID),
			zap.Uint("productID", product.ID))
	} else {
		logger.Log.Warn("Item already in wishlist",
			zap.Uint("wishlistID", wishlist.ID),
			zap.Uint("wishlistItemID", wishlistItem.ID),
			zap.Uint("productID", product.ID))
	}

	tx.Commit()
	logger.Log.Info("Wishlist addition completed",
		zap.Uint("userID", userID),
		zap.Uint("productID", uint(productID)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Added to Wishlist Successfully",
		"code":    http.StatusOK,
	})
}

func RemoveFromWishlist(c *gin.Context) {
	logger.Log.Info("Removing from wishlist")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var wishlist models.Wishlist
	if err := config.DB.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Wishlist not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Wishlist Not Found", "Something Went Wrong", "")
		return
	}

	itemID := c.Param("id")

	var wishlistItem models.WishlistItem
	if err := config.DB.First(&wishlistItem, "product_variant_id = ? AND wishlist_id = ?", itemID, wishlist.ID).Error; err != nil {
		logger.Log.Warn("Wishlist item not found",
			zap.String("itemID", itemID),
			zap.Uint("wishlistID", wishlist.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Something Went Wrong", "")
		return
	}

	if err := config.DB.Unscoped().
		Where("wishlist_id = ? AND product_variant_id = ? AND product_id = ?",
			wishlist.ID, wishlistItem.ProductVariantID, wishlistItem.ProductID).
		Delete(&models.WishlistItem{}).Error; err != nil {
		logger.Log.Error("Failed to delete wishlist item",
			zap.Uint("wishlistID", wishlist.ID),
			zap.Uint("productVariantID", wishlistItem.ProductVariantID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
		return
	}

	logger.Log.Info("Item removed from wishlist",
		zap.Uint("userID", userID),
		zap.Uint("wishlistID", wishlist.ID),
		zap.String("itemID", itemID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Item Removed Successfully",
		"code":    http.StatusOK,
	})
}

func WishlistTOCart(c *gin.Context) {
	logger.Log.Info("Moving wishlist item to cart")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	tx := config.DB.Begin()

	var wishlist models.Wishlist
	if err := tx.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Wishlist not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Wishlist Not Found", "Something Went Wrong", "")
		return
	}

	itemID := c.Param("id")

	var wishlistItem models.WishlistItem
	if err := tx.First(&wishlistItem, "wishlist_id = ? AND id = ?", wishlist.ID, itemID).Error; err != nil {
		logger.Log.Warn("Wishlist item not found",
			zap.String("itemID", itemID),
			zap.Uint("wishlistID", wishlist.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Something Went Wrong", "")
		return
	}

	var cart models.Cart
	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart Not Found", "Something Went Wrong", "")
		return
	}

	var product models.ProductVariantDetails
	if err := tx.First(&product, "id = ? AND product_id = ?", wishlistItem.ProductVariantID, wishlistItem.ProductID).Error; err != nil {
		logger.Log.Error("Product not found",
			zap.Uint("productVariantID", wishlistItem.ProductVariantID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Product Not Found", "")
		return
	}

	if product.StockQuantity == 0 {
		logger.Log.Warn("Product out of stock",
			zap.Uint("productID", product.ID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusConflict, "Product Out Of Stock", "Product Out Of Stock", "")
		return
	}

	var cartItem models.CartItem
	if err := tx.First(&cartItem, "product_id = ? AND product_variant_id = ? AND cart_id = ?", product.ProductID, product.ID, cart.ID).Error; err != nil {
		cartItem = models.CartItem{
			CartID:           cart.ID,
			ProductID:        product.ProductID,
			ProductVariantID: product.ID,
			Quantity:         1,
		}
		if createErr := tx.Create(&cartItem).Error; createErr != nil {
			logger.Log.Error("Failed to add item to cart",
				zap.Uint("cartID", cart.ID),
				zap.Uint("productID", product.ID),
				zap.Error(createErr))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
			return
		}

		if err := config.DB.Unscoped().Delete(&wishlistItem).Error; err != nil {
			logger.Log.Error("Failed to delete wishlist item after moving to cart",
				zap.Uint("wishlistItemID", wishlistItem.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
			return
		}
	} else {
		logger.Log.Warn("Product already in cart",
			zap.Uint("cartID", cart.ID),
			zap.Uint("productID", product.ID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Product Already In Your Cart", "Product Already In Your Cart", "")
		return
	}

	tx.Commit()
	logger.Log.Info("Item moved from wishlist to cart",
		zap.Uint("userID", userID),
		zap.Uint("wishlistID", wishlist.ID),
		zap.Uint("cartID", cart.ID),
		zap.Uint("productID", product.ID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Product Moved TO Cart Successfully",
		"code":    http.StatusOK,
	})
}

func WishlistAllTOCart(c *gin.Context) {
	logger.Log.Info("Moving all wishlist items to cart")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	tx := config.DB.Begin()

	var wishlist models.Wishlist
	if err := tx.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Wishlist not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Wishlist Not Found", "Something Went Wrong", "")
		return
	}

	var wishlistItems []models.WishlistItem
	if err := tx.Find(&wishlistItems, "wishlist_id = ?", wishlist.ID).Error; err != nil {
		logger.Log.Warn("No wishlist items found",
			zap.Uint("wishlistID", wishlist.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request", "")
		return
	}

	var cart models.Cart
	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart Not Found", "Something Went Wrong", "")
		return
	}

	movedCount := 0
	for _, item := range wishlistItems {
		var product models.ProductVariantDetails
		if err := tx.First(&product, item.ProductVariantID).Error; err != nil {
			logger.Log.Warn("Product not found during bulk move",
				zap.Uint("productVariantID", item.ProductVariantID),
				zap.Error(err))
			continue
		}

		if product.StockQuantity == 0 {
			logger.Log.Warn("Product out of stock during bulk move",
				zap.Uint("productID", product.ID))
			continue
		}

		var cartItem models.CartItem
		if err := tx.First(&cartItem, "product_id = ? AND product_variant_id = ? AND cart_id = ?", product.ProductID, product.ID, cart.ID).Error; err != nil {
			cartItem = models.CartItem{
				CartID:           cart.ID,
				ProductID:        product.ProductID,
				ProductVariantID: product.ID,
				Quantity:         1,
			}
			if createErr := tx.Create(&cartItem).Error; createErr != nil {
				logger.Log.Error("Failed to add item to cart during bulk move",
					zap.Uint("cartID", cart.ID),
					zap.Uint("productID", product.ID),
					zap.Error(createErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
				return
			}

			if err := config.DB.Unscoped().Delete(&item).Error; err != nil {
				logger.Log.Error("Failed to delete wishlist item after bulk move",
					zap.Uint("wishlistItemID", item.ID),
					zap.Error(err))
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
				return
			}
			movedCount++
		}
	}

	tx.Commit()
	logger.Log.Info("All wishlist items moved to cart",
		zap.Uint("userID", userID),
		zap.Uint("wishlistID", wishlist.ID),
		zap.Uint("cartID", cart.ID),
		zap.Int("movedCount", movedCount))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "All Products Moved TO Cart Successfully",
		"code":    http.StatusOK,
	})
}
