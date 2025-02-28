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
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PaymentPage(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var request struct {
		AddressID string `json:"addressId"`
	}

	tx := config.DB.Begin()
	var expiredReservations []models.ReservedStock
	err := tx.Where("is_confirmed = ? AND reserve_till >= ?", false, time.Now().Add(time.Duration(25)*time.Millisecond)).
		Find(&expiredReservations).Error
	if err != nil {
		tx.Rollback()
		log.Println("Error finding expired reservations:", err)
		return
	}
	for _, reservation := range expiredReservations {
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

	if err := c.ShouldBind(&request); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Binding the data", "")
		return
	}

	var address models.UserAddress
	if err := tx.First(&address, "id = ? AND user_id = ?", request.AddressID, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Address not found", "Address not found", "")
	}

	var (
		regularPrice    float64
		salePrice       float64
		productDiscount float64
		totalDiscount   float64
		tax             float64
		total           float64
	)
	shippingCharge := 100
	var cart models.Cart
	if err := tx.First(&cart, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "", "Something Went Wrong", "/checkout")
		return
	}
	var cartItems []models.CartItem
	if err := tx.Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "", "Product Not Found", "/checkout")
		return
	}

	if len(cartItems) == 0 {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/checkout")
		return
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
		}
		if err := tx.Create(&reserveStock).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to store reserved stock", "Something Went Wrong", "/checkout")
			return
		}
	}

	for _, items := range cartItems {
		regularPrice += items.ProductVariant.RegularPrice * float64(items.Quantity)
		salePrice += items.ProductVariant.SalePrice * float64(items.Quantity)
	}
	tax = (salePrice * 18) / 100
	productDiscount = regularPrice - salePrice

	if salePrice < 1000 {
		totalDiscount = productDiscount
	} else {
		totalDiscount = productDiscount + float64(shippingCharge)
		shippingCharge = 0
	}
	total = salePrice + tax

	tx.Commit()

	c.HTML(http.StatusOK, "paymentPage.html", gin.H{
		"status":          "OK",
		"message":         "Checkout fetch success",
		"Address":         address,
		"CartItem":        cartItems,
		"SubTotal":        regularPrice,
		"Shipping":        shippingCharge,
		"Tax":             tax,
		"ProductDiscount": productDiscount,
		"TotalDiscount":   totalDiscount,
		"Total":           total,
		"code":            http.StatusOK,
	})
}

var paymentRequest struct {
	PaymentMethod string `json:"paymentMethod"`
	AddressID     string `json:"addressId"`
}

func ProceedToPayment(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

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
	reservedMap, shippingCharge, tax, total := ReservedProductCheck(c, reservedProducts, cartItems)

	switch paymentRequest.PaymentMethod {
	case "COD":
		paymentStatus := true
		tx := config.DB.Begin()
		orderID := CreateOrder(c, tx, userDetails.ID, total, tax, float64(shippingCharge), currentTime)
		SaveOrderAddress(c, tx, orderID, userDetails.ID, paymentRequest.AddressID)
		CreateOrderItems(c, tx, reservedProducts, float64(shippingCharge), orderID, userDetails.ID, currentTime)
		orderItems := FetchOrderItems(c, tx, orderID)

		for _, ordItems := range orderItems {
			if err := CODPayment(c, tx, userDetails.ID, ordItems.ID, ordItems.Total, "Cash On Delivery"); err != nil {
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

		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"message":  "Payment processed",
			"redirect": "/order/success",
			"code":     http.StatusOK,
		})

	case "Razorpay":
		address := FetchAddressByIDAndUserID(c, userID, paymentRequest.AddressID)
		razorpayOrder, err := CreateRazorpayOrder(c, total)
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
			"amount":   total * 100, 
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/checkout")
		return
	}
}

