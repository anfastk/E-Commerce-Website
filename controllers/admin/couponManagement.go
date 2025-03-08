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
)

func ShowCoupon(c *gin.Context) {
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
	config.DB.Find(&category)
	c.HTML(http.StatusOK, "couponManagement.html", gin.H{
		"Coupons":  couponsDetail,
		"Count":    count,
		"Category": category,
	})
}

func AddCoupon(c *gin.Context) {
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Something Went Wrong", "")
		return
	}
	discountValue, err := strconv.ParseFloat(couponInput.DiscountValue, 64)
	maxDiscountValue, err := strconv.ParseFloat(couponInput.MaxDiscountValue, 64)
	minOrderValue, err := strconv.ParseFloat(couponInput.MinOrdervalue, 64)
	maxUseCount, err := strconv.Atoi(couponInput.MaxUseCount)
	layout := "2006-01-02"
	validFrom, err := time.Parse(layout, couponInput.ValidFrom)
	expirationDate, err := time.Parse(layout, couponInput.ExpirationDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Something Went Wrong", "")
		return
	}
	if validFrom.Before(time.Now().Truncate(24 * time.Hour)) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Starting Date", "Start date must be today or in the future", "")
		return
	}
	if time.Now().After(expirationDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "End date must be in the future", "")
		return
	}
	if validFrom.After(expirationDate) {
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
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to create coupon code already exist", "Failed to create coupon code already exist", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon Added Successfully",
		"code":    http.StatusOK,
	})

}

func DeleteCoupon(c *gin.Context) {
	couponID := c.Param("id")
	var coupon models.Coupon
	if err := config.DB.First(&coupon, couponID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Coupon not found", "Coupon not found", "")
		return
	}

	coupon.Status = "Deleted"

	if err := config.DB.Save(&coupon).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete coupon", "Failed to delete coupon", "")
		return
	}

	if err := config.DB.Delete(&coupon).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete coupon", "Failed to delete coupon", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon deleted successfully",
		"code":    200,
	})
}

func CouponDetails(c *gin.Context) {
	couponID := c.Param("id")
	var coupon models.Coupon
	if err := config.DB.First(&coupon, couponID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Coupon not found", "Coupon not found", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon Fetched Successfully",
		"code":    200,
		"coupon":  coupon,
	})
}

func UpdateCoupon(c *gin.Context) {
	couponID := c.Param("id")

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
		fmt.Println(request)
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request", "")
		return
	}

	var coupon models.Coupon
	if err := config.DB.First(&coupon, couponID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Coupon not found", "Coupon not found", "")
		return
	}

	validFrom, err := time.Parse("2006-01-02", request.ValidDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid valid date format", "Something Went Wrong", "")
		return
	}

	expiryDate, err := time.Parse("2006-01-02", request.ExpiryDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid expiry date format", "Something Went Wrong", "")
		return
	}
	discountValue, err := strconv.ParseFloat(request.DiscountValue, 64)
	maxDiscountValue, err := strconv.ParseFloat(request.MaxDiscountValue, 64)
	minOrderValue, err := strconv.ParseFloat(request.MinOrderValue, 64)
	maxUseCount, err := strconv.Atoi(request.UsageLimit)
	if err != nil {
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Starting Date", "Start date must be today or in the future", "")
		return
	}
	if time.Now().After(expiryDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "End date must be in the future", "")
		return
	}
	if validFrom.After(expiryDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Date Range", "Start date must be before end date", "")
		return
	}
	coupon.Status = "Active"
	if validFrom.After(time.Now()) {
		coupon.Status = "Scheduled"
	}

	if err := config.DB.Save(&coupon).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update coupon", "Something Went Wrong", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Coupon updated successfully",
		"code":    200,
	})
}
