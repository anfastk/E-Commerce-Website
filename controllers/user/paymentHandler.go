package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func PaymentPage(c *gin.Context) {
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

	var request struct {
		AddressID            string  `json:"addressId"`
		CouponCode           string  `json:"couponCode"`
		CouponId             int     `json:"couponId"`
		CouponDiscountAmount float64 `json:"couponDiscountAmount"`
	}

	tx := config.DB.Begin()
	var expiredReservations []models.ReservedStock
	err := tx.Where("is_confirmed = ? AND reserve_till >= ?", false, time.Now()).
		Find(&expiredReservations).Error
	if err != nil {
		tx.Rollback()
		log.Println("Error finding expired reservations:", err)
		return
	}
	for _, reservation := range expiredReservations {
		var coupon models.ReservedCoupon
		tx.First(&coupon, reservation.ReservedCouponID)
		tx.Exec("UPDATE coupons SET users_used_count = users_used_count + ? WHERE id = ?", 1, coupon.CouponID)
		tx.Unscoped().Delete(&coupon)
		if err := tx.Exec(
			"UPDATE product_variant_details SET stock_quantity = stock_quantity + ? WHERE id = ?",
			reservation.Quantity, reservation.ProductVariantID,
		).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed releasing stock", "Something Went Wrong", "")
			tx.Rollback()
			return
		}
		if err := tx.Unscoped().Delete(&reservation).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete  released stock", "Something Went Wrong", "")
			tx.Rollback()
			return
		}
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Binding the data", "")
		return
	}

	var address models.UserAddress
	if err := tx.First(&address, "id = ? AND user_id = ?", request.AddressID, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Address not found", "Address not found", "")
	}

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
	var couponId uint
	if request.CouponId != 0 {
		var coupon models.Coupon
		if productErr := tx.First(&coupon, request.CouponId).Error; productErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Coupon not found", "Something Went Wrong", "/checkout")
			return
		}
		coupon.UsersUsedCount += 1
		if err := tx.Save(&coupon).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusConflict, "Failed to Reserve coupon", "Something Went Wrong", "/checkout")
			return
		}

		reserveCoupon := models.ReservedCoupon{
			CouponCode:           request.CouponCode,
			Discription:          coupon.Discription,
			CouponDiscountAmount: request.CouponDiscountAmount,
			CouponID:             uint(request.CouponId),
		}
		if err := tx.Create(&reserveCoupon).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to coupon reserved stock", "Something Went Wrong", "/checkout")
			return
		}
		couponId = reserveCoupon.ID
	}

	for _, item := range cartItems {
		if item.ProductVariant.StockQuantity < item.Quantity {
			helper.RespondWithError(c, http.StatusConflict, "Stock unavailable", "One or more items in your cart are out of stock. Please update your cart.", "/checkout")
			return
		}
		var product models.ProductVariantDetails
		if productErr := tx.First(&product, item.ProductID).Error; productErr != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Product not found", "Something Went Wrong", "/checkout")
			return
		}
		product.StockQuantity -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusConflict, "Failed to Reserve stock", "Something Went Wrong", "/checkout")
			return
		}
		reserveStock := models.ReservedStock{
			UserID:           userID,
			ProductVariantID: item.ProductID,
			Quantity:         item.Quantity,
			ReservedAt:       time.Now(),
			ReserveTill:      time.Now().Add(15 * time.Minute),
			IsConfirmed:      false,
			ReservedCouponID: couponId,
		}
		if err := tx.Create(&reserveStock).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to store reserved stock", "Something Went Wrong", "/checkout")
			return
		}
	}
	var CartItems []CartItemWithDiscount
	for _, items := range cartItems {
		discountAmount, _, _ := helper.DiscountCalculation(items.ProductID, items.ProductVariant.CategoryID, items.ProductVariant.RegularPrice, items.ProductVariant.SalePrice)
		CartItems = append(CartItems, CartItemWithDiscount{
			Item:          items,
			DiscountPrice: items.ProductVariant.SalePrice - discountAmount,
		})
	}
	regularPrice, salePrice, tax, productDiscount, totalDiscount, shippingCharge := services.CalculateCartPrices(cartItems)

	TotalDiscount := totalDiscount + request.CouponDiscountAmount
	total := (salePrice + tax) - request.CouponDiscountAmount
	IsCodAvailable := true

	for _, itm := range cartItems {
		var productDetail models.ProductDetail
		if err := config.DB.First(&productDetail, itm.ProductID).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Prduct Not Found", "Something Went Wrong", "/checkout")
			return
		}
		if !productDetail.IsCODAvailable {
			IsCodAvailable = false
			break
		}
	}

	if total > 75000 {
		IsCodAvailable = false
	}

	tx.Commit()
	CheckForReferrer(c)
	CheckForJoinee(c)

	c.HTML(http.StatusOK, "paymentPage.html", gin.H{
		"status":          "OK",
		"message":         "Checkout fetch success",
		"Address":         address,
		"CartItem":        CartItems,
		"SubTotal":        regularPrice,
		"Shipping":        shippingCharge,
		"Tax":             tax,
		"CouponID":        request.CouponId,
		"CouponCode":      request.CouponCode,
		"CouponDiscount":  request.CouponDiscountAmount,
		"ProductDiscount": productDiscount,
		"TotalDiscount":   TotalDiscount,
		"IsCodAvailable":  IsCodAvailable,
		"Total":           total,
		"code":            http.StatusOK,
	})
}

