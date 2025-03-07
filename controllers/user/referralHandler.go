package controllers

import (
	"fmt"
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


func ShowReferralPage(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var refferal models.ReferralAccount
	if err := config.DB.First(&refferal, "user_id = ?", userauth.ID).Error; err != nil {
		createReferral := models.ReferralAccount{
			UserID: userauth.ID,
		}
		if err := config.DB.Create(&createReferral).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Referral Account Creation Failed", "Something Went Wrong", "")
			return
		}
	}

	var referalHistory []models.ReferalHistory
	config.DB.Preload("JoinedUser").Find(&referalHistory, "referral_id = ?", refferal.ID)
	CheckForReferrer(c)
	CheckForJoinee(c)

	c.HTML(http.StatusOK, "profileReferral.html", gin.H{
		"User":            userauth,
		"ReferralAccount": refferal,
		"ReferalHistory":  referalHistory,
	})
}

func AddReferral(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	var inputReferralCode struct {
		ReferralCode string `json:"referralCode"`
	}

	if err := c.ShouldBindJSON(&inputReferralCode); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "invalid data", "Enter Correct Code", "")
		return
	}

	tx := config.DB.Begin()

	var userauth models.UserAuth
	if err := tx.First(&userauth, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "Login First", "")
		return
	}

	if userauth.IsRefered {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Oops! You Have Already Been Referred.", "Oops! You Have Already Been Referred.", "")
		return
	}

	var referrer models.UserAuth
	if err := tx.First(&referrer, "referral_code = ? ", inputReferralCode.ReferralCode).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Referral Code", "Invalid Referral Code", "")
		return
	}

	if userauth.ReferralCode == referrer.ReferralCode {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Referral Code", "Don't Use Your Own Referral Code", "")
		return
	}

	userauth.IsRefered = true
	if err := tx.Save(&userauth).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Updateing Failed", "Something Went Wrong", "")
		return
	}

	var referrerAccountDetails models.ReferralAccount
	if err := tx.First(&referrerAccountDetails, "user_id = ? ", referrer.ID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
		return
	}

	referrerAccountDetails.Count += 1

	if err := tx.Save(&referrerAccountDetails).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Updateing Failed", "Something Went Wrong", "")
		return
	}

	createHistory := models.ReferalHistory{
		ReferralID:   referrerAccountDetails.ID,
		JoinedUserId: userauth.ID,
		Status:       "Pending",
		Reward:       250,
	}

	if err := tx.Create(&createHistory).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed To Create History", "Something Went Wrong", "")
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Referral Code Added Successfully",
		"code":    http.StatusOK,
	})
}

func CheckForReferrer(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	tx := config.DB.Begin()

	var userauth models.UserAuth
	if err := tx.First(&userauth, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "Login First", "")
		return
	}

	var referralAccountDetails models.ReferralAccount
	if err := tx.First(&referralAccountDetails, "user_id = ? ", userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
		return
	}

	var referralHistory []models.ReferalHistory
	tx.Find(&referralHistory, "referral_id = ? AND status = ?", referralAccountDetails.ID, "Pending")

	for _, refRow := range referralHistory {
		var order models.OrderItem
		if err := tx.First(&order, "user_id = ? AND order_status = ?", refRow.JoinedUserId, "Delivered").Error; err == nil {
			var joineDetails models.UserAuth
			if err := tx.First(&joineDetails, refRow.JoinedUserId).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Joinee Not Found", "Something Went Wrong", "")
				return
			}
			refRow.Status = "Complete"
			if saveErr := tx.Save(&refRow).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral History Update Failed", "Something Went Wrong", "")
				return
			}
			referralAccountDetails.Balance += 250
			if saveErr := tx.Save(&referralAccountDetails).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral Account Details Update Failed", "Something Went Wrong", "")
				return
			}
			var referrerWallet models.Wallet
			if err := tx.First(&referrerWallet, "user_id = ? ", userID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}
			referrerWallet.Balance += 250
			if saveErr := tx.Save(&referrerWallet).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}
			receiptID := "rcpt_" + uuid.New().String()
			transactionID := "TXN_" + uuid.New().String()
			createReferrerWalletHistory := models.WalletTransaction{
				UserID:        userID,
				WalletID:      referrerWallet.ID,
				Amount:        250,
				Description:   fmt.Sprintf("Referral Bonus - " + joineDetails.FullName + " Joined"),
				Type:          "Referral",
				Receipt:       receiptID,
				TransactionID: transactionID,
			}
			if createErr := tx.Create(&createReferrerWalletHistory).Error; createErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}
			var joineeWallet models.Wallet
			if err := tx.First(&joineeWallet, "user_id = ? ", joineDetails.ID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}
			joineeWallet.Balance += 100
			if saveErr := tx.Save(&joineeWallet).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}
			createJoineeWalletHistory := models.WalletTransaction{
				UserID:        joineDetails.ID,
				WalletID:      joineeWallet.ID,
				Amount:        100,
				Description:   fmt.Sprintf("Referral Bonus - " + userauth.FullName + " Added You"),
				Type:          "Referral",
				Receipt:       receiptID,
				TransactionID: transactionID,
			}
			if createErr := tx.Create(&createJoineeWalletHistory).Error; createErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}
		}
	}

	tx.Commit()
}

