package controllers

import (
	"math"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
)

func DashboardHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "adminDashboard.html", gin.H{
		"title": "Admin Dashboard",
	})
}

func StatsHandler(c *gin.Context) {
	period := c.Query("period")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var startTime, endTime time.Time
	now := time.Now().UTC()

	switch period {
	case "daily":
		startTime = now.Truncate(24 * time.Hour)
		endTime = now
	case "weekly":
		startTime = now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		endTime = now
	case "yearly":
		startTime = now.AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endTime = now
	case "custom":
		var err1, err2 error
		startTime, err1 = time.Parse("2006-01-02", startDate)
		endTime, err2 = time.Parse("2006-01-02", endDate)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endTime = endTime.Add(24*time.Hour - time.Second)
	default:
		startTime = now.AddDate(0, -1, 0).Truncate(24 * time.Hour)
		endTime = now
	}

	var orders []models.Order
	var users []models.UserAuth
	var products []models.ProductDetail
	var categories []models.Categories

	if err := config.DB.Where("order_date BETWEEN ? AND ?", startTime, endTime).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}
	totalAmount := 0.0
	for _, order := range orders {
		totalAmount += order.TotalAmount
	}

	var revenue float64
	config.DB.Raw(`
		SELECT SUM(total_amount) 
		FROM orders 
		WHERE order_date BETWEEN ? AND ? 
		AND id IN (SELECT order_id FROM order_items WHERE order_status = 'Delivered')`,
		startTime, endTime).Scan(&revenue)

	config.DB.Where("created_at BETWEEN ? AND ?", startTime, endTime).Find(&users)
	userCount := len(users)

	config.DB.Find(&products)
	config.DB.Find(&categories)

	var cancelledCount int64
	config.DB.Model(&models.OrderItem{}).Where("order_status = 'Cancelled' AND created_at BETWEEN ? AND ?",
		startTime, endTime).Count(&cancelledCount)

	var totalDiscount, couponDiscount float64
	config.DB.Raw("SELECT SUM(total_discount) FROM orders WHERE order_date BETWEEN ? AND ?",
		startTime, endTime).Scan(&totalDiscount)
	config.DB.Raw("SELECT SUM(coupon_discount_amount) FROM orders WHERE order_date BETWEEN ? AND ? AND is_coupon_applied = true",
		startTime, endTime).Scan(&couponDiscount)

	avgOrderValue := 0.0
	if len(orders) > 0 {
		avgOrderValue = totalAmount / float64(len(orders))
	}

	c.JSON(http.StatusOK, gin.H{
		"total_amount":          totalAmount,
		"order_count":           len(orders),
		"revenue":               revenue,
		"avg_order_value":       math.Round(avgOrderValue*100) / 100,
		"user_count":            userCount,
		"products_count":        len(products),
		"category_count":        len(categories),
		"cancelled_order_count": cancelledCount,
		"total_discount":        totalDiscount,
		"coupon_discount":       couponDiscount,
	})
}

func OrdersHandler(c *gin.Context) {
	var orderItems []models.OrderItem
	if err := config.DB.Preload("UserAuth").Order("created_at desc").Limit(10).Find(&orderItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	ordersResponse := make([]map[string]interface{}, 0)
	for _, item := range orderItems {
		ordersResponse = append(ordersResponse, map[string]interface{}{
			"order_id": item.OrderUID,
			"customer": item.UserAuth.FullName,
			"products": item.ProductName,
			"amount":   item.Total,
			"date":     item.CreatedAt.Format("2006-01-02"),
			"status":   item.OrderStatus,
		})
	}

	c.JSON(http.StatusOK, gin.H{"orders": ordersResponse})
}

func ChartsHandler(c *gin.Context) {
	period := c.Query("period")
	section := c.Query("section")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var startTime, endTime time.Time
	now := time.Now().UTC()

	switch period {
	case "daily":
		startTime = now.Truncate(24 * time.Hour)
		endTime = now
	case "weekly":
		startTime = now.AddDate(0, 0, -7).Truncate(24 * time.Hour)
		endTime = now
	case "yearly":
		startTime = now.AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endTime = now
	case "custom":
		var err1, err2 error
		startTime, err1 = time.Parse("2006-01-02", startDate)
		endTime, err2 = time.Parse("2006-01-02", endDate)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}
		endTime = endTime.Add(24*time.Hour - time.Second)
	default:
		startTime = now.AddDate(0, -1, 0).Truncate(24 * time.Hour)
		endTime = now
	}

	switch section {
	case "sales":
		var results []struct {
			Date  time.Time
			Total float64
		}
		config.DB.Raw(`
			SELECT DATE(order_date) as date, SUM(total_amount) as total 
			FROM orders 
			WHERE order_date BETWEEN ? AND ? 
			GROUP BY DATE(order_date)
			ORDER BY date ASC`,
			startTime, endTime).Scan(&results)

		salesData := make(map[string]float64)
		for _, result := range results {
			salesData[result.Date.Format("2006-01-02")] = result.Total
		}
		c.JSON(http.StatusOK, gin.H{"sales": salesData})

	case "products":
		var results []struct {
			ProductName string
			Quantity    int
		}
		config.DB.Raw(`
			SELECT product_name, SUM(quantity) as quantity 
			FROM order_items 
			WHERE created_at BETWEEN ? AND ? 
			GROUP BY product_name 
			ORDER BY quantity DESC 
			LIMIT 10`,
			startTime, endTime).Scan(&results)

		productsData := make(map[string]int)
		for _, result := range results {
			productsData[result.ProductName] = result.Quantity
		}
		c.JSON(http.StatusOK, gin.H{"products": productsData})

	case "categories":
		var results []struct {
			CategoryName string
			Total        float64
		}
		config.DB.Raw(`
			SELECT c.name as category_name, SUM(oi.total) as total 
			FROM order_items oi 
			JOIN product_details pd ON oi.product_variant_id = pd.id 
			JOIN categories c ON pd.category_id = c.id 
			WHERE oi.created_at BETWEEN ? AND ? 
			GROUP BY c.name 
			ORDER BY total DESC 
			LIMIT 10`,
			startTime, endTime).Scan(&results)

		categoriesData := make(map[string]float64)
		for _, result := range results {
			categoriesData[result.CategoryName] = result.Total
		}
		c.JSON(http.StatusOK, gin.H{"categories": categoriesData})

	case "brands":
		var results []struct {
			BrandName string
			Total     float64
		}
		config.DB.Raw(`
			SELECT pd.brand_name, SUM(oi.total) as total 
			FROM order_items oi 
			JOIN product_details pd ON oi.product_variant_id = pd.id 
			WHERE oi.created_at BETWEEN ? AND ? 
			GROUP BY pd.brand_name 
			ORDER BY total DESC 
			LIMIT 10`,
			startTime, endTime).Scan(&results)

		brandsData := make(map[string]float64)
		for _, result := range results {
			brandsData[result.BrandName] = result.Total
		}
		c.JSON(http.StatusOK, gin.H{"brands": brandsData})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid section"})
	}
}
