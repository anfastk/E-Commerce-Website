package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
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

// Add this function to create a Razorpay order
func CreateRazorpayOrder(c *gin.Context, amount float64) (map[string]interface{}, error) {

	client := razorpay.NewClient(config.RAZORPAY_KEY_ID, config.RAZORPAY_KEY_SECRET)
	receiptID := uuid.New().String()[:30] // Taking only the first 30 characters

	data := map[string]interface{}{
		"amount":   int64(amount * 100), // Amount in smallest currency unit (paise for INR)
		"currency": "INR",
		"receipt":  "rcpt_" + receiptID,
	}
	log.Println("Creating Razorpay Order with Data:", data) // Debug log

	order, err := client.Order.Create(data, nil)
	if err != nil {
		log.Println("Razorpay Order Creation Error:", err) // Log the exact error
		return nil, err
	}

	return order, nil
}

var razorPayOrderID string

// Add this function for verifying Razorpay payments
func VerifyRazorpayPayment(c *gin.Context) {
	currentTime := time.Now()
	userID := c.MustGet("userid").(uint)
	var verifyRequest struct {
		PaymentID    string `json:"razorpay_payment_id"`
		OrderID      string `json:"razorpay_order_id"`
		Signature    string `json:"razorpay_signature"`
		OrderItemIDs []uint `json:"order_item_ids"`
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
	reservedMap, shippingCharge, tax, total := ReservedProductCheck(c, reservedProducts, cartItems)

	tx := config.DB.Begin()

	orderID := CreateOrder(c, tx, userDetails.ID, total, tax, float64(shippingCharge), currentTime)
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
			OrderId:       razorPayOrderID,
			TransactionID: verifyRequest.PaymentID,
			Receipt:       receiptID,
		}

		if err := tx.Create(&createPayment).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Payment creation failed", "Something Went Wrong", "/cart")
			return
		}

		// Update order status
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
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment verified successfully",
	})
}

func PaymentFailureHandler(c *gin.Context) {
	currentTime := time.Now()
	userID := c.MustGet("userid").(uint)

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
	reservedMap, shippingCharge, tax, total := ReservedProductCheck(c, reservedProducts, cartItems)

	tx := config.DB.Begin()

	orderID := CreateOrder(c, tx, userDetails.ID, total, tax, float64(shippingCharge), currentTime)
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
			OrderId:       razorPayOrderID,
			TransactionID: verifyRequest.PaymentID,
			Receipt:       receiptID,
		}

		if err := tx.Create(&createPayment).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Payment creation failed", "Something Went Wrong", "/cart")
			return
		}

		// Update order status
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
	tx.Commit()

	c.Redirect(http.StatusSeeOther, "/profile/order/details")
}
