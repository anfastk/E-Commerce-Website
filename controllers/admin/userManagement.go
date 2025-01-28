package controllers

import (
	"fmt"
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
			"status": "InternalServerError",
			"error":  "Could not fetch users",
			"code":   500,
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
			"status":  "Not Found",
			"message": "User not found",
			"code":    500,
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
			"status":   "InternalServerError",
			"messeage": "Failed to block user/unblock user",
			"code":     500,
		})
	}

}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.UserAuth

	if err := config.DB.Unscoped().First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "User not found",
			"code":   http.StatusNotFound,
		})
		return
	}

	user.IsDeleted = !user.IsDeleted
	if user.IsDeleted {
		user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	} else {
		user.DeletedAt = gorm.DeletedAt{}
	}

	if err := config.DB.Save(&user).Error; err != nil {
		action := "delete"
		if !user.IsDeleted {
			action = "restore"
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  fmt.Sprintf("Failed to %s user", action),
			"code":   http.StatusInternalServerError,
		})
		return
	}

	message := "User deleted successfully"
	if !user.IsDeleted {
		message = "User restored successfully"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": message,
		"code":    http.StatusOK,
	})
}
