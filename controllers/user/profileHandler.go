package controllers

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func ProfileDetails(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

	var authDetails models.UserAuth
	if err := config.DB.Preload("UserProfile").
		Where("id = ? AND is_blocked = ?", userID, false).
		First(&authDetails).
		Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	CreateWallet(c, userID)
	CheckForReferrer(c)
	CheckForJoinee(c)
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

func ProfileImageUpdate(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

	tx := config.DB.Begin()
	file, fileHeader, err := c.Request.FormFile("image")
	if err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid file", "Invalid file choosen", "")
		return
	}
	cld := config.InitCloudinary()

	cloudinaryURL, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "ProfilePicture", "")
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	var userDetails models.UserAuth
	if userErr := tx.First(&userDetails, "id = ?", userID).Error; userErr != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	currentAvathar := userDetails.ProfilePic
	userDetails.ProfilePic = cloudinaryURL
	config.DB.Save(&userDetails)

	defaultAvathar := os.Getenv("DEFAULT_PROFILE_PIC")
	if currentAvathar != defaultAvathar {
		publicID, err := helper.ExtractCloudinaryPublicID(currentAvathar)
		if err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Failed to extract Cloudinary public ID", "")
			return
		}
		if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Failed to delete image from Cloudinary", "")
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Profile image updated",
		"code":    http.StatusOK,
	})
}

func Settings(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
		"User":    authDetails,
		"Address": address,
	})
}

func ShowAddAddress(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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

func OrderDetails(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

	type OrderResponse struct {
		Slno                 int     `json:"slno"`
		ID                   uint    `json:"id"`
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
		ExpectedDeliveryDate string  `json:"expecteddeliverydate"`
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, "id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	var order []models.Order
	if err := config.DB.Order("created_at DESC").Find(&order, "user_id = ?", userID).Error; err != nil {
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
			helper.RespondWithError(c, http.StatusInternalServerError, "Order items Not found", "Something Went Wrong", "")
			return
		}
		var shippingAddress models.ShippingAddress
		if addErr := config.DB.First(&shippingAddress, "user_id = ? AND order_id = ?", userauth.ID, order.ID).Error; addErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Address Not found", "Something Went Wrong", "")
			return
		}
		for i, item := range orderItems {
			var payment models.PaymentDetail
			if payErr := config.DB.First(&payment, "user_id = ? AND order_item_id = ?", userID, item.ID).Error; payErr != nil {
				helper.RespondWithError(c, http.StatusInternalServerError, "Payment details Not found", "Something Went Wrong", "")
				return
			}
			formattedOrderDate := item.CreatedAt.Format("2 January 2006")
			formattedReturnDate := item.ReturnDate.Format("2 January 2006")
			formattedDeliveryDate := item.DeliveryDate.Format("2 January 2006")
			formattedCancelDate := item.CancelDate.Format("2 January 2006")
			formattedExpectedDeliveryDate := item.ExpectedDeliveryDate.Format("2 January 2006")
			orderResponse := OrderResponse{
				Slno:                 i + 1,
				ID:                   item.ID,
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
				OrderDate:            formattedOrderDate,
				DeliveryDate:         formattedDeliveryDate,
				ReturnDate:           formattedReturnDate,
				CancelDate:           formattedCancelDate,
				ExpectedDeliveryDate: formattedExpectedDeliveryDate,
			}
			orderResponses = append(orderResponses, orderResponse)
		}
	}

	c.HTML(http.StatusOK, "profileOrderDetails.html", gin.H{
		"status":  "success",
		"message": "Order details fetched successfully",
		"User":    userauth,
		"data":    orderResponses,
	})
}

func OrderHistory(c *gin.Context) {
	userIDInterface, exists := c.Get("userid")
	if !exists {
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}

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
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var orderItems []models.OrderItem
	if err := config.DB.Order("created_at DESC").Find(&orderItems, "user_id = ?", userID).Error; err != nil {
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
	c.HTML(http.StatusOK, "profileOrderHistory.html", gin.H{
		"status":  "success",
		"message": "Order details fetched successfully",
		"User":    userauth,
		"Order":   orderHistoryResponse,
	})
}