var paymentRequest struct {
	PaymentMethod        string `json:"paymentMethod"`
	AddressID            string `json:"addressId"`
	CouponCode           string `json:"couponCode"`
	CouponId             string `json:"couponId"`
	CouponDiscountAmount string `json:"couponDiscountAmount"`
}

func ProceedToPayment(c *gin.Context) {
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

	if err := c.ShouldBind(&paymentRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "")
		return
	}

	var userDetails models.UserAuth

	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

	currentTime := time.Now()

	cart := FetchCartByUserID(c, userDetails.ID)

	cartItems := FetchCartItemByCartID(c, cart.ID)

	if len(cartItems) == 0 {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/cart")
		return
	}
	reservedProducts := FetchReservedProducts(c, userID)

	if len(reservedProducts) != len(cartItems) {
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 11", "Something Went Wrong", "/cart")
		return
	}
	reservedMap, subtotal, totalProductDiscount, totalDiscount, shippingCharge, tax, total := ReservedProductCheck(c, reservedProducts, cartItems)

	var coupon models.ReservedCoupon
	config.DB.First(&coupon, paymentRequest.CouponId)
	couponDiscountAmount, err := strconv.ParseFloat(paymentRequest.CouponDiscountAmount, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Coverting Failed", "Something Went Wrong", "/cart")
		return
	}

	switch paymentRequest.PaymentMethod {
	case "COD":
		paymentStatus := true
		tx := config.DB.Begin()
		orderID := CreateOrder(c, tx, userDetails.ID, subtotal, totalProductDiscount, totalDiscount+couponDiscountAmount, tax, float64(shippingCharge), total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
		SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
		CreateOrderItems(c, tx, reservedProducts, float64(shippingCharge), orderID, userDetails.ID, currentTime, couponDiscountAmount)
		orderItems := FetchOrderItems(c, tx, orderID)

		for _, ordItems := range orderItems {
			if err := CODPayment(c, tx, userDetails.ID, ordItems.OrderUID, ordItems.ID, ordItems.Total, "Cash On Delivery"); err != nil {
				tx.Rollback()
				paymentStatus = false
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment failed", "Something Went Wrong", "/checkout")
				return
			}
			DeleteReservedItems(c, tx, ordItems.ProductVariantID, userID)
		}

		if paymentStatus {
			ClearCart(c, tx, reservedMap)
		}
		tx.Unscoped().Delete(&coupon, paymentRequest.CouponId)
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"message":  "Payment processed",
			"redirect": "/order/success",
			"code":     http.StatusOK,
		})

	case "Razorpay":
		address := FetchAddressByIDAndUserID(c, userID, paymentRequest.AddressID)
		razorpayOrder, err := CreateRazorpayOrder(c, total-couponDiscountAmount)
		if err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create Razorpay order", "Something Went Wrong", "/checkout")
			return
		}

		razorpayOrderID, ok := razorpayOrder["id"].(string)
		if !ok {
			log.Println("Failed to extract order ID from Razorpay response:", razorpayOrder)
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Razorpay response", "Something Went Wrong", "/checkout")
			return
		}
		RazorPayOrderID = razorpayOrderID

		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"order_id": razorpayOrderID,
			"amount":   total - couponDiscountAmount*100,
			"currency": "INR",
			"key_id":   config.RAZORPAY_KEY_ID,
			"prefill": gin.H{
				"name":    userDetails.FullName,
				"email":   userDetails.Email,
				"contact": address.Mobile,
			},
			"notes": gin.H{
				"address": address.Address,
				"user_id": userDetails.ID,
			},
		})
	case "Wallet":
		tx := config.DB.Begin()
		var walletDetails models.Wallet
		if err := tx.First(&walletDetails, "user_id = ?", userID).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Not Found", "Something Went Wrong", "/cart")
			return
		}
		orderID := CreateOrder(c, tx, userDetails.ID, subtotal, totalProductDiscount, totalDiscount+couponDiscountAmount, tax, float64(shippingCharge), total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
		SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
		CreateOrderItems(c, tx, reservedProducts, float64(shippingCharge), orderID, userDetails.ID, currentTime, couponDiscountAmount)
		orderItems := FetchOrderItems(c, tx, orderID)

		var orderDetails models.Order
		if err := tx.First(&orderDetails, orderID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "")
			return
		}

		for _, orderItem := range orderItems {

			receiptID := "rcpt-" + uuid.New().String()
			transactionID := "TXN-" + uuid.New().String()

			createPayment := models.PaymentDetail{
				UserID:        userID,
				OrderItemID:   orderItem.ID,
				PaymentStatus: "Completed",
				PaymentAmount: orderItem.Total,
				PaymentMethod: "Wallet",
				OrderId:       orderDetails.OrderUID,
				TransactionID: transactionID,
				Receipt:       receiptID,
			}

			if err := tx.Create(&createPayment).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment creation failed", "Something Went Wrong", "/cart")
				return
			}

			var orderedItem models.OrderItem
			if err := tx.First(&orderedItem, orderItem.ID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "")
				return
			}

			orderedItem.OrderStatus = "Confirmed"
			if err := tx.Save(&orderedItem).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Something Went Wrong", "")
				return
			}
			DeleteReservedItems(c, tx, orderItem.ProductVariantID, userID)
		}

		lastBalance := walletDetails.Balance
		walletDetails.Balance -= total
		if err := tx.Save(&walletDetails).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update wallet details", "Failed to update order", "")
			return
		}
		walletReceiptID := "rcpt-" + uuid.New().String()
		walletTransactionID := "TXN-" + uuid.New().String()
		walletHistory := models.WalletTransaction{
			UserID:        userID,
			WalletID:      walletDetails.ID,
			Amount:        total,
			LastBalance:   lastBalance,
			Description:   "Product Purchase ORD ID" + orderDetails.OrderUID,
			Type:          "Debited",
			Receipt:       walletReceiptID,
			OrderId:       orderDetails.OrderUID,
			TransactionID: walletTransactionID,
			PaymentMethod: "Wallet",
		}
		if err := tx.Create(&walletHistory).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Creation Failed", "Something Went Wrong", "/cart")
			return
		}

		ClearCart(c, tx, reservedMap)
		tx.Unscoped().Delete(&coupon, paymentRequest.CouponId)
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"message":  "Payment processed",
			"redirect": "/order/success",
			"code":     http.StatusOK,
		})

	default:
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/checkout")
		return
	}
}

