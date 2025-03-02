package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ShowOrderManagent(c *gin.Context) {
	type OrderListResponce struct {
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
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders", "Something Went Wrong", "")
		return
	}

	var orderListResponce []OrderListResponce
	for _, item := range orderDetails {
		var paymentDetails models.PaymentDetail
		if err := config.DB.First(&paymentDetails, "order_item_id = ?", item.ID).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders payment detail", "Something Went Wrong", "")
			return
		}
		var orderDetails models.Order
		if err := config.DB.First(&orderDetails, "id = ?", item.OrderID).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders detail", "Something Went Wrong", "")
			return
		}
		var userDetails models.UserAuth
		if err := config.DB.First(&userDetails, "id = ?", orderDetails.UserID).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch userdetail", "Something Went Wrong", "")
			return
		}
		orderListResponces := OrderListResponce{
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
		orderListResponce = append(orderListResponce, orderListResponces)
	}
	c.HTML(http.StatusOK, "orderList.html", gin.H{
		"status":  "success",
		"message": "Order details fetched successfully",
		"data":    orderListResponce,
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

	var orderItemDetails models.OrderItem
	if err := config.DB.First(&orderItemDetails, "id = ?", orderItemID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders items details", "Something Went Wrong", "")
		return
	}
	var UserDetails models.UserAuth
	if err := config.DB.Unscoped().First(&UserDetails, "id = ?", orderItemDetails.UserID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch user details", "Something Went Wrong", "")
		return
	}
	var orderDetails models.Order
	if err := config.DB.First(&orderDetails, "id = ? AND user_id = ?", orderItemDetails.OrderID, UserDetails.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders details", "Something Went Wrong", "")
		return
	}
	var shippingAddress models.ShippingAddress
	if err := config.DB.First(&shippingAddress, "order_id = ? AND user_id = ?", orderDetails.ID, orderDetails.UserID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders details", "Something Went Wrong", "")
		return
	}
	var paymentDetails models.PaymentDetail
	if err := config.DB.First(&paymentDetails, "user_id = ? AND order_item_id = ?", UserDetails.ID, orderItemDetails.ID).Error; err != nil {
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
	type orderManageData struct {
		OrderId       string `json:"orderId"`
		NewStatus     string `json:"status"`
		CurrentStatus string `json:"previousStatus"`
		CancelReason  string `json:"cancelReason"`
		OtherReason   string `json:"otherReason"`
	}

	var updateOrderStatus orderManageData
	if err := c.ShouldBindJSON(&updateOrderStatus); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Invalid Data", "")
		return
	}
	orderItemID, err := strconv.Atoi(updateOrderStatus.OrderId)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Order Id", "Invalid Order Id", "")
		return
	}
	tx := config.DB.Begin()
	var orderItemDetails models.OrderItem
	if err := tx.First(&orderItemDetails, "id = ?", orderItemID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch orders items details", "Something Went Wrong", "")
		return
	}
	var paymentDetails models.PaymentDetail
	if err := tx.First(&paymentDetails, "order_item_id = ?", orderItemDetails.ID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch payment details", "Something Went Wrong", "")
		return
	}

	switch updateOrderStatus.NewStatus {
	case "Confirmed":
		if err := tx.Model(&orderItemDetails).Update("order_status", "Confirmed").Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
	case "Shipped":
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status": "Shipped",
			"shipped_date": time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
	case "Out for Delivery":
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status":         "Out For Delivery",
			"out_of_delivery_date": time.Now(),
		}).Error; err != nil {
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
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&paymentDetails).Update("payment_status", "Completed").Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
			return
		}
	case "Return":
		if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
			"order_status": "Returned",
			"return_date":  time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&paymentDetails).Update("payment_status", "Refunded").Error; err != nil {
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
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
				return
			}
		} else {
			if err := tx.Model(&orderItemDetails).Updates(map[string]interface{}{
				"order_status": "Cancelled",
				"cancel_date":  time.Now(),
			}).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order status ", "Something Went Wrong", "")
				return
			}
		}
		if paymentDetails.PaymentMethod == "Cash On Delivery" && paymentDetails.PaymentStatus == "Paid" {
			if err := tx.Model(&paymentDetails).Update("payment_status", "Refunded").Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
				return
			}
		} else if paymentDetails.PaymentMethod == "Cash On Delivery" && paymentDetails.PaymentStatus != "Paid" {
			if err := tx.Model(&paymentDetails).Update("payment_status", "Cancelled").Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update payment status ", "Something Went Wrong", "")
				return
			}
		}
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Order Status Updated Successfully",
		"code":    200,
	})
}

