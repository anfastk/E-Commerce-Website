package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TrackingPage(c *gin.Context) {
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

	orderID := c.Param("id")

	var orderItem models.OrderItem
	if err := config.DB.Preload("ProductVariantDetails").
		Preload("ProductVariantDetails.VariantsImages").
		First(&orderItem, "id = ? AND user_id = ?", orderID, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
		return
	}

	var order models.Order
	if err := config.DB.First(&order, "user_id = ? AND id = ?", userID, orderItem.OrderID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Order Not found", "Something Went Wrong", "")
		return
	}

	var allOrderItems []models.OrderItem
	if err := config.DB.Preload("ProductVariantDetails").
		Preload("ProductVariantDetails.VariantsImages").
		Find(&allOrderItems, "id != ? AND user_id = ? AND order_id = ?", orderID, userID, order.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
		return
	}

	var shippingAddress models.ShippingAddress
	if err := config.DB.First(&shippingAddress, "user_id = ? AND order_id = ?", userID, order.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Address Not found", "Something Went Wrong", "")
		return
	}

	var payment models.PaymentDetail
	if err := config.DB.First(&payment, "user_id = ? AND order_item_id = ?", userID, orderItem.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Payment details Not found", "Something Went Wrong", "")
		return
	}
	isPaid := false
	if payment.PaymentStatus == "Paid" {
		isPaid = true
	}
	isDelivered := false
	if orderItem.OrderStatus == "Delivered" {
		isDelivered = true
	}
	productDiscount := (orderItem.ProductRegularPrice - orderItem.ProductSalePrice) * float64(orderItem.Quantity)
	totalDiscount := productDiscount + order.CouponDiscountAmount

	var (
		allSubTotal             float64
		allSalePrice            float64
		allProductDiscount      float64
		allProductTotalDiscount float64
		shipCharge              float64
	)
	for _, allItems := range allOrderItems {
		allSubTotal += allItems.ProductRegularPrice
		allSalePrice += allItems.ProductSalePrice
	}
	allSubTotal += orderItem.ProductRegularPrice
	allSalePrice += orderItem.ProductSalePrice
	allProductDiscount = (allSubTotal - allSalePrice)
	if order.ShippingCharge == 0.0 {
		shipCharge = 100
	}
	allProductTotalDiscount = allProductDiscount + order.ShippingCharge + shipCharge + order.CouponDiscountAmount

	IsCancelSpecificOrder := true
	cancelCheckAmount := order.SubTotal - orderItem.SubTotal
	if cancelCheckAmount-order.CouponDiscountAmount < 0 {
		IsCancelSpecificOrder = false
	}
	isAllOrderCancel := false
	count := 0
	var allItems []models.OrderItem
	if err := config.DB.
		Find(&allItems, " user_id = ? AND order_id = ?", userID, order.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
		return
	}
	for _, itm := range allItems {
		if itm.OrderStatus != "Cancelled" && itm.OrderStatus != "Failed" && itm.OrderStatus != "Returned" {
			count++
		}
		if itm.OrderStatus == "Delivered" || itm.OrderStatus == "Cancelled" || itm.OrderStatus == "Failed" || itm.OrderStatus == "Returned" {
			isAllOrderCancel = false
			break
		}
		if count >= 1 {
			isAllOrderCancel = true
		}
	}
	isAlreadyRequested := true
	var returnRequest models.ReturnRequest
	if err := config.DB.First(&returnRequest, "order_item_id = ? AND product_variant_id = ? AND user_id = ?", orderItem.ID, orderItem.ProductVariantID, userID).Error; err != nil {
		isAlreadyRequested = false
	}

	c.HTML(http.StatusOK, "trackOrder.html", gin.H{
		"status":                  "success",
		"message":                 "Order details fetched successfully",
		"Order":                   order,
		"Address":                 shippingAddress,
		"OrderItem":               orderItem,
		"Payment":                 payment,
		"IsPaid":                  isPaid,
		"IsDelivered":             isDelivered,
		"ProductDiscount":         productDiscount,
		"TotalDiscount":           totalDiscount,
		"isAllOrderCancel":        isAllOrderCancel,
		"isAlreadyRequested":      isAlreadyRequested,
		"IsCancelSpecificOrder":   IsCancelSpecificOrder,
		"OrderDate":               orderItem.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
		"ExpectedDeliveryDate":    orderItem.ExpectedDeliveryDate.Format("2006-01-02T15:04:05.000-07:00"),
		"ReturnDate":              orderItem.ReturnDate.Format("2006-01-02T15:04:05.000-07:00"),
		"ShippedDate":             orderItem.ShippedDate.Format("2006-01-02T15:04:05.000-07:00"),
		"OutOfDeliveryDate":       orderItem.OutOfDeliveryDate.Format("2006-01-02T15:04:05.000-07:00"),
		"DeliveryDate":            orderItem.DeliveryDate.Format("2006-01-02T15:04:05.000-07:00"),
		"CancelDate":              orderItem.CancelDate.Format("2006-01-02T15:04:05.000-07:00"),
		"AllProduct":              allOrderItems,
		"AllSubTotal":             allSubTotal,
		"AllProductDiscount":      allProductDiscount,
		"AllProductTotalDiscount": allProductTotalDiscount,
	})
}

