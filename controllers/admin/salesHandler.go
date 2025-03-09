package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type SalesOverviewDTO struct {
	Month   string  `json:"month"`
	Revenue float64 `json:"revenue"`
	Orders  int     `json:"orders"`
}

type CategorySalesDTO struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

type SalesDashboardDTO struct {
	TotalAmount     float64            `json:"totalAmount"`
	OrderCount      int                `json:"orderCount"`
	AverageOrder    float64            `json:"averageOrder"`
	DiscountApplied float64            `json:"discountApplied"`
	CouponDiscount  float64            `json:"couponDiscount"`
	TotalRevenue    float64            `json:"totalRevenue"`
	SalesOverview   []SalesOverviewDTO `json:"salesOverview"`
	CategorySales   []CategorySalesDTO `json:"categorySales"`
	RecentOrders    []RecentOrderDTO   `json:"recentOrders"`
	TotalOrderCount int                `json:"totalOrderCount"`
	CurrentPage     int                `json:"currentPage"`
	TotalPages      int                `json:"totalPages"`
	ItemsPerPage    int                `json:"itemsPerPage"`
}

type RecentOrderDTO struct {
	OrderID     string    `json:"orderId"`
	Customer    string    `json:"customer"`
	Date        time.Time `json:"date"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	OrderNumber string    `json:"orderNumber"`
}

type Filter struct {
	Period   string     `json:"period"`
	FromDate *time.Time `json:"fromDate"`
	ToDate   *time.Time `json:"toDate"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
}

func GetSalesDashboard(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)

	var fromDate, toDate *time.Time
	fromDateStr := c.Query("fromDate")
	toDateStr := c.Query("toDate")

	if fromDateStr != "" && toDateStr != "" {
		parsedFromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err == nil {
			fromDate = &parsedFromDate
		}

		parsedToDate, err := time.Parse("2006-01-02", toDateStr)
		if err == nil {
			toDate = &parsedToDate
		}
	}

	filter := Filter{
		Period:   period,
		FromDate: fromDate,
		ToDate:   toDate,
		Page:     page,
		PageSize: 5, 
	}

	data, err := GetSalesDashboardData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "sales.html", data)
}

func GetSalesData(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)

	var fromDate, toDate *time.Time
	fromDateStr := c.Query("fromDate")
	toDateStr := c.Query("toDate")

	if fromDateStr != "" && toDateStr != "" {
		parsedFromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err == nil {
			fromDate = &parsedFromDate
		}

		parsedToDate, err := time.Parse("2006-01-02", toDateStr)
		if err == nil {
			toDate = &parsedToDate
		}
	}

	filter := Filter{
		Period:   period,
		FromDate: fromDate,
		ToDate:   toDate,
		Page:     page,
		PageSize: 5,  
	}

	data, err := GetSalesDashboardData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}


