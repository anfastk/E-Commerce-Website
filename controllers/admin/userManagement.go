package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListUsers(c *gin.Context) {
	var users []models.UserAuth
	if err := config.DB.Unscoped().Order("id ASC").Find(&users).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Could not fetch users")
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
		helper.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	if user.IsDeleted {
		helper.RespondWithError(c, http.StatusBadRequest, "Can't unblock the user. The user's account is deleted.")
		return
	}

	if user.IsBlocked {
		user.IsBlocked = false
		user.Status = "Active"
		c.JSON(http.StatusOK, gin.H{
			"status":  "Success",
			"message": "User's account unblocked",
			"code":    http.StatusOK,
		})
	} else {
		user.IsBlocked = true
		user.Status = "Blocked"
		c.JSON(http.StatusOK, gin.H{
			"status":  "Success",
			"message": "User's account blocked",
			"code":    http.StatusOK,
		})
	}

	if err := config.DB.Save(&user).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to block/unblock user")
	}
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.UserAuth

	if err := config.DB.Unscoped().First(&user, "id = ?", id).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	user.IsDeleted = !user.IsDeleted
	if user.IsDeleted {
		user.Status = "Deleted"
		user.IsBlocked = true
		user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	} else {
		user.Status = "Active"
		user.IsBlocked = false
		user.DeletedAt = gorm.DeletedAt{}
	}

	if err := config.DB.Save(&user).Error; err != nil {
		action := "delete"
		if !user.IsDeleted {
			action = "restore"
		}
		helper.RespondWithError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to %s user", action))
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