func PayNow(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

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

		razorpayOrder, err := CreateRazorpayOrder(c, order.OrderTotalAmount)
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
		fmt.Println(address.Mobile, address.Address)
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"order_id": razorpayOrderID,
			"amount":   order.OrderTotalAmount * 100, 
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

func generateOrderID() string {
	rand.Seed(time.Now().UnixNano())
	section1 := rand.Intn(900) + 100
	section2 := rand.Intn(9000000) + 1000000
	section3 := time.Now().UnixNano() % 10000000

	return fmt.Sprintf("#%d-%d-%07d", section1, section2, section3)
}

func ShowSuccessPage(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

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

func ClearCart(c *gin.Context, tx *gorm.DB, ordered map[uint]int) {
	userID := c.MustGet("userid").(uint)

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

func CreateOrder(c *gin.Context, tx *gorm.DB, userID uint, total float64, tax float64, shippingCharge float64, currentTime time.Time) uint {
	var order models.Order
	order = models.Order{
		UserID:           userID,
		OrderTotalAmount: total,
		ShippingCharge:   shippingCharge,
		Tax:              tax,
		OrderDate:        currentTime,
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
		return 0
	}
	return order.ID
}
func CreateOrderItems(c *gin.Context, tx *gorm.DB, reservedProducts []models.ReservedStock, shippingCharge float64, orderID uint, userID uint, currentTime time.Time) {
	for _, item := range reservedProducts {
		orderUID := generateOrderID()
		regularPrice := item.ProductVariant.RegularPrice * float64(item.Quantity)
		salePrice := item.ProductVariant.SalePrice * float64(item.Quantity)
		tax := (salePrice * 18) / 100
		if salePrice > 1000 {
			shippingCharge = 0
		}
		total := salePrice + tax + shippingCharge
		var firstImage string

		var firstVariantImage models.ProductVariantsImage
		err := config.DB.Where("product_variant_id = ?", item.ProductVariant.ID).Order("id ASC").First(&firstVariantImage).Error

		if err == nil {
			firstImage = firstVariantImage.ProductVariantsImages
		}
		var mainProduct models.ProductDetail
		if err := tx.First(&mainProduct, "id = ?", item.ProductVariant.ProductID).Error; err != nil {
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
			OrderID:         orderID,
			UserID:          userID,
			OrderUID:        orderUID,
			ProductName:     item.ProductVariant.ProductName,
			ProductSummary:  item.ProductVariant.ProductSummary,
			ProductCategory: category.Name,
			/* CouponDiscount: , */
			ProductImage:         firstImage,
			ProductRegularPrice:  item.ProductVariant.RegularPrice,
			ProductSalePrice:     item.ProductVariant.SalePrice,
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
	if err := config.DB.Preload("ProductVariant").
		Find(&reservedProducts, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to fetch cart items", "/cart")
		return nil
	}
	return reservedProducts
}

func ReservedProductCheck(c *gin.Context, reservedProducts []models.ReservedStock, cartItems []models.CartItem) (map[uint]int, float64, float64, float64) {
	shippingCharge := 100
	var (
		salePrice float64
		tax       float64
		total     float64
	)
	reservedMap := make(map[uint]int)
	for _, r := range reservedProducts {
		reservedMap[r.ProductVariantID] = r.Quantity
		salePrice += r.ProductVariant.SalePrice * float64(r.Quantity)
	}

	for _, item := range cartItems {
		reservedQty, exists := reservedMap[item.ProductID]
		if exists {
			if item.Quantity != reservedQty {
				helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 1", "Something Went Wrong", "/cart")
				return nil, 0, 0, 0
			}
		} else {
			helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 2", "Something Went Wrong", "/cart")
			return nil, 0, 0, 0
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
			return nil, 0, 0, 0
		}
	}
	tax = (salePrice * 18) / 100

	if salePrice > 1000 {
		shippingCharge = 0
	}
	total = salePrice + tax + float64(shippingCharge)
	return reservedMap, float64(shippingCharge), tax, total
}

func DeleteReservedItems(c *gin.Context, tx *gorm.DB, productVariantID uint, userID uint) {
	if err := tx.Unscoped().Where("product_variant_id = ? AND user_id = ?", productVariantID, userID).Delete(&models.ReservedStock{}).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete reservations", "Something Went Wrong", "/cart")
		return
	}
}