func ApproveReturn(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var input struct {
		ReturnRequestID string `json:"requestUID"`
		OrderID         string `json:"orderId"`
		Status          string `json:"action"`
		AdminNotes      string `json:"adminNotes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Something Went Wrong", "")
		return
	}
	tx := config.DB.Begin()
	ordid, err := strconv.ParseUint(input.OrderID, 10, 32) // 10 → Base 10, 32 → uint32 range
	if err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Something Went Wrong", "Something Went Wrong", "")
		return
	}

	var returnRequest models.ReturnRequest
	if err := tx.First(&returnRequest, "order_item_id = ? AND request_uid = ? AND user_id = ?", ordid, input.ReturnRequestID, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Return request not found", "Something Went Wrong", "")
		return
	}
	if returnRequest.Status != "Pending" {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusConflict, "Return request already processed", "Something Went Wrong", "")
		return
	}

	if input.Status == "Approved" {
		var orderItems models.OrderItem
		if err := tx.First(&orderItems, "id = ? AND user_id = ?", ordid, userID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Order Items Not Found", "Something Went Wrong", "")
			return
		}

		var order models.Order
		if err := tx.First(&order, "user_id = ? AND id = ?", userID, orderItems.OrderID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
			return
		}

		var refundAmount float64
		total := order.SubTotal
		productTotal := orderItems.SubTotal
		IscouponRemoved := false

		if order.IsCouponApplied {
			var couponDetails models.Coupon
			if err := tx.Unscoped().First(&couponDetails, order.CouponID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Coupon Not Found", "Something Went Wrong", "")
				return
			}

			if (total - productTotal) < couponDetails.MinOrderPrice {
				refundAmount = orderItems.Total - order.CouponDiscountAmount
				if refundAmount < 0 {
					tx.Rollback()
					helper.RespondWithError(c, http.StatusBadRequest, "Sorry, you cannot cancel this order individually.", "Something Went Wrong", "")
					return
				}
				couponDetails.UsersUsedCount -= 1
				if err := tx.Unscoped().Save(&couponDetails).Error; err != nil {
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
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&product).Where("id = ?", orderItems.ProductVariantID).
			Update("stock_quantity", product.StockQuantity+orderItems.Quantity).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Stock Reverse Failed", "Stock Update Failed", "/checkout")
			return
		}

		var payment models.PaymentDetail
		if err := tx.First(&payment, "order_item_id = ? AND user_id = ?", orderItems.ID, userID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Payment Details Not Found", "Something Went Wrong", "")
			return
		}
		if payment.PaymentStatus == "Completed" {
			var wallet models.Wallet
			if err := tx.First(&wallet, "user_id = ?", userID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "User Wallet Not Found", "Something Went Wrong", "")
				return
			}
			wallet.Balance += refundAmount
			if err := tx.Save(&wallet).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
				return
			}
			receiptID := "rcpt_" + uuid.New().String()
			transactionID := "TXN_" + uuid.New().String()

			walletTransaction := models.WalletTransaction{
				UserID:        userID,
				WalletID:      wallet.ID,
				Amount:        refundAmount,
				Description:   fmt.Sprintf("Order Refund ORD ID " + orderItems.OrderUID),
				Type:          "Refund",
				Receipt:       receiptID,
				OrderId:       order.OrderUID,
				TransactionID: transactionID,
				PaymentMethod: payment.PaymentMethod,
			}
			if err := tx.Create(&walletTransaction).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
				return
			}
			if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, orderItems.ID).
				Update("payment_status", "Refunded").Error; err != nil {
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
			if err := tx.Model(&order).Where("user_id = ? AND id = ?", userID, order.ID).
				Updates(map[string]interface{}{
					"coupon_code":            gorm.Expr("NULL"),
					"coupon_id":              gorm.Expr("NULL"),
					"shipping_charge":        shipCharge,
					"coupon_discount_amount": gorm.Expr("NULL"),
					"is_coupon_applied":      false,
				}).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
				return
			}
		} else {
			if err := tx.Model(&order).Where("user_id = ? AND id = ?", userID, order.ID).
				Updates(map[string]interface{}{
					"shipping_charge":        shipCharge,
					"coupon_discount_amount": gorm.Expr("NULL"),
				}).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
				return
			}
		}

		if err := tx.Model(&orderItems).Where("user_id = ? AND order_id = ?", userID, order.ID).
			Updates(map[string]interface{}{
				"order_status":           "Returned",
				"reason":                 returnRequest.Reason,
				"cancel_date":            time.Now(),
				"expected_delivery_date": time.Now(),
				"return_date":            time.Now(),
			}).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
			return
		}
	}

	if err := tx.Model(&returnRequest).Where("request_uid = ?", input.ReturnRequestID).Updates(map[string]interface{}{
		"status":      input.Status,
		"admin_notes": input.AdminNotes,
	}).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update status", "Something Went Wrong", "")
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Return request processed successfully.",
		"code":    http.StatusOK,
	})
}
