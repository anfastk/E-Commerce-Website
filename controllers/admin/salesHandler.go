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
	"github.com/jinzhu/now"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type SalesOverviewDTO struct {
	Month       string  `json:"month"`
	TotalAmount float64 `json:"totalAmount"`
	Revenue     float64 `json:"revenue"`
	Orders      int     `json:"orders"`
}

type CategorySalesDTO struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Count    int     `json:"count"`
}

type SalesDashboardDTO struct {
	TotalAmount     float64            `json:"totalAmount"`
	TotalRevenue    float64            `json:"totalRevenue"`
	OrderCount      int                `json:"orderCount"`
	AverageOrder    float64            `json:"averageOrder"`
	DiscountApplied float64            `json:"discountApplied"`
	CouponDiscount  float64            `json:"couponDiscount"`
	SalesOverview   []SalesOverviewDTO `json:"salesOverview"`
	CategorySales   []CategorySalesDTO `json:"categorySales"`
	RecentOrders    []RecentOrderDTO   `json:"recentOrders"`
	TotalOrderCount int                `json:"totalOrderCount"`
	CurrentPage     int                `json:"currentPage"`
	TotalPages      int                `json:"totalPages"`
	ItemsPerPage    int                `json:"itemsPerPage"`
}

type RecentOrderDTO struct {
	OrderID     string    `json:"orderID"`
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
	Status   string     `json:"status"`
}

func GetSalesDashboard(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")
	pageStr := c.DefaultQuery("page", "1")
	status := c.DefaultQuery("status", "Delivered")
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
		Status:   status,
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
	status := c.DefaultQuery("status", "Delivered")
	page, _ := strconv.Atoi(pageStr)

	var fromDate, toDate *time.Time
	if fromDateStr := c.Query("fromDate"); fromDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			fromDate = &parsed
		}
	}
	if toDateStr := c.Query("toDate"); toDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", toDateStr); err == nil {
			toDate = &parsed
		}
	}

	filter := Filter{
		Period:   period,
		FromDate: fromDate,
		ToDate:   toDate,
		Page:     page,
		PageSize: 5,
		Status:   status,
	}

	data, err := GetSalesDashboardData(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func GetRecentOrdersUnfiltered(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	pageSize := 5

	db := config.DB
	var orders []RecentOrderDTO
	var totalCount int64

	countQuery := db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Joins("JOIN user_auths ON orders.user_id = user_auths.id")
	countQuery.Count(&totalCount)

	dataQuery := db.Model(&models.OrderItem{}).
		Select(`
            orders.order_uid as order_number,
            CONCAT(user_auths.full_name) as customer,
            orders.order_date as date,
            SUM(order_items.total) - COALESCE(orders.coupon_discount_amount, 0) as amount,
            order_items.order_status as status
        `).
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Joins("JOIN user_auths ON orders.user_id = user_auths.id")

	offset := (page - 1) * pageSize
	dataQuery.Group("orders.order_uid, user_auths.full_name, orders.order_date, order_items.order_status, orders.coupon_discount_amount").
		Order("orders.order_date DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&orders)

	for i := range orders {
		orders[i].OrderID = fmt.Sprintf("ORD-%s", orders[i].OrderNumber[0:8])
	}

	totalPages := (int(totalCount) + pageSize - 1) / pageSize
	response := gin.H{
		"recentOrders":    orders,
		"totalOrderCount": int(totalCount),
		"currentPage":     page,
		"totalPages":      totalPages,
		"itemsPerPage":    pageSize,
	}

	c.JSON(http.StatusOK, response)
}

func DownloadSalesReport(c *gin.Context) {
	format := c.DefaultQuery("format", "excel")
	period := c.DefaultQuery("period", "monthly")
	status := c.DefaultQuery("status", "Delivered")

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
		Status:   status,
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
		TotalAmount     float64
		TotalRevenue    float64
		OrderCount      int64
		DiscountApplied float64
		CouponDiscount  float64
	}

	query := db.Model(&models.Order{}).
		Select(`
            SUM(oi.total_amount) as total_amount,
            SUM(orders.total_amount) as total_revenue,
            COUNT(DISTINCT orders.id) as order_count,
            SUM(oi.discount_applied) as discount_applied,
            SUM(COALESCE(orders.coupon_discount_amount, 0)) as coupon_discount
        `).
		Joins(`
            JOIN (
                SELECT 
                    order_id,
                    SUM(sub_total + tax) as total_amount,
                    SUM(product_regular_price - product_sale_price) as discount_applied
                FROM order_items 
                WHERE order_status = ?
                GROUP BY order_id
            ) oi ON oi.order_id = orders.id
        `, filter.Status)

	if startDate != nil && endDate != nil {
		query = query.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Scan(&totalStats).Error; err != nil {
		return result, err
	}

	result.TotalAmount = totalStats.TotalAmount
	result.TotalRevenue = totalStats.TotalRevenue
	result.OrderCount = int(totalStats.OrderCount)
	result.DiscountApplied = totalStats.DiscountApplied
	result.CouponDiscount = totalStats.CouponDiscount
	if result.OrderCount > 0 {
		result.AverageOrder = result.TotalRevenue / float64(result.OrderCount)
	}

	result.SalesOverview = getSalesOverviewData(db, filter, startDate, endDate)
	result.CategorySales = getCategorySalesData(db, startDate, endDate, filter.Status)
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

	var startDate time.Time
	endDate := now

	switch filter.Period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 0, 1).Add(-time.Second)
	case "weekly":
		startDate = now.AddDate(0, 0, -7*7)
		weekday := int(startDate.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = startDate.AddDate(0, 0, -weekday+1)
		endDate = now
	case "monthly":
		startDate = now.AddDate(0, -7, 0)
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
		endDate = now
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -7)
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	case "yearly":
		startDate = now.AddDate(-6, 0, 0)
		startDate = time.Date(startDate.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = now
	default:
		startDate = now.AddDate(0, -7, 0)
		startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())
		endDate = now
	}
	return &startDate, &endDate
}

