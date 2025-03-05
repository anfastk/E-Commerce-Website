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

func ListCategory(c *gin.Context) {
	type categoryResponce struct {
		ID           uint
		Name         string
		Description  string
		Status       string
		ProductCount int
	}

	var categories []models.Categories
	if err := config.DB.Find(&categories).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories", "Failed to fetch categories", "")
		return
	}
	var (
		activeCount       int
		totalProductCount int
	)
	var responce []categoryResponce
	for _, category := range categories {
		var products []models.ProductVariantDetails
		config.DB.Find(&products, "category_id = ?", category.ID)
		count := len(products)
		row := categoryResponce{
			ID:           category.ID,
			Name:         category.Name,
			Description:  category.Description,
			Status:       category.Status,
			ProductCount: count,
		}
		totalProductCount += count
		responce = append(responce, row)
		if category.Status == "Active" {
			activeCount++
		}
	}
	categoryCount := len(categories)
	c.HTML(http.StatusOK, "categoryManagement.html", gin.H{
		"categories":          responce,
		"CategoryCount":       categoryCount,
		"ActiveCategoryCount": activeCount,
		"TotalProductCount":   totalProductCount,
	})
}

func AddCategory(c *gin.Context) {
	c.Request.ParseForm()

	var category models.Categories
	if err := c.ShouldBind(&category); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Invalid input data", "")
		return
	}

	if category.Name == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Category name is required", "Category name is required", "")
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create category", "Failed to create category", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Category created successfully",
		"code":    200,
	})
}

func EditCategory(c *gin.Context) {
	categoryID := c.Param("id")

	if categoryID == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Category ID is missing", "Category ID is missing", "")
		return
	}

	if _, err := strconv.ParseInt(categoryID, 10, 64); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Category ID", "Invalid Category ID", "")
		return
	}

	var category models.Categories

	if err := config.DB.First(&category, categoryID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Category not found", "Category not found", "")
		return
	}

	if err := c.Bind(&category); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to bind form data", "Failed to bind form data", "")
		return
	}

	if err := config.DB.Model(&category).Where("id = ?", categoryID).Updates(category).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update category", "Failed to update category", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Category updated successfully",
		"code":    200,
	})
}

func DeleteCategory(c *gin.Context) {
	var category models.Categories
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid category ID", "Invalid category ID", "")
		return
	}

	tx := config.DB.Begin()

	// Find the category
	if err := tx.Unscoped().First(&category, categoryID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Category not found", "Category not found", "")
		return
	}

	category.IsDeleted = !category.IsDeleted
	if category.IsDeleted {
		category.Status = "Blocked"
	} else {
		category.Status = "Active"
	}

	if err := tx.Save(&category).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update category status", "Failed to update category status", "")
		return
	}

	var product models.ProductDetail
	if err := tx.Unscoped().First(&product, "category_id = ?", categoryID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Invalid input data", "")
		return
	}

	newDeleteStatus := !product.IsDeleted
	updateData := map[string]interface{}{
		"is_deleted": newDeleteStatus,
	}
	if newDeleteStatus {
		updateData["deleted_at"] = time.Now()
	} else {
		updateData["deleted_at"] = nil
	}

	if err := tx.Unscoped().Model(&models.ProductDetail{}).Where("category_id = ?", categoryID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product", "Failed to update product", "")
		return
	}

	if err := tx.Unscoped().Model(&models.ProductVariantDetails{}).Where("category_id = ?", categoryID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product variants", "Failed to update product variants", "")
		return
	}

	if err := tx.Commit().Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Transaction commit failed", "Transaction commit failed", "")
		return
	}

	message := "Category deleted successfully"
	if !category.IsDeleted {
		message = "Category recovered successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": message,
		"code":    http.StatusOK,
	})
}

func ShowCategoryDetails(c *gin.Context) {
	categoryID := c.Param("id")

	var category models.Categories
	if err := config.DB.First(&category, categoryID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Categoey not found", "Something Weint Wrong", "")
		return
	}

	var product []models.ProductVariantDetails
	config.DB.Preload("VariantsImages").Find(&product, "category_id = ?", categoryID)

	isOfferNotAvailable := false
	var offer models.OfferByCategory
	if err := config.DB.First(&offer, "category_id = ? AND offer_status = ?", categoryID, "Active").Error; err != nil {
		isOfferNotAvailable = true
	}

	c.HTML(http.StatusOK, "categoryDetails.html", gin.H{
		"Category":            category,
		"Products":            product,
		"IsOfferNotAvailable": isOfferNotAvailable,
		"OfferName":           offer.CategoryOfferName,
		"OfferPercentage":     offer.CategoryOfferPercentage,
		"OfferDescription":    offer.OfferDescription,
		"OfferId":             offer.ID,
		"OfferStatus":         offer.OfferStatus,
		"OfferStartDate":      offer.StartDate.Format("2006-01-02"),
		"OfferEndDate":        offer.EndDate.Format("2006-01-02"),
	})
}
