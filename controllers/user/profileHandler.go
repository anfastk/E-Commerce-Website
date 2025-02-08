package controllers

import (
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ProfileDetails(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var authDetails models.UserAuth
	if err := config.DB.Preload("UserProfile").
		Where("id = ? AND is_blocked = ?", userID, false).
		First(&authDetails).
		Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Invalid data", "")
		return
	}
	userID, err := strconv.ParseUint(userUpdate.Id, 10, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user_id", "Invalid user_id", "")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to update user", "Failed to update user", "")
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
		if err := tx.Create(&profile).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create user", "Failed to create user", "")
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update user", "Failed to update user", "")
			return
		}
	}

	tx.Commit()
	helper.RespondWithError(c, http.StatusOK, "User updated successfully", "User updated successfully", "")
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

func ManageAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var authDetails models.UserAuth
	if err := config.DB.First(&authDetails, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	var address []models.UserAddress
	if err := config.DB.Order("updated_at DESC").Find(&address, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	c.HTML(http.StatusOK, "profileManageAddress.html", gin.H{
		"User": authDetails,
		"Address":address,
	})
}

func ShowAddAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var userauth models.UserAuth
	if err := config.DB.Find(&userauth, "id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	c.HTML(http.StatusOK, "addNewAddress.html", gin.H{
		"User": userauth,
	})
}

func AddAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var addAddress struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Mobile    string `json:"phoneNumber"`
		Address   string `json:"address"`
		Landmark  string `json:"landmark"`
		Country   string `json:"country"`
		State     string `json:"state"`
		City      string `json:"city"`
		PinCode   string `json:"zipCode"`
	}
	var address models.UserAddress

	if err := c.ShouldBindJSON(&addAddress); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Invalid Data", "")
		return
	}

	address = models.UserAddress{
		FirstName: addAddress.FirstName,
		LastName:  addAddress.LastName,
		Mobile:    addAddress.Mobile,
		Address:   addAddress.Address,
		Landmark:  addAddress.Landmark,
		Country:   addAddress.Country,
		State:     addAddress.State,
		City:      addAddress.City,
		PinCode:   addAddress.PinCode,
		UserID:    userID,
	}
	if err := config.DB.Create(&address).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Address create failed", "Address create failed", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Address Added Successfully",
		"code":    http.StatusOK,
	})
}

func ShowEditAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	id := c.Param("id")
	var userauth models.UserAuth
	if err := config.DB.Find(&userauth, "id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var userAddress models.UserAddress
	if err := config.DB.First(&userAddress, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}
	c.HTML(http.StatusOK, "editAddress.html", gin.H{
		"User":    userauth,
		"Address": userAddress,
	})
}

func EditAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var UpdateAddress struct {
		Id        string `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Mobile    string `json:"phoneNumber"`
		Address   string `json:"address"`
		Landmark  string `json:"landmark"`
		Country   string `json:"country"`
		State     string `json:"state"`
		City      string `json:"city"`
		PinCode   string `json:"zipCode"`
	}

	if err := c.ShouldBindJSON(&UpdateAddress); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "invalid data", "invalid data", "")
		return
	}
	ID, err := strconv.ParseUint(UpdateAddress.Id, 10, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "invalid id", "invalid id", "")
		return
	}
	var address models.UserAddress

	if err := config.DB.Model(&address).Where("id = ? AND user_id = ?", ID, userID).Updates(map[string]interface{}{
		"first_name": UpdateAddress.FirstName,
		"last_name":  UpdateAddress.LastName,
		"mobile":     UpdateAddress.Mobile,
		"address":    UpdateAddress.Address,
		"landmark":   UpdateAddress.Landmark,
		"country":    UpdateAddress.Country,
		"state":      UpdateAddress.State,
		"city":       UpdateAddress.City,
		"pin_code":   UpdateAddress.PinCode,
	}).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update address", "Failed to update address", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Address Edited Successfully",
		"code":    http.StatusOK,
	})
}

func SetAsDefaultAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	addressID := c.Param("id")
	tx := config.DB.Begin()
	if err := tx.Model(&models.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to Update Address", "Failed to Update Address", "")
		return
	}
	if err := tx.Model(&models.UserAddress{}).Where("user_id = ? AND id = ?", userID, addressID).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed To Update Default Address", "Failed To Update Default Address", "")
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Default address updated successfully",
		"code":    http.StatusOK,
	})
}

func DeleteAddress(c *gin.Context) {
	id := c.Param("id")

	var address models.UserAddress
	if err := config.DB.First(&address, id).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&address).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete address", "Failed to delete address", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Address Deleted Successfully",
		"code":    http.StatusOK,
	})
}

func ShowChangePassword(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	var userAuth models.UserAuth
	if err := config.DB.First(&userAuth, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	c.HTML(http.StatusOK, "profileChangePassword.html", gin.H{
		"status": "OK",
		"User":   userAuth,
		"code":   http.StatusOK,
	})
}

func ChangePassword(c *gin.Context) {
	userID := c.PostForm("user_id")
	currentPassword := c.PostForm("current_password")
	password := c.PostForm("password")
	conformPassword := c.PostForm("conform_password")
	if currentPassword == "" || password == "" || conformPassword == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Invalid input data", "")
		return
	}
	var userAuth models.UserAuth
	if err := config.DB.First(&userAuth, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	if userAuth.Password != "" {
		if !CheckPasswordHash(currentPassword, userAuth.Password) {
			helper.RespondWithError(c, http.StatusBadRequest, "Enter correct current password", "Enter correct current password", "")
			return
		}
	}
	if password != conformPassword {
		helper.RespondWithError(c, http.StatusBadRequest, "Password not match", "Password not match", "")
		return
	}
	hashedPassowrd, err := HashPassword(conformPassword)
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to process password", "Failed to process password", "")
		return
	}
	userAuth = models.UserAuth{
		Password: hashedPassowrd,
	}
	if err := config.DB.Model(&userAuth).
		Where("id = ?", userID).
		Updates(userAuth).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to change password", "Failed to change password", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Password Changed successfully",
		"code":    http.StatusOK,
	})
}
