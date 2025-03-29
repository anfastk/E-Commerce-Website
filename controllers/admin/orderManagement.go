package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ShowOrderManagement(c *gin.Context) {
	logger.Log.Info("Requested to show order management")

	type OrderListResponse struct {
		ID            uint    `json:"id"`
		ProductImage  string  `json:"productimage"`
		ProductName   string  `json:"productname"`
		OrderID       string  `json:"orderid"`
		OrderDate     string  `json:"orderdate"`
		PaymentMethod string  `json:"paymentmethod"`
		ProfilePic    string  `json:"profilepic"`
		UserName      string  `json:"username"`
		Status        string  `json:"status"`
		Amount        float64 `json:"amount"`
	}

	var orderDetails []models.OrderItem
	if err := config.DB.Order("created_at DESC").Find(&orderDetails).Error; err != nil {
		logger.Log.Error("Failed to fetch orders", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders", "Something Went Wrong", "")
		return
	}

	var orderListResponse []OrderListResponse
	for _, item := range orderDetails {
		var paymentDetails models.PaymentDetail
		if err := config.DB.First(&paymentDetails, "order_item_id = ?", item.ID).Error; err != nil {
			logger.Log.Error("Failed to fetch payment details", zap.Uint("orderItemID", item.ID), zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders payment detail", "Something Went Wrong", "")
			return
		}

		var orderDetails models.Order
		if err := config.DB.First(&orderDetails, "id = ?", item.OrderID).Error; err != nil {
			logger.Log.Error("Failed to fetch order details", zap.Uint("orderID", item.OrderID), zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders detail", "Something Went Wrong", "")
			return
		}

		var userDetails models.UserAuth
		if err := config.DB.First(&userDetails, "id = ?", orderDetails.UserID).Error; err != nil {
			logger.Log.Error("Failed to fetch user details", zap.Uint("userID", orderDetails.UserID), zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch userdetail", "Something Went Wrong", "")
			return
		}

		orderListResponses := OrderListResponse{
			ID:            item.ID,
			ProductImage:  item.ProductImage,
			ProductName:   item.ProductName,
			OrderID:       item.OrderUID,
			OrderDate:     fmt.Sprintf("%s %d%s, %d", item.CreatedAt.Format("Jan"), item.CreatedAt.Day(), ordinalSuffix(item.CreatedAt.Day()), item.CreatedAt.Year()),
			PaymentMethod: paymentDetails.PaymentMethod,
			ProfilePic:    userDetails.ProfilePic,
			UserName:      userDetails.FullName,
			Status:        item.OrderStatus,
			Amount:        item.Total,
		}
		orderListResponse = append(orderListResponse, orderListResponses)
	}

	logger.Log.Info("Order management data fetched successfully", zap.Int("orderCount", len(orderListResponse)))
	c.HTML(http.StatusOK, "orderList.html", gin.H{
		"status":  "success",
		"message": "Order details fetched successfully",
		"data":    orderListResponse,
	})
}

func SearchOrders(c *gin.Context) {
	searchQuery := c.Query("search")
	logger.Log.Info("Search request received", zap.String("query", searchQuery))

	type OrderListResponse struct {
		ID            uint    `json:"id"`
		ProductImage  string  `json:"productimage"`
		ProductName   string  `json:"productname"`
		OrderID       string  `json:"orderid"`
		OrderDate     string  `json:"orderdate"`
		PaymentMethod string  `json:"paymentmethod"`
		ProfilePic    string  `json:"profilepic"`
		UserName      string  `json:"username"`
		Status        string  `json:"status"`
		Amount        float64 `json:"amount"`
	}

	var orderDetails []models.OrderItem
	query := config.DB.Order("created_at DESC")

	if searchQuery != "" {
		query = query.Joins("JOIN orders ON orders.id = order_items.order_id").
			Joins("JOIN user_auths ON user_auths.id = orders.user_id").
			Where("order_items.product_name ILIKE ? OR order_items.order_uid ILIKE ? OR user_auths.full_name ILIKE ?",
				"%"+searchQuery+"%", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}
 
	if err := query.Find(&orderDetails).Error; err != nil {
		logger.Log.Error("Failed to fetch orders", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders", "Something Went Wrong", "")
		return
	}

	var orderListResponse []OrderListResponse
	for _, item := range orderDetails {
		var paymentDetails models.PaymentDetail
		if err := config.DB.First(&paymentDetails, "order_item_id = ?", item.ID).Error; err != nil {
			logger.Log.Error("Failed to fetch payment details", zap.Uint("orderItemID", item.ID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch payment details"})
			return
		}

		var orderDetails models.Order
		if err := config.DB.First(&orderDetails, "id = ?", item.OrderID).Error; err != nil {
			logger.Log.Error("Failed to fetch order details", zap.Uint("orderID", item.OrderID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch order details"})
			return
		}

		var userDetails models.UserAuth
		if err := config.DB.First(&userDetails, "id = ?", orderDetails.UserID).Error; err != nil {
			logger.Log.Error("Failed to fetch user details", zap.Uint("userID", orderDetails.UserID), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to fetch user details"})
			return
		}

		orderListResponses := OrderListResponse{
			ID:            item.ID,
			ProductImage:  item.ProductImage,
			ProductName:   item.ProductName,
			OrderID:       item.OrderUID,
			OrderDate:     fmt.Sprintf("%s %d%s, %d", item.CreatedAt.Format("Jan"), item.CreatedAt.Day(), ordinalSuffix(item.CreatedAt.Day()), item.CreatedAt.Year()),
			PaymentMethod: paymentDetails.PaymentMethod,
			ProfilePic:    userDetails.ProfilePic,
			UserName:      userDetails.FullName,
			Status:        item.OrderStatus,
			Amount:        item.Total,
		}
		orderListResponse = append(orderListResponse, orderListResponses)
	}

	logger.Log.Info("Search completed successfully", zap.Int("resultCount", len(orderListResponse)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Search results fetched successfully",
		"data":    orderListResponse,
	})
}

func ordinalSuffix(day int) string {
	if day >= 11 && day <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func ShowOrderDetailManagement(c *gin.Context) {
	orderItemID := c.Param("id")
	logger.Log.Info("Requested to show order detail management", zap.String("orderItemID", orderItemID))

	var orderItemDetails models.OrderItem
	if err := config.DB.First(&orderItemDetails, "id = ?", orderItemID).Error; err != nil {
		logger.Log.Error("Failed to fetch order item details", zap.String("orderItemID", orderItemID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders items details", "Something Went Wrong", "")
		return
	}

	var UserDetails models.UserAuth
	if err := config.DB.Unscoped().First(&UserDetails, "id = ?", orderItemDetails.UserID).Error; err != nil {
		logger.Log.Error("Failed to fetch user details", zap.Uint("userID", orderItemDetails.UserID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch user details", "Something Went Wrong", "")
		return
	}

	var orderDetails models.Order
	if err := config.DB.First(&orderDetails, "id = ? AND user_id = ?", orderItemDetails.OrderID, UserDetails.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch order details", zap.Uint("orderID", orderItemDetails.OrderID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders details", "Something Went Wrong", "")
		return
	}

	var shippingAddress models.ShippingAddress
	if err := config.DB.First(&shippingAddress, "order_id = ? AND user_id = ?", orderDetails.ID, orderDetails.UserID).Error; err != nil {
		logger.Log.Error("Failed to fetch shipping address", zap.Uint("orderID", orderDetails.ID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders details", "Something Went Wrong", "")
		return
	}

	var paymentDetails models.PaymentDetail
	if err := config.DB.First(&paymentDetails, "user_id = ? AND order_item_id = ?", UserDetails.ID, orderItemDetails.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch payment details", zap.Uint("orderItemID", orderItemDetails.ID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch payment details", "Something Went Wrong", "")
		return
	}

	productDiscount := (orderItemDetails.ProductRegularPrice - orderItemDetails.ProductSalePrice) * float64(orderItemDetails.Quantity)
	totalDiscount := productDiscount + orderDetails.CouponDiscountAmount
	if orderDetails.ShippingCharge == 0 {
		totalDiscount += 100
	}

	IsReurnRequested := false
	var returnRequest models.ReturnRequest
	if err := config.DB.First(&returnRequest, "user_id = ? AND order_item_id = ? AND product_variant_id = ?", UserDetails.ID, orderItemDetails.ID, orderItemDetails.ProductVariantID).Error; err == nil {
		IsReurnRequested = true
	}

	logger.Log.Info("Order details fetched successfully", zap.String("orderItemID", orderItemID))
	c.HTML(http.StatusOK, "orderDetailsManagement.html", gin.H{
		"status":            "success",
		"message":           "Order details fetched successfully",
		"IsReurnRequested":  IsReurnRequested,
		"ReturnRequest":     returnRequest,
		"ReturnRequestDate": returnRequest.CreatedAt.Format("Jan 02, 2006 - 03:04 PM"),
		"OrderItem":         orderItemDetails,
		"OrderDate":         orderItemDetails.CreatedAt.Format("02 Jan 2006"),
		"Order":             orderDetails,
		"User":              UserDetails,
		"Payment":           paymentDetails,
		"Address":           shippingAddress,
		"ProductDiscount":   productDiscount,
		"TotalDiscount":     totalDiscount,
	})
}

func ChangeOrderStatus(c *gin.Context) {
	logger.Log.Info("Requested to change order status")

	type orderManageData struct {
		OrderId       string `json:"orderId"`
		NewStatus     string `json:"status"`
		CurrentStatus string `json:"previousStatus"`
		CancelReason  string `json:"cancelReason"`
		OtherReason   string `json:"otherReason"`
	}

	var updateOrderStatus orderManageData
	if err := c.ShouldBindJSON(&updateOrderStatus); err != nil {
		logger.Log.Error("Invalid request data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Invalid Data", "")
		return
	}

	orderItemID, err := strconv.Atoi(updateOrderStatus.OrderId)
	if err != nil {
		logger.Log.Error("Invalid order ID", zap.String("orderId", updateOrderStatus.OrderId), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Order Id", "Invalid Order Id", "")
		return
	}

	tx := config.DB.Begin()
	var orderItemDetails models.OrderItem
	if err := tx.First(&orderItemDetails, "id = ?", orderItemID).Error; err != nil {
		logger.Log.Error("Failed to fetch order item details", zap.Int("orderItemID", orderItemID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders items details", "Something Went Wrong", "")
		return
	}

	var paymentDetails models.PaymentDetail
	if err := tx.First(&paymentDetails, "order_item_id = ?", orderItemDetails.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch payment details", zap.Uint("orderItemID", orderItemDetails.ID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch payment details", "Something Went Wrong", "")
		return
	}

	switch updateOrderStatus.NewStatus {
	case "Confirmed":
		if err := tx.Model(&orderItemDetails).Update("order_status", "Confirmed").Error; err != nil {
			logger.Log.Error("Failed to update order status to Confirmed", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
	case "Shipped":
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status": "Shipped",
			"shipped_date": time.Now(),
		}).Error; err != nil {
			logger.Log.Error("Failed to update order status to Shipped", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
	case "Out for Delivery":
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status":         "Out For Delivery",
			"out_of_delivery_date": time.Now(),
		}).Error; err != nil {
			logger.Log.Error("Failed to update order status to Out for Delivery", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
	case "Delivered":
		currentTime := time.Now()
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status":  "Delivered",
			"delivery_date": currentTime,
			"return_date":   currentTime.AddDate(0, 0, 7),
		}).Error; err != nil {
			logger.Log.Error("Failed to update order status to Delivered", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&paymentDetails).Update("payment_status", "Completed").Error; err != nil {
			logger.Log.Error("Failed to update payment status to Completed", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
			return
		}
	case "Return":
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status": "Returned",
			"return_date":  time.Now(),
		}).Error; err != nil {
			logger.Log.Error("Failed to update order status to Returned", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&paymentDetails).Update("payment_status", "Refunded").Error; err != nil {
			logger.Log.Error("Failed to update payment status to Refunded", zap.Int("orderItemID", orderItemID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
			return
		}
	case "Cancel":
		if updateOrderStatus.CancelReason == "other" {
			if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
				"order_status": "Cancelled",
				"reason":       updateOrderStatus.OtherReason,
				"cancel_date":  time.Now(),
			}).Error; err != nil {
				logger.Log.Error("Failed to update order status to Cancelled with other reason", zap.Int("orderItemID", orderItemID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
				return
			}
		} else {
			if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
				"order_status": "Cancelled",
				"cancel_date":  time.Now(),
			}).Error; err != nil {
				logger.Log.Error("Failed to update order status to Cancelled", zap.Int("orderItemID", orderItemID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
				return
			}
		}
		if paymentDetails.PaymentMethod == "Cash On Delivery" && paymentDetails.PaymentStatus == "Paid" {
			if err := tx.Model(&paymentDetails).Update("payment_status", "Refunded").Error; err != nil {
				logger.Log.Error("Failed to update COD payment status to Refunded", zap.Int("orderItemID", orderItemID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
				return
			}
		} else if paymentDetails.PaymentMethod == "Cash On Delivery" && paymentDetails.PaymentStatus != "Paid" {
			if err := tx.Model(&paymentDetails).Update("payment_status", "Cancelled").Error; err != nil {
				logger.Log.Error("Failed to update COD payment status to Cancelled", zap.Int("orderItemID", orderItemID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
				return
			}
		}
	}

	tx.Commit()
	logger.Log.Info("Order status updated successfully",
		zap.Int("orderItemID", orderItemID),
		zap.String("newStatus", updateOrderStatus.NewStatus),
		zap.String("previousStatus", updateOrderStatus.CurrentStatus))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Order Status Updated Successfully",
		"code":    200,
	})
}

func ApproveReturn(c *gin.Context) {
	logger.Log.Info("Requested to approve return")

	var input struct {
		ReturnRequestID string `json:"requestUID"`
		OrderID         string `json:"orderId"`
		Status          string `json:"action"`
		AdminNotes      string `json:"adminNotes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Invalid request payload", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Something Went Wrong", "")
		return
	}

	tx := config.DB.Begin()
	ordid, err := strconv.ParseUint(input.OrderID, 10, 32)
	if err != nil {
		logger.Log.Error("Invalid order ID", zap.String("orderID", input.OrderID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Something Went Wrong", "Something Went Wrong", "")
		return
	}

	var returnRequest models.ReturnRequest
	if err := tx.First(&returnRequest, "order_item_id = ? AND request_uid = ?", ordid, input.ReturnRequestID).Error; err != nil {
		logger.Log.Error("Return request not found", zap.String("requestUID", input.ReturnRequestID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Return request not found", "Something Went Wrong", "")
		return
	}

	if returnRequest.Status != "Pending" {
		logger.Log.Warn("Return request already processed", zap.String("requestUID", input.ReturnRequestID), zap.String("status", returnRequest.Status))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusConflict, "Return request already processed", "Something Went Wrong", "")
		return
	}

	if input.Status == "Approved" {
		var orderItems models.OrderItem
		if err := tx.First(&orderItems, "id = ? AND user_id = ?", ordid, returnRequest.UserID).Error; err != nil {
			logger.Log.Error("Order items not found", zap.Uint64("orderItemID", ordid), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Order Items Not Found", "Something Went Wrong", "")
			return
		}

		var order models.Order
		if err := tx.First(&order, "user_id = ? AND id = ?", returnRequest.UserID, orderItems.OrderID).Error; err != nil {
			logger.Log.Error("Order not found", zap.Uint("orderID", orderItems.OrderID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
			return
		}

		var refundAmount float64
		total := order.SubTotal - order.TotalProductDiscount
		productTotal := orderItems.ProductSalePrice
		IscouponRemoved := false
		IsMinusAmount := false

		if order.IsCouponApplied {
			var couponDetails models.Coupon
			if err := tx.Unscoped().First(&couponDetails, "coupon_code = ?", order.CouponCode).Error; err != nil {
				logger.Log.Error("Coupon not found",
					zap.String("couponCode", order.CouponCode),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Coupon Not Found", "Something Went Wrong", "")
				return
			}
			if (total - productTotal) < couponDetails.MinOrderValue {
				refundAmount = orderItems.Total - order.CouponDiscountAmount
				if refundAmount <= 0 {
					IsMinusAmount = true
				}
				couponDetails.UsersUsedCount -= 1
				if err := tx.Unscoped().Save(&couponDetails).Error; err != nil {
					logger.Log.Error("Failed to update coupon",
						zap.String("couponCode", order.CouponCode),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Coupon", "Something Went Wrong", "")
					return
				}
				IscouponRemoved = true
			}
		} else {
			refundAmount = orderItems.Total
		}

		var product models.ProductVariantDetails
		if err := tx.First(&product, orderItems.ProductVariantID).Error; err != nil {
			logger.Log.Error("Product not found", zap.Uint("productVariantID", orderItems.ProductVariantID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&product).Where("id = ?", orderItems.ProductVariantID).
			Update("stock_quantity", product.StockQuantity+orderItems.Quantity).Error; err != nil {
			logger.Log.Error("Failed to update stock quantity", zap.Uint("productVariantID", orderItems.ProductVariantID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Stock Reverse Failed", "Stock Update Failed", "/checkout")
			return
		}

		var payment models.PaymentDetail
		if err := tx.First(&payment, "order_item_id = ? AND user_id = ?", orderItems.ID, returnRequest.UserID).Error; err != nil {
			logger.Log.Error("Payment details not found", zap.Uint("orderItemID", orderItems.ID), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Payment Details Not Found", "Something Went Wrong", "")
			return
		}

		if payment.PaymentStatus == "Completed" {
			var wallet models.Wallet
			if err := tx.First(&wallet, "user_id = ?", returnRequest.UserID).Error; err != nil {
				logger.Log.Error("Wallet not found", zap.Uint("userID", returnRequest.UserID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "User Wallet Not Found", "Something Went Wrong", "")
				return
			}
			if IsMinusAmount {
				lastBalance := wallet.Balance
				wallet.Balance += -order.CouponDiscountAmount
				if err := tx.Save(&wallet).Error; err != nil {
					logger.Log.Error("Failed to update wallet",
						zap.Uint("userID", returnRequest.UserID),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
					return
				}

				receiptID := "rcpt_" + uuid.New().String()
				rand.Seed(time.Now().UnixNano())
				transactionID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))

				walletTransaction := models.WalletTransaction{
					UserID:        returnRequest.UserID,
					WalletID:      wallet.ID,
					Amount:        order.CouponDiscountAmount,
					Description:   fmt.Sprintf("Coupon discount adjustment due to partial order cancellation. ₹%.2f deducted from wallet as per refund policy. ORD ID %s", order.CouponDiscountAmount, orderItems.OrderUID),
					Type:          "Deduct",
					Receipt:       receiptID,
					OrderId:       orderItems.OrderUID,
					LastBalance:   lastBalance,
					TransactionID: strings.ToUpper(transactionID),
					PaymentMethod: payment.PaymentMethod,
				}
				if err := tx.Create(&walletTransaction).Error; err != nil {
					logger.Log.Error("Failed to create wallet transaction",
						zap.Uint("userID", returnRequest.UserID),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
					return
				}

				lastBalance = wallet.Balance
				wallet.Balance += orderItems.ProductSalePrice + orderItems.Tax
				if err := tx.Save(&wallet).Error; err != nil {
					logger.Log.Error("Failed to update wallet",
						zap.Uint("userID", returnRequest.UserID),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
					return
				}

				receiptID = "rcpt_" + uuid.New().String()
				rand.Seed(time.Now().UnixNano())
				transactionID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))

				walletTransaction = models.WalletTransaction{
					UserID:        returnRequest.UserID,
					WalletID:      wallet.ID,
					Amount:        orderItems.Total,
					Description:   fmt.Sprintf("Order Refund ORD ID " + orderItems.OrderUID),
					Type:          "Refund",
					Receipt:       receiptID,
					OrderId:       orderItems.OrderUID,
					LastBalance:   lastBalance,
					TransactionID: strings.ToUpper(transactionID),
					PaymentMethod: payment.PaymentMethod,
				}
				if err := tx.Create(&walletTransaction).Error; err != nil {
					logger.Log.Error("Failed to create wallet transaction",
						zap.Uint("userID", returnRequest.UserID),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
					return
				}

			} else {
				lastBalance := wallet.Balance
				wallet.Balance += refundAmount
				if err := tx.Save(&wallet).Error; err != nil {
					logger.Log.Error("Failed to update wallet",
						zap.Uint("userID", returnRequest.UserID),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
					return
				}

				receiptID := "rcpt_" + uuid.New().String()
				rand.Seed(time.Now().UnixNano())
				transactionID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))

				walletTransaction := models.WalletTransaction{
					UserID:        returnRequest.UserID,
					WalletID:      wallet.ID,
					Amount:        refundAmount,
					Description:   fmt.Sprintf("Order Refund ORD ID " + orderItems.OrderUID),
					Type:          "Refund",
					Receipt:       receiptID,
					OrderId:       orderItems.OrderUID,
					LastBalance:   lastBalance,
					TransactionID: strings.ToUpper(transactionID),
					PaymentMethod: payment.PaymentMethod,
				}
				if err := tx.Create(&walletTransaction).Error; err != nil {
					logger.Log.Error("Failed to create wallet transaction",
						zap.Uint("userID", returnRequest.UserID),
						zap.Error(err))
					tx.Rollback()
					helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
					return
				}
			}
			if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", returnRequest.UserID, orderItems.ID).
				Update("payment_status", "Refunded").Error; err != nil {
				logger.Log.Error("Failed to update payment status to Refunded", zap.Uint("orderItemID", orderItems.ID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment Status Update Failed", "Payment Status Update Failed", "/checkout")
				return
			}
		}

		shipCharge := 100
		if order.SubTotal-orderItems.SubTotal > 1000 {
			shipCharge = 0
		}

		if IscouponRemoved {
			if err := tx.Model(&order).Where("user_id = ? AND id = ?", returnRequest.UserID, order.ID).
				Updates(map[string]interface{}{
					"coupon_code":            gorm.Expr("NULL"),
					"coupon_id":              gorm.Expr("NULL"),
					"shipping_charge":        shipCharge,
					"coupon_discount_amount": gorm.Expr("NULL"),
					"is_coupon_applied":      false,
				}).Error; err != nil {
				logger.Log.Error("Failed to update order with coupon removal", zap.Uint("orderID", order.ID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
				return
			}
		} else {
			if err := tx.Model(&order).Where("user_id = ? AND id = ?", returnRequest.UserID, order.ID).
				Updates(map[string]interface{}{
					"shipping_charge":        shipCharge,
					"coupon_discount_amount": gorm.Expr("NULL"),
				}).Error; err != nil {
				logger.Log.Error("Failed to update order shipping", zap.Uint("orderID", order.ID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
				return
			}
		}

		if err := tx.Model(&orderItems).Where("user_id = ? AND order_id = ?", returnRequest.UserID, order.ID).
			Updates(map[string]interface{}{
				"order_status":           "Returned",
				"reason":                 returnRequest.Reason,
				"cancel_date":            time.Now(),
				"expected_delivery_date": time.Now(),
				"return_date":            time.Now(),
			}).Error; err != nil {
			logger.Log.Error("Failed to update order item to Returned", zap.Uint64("orderItemID", ordid), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
			return
		}
	}

	if err := tx.Model(&returnRequest).Where("request_uid = ?", input.ReturnRequestID).Updates(map[string]interface{}{
		"status":      input.Status,
		"admin_notes": input.AdminNotes,
	}).Error; err != nil {
		logger.Log.Error("Failed to update return request status", zap.String("requestUID", input.ReturnRequestID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update status", "Something Went Wrong", "")
		return
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed to commit transaction", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Transaction Failed", "Order cancellation failed", "")
		return
	}
	logger.Log.Info("Return request processed successfully",
		zap.String("requestUID", input.ReturnRequestID),
		zap.String("status", input.Status))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Return request processed successfully.",
		"code":    http.StatusOK,
	})
}