func PayNow(c *gin.Context) {
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

	var PayNowRequest struct {
		Method      string `json:"method"`
		OrderItemID string `json:"orderId"`
		AddressID   string `json:"addressId"`
	}

	if err := c.ShouldBind(&PayNowRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "/profile/order/details")
		return
	}

	var userDetails models.UserAuth

	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

	if PayNowRequest.Method == "" || PayNowRequest.OrderItemID == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "/profile/order/details")
		return
	}
	switch PayNowRequest.Method {
	case "Razorpay":
		var orderItems models.OrderItem
		if err := config.DB.First(&orderItems, PayNowRequest.OrderItemID).Error; err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Order Item Not Fount", "Something Went Wrong", "/profile/order/details")
			return
		}
		var order models.Order
		if err := config.DB.First(&order, orderItems.OrderID).Error; err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Order Not Fount", "Something Went Wrong", "/profile/order/details")
			return
		}
		addressId, adErr := strconv.Atoi(PayNowRequest.AddressID)
		if adErr != nil {
			helper.RespondWithError(c, http.StatusBadRequest, "Invalid Address", "Something Went Wrong", "/profile/order/details")
			return
		}
		var address models.ShippingAddress
		if err := config.DB.First(&address, "id = ?", addressId).Error; err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Address Not Found", "Something Went Wrong", "/profile/order/details")
			return
		}

		razorpayOrder, err := CreateRazorpayOrder(c, order.TotalAmount)
		if err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create Razorpay order", "Something Went Wrong", "/profile/order/details")
			return
		}

		razorpayOrderID, ok := razorpayOrder["id"].(string)
		if !ok {
			log.Println("Failed to extract order ID from Razorpay response:", razorpayOrder)
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Razorpay response", "Something Went Wrong", "/checkout")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"order_id": razorpayOrderID,
			"amount":   order.TotalAmount * 100,
			"currency": "INR",
			"key_id":   config.RAZORPAY_KEY_ID,
			"prefill": gin.H{
				"name":    userDetails.FullName,
				"email":   userDetails.Email,
				"contact": address.Mobile,
			},
			"notes": gin.H{
				"address": address.Address,
				"user_id": userDetails.ID,
			},
		})

	default:
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/profile/order/details")
		return
	}

}
