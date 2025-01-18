package controllers

import (
	"fmt"
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"strconv"
)

func ListCategory(c *gin.Context) {
	var categorys []models.Categories

	if err := config.DB.Where("is_deleted = ? AND status = ?", false,"Active").Find(&categorys).Error; err != nil {
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
	fmt.Println("Form Data:", c.Request.Form)

	var category models.Categories
	if err := c.ShouldBind(&category); err != nil {
		fmt.Println("Error in Binding:", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid input data",
			"code":   400,
		})
		return
	}

	fmt.Println("Parsed Category:", category)

	if category.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Category name is required",
			"code":   400,
		})
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		fmt.Println("Error in DB Operation:", err)
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

	fmt.Println(categoryID)

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
	var blockCategory models.Categories
	id := c.Param("id")
	fmt.Println(id)
	if err := config.DB.First(&blockCategory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not found",
			"error":  "Category not found",
			"code":   404,
		})
		return
	}
	if blockCategory.IsDeleted {
		blockCategory.IsDeleted = false
		blockCategory.Status = "Active"
		c.JSON(http.StatusOK, gin.H{
			"status":  "Ok",
			"message": "Category deleted",
			"code":    200,
		})
	} else {
		blockCategory.IsDeleted = true
		blockCategory.Status = "Blocked"
		c.JSON(http.StatusOK, gin.H{
			"status":  "Ok",
			"message": "Category deleted",
			"code":    200,
		})
	}
	if err := config.DB.Save(&blockCategory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Delete or recover failed",
			"code":   500,
		})
	}
}
