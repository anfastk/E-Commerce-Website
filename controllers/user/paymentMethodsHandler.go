package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
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
	"github.com/razorpay/razorpay-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CODPayment(c *gin.Context, tx *gorm.DB, userId uint, orderId string, order_item_id uint, paymentAmount float64, method string) error {
	logger.Log.Info("Processing COD payment",
		zap.Uint("userID", userId),
		zap.String("orderID", orderId),
		zap.Uint("orderItemID", order_item_id))

	paymentDetail := models.PaymentDetail{
		UserID:        userId,
		OrderItemID:   order_item_id,
		OrderId:       orderId,
		PaymentStatus: "Pending",
		PaymentAmount: paymentAmount,
		PaymentMethod: method,
	}
	if err := tx.Create(&paymentDetail).Error; err != nil {
		logger.Log.Error("Failed to create COD payment",
			zap.Uint("userID", userId),
			zap.String("orderID", orderId),
			zap.Error(err))
		tx.Rollback()
		return errors.New("Payment Creation Failed")
	}

	logger.Log.Info("COD payment created successfully",
		zap.Uint("paymentID", paymentDetail.ID))
	return nil
}

func FetchWalletBalance(c *gin.Context) {
	logger.Log.Info("Fetching wallet balance")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var walletDetails models.Wallet
	if err := config.DB.First(&walletDetails, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Wallet not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Wallet balance fetched successfully",
		zap.Uint("userID", userID),
		zap.Float64("balance", walletDetails.Balance))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"balance": walletDetails.Balance,
		"code":    http.StatusOK,
	})
}

func CreateRazorpayOrder(c *gin.Context, amount float64) (map[string]interface{}, error) {
	logger.Log.Info("Creating Razorpay order",
		zap.Float64("amount", amount))

	client := razorpay.NewClient(config.RAZORPAY_KEY_ID, config.RAZORPAY_KEY_SECRET)
	receiptID := uuid.New().String()[:30]

	data := map[string]interface{}{
		"amount":   int64(amount * 100),
		"currency": "INR",
		"receipt":  "rcpt_" + receiptID,
	}
	log.Println("Creating Razorpay Order with Data:", data)

	order, err := client.Order.Create(data, nil)
	if err != nil {
		logger.Log.Error("Failed to create Razorpay order",
			zap.Float64("amount", amount),
			zap.Error(err))
		log.Println("Razorpay Order Creation Error:", err)
		return nil, err
	}

	orderID, _ := order["id"].(string)
	logger.Log.Info("Razorpay order created successfully",
		zap.String("orderID", orderID))
	return order, nil
}

var RazorPayOrderID string

