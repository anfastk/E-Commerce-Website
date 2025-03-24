package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TrackingPage(c *gin.Context) {
	logger.Log.Info("Requested tracking page")

	userID := helper.FetchUserID(c)
	orderID := c.Param("id")
	logger.Log.Debug("Fetched user ID and order ID", zap.Uint("userID", userID), zap.String("orderID", orderID))

	var orderItem models.OrderItem
	if err := config.DB.
		First(&orderItem, "id = ? AND user_id = ?", orderID, userID).Error; err != nil {
		logger.Log.Error("Failed to fetch order item",
			zap.String("orderID", orderID),
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
		return
	}

	var order models.Order
	if err := config.DB.First(&order, "user_id = ? AND id = ?", userID, orderItem.OrderID).Error; err != nil {
		logger.Log.Error("Failed to fetch order",
			zap.Uint("orderID", orderItem.OrderID),
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Order Not found", "Something Went Wrong", "")
		return
	}

	var allOrderItems []models.OrderItem
	if err := config.DB.Preload("ProductVariantDetails").
		Preload("ProductVariantDetails.VariantsImages").
		Find(&allOrderItems, "id != ? AND user_id = ? AND order_id = ?", orderID, userID, order.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch all order items",
			zap.String("orderID", orderID),
			zap.Uint("userID", userID),
			zap.Uint("orderID", order.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
		return
	}

	var shippingAddress models.ShippingAddress
	if err := config.DB.First(&shippingAddress, "user_id = ? AND order_id = ?", userID, order.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch shipping address",
			zap.Uint("userID", userID),
			zap.Uint("orderID", order.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Address Not found", "Something Went Wrong", "")
		return
	}

	var payment models.PaymentDetail
	if err := config.DB.First(&payment, "user_id = ? AND order_item_id = ?", userID, orderItem.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch payment details",
			zap.Uint("userID", userID),
			zap.Uint("orderItemID", orderItem.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Payment details Not found", "Something Went Wrong", "")
		return
	}

	isPaid := payment.PaymentStatus == "Paid"
	isDelivered := orderItem.OrderStatus == "Delivered"
	productDiscount := (orderItem.ProductRegularPrice - orderItem.ProductSalePrice) * float64(orderItem.Quantity)
	totalDiscount := productDiscount + order.CouponDiscountAmount

	var (
		allSubTotal             float64
		allProductDiscount      float64
		allProductTotalDiscount float64
		shipCharge              float64
	)

	allSubTotal += order.SubTotal
	allProductDiscount = order.TotalProductDiscount
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
	if err := config.DB.Find(&allItems, "user_id = ? AND order_id = ?", userID, order.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch all items for cancel check",
			zap.Uint("userID", userID),
			zap.Uint("orderID", order.ID),
			zap.Error(err))
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

	currentTime := time.Now()
	daysSinceDelivery := currentTime.Sub(orderItem.DeliveryDate).Hours() / 24
	if daysSinceDelivery > 7 {
		orderItem.ReturnableStatus = false
		if err := config.DB.Save(&orderItem).Error; err != nil {
			logger.Log.Error("Failed to update order item return status",
				zap.Uint("orderItemID", orderItem.ID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Updated returnable status to false", zap.Uint("orderItemID", orderItem.ID))
	}

	logger.Log.Info("Tracking page loaded successfully",
		zap.Uint("userID", userID),
		zap.String("orderID", orderID),
		zap.Bool("isPaid", isPaid),
		zap.Bool("isDelivered", isDelivered))
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
		"IsReturnAvailable":       orderItem.ProductVariantDetails.Product.IsReturnable && orderItem.ReturnableStatus,
		"isAllOrderCancel":        isAllOrderCancel,
		"isAlreadyRequested":      isAlreadyRequested,
		"IsCancelSpecificOrder":   IsCancelSpecificOrder,
		"PaymentDate":             payment.CreatedAt.Format("January 02, 2006 at 03:04 PM"),
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
	logger.Log.Info("Creating new order", zap.Uint("userID", userID))
	IsCouponApplied := false
	if CouponDiscountAmount > 0 {
		IsCouponApplied = true
	}

	orderUID := helper.GenerateOrderID()
	order := models.Order{
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
		IsCouponApplied:      IsCouponApplied,
		CouponDiscountAmount: CouponDiscountAmount,
		CouponDiscription:    CouponDiscription,
	}

	if err := tx.Create(&order).Error; err != nil {
		logger.Log.Error("Failed to create order",
			zap.Uint("userID", userID),
			zap.String("orderUID", orderUID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
		return 0
	}

	if CouponCode != "" && CouponDiscountAmount != 0 && CouponDiscription != "" {
		order.IsCouponApplied = true
		if err := tx.Save(&order).Error; err != nil {
			logger.Log.Error("Failed to update order with coupon",
				zap.Uint("orderID", order.ID),
				zap.String("couponCode", CouponCode),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update order", "Something Went Wrong", "/checkout")
			return 0
		}
		logger.Log.Info("Coupon applied to order",
			zap.Uint("orderID", order.ID),
			zap.String("couponCode", CouponCode))
	}

	logger.Log.Info("Order created successfully",
		zap.Uint("orderID", order.ID),
		zap.String("orderUID", orderUID))
	return order.ID
}

func CreateOrderItems(c *gin.Context, tx *gorm.DB, reservedProducts []models.ReservedStock, shippingCharge float64, orderID uint, userID uint, currentTime time.Time, couponDiscount float64) {
	logger.Log.Info("Creating order items", zap.Uint("orderID", orderID))

	for _, item := range reservedProducts {
		orderUID := helper.GenerateOrderID()
		discountAmount, _, _ := helper.DiscountCalculation(item.ProductVariant.ProductID, item.ProductVariant.CategoryID, item.ProductVariant.RegularPrice, item.ProductVariant.SalePrice)
		regularPrice := item.ProductVariant.RegularPrice * float64(item.Quantity)
		salePrice := (item.ProductVariant.SalePrice - discountAmount) * float64(item.Quantity)
		tax := (salePrice * 18) / 100
		if salePrice > 1000 {
			shippingCharge = 0
		}
		total := salePrice + tax + shippingCharge

		var firstImage string
		var firstVariantImage models.ProductVariantsImage
		if err := config.DB.Unscoped().Where("product_variant_id = ?", item.ProductVariant.ID).Order("id ASC").First(&firstVariantImage).Error; err == nil {
			firstImage = firstVariantImage.ProductVariantsImages
		} else {
			logger.Log.Warn("Failed to fetch first variant image",
				zap.Uint("productVariantID", item.ProductVariant.ID),
				zap.Error(err))
		}

		var mainProduct models.ProductDetail
		if err := tx.Unscoped().First(&mainProduct, "id = ?", item.ProductVariant.ProductID).Error; err != nil {
			logger.Log.Error("Failed to fetch main product",
				zap.Uint("productID", item.ProductVariant.ProductID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch main product", "Something Went Wrong", "/checkout")
			return
		}

		var category models.Categories
		if err := tx.First(&category, "id = ?", mainProduct.CategoryID).Error; err != nil {
			logger.Log.Error("Failed to fetch category",
				zap.Uint("categoryID", mainProduct.CategoryID),
				zap.Error(err))
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
			logger.Log.Error("Failed to create order item",
				zap.Uint("orderID", orderID),
				zap.Uint("productVariantID", item.ProductVariantID),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create order", "Something Went Wrong", "/checkout")
			return
		}
		logger.Log.Info("Order item created",
			zap.Uint("orderItemID", orderItems.ID),
			zap.Uint("productVariantID", item.ProductVariantID))
	}
}

func SaveOrderAddress(c *gin.Context, tx *gorm.DB, orderID uint, userID uint, addressID string) {
	logger.Log.Info("Saving order address",
		zap.Uint("orderID", orderID),
		zap.Uint("userID", userID))

	address := FetchAddressByIDAndUserID(c, userID, addressID)
	if address == nil {
		logger.Log.Error("Address fetch returned nil",
			zap.Uint("userID", userID),
			zap.String("addressID", addressID))
		return
	}

	shippingAddress := models.ShippingAddress{
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
		logger.Log.Error("Failed to create shipping address",
			zap.Uint("orderID", orderID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create address", "Something Went Wrong", "/checkout")
		return
	}
	logger.Log.Info("Shipping address saved successfully",
		zap.Uint("orderID", orderID),
		zap.Uint("userID", userID))
}

func ClearCart(c *gin.Context, tx *gorm.DB, ordered map[uint]int) {
	logger.Log.Info("Clearing cart")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var cart models.Cart
	if err := tx.First(&cart, "user_ID = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Cart not found", "Cart not found", "/cart")
		return
	}

	for pId := range ordered {
		var cartItems models.CartItem
		if err := tx.First(&cartItems, "cart_id = ? AND product_variant_id = ?", cart.ID, pId).Error; err != nil {
			logger.Log.Error("Cart item not found",
				zap.Uint("cartID", cart.ID),
				zap.Uint("productVariantID", pId),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusBadRequest, "Product Not Found In Cart", "Something Went Wrong", "/cart")
			return
		}
		if err := tx.Unscoped().Where("product_variant_id = ? AND cart_id = ?", pId, cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			logger.Log.Error("Failed to delete cart item",
				zap.Uint("cartID", cart.ID),
				zap.Uint("productVariantID", pId),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Something Went Wrong", "/cart")
			return
		}
		logger.Log.Info("Cart item cleared",
			zap.Uint("cartID", cart.ID),
			zap.Uint("productVariantID", pId))
	}
}

func FetchOrderItems(c *gin.Context, tx *gorm.DB, orderID uint) []models.OrderItem {
	logger.Log.Info("Fetching order items", zap.Uint("orderID", orderID))

	var orderItems []models.OrderItem
	if err := tx.Find(&orderItems, "order_id = ?", orderID).Error; err != nil {
		logger.Log.Error("Failed to fetch order items",
			zap.Uint("orderID", orderID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Order not found", "Something Went Wrong", "/profile/order/details")
		return nil
	}
	logger.Log.Info("Order items fetched successfully",
		zap.Uint("orderID", orderID),
		zap.Int("itemCount", len(orderItems)))
	return orderItems
}

func FetchCartByUserID(c *gin.Context, userID uint) *models.Cart {
	logger.Log.Info("Fetching cart by user ID", zap.Uint("userID", userID))

	var cart models.Cart
	if err := config.DB.First(&cart, "user_ID = ?", userID).Error; err != nil {
		logger.Log.Error("Cart not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Cart not found", "Cart not found", "/")
		return nil
	}
	logger.Log.Info("Cart fetched successfully",
		zap.Uint("cartID", cart.ID),
		zap.Uint("userID", userID))
	return &cart
}

func FetchCartItemByCartID(c *gin.Context, cartID uint) []models.CartItem {
	logger.Log.Info("Fetching cart items by cart ID", zap.Uint("cartID", cartID))

	var cartItems []models.CartItem
	if err := config.DB.Find(&cartItems, "cart_id = ?", cartID).Error; err != nil {
		logger.Log.Error("Failed to fetch cart items",
			zap.Uint("cartID", cartID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to fetch cart items", "/cart")
		return nil
	}
	logger.Log.Info("Cart items fetched successfully",
		zap.Uint("cartID", cartID),
		zap.Int("itemCount", len(cartItems)))
	return cartItems
}

func FetchAddressByIDAndUserID(c *gin.Context, userID uint, addressID string) *models.UserAddress {
	logger.Log.Info("Fetching address by ID and user ID",
		zap.Uint("userID", userID),
		zap.String("addressID", addressID))

	addressId, adErr := strconv.Atoi(addressID)
	if adErr != nil {
		logger.Log.Error("Invalid address ID",
			zap.String("addressID", addressID),
			zap.Error(adErr))
		return nil
	}

	var address models.UserAddress
	if err := config.DB.First(&address, "user_id = ? AND id = ?", userID, addressId).Error; err != nil {
		logger.Log.Error("Failed to fetch address",
			zap.Uint("userID", userID),
			zap.Int("addressID", addressId),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Address not found", "Something Went Wrong", "/checkout")
		return nil
	}
	logger.Log.Info("Address fetched successfully",
		zap.Uint("userID", userID),
		zap.Int("addressID", addressId))
	return &address
}

func FetchReservedProducts(c *gin.Context, userID uint) []models.ReservedStock {
	logger.Log.Info("Fetching reserved products", zap.Uint("userID", userID))

	var reservedProducts []models.ReservedStock
	if err := config.DB.Unscoped().Preload("ProductVariant").
		Find(&reservedProducts, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Failed to fetch reserved products",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Database Error", "Unable to fetch cart items", "/cart")
		return nil
	}
	logger.Log.Info("Reserved products fetched successfully",
		zap.Uint("userID", userID),
		zap.Int("productCount", len(reservedProducts)))
	return reservedProducts
}

type ReservedProductCheckResult struct {
	ReservedMap     map[uint]int
	RegularPrice    float64
	ProductDiscount float64
	TotalDiscount   float64
	ShippingCharge  float64
	Tax             float64
	Total           float64
}

func ReservedProductCheck(c *gin.Context, reservedProducts []models.ReservedStock, cartItems []services.CartItemDetailWithDiscount) (*ReservedProductCheckResult, error) {
	logger.Log.Info("Checking reserved products")

	shippingCharge := 100.0
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
		salePrice += (r.ProductVariant.SalePrice - discountAmount) * float64(r.Quantity)
	}

	for _, item := range cartItems {
		reservedQty, exists := reservedMap[item.CartItem.ProductID]
		if exists {
			if item.CartItem.Quantity != reservedQty {
				logger.Log.Error("Mismatch between cart and reserved quantity",
					zap.Uint("productID", item.CartItem.ProductID),
					zap.Int("cartQty", int(item.CartItem.Quantity)),
					zap.Int("reservedQty", reservedQty))
				helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product ", "Something Went Wrong", "/cart")
				return nil, errors.New("cart-reserved quantity mismatch")
			}
		} else {
			logger.Log.Error("Product in cart not found in reserved",
				zap.Uint("productID", item.CartItem.ProductID))
			helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 2", "Something Went Wrong", "/cart")
			return nil, errors.New("cart product not in reserved")
		}
	}

	for pID := range reservedMap {
		found := false
		for _, items := range cartItems {
			if items.CartItem.ProductID == pID {
				found = true
				break
			}
		}
		if !found {
			logger.Log.Error("Product in reserved not found in cart",
				zap.Uint("productID", pID))
			helper.RespondWithError(c, http.StatusBadRequest, "Mismatch cart items and reserved product 3", "Something Went Wrong", "/cart")
			return nil, errors.New("reserved product not in cart")
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

	logger.Log.Info("Reserved products checked successfully",
		zap.Float64("total", total),
		zap.Int("productCount", len(reservedMap)))

	return &ReservedProductCheckResult{
		ReservedMap:     reservedMap,
		RegularPrice:    regularPrice,
		ProductDiscount: productDiscount,
		TotalDiscount:   totalDiscount,
		ShippingCharge:  shippingCharge,
		Tax:             tax,
		Total:           total,
	}, nil
}

func DeleteReservedItems(c *gin.Context, tx *gorm.DB, productVariantID uint, userID uint) {
	logger.Log.Info("Deleting reserved items",
		zap.Uint("productVariantID", productVariantID),
		zap.Uint("userID", userID))

	if err := tx.Unscoped().Where("product_variant_id = ? AND user_id = ?", productVariantID, userID).Delete(&models.ReservedStock{}).Error; err != nil {
		logger.Log.Error("Failed to delete reserved items",
			zap.Uint("productVariantID", productVariantID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete reservations", "Something Went Wrong", "/cart")
		return
	}
	logger.Log.Info("Reserved items deleted successfully",
		zap.Uint("productVariantID", productVariantID),
		zap.Uint("userID", userID))
}

func CancelSpecificOrder(c *gin.Context) {
	logger.Log.Info("Requested specific order cancellation")

	userID := helper.FetchUserID(c)
	orderItemID := c.Param("id")
	logger.Log.Debug("Fetched user ID and order item ID",
		zap.Uint("userID", userID),
		zap.String("orderItemID", orderItemID))

	var inputReason struct {
		Reason string `json:"cancelReason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&inputReason); err != nil {
		logger.Log.Error("Failed to bind cancellation reason", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Reason Not Found", "Reason is required", "")
		return
	}

	tx := config.DB.Begin()

	var orderItems models.OrderItem
	if err := tx.First(&orderItems, "id = ? AND user_id = ?", orderItemID, userID).Error; err != nil {
		logger.Log.Error("Order item not found",
			zap.String("orderItemID", orderItemID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Order Items Not Found", "Something Went Wrong", "")
		return
	}

	var order models.Order
	if err := tx.First(&order, "user_id = ? AND id = ?", userID, orderItems.OrderID).Error; err != nil {
		logger.Log.Error("Order not found",
			zap.Uint("orderID", orderItems.OrderID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
		return
	}

	switch orderItems.OrderStatus {
	case "Returned":
		logger.Log.Warn("Cannot cancel returned order",
			zap.Uint("orderItemID", orderItems.ID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "You cannot cancel this order as it has already been returned.", "You cannot cancel this order as it has already been returned.", "")
		return
	case "Delivered":
		logger.Log.Warn("Cannot cancel delivered order",
			zap.Uint("orderItemID", orderItems.ID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "You cannot cancel this order as it has already been delivered.", "You cannot cancel this order as it has already been delivered.", "")
		return
	case "Cancelled":
		logger.Log.Warn("Order already cancelled",
			zap.Uint("orderItemID", orderItems.ID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "You have already cancelled this order.", "You have already cancelled this order.", "")
		return
	case "Failed":
		logger.Log.Warn("Cannot cancel failed order",
			zap.Uint("orderItemID", orderItems.ID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Sorry, you cannot cancel this order as it has already failed.", "Sorry, you cannot cancel this order as it has already failed.", "")
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
		logger.Log.Error("Product not found",
			zap.Uint("productVariantID", orderItems.ProductVariantID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
		return
	}
	if err := tx.Model(&product).Where("id = ?", orderItems.ProductVariantID).
		Update("stock_quantity", product.StockQuantity+orderItems.Quantity).Error; err != nil {
		logger.Log.Error("Failed to update stock quantity",
			zap.Uint("productVariantID", orderItems.ProductVariantID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Stock Reverse Failed", "Stock Update Failed", "/checkout")
		return
	}

	var payment models.PaymentDetail
	if err := tx.First(&payment, "order_item_id = ? AND user_id = ?", orderItems.ID, userID).Error; err != nil {
		logger.Log.Error("Payment details not found",
			zap.Uint("orderItemID", orderItems.ID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Payment Details Not Found", "Something Went Wrong", "")
		return
	}

	if payment.PaymentStatus == "Completed" {
		var wallet models.Wallet
		if err := tx.First(&wallet, "user_id = ?", userID).Error; err != nil {
			logger.Log.Error("Wallet not found",
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "User Wallet Not Found", "Something Went Wrong", "")
			return
		}

		if IsMinusAmount {
			lastBalance := wallet.Balance
			wallet.Balance += -order.CouponDiscountAmount
			if err := tx.Save(&wallet).Error; err != nil {
				logger.Log.Error("Failed to update wallet",
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
				return
			}

			receiptID := "rcpt_" + uuid.New().String()
			rand.Seed(time.Now().UnixNano())
			transactionID := fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

			walletTransaction := models.WalletTransaction{
				UserID:        userID,
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
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
				return
			}

			lastBalance = wallet.Balance
			wallet.Balance += orderItems.ProductSalePrice + orderItems.Tax
			if err := tx.Save(&wallet).Error; err != nil {
				logger.Log.Error("Failed to update wallet",
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
				return
			}

			receiptID = "rcpt_" + uuid.New().String()
			rand.Seed(time.Now().UnixNano())
			transactionID = fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

			walletTransaction = models.WalletTransaction{
				UserID:        userID,
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
					zap.Uint("userID", userID),
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
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
				return
			}

			receiptID := "rcpt_" + uuid.New().String()
			rand.Seed(time.Now().UnixNano())
			transactionID := fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

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
				logger.Log.Error("Failed to create wallet transaction",
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
				return
			}
		}

		if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, orderItems.ID).
			Update("payment_status", "Refunded").Error; err != nil {
			logger.Log.Error("Failed to update payment status to Refunded",
				zap.Uint("orderItemID", orderItems.ID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Payment Status Update Failed", "Payment Status Update Failed", "/checkout")
			return
		}
	} else if payment.PaymentStatus == "Pending" || payment.PaymentStatus == "Failed" {
		if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, orderItems.ID).
			Update("payment_status", "Cancelled").Error; err != nil {
			logger.Log.Error("Failed to update payment status to Cancelled",
				zap.Uint("orderItemID", orderItems.ID),
				zap.Error(err))
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
				"shipping_charge":        shipCharge,
				"coupon_discount_amount": gorm.Expr("NULL"),
				"is_coupon_applied":      false,
			}).Error; err != nil {
			logger.Log.Error("Failed to update order with coupon removal",
				zap.Uint("orderID", order.ID),
				zap.Error(err))
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
			logger.Log.Error("Failed to update order shipping charge",
				zap.Uint("orderID", order.ID),
				zap.Error(err))
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
		logger.Log.Error("Failed to update order item status",
			zap.Uint("orderItemID", orderItems.ID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
		return
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed to commit transaction", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Transaction Failed", "Order cancellation failed", "")
		return
	}
	logger.Log.Info("Specific order cancelled successfully",
		zap.Uint("userID", userID),
		zap.Uint("orderItemID", orderItems.ID),
		zap.Float64("refundAmount", refundAmount))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order cancelled successfully",
		"code":    http.StatusOK,
	})
}

func CancelAllOrderItems(c *gin.Context) {
	logger.Log.Info("Requested cancellation of all order items")

	userID := helper.FetchUserID(c)
	orderID := c.Param("id")
	logger.Log.Debug("Fetched user ID and order ID",
		zap.Uint("userID", userID),
		zap.String("orderID", orderID))

	var inputReason struct {
		Reason string `form:"cancelAllReason" json:"cancelAllReason" binding:"required"`
	}
	if err := c.ShouldBind(&inputReason); err != nil {
		logger.Log.Error("Failed to bind cancellation reason", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Reason Not Found", "Reason is required", "")
		return
	}

	tx := config.DB.Begin()

	var order models.Order
	if err := tx.First(&order, "user_id = ? AND id = ?", userID, orderID).Error; err != nil {
		logger.Log.Error("Order not found",
			zap.String("orderID", orderID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
		return
	}

	var orderItems []models.OrderItem
	if err := tx.Find(&orderItems, "user_id = ? AND order_id = ?", userID, order.ID).Error; err != nil {
		logger.Log.Error("Order items not found",
			zap.Uint("orderID", order.ID),
			zap.Uint("userID", userID),
			zap.Error(err))
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
			logger.Log.Warn("Cannot cancel order due to item status",
				zap.Uint("orderItemID", itm.ID),
				zap.String("status", itm.OrderStatus))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusBadRequest, "Something Went Wrong", "Something Went Wrong", "")
			return
		}

		var product models.ProductVariantDetails
		if err := tx.First(&product, itm.ProductVariantID).Error; err != nil {
			logger.Log.Error("Product not found",
				zap.Uint("productVariantID", itm.ProductVariantID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Product Not Found", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&product).Where("id = ?", itm.ProductVariantID).
			Update("stock_quantity", product.StockQuantity+itm.Quantity).Error; err != nil {
			logger.Log.Error("Failed to update stock quantity",
				zap.Uint("productVariantID", itm.ProductVariantID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Stock Reverse Failed", "Stock Update Failed", "/checkout")
			return
		}

		var payment models.PaymentDetail
		if err := tx.First(&payment, "order_item_id = ? AND user_id = ?", itm.ID, userID).Error; err != nil {
			logger.Log.Error("Payment details not found",
				zap.Uint("orderItemID", itm.ID),
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Payment Details Not Found", "Something Went Wrong", "")
			return
		}
		paymentMethod = payment.PaymentMethod
		if payment.PaymentStatus == "Completed" {
			IsPaymentCompleted = true
			if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, itm.ID).
				Update("payment_status", "Refunded").Error; err != nil {
				logger.Log.Error("Failed to update payment status to Refunded",
					zap.Uint("orderItemID", itm.ID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment Status Update Failed", "Payment Status Update Failed", "/checkout")
				return
			}
		} else if payment.PaymentStatus == "Pending" || payment.PaymentStatus == "Failed" {
			if err := tx.Model(&payment).Where("user_id = ? AND order_item_id = ?", userID, itm.ID).
				Update("payment_status", "Cancelled").Error; err != nil {
				logger.Log.Error("Failed to update payment status to Cancelled",
					zap.Uint("orderItemID", itm.ID),
					zap.Error(err))
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
			logger.Log.Error("Wallet not found",
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "User Wallet Not Found", "Something Went Wrong", "")
			return
		}
		lastBalance := wallet.Balance
		wallet.Balance += refundAmount
		if err := tx.Save(&wallet).Error; err != nil {
			logger.Log.Error("Failed to update wallet",
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Wallet", "Something Went Wrong", "")
			return
		}

		receiptID := "rcpt_" + uuid.New().String()
		rand.Seed(time.Now().UnixNano())
		transactionID := fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))

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
			logger.Log.Error("Failed to create wallet transaction",
				zap.Uint("userID", userID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Create Transaction History", "Something Went Wrong", "")
			return
		}
	}

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
		couponDetails.UsersUsedCount -= 1
		if err := tx.Unscoped().Save(&couponDetails).Error; err != nil {
			logger.Log.Error("Failed to update coupon",
				zap.String("couponCode", order.CouponCode),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusNotFound, "Failed to Update Coupon", "Something Went Wrong", "")
			return
		}
		if err := tx.Model(&order).Where("user_id = ? AND id = ?", userID, order.ID).
			Updates(map[string]interface{}{
				"coupon_code":            gorm.Expr("NULL"),
				"coupon_discount_amount": gorm.Expr("NULL"),
				"is_coupon_applied":      false,
			}).Error; err != nil {
			logger.Log.Error("Failed to update order with coupon removal",
				zap.Uint("orderID", order.ID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
			return
		}
	} else {
		if err := tx.Model(&order).Where("user_id = ? AND id = ?", userID, order.ID).
			Updates(map[string]interface{}{
				"coupon_discount_amount": gorm.Expr("NULL"),
			}).Error; err != nil {
			logger.Log.Error("Failed to update order coupon discount",
				zap.Uint("orderID", order.ID),
				zap.Error(err))
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
		logger.Log.Error("Failed to update order items status",
			zap.Uint("orderID", order.ID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Order Status Update Failed", "Order Status Update Failed", "/checkout")
		return
	}

	tx.Commit()
	logger.Log.Info("All order items cancelled successfully",
		zap.Uint("userID", userID),
		zap.Uint("orderID", order.ID),
		zap.Int("itemCount", len(orderItems)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Order cancelled successfully",
		"code":    http.StatusOK,
	})
}

func ReturnOrder(c *gin.Context) {
	logger.Log.Info("Requested order return")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var input struct {
		Reason            string `json:"reason" binding:"required"`
		AdditionalDetails string `json:"additionalDetails" binding:"required"`
		ProductId         string `json:"productId" binding:"required"`
		OrderId           string `json:"orderId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		logger.Log.Error("Failed to bind return request data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Enter All Fields", "")
		return
	}

	ordId, err := strconv.ParseUint(input.OrderId, 10, 32)
	if err != nil {
		logger.Log.Error("Invalid order ID",
			zap.String("orderId", input.OrderId),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Something Went Wrong", "")
		return
	}

	prdtId, err := strconv.ParseUint(input.ProductId, 10, 32)
	if err != nil {
		logger.Log.Error("Invalid product ID",
			zap.String("productId", input.ProductId),
			zap.Error(err))
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
		logger.Log.Error("Failed to create return request",
			zap.Uint("orderItemID", returnRequest.OrderItemID),
			zap.Uint("productVariantID", returnRequest.ProductVariantID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to submit return request", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Return request submitted successfully",
		zap.String("requestUID", reqstId),
		zap.Uint("userID", userID),
		zap.Uint("orderItemID", uint(ordId)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Return request submitted successfully. Awaiting admin approval.",
		"code":    http.StatusOK,
	})
}
