package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func PaymentPage(c *gin.Context) {
	logger.Log.Info("Requested payment page")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var request struct {
		AddressID            string  `json:"addressId"`
		CouponCode           string  `json:"couponCode"`
		CouponId             int     `json:"couponId"`
		CouponDiscountAmount float64 `json:"couponDiscountAmount"`
	}

	tx := config.DB.Begin()

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Log.Error("Failed to bind request data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Binding the data", "")
		return
	}

	var address models.UserAddress
	if err := tx.First(&address, "id = ? AND user_id = ?", request.AddressID, userID).Error; err != nil {
		logger.Log.Error("Address not found",
			zap.String("addressID", request.AddressID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Address not found", "Address not found", "")
		return
	}

	shippingCharge := 100

	_, cartItems, err := services.FetchCartItems(userID)
	if err != nil {
		logger.Log.Error("Failed to fetch cart items",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Cart Error", err.Error(), "/cart")
		return
	}

	if len(cartItems) == 0 {
		logger.Log.Warn("Cart is empty", zap.Uint("userID", userID))
		helper.RespondWithError(c, http.StatusNotFound, "Cart is empty", "Cart is empty", "/cart")
		return
	}

	reservedProducts := FetchReservedProducts(c, userID)
	if reservedProducts == nil {
		return
	}

	if len(reservedProducts) != len(cartItems) {
		logger.Log.Error("Mismatch between cart items and reserved products",
			zap.Int("cartItemCount", len(cartItems)),
			zap.Int("reservedProductCount", len(reservedProducts)))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product", "Something Went Wrong", "/cart")
		return
	}
	_, err = ReservedProductCheck(c, reservedProducts, cartItems)
	if err != nil {
		return
	}

	var couponId uint
	if request.CouponId != 0 {
		var coupon models.Coupon
		if err := tx.First(&coupon, request.CouponId).Error; err != nil {
			logger.Log.Error("Coupon not found",
				zap.Int("couponID", request.CouponId),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Coupon not found", "Something Went Wrong", "/checkout")
			return
		}
		coupon.UsersUsedCount += 1
		if err := tx.Save(&coupon).Error; err != nil {
			logger.Log.Error("Failed to reserve coupon",
				zap.Int("couponID", request.CouponId),
				zap.Error(err))
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
			logger.Log.Error("Failed to create reserved coupon",
				zap.String("couponCode", request.CouponCode),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to coupon reserved stock", "Something Went Wrong", "/checkout")
			return
		}
		couponId = reserveCoupon.ID
		logger.Log.Info("Coupon reserved",
			zap.Uint("reservedCouponID", couponId),
			zap.String("couponCode", request.CouponCode))
	}

	for _, itm := range reservedProducts {
		itm.ReservedCouponID = couponId

		if err := tx.Save(&itm).Error; err != nil {
			logger.Log.Error("Product detail not found",
				zap.Uint("productID", itm.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save reserved coupon id", "Something Went Wrong", "/checkout")
			return
		}
	}

	regularPrice, salePrice, tax, productDiscount, totalDiscount, shippingCharge := services.CalculateCartPrices(cartItems)
	TotalDiscount := totalDiscount + request.CouponDiscountAmount
	total := (salePrice + tax) - request.CouponDiscountAmount

	IsCodAvailable := true
	for _, itm := range cartItems {
		var productDetail models.ProductDetail
		if err := config.DB.First(&productDetail, itm.CartItem.ProductID).Error; err != nil {
			logger.Log.Error("Product detail not found",
				zap.Uint("productID", itm.CartItem.ProductID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Product Not Found", "Something Went Wrong", "/checkout")
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
	var referralDetails models.ReferralAccount
	if err := config.DB.First(&referralDetails, "user_id = ?", userID).Error; err == nil {
		CheckForReferrer(c)
		CheckForJoinee(c)
	}

	logger.Log.Info("Payment page loaded successfully",
		zap.Uint("userID", userID),
		zap.Float64("total", total),
		zap.Int("itemCount", len(cartItems)))
	c.HTML(http.StatusOK, "paymentPage.html", gin.H{
		"status":          "OK",
		"message":         "Checkout fetch success",
		"Address":         address,
		"CartItem":        cartItems,
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
	logger.Log.Info("Proceeding to payment")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	if err := c.ShouldBind(&paymentRequest); err != nil {
		logger.Log.Error("Failed to bind payment request", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "")
		return
	}

	var userDetails models.UserAuth
	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

	currentTime := time.Now()
	_, cartItems, err := services.FetchCartItems(userID)

	if len(cartItems) == 0 {
		logger.Log.Warn("Cart is empty", zap.Uint("userID", userID))
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/cart")
		return
	}

	reservedProducts := FetchReservedProducts(c, userID)
	if reservedProducts == nil {
		return
	}

	if len(reservedProducts) != len(cartItems) {
		logger.Log.Error("Mismatch between cart items and reserved products",
			zap.Int("cartItemCount", len(cartItems)),
			zap.Int("reservedProductCount", len(reservedProducts)))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 11", "Something Went Wrong", "/cart")
		return
	}

	result, err := ReservedProductCheck(c, reservedProducts, cartItems)
	if err != nil {
		return
	}

	var coupon models.ReservedCoupon
	if err := config.DB.First(&coupon, paymentRequest.CouponId).Error; err != nil {
		logger.Log.Warn("Reserved coupon not found",
			zap.String("couponID", paymentRequest.CouponId),
			zap.Error(err))
	}

	couponDiscountAmount, err := strconv.ParseFloat(paymentRequest.CouponDiscountAmount, 64)
	if err != nil {
		logger.Log.Error("Failed to parse coupon discount amount",
			zap.String("couponDiscountAmount", paymentRequest.CouponDiscountAmount),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Coverting Failed", "Something Went Wrong", "/cart")
		return
	}

	switch paymentRequest.PaymentMethod {
	case "COD":
		paymentStatus := true
		tx := config.DB.Begin()
		orderID := CreateOrder(c, tx, userDetails.ID, result.RegularPrice, result.ProductDiscount, result.TotalDiscount+couponDiscountAmount, result.Tax, float64(result.ShippingCharge), result.Total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
		if orderID == 0 {
			return
		}
		SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
		CreateOrderItems(c, tx, reservedProducts, float64(result.ShippingCharge), orderID, userDetails.ID, currentTime, couponDiscountAmount)
		orderItems := FetchOrderItems(c, tx, orderID)
		if orderItems == nil {
			return
		}

		for _, ordItems := range orderItems {
			if err := CODPayment(c, tx, userDetails.ID, ordItems.OrderUID, ordItems.ID, ordItems.Total, "Cash On Delivery"); err != nil {
				logger.Log.Error("COD payment failed",
					zap.Uint("orderItemID", ordItems.ID),
					zap.Error(err))
				tx.Rollback()
				paymentStatus = false
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment failed", "Something Went Wrong", "/checkout")
				return
			}
			DeleteReservedItems(c, tx, ordItems.ProductVariantID, userID)
		}

		if paymentStatus {
			ClearCart(c, tx, result.ReservedMap)
		}
		if err := tx.Unscoped().Delete(&coupon, paymentRequest.CouponId).Error; err != nil {
			logger.Log.Warn("Failed to delete reserved coupon",
				zap.String("couponID", paymentRequest.CouponId),
				zap.Error(err))
		}
		var orderDetails models.Order
		if err := tx.First(&orderDetails, orderID).Error; err != nil {
			logger.Log.Error("Order not found",
				zap.Uint("orderID", orderID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "")
			return
		}
		tx.Commit()
		logger.Log.Info("COD payment processed successfully",
			zap.Uint("userID", userID),
			zap.Uint("orderID", orderID))
		c.HTML(http.StatusOK, "orderSuccess.html", gin.H{
			"status":        "Success",
			"message":       "Order Success",
			"OrderID":       orderDetails.OrderUID,
			"PaymentMethod": "Cash On Delivery",
			"OrderDate":     orderDetails.CreatedAt.Format("January 2, 2006"),
			"ExpextedDate":  orderDetails.CreatedAt.AddDate(0, 0, 7).Format("January 2, 2006"),
			"code":          http.StatusOK,
		})

	case "Razorpay":
		address := FetchAddressByIDAndUserID(c, userID, paymentRequest.AddressID)
		if address == nil {
			return
		}
		razorpayOrder, err := CreateRazorpayOrder(c, result.Total-couponDiscountAmount)
		if err != nil {
			logger.Log.Error("Failed to create Razorpay order",
				zap.Float64("amount", result.Total-couponDiscountAmount),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create Razorpay order", "Something Went Wrong", "/checkout")
			return
		}

		razorpayOrderID, ok := razorpayOrder["id"].(string)
		if !ok {
			logger.Log.Error("Failed to extract Razorpay order ID",
				zap.Any("razorpayOrder", razorpayOrder))
			log.Println("Failed to extract order ID from Razorpay response:", razorpayOrder)
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Razorpay response", "Something Went Wrong", "/checkout")
			return
		}
		RazorPayOrderID = razorpayOrderID
		logger.Log.Info("Razorpay payment initiated",
			zap.String("razorpayOrderID", razorpayOrderID),
			zap.Float64("amount", result.Total-couponDiscountAmount))
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"order_id": razorpayOrderID,
			"amount":   result.Total * 100,
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
			logger.Log.Error("Wallet not found",
				zap.Uint("userID", userID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Not Found", "Something Went Wrong", "/cart")
			return
		}

		orderID := CreateOrder(c, tx, userDetails.ID, result.RegularPrice, result.ProductDiscount, result.TotalDiscount+couponDiscountAmount, result.Tax, float64(result.ShippingCharge), result.Total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
		if orderID == 0 {
			return
		}
		SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
		CreateOrderItems(c, tx, reservedProducts, float64(result.ShippingCharge), orderID, userDetails.ID, currentTime, couponDiscountAmount)
		orderItems := FetchOrderItems(c, tx, orderID)
		if orderItems == nil {
			return
		}

		var orderDetails models.Order
		if err := tx.First(&orderDetails, orderID).Error; err != nil {
			logger.Log.Error("Order not found",
				zap.Uint("orderID", orderID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "")
			return
		}

		for _, orderItem := range orderItems {
			receiptID := "rcpt-" + uuid.New().String()
			rand.Seed(time.Now().UnixNano())
			transactionID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))

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
				logger.Log.Error("Failed to create payment",
					zap.Uint("orderItemID", orderItem.ID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment creation failed", "Something Went Wrong", "/cart")
				return
			}

			var orderedItem models.OrderItem
			if err := tx.First(&orderedItem, orderItem.ID).Error; err != nil {
				logger.Log.Error("Order item not found",
					zap.Uint("orderItemID", orderItem.ID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "")
				return
			}

			orderedItem.OrderStatus = "Confirmed"
			if err := tx.Save(&orderedItem).Error; err != nil {
				logger.Log.Error("Failed to update order item status",
					zap.Uint("orderItemID", orderItem.ID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Something Went Wrong", "")
				return
			}
			DeleteReservedItems(c, tx, orderItem.ProductVariantID, userID)
		}

		lastBalance := walletDetails.Balance
		walletDetails.Balance -= (result.Total - couponDiscountAmount)
		if err := tx.Save(&walletDetails).Error; err != nil {
			logger.Log.Error("Failed to update wallet balance",
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update wallet details", "Failed to update order", "")
			return
		}

		walletReceiptID := "rcpt-" + uuid.New().String()
		rand.Seed(time.Now().UnixNano())
		walletTransactionID := fmt.Sprintf("-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

		walletHistory := models.WalletTransaction{
			UserID:        userID,
			WalletID:      walletDetails.ID,
			Amount:        result.Total - couponDiscountAmount,
			LastBalance:   lastBalance,
			Description:   "Product Purchase ORD ID" + orderDetails.OrderUID,
			Type:          "Debited",
			Receipt:       walletReceiptID,
			OrderId:       orderDetails.OrderUID,
			TransactionID: walletTransactionID,
			PaymentMethod: "Wallet",
		}
		if err := tx.Create(&walletHistory).Error; err != nil {
			logger.Log.Error("Failed to create wallet transaction",
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Creation Failed", "Something Went Wrong", "/cart")
			return
		}

		ClearCart(c, tx, result.ReservedMap)
		if err := tx.Unscoped().Delete(&coupon, paymentRequest.CouponId).Error; err != nil {
			logger.Log.Warn("Failed to delete reserved coupon",
				zap.String("couponID", paymentRequest.CouponId),
				zap.Error(err))
		}
		tx.Commit()
		logger.Log.Info("Wallet payment processed successfully",
			zap.Uint("userID", userID),
			zap.Uint("orderID", orderID),
			zap.Float64("amount", result.Total))
		c.HTML(http.StatusOK, "orderSuccess.html", gin.H{
			"status":        "Success",
			"message":       "Order Success",
			"OrderID":       orderDetails.OrderUID,
			"PaymentMethod": "Wallet",
			"OrderDate":     orderDetails.CreatedAt.Format("January 2, 2006"),
			"ExpextedDate":  orderDetails.CreatedAt.AddDate(0, 0, 7).Format("January 2, 2006"),
			"code":          http.StatusOK,
		})
	default:
		logger.Log.Warn("Invalid payment method",
			zap.String("method", paymentRequest.PaymentMethod))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/checkout")
		return
	}
}

func PayNow(c *gin.Context) {
	logger.Log.Info("Requested pay now")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var PayNowRequest struct {
		Method      string `json:"method"`
		OrderItemID string `json:"orderId"`
		AddressID   string `json:"addressId"`
	}

	if err := c.ShouldBind(&PayNowRequest); err != nil {
		logger.Log.Error("Failed to bind pay now request", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "/profile/order/details")
		return
	}

	var userDetails models.UserAuth
	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

	if PayNowRequest.Method == "" || PayNowRequest.OrderItemID == "" {
		logger.Log.Error("Missing required fields in pay now request",
			zap.String("method", PayNowRequest.Method),
			zap.String("orderItemID", PayNowRequest.OrderItemID))
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "/profile/order/details")
		return
	}

	switch PayNowRequest.Method {
	case "Razorpay":
		var orderItems models.OrderItem
		if err := config.DB.First(&orderItems, PayNowRequest.OrderItemID).Error; err != nil {
			logger.Log.Error("Order item not found",
				zap.String("orderItemID", PayNowRequest.OrderItemID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Order Item Not Found", "Something Went Wrong", "/profile/order/details")
			return
		}

		var order models.Order
		if err := config.DB.First(&order, orderItems.OrderID).Error; err != nil {
			logger.Log.Error("Order not found",
				zap.Uint("orderID", orderItems.OrderID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "/profile/order/details")
			return
		}

		addressId, adErr := strconv.Atoi(PayNowRequest.AddressID)
		if adErr != nil {
			logger.Log.Error("Invalid address ID",
				zap.String("addressID", PayNowRequest.AddressID),
				zap.Error(adErr))
			helper.RespondWithError(c, http.StatusBadRequest, "Invalid Address", "Something Went Wrong", "/profile/order/details")
			return
		}

		var address models.ShippingAddress
		if err := config.DB.First(&address, "id = ?", addressId).Error; err != nil {
			logger.Log.Error("Address not found",
				zap.Int("addressID", addressId),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Address Not Found", "Something Went Wrong", "/profile/order/details")
			return
		}

		razorpayOrder, err := CreateRazorpayOrder(c, order.TotalAmount)
		if err != nil {
			logger.Log.Error("Failed to create Razorpay order",
				zap.Float64("amount", order.TotalAmount),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create Razorpay order", "Something Went Wrong", "/profile/order/details")
			return
		}

		razorpayOrderID, ok := razorpayOrder["id"].(string)
		if !ok {
			logger.Log.Error("Failed to extract Razorpay order ID",
				zap.Any("razorpayOrder", razorpayOrder))
			log.Println("Failed to extract order ID from Razorpay response:", razorpayOrder)
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Razorpay response", "Something Went Wrong", "/checkout")
			return
		}

		logger.Log.Info("Razorpay pay now initiated",
			zap.String("razorpayOrderID", razorpayOrderID),
			zap.Uint("orderItemID", orderItems.ID))
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
		logger.Log.Warn("Invalid payment method for pay now",
			zap.String("method", PayNowRequest.Method))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/profile/order/details")
		return
	}
}
