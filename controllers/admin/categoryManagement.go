package controllers

import (
	"net/http"
	"time"

	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
)

func ListCategory(c *gin.Context) {
	var categorys []models.Categories

	if err := config.DB.Order("id ASC").Find(&categorys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Status": "InternalServerError",
			"error":  "Failed to fetch categories",
		})
		return
	}
	c.HTML(http.StatusOK, "categoryManagement.html", gin.H{
		"categories": categorys,
	})
}

func AddCategory(c *gin.Context) {
	c.Request.ParseForm()

	var category models.Categories
	if err := c.ShouldBind(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid input data",
			"code":   400,
		})
		return
	}

	if category.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Category name is required",
			"code":   400,
		})
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to create category",
			"code":   500,
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Category ID is missing",
			"code":   400,
		})
		return
	}

	if _, err := strconv.ParseInt(categoryID, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid Category ID",
			"code":   400,
		})
		return
	}

	var category models.Categories

	if err := config.DB.First(&category, categoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Category not found",
			"code":   404,
		})
		return
	}

	if err := c.Bind(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Failed to bind form data",
			"code":   400,
		})
		return
	}

	if err := config.DB.Model(&category).Where("id = ?", categoryID).Updates(category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to update category",
			"code":   500,
		})
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
	CategoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   http.StatusBadRequest,
		})
		return
	}
	tx:=config.DB.Begin()
	// Find the category
	if err := tx.Unscoped().First(&category, CategoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Category not found",
			"code":   http.StatusNotFound,
		})
		return
	}

	category.IsDeleted = !category.IsDeleted
    if category.IsDeleted {
        category.Status = "Blocked"
    } else {
        category.Status = "Active"
    }

    if err := config.DB.Save(&category).Error; err != nil {
		tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{
            "status": "Internal Server Error",
            "error":  "Failed to update category status",
            "code":   http.StatusInternalServerError,
        })
        return
    }


	var product models.ProductDetail
	if err := tx.Unscoped().First(&product,"category_id = ?", CategoryID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid input data",
			"code":   http.StatusBadRequest,
		})
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

	if err := tx.Unscoped().Model(&models.ProductDetail{}).Where("category_id = ?", CategoryID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Error",
			"error":  "Failed to update product",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	if err := tx.Unscoped().Model(&models.ProductVariantDetails{}).Where("category_id = ?", CategoryID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Error",
			"error":  "Failed to update product variants",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Error",
			"error":  "Transaction commit failed",
			"code":   http.StatusInternalServerError,
		})
		return
	}
	message := "Category deleted successfully"
	if !category.IsDeleted {
		message = "Category recovered successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Ok",
		"message": message,
		"code":    http.StatusOK,
	})
}