func CreateOrder(c *gin.Context, tx *gorm.DB, userID uint, subTotal float64, totalProductDiscount float64, totalDiscount float64, tax float64, shippingCharge float64, totalAmount float64, currentTime time.Time, CouponCode string, CouponDiscountAmount float64, CouponDiscription string) uint {
	orderUID := helper.GenerateOrderID()
	var order models.Order
	order = models.Order{
		OrderUID:             orderUID,
		UserID:               userID,
		SubTotal:             subTotal,
		TotalProductDiscount: totalProductDiscount,
		TotalDiscount:        totalDiscount,
		Tax:                  tax,
		ShippingCharge:       shippingCharge,
		TotalAmount:          totalAmount,
		OrderDate:            currentTime,
		CouponCode:           CouponCode,
		CouponDiscountAmount: CouponDiscountAmount,
		CouponDiscription:    CouponDiscription,
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
		return 0
	}
	if CouponCode != "" && CouponDiscountAmount != 0 && CouponDiscription != "" {
		order.IsCouponApplied = true
		if err := tx.Save(&order).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Something Went Wrong", "/checkout")
			return 0
		}
	}
	return order.ID
}
func CreateOrderItems(c *gin.Context, tx *gorm.DB, reservedProducts []models.ReservedStock, shippingCharge float64, orderID uint, userID uint, currentTime time.Time, couponDiscount float64) {
	for _, item := range reservedProducts {
		orderUID := helper.GenerateOrderID()
		discountAmount, _, _ := helper.DiscountCalculation(item.ProductVariant.ProductID, item.ProductVariant.CategoryID, item.ProductVariant.RegularPrice, item.ProductVariant.SalePrice)
		regularPrice := item.ProductVariant.RegularPrice * float64(item.Quantity)
		salePrice := (item.ProductVariant.SalePrice - discountAmount) * float64(item.Quantity)
		tax := (salePrice * 18) / 100
		if salePrice > 1000 {
			shippingCharge = 0
		}
		total := salePrice + tax + shippingCharge - couponDiscount
		var firstImage string

		var firstVariantImage models.ProductVariantsImage
		err := config.DB.Unscoped().Where("product_variant_id = ?", item.ProductVariant.ID).Order("id ASC").First(&firstVariantImage).Error

		if err == nil {
			firstImage = firstVariantImage.ProductVariantsImages
		}
		var mainProduct models.ProductDetail
		if err := tx.Unscoped().First(&mainProduct, "id = ?", item.ProductVariant.ProductID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch main product", "Something Went Wrong", "/checkout")
			return
		}
		cID := mainProduct.CategoryID
		var category models.Categories
		if err := tx.First(&category, "id = ?", cID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch category", "Something Went Wrong", "/checkout")
			return
		}

		orderItems := models.OrderItem{
			OrderID:              orderID,
			UserID:               userID,
			OrderUID:             orderUID,
			ProductName:          item.ProductVariant.ProductName,
			ProductSummary:       item.ProductVariant.ProductSummary,
			ProductCategory:      category.Name,
			ProductImage:         firstImage,
			ProductRegularPrice:  item.ProductVariant.RegularPrice,
			ProductSalePrice:     item.ProductVariant.SalePrice - discountAmount,
			ProductVariantID:     item.ProductVariantID,
			Quantity:             item.Quantity,
			SubTotal:             regularPrice,
			Tax:                  tax,
			Total:                total,
			OrderStatus:          "Pending",
			ExpectedDeliveryDate: currentTime.AddDate(0, 0, 7),
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
			return
		}
	}
}
func SaveOrderAddress(c *gin.Context, tx *gorm.DB, orderID uint, userID uint, addressID string) {
	address := FetchAddressByIDAndUserID(c, userID, paymentRequest.AddressID)
	var shippingAddress models.ShippingAddress
	shippingAddress = models.ShippingAddress{
		UserID:    address.UserID,
		OrderID:   orderID,
		FirstName: address.FirstName,
		LastName:  address.LastName,
		Mobile:    address.Mobile,
		Address:   address.Address,
		Landmark:  address.Landmark,
		Country:   address.Country,
		State:     address.State,
		City:      address.City,
		PinCode:   address.PinCode,
	}
	if err := tx.Create(&shippingAddress).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create address", "Something Went Wrong", "/checkout")
		return
	}
}