func getSalesOverviewData(db *gorm.DB, filter Filter, startDate, endDate *time.Time) []SalesOverviewDTO {
	var results []SalesOverviewDTO
	var groupBy, orderBy string
	var labelFunc func(time.Time) string

	switch filter.Period {
	case "today":
		groupBy = "DATE_TRUNC('hour', orders.order_date)"
		orderBy = "DATE_TRUNC('hour', orders.order_date)"
		labelFunc = func(t time.Time) string {
			hour := t.Hour() - (t.Hour() % 4)
			return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location()).Format("03PM")
		}
	case "weekly":
		groupBy = "DATE_TRUNC('week', orders.order_date)"
		orderBy = "DATE_TRUNC('week', orders.order_date)"
		labelFunc = func(t time.Time) string {
			weekStart := t.AddDate(0, 0, -int(t.Weekday())+1)
			return weekStart.Format("Jan 02")
		}
	case "monthly":
		groupBy = "DATE_TRUNC('month', orders.order_date)"
		orderBy = "DATE_TRUNC('month', orders.order_date)"
		labelFunc = func(t time.Time) string { return t.Format("Jan 2006") }
	case "daily":
		groupBy = "DATE_TRUNC('day', orders.order_date)"
		orderBy = "DATE_TRUNC('day', orders.order_date)"
		labelFunc = func(t time.Time) string { return t.Format("2006-01-02") }
	case "yearly":
		groupBy = "DATE_TRUNC('year', orders.order_date)"
		orderBy = "DATE_TRUNC('year', orders.order_date)"
		labelFunc = func(t time.Time) string { return t.Format("2006") }
	case "custom":
		daysDiff := endDate.Sub(*startDate).Hours() / 24
		if daysDiff <= 1 {
			groupBy = "DATE_TRUNC('hour', orders.order_date)"
			orderBy = "DATE_TRUNC('hour', orders.order_date)"
			labelFunc = func(t time.Time) string {
				hour := t.Hour() - (t.Hour() % 4)
				return time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location()).Format("03PM")
			}
		} else if daysDiff <= 7 {
			groupBy = "DATE_TRUNC('day', orders.order_date)"
			orderBy = "DATE_TRUNC('day', orders.order_date)"
			labelFunc = func(t time.Time) string { return t.Format("2006-01-02") }
		} else if daysDiff <= 180 {
			groupBy = "DATE_TRUNC('week', orders.order_date)"
			orderBy = "DATE_TRUNC('week', orders.order_date)"
			labelFunc = func(t time.Time) string {
				weekStart := t.AddDate(0, 0, -int(t.Weekday())+1)
				return weekStart.Format("Jan 02")
			}
		} else {
			groupBy = "DATE_TRUNC('month', orders.order_date)"
			orderBy = "DATE_TRUNC('month', orders.order_date)"
			labelFunc = func(t time.Time) string { return t.Format("Jan 2006") }
		}
	default:
		groupBy = "DATE_TRUNC('month', orders.order_date)"
		orderBy = "DATE_TRUNC('month', orders.order_date)"
		labelFunc = func(t time.Time) string { return t.Format("Jan 2006") }
	}

	query := db.Model(&models.Order{}).
		Select(fmt.Sprintf(`
            %s as month,
            SUM(oi.total_amount) as total_amount,
            SUM(oi.total_revenue) - SUM(COALESCE(orders.coupon_discount_amount, 0)) as revenue,
            COUNT(DISTINCT orders.id) as orders
        `, groupBy)).
		Joins(`
            JOIN (
                SELECT 
                    order_id,
                    SUM(sub_total + tax) as total_amount,
                    SUM(total) as total_revenue
                FROM order_items 
                WHERE order_status = ?
                GROUP BY order_id
            ) oi ON oi.order_id = orders.id
        `, filter.Status)

	if startDate != nil && endDate != nil {
		query = query.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}

	var tempResults []struct {
		Month       time.Time
		TotalAmount float64
		Revenue     float64
		Orders      int
	}
	query.Group(groupBy).Order(orderBy + " ASC").Scan(&tempResults)

	if filter.Period == "daily" {
		results = make([]SalesOverviewDTO, 8)
		today := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
		start := today.AddDate(0, 0, -7)
		dateMap := make(map[string]SalesOverviewDTO)

		for _, temp := range tempResults {
			label := labelFunc(temp.Month)
			dateMap[label] = SalesOverviewDTO{
				Month:       label,
				TotalAmount: temp.TotalAmount,
				Revenue:     temp.Revenue,
				Orders:      temp.Orders,
			}
		}

		for i := 0; i < 8; i++ {
			day := start.AddDate(0, 0, i)
			label := labelFunc(day)
			if data, exists := dateMap[label]; exists {
				results[i] = data
			} else {
				results[i] = SalesOverviewDTO{
					Month: label,
				}
			}
		}
	} else if filter.Period == "today" {
		results = make([]SalesOverviewDTO, 6)
		today := time.Date(now.BeginningOfHour().Year(), now.BeginningOfHour().Month(), now.BeginningOfHour().Day(), 0, 0, 0, 0, now.BeginningOfHour().Location())
		for i := 0; i < 6; i++ {
			intervalStart := today.Add(time.Hour * time.Duration(i*4))
			results[i] = SalesOverviewDTO{
				Month: labelFunc(intervalStart),
			}
		}
		for _, temp := range tempResults {
			hour := temp.Month.Hour()
			index := hour / 4
			if index < 6 {
				results[index].TotalAmount += temp.TotalAmount
				results[index].Revenue += temp.Revenue
				results[index].Orders += temp.Orders
			}
		}
	} else if filter.Period == "weekly" || filter.Period == "monthly" {
		expectedCount := 8
		results = make([]SalesOverviewDTO, expectedCount)
		start := *startDate
		interval := time.Duration(7*24) * time.Hour
		if filter.Period == "monthly" {
			interval = 30 * 24 * time.Hour
		}

		tempMap := make(map[string]SalesOverviewDTO)
		for _, temp := range tempResults {
			label := labelFunc(temp.Month)
			tempMap[label] = SalesOverviewDTO{
				Month:       label,
				TotalAmount: temp.TotalAmount,
				Revenue:     temp.Revenue,
				Orders:      temp.Orders,
			}
		}

		for i := 0; i < expectedCount; i++ {
			var periodStart time.Time
			if filter.Period == "weekly" {
				periodStart = start.Add(time.Duration(i) * interval)
				periodStart = periodStart.AddDate(0, 0, -int(periodStart.Weekday())+1)
			} else {
				periodStart = start.AddDate(0, i, 0)
				periodStart = time.Date(periodStart.Year(), periodStart.Month(), 1, 0, 0, 0, 0, periodStart.Location())
			}
			label := labelFunc(periodStart)
			if data, exists := tempMap[label]; exists {
				results[i] = data
			} else {
				results[i] = SalesOverviewDTO{
					Month: label,
				}
			}
		}
	} else {
		for _, temp := range tempResults {
			results = append(results, SalesOverviewDTO{
				Month:       labelFunc(temp.Month),
				TotalAmount: temp.TotalAmount,
				Revenue:     temp.Revenue,
				Orders:      temp.Orders,
			})
		}
	}

	return results
}

