package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"go.uber.org/zap"
)

func DownloadInvoice(c *gin.Context) {
	logger.Log.Info("Requested invoice download")

	orderIdStr := c.Param("id")
	orderId, err := strconv.ParseUint(orderIdStr, 10, 64)
	if err != nil {
		logger.Log.Error("Invalid order ID", zap.String("orderIdStr", orderIdStr), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var order models.Order
	result := config.DB.Preload("ShippingAddress").First(&order, orderId)
	if result.Error != nil {
		logger.Log.Error("Order not found", zap.Uint64("orderID", orderId), zap.Error(result.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var user models.UserAuth
	result = config.DB.First(&user, order.UserID)
	if result.Error != nil {
		logger.Log.Error("User not found for order",
			zap.Uint("userID", order.UserID),
			zap.Uint64("orderID", orderId),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	var orderItems []models.OrderItem
	result = config.DB.Where("order_id = ?", orderId).Find(&orderItems)
	if result.Error != nil {
		logger.Log.Error("Failed to fetch order items",
			zap.Uint64("orderID", orderId),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order items"})
		return
	}

	productIds := make([]uint, len(orderItems))
	for i, item := range orderItems {
		productIds[i] = item.ProductVariantID
	}

	var products []models.ProductVariantDetails
	result = config.DB.Where("id IN ?", productIds).Find(&products)
	if result.Error != nil {
		logger.Log.Error("Failed to fetch products for order",
			zap.Any("productIDs", productIds),
			zap.Uint64("orderID", orderId),
			zap.Error(result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	productMap := make(map[uint]models.ProductVariantDetails)
	for _, product := range products {
		productMap[product.ID] = product
	}

	pdf, err := GenerateInvoice(order, user, orderItems, productMap)
	if err != nil {
		logger.Log.Error("Failed to generate invoice PDF",
			zap.Uint64("orderID", orderId),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invoice"})
		return
	}

	fileName := fmt.Sprintf("invoice_%s.pdf", order.OrderUID)
	err = pdf.OutputFileAndClose(fileName)
	if err != nil {
		logger.Log.Error("Failed to save invoice PDF",
			zap.String("fileName", fileName),
			zap.Uint64("orderID", orderId),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save invoice"})
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")

	c.File(fileName)
	logger.Log.Info("Invoice downloaded successfully",
		zap.Uint64("orderID", orderId),
		zap.String("fileName", fileName))

	go func() {
		time.Sleep(5 * time.Second)
		if err := os.Remove(fileName); err != nil {
			logger.Log.Warn("Failed to remove temporary invoice file",
				zap.String("fileName", fileName),
				zap.Error(err))
		} else {
			logger.Log.Debug("Temporary invoice file removed",
				zap.String("fileName", fileName))
		}
	}()
}

func GenerateInvoice(order models.Order, user models.UserAuth, orderItems []models.OrderItem, products map[uint]models.ProductVariantDetails) (*gofpdf.Fpdf, error) {
	logger.Log.Info("Generating invoice PDF", zap.Uint("orderID", order.ID))

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pageWidth := 210.0

	if _, err := os.Stat(config.CompanyConfig.LogoFilePath); err == nil {
		pdf.Image(config.CompanyConfig.LogoFilePath, 10, 10, 40, 0, false, "", 0, "")
		logger.Log.Debug("Added company logo to invoice",
			zap.String("logoPath", config.CompanyConfig.LogoFilePath))
	} else {
		logger.Log.Warn("Company logo file not found",
			zap.String("logoPath", config.CompanyConfig.LogoFilePath),
			zap.Error(err))
	}

	pdf.SetFont("Arial", "B", 20)
	title := "INVOICE"
	titleWidth := pdf.GetStringWidth(title)
	pdf.SetXY((pageWidth-titleWidth)/2, 20)
	pdf.Cell(titleWidth, 10, title)
	pdf.Ln(20)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetXY(10, 40)
	pdf.Cell(80, 10, "ORDER DETAILS")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetX(10)
	pdf.Cell(80, 6, fmt.Sprintf("Order #: %s", order.OrderUID))
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(80, 6, fmt.Sprintf("Date: %s", order.CreatedAt.Format("2006-01-02")))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetX(10)
	pdf.Cell(90, 10, "COMPANY DETAILS")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetX(10)
	pdf.Cell(90, 6, config.CompanyConfig.Name)
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, config.CompanyConfig.Address1)
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, config.CompanyConfig.Address2)
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, "Email: "+config.CompanyConfig.Email)
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetX(10)
	pdf.Cell(90, 10, "SHIPPING ADDRESS")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetX(10)
	pdf.Cell(90, 6, user.FullName)
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, fmt.Sprintf("Email: %s", user.Email))
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, fmt.Sprintf("Phone: %s", order.ShippingAddress.Mobile))
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, order.ShippingAddress.Address)
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, fmt.Sprintf("%s", order.ShippingAddress.State))
	pdf.Ln(6)
	pdf.SetX(10)
	pdf.Cell(90, 6, fmt.Sprintf("%s - %s", order.ShippingAddress.Country, order.ShippingAddress.PinCode))
	pdf.Ln(6)
	if order.ShippingAddress.Landmark != "" {
		pdf.SetX(10)
		pdf.Cell(90, 6, fmt.Sprintf("Landmark: %s", order.ShippingAddress.Landmark))
		pdf.Ln(6)
	}
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "PRODUCT DETAILS")
	pdf.Ln(10)

	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 10)

	pdf.CellFormat(75, 8, "Product", "1", 0, "L", true, 0, "")
	pdf.CellFormat(20, 8, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(25, 8, "Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 8, "Discount", "1", 0, "R", true, 0, "")
	pdf.CellFormat(20, 8, "Tax", "1", 0, "R", true, 0, "")
	pdf.CellFormat(25, 8, "Total", "1", 1, "R", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	var subtotal float64
	var totalDiscount float64
	var totalTax float64

	for _, item := range orderItems {
		product := item.ProductName
		regularPrice := item.ProductRegularPrice
		discountPrice := item.ProductSalePrice
		discount := regularPrice - discountPrice
		quantity := float64(item.Quantity)

		taxAmount := item.Tax
		lineTotal := (discountPrice + taxAmount) * quantity

		subtotal += regularPrice * quantity
		totalDiscount += order.TotalProductDiscount
		totalTax += taxAmount

		pdf.CellFormat(75, 8, product, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 8, strconv.Itoa(item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", regularPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", discount*quantity), "1", 0, "R", false, 0, "")
		pdf.CellFormat(20, 8, fmt.Sprintf("%.2f", taxAmount*quantity), "1", 0, "R", false, 0, "")
		pdf.CellFormat(25, 8, fmt.Sprintf("%.2f", lineTotal), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "ORDER SUMMARY")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	leftColWidth := 150.0
	rightColWidth := 40.0

	pdf.CellFormat(leftColWidth, 6, "Subtotal:", "", 0, "R", false, 0, "")
	pdf.CellFormat(rightColWidth, 6, fmt.Sprintf("%.2f", subtotal), "", 1, "R", false, 0, "")

	pdf.CellFormat(leftColWidth, 6, "Product Discount:", "", 0, "R", false, 0, "")
	pdf.CellFormat(rightColWidth, 6, fmt.Sprintf("%.2f", totalDiscount), "", 1, "R", false, 0, "")

	pdf.CellFormat(leftColWidth, 6, "Tax:", "", 0, "R", false, 0, "")
	pdf.CellFormat(rightColWidth, 6, fmt.Sprintf("%.2f", totalTax), "", 1, "R", false, 0, "")

	if order.CouponDiscountAmount > 0 {
		pdf.CellFormat(leftColWidth, 6, "Coupon Discount:", "", 0, "R", false, 0, "")
		pdf.CellFormat(rightColWidth, 6, fmt.Sprintf("%.2f", order.CouponDiscountAmount), "", 1, "R", false, 0, "")
	}

	pdf.CellFormat(leftColWidth, 6, "Shipping Charge:", "", 0, "R", false, 0, "")
	if order.ShippingCharge == 0 {
		pdf.CellFormat(rightColWidth, 6, "FREE", "", 1, "R", false, 0, "")
	} else {
		pdf.CellFormat(rightColWidth, 6, fmt.Sprintf("%.2f", order.ShippingCharge), "", 1, "R", false, 0, "")
	}
	shippingCharge := 0.0
	if order.ShippingCharge == 0 {
		shippingCharge = 100
	}

	totalAllDiscounts := totalDiscount + order.CouponDiscountAmount + shippingCharge
	pdf.CellFormat(leftColWidth, 6, "Total Discount:", "", 0, "R", false, 0, "")
	pdf.CellFormat(rightColWidth, 6, fmt.Sprintf("%.2f", totalAllDiscounts), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(leftColWidth, 8, "Total Amount:", "T", 0, "R", false, 0, "")
	pdf.CellFormat(rightColWidth, 8, fmt.Sprintf("%.2f", order.TotalAmount), "T", 1, "R", false, 0, "")

	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 12)
	thankYouMsg := "Thank You For Shopping With Us!"
	msgWidth := pdf.GetStringWidth(thankYouMsg)
	pdf.SetX((pageWidth - msgWidth) / 2)
	pdf.Cell(msgWidth, 10, thankYouMsg)
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	contactMsg := "Contact Us At laptixinfo@gmail.com for any queries."
	contactWidth := pdf.GetStringWidth(contactMsg)
	pdf.SetX((pageWidth - contactWidth) / 2)
	pdf.Cell(contactWidth, 6, contactMsg)

	logger.Log.Info("Invoice PDF generated successfully",
		zap.Uint("orderID", order.ID),
		zap.Int("itemCount", len(orderItems)),
		zap.Float64("totalAmount", order.TotalAmount))
	return pdf, nil
}
