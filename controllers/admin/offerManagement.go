package controllers

import (
	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func AddProductOffer(c *gin.Context) {
	productID, parseErr := strconv.Atoi(c.PostForm("product_id"))
	if parseErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	offerName := c.PostForm("offer_name")
	offerDetails := c.PostForm("offer_details")
	startDate := c.PostForm("offer_start_date")
	endDate := c.PostForm("offer_end_date")
	offerPercentage := c.PostForm("offer_percentage")

	percentageValue, percentageErr := strconv.ParseFloat(offerPercentage, 64)
	if percentageErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid percentage", "Invalid percentage", "")
		return
	}

	startDates, startErr := time.Parse("2006-01-02", startDate)
	if startErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid date format", "Invalid date format", "")
		return
	}

	endDates, endErr := time.Parse("2006-01-02", endDate)
	if endErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid date format", "Invalid date format", "")
		return
	}

	productOffer := models.ProductOffer{
		OfferName:       offerName,
		OfferDetails:    offerDetails,
		StartDate:       startDates,
		EndDate:         endDates,
		OfferPercentage: percentageValue,
		ProductID:       uint(productID),
	}

	if err := config.DB.Create(&productOffer).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product", "Failed to save product", "")
		return
	}

	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(productID)
	c.Redirect(http.StatusFound, redirectURL)
}

func AddCategoryOffer(c *gin.Context) {
	var addOfferInput struct {
		Id               string `json:"categoryId"`
		OfferName        string `json:"offerName"`
		OfferDescription string `json:"offerDescription"`
		OfferValue       string `json:"discount"`
		StartDate        string `json:"startDate"`
		EndDate          string `json:"endDate"`
	}
	if err := c.ShouldBindJSON(&addOfferInput); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Enter Details Correctly", "")
		return
	}
	offerValue, err := strconv.ParseFloat(addOfferInput.OfferValue, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Offer Value", "Offer Value must be a valid number", "")
		return
	}

	categoryId, err := strconv.Atoi(addOfferInput.Id)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Category ID", "Category ID must be a number", "")
		return
	}

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, addOfferInput.StartDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Start Date", "Enter a valid start date", "")
		return
	}

	endDate, err := time.Parse(layout, addOfferInput.EndDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "Enter a valid end date", "")
		return
	}

	var existingOffers models.OfferByCategory
	if err := config.DB.First(&existingOffers, "category_id = ?", categoryId).Error; err == nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Can't Add Offer", "Can't Add Offer.If You Want TO Add Offer Delete Existing Offer", "")
		return
	}

	if time.Now().After(startDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Starting Date", "Start date must be in the future", "")
		return
	}

	if time.Now().After(endDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "End date must be in the future", "")
		return
	}

	if startDate.After(endDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Date Range", "Start date must be before end date", "")
		return
	}

	addOffer := models.OfferByCategory{
		CategoryOfferName:       addOfferInput.OfferName,
		CategoryOfferPercentage: offerValue,
		OfferDescription:        addOfferInput.OfferDescription,
		CategoryID:              uint(categoryId),
		OfferStatus:             "Active",
		StartDate:               startDate,
		EndDate:                 endDate,
	}

	if err := config.DB.Create(&addOffer).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create offer", "Failed to create offer", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Offer Added Successfully",
		"code":    http.StatusOK,
	})
}

func UpdateCategoryOffer(c *gin.Context) {
	var editOfferInput struct {
		Id               string `json:"offerId"`
		OfferName        string `json:"offerName"`
		OfferDescription string `json:"offerDescription"`
		OfferValue       string `json:"discount"`
		StartDate        string `json:"startDate"`
		EndDate          string `json:"endDate"`
	}

	if err := c.ShouldBindJSON(&editOfferInput); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Enter Details Correctly", "")
		return
	}

	offerId, err := strconv.Atoi(editOfferInput.Id)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Offer ID", "Offer ID must be a number", "")
		return
	}

	offerValue, err := strconv.ParseFloat(editOfferInput.OfferValue, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Offer Value", "Offer Value must be a valid number", "")
		return
	}

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, editOfferInput.StartDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Start Date", "Enter a valid start date", "")
		return
	}

	endDate, err := time.Parse(layout, editOfferInput.EndDate)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "Enter a valid end date", "")
		return
	}

	var offer models.OfferByCategory
	if err := config.DB.First(&offer, offerId).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Offer Not Found", "Offer Not Found", "")
		return
	}

	if time.Now().After(startDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Starting Date", "Start date must be in the future", "")
		return
	}

	if time.Now().After(endDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid End Date", "End date must be in the future", "")
		return
	}

	if startDate.After(endDate) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Date Range", "Start date must be before end date", "")
		return
	}

	offer.CategoryOfferName = editOfferInput.OfferName
	offer.CategoryOfferPercentage = offerValue
	offer.OfferDescription = editOfferInput.OfferDescription
	offer.OfferStatus = "Active"
	offer.StartDate = startDate
	offer.EndDate = endDate

	if err := config.DB.Save(&offer).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update offer", "Something Went Wrong", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Offer Updated Successfully",
		"code":    http.StatusOK,
	})
}

func DeleteCategoryOffer(c *gin.Context) {
	var DeleteOfferInput struct{
		OfferId string `json:"offerId"`
	}

	if err := c.ShouldBindJSON(&DeleteOfferInput); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Something Went Wrong", "")
		return
	}

	offerId, err := strconv.Atoi(DeleteOfferInput.OfferId)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Offer ID", "Offer ID must be a number", "")
		return
	}

	var Offer models.OfferByCategory
	if err := config.DB.First(&Offer, offerId).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Offer not found", "Offer not found", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&Offer).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed To Delete Offer", "Delete Offer Failed", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Offer deleted successfully",
		"code":    200,
	})
}