func getCategorySalesData(db *gorm.DB, startDate, endDate *time.Time, status string) []CategorySalesDTO {
	var results []CategorySalesDTO

	query := db.Model(&models.OrderItem{}).
		Select(`
            product_category as category,
            SUM(order_items.total) - SUM(COALESCE(orders.coupon_discount_amount, 0)) as amount,
            COUNT(*) as count
        `).
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Where("order_items.order_status = ?", status)

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

	countQuery := db.Model(&models.OrderItem{}).
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Joins("JOIN user_auths ON orders.user_id = user_auths.id").
		Where("order_items.order_status = ?", filter.Status)

	if startDate != nil && endDate != nil {
		countQuery = countQuery.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}
	countQuery.Count(&totalCount)

	dataQuery := db.Model(&models.OrderItem{}).
		Select(`
            orders.order_uid as order_number,
            CONCAT(user_auths.full_name) as customer,
            orders.order_date as date,
            SUM(order_items.total) - COALESCE(orders.coupon_discount_amount, 0) as amount,
            order_items.order_status as status
        `).
		Joins("JOIN orders ON order_items.order_id = orders.id").
		Joins("JOIN user_auths ON orders.user_id = user_auths.id").
		Where("order_items.order_status = ?", filter.Status)

	if startDate != nil && endDate != nil {
		dataQuery = dataQuery.Where("orders.order_date BETWEEN ? AND ?", startDate, endDate)
	}

	offset := (filter.Page - 1) * filter.PageSize
	dataQuery.Group("orders.order_uid, user_auths.full_name, orders.order_date, order_items.order_status, orders.coupon_discount_amount").
		Order("orders.order_date DESC").
		Limit(filter.PageSize).
		Offset(offset).
		Scan(&orders)

	for i := range orders {
		orders[i].OrderID = fmt.Sprintf("ORD-%s", orders[i].OrderNumber[0:8])
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

	f.SetCellValue(summarySheet, "A1", "Sales Summary")
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
	pdf.Cell(190, 10, "Sales Report")
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
