package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
)

func AddProductOffer(c *gin.Context) {
	productID, parseErr := strconv.Atoi(c.PostForm("product_id"))
	if parseErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}
	offerName := c.PostForm("offer_name")
	offerDetails := c.PostForm("offer_details")
	startDate := c.PostForm("offer_start_date")
	endDate := c.PostForm("offer_end_date")
	offerAmount := c.PostForm("offer_amount")
	offerPercentage := c.PostForm("offer_percentage")

	priceValue, amountErr := strconv.ParseFloat(offerAmount, 64)
	if amountErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid price",
			"code":   400,
		})
		return
	}
	percentageValue, percentageErr := strconv.ParseFloat(offerPercentage, 64)
	if percentageErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid percentage",
			"code":   400,
		})
		return
	}
	startDates, startErr := time.Parse("2006-01-02", startDate)
	if startErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid date format",
			"code":   400,
		})
	}

	endDates, endErr := time.Parse("2006-01-02", endDate)
	if endErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid date format",
			"code":   400,
		})
	}

	productOffer := models.ProductOffer{
		OfferName:       offerName,
		OfferDetails:    offerDetails,
		StartDate:       startDates,
		EndDate:         endDates,
		OfferAmount:     priceValue,
		OfferPercentage: percentageValue,
		ProductID:       uint(productID),
	}

	if err := config.DB.Create(&productOffer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to save product",
			"code":   500,
		})
	}
	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(int(productID))
	c.Redirect(http.StatusFound, redirectURL)
}
