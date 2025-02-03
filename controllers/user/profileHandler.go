package controllers

import (
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
)

func ProfileDetails(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var authDetails models.UserAuth
	if err := config.DB.Preload("UserProfile").
		Where("id = ? AND is_blocked = ?", userID, false).
		First(&authDetails).
		Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "Bad Request",
			"message": "User not found",
			"code":    http.StatusBadRequest,
		})
		return
	}
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"User": authDetails,
	})
}

func ProfileUpdate(c *gin.Context) {
	var userUpdate struct {
		Id       string `json:"userid"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
		Mobile   string `json:"phone"`
		Country  string `json:"country"`
		State    string `json:"state"`
		Pincode  string `json:"zipcode"`
	}
	if err := c.ShouldBindJSON(&userUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Data",
		})
		return
	}
	userID, err := strconv.ParseUint(userUpdate.Id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
		return
	}

	var authDetails models.UserAuth
	var profile models.UserProfile

	tx := config.DB.Begin()

	if err := tx.Model(&authDetails).Where("id = ? ", userID).Updates(map[string]interface{}{
		"full_name": userUpdate.FullName,
		"email":     userUpdate.Email,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"statue":  "Internal Server Error",
			"message": "Failed to update user",
			"code":    http.StatusInternalServerError,
		})
		return
	}

	userprofile := tx.First(&profile, "user_id = ?", userID)

	if userprofile.Error != nil {
		profile = models.UserProfile{
			UserID:  uint(userID),
			Mobile:  userUpdate.Mobile,
			Country: userUpdate.Country,
			State:   userUpdate.State,
			Pincode: userUpdate.Pincode,
		}
		if err := config.DB.Create(&profile).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, gin.H{
				"status":  "Status OK",
				"message": "User updated successfully",
				"code":    http.StatusOK,
			})
			return
		}
	} else {
		if err := tx.Model(&profile).Where("user_id = ? ", userID).Updates(map[string]interface{}{
			"mobile":  userUpdate.Mobile,
			"country": userUpdate.Country,
			"state":   userUpdate.State,
			"pincode": userUpdate.Pincode,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"statue":  "Internal Server Error",
				"message": "Failed to update user",
				"code":    http.StatusInternalServerError,
			})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "User updated successfully",
		"code":    http.StatusOK,
	})
}

func Settings(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var userDetails models.UserAuth
	if err := config.DB.First(&userDetails, "id=? AND is_blocked = ?", userID, false).Error; err != nil {
		c.Redirect(http.StatusSeeOther, "/user/login")
		return
	}

	c.HTML(http.StatusOK, "profileSettings.html", gin.H{
		"User": userDetails,
	})
}
