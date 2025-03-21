package controllers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ProfileDetails(c *gin.Context) {
	logger.Log.Info("Fetching profile details")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var authDetails models.UserAuth
	if err := config.DB.Preload("UserProfile").
		Where("id = ? AND is_blocked = ?", userID, false).
		First(&authDetails).
		Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	logger.Log.Info("Profile details loaded", zap.Uint("userID", userID))
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"User": authDetails,
	})
}

func ProfileUpdate(c *gin.Context) {
	logger.Log.Info("Updating profile")

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
		logger.Log.Error("Failed to bind profile update data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Invalid data", "")
		return
	}

	userID, err := strconv.ParseUint(userUpdate.Id, 10, 64)
	if err != nil {
		logger.Log.Error("Invalid user ID",
			zap.String("userID", userUpdate.Id),
			zap.Error(err))
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
		logger.Log.Error("Failed to update user auth details",
			zap.Uint("userID", uint(userID)),
			zap.Error(err))
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
			logger.Log.Error("Failed to create user profile",
				zap.Uint("userID", uint(userID)),
				zap.Error(err))
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
			logger.Log.Error("Failed to update user profile",
				zap.Uint("userID", uint(userID)),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update user", "Failed to update user", "")
			return
		}
	}

	tx.Commit()
	logger.Log.Info("Profile updated successfully", zap.Uint("userID", uint(userID)))
	helper.RespondWithError(c, http.StatusOK, "User updated successfully", "User updated successfully", "")
}

func ProfileImageUpdate(c *gin.Context) {
	logger.Log.Info("Updating profile image")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	tx := config.DB.Begin()
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		logger.Log.Error("Invalid file upload", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid file", "Invalid file choosen", "")
		return
	}

	cld := config.InitCloudinary()
	cloudinaryURL, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "ProfilePicture", "")
	if err != nil {
		logger.Log.Error("Failed to upload image to Cloudinary",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	var userDetails models.UserAuth
	if userErr := tx.First(&userDetails, "id = ?", userID).Error; userErr != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(userErr))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	currentAvathar := userDetails.ProfilePic
	userDetails.ProfilePic = cloudinaryURL
	if err := config.DB.Save(&userDetails).Error; err != nil {
		logger.Log.Error("Failed to save profile picture update",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile picture", "Something Went Wrong", "")
		return
	}

	defaultAvathar := os.Getenv("DEFAULT_PROFILE_PIC")
	if currentAvathar != defaultAvathar {
		publicID, err := helper.ExtractCloudinaryPublicID(currentAvathar)
		if err != nil {
			logger.Log.Error("Failed to extract Cloudinary public ID",
				zap.String("currentAvatar", currentAvathar),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Failed to extract Cloudinary public ID", "")
			return
		}
		if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
			logger.Log.Error("Failed to delete old image from Cloudinary",
				zap.String("publicID", publicID),
				zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Failed to delete image from Cloudinary", "")
			return
		}
	}

	tx.Commit()
	logger.Log.Info("Profile image updated successfully",
		zap.Uint("userID", userID),
		zap.String("newImageURL", cloudinaryURL))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile image updated",
		"code":    http.StatusOK,
	})
}

func Settings(c *gin.Context) {
	logger.Log.Info("Loading settings page")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var userDetails models.UserAuth
	if err := config.DB.First(&userDetails, "id=? AND is_blocked = ?", userID, false).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}

	logger.Log.Info("Settings page loaded", zap.Uint("userID", userID))
	c.HTML(http.StatusOK, "profileSettings.html", gin.H{
		"User": userDetails,
	})
}

