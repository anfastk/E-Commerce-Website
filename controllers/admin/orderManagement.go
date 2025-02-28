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
	totalDiscount := productDiscount + orderItemDetails.CouponDiscount
	if orderItemDetails.ShippingCharge == 0 {
		totalDiscount += 100
	}
	c.HTML(http.StatusOK, "orderDetailsManagement.html", gin.H{
		"status":          "success",
		"message":         "Order details fetched successfully",
		"OrderItem":       orderItemDetails,
		"OrderDate":       orderItemDetails.CreatedAt.Format("02 Jan 2006"),
		"Order":           orderDetails,
		"User":            UserDetails,
		"Payment":         paymentDetails,
		"Address":         shippingAddress,
		"ProductDiscount": productDiscount,
		"TotalDiscount":   totalDiscount,
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
		if err := tx.Model(&paymentDetails).Update("payment_status", "Paid").Error; err != nil {
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
