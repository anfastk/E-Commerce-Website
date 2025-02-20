package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func PaymentPage(c *gin.Context) {
	userID := c.MustGet("userid")
	var request struct {
		AddressID string `json:"addressId"`
	}

	if err := c.ShouldBind(&request); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Binding the data", "")
		return
	}

	var address models.UserAddress
	if err := config.DB.First(&address, "id = ? AND user_id = ?", request.AddressID, userID).Error; err != nil {
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
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "", "Something Went Wrong", "/checkout")
		return
	}
	var cartItems []models.CartItem
	if err := config.DB.Preload("ProductVariant").
		Preload("ProductVariant.VariantsImages").
		Preload("ProductVariant.Category").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
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

func ProceedToPayment(c *gin.Context) {
	userID := c.MustGet("userid")
	var (
		regularPrice float64
		salePrice    float64
		tax          float64
		total        float64
	)
	shippingCharge := 100
	var paymentRequest struct {
		PaymentMethod string `json:"paymentMethod"`
		AddressID     string `json:"addressId"`
	}

	if err := c.ShouldBind(&paymentRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Request Not Found", "")
		return
	}
	tx := config.DB.Begin()
	var userDetails models.UserAuth

	if err := tx.First(&userDetails, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/cart")
		return
	}

	var cart models.Cart

	if err := tx.First(&cart, "user_ID = ?", userDetails.ID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart not found", "Cart not found", "/cart")
		return
	}
	var cartItems []models.CartItem
	if err := tx.Preload("ProductVariant").
		Find(&cartItems, "cart_id = ?", cart.ID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to fetch cart items", "/cart")
		return
	}
	if len(cartItems) == 0 {
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Add Product in Your Cart", "/cart")
		return
	}

	for _, items := range cartItems {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", items.ProductVariantID).
			First(&items.ProductVariant).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusConflict, "Stock unavailable", "Some items are out of stock. Please update your cart.", "/cart")
			return
		}
		if items.ProductVariant.StockQuantity < items.Quantity {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusConflict, "Stock unavailable", "Some items are out of stock. Please update your cart.", "/cart")
			return
		}
		items.ProductVariant.StockQuantity -= items.Quantity
		tx.Save(&items.ProductVariant)
		salePrice += items.ProductVariant.SalePrice * float64(items.Quantity)
	}
	tax = (salePrice * 18) / 100

	if salePrice > 1000 {
		shippingCharge = 0
	}
	total = salePrice + tax
	currentTime := time.Now()
	var order models.Order
	order = models.Order{
		UserID:         userDetails.ID,
		OrderAmount:    total,
		ShippingCharge: float64(shippingCharge),
		Tax:            tax,
		OrderDate:      currentTime,
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
		return
	}
	var address models.UserAddress
	if err := tx.First(&address, "user_id = ? AND id = ?", userDetails.ID, paymentRequest.AddressID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Address not found", "Something Went Wrong", "/checkout")
		return
	}
	var shippingAddress models.ShippingAddress
	shippingAddress = models.ShippingAddress{
		UserID:    address.UserID,
		OrderID:   order.ID,
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

	for _, item := range cartItems {
		orderUID := generateOrderID()
		regularPrice = item.ProductVariant.RegularPrice * float64(item.Quantity)
		salePrice = item.ProductVariant.SalePrice * float64(item.Quantity)
		tax = (salePrice * 18) / 100
		if salePrice > 1000 {
			shippingCharge = 0
		}
		total = salePrice + tax
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
			OrderID:              order.ID,
			UserID:               userDetails.ID,
			OrderUID:             orderUID,
			ProductName:          item.ProductVariant.ProductName,
			ProductSummary:       item.ProductVariant.ProductSummary,
			ProductCategory:      category.Name,
			/* CouponDiscount: , */
			ProductImage:         firstImage,
			ProductRegularPrice:  item.ProductVariant.RegularPrice,
			ProductSalePrice:     item.ProductVariant.SalePrice,
			ProductVariantID:     item.ProductVariantID,
			Quantity:             item.Quantity,
			SubTotal:             regularPrice,
			ShippingCharge:       float64(shippingCharge),
			Tax:                  tax,
			Total:                total,
			ExpectedDeliveryDate: currentTime.AddDate(0, 0, 7),
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
			return
		}
		if err := tx.First(&orderItems, "id = ? AND order_id=? AND product_variant_id = ?", orderItems.ID, order.ID, item.ProductVariantID).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "/profile/order/details")
			return
		}
		paymentStatus := false
		switch paymentRequest.PaymentMethod {
		case "COD":
			if err := CODPayment(c, tx, userDetails.ID, orderItems.ID, orderItems.Total, "Cash On Delivery"); err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment failed", "Something Went Wrong", "/checkout")
				return
			}
			paymentStatus = true
			/* 	case "Wallet":
			   	case "Others": */
		default:
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Payment Method", "Invalid Payment Method", "/checkout")
			return
		}
		if paymentStatus {
			for _, item := range cartItems {
				if err := tx.Unscoped().Where("id = ?", item.ID).Delete(&models.CartItem{}).Error; err != nil {
					tx.Rollback()
					helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to delete cart items", "/cart")
					return
				}
			}

		}
	}

	tx.Commit()
	c.HTML(http.StatusOK, "orderSuccess.html", gin.H{
		"status":  "OK",
		"message": "Payment processed",
		"code":    http.StatusOK,
	})
}

func generateOrderID() string {
	rand.Seed(time.Now().UnixNano())
	section1 := rand.Intn(900) + 100
	section2 := rand.Intn(9000000) + 1000000
	section3 := time.Now().UnixNano() % 10000000

	return fmt.Sprintf("#%d-%d-%07d", section1, section2, section3)
}

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