func VerifyRazorpayPayment(c *gin.Context) {
	logger.Log.Info("Verifying Razorpay payment")

	currentTime := time.Now()
	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var verifyRequest struct {
		PaymentID string `json:"razorpay_payment_id"`
		OrderID   string `json:"razorpay_order_id"`
		Signature string `json:"razorpay_signature"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		logger.Log.Error("Failed to bind verification request", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
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

	data := verifyRequest.OrderID + "|" + verifyRequest.PaymentID
	expectedSignature := hmac.New(sha256.New, []byte(config.RAZORPAY_KEY_SECRET))
	expectedSignature.Write([]byte(data))
	calculatedSignature := hex.EncodeToString(expectedSignature.Sum(nil))

	if calculatedSignature != verifyRequest.Signature {
		logger.Log.Error("Payment signature verification failed",
			zap.String("orderID", verifyRequest.OrderID),
			zap.String("paymentID", verifyRequest.PaymentID))
		fmt.Println("Payment verification failed")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid signature", "Payment verification failed", "")
		return
	}

	_, cartItems, err := services.FetchCartItems(userID)

	if len(cartItems) == 0 {
		logger.Log.Warn("Cart is empty", zap.Uint("userID", userID))
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/cart")
		return
	}

	reservedProducts := FetchReservedProducts(c, userID)
	if reservedProducts == nil {
		helper.RespondWithError(c, http.StatusNotFound, "Failed to fetch reserved stock", "Something Went Wrong", "/cart")
		return
	}

	if len(reservedProducts) != len(cartItems) {
		logger.Log.Error("Mismatch between cart items and reserved products",
			zap.Int("cartItemCount", len(cartItems)),
			zap.Int("reservedProductCount", len(reservedProducts)))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product", "Something Went Wrong", "/cart")
		return
	}

	result, err := ReservedProductCheck(c, reservedProducts, cartItems)
	if err != nil {
		logger.Log.Error(err.Error(),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, err.Error(), "Something Went Wrong", "/cart")
		return
	}

	tx := config.DB.Begin()
	var coupon models.ReservedCoupon
	if err := tx.First(&coupon, paymentRequest.CouponId).Error; err != nil {
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

	orderID := CreateOrder(c, tx, userDetails.ID, result.RegularPrice, result.ProductDiscount, result.TotalDiscount+couponDiscountAmount, result.Tax, float64(result.ShippingCharge), result.Total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
	if orderID == 0 {
		helper.RespondWithError(c, http.StatusNotFound, "Order not found", "Something Went Wrong", "/cart")
		return
	}

	SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
	CreateOrderItems(c, tx, reservedProducts, float64(result.ShippingCharge), orderID, userDetails.ID, currentTime, couponDiscountAmount)
	orderItems := FetchOrderItems(c, tx, orderID)
	if orderItems == nil {
		helper.RespondWithError(c, http.StatusNotFound, "Failed to fetch order items", "Something Went Wrong", "/cart")
		return
	}

	var OrderItemIDs []int
	for _, item := range orderItems {
		OrderItemIDs = append(OrderItemIDs, int(item.ID))
	}

	for _, orderItem := range orderItems {
		receiptID := "rcpt_" + uuid.New().String()
		createPayment := models.PaymentDetail{
			UserID:        userID,
			OrderItemID:   orderItem.ID,
			PaymentStatus: "Completed",
			PaymentAmount: orderItem.Total,
			PaymentMethod: "Razorpay",
			OrderId:       RazorPayOrderID,
			TransactionID: verifyRequest.PaymentID,
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Order not found", "")
			return
		}

		orderedItem.OrderStatus = "Confirmed"
		if err := tx.Save(&orderedItem).Error; err != nil {
			logger.Log.Error("Failed to update order status",
				zap.Uint("orderItemID", orderItem.ID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Failed to update order", "")
			return
		}
		DeleteReservedItems(c, tx, orderItem.ProductVariantID, userID)
	}

	ClearCart(c, tx, result.ReservedMap)
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

	logger.Log.Info("Razorpay payment verified successfully",
		zap.Uint("userID", userID),
		zap.Uint("orderID", orderID),
		zap.String("paymentID", verifyRequest.PaymentID))
	c.HTML(http.StatusOK, "orderSuccess.html", gin.H{
		"status":        "Success",
		"message":       "Order Success",
		"OrderID":       orderDetails.OrderUID,
		"PaymentMethod": "Razorpay",
		"OrderDate":     orderDetails.CreatedAt.Format("January 2, 2006"),
		"ExpextedDate":  orderDetails.CreatedAt.AddDate(0, 0, 7).Format("January 2, 2006"),
		"code":          http.StatusOK,
	})
}

func PaymentFailureHandler(c *gin.Context) {
	logger.Log.Info("Handling payment failure")

	currentTime := time.Now()
	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var verifyRequest struct {
		PaymentID string `json:"razorpay_payment_id"`
		OrderID   string `json:"razorpay_order_id"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		logger.Log.Error("Failed to bind failure request", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
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

	_, cartItems, err := services.FetchCartItems(userID)

	if len(cartItems) == 0 {
		logger.Log.Warn("Cart is empty", zap.Uint("userID", userID))
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/cart")
		return
	}

	reservedProducts := FetchReservedProducts(c, userID)
	if reservedProducts == nil {
		helper.RespondWithError(c, http.StatusNotFound, "Failed to fetch reserved stock", "Something Went Wrong", "/cart")
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
		logger.Log.Error(err.Error(),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, err.Error(), "Something Went Wrong", "/cart")
		return
	}

	tx := config.DB.Begin()
	var coupon models.ReservedCoupon
	if err := tx.First(&coupon, paymentRequest.CouponId).Error; err != nil {
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

	orderID := CreateOrder(c, tx, userDetails.ID, result.RegularPrice, result.ProductDiscount, result.TotalDiscount+couponDiscountAmount, result.Tax, float64(result.ShippingCharge), result.Total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
	if orderID == 0 {
		helper.RespondWithError(c, http.StatusNotFound, "Order not found", "Something Went Wrong", "/cart")
		return
	}

	SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
	CreateOrderItems(c, tx, reservedProducts, float64(result.ShippingCharge), orderID, userDetails.ID, currentTime, couponDiscountAmount)
	orderItems := FetchOrderItems(c, tx, orderID)
	if orderItems == nil {
		helper.RespondWithError(c, http.StatusNotFound, "Failed to fetch order items", "Something Went Wrong", "/cart")
		return
	}

	var OrderItemIDs []int
	for _, item := range orderItems {
		OrderItemIDs = append(OrderItemIDs, int(item.ID))
	}

	for _, orderItem := range orderItems {
		receiptID := "rcpt_" + uuid.New().String()
		createPayment := models.PaymentDetail{
			UserID:        userID,
			OrderItemID:   orderItem.ID,
			PaymentStatus: "Failed",
			PaymentAmount: result.Total,
			PaymentMethod: "Razorpay",
			OrderId:       RazorPayOrderID,
			TransactionID: verifyRequest.PaymentID,
			Receipt:       receiptID,
		}

		if err := tx.Create(&createPayment).Error; err != nil {
			logger.Log.Error("Failed to create failed payment",
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Order not found", "")
			return
		}

		orderedItem.OrderStatus = "Order Not Placed"
		if err := tx.Save(&orderedItem).Error; err != nil {
			logger.Log.Error("Failed to update order status",
				zap.Uint("orderItemID", orderItem.ID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Failed to update order", "")
			return
		}
		DeleteReservedItems(c, tx, orderItem.ProductVariantID, userID)
	}

	ClearCart(c, tx, result.ReservedMap)
	if err := tx.Unscoped().Delete(&coupon, paymentRequest.CouponId).Error; err != nil {
		logger.Log.Warn("Failed to delete reserved coupon",
			zap.String("couponID", paymentRequest.CouponId),
			zap.Error(err))
	}
	tx.Commit()

	logger.Log.Info("Payment failure handled",
		zap.Uint("userID", userID),
		zap.Uint("orderID", orderID),
		zap.String("paymentID", verifyRequest.PaymentID))
	c.Redirect(http.StatusSeeOther, "/profile/order/details")
}

func VerifyPayNowRazorpayPayment(c *gin.Context) {
	logger.Log.Info("Verifying PayNow Razorpay payment")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var verifyRequest struct {
		PaymentID   string `json:"razorpay_payment_id"`
		OrderID     string `json:"razorpay_order_id"`
		Signature   string `json:"razorpay_signature"`
		OrderItemID string `json:"order_id"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		logger.Log.Error("Failed to bind verification request", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}

	var userDetails models.UserAuth
	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/profile/order/details")
		return
	}

	data := verifyRequest.OrderID + "|" + verifyRequest.PaymentID
	expectedSignature := hmac.New(sha256.New, []byte(config.RAZORPAY_KEY_SECRET))
	expectedSignature.Write([]byte(data))
	calculatedSignature := hex.EncodeToString(expectedSignature.Sum(nil))

	if calculatedSignature != verifyRequest.Signature {
		logger.Log.Error("Payment signature verification failed",
			zap.String("orderID", verifyRequest.OrderID),
			zap.String("paymentID", verifyRequest.PaymentID))
		fmt.Println("Payment verification failed")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid signature", "Payment verification failed", "/profile/order/details")
		return
	}

	var orderItem models.OrderItem
	ordrID, err := strconv.Atoi(verifyRequest.OrderItemID)
	if err != nil {
		logger.Log.Error("Invalid order item ID",
			zap.String("orderItemID", verifyRequest.OrderItemID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid order ID", "Something Went Wrong", "/profile/order/details")
		return
	}

	if err := config.DB.First(&orderItem, ordrID).Error; err != nil {
		logger.Log.Error("Order item not found",
			zap.Int("orderItemID", ordrID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Order Item not found", "Something Went Wrong", "/profile/order/details")
		return
	}

	var orderDetails models.Order
	if err := config.DB.First(&orderDetails, orderItem.OrderID).Error; err != nil {
		logger.Log.Error("Order not found",
			zap.Uint("orderID", orderItem.OrderID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Order not found", "Something Went Wrong", "/profile/order/details")
		return
	}

	var allOrderItem []models.OrderItem
	if err := config.DB.Find(&allOrderItem, "order_id = ?", orderDetails.ID).Error; err != nil {
		logger.Log.Error("Order items not found",
			zap.Uint("orderID", orderDetails.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Order Items not found", "Something Went Wrong", "/profile/order/details")
		return
	}

	for _, items := range allOrderItem {
		var paymentDetails models.PaymentDetail
		if err := config.DB.First(&paymentDetails, "order_item_id = ?", items.ID).Error; err != nil {
			logger.Log.Error("Payment details not found",
				zap.Uint("orderItemID", items.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusNotFound, "Payment Details not found", "Something Went Wrong", "/profile/order/details")
			return
		}

		paymentDetails.PaymentStatus = "Completed"
		if err := config.DB.Save(&paymentDetails).Error; err != nil {
			logger.Log.Error("Failed to update payment details",
				zap.Uint("paymentID", paymentDetails.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment details", "Failed to update order", "/profile/order/details")
			return
		}

		items.OrderStatus = "Confirmed"
		if err := config.DB.Save(&items).Error; err != nil {
			logger.Log.Error("Failed to update order item status",
				zap.Uint("orderItemID", items.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Failed to update order", "/profile/order/details")
			return
		}
	}

	logger.Log.Info("PayNow Razorpay payment verified successfully",
		zap.Uint("userID", userID),
		zap.Int("orderItemID", ordrID),
		zap.String("paymentID", verifyRequest.PaymentID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment verified successfully",
	})
}
