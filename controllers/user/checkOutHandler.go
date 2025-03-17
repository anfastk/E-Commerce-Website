package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ShowCheckoutPage(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	type CartItemWithDiscount struct {
		Item          models.CartItem
		DiscountPrice float64
	}
	shippingCharge := 100
	_, cartItems, err := services.FetchCartItems(userID)
	if err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Cart Error", err.Error(), "/cart")
		return
	}

	for _, item := range cartItems {
		if item.ProductVariant.StockQuantity < item.Quantity || item.ProductVariant.StockQuantity == 0 || item.Quantity == 0 {
			helper.RespondWithError(c, http.StatusConflict, "Stock unavailable", "One or more items in your cart are out of stock. Please update your cart.", "/cart")
			return
		}
	}

	var categoryIdForOffer uint

	var CartItems []CartItemWithDiscount
	for _, items := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(items.ProductID, items.ProductVariant.CategoryID, items.ProductVariant.RegularPrice, items.ProductVariant.SalePrice)
		CartItems = append(CartItems, CartItemWithDiscount{
			Item:          items,
			DiscountPrice: items.ProductVariant.SalePrice - discountAmount,
		})
		categoryIdForOffer = items.ProductVariant.CategoryID
	}

	regularPrice, salePrice, tax, productDiscount, totalDiscount, shippingCharge := services.CalculateCartPrices(cartItems)

	total := salePrice + tax

	var address []models.UserAddress
	if err := config.DB.Order("updated_at DESC").Find(&address, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}
	isAllCategorySame := true
	for _, items := range cartItems {
		if categoryIdForOffer != items.ProductVariant.CategoryID {
			isAllCategorySame = false
			break
		}
	}
	var allResponceCoupons []models.Coupon
	if isAllCategorySame {
		var category models.Categories
		if err := config.DB.First(&category, categoryIdForOffer).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Category not found", "Something Went Wrong", "")
			return
		}

		var categoryCoupons []models.Coupon
		if err := config.DB.Where(
			"min_order_value <= ? AND users_used_count < max_use_count AND applicable_for = ? AND expiration_date >= ? AND status = ?",
			salePrice, category.Name, time.Now(), "Active",
		).Find(&categoryCoupons).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Error fetching coupons", "Something Went Wrong", "")
			return
		}
		for _, coupon := range categoryCoupons {
			allResponceCoupons = append(allResponceCoupons, coupon)
		}
	}

	var allProductCoupons []models.Coupon
	if err := config.DB.Where(
		"min_order_value <= ? AND users_used_count < max_use_count AND applicable_for = ? AND expiration_date >= ? AND status = ?",
		salePrice, "AllProducts", time.Now(), "Active",
	).Find(&allProductCoupons).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Error fetching coupons", "Something Went Wrong", "")
		return
	}

	for _, coupon := range allProductCoupons {
		allResponceCoupons = append(allResponceCoupons, coupon)
	}

	CreateWallet(c, userID)
	c.HTML(http.StatusOK, "checkOut.html", gin.H{
		"status":          "OK",
		"message":         "Checkout fetch success",
		"Address":         address,
		"CartItem":        CartItems,
		"SubTotal":        regularPrice,
		"Shipping":        shippingCharge,
		"Tax":             tax,
		"ProductDiscount": productDiscount,
		"TotalDiscount":   totalDiscount,
		"Total":           total,
		"Coupons":         allResponceCoupons,
		"code":            http.StatusOK,
	})
}

func CheckCoupon(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	var couponInput struct {
		CouponCode      string  `json:"couponCode"`
		SubTotal        float64 `json:"subTotal"`
		ProductDiscount float64 `json:"productDiscount"`
	}
	if err := c.ShouldBindJSON(&couponInput); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Invalid data entered", "")
		return
	}

	couponCode := strings.TrimSpace(strings.ToUpper(couponInput.CouponCode))
	var coupon models.Coupon
	if err := config.DB.First(&coupon, "UPPER(coupon_code) = ?", couponCode).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Coupon Code", "Invalid Coupon Code", "")
		return
	}

	if coupon.Status == "Deleted" {
		helper.RespondWithError(c, http.StatusNotFound, "Coupon Not Found", "Coupon Not Found", "")
		return
	}
	if coupon.Status == "Expired" || coupon.UsersUsedCount >= coupon.MaxUseCount {
		helper.RespondWithError(c, http.StatusNotFound, "Coupon Expired", "Coupon Expired", "")
		return
	}

	if coupon.ExpirationDate.Before(time.Now().Truncate(24 * time.Hour)) {
		helper.RespondWithError(c, http.StatusBadRequest, "Coupon Expired", "Coupon Expired", "")
		return
	}

	if time.Now().Before(coupon.ValidFrom) {
		helper.RespondWithError(c, http.StatusBadRequest, "Coupon Not Started", "Coupon Not Started", "")
		return
	}

	if coupon.ApplicableFor != "AllProducts" {
		_, cartItems, err := services.FetchCartItems(userID)
		if err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Cart Error", err.Error(), "/cart")
			return
		}
		if len(cartItems) == 0 {
			helper.RespondWithError(c, http.StatusBadRequest, "Cart Empty", "Add products to apply coupon", "/cart")
			return
		}
		categoryIdForOffer := cartItems[0].ProductVariant.CategoryID
		isAllCategorySame := true
		for _, items := range cartItems {
			if categoryIdForOffer != items.ProductVariant.CategoryID {
				isAllCategorySame = false
				break
			}
		}
		if !isAllCategorySame {
			helper.RespondWithError(c, http.StatusBadRequest, "Coupon Not Applicable", "Coupon Not Applicable", "/cart")
			return
		}
	}

	purchaseAmount := couponInput.SubTotal - couponInput.ProductDiscount
	if purchaseAmount < coupon.MinOrderValue {
		helper.RespondWithError(c, http.StatusBadRequest, "Coupon Not Applicable", "Coupon Not Applicable", "/cart")
		return
	}

	var Discount float64
	if coupon.IsFixedCoupon {
		Discount = coupon.MaxDiscountValue
	} else {
		Discount = purchaseAmount * coupon.DiscountValue / 100
		if Discount > coupon.MaxDiscountValue {
			Discount = coupon.MaxDiscountValue
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"message":        "Coupon Applied",
		"CouponID":       coupon.ID,
		"description":    coupon.Discription,
		"discountAmount": Discount,
		"code":           http.StatusOK,
	})
}
