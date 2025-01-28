package controllers

import (
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListUsers(c *gin.Context) {
	var users []models.UserAuth
	if err := config.DB.Unscoped().Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":"InternalServerError",
			"error": "Could not fetch users",
			"code":500,
		})
		return
	}
	c.HTML(http.StatusOK, "user_management.html", gin.H{
		"users": users,
	})
}

func BlockUser(c *gin.Context) {
	id := c.Param("id")
	var user models.UserAuth
	if err := config.DB.Unscoped().First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"message":  "User not found",
			"code":   500,
		})
		return
	}

	if user.IsBlocked {
		user.IsBlocked = false
		user.Status = "Active"
		c.JSON(http.StatusOK, gin.H{
			"status":  "Success",
			"message": "User's account unblocked",
			"code":    200,
		})
	} else {
		user.IsBlocked = true
		user.Status = "Blocked"
		c.JSON(http.StatusOK, gin.H{
			"status":  "Success",
			"message": "User's account blocked",
			"code":    200,
		})
	}
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "InternalServerError",
			"messeage":  "Failed to block user/unblock user",
			"code": 500,
		})
	}
	
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id") 

	var user models.UserAuth
	
	if err := config.DB.Unscoped().First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":"Not Found",
			"error": "User not found",
			"code":500,
		})
		return
	}

	user.IsDeleted = true
	user.DeletedAt = gorm.DeletedAt{Time: time.Now(),Valid: true}
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status" : "InternalServerError",
			"error": "Failed to delete user",
			"code":500,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":"Ok",
		"message": "User deleted successfully",
		"code":200,
	})
}