func ClearCart(c *gin.Context, tx *gorm.DB, ordered map[uint]int) {
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

	var cart models.Cart

	if err := tx.First(&cart, "user_ID = ?", userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart not found", "Cart not found", "/cart")
		return
	}

	for pId, _ := range ordered {
		var cartItems models.CartItem
		if err := tx.First(&cartItems, "cart_id = ? AND product_variant_id = ?", cart.ID, pId).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusBadRequest, "Product Not Found In Cart", "Something Went Wrong", "/cart")
			return
		}
		if err := tx.Unscoped().Where("product_variant_id = ? AND cart_id = ?", pId, cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Something Went Wrong", "/cart")
			return
		}

	}
}

func ShowSuccessPage(c *gin.Context) {
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

	today := time.Now().Format("2006-01-02")
	var order models.Order
	if err := config.DB.First(&order, "user_id = ? AND DATE(created_at) = ?", userID, today).Error; err != nil {
		c.HTML(http.StatusForbidden, "error.html", gin.H{"message": "No recent order found"})
		return
	}

	c.HTML(http.StatusOK, "orderSuccess.html", gin.H{
		"status":  "OK",
		"message": "Payment processed",
	})
}

func FetchOrderItems(c *gin.Context, tx *gorm.DB, orderID uint) []models.OrderItem {
	var orderItems []models.OrderItem
	if err := tx.Find(&orderItems, "order_id = ?", orderID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "/profile/order/details")
		return nil
	}
	return orderItems
}
func FetchCartByUserID(c *gin.Context, userID uint) *models.Cart {
	var cart models.Cart

	if err := config.DB.First(&cart, "user_ID = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Cart not found", "Cart not found", "/")
		return nil
	}
	return &cart
}

func FetchCartItemByCartID(c *gin.Context, cartID uint) []models.CartItem {
	var cartItems []models.CartItem
	if err := config.DB.Find(&cartItems, "cart_id = ?", cartID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to fetch cart items", "/cart")
		return nil
	}
	return cartItems
}

func FetchAddressByIDAndUserID(c *gin.Context, userID uint, addressID string) *models.UserAddress {
	addressId, adErr := strconv.Atoi(addressID)
	if adErr != nil {
		fmt.Println("Invalid Address:", adErr)
		return nil
	}
	var address models.UserAddress
	if err := config.DB.First(&address, "user_id = ? AND id = ?", userID, addressId).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Address not found", "Something Went Wrong", "/checkout")
		return nil
	}
	return &address
}

func FetchReservedProducts(c *gin.Context, userID uint) []models.ReservedStock {
	var reservedProducts []models.ReservedStock
	if err := config.DB.Unscoped().Preload("ProductVariant").
		Find(&reservedProducts, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to fetch cart items", "/cart")
		return nil
	}
	return reservedProducts
}

func ReservedProductCheck(c *gin.Context, reservedProducts []models.ReservedStock, cartItems []models.CartItem) (map[uint]int, float64, float64, float64, float64, float64, float64) {
	shippingCharge := 100
	var (
		regularPrice    float64
		salePrice       float64
		tax             float64
		productDiscount float64
		totalDiscount   float64
		total           float64
	)
	reservedMap := make(map[uint]int)
	for _, r := range reservedProducts {
		discountAmount, _, _ := helper.DiscountCalculation(r.ProductVariantID, r.ProductVariant.CategoryID, r.ProductVariant.RegularPrice, r.ProductVariant.SalePrice)
		reservedMap[r.ProductVariantID] = r.Quantity
		regularPrice += r.ProductVariant.RegularPrice * float64(r.Quantity)
		salePrice += (r.ProductVariant.SalePrice * float64(r.Quantity)) - discountAmount
	}

	for _, item := range cartItems {
		reservedQty, exists := reservedMap[item.ProductID]
		if exists {
			if item.Quantity != reservedQty {
				helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 1", "Something Went Wrong", "/cart")
				return nil, 0, 0, 0, 0, 0, 0
			}
		} else {
			helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 2", "Something Went Wrong", "/cart")
			return nil, 0, 0, 0, 0, 0, 0
		}
	}
	for pID, _ := range reservedMap {
		found := false
		for _, items := range cartItems {
			if items.ProductID == pID {
				found = true
				break
			}
		}
		if !found {
			helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 3", "Something Went Wrong", "/cart")
			return nil, 0, 0, 0, 0, 0, 0
		}
	}
	tax = (salePrice * 18) / 100
	productDiscount = regularPrice - salePrice
	if salePrice > 1000 {
		shippingCharge = 0
	}
	totalDiscount = productDiscount + float64(shippingCharge)
	if shippingCharge == 0 {
		totalDiscount = productDiscount + 100
	}
	total = salePrice + tax + float64(shippingCharge)
	return reservedMap, regularPrice, productDiscount, totalDiscount, float64(shippingCharge), tax, total
}

