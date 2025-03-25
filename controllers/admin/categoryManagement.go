package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ListCategory(c *gin.Context) {
	type categoryResponce struct {
		ID           uint
		Name         string
		Description  string
		Status       string
		ProductCount int
	}

	logger.Log.Info("Requested TO Open Category Management Page")

	var categories []models.Categories
	if err := config.DB.Find(&categories).Error; err != nil {
		logger.Log.Error("Failed to fetch categories", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories", "Something Went Wrong", "")
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
	logger.Log.Info("Admin Open Category Management Successfully")
	c.HTML(http.StatusOK, "categoryManagement.html", gin.H{
		"categories":          responce,
		"CategoryCount":       categoryCount,
		"ActiveCategoryCount": activeCount,
		"TotalProductCount":   totalProductCount,
	})
}

func AddCategory(c *gin.Context) {
	c.Request.ParseForm()

	logger.Log.Info("Requested TO Add Category")

	var categoryInput struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&categoryInput); err != nil {
		logger.Log.Error("Invaild Data Entered", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data Entered", "Invalid data Entered", "")
		return
	}

	if categoryInput.Name == "" {
		logger.Log.Error("Category Name Is Required")
		helper.RespondWithError(c, http.StatusBadRequest, "Category name is required", "Category name is required", "")
		return
	}

	category := models.Categories{
		Name:        categoryInput.Name,
		Description: categoryInput.Description,
	}

	if err := config.DB.Create(&category).Error; err != nil {
		logger.Log.Error("Failed To Create Category", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create category", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Categoty Created Successfully")

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Category created successfully",
		"code":    200,
	})
}

func EditCategory(c *gin.Context) {
	logger.Log.Info("Requested To Edit Category")
	categoryID := c.Param("id")

	if categoryID == "" {
		logger.Log.Error("Category ID Is Missing")
		helper.RespondWithError(c, http.StatusBadRequest, "Category ID is missing", "Something Went Wrong", "")
		return
	}

	if _, err := strconv.ParseInt(categoryID, 10, 64); err != nil {
		logger.Log.Error("Invalid Category ID", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Category ID", "Invalid Category ID", "")
		return
	}

	var category models.Categories

	logger.Log.Info("Category requested", zap.String("CategoryID", categoryID))

	if err := config.DB.First(&category, categoryID).Error; err != nil {
		logger.Log.Error("Failed To Fetch Category,Category not found", zap.String("CategoryID", categoryID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Category not found", "Category not found", "")
		return
	}

	logger.Log.Info("Category viewed", zap.String("CategoryID", categoryID))

	if err := c.Bind(&category); err != nil {
		logger.Log.Error("Failed to bind data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to bind form data", "Invalid input data", "")
		return
	}

	if err := config.DB.Model(&category).Where("id = ?", categoryID).Updates(category).Error; err != nil {
		logger.Log.Error("Failed to update category", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update category", "Something Went Wrong", "")
		return
	}
	logger.Log.Info("Categoty Updated Successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Category updated successfully",
		"code":    200,
	})
}

func DeleteCategory(c *gin.Context) {
	logger.Log.Info("Requested To Change Category Status")
	var category models.Categories
	categoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		logger.Log.Error("Invalid Category ID", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid category ID", "Soemthing Wewnt Wrong", "")
		return
	}

	tx := config.DB.Begin()

	logger.Log.Info("Category Requested", zap.Uint64("CategoryID", categoryID))

	if err := tx.Unscoped().First(&category, categoryID).Error; err != nil {
		tx.Rollback()
		logger.Log.Error("Failed To Fetch Category", zap.Uint64("CategoryID", categoryID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Category not found", "Category not found", "")
		return
	}

	logger.Log.Info("Category Viewed", zap.Uint64("CategoryID", categoryID))

	category.IsDeleted = !category.IsDeleted
	if category.IsDeleted {
		logger.Log.Info("Blocked Successfully")
		category.Status = "Blocked"
	} else {
		logger.Log.Info("Active Successfully")
		category.Status = "Active"
	}

	if err := tx.Save(&category).Error; err != nil {
		logger.Log.Error("Failed To Update Category Status", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update category status", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Requested To Change Product Status", zap.Uint64("CategoryID", categoryID))

	var product models.ProductDetail
	if err := tx.Unscoped().First(&product, "category_id = ?", categoryID).Error; err != nil {
		logger.Log.Error("Failed To Fetch Category", zap.Uint64("CategoryID", categoryID), zap.Error(err))
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
		logger.Log.Error("Failed To Update Product Status", zap.Uint64("CategoryID", categoryID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product", "Failed to update product", "")
		return
	}

	if err := tx.Unscoped().Model(&models.ProductVariantDetails{}).Where("category_id = ?", categoryID).Updates(updateData).Error; err != nil {
		logger.Log.Error("Failed To Update Product Variant Status", zap.Uint64("CategoryID", categoryID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product variants", "Failed to update product variants", "")
		return
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed To Commit Transaction", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Transaction commit failed", "Transaction commit failed", "")
		return
	}

	message := "Category deleted successfully"
	if !category.IsDeleted {
		message = "Category recovered successfully"
	}

	logger.Log.Info(message)

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": message,
		"code":    http.StatusOK,
	})
}

func ShowCategoryDetails(c *gin.Context) {
	categoryID := c.Param("id")

	logger.Log.Info("Category Requested", zap.String("CategoryID", categoryID))

	var category models.Categories
	if err := config.DB.First(&category, categoryID).Error; err != nil {
		logger.Log.Error("Failed To Fetch Category", zap.String("CategoryID", categoryID), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Categoey not found", "Something Weint Wrong", "")
		return
	}

	logger.Log.Info("Category Viewed", zap.String("CategoryID", categoryID))

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
