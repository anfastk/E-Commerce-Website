package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
) 

func ShowCheckoutPage(c *gin.Context) {
	logger.Log.Info("Requested checkout page")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	shippingCharge := 100

	_, cartItems, err := services.FetchCartItems(userID)
	if err != nil {
		logger.Log.Error("Failed to fetch cart items", zap.Uint("userID", userID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Cart Error", err.Error(), "/cart")
		return
	}

	if len(cartItems) == 0 {
		logger.Log.Warn("Cart is empty", zap.Uint("userID", userID))
		helper.RespondWithError(c, http.StatusNotFound, "Cart is empty", "Cart is empty", "/cart")
		return
	}

	for _, item := range cartItems {
		if item.ProductDetails.StockQuantity < item.CartItem.Quantity || item.ProductDetails.StockQuantity == 0 || item.CartItem.Quantity == 0 {
			logger.Log.Warn("Product Out Of Stock",
				zap.Uint("productVariantID", item.ProductDetails.ID),
				zap.Int("stockQuantity", int(item.ProductDetails.StockQuantity)),
				zap.Int("cartQuantity", int(item.CartItem.Quantity)))
			helper.RespondWithError(c, http.StatusConflict, "Out of stock", "Some items in your cart are out of stock or have lower availability than your selected quantity. Please update your cart.", "/cart")
			return
		}
		if item.ProductDetails.IsDeleted {
			logger.Log.Warn("Product Unavailable",
				zap.Uint("productVariantID", item.ProductDetails.ID),
				zap.Int("stockQuantity", int(item.ProductDetails.StockQuantity)),
				zap.Int("cartQuantity", int(item.CartItem.Quantity)))
			helper.RespondWithError(c, http.StatusConflict, "Product unavailable", "Some items in your cart are unavailable. Please update your cart.", "/cart")
			return
		}
	}

	var categoryIdForOffer uint
	for _, items := range cartItems {
		categoryIdForOffer = items.ProductDetails.CategoryID
	}

	regularPrice, salePrice, tax, productDiscount, totalDiscount, shippingCharge := services.CalculateCartPrices(cartItems)
	total := salePrice + tax

	isAllCategorySame := true
	for _, items := range cartItems {
		if categoryIdForOffer != items.ProductDetails.CategoryID {
			isAllCategorySame = false
			break
		}
	}

	var allResponceCoupons []models.Coupon
	if isAllCategorySame {
		var category models.Categories
		if err := config.DB.First(&category, categoryIdForOffer).Error; err != nil {
			logger.Log.Error("Failed to fetch category",
				zap.Uint("categoryID", categoryIdForOffer),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Category not found", "Something Went Wrong", "")
			return
		}

		var categoryCoupons []models.Coupon
		if err := config.DB.Where(
			"min_order_value <= ? AND users_used_count < max_use_count AND applicable_for = ? AND expiration_date >= ? AND status = ?",
			salePrice, category.Name, time.Now(), "Active",
		).Find(&categoryCoupons).Error; err != nil {
			logger.Log.Error("Failed to fetch category coupons",
				zap.Float64("salePrice", salePrice),
				zap.String("category", category.Name),
				zap.Error(err))
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
		logger.Log.Error("Failed to fetch all products coupons",
			zap.Float64("salePrice", salePrice),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Error fetching coupons", "Something Went Wrong", "")
		return
	}

	for _, coupon := range allProductCoupons {
		allResponceCoupons = append(allResponceCoupons, coupon)
	}

	tx := config.DB.Begin()
	var expiredReservations []models.ReservedStock
	if err := tx.Where("is_confirmed = ? AND reserve_till >= ?", false, time.Now()).
		Find(&expiredReservations).Error; err != nil {
		logger.Log.Error("Failed to find expired reservations", zap.Error(err))
		tx.Rollback()
		return
	}

	for _, reservation := range expiredReservations {
		var coupon models.ReservedCoupon
		if err := tx.First(&coupon, reservation.ReservedCouponID).Error; err != nil {
			logger.Log.Warn("Failed to fetch reserved coupon",
				zap.Uint("reservedCouponID", reservation.ReservedCouponID),
				zap.Error(err))
		}
		if err := tx.Exec("UPDATE coupons SET users_used_count = users_used_count + ? WHERE id = ?", 1, coupon.CouponID).Error; err != nil {
			logger.Log.Error("Failed to update coupon users count",
				zap.Uint("couponID", coupon.CouponID),
				zap.Error(err))
		}
		if err := tx.Unscoped().Delete(&coupon).Error; err != nil {
			logger.Log.Warn("Failed to delete reserved coupon",
				zap.Uint("reservedCouponID", reservation.ReservedCouponID),
				zap.Error(err))
		}
		if err := tx.Exec(
			"UPDATE product_variant_details SET stock_quantity = stock_quantity + ? WHERE id = ?",
			reservation.Quantity, reservation.ProductVariantID,
		).Error; err != nil {
			logger.Log.Error("Failed to release stock",
				zap.Uint("productVariantID", reservation.ProductVariantID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed releasing stock", "Something Went Wrong", "")
			tx.Rollback()
			return
		}
		if err := tx.Unscoped().Delete(&reservation).Error; err != nil {
			logger.Log.Error("Failed to delete released stock",
				zap.Uint("reservationID", reservation.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete released stock", "Something Went Wrong", "")
			tx.Rollback()
			return
		}
		logger.Log.Info("Expired reservation cleaned up",
			zap.Uint("reservationID", reservation.ID))
	}

	for _, item := range cartItems {
		if item.ProductDetails.StockQuantity < item.CartItem.Quantity {
			logger.Log.Warn("Stock unavailable",
				zap.Uint("productID", item.CartItem.ProductID),
				zap.Int("stockQty", int(item.CartItem.ProductVariant.StockQuantity)),
				zap.Int("requestedQty", int(item.CartItem.Quantity)))
			helper.RespondWithError(c, http.StatusConflict, "Stock unavailable", "One or more items in your cart are out of stock. Please update your cart.", "/checkout")
			return
		}
		var product models.ProductVariantDetails
		if err := tx.First(&product, item.CartItem.ProductID).Error; err != nil {
			logger.Log.Error("Product not found",
				zap.Uint("productID", item.CartItem.ProductID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Product not found", "Something Went Wrong", "/checkout")
			return
		}
		product.StockQuantity -= item.CartItem.Quantity
		if err := tx.Save(&product).Error; err != nil {
			logger.Log.Error("Failed to reserve stock",
				zap.Uint("productID", item.CartItem.ProductID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusConflict, "Failed to Reserve stock", "Something Went Wrong", "/checkout")
			return
		}
		reserveStock := models.ReservedStock{
			UserID:           userID,
			ProductVariantID: item.CartItem.ProductID,
			Quantity:         item.CartItem.Quantity,
			ReservedAt:       time.Now(),
			ReserveTill:      time.Now().Add(15 * time.Minute),
			IsConfirmed:      false,
		}

		if err := tx.Create(&reserveStock).Error; err != nil {
			logger.Log.Error("Failed to store reserved stock",
				zap.Uint("productID", item.CartItem.ProductID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to store reserved stock", "Something Went Wrong", "/checkout")
			return
		}
		logger.Log.Info("Stock reserved",
			zap.Uint("productID", item.CartItem.ProductID),
			zap.Int("quantity", int(item.CartItem.Quantity)))
	}

	tx.Commit()

	helper.CreateWallet(c, userID)
	logger.Log.Info("Checkout page loaded successfully",
		zap.Uint("userID", userID),
		zap.Int("cartItemCount", len(cartItems)),
		zap.Int("couponCount", len(allResponceCoupons)),
		zap.Float64("total", total))
	c.HTML(http.StatusOK, "checkOut.html", gin.H{
		"status":          "OK",
		"message":         "Checkout fetch success",
		"CartItem":        cartItems,
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

func ShippingAddress(c *gin.Context) {
	logger.Log.Info("Requested to address")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var address []models.UserAddress
	if err := config.DB.Order("updated_at DESC").Find(&address, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Failed to fetch user addresses", zap.Uint("userID", userID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "Success",
		"Addresses": address,
		"code":      http.StatusOK,
	})
}

func CheckCoupon(c *gin.Context) {
	logger.Log.Info("Requested to check coupon")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var couponInput struct {
		CouponCode      string  `json:"couponCode"`
		SubTotal        float64 `json:"subTotal"`
		ProductDiscount float64 `json:"productDiscount"`
	}
	if err := c.ShouldBindJSON(&couponInput); err != nil {
		logger.Log.Error("Failed to bind coupon input", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Invalid data entered", "")
		return
	}

	couponCode := strings.TrimSpace(strings.ToUpper(couponInput.CouponCode))
	var coupon models.Coupon
	if err := config.DB.First(&coupon, "UPPER(coupon_code) = ?", couponCode).Error; err != nil {
		logger.Log.Warn("Invalid coupon code", zap.String("couponCode", couponCode), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Coupon Code", "Invalid Coupon Code", "")
		return
	}

	if coupon.Status == "Deleted" {
		logger.Log.Warn("Coupon is deleted", zap.String("couponCode", couponCode))
		helper.RespondWithError(c, http.StatusNotFound, "Coupon Not Found", "Coupon Not Found", "")
		return
	}
	if coupon.Status == "Expired" || coupon.UsersUsedCount >= coupon.MaxUseCount {
		logger.Log.Warn("Coupon is expired or max uses reached",
			zap.String("couponCode", couponCode),
			zap.String("status", coupon.Status),
			zap.Int("usersUsedCount", int(coupon.UsersUsedCount)),
			zap.Int("maxUseCount", int(coupon.MaxUseCount)))
		helper.RespondWithError(c, http.StatusNotFound, "Coupon Expired", "Coupon Expired", "")
		return
	}

	if coupon.ExpirationDate.Before(time.Now().Truncate(24 * time.Hour)) {
		logger.Log.Warn("Coupon has expired",
			zap.String("couponCode", couponCode),
			zap.Time("expirationDate", coupon.ExpirationDate))
		helper.RespondWithError(c, http.StatusBadRequest, "Coupon Expired", "Coupon Expired", "")
		return
	}

	if time.Now().Before(coupon.ValidFrom) {
		logger.Log.Warn("Coupon not yet valid",
			zap.String("couponCode", couponCode),
			zap.Time("validFrom", coupon.ValidFrom))
		helper.RespondWithError(c, http.StatusBadRequest, "Coupon Not Started", "Coupon Not Started", "")
		return
	}

	if coupon.ApplicableFor != "AllProducts" {
		_, cartItems, err := services.FetchCartItems(userID)
		if err != nil {
			logger.Log.Error("Failed to fetch cart items for coupon check",
				zap.Uint("userID", userID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Cart Error", err.Error(), "/cart")
			return
		}

		if len(cartItems) == 0 {
			logger.Log.Warn("Cart is empty for coupon check", zap.Uint("userID", userID))
			helper.RespondWithError(c, http.StatusBadRequest, "Cart Empty", "Add products to apply coupon", "/cart")
			return
		}

		categoryIdForOffer := cartItems[0].ProductDetails.CategoryID
		isAllCategorySame := true
		for _, items := range cartItems {
			if categoryIdForOffer != items.ProductDetails.CategoryID {
				isAllCategorySame = false
				break
			}
		}
		if !isAllCategorySame {
			logger.Log.Warn("Coupon not applicable due to mixed categories",
				zap.String("couponCode", couponCode),
				zap.String("applicableFor", coupon.ApplicableFor))
			helper.RespondWithError(c, http.StatusBadRequest, "Coupon Not Applicable", "Coupon Not Applicable", "/cart")
			return
		}
	}

	purchaseAmount := couponInput.SubTotal - couponInput.ProductDiscount
	if purchaseAmount < coupon.MinOrderValue {
		logger.Log.Warn("Purchase amount below minimum for coupon",
			zap.String("couponCode", couponCode),
			zap.Float64("purchaseAmount", purchaseAmount),
			zap.Float64("minOrderValue", coupon.MinOrderValue))
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

	logger.Log.Info("Coupon applied successfully",
		zap.String("couponCode", couponCode),
		zap.Uint("couponID", coupon.ID),
		zap.Float64("discountAmount", Discount))
	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"message":        "Coupon Applied",
		"CouponID":       coupon.ID,
		"description":    coupon.Discription,
		"discountAmount": Discount,
		"code":           http.StatusOK,
	})
}
