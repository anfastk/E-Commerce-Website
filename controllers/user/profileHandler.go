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
		helper.RespondWithError(c, http.StatusBadRequest, "User not found")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data")
		return
	}
	userID, err := strconv.ParseUint(userUpdate.Id, 10, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user_id")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to update user")
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create user")
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update user")
			return
		}
	}

	tx.Commit()
	helper.RespondWithError(c, http.StatusOK, "User updated successfully")
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
	if err := config.DB.Preload("UserAddress", "user_id = ?", userID).
		Where("id = ? AND is_blocked = ?", userID, false).
		First(&authDetails).
		Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "User not found")
		return
	}

	c.HTML(http.StatusOK, "profileManageAddress.html", gin.H{
		"User": authDetails,
	})
}

func ShowAddAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var userauth models.UserAuth
	if err := config.DB.Find(&userauth, "id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "User not found")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data")
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
		helper.RespondWithError(c, http.StatusInternalServerError, "User create failed")
		return
	}
	helper.RespondWithError(c, http.StatusOK, "User create successfully")
}

func ShowEditAddress(c *gin.Context) {
	userID := c.MustGet("userid").(uint)
	id := c.Param("id")
	var userauth models.UserAuth
	if err := config.DB.Find(&userauth, "id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "User not found")
		return
	}

	var userAddress models.UserAddress
	if err := config.DB.First(&userAddress, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Address not found")
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
		helper.RespondWithError(c, http.StatusBadRequest, "invalid data")
		return
	}
	ID, err := strconv.ParseUint(UpdateAddress.Id, 10, 64)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "invalid id")
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
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update address")
		return
	}
	helper.RespondWithError(c, http.StatusOK, "User updated successfully")
}

func DeleteAddress(c *gin.Context) {
	id := c.Param("id")

	var address models.UserAddress
	if err := config.DB.First(&address, id).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Address not found")
		return
	}

	if err := config.DB.Unscoped().Delete(&address).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete address")
		return
	}
	helper.RespondWithError(c, http.StatusOK, "Address deleted successfully")
}