func CheckForJoinee(c *gin.Context) {
	userID := c.MustGet("userid").(uint)

	tx := config.DB.Begin()

	var userauth models.UserAuth
	if err := tx.First(&userauth, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "Login First", "")
		return
	}

	var referralHistory models.ReferalHistory
	err := tx.First(&referralHistory, "joined_user_id = ? AND status = ?", userID, "Pending")
	if err.Error == nil {
		var order models.OrderItem
		if err := tx.First(&order, "user_id = ? AND order_status = ?", userID, "Delivered").Error; err == nil {
			var joineeWallet models.Wallet
			if err := tx.First(&joineeWallet, "user_id = ? ", userID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}
			joineeWallet.Balance += 100
			if saveErr := tx.Save(&joineeWallet).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}
			receiptID := "rcpt_" + uuid.New().String()
			transactionID := "TXN_" + uuid.New().String()
			createJoineeWalletHistory := models.WalletTransaction{
				UserID:        userID,
				WalletID:      joineeWallet.ID,
				Amount:        100,
				Description:   fmt.Sprintf("Referral Bonus - " + userauth.FullName + " Added You"),
				Type:          "Referral",
				Receipt:       receiptID,
				TransactionID: transactionID,
			}
			if createErr := tx.Create(&createJoineeWalletHistory).Error; createErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}
			var referrerAccountDetails models.ReferralAccount
			if err := tx.First(&referrerAccountDetails, referralHistory.ReferralID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}
			var referrerDetails models.UserAuth
			if err := tx.First(&referrerDetails, referrerAccountDetails.UserID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Not Found", "Something Went Wrong", "")
				return
			}
			referralHistory.Status = "Complete"
			if saveErr := tx.Save(&referralHistory).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral History Update Failed", "Something Went Wrong", "")
				return
			}
			referrerAccountDetails.Balance += 250
			if saveErr := tx.Save(&referrerAccountDetails).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral Account Details Update Failed", "Something Went Wrong", "")
				return
			}
			var referrerWallet models.Wallet
			if err := tx.First(&referrerWallet, "user_id = ? ", referrerDetails.ID).Error; err != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}
			referrerWallet.Balance += 250
			if saveErr := tx.Save(&referrerWallet).Error; saveErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}
			createReferrerWalletHistory := models.WalletTransaction{
				UserID:        referrerDetails.ID,
				WalletID:      referrerWallet.ID,
				Amount:        250,
				Description:   fmt.Sprintf("Referral Bonus - " + userauth.FullName + " Joined"),
				Type:          "Referral",
				Receipt:       receiptID,
				TransactionID: transactionID,
			}
			if createErr := tx.Create(&createReferrerWalletHistory).Error; createErr != nil {
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}
		}
	}

	tx.Commit()
}
