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
	var categories []models.Categories

	if err := config.DB.Order("id ASC").Find(&categories).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}
	c.HTML(http.StatusOK, "categoryManagement.html", gin.H{
		"categories": categories,
	})
}

func AddCategory(c *gin.Context) {
	c.Request.ParseForm()

	var category models.Categories
	if err := c.ShouldBind(&category); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	if category.Name == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Category name is required")
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create category")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Category ID is missing")
		return
	}

	if _, err := strconv.ParseInt(categoryID, 10, 64); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Category ID")
		return
	}

	var category models.Categories

	if err := config.DB.First(&category, categoryID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Category not found")
		return
	}

	if err := c.Bind(&category); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to bind form data")
		return
	}

	if err := config.DB.Model(&category).Where("id = ?", categoryID).Updates(category).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update category")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	tx := config.DB.Begin()

	// Find the category
	if err := tx.Unscoped().First(&category, categoryID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Category not found")
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
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update category status")
		return
	}

	var product models.ProductDetail
	if err := tx.Unscoped().First(&product, "category_id = ?", categoryID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data")
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
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product")
		return
	}

	if err := tx.Unscoped().Model(&models.ProductVariantDetails{}).Where("category_id = ?", categoryID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product variants")
		return
	}

	if err := tx.Commit().Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Transaction commit failed")
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