package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func AddProductOffer(c *gin.Context) {
	productID, parseErr := strconv.Atoi(c.PostForm("product_id"))
	if parseErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid price")
		return
	}

	percentageValue, percentageErr := strconv.ParseFloat(offerPercentage, 64)
	if percentageErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid percentage")
		return
	}

	startDates, startErr := time.Parse("2006-01-02", startDate)
	if startErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid date format")
		return
	}

	endDates, endErr := time.Parse("2006-01-02", endDate)
	if endErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid date format")
		return
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
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product")
		return
	}

	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(productID)
	c.Redirect(http.StatusFound, redirectURL)
}
