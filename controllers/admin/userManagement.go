package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserAuth struct {
	ID        string `gorm:"primaryKey"`
	FullName  string
	Email     string
	Status    string
	IsBlocked bool
	IsDeleted bool
}

func ListUsers(c *gin.Context) {
	logger.Log.Info("Requested to list users")

	var users []UserAuth
	if err := config.DB.Unscoped().Order("id ASC").Find(&users).Error; err != nil {
		logger.Log.Error("Failed to fetch users", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Could not fetch users", "Could not fetch users", "")
		return
	}

	logger.Log.Info("Users fetched successfully", zap.Int("userCount", len(users)))
	c.HTML(http.StatusOK, "user_management.html", gin.H{
		"users": users,
	})
}

func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	logger.Log.Info("Search users requested", zap.String("query", query))

	var users []UserAuth
	queryDB := config.DB.Unscoped().Order("id ASC")

	if query != "" {
		searchTerm := "%" + query + "%"
		queryDB = queryDB.Where("full_name LIKE ? OR email LIKE ?", searchTerm, searchTerm)
	}

	if err := queryDB.Find(&users).Error; err != nil {
		logger.Log.Error("Failed to search users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not search users",
		})
		return
	}

	logger.Log.Info("Users search completed", zap.Int("userCount", len(users)))
	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func BlockUser(c *gin.Context) {
	logger.Log.Info("Requested to block/unblock user")

	id := c.Param("id")
	var user models.UserAuth
	if err := config.DB.Unscoped().First(&user, "id = ?", id).Error; err != nil {
		logger.Log.Error("User not found", zap.String("userID", id), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	if user.IsDeleted {
		logger.Log.Warn("Attempted to unblock deleted user", zap.String("userID", id))
		helper.RespondWithError(c, http.StatusBadRequest, "Can't unblock the user. The user's account is deleted.", "Can't unblock the user. The user's account is deleted.", "")
		return
	}

	var message string
	if user.IsBlocked {
		user.IsBlocked = false
		user.Status = "Active"
		message = "User's account unblocked"
	} else {
		user.IsBlocked = true
		user.Status = "Blocked"
		message = "User's account blocked"
	}

	if err := config.DB.Save(&user).Error; err != nil {
		logger.Log.Error("Failed to update user block status",
			zap.String("userID", id),
			zap.Bool("isBlocked", user.IsBlocked),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to block/unblock user", "Failed to block/unblock user", "")
		return
	}

	logger.Log.Info("User block status updated successfully",
		zap.String("userID", id),
		zap.Bool("isBlocked", user.IsBlocked))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": message,
		"code":    http.StatusOK,
	})
}

func DeleteUser(c *gin.Context) {
	logger.Log.Info("Requested to delete/restore user")

	id := c.Param("id")
	var user models.UserAuth

	if err := config.DB.Unscoped().First(&user, "id = ?", id).Error; err != nil {
		logger.Log.Error("User not found", zap.String("userID", id), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	user.IsDeleted = !user.IsDeleted
	var message string
	if user.IsDeleted {
		user.Status = "Deleted"
		user.IsBlocked = true
		user.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
		message = "User deleted successfully"
	} else {
		user.Status = "Active"
		user.IsBlocked = false
		user.DeletedAt = gorm.DeletedAt{}
		message = "User restored successfully"
	}

	if err := config.DB.Save(&user).Error; err != nil {
		action := "delete"
		if !user.IsDeleted {
			action = "restore"
		}
		errMsg := fmt.Sprintf("Failed to %s user", action)
		logger.Log.Error("Failed to update user delete status",
			zap.String("userID", id),
			zap.String("action", action),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, errMsg, errMsg, "")
		return
	}

	logger.Log.Info("User delete status updated successfully",
		zap.String("userID", id),
		zap.Bool("isDeleted", user.IsDeleted))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": message,
		"code":    http.StatusOK,
	})
}
