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
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

func CODPayment(c *gin.Context, tx *gorm.DB, userId uint, orderId uint, paymentAmount float64, method string) error {
	var paymentDetail models.PaymentDetail
	paymentDetail = models.PaymentDetail{
		UserID:        userId,
		OrderItemID:   orderId,
		PaymentStatus: "Pending",
		PaymentAmount: paymentAmount,
		PaymentMethod: method,
	}
	if err := tx.Create(&paymentDetail).Error; err != nil {
		tx.Rollback()
		return errors.New("Payment Creation Failed")
	}
	return nil
}

func CreateRazorpayOrder(c *gin.Context, amount float64) (map[string]interface{}, error) {

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
		log.Println("Razorpay Order Creation Error:", err)
		return nil, err
	}

	return order, nil
}

var RazorPayOrderID string

func VerifyRazorpayPayment(c *gin.Context) {
	currentTime := time.Now()
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

	var verifyRequest struct {
		PaymentID string `json:"razorpay_payment_id"`
		OrderID   string `json:"razorpay_order_id"`
		Signature string `json:"razorpay_signature"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}
	var userDetails models.UserAuth

	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

	/* 	client := razorpay.NewClient(RAZORPAY_KEY_ID, RAZORPAY_SECRET)*/
	data := verifyRequest.OrderID + "|" + verifyRequest.PaymentID
	expectedSignature := hmac.New(sha256.New, []byte(config.RAZORPAY_KEY_SECRET))
	expectedSignature.Write([]byte(data))
	calculatedSignature := hex.EncodeToString(expectedSignature.Sum(nil))

	if calculatedSignature != verifyRequest.Signature {
		fmt.Println("Payment verification failed")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid signature", "Payment verification failed", "")
		return
	}

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

	tx := config.DB.Begin()

	var coupon models.ReservedCoupon
	tx.First(&coupon, paymentRequest.CouponId)
	couponDiscountAmount, err := strconv.ParseFloat(paymentRequest.CouponDiscountAmount, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Coverting Failed", "Something Went Wrong", "/cart")
		return
	}
	orderID := CreateOrder(c, tx, userDetails.ID, subtotal, totalProductDiscount, totalDiscount+couponDiscountAmount, tax, float64(shippingCharge), total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
	SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
	CreateOrderItems(c, tx, reservedProducts, float64(shippingCharge), orderID, userDetails.ID, currentTime)
	orderItems := FetchOrderItems(c, tx, orderID)
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
			PaymentAmount: total,
			PaymentMethod: "Razorpay",
			OrderId:       RazorPayOrderID,
			TransactionID: verifyRequest.PaymentID,
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Order not found", "")
			return
		}

		orderedItem.OrderStatus = "Confirmed"
		if err := tx.Save(&orderedItem).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Failed to update order", "")
			return
		}
		DeleteReservedItems(c, tx, orderItem.ProductVariantID, userID)
	}
	ClearCart(c, tx, reservedMap)
	tx.Unscoped().Delete(&coupon, paymentRequest.CouponId)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment verified successfully",
	})
}

func PaymentFailureHandler(c *gin.Context) {
	currentTime := time.Now()
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

	var verifyRequest struct {
		PaymentID string `json:"razorpay_payment_id"`
		OrderID   string `json:"razorpay_order_id"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}

	var userDetails models.UserAuth

	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

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

	tx := config.DB.Begin()
	var coupon models.ReservedCoupon
	tx.First(&coupon, paymentRequest.CouponId)
	couponDiscountAmount, err := strconv.ParseFloat(paymentRequest.CouponDiscountAmount, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Coverting Failed", "Something Went Wrong", "/cart")
		return
	}
	orderID := CreateOrder(c, tx, userDetails.ID, subtotal, totalProductDiscount, totalDiscount+couponDiscountAmount, tax, float64(shippingCharge), total-couponDiscountAmount, currentTime, paymentRequest.CouponCode, couponDiscountAmount, coupon.Discription)
	SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
	CreateOrderItems(c, tx, reservedProducts, float64(shippingCharge), orderID, userDetails.ID, currentTime)
	orderItems := FetchOrderItems(c, tx, orderID)
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
			PaymentAmount: total,
			PaymentMethod: "Razorpay",
			OrderId:       RazorPayOrderID,
			TransactionID: verifyRequest.PaymentID,
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Order not found", "")
			return
		}

		orderedItem.OrderStatus = "Order Not Placed"
		if err := tx.Save(&orderedItem).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Failed to update order", "")
			return
		}
		DeleteReservedItems(c, tx, orderItem.ProductVariantID, userID)
	}
	ClearCart(c, tx, reservedMap)
	tx.Unscoped().Delete(&coupon, paymentRequest.CouponId)

	tx.Commit()

	c.Redirect(http.StatusSeeOther, "/profile/order/details")
}

func VerifyPayNowRazorpayPayment(c *gin.Context) {
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

	var verifyRequest struct {
		PaymentID   string `json:"razorpay_payment_id"`
		OrderID     string `json:"razorpay_order_id"`
		Signature   string `json:"razorpay_signature"`
		OrderItemID string `json:"order_id"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}

	var userDetails models.UserAuth

	if err := config.DB.First(&userDetails, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/profile/order/details")
		return
	}

	data := verifyRequest.OrderID + "|" + verifyRequest.PaymentID
	expectedSignature := hmac.New(sha256.New, []byte(config.RAZORPAY_KEY_SECRET))
	expectedSignature.Write([]byte(data))
	calculatedSignature := hex.EncodeToString(expectedSignature.Sum(nil))

	if calculatedSignature != verifyRequest.Signature {
		fmt.Println("Payment verification failed")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid signature", "Payment verification failed", "/profile/order/details")
		return
	}
	var orderItem models.OrderItem
	ordrID, _ := strconv.Atoi(verifyRequest.OrderItemID)
	if err := config.DB.First(&orderItem, ordrID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Order Item not found", "Something Went Wrong", "/profile/order/details")
		return
	}
	var orderDetails models.Order
	if err := config.DB.First(&orderDetails, orderItem.OrderID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Order not found", "Something Went Wrong", "/profile/order/details")
		return
	}

	var allOrderItem []models.OrderItem
	if err := config.DB.Find(&allOrderItem, "order_id = ?", orderDetails.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Order Items not found", "Something Went Wrong", "/profile/order/details")
		return
	}

	for _, items := range allOrderItem {

		var paymentDetails models.PaymentDetail
		if err := config.DB.First(&paymentDetails, "order_item_id = ?", items.ID).Error; err != nil {
			helper.RespondWithError(c, http.StatusNotFound, "Payment Details not found", "Something Went Wrong", "/profile/order/details")
			return
		}
		paymentDetails.PaymentStatus = "Completed"
		if err := config.DB.Save(&paymentDetails).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment details", "Failed to update order", "/profile/order/details")
			return
		}

		items.OrderStatus = "Confirmed"
		if err := config.DB.Save(&items).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Failed to update order", "/profile/order/details")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Payment verified successfully",
		})
	}
}