func DeleteReservedItems(c *gin.Context, tx *gorm.DB, productVariantID uint, userID uint) {
	if err := tx.Unscoped().Where("product_variant_id = ? AND user_id = ?", productVariantID, userID).Delete(&models.ReservedStock{}).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete reservations", "Something Went Wrong", "/cart")
		return
	}
}

func CancelSpecificOrder(c *gin.Context) {
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

	orderItemID := c.Param("id")
	var inputReason struct {
		Reason string ` json:"cancelReason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&inputReason); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Reason Not Found", "Reason is required", "")
		return
	}

	tx := config.DB.Begin()

	var orderItems models.OrderItem
	if err := tx.First(&orderItems, "id = ? AND user_id = ?", orderItemID, userID).Error; err != nil {
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
	if orderItems.OrderStatus == "Returned" {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "You cannot cancel this order as it has already been returned.", "You cannot cancel this order as it has already been returned.", "")
		return
	}

	if orderItems.OrderStatus == "Delivered" {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "You cannot cancel this order as it has already been delivered.", "You cannot cancel this order as it has already been delivered.", "")
		return
	}

	if orderItems.OrderStatus == "Cancelled" {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "You have already cancelled this order.", "You have already cancelled this order.", "")
		return
	}

	if orderItems.OrderStatus == "Failed" {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Sorry, you cannot cancel this order as it has already failed.", "Sorry, you cannot cancel this order as it has already failed.", "")
		return
	}
	var refundAmount float64
	total := order.SubTotal
	productTotal := orderItems.SubTotal
	IscouponRemoved := false

	if order.IsCouponApplied {
		var couponDetails models.Coupon
		if err := tx.Unscoped().First(&couponDetails, "coupon_code = ?", order.CouponCode).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Coupon Not Found", "Something Went Wrong", "")
			return
		}

		if (total - productTotal) < couponDetails.MinOrderValue {
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
		lastBalance := wallet.Balance
		wallet.Balance += refundAmount
		if err := tx.Save(&wallet).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
			return
		}
		receiptID := "rcpt_" + uuid.New().String()
		transactionID := "TXN-" + uuid.New().String()

		walletTransaction := models.WalletTransaction{
			UserID:        userID,
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
	if payment.PaymentStatus == "Pending" || payment.PaymentStatus == "Failed" {
		if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, orderItems.ID).
			Update("payment_status", "Cancelled").Error; err != nil {
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
			"order_status":           "Cancelled",
			"reason":                 inputReason.Reason,
			"cancel_date":            time.Now(),
			"expected_delivery_date": time.Now(),
			"return_date":            time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order cancelled successfully",
		"code":    http.StatusOK,
	})
}

func CancelAllOrderItems(c *gin.Context) {
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

	orderID := c.Param("id")
	var inputReason struct {
		Reason string `form:"cancelAllReason" json:"cancelAllReason" binding:"required"`
	}
	if err := c.ShouldBind(&inputReason); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Reason Not Found", "Reason is required", "")
		return
	}

	tx := config.DB.Begin()

	var order models.Order
	if err := tx.First(&order, "user_id = ? AND id = ?", userID, orderID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
		return
	}

	var orderItems []models.OrderItem

	if err := tx.Find(&orderItems, "user_id = ? AND order_id = ?", userID, order.ID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
		return
	}
	IsPaymentCompleted := false
	var paymentMethod string
	var subTotal float64
	for _, itm := range orderItems {
		subTotal += itm.SubTotal
		if itm.OrderStatus == "Delivered" || itm.OrderStatus == "Cancelled" || itm.OrderStatus == "Failed" || itm.OrderStatus == "Returned" {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
			return
		}
		var product models.ProductVariantDetails
		if err := tx.First(&product, itm.ProductVariantID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&product).Where("id = ?", itm.ProductVariantID).
			Update("stock_quantity", product.StockQuantity+itm.Quantity).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Stock Reverse Failed", "Stock Update Failed", "/checkout")
			return
		}

		var payment models.PaymentDetail
		if err := tx.First(&payment, "order_item_id = ? AND user_id = ?", itm.ID, userID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Payment Details Not Found", "Something Went Wrong", "")
			return
		}
		paymentMethod = payment.PaymentMethod
		if payment.PaymentStatus == "Completed" {
			IsPaymentCompleted = true
			if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, itm.ID).
				Update("payment_status", "Refunded").Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment Status Update Failed", "Payment Status Update Failed", "/checkout")
				return
			}
		}
		if payment.PaymentStatus == "Pending" || payment.PaymentStatus == "Failed" {
			if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, itm.ID).
				Update("payment_status", "Cancelled").Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment Status Update Failed", "Payment Status Update Failed", "/checkout")
				return
			}
		}

	}

	if IsPaymentCompleted {
		refundAmount := order.TotalAmount + order.ShippingCharge
		var wallet models.Wallet
		if err := tx.First(&wallet, "user_id = ?", userID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "User Wallet Not Found", "Something Went Wrong", "")
			return
		}
		lastBalance := wallet.Balance
		wallet.Balance += refundAmount
		if err := tx.Save(&wallet).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
			return
		}
		receiptID := "rcpt_" + uuid.New().String()
		transactionID := "TXN-" + uuid.New().String()

		walletTransaction := models.WalletTransaction{
			UserID:        userID,
			WalletID:      wallet.ID,
			Amount:        refundAmount,
			Description:   fmt.Sprintf("Order Refund ORD ID " + order.OrderUID),
			Type:          "Refund",
			Receipt:       receiptID,
			OrderId:       order.OrderUID,
			LastBalance:   lastBalance,
			TransactionID: strings.ToUpper(transactionID),
			PaymentMethod: paymentMethod,
		}
		if err := tx.Create(&walletTransaction).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
			return
		}
	}
	if order.IsCouponApplied {
		var couponDetails models.Coupon
		if err := tx.Unscoped().First(&couponDetails, "coupon_code = ?", order.CouponCode).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Coupon Not Found", "Something Went Wrong", "")
			return
		}
		couponDetails.UsersUsedCount -= 1
		if err := tx.Unscoped().Save(&couponDetails).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Coupon", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&order).Where("user_id = ? AND id = ?", userID, order.ID).
			Updates(map[string]interface{}{
				"coupon_code":            gorm.Expr("NULL"),
				"coupon_id":              gorm.Expr("NULL"),
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
				"coupon_discount_amount": gorm.Expr("NULL"),
			}).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
			return
		}
	}

	if err := tx.Model(&orderItems).Where("user_id = ? AND order_id = ?", userID, order.ID).
		Updates(map[string]interface{}{
			"order_status":           "Cancelled",
			"reason":                 inputReason.Reason,
			"cancel_date":            time.Now(),
			"expected_delivery_date": time.Now(),
			"return_date":            time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order cancelled successfully",
		"code":    http.StatusOK,
	})
}

func ReturnOrder(c *gin.Context) {
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

	var input struct {
		Reason            string `json:"reason" binding:"required"`
		AdditionalDetails string `json:"additionalDetails" binding:"required"`
		ProductId         string `json:"productId" binding:"required"`
		OrderId           string `json:"orderId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Enter All Fields", "")
		return
	}

	ordId, err := strconv.ParseUint(input.OrderId, 10, 32)
	prdtId, err := strconv.ParseUint(input.ProductId, 10, 32)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Something Went Wrong", "")
		return
	}
	reqstId := "RTN-" + uuid.New().String()
	returnRequest := models.ReturnRequest{
		RequestUID:        reqstId,
		OrderItemID:       uint(ordId),
		ProductVariantID:  uint(prdtId),
		UserID:            userID,
		Reason:            input.Reason,
		AdditionalDetails: input.AdditionalDetails,
		Status:            "Pending",
	}
	if err := config.DB.Create(&returnRequest).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to submit return request", "Something Went Wrong", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Return request submitted successfully. Awaiting admin approval.",
		"code":    http.StatusOK,
	})

}