func ManageAddress(c *gin.Context) {
	logger.Log.Info("Managing addresses")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var authDetails models.UserAuth
	if err := config.DB.First(&authDetails, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var address []models.UserAddress
	if err := config.DB.Order("updated_at DESC").Find(&address, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Addresses not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	logger.Log.Info("Address management page loaded",
		zap.Uint("userID", userID),
		zap.Int("addressCount", len(address)))
	c.HTML(http.StatusOK, "profileManageAddress.html", gin.H{
		"User":    authDetails,
		"Address": address,
	})
}

func ShowAddAddress(c *gin.Context) {
	logger.Log.Info("Showing add address page")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var userauth models.UserAuth
	if err := config.DB.Find(&userauth, "id = ?", userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	logger.Log.Info("Add address page loaded", zap.Uint("userID", userID))
	c.HTML(http.StatusOK, "addNewAddress.html", gin.H{
		"User": userauth,
	})
}

func AddAddress(c *gin.Context) {
	logger.Log.Info("Adding new address")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

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

	if err := c.ShouldBindJSON(&addAddress); err != nil {
		logger.Log.Error("Failed to bind address data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Data", "Invalid Data", "")
		return
	}

	address := models.UserAddress{
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
		logger.Log.Error("Failed to create address",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Address create failed", "Address create failed", "")
		return
	}

	logger.Log.Info("Address added successfully",
		zap.Uint("userID", userID),
		zap.Uint("addressID", address.ID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Address Added Successfully",
		"code":    http.StatusOK,
	})
}

func ShowEditAddress(c *gin.Context) {
	logger.Log.Info("Showing edit address page")

	userID := helper.FetchUserID(c)
	id := c.Param("id")
	logger.Log.Debug("Fetched user ID and address ID",
		zap.Uint("userID", userID),
		zap.String("addressID", id))

	var userauth models.UserAuth
	if err := config.DB.Find(&userauth, "id = ?", userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var userAddress models.UserAddress
	if err := config.DB.First(&userAddress, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		logger.Log.Error("Address not found",
			zap.String("addressID", id),
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	logger.Log.Info("Edit address page loaded",
		zap.Uint("userID", userID),
		zap.Uint("addressID", userAddress.ID))
	c.HTML(http.StatusOK, "editAddress.html", gin.H{
		"User":    userauth,
		"Address": userAddress,
	})
}

func EditAddress(c *gin.Context) {
	logger.Log.Info("Editing address")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

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
		logger.Log.Error("Failed to bind address update data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "invalid data", "invalid data", "")
		return
	}

	ID, err := strconv.ParseUint(UpdateAddress.Id, 10, 64)
	if err != nil {
		logger.Log.Error("Invalid address ID",
			zap.String("addressID", UpdateAddress.Id),
			zap.Error(err))
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
		logger.Log.Error("Failed to update address",
			zap.Uint("addressID", uint(ID)),
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update address", "Failed to update address", "")
		return
	}

	logger.Log.Info("Address edited successfully",
		zap.Uint("userID", userID),
		zap.Uint("addressID", uint(ID)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Address Edited Successfully",
		"code":    http.StatusOK,
	})
}

func SetAsDefaultAddress(c *gin.Context) {
	logger.Log.Info("Setting default address")

	userID := helper.FetchUserID(c)
	addressID := c.Param("id")
	logger.Log.Debug("Fetched user ID and address ID",
		zap.Uint("userID", userID),
		zap.String("addressID", addressID))

	tx := config.DB.Begin()
	if err := tx.Model(&models.UserAddress{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
		logger.Log.Error("Failed to reset default addresses",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to Update Address", "Failed to Update Address", "")
		return
	}

	if err := tx.Model(&models.UserAddress{}).Where("user_id = ? AND id = ?", userID, addressID).Update("is_default", true).Error; err != nil {
		logger.Log.Error("Failed to set default address",
			zap.String("addressID", addressID),
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed To Update Default Address", "Failed To Update Default Address", "")
		return
	}

	tx.Commit()
	logger.Log.Info("Default address set successfully",
		zap.Uint("userID", userID),
		zap.String("addressID", addressID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Default address updated successfully",
		"code":    http.StatusOK,
	})
}

func DeleteAddress(c *gin.Context) {
	logger.Log.Info("Deleting address")

	id := c.Param("id")
	logger.Log.Debug("Fetched address ID", zap.String("addressID", id))

	var address models.UserAddress
	if err := config.DB.First(&address, id).Error; err != nil {
		logger.Log.Error("Address not found",
			zap.String("addressID", id),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Address not found", "Address not found", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&address).Error; err != nil {
		logger.Log.Error("Failed to delete address",
			zap.Uint("addressID", address.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete address", "Failed to delete address", "")
		return
	}

	logger.Log.Info("Address deleted successfully",
		zap.Uint("addressID", address.ID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Address Deleted Successfully",
		"code":    http.StatusOK,
	})
}

func ShowChangePassword(c *gin.Context) {
	logger.Log.Info("Showing change password page")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var userAuth models.UserAuth
	if err := config.DB.First(&userAuth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	logger.Log.Info("Change password page loaded", zap.Uint("userID", userID))
	c.HTML(http.StatusOK, "profileChangePassword.html", gin.H{
		"status": "OK",
		"User":   userAuth,
		"code":   http.StatusOK,
	})
}

func ChangePassword(c *gin.Context) {
	logger.Log.Info("Changing password")

	userID := c.PostForm("user_id")
	currentPassword := c.PostForm("current_password")
	password := c.PostForm("password")
	conformPassword := c.PostForm("conform_password")

	if currentPassword == "" || password == "" || conformPassword == "" {
		logger.Log.Error("Invalid password input data")
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Invalid input data", "")
		return
	}

	var userAuth models.UserAuth
	if err := config.DB.First(&userAuth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.String("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	if userAuth.Password != "" {
		if !CheckPasswordHash(currentPassword, userAuth.Password) {
			logger.Log.Warn("Incorrect current password",
				zap.Uint("userID", userAuth.ID))
			helper.RespondWithError(c, http.StatusBadRequest, "Enter correct current password", "Enter correct current password", "")
			return
		}
	}

	if password != conformPassword {
		logger.Log.Warn("Password mismatch",
			zap.Uint("userID", userAuth.ID))
		helper.RespondWithError(c, http.StatusBadRequest, "Password not match", "Password not match", "")
		return
	}

	hashedPassword, err := HashPassword(conformPassword)
	if err != nil {
		logger.Log.Error("Failed to hash password",
			zap.Uint("userID", userAuth.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to process password", "Failed to process password", "")
		return
	}

	userAuth.Password = hashedPassword
	if err := config.DB.Model(&userAuth).
		Where("id = ?", userID).
		Updates(userAuth).Error; err != nil {
		logger.Log.Error("Failed to update password",
			zap.Uint("userID", userAuth.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to change password", "Failed to change password", "")
		return
	}

	logger.Log.Info("Password changed successfully",
		zap.Uint("userID", userAuth.ID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Password Changed successfully",
		"code":    http.StatusOK,
	})
}

func OrderDetails(c *gin.Context) {
	logger.Log.Info("Fetching order details")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	type OrderResponse struct {
		Slno                 int     `json:"slno"`
		ID                   uint    `json:"id"`
		OrderID              uint    `json:"order_id"`
		ProductId            uint    `json:"productid"`
		FirstName            string  `json:"firstname"`
		LastName             string  `json:"lastname"`
		OrderUID             string  `json:"orderid"`
		Image                string  `json:"image"`
		Quantity             uint    `json:"quantity"`
		ProductSummary       string  `json:"productsummary"`
		CategoryName         string  `json:"categoryname"`
		OrderStatus          string  `json:"orderststus"`
		PaymentStatus        string  `json:"paymentststus"`
		SubTotal             float64 `json:"subtotal"`
		OrderDate            string  `json:"orderdate"`
		DeliveryDate         string  `json:"deliverydate"`
		ReturnDate           string  `json:"returndate"`
		ReturnableDate       string  `json:"returnabledate"`
		CancelDate           string  `json:"canceldate"`
		IsReturnable         bool    `json:"is_returnable"`
		ExpectedDeliveryDate string  `json:"expecteddeliverydate"`
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, "id = ?", userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var order []models.Order
	if err := config.DB.Order("created_at DESC").Find(&order, "user_id = ?", userID).Error; err != nil {
		logger.Log.Warn("No orders found",
			zap.Uint("userID", userID),
			zap.Error(err))
		c.HTML(http.StatusOK, "profileOrderDetails.html", gin.H{
			"status":  "success",
			"message": "No orders found",
			"User":    userauth,
			"data":    []OrderResponse{},
		})
		return
	}

	var orderResponses []OrderResponse
	for _, order := range order {
		var orderItems []models.OrderItem
		if ordErr := config.DB.Order("created_at DESC").
			Find(&orderItems, "user_id = ? AND order_id = ?", userID, order.ID).Error; ordErr != nil {
			logger.Log.Error("Order items not found",
				zap.Uint("orderID", order.ID),
				zap.Error(ordErr))
			helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
			return
		}

		var shippingAddress models.ShippingAddress
		if addErr := config.DB.First(&shippingAddress, "user_id = ? AND order_id = ?", userauth.ID, order.ID).Error; addErr != nil {
			logger.Log.Error("Shipping address not found",
				zap.Uint("orderID", order.ID),
				zap.Error(addErr))
			helper.RespondWithError(c, http.StatusInternalServerError, "Address Not found", "Something Went Wrong", "")
			return
		}

		for i, item := range orderItems {
			currentTime := time.Now()
			daysSinceDelivery := currentTime.Sub(item.DeliveryDate).Hours() / 24
			if daysSinceDelivery > 7 {
				item.ReturnableStatus = false
				if err := config.DB.Save(&item).Error; err != nil {
					logger.Log.Error("Failed to update returnable status",
						zap.Uint("orderItemID", item.ID),
						zap.Error(err))
					helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product", "Something Went Wrong", "")
					return
				}
			}

			var payment models.PaymentDetail
			if payErr := config.DB.First(&payment, "user_id = ? AND order_item_id = ?", userID, item.ID).Error; payErr != nil {
				logger.Log.Error("Payment details not found",
					zap.Uint("orderItemID", item.ID),
					zap.Error(payErr))
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment details Not found", "Something Went Wrong", "")
				return
			}

			orderResponse := OrderResponse{
				Slno:                 i + 1,
				ID:                   item.ID,
				OrderID:              item.OrderID,
				ProductId:            item.ProductVariantID,
				FirstName:            shippingAddress.FirstName,
				LastName:             shippingAddress.LastName,
				OrderUID:             item.OrderUID,
				Quantity:             uint(item.Quantity),
				Image:                item.ProductImage,
				ProductSummary:       item.ProductSummary,
				CategoryName:         item.ProductCategory,
				OrderStatus:          item.OrderStatus,
				PaymentStatus:        strings.ToUpper(payment.PaymentStatus),
				SubTotal:             item.Total,
				OrderDate:            item.CreatedAt.Format("2 January 2006"),
				DeliveryDate:         item.DeliveryDate.Format("2 January 2006"),
				ReturnDate:           item.ReturnDate.Format("2 January 2006"),
				CancelDate:           item.CancelDate.Format("2 January 2006"),
				IsReturnable:         item.ReturnableStatus,
				ExpectedDeliveryDate: item.ExpectedDeliveryDate.Format("2 January 2006"),
			}
			orderResponses = append(orderResponses, orderResponse)
		}
	}

	logger.Log.Info("Order details loaded",
		zap.Uint("userID", userID),
		zap.Int("orderCount", len(orderResponses)))
	c.HTML(http.StatusOK, "profileOrderDetails.html", gin.H{
		"status":  "success",
		"message": "Order details fetched successfully",
		"User":    userauth,
		"data":    orderResponses,
	})
}

func OrderHistory(c *gin.Context) {
	logger.Log.Info("Fetching order history")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	type OrderHistoryResponse struct {
		ID          uint    `json:"id"`
		OrderUID    string  `json:"orderid"`
		ProductName string  `json:"productname"`
		Image       string  `json:"image"`
		Quantity    uint    `json:"quantity"`
		OrderStatus string  `json:"orderststus"`
		Total       float64 `json:"total"`
		OrderDate   string  `json:"orderdate"`
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var orderItems []models.OrderItem
	if err := config.DB.Order("created_at DESC").Find(&orderItems, "user_id = ?", userID).Error; err != nil {
		logger.Log.Warn("Order items not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Order Not Found", "Something Went Wrong", "")
		return
	}

	var orderHistoryResponse []OrderHistoryResponse
	for _, item := range orderItems {
		historyResponce := OrderHistoryResponse{
			ID:          item.ID,
			OrderUID:    item.OrderUID,
			ProductName: item.ProductName,
			Image:       item.ProductImage,
			Quantity:    uint(item.Quantity),
			OrderStatus: item.OrderStatus,
			Total:       item.Total,
			OrderDate:   item.CreatedAt.Format("2006-01-02T15:04:05.000-07:00"),
		}
		orderHistoryResponse = append(orderHistoryResponse, historyResponce)
	}

	logger.Log.Info("Order history loaded",
		zap.Uint("userID", userID),
		zap.Int("orderCount", len(orderHistoryResponse)))
	c.HTML(http.StatusOK, "profileOrderHistory.html", gin.H{
		"status":  "success",
		"message": "Order details fetched successfully",
		"User":    userauth,
		"Order":   orderHistoryResponse,
	})
}
