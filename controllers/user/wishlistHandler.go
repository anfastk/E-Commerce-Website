package controllers

import (
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ShowWishlist(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	type responce struct {
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create wishlist", "Something Went Wrong", "")
			return
		}
	}

	var wishlistItem []models.WishlistItem
	if err := config.DB.Preload("ProductVariantDetails").
		Preload("ProductVariantDetails.VariantsImages").
		Find(&wishlistItem, "wishlist_id = ?", wishlist.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
		return
	}
	var wishlistResponce []responce
	for _, item := range wishlistItem {
		var cart models.Cart
		if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Cart Not Found", "Something Went Wrong", "")
			return
		}
		isInCart := true
		var cartItems models.CartItem
		if err := config.DB.First(&cartItems, "cart_id = ? AND product_id = ? AND product_variant_id = ?", cart.ID, item.ProductID, item.ProductVariantID).Error; err != nil {
			isInCart = false
		}
		discountAmount, discountPercentage, _ := helper.DiscountCalculation(item.ProductVariantID, item.ProductVariantDetails.CategoryID, item.ProductVariantDetails.RegularPrice, item.ProductVariantDetails.SalePrice)
		items := responce{
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
		wishlistResponce = append(wishlistResponce, items)
	}
	count:=len(wishlistResponce)
	c.HTML(http.StatusOK, "wishlist.html", gin.H{
		"status":  "Success",
		"Data":    wishlistResponce,
		"Count":count,
		"message": "Added to Wishlist Successfully",
		"code":    http.StatusOK,
	})
}

func AddToWishlist(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	tx := config.DB.Begin()

	var wishlist models.Wishlist
	if err := tx.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		wishlist = models.Wishlist{UserID: userID}
		if createErr := tx.Create(&wishlist).Error; createErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create wishlist", "Something Went Wrong", "")
			return
		}
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Product ID", "Product ID Not Found", "")
		return
	}

	var product models.ProductVariantDetails
	if err := tx.First(&product, "id = ?", productID).Error; err != nil {
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
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to add to wishlist", "Something Went Wrong", "")
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Added to Wishlist Successfully",
		"code":    http.StatusOK,
	})
}

func RemoveFromWishlist(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	var wishlist models.Wishlist
	if err := config.DB.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Wishlist Not Found", "Something Went Wrong", "")
		return
	}

	itemID := c.Param("id")

	var wishlistItem models.WishlistItem
	if err := config.DB.First(&wishlistItem, "product_variant_id = ? AND wishlist_id = ?", itemID, wishlist.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Inavlid request", "Something Went Wrong", "")
		return
	}

	if err := config.DB.Unscoped().
		Where("wishlist_id = ? AND product_variant_id = ? AND product_id = ?",
			wishlist.ID, wishlistItem.ProductVariantID, wishlistItem.ProductID).
		Delete(&models.WishlistItem{}).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Item Removed Successfully",
		"code":    http.StatusOK,
	})
}

func WishlistTOCart(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	tx := config.DB.Begin()

	var wishlist models.Wishlist
	if err := tx.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Wishlist Not Found", "Something Went Wrong", "")
		return
	}

	itemID := c.Param("id")

	var wishlistItem models.WishlistItem
	if err := tx.First(&wishlistItem, "wishlist_id = ? AND id = ?", wishlist.ID, itemID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Inavlid request", "Something Went Wrong", "")
	}

	var cart models.Cart

	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart Not Found", "Something Went Wrong", "")
		return
	}

	var product models.ProductVariantDetails
	if err := tx.First(&product, "id = ? AND product_id = ?", wishlistItem.ProductVariantID, wishlistItem.ProductID).Error; err != nil {
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
			ProductVariantID: product.ID,
			Quantity:         1,
		}
		if createErr := tx.Create(&cartItems).Error; createErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
			return
		}

		if err := config.DB.Unscoped().Delete(&wishlistItem).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
		}
	} else {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Product Already In Your Cart", "Product Already In Your Cart", "")
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "Product Moved TO Cart Successfully",
		"code":    http.StatusOK,
	})
}

func WishlistAllTOCart(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	tx := config.DB.Begin()
	var wishlist models.Wishlist
	if err := tx.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Wishlist Not Found", "Something Went Wrong", "")
		return
	}

	var wishlistItem []models.WishlistItem
	if err := tx.Find(&wishlistItem, "wishlist_id = ?", wishlist.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Inavlid request", "Inavlid request", "")
	}
	var cart models.Cart

	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart Not Found", "Something Went Wrong", "")
		return
	}

	for _, items := range wishlistItem {
		var product models.ProductVariantDetails
		if err := tx.First(&product, items.ProductVariantID).Error; err != nil {
			break
		}
		if product.StockQuantity == 0 {
			break
		}

		var cartItems models.CartItem
		if err := tx.First(&cartItems, "product_id = ? AND product_variant_id = ? AND cart_id = ?", product.ProductID, product.ID, cart.ID).Error; err != nil {
			cartItems = models.CartItem{
				CartID:           cart.ID,
				ProductID:        product.ProductID,
				ProductVariantID: product.ID,
				Quantity:         1,
			}
			if createErr := tx.Create(&cartItems).Error; createErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Add to Cart Failed", "Add to Cart Failed", "")
				return
			}

			if err := config.DB.Unscoped().Delete(&items).Error; err != nil {
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete item", "Failed to delete item", "")
			}
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "All Products Moved TO Cart Successfully",
		"code":    http.StatusOK,
	})
}
