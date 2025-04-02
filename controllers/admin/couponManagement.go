package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ShowCoupon(c *gin.Context) {
	logger.Log.Info("Requested to Show Coupons")

	type couponDetail struct {
		ID             uint
		CouponCode     string
		Discription    string
		DiscountValue  float64
		CouponType     string
		UsersUsedCount int
		MaxUseCount    int
		ExpirationDate string
		ApplicableFor  string
		Status         string
	}

	var coupons []models.Coupon
	if err := config.DB.Find(&coupons).Error; err != nil {
		logger.Log.Error("Failed to fetch coupons", zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Coupons Not Found", "Something Went Wrong", "")
		return
	}

	var couponsDetail []couponDetail
	for _, coupon := range coupons {
		cpn := couponDetail{
			ID:             coupon.ID,
			CouponCode:     coupon.CouponCode,
			Discription:    coupon.Discription,
			DiscountValue:  coupon.DiscountValue,
			CouponType:     coupon.CouponType,
			UsersUsedCount: coupon.UsersUsedCount,
			MaxUseCount:    coupon.MaxUseCount,
			Status:         coupon.Status,
			ApplicableFor:  coupon.ApplicableFor,
			ExpirationDate: coupon.ExpirationDate.In(time.UTC).Format("2006-01-02"),
		}
		couponsDetail = append(couponsDetail, cpn)
	}

	count := len(coupons)
	var category []models.Categories
	if err := config.DB.Find(&category).Error; err != nil {
		logger.Log.Error("Failed to fetch categories", zap.Error(err))
	}

	logger.Log.Info("Coupons displayed successfully", zap.Int("count", count))
	c.HTML(http.StatusOK, "couponManagement.html", gin.H{
		"Coupons":  couponsDetail,
		"Count":    count,
		"Category": category,
	})
}

func AddCoupon(c *gin.Context) {
	logger.Log.Info("Requested to Add Coupon")

	var couponInput struct {
		CouponCode        string `json:"code" binding:"required"`
		Discription       string `json:"description" binding:"required"`
		CouponType        string `json:"type" binding:"required"`
		DiscountValue     string `json:"value" binding:"required"`
		ApplicableProduct string `json:"appliedTo" binding:"required"`
		MinOrdervalue     string `json:"minOrderValue" binding:"required"`
		MaxDiscountValue  string `json:"maxDiscount" binding:"required"`
		MaxUseCount       string `json:"usageLimit" binding:"required"`
		ValidFrom         string `json:"validDate" binding:"required"`
		ExpirationDate    string `json:"expiryDate" binding:"required"`
	}

	if err := c.ShouldBindJSON(&couponInput); err != nil {
		logger.Log.Error("Invalid request payload", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Something Went Wrong", "")
		return
	}

	discountValue, err := strconv.ParseFloat(couponInput.DiscountValue, 64)
	if (couponInput.CouponType != "Fixed" && (discountValue > 90 || discountValue < 1)) || discountValue < 1 {
		logger.Log.Error("Invalid data.Discount value should be greater than 0 and less than 90")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data. Discount value should be greater than 0 and less than 90", "Invalid data. Discount value should be greater than 0 and less than 90", "")
		return
	}

	maxDiscountValue, err := strconv.ParseFloat(couponInput.MaxDiscountValue, 64)
	if maxDiscountValue < 1 {
		logger.Log.Error("Invalid data.Max discount value should be greater than 0")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data.Max discount value should be greater than 0", "Invalid data.Max discount value should be greater yhan 0 ", "")
		return
	}
	minOrderValue, err := strconv.ParseFloat(couponInput.MinOrdervalue, 64)
	if minOrderValue < 1 {
		logger.Log.Error("Invalid data.Min order value should be greater than 0")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data.Min order value should be greater than 0", "Invalid data.Min order value should be greater yhan 0 ", "")
		return
	}
	maxUseCount, err := strconv.Atoi(couponInput.MaxUseCount)
	if maxUseCount < 1 {
		logger.Log.Error("Invalid data.Amount should be greater than 0")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data.Max use count should be greater than 0", "Invalid data.Max use count should be greater yhan 0 ", "")
		return
	}
	layout := "2006-01-02"
	validFrom, err := time.Parse(layout, couponInput.ValidFrom)
	expirationDate, err := time.Parse(layout, couponInput.ExpirationDate)
	if err != nil {
		logger.Log.Error("Invalid data format", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Something Went Wrong", "")
		return
	}

	if validFrom.Before(time.Now().Truncate(24 * time.Hour)) {
		logger.Log.Error("Invalid starting date - date in past")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Starting Date", "Start date must be today or in the future", "")
		return
	}
	if time.Now().After(expirationDate) {
		logger.Log.Error("Invalid end date - date in past")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "End date must be in the future", "")
		return
	}
	if validFrom.After(expirationDate) {
		logger.Log.Error("Invalid date range - start after end")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Date Range", "Start date must be before end date", "")
		return
	}

	status := "Active"
	if validFrom.After(time.Now()) {
		status = "Scheduled"
	}

	couponFixed := models.Coupon{
		CouponCode:       strings.ToUpper(couponInput.CouponCode),
		Discription:      couponInput.Discription,
		DiscountValue:    discountValue,
		MaxDiscountValue: maxDiscountValue,
		MinOrderValue:    minOrderValue,
		MaxUseCount:      maxUseCount,
		ValidFrom:        validFrom,
		ExpirationDate:   expirationDate,
		IsFixedCoupon:    couponInput.CouponType == "Fixed",
		CouponType:       couponInput.CouponType,
		ApplicableFor:    couponInput.ApplicableProduct,
		Status:           status,
	}

	if err := config.DB.Create(&couponFixed).Error; err != nil {
		logger.Log.Error("Failed to create coupon", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to create coupon code already exist", "Failed to create coupon code already exist", "")
		return
	}

	logger.Log.Info("Coupon added successfully", zap.String("couponCode", couponFixed.CouponCode))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon Added Successfully",
		"code":    http.StatusOK,
	})
}

func DeleteCoupon(c *gin.Context) {
	couponID := c.Param("id")
	logger.Log.Info("Requested to Delete Coupon", zap.String("couponID", couponID))

	var coupon models.Coupon
	if err := config.DB.First(&coupon, couponID).Error; err != nil {
		logger.Log.Error("Coupon not found", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Coupon not found", "Coupon not found", "")
		return
	}

	coupon.Status = "Deleted"

	if err := config.DB.Save(&coupon).Error; err != nil {
		logger.Log.Error("Failed to update coupon status", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete coupon", "Failed to delete coupon", "")
		return
	}

	if err := config.DB.Delete(&coupon).Error; err != nil {
		logger.Log.Error("Failed to delete coupon", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete coupon", "Failed to delete coupon", "")
		return
	}

	logger.Log.Info("Coupon deleted successfully", zap.String("couponID", couponID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon deleted successfully",
		"code":    200,
	})
}

func CouponDetails(c *gin.Context) {
	couponID := c.Param("id")
	logger.Log.Info("Requested Coupon Details", zap.String("couponID", couponID))

	var coupon models.Coupon
	if err := config.DB.First(&coupon, couponID).Error; err != nil {
		logger.Log.Error("Failed to fetch coupon", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Coupon not found", "Coupon not found", "")
		return
	}

	logger.Log.Info("Coupon details fetched successfully", zap.String("couponID", couponID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon Fetched Successfully",
		"code":    200,
		"coupon":  coupon,
	})
}

func UpdateCoupon(c *gin.Context) {
	couponID := c.Param("id")
	logger.Log.Info("Requested to Update Coupon", zap.String("couponID", couponID))

	var request struct {
		CouponCode       string `json:"code"`
		Description      string `json:"description"`
		CouponType       string `json:"type"`
		DiscountValue    string `json:"value"`
		AppliedTo        string `json:"appliedTo"`
		MinOrderValue    string `json:"minOrderValue"`
		MaxDiscountValue string `json:"maxDiscount"`
		UsageLimit       string `json:"usageLimit"`
		ValidDate        string `json:"validDate"`
		ExpiryDate       string `json:"expiryDate"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Log.Error("Invalid request data", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request", "")
		return
	}

	var coupon models.Coupon
	if err := config.DB.First(&coupon, couponID).Error; err != nil {
		logger.Log.Error("Coupon not found", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Coupon not found", "Coupon not found", "")
		return
	}

	validFrom, err := time.Parse("2006-01-02", request.ValidDate)
	if err != nil {
		logger.Log.Error("Invalid valid date format", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid valid date format", "Something Went Wrong", "")
		return
	}

	expiryDate, err := time.Parse("2006-01-02", request.ExpiryDate)
	if err != nil {
		logger.Log.Error("Invalid expiry date format", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid expiry date format", "Something Went Wrong", "")
		return
	}

	discountValue, err := strconv.ParseFloat(request.DiscountValue, 64)
	maxDiscountValue, err := strconv.ParseFloat(request.MaxDiscountValue, 64)
	minOrderValue, err := strconv.ParseFloat(request.MinOrderValue, 64)
	maxUseCount, err := strconv.Atoi(request.UsageLimit)
	if err != nil {
		logger.Log.Error("Invalid data format", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Something Went Wrong", "")
		return
	}

	coupon.CouponCode = strings.ToUpper(request.CouponCode)
	coupon.Discription = request.Description
	coupon.CouponType = request.CouponType
	coupon.DiscountValue = discountValue
	coupon.ApplicableFor = request.AppliedTo
	coupon.MinOrderValue = minOrderValue
	coupon.MaxDiscountValue = maxDiscountValue
	coupon.MaxUseCount = maxUseCount
	coupon.ValidFrom = validFrom
	coupon.ExpirationDate = expiryDate
	coupon.IsFixedCoupon = request.CouponType == "Fixed"

	if validFrom.Before(time.Now().Truncate(24 * time.Hour)) {
		logger.Log.Error("Invalid starting date - date in past")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Starting Date", "Start date must be today or in the future", "")
		return
	}
	if time.Now().After(expiryDate) {
		logger.Log.Error("Invalid end date - date in past")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "End date must be in the future", "")
		return
	}
	if validFrom.After(expiryDate) {
		logger.Log.Error("Invalid date range - start after end")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Date Range", "Start date must be before end date", "")
		return
	}

	coupon.Status = "Active"
	if validFrom.After(time.Now()) {
		coupon.Status = "Scheduled"
	}

	if err := config.DB.Save(&coupon).Error; err != nil {
		logger.Log.Error("Failed to update coupon", zap.String("couponID", couponID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update coupon", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Coupon updated successfully", zap.String("couponID", couponID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon updated successfully",
		"code":    200,
	})
}