func DownloadSalesReport(c *gin.Context) {
	format := c.DefaultQuery("format", "excel")
	period := c.DefaultQuery("period", "monthly")

	var fromDate, toDate *time.Time
	fromDateStr := c.Query("fromDate")
	toDateStr := c.Query("toDate")

	if fromDateStr != "" && toDateStr != "" {
		parsedFromDate, err := time.Parse("2006-01-02", fromDateStr)
		if err == nil {
			fromDate = &parsedFromDate
		}

		parsedToDate, err := time.Parse("2006-01-02", toDateStr)
		if err == nil {
			toDate = &parsedToDate
		}
	}

	filter := Filter{
		Period:   period,
		FromDate: fromDate,
		ToDate:   toDate,
		Page:     1,
		PageSize: 1000, 
	}

	fileBytes, fileName, err := GenerateSalesReport(filter, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var contentType string
	if format == "pdf" {
		contentType = "application/pdf"
	} else {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Data(http.StatusOK, contentType, fileBytes)
}

func GetSalesDashboardData(filter Filter) (SalesDashboardDTO, error) {
	db := config.DB 

	startDate, endDate := calculateDateRange(filter)

	result := SalesDashboardDTO{
		CurrentPage:  filter.Page,
		ItemsPerPage: filter.PageSize,
	}

	var totalStats struct {
		TotalAmount      float64
		OrderCount       int64
		DiscountApplied  float64
		CouponDiscount   float64
	}

	query := db.Model(&models.Order{})
	if startDate != nil && endDate != nil {
		query = query.Where("order_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Select("SUM(total_amount) as total_amount, COUNT(*) as order_count, " +
		"SUM(total_discount) as discount_applied, SUM(coupon_discount_amount) as coupon_discount").
		Scan(&totalStats).Error; err != nil {
		return result, err
	}

	result.TotalAmount = totalStats.TotalAmount
	result.OrderCount = int(totalStats.OrderCount)
	result.DiscountApplied = totalStats.DiscountApplied
	result.CouponDiscount = totalStats.CouponDiscount
	
	if result.OrderCount > 0 {
		result.AverageOrder = result.TotalAmount / float64(result.OrderCount)
	}
	
	result.TotalRevenue = result.TotalAmount - result.DiscountApplied - result.CouponDiscount

	result.SalesOverview = getSalesOverviewData(db, filter, startDate, endDate)
	
	result.CategorySales = getCategorySalesData(db, startDate, endDate)
	
	recentOrders, totalOrders := getRecentOrders(db, filter, startDate, endDate)
	result.RecentOrders = recentOrders
	result.TotalOrderCount = totalOrders
	
	result.TotalPages = (totalOrders + filter.PageSize - 1) / filter.PageSize
	
	return result, nil
}

func calculateDateRange(filter Filter) (*time.Time, *time.Time) {
	now := time.Now()
	
	if filter.FromDate != nil && filter.ToDate != nil {
		return filter.FromDate, filter.ToDate
	}
	
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	var startDate time.Time
	
	switch filter.Period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 { 
			weekday = 7
		}
		startDate = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "yearly":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	default:
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	}
	
	return &startDate, &endDate
}

func getSalesOverviewData(db *gorm.DB, filter Filter, startDate, endDate *time.Time) []SalesOverviewDTO {
	var results []SalesOverviewDTO
	
	query := db.Model(&models.Order{})
	if startDate != nil && endDate != nil {
		query = query.Where("order_date BETWEEN ? AND ?", startDate, endDate)
	}
	
	var timeFormat string
	switch filter.Period {
	case "daily":
		timeFormat = "2006-01-02 15" 
	case "weekly":
		timeFormat = "2006-01-02" 
	case "monthly":
		timeFormat = "2006-01-02" 
	case "yearly":
		timeFormat = "2006-01" 
	default:
		timeFormat = "2006-01" 
	}
	
	rows, err := query.Select(fmt.Sprintf("TO_CHAR(order_date, '%s'), SUM(total_amount), COUNT(*)", timeFormat)).
	Group(fmt.Sprintf("TO_CHAR(order_date, '%s')", timeFormat)).
	Order(fmt.Sprintf("TO_CHAR(order_date, '%s')", timeFormat)).
	Rows()

	
	if err != nil {
		return results
	}
	defer rows.Close()
	
	for rows.Next() {
		var overview SalesOverviewDTO
		var timeStr string
		rows.Scan(&timeStr, &overview.Revenue, &overview.Orders)
		
		overview.Month = timeStr
		results = append(results, overview)
	}
	
	return results
}

func getCategorySalesData(db *gorm.DB, startDate, endDate *time.Time) []CategorySalesDTO {
	var results []CategorySalesDTO
	
	query := db.Model(&models.OrderItem{}).
		Select("product_category as category, SUM(total) as amount, COUNT(*) as count").
		Joins("JOIN orders ON order_items.order_id = orders.id")
	
	if startDate != nil && endDate != nil {
		query = query.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}
	
	query.Group("product_category").
		Order("amount DESC").
		Limit(5).
		Scan(&results)
	
	return results
}

func getRecentOrders(db *gorm.DB, filter Filter, startDate, endDate *time.Time) ([]RecentOrderDTO, int) {
	var orders []RecentOrderDTO
	var totalCount int64
	
	countQuery := db.Model(&models.Order{}).
		Joins("JOIN user_auths ON orders.user_id = user_auths.id")
	
	if startDate != nil && endDate != nil {
		countQuery = countQuery.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}
	
	countQuery.Count(&totalCount)
	
	dataQuery := db.Model(&models.Order{}).
		Select("orders.order_uid as order_number, CONCAT(user_auths.full_name) as customer, " +
			"orders.order_date as date, orders.total_amount as amount, " +
			"(SELECT order_status FROM order_items WHERE order_items.order_id = orders.id LIMIT 1) as status").
		Joins("JOIN user_auths ON orders.user_id = user_auths.id")
	
	if startDate != nil && endDate != nil {
		dataQuery = dataQuery.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}
	
	offset := (filter.Page - 1) * filter.PageSize
	dataQuery.Order("orders.order_date DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Scan(&orders)
	
	for i := range orders {
		orders[i].OrderID = fmt.Sprintf("#ORD-%s", orders[i].OrderNumber[0:5])
	}
	
	return orders, int(totalCount)
}

func GenerateSalesReport(filter Filter, format string) ([]byte, string, error) {
	data, err := GetSalesDashboardData(filter)
	if err != nil {
		return nil, "", err
	}
	
	now := time.Now()
	fileName := fmt.Sprintf("sales_report_%s_%s", filter.Period, now.Format("2006-01-02"))
	
	if format == "excel" {
		return generateExcelReport(data, fileName+".xlsx")
	} else {
		return generatePDFReport(data, fileName+".pdf")
	}
}

func generateExcelReport(data SalesDashboardDTO, fileName string) ([]byte, string, error) {
	f := excelize.NewFile()
	
	summarySheet := "Summary"
	f.NewSheet(summarySheet)
	f.DeleteSheet("Sheet1") 
	
	f.SetCellValue(summarySheet, "A1", "Sales Dashboard Summary")
	f.SetCellValue(summarySheet, "A3", "Metric")
	f.SetCellValue(summarySheet, "B3", "Value")
	
	f.SetCellValue(summarySheet, "A4", "Total Amount")
	f.SetCellValue(summarySheet, "B4", data.TotalAmount)
	
	f.SetCellValue(summarySheet, "A5", "Total Orders")
	f.SetCellValue(summarySheet, "B5", data.OrderCount)
	
	f.SetCellValue(summarySheet, "A6", "Average Order Value")
	f.SetCellValue(summarySheet, "B6", data.AverageOrder)
	
	f.SetCellValue(summarySheet, "A7", "Total Discount Applied")
	f.SetCellValue(summarySheet, "B7", data.DiscountApplied)
	
	f.SetCellValue(summarySheet, "A8", "Total Coupon Discount")
	f.SetCellValue(summarySheet, "B8", data.CouponDiscount)
	
	f.SetCellValue(summarySheet, "A9", "Total Revenue")
	f.SetCellValue(summarySheet, "B9", data.TotalRevenue)
	
	overviewSheet := "Sales Overview"
	f.NewSheet(overviewSheet)
	
	f.SetCellValue(overviewSheet, "A1", "Period")
	f.SetCellValue(overviewSheet, "B1", "Revenue")
	f.SetCellValue(overviewSheet, "C1", "Orders")
	
	for i, item := range data.SalesOverview {
		row := i + 2
		f.SetCellValue(overviewSheet, fmt.Sprintf("A%d", row), item.Month)
		f.SetCellValue(overviewSheet, fmt.Sprintf("B%d", row), item.Revenue)
		f.SetCellValue(overviewSheet, fmt.Sprintf("C%d", row), item.Orders)
	}
	
	catSheet := "Categories"
	f.NewSheet(catSheet)
	
	f.SetCellValue(catSheet, "A1", "Category")
	f.SetCellValue(catSheet, "B1", "Amount")
	f.SetCellValue(catSheet, "C1", "Count")
	
	for i, item := range data.CategorySales {
		row := i + 2
		f.SetCellValue(catSheet, fmt.Sprintf("A%d", row), item.Category)
		f.SetCellValue(catSheet, fmt.Sprintf("B%d", row), item.Amount)
		f.SetCellValue(catSheet, fmt.Sprintf("C%d", row), item.Count)
	}
	
	ordersSheet := "Recent Orders"
	f.NewSheet(ordersSheet)
	
	f.SetCellValue(ordersSheet, "A1", "Order ID")
	f.SetCellValue(ordersSheet, "B1", "Customer")
	f.SetCellValue(ordersSheet, "C1", "Date")
	f.SetCellValue(ordersSheet, "D1", "Amount")
	f.SetCellValue(ordersSheet, "E1", "Status")
	
	for i, order := range data.RecentOrders {
		row := i + 2
		f.SetCellValue(ordersSheet, fmt.Sprintf("A%d", row), order.OrderID)
		f.SetCellValue(ordersSheet, fmt.Sprintf("B%d", row), order.Customer)
		f.SetCellValue(ordersSheet, fmt.Sprintf("C%d", row), order.Date.Format("2006-01-02"))
		f.SetCellValue(ordersSheet, fmt.Sprintf("D%d", row), order.Amount)
		f.SetCellValue(ordersSheet, fmt.Sprintf("E%d", row), order.Status)
	}
	
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}
	
	return buffer.Bytes(), fileName, nil
}

func generatePDFReport(data SalesDashboardDTO, fileName string) ([]byte, string, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Sales Dashboard Report")
	pdf.Ln(15)
	
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Summary")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 12)
	
	pdf.CellFormat(60, 8, "Metric", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, "Value", "1", 1, "", false, 0, "")
	
	pdf.CellFormat(60, 8, "Total Amount", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("%.2f", data.TotalAmount), "1", 1, "", false, 0, "")
	
	pdf.CellFormat(60, 8, "Total Orders", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("%d", data.OrderCount), "1", 1, "", false, 0, "")
	
	pdf.CellFormat(60, 8, "Average Order Value", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("%.2f", data.AverageOrder), "1", 1, "", false, 0, "")
	
	pdf.CellFormat(60, 8, "Total Discount Applied", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("%.2f", data.DiscountApplied), "1", 1, "", false, 0, "")
	
	pdf.CellFormat(60, 8, "Total Coupon Discount", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("%.2f", data.CouponDiscount), "1", 1, "", false, 0, "")
	
	pdf.CellFormat(60, 8, "Total Revenue", "1", 0, "", false, 0, "")
	pdf.CellFormat(60, 8, fmt.Sprintf("%.2f", data.TotalRevenue), "1", 1, "", false, 0, "")
	
	pdf.Ln(15)
	
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Recent Orders")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(30, 8, "Order ID", "1", 0, "", false, 0, "")
	pdf.CellFormat(50, 8, "Customer", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 8, "Date", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 8, "Amount", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 8, "Status", "1", 1, "", false, 0, "")
	
	for _, order := range data.RecentOrders {
		pdf.CellFormat(30, 8, order.OrderID, "1", 0, "", false, 0, "")
		pdf.CellFormat(50, 8, order.Customer, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, order.Date.Format("2006-01-02"), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%.2f", order.Amount), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, order.Status, "1", 1, "", false, 0, "")
	}
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, "", err
	}
	
	return buf.Bytes(), fileName, nil
}