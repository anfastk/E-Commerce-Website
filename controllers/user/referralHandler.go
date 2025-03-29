package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)
 
func ShowReferralPage(c *gin.Context) {
	logger.Log.Info("Showing referral page")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	var referral models.ReferralAccount
	if err := config.DB.First(&referral, "user_id = ?", userauth.ID).Error; err != nil {
		logger.Log.Warn("Referral account not found, creating new",
			zap.Uint("userID", userauth.ID))
		CreateReferralAccount(c, userauth.ID)
		// Re-fetch after creation
		if err := config.DB.First(&referral, "user_id = ?", userauth.ID).Error; err != nil {
			logger.Log.Error("Failed to fetch newly created referral account",
				zap.Uint("userID", userauth.ID),
				zap.Error(err))
			return
		}
	}

	var referralHistory []models.ReferalHistory
	if err := config.DB.Preload("JoinedUser").Find(&referralHistory, "referral_id = ?", referral.ID).Error; err != nil {
		logger.Log.Warn("Failed to fetch referral history",
			zap.Uint("referralID", referral.ID),
			zap.Error(err))
	}

	CheckForReferrer(c)
	CheckForJoinee(c)

	logger.Log.Info("Referral page loaded",
		zap.Uint("userID", userID),
		zap.Int("referralHistoryCount", len(referralHistory)))
	c.HTML(http.StatusOK, "profileReferral.html", gin.H{
		"User":            userauth,
		"ReferralAccount": referral,
		"ReferalHistory":  referralHistory,
	})
}

func CreateReferralAccount(c *gin.Context, userID uint) {
	logger.Log.Info("Creating referral account", zap.Uint("userID", userID))

	createReferral := models.ReferralAccount{
		UserID: userID,
	}
	if err := config.DB.Create(&createReferral).Error; err != nil {
		logger.Log.Error("Failed to create referral account",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Referral Account Creation Failed", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Referral account created successfully",
		zap.Uint("userID", userID),
		zap.Uint("referralID", createReferral.ID))
}

func AddReferral(c *gin.Context) {
	logger.Log.Info("Adding referral")

	userIDInterface, exists := c.Get("userid")
	if !exists {
		logger.Log.Error("User ID not found in context")
		helper.RespondWithError(c, http.StatusBadRequest, "Unauthorized", "Login First", "")
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		logger.Log.Error("Invalid user ID type", zap.Any("userID", userIDInterface))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid user ID type", "Something Went Wrong", "")
		return
	}
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var inputReferralCode struct {
		ReferralCode string `json:"referralCode"`
	}
	if err := c.ShouldBindJSON(&inputReferralCode); err != nil {
		logger.Log.Error("Failed to bind referral code data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "invalid data", "Enter Correct Code", "")
		return
	}

	tx := config.DB.Begin()

	var userauth models.UserAuth
	if err := tx.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "Login First", "")
		return
	}

	if userauth.IsRefered {
		logger.Log.Warn("User already referred",
			zap.Uint("userID", userID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Oops! You Have Already Been Referred.", "Oops! You Have Already Been Referred.", "")
		return
	}

	var referrer models.UserAuth
	if err := tx.First(&referrer, "referral_code = ? ", inputReferralCode.ReferralCode).Error; err != nil {
		logger.Log.Error("Invalid referral code",
			zap.String("referralCode", inputReferralCode.ReferralCode),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Referral Code", "Invalid Referral Code", "")
		return
	}

	if userauth.ReferralCode == referrer.ReferralCode {
		logger.Log.Warn("User attempted to use own referral code",
			zap.Uint("userID", userID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Referral Code", "Don't Use Your Own Referral Code", "")
		return
	}

	userauth.IsRefered = true
	if err := tx.Save(&userauth).Error; err != nil {
		logger.Log.Error("Failed to update user referral status",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Updateing Failed", "Something Went Wrong", "")
		return
	}

	var referrerAccountDetails models.ReferralAccount
	if err := tx.First(&referrerAccountDetails, "user_id = ? ", referrer.ID).Error; err != nil {
		logger.Log.Warn("Referrer account not found, creating new",
			zap.Uint("referrerID", referrer.ID))
		CreateReferralAccount(c, referrer.ID)
		// Re-fetch after creation
		if err := tx.First(&referrerAccountDetails, "user_id = ? ", referrer.ID).Error; err != nil {
			logger.Log.Error("Failed to fetch newly created referrer account",
				zap.Uint("referrerID", referrer.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}
	}

	referrerAccountDetails.Count += 1
	if err := tx.Save(&referrerAccountDetails).Error; err != nil {
		logger.Log.Error("Failed to update referrer account count",
			zap.Uint("referralID", referrerAccountDetails.ID),
			zap.Error(err))
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
		logger.Log.Error("Failed to create referral history",
			zap.Uint("referralID", referrerAccountDetails.ID),
			zap.Uint("joinedUserID", userauth.ID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed To Create History", "Something Went Wrong", "")
		return
	}

	tx.Commit()
	logger.Log.Info("Referral added successfully",
		zap.Uint("userID", userID),
		zap.Uint("referrerID", referrer.ID),
		zap.String("referralCode", inputReferralCode.ReferralCode))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Referral Code Added Successfully",
		"code":    http.StatusOK,
	})
}

func CheckForReferrer(c *gin.Context) {
	logger.Log.Info("Checking for referrer rewards")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	tx := config.DB.Begin()

	var userauth models.UserAuth
	if err := tx.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "Login First", "")
		return
	}

	var referralAccountDetails models.ReferralAccount
	if err := tx.First(&referralAccountDetails, "user_id = ? ", userID).Error; err != nil {
		logger.Log.Error("Referral account not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Account not found", "Something Went Wrong", "")
		return
	}

	var referralHistory []models.ReferalHistory
	if err := tx.Find(&referralHistory, "referral_id = ? AND status = ?", referralAccountDetails.ID, "Pending").Error; err != nil {
		logger.Log.Warn("No pending referral history found",
			zap.Uint("referralID", referralAccountDetails.ID),
			zap.Error(err))
	}

	for _, refRow := range referralHistory {
		var order models.OrderItem
		if err := tx.First(&order, "created_at >= ? AND user_id = ? AND order_status = ?", refRow.CreatedAt, refRow.JoinedUserId, "Delivered").Error; err == nil {
			var joineeDetails models.UserAuth
			if err := tx.First(&joineeDetails, refRow.JoinedUserId).Error; err != nil {
				logger.Log.Error("Joinee not found",
					zap.Uint("joinedUserID", refRow.JoinedUserId),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Joinee Not Found", "Something Went Wrong", "")
				return
			}

			refRow.Status = "Complete"
			if saveErr := tx.Save(&refRow).Error; saveErr != nil {
				logger.Log.Error("Failed to update referral history status",
					zap.Uint("referralHistoryID", refRow.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral History Update Failed", "Something Went Wrong", "")
				return
			}

			referralAccountDetails.Balance += 250
			if saveErr := tx.Save(&referralAccountDetails).Error; saveErr != nil {
				logger.Log.Error("Failed to update referral account balance",
					zap.Uint("referralID", referralAccountDetails.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral Account Details Update Failed", "Something Went Wrong", "")
				return
			}

			var referrerWallet models.Wallet
			if err := tx.First(&referrerWallet, "user_id = ? ", userID).Error; err != nil {
				logger.Log.Error("Referrer wallet not found",
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}

			lastBalance := referrerWallet.Balance
			referrerWallet.Balance += 250
			if saveErr := tx.Save(&referrerWallet).Error; saveErr != nil {
				logger.Log.Error("Failed to update referrer wallet",
					zap.Uint("walletID", referrerWallet.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}

			receiptID := "rcpt_" + uuid.New().String()
			rand.Seed(time.Now().UnixNano())
			transactionID := fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
			createReferrerWalletHistory := models.WalletTransaction{
				UserID:        userID,
				WalletID:      referrerWallet.ID,
				Amount:        250,
				Description:   fmt.Sprintf("Referral Bonus - %s Joined", joineeDetails.FullName),
				Type:          "Referral",
				Receipt:       receiptID,
				LastBalance:   lastBalance,
				TransactionID: strings.ToUpper(transactionID),
			}
			if createErr := tx.Create(&createReferrerWalletHistory).Error; createErr != nil {
				logger.Log.Error("Failed to create referrer wallet transaction",
					zap.Uint("walletID", referrerWallet.ID),
					zap.Error(createErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}

			var joineeWallet models.Wallet
			if err := tx.First(&joineeWallet, "user_id = ? ", joineeDetails.ID).Error; err != nil {
				logger.Log.Error("Joinee wallet not found",
					zap.Uint("joinedUserID", joineeDetails.ID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}

			lastBalanceJoinee := joineeWallet.Balance
			joineeWallet.Balance += 100
			if saveErr := tx.Save(&joineeWallet).Error; saveErr != nil {
				logger.Log.Error("Failed to update joinee wallet",
					zap.Uint("walletID", joineeWallet.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}

			createJoineeWalletHistory := models.WalletTransaction{
				UserID:        joineeDetails.ID,
				WalletID:      joineeWallet.ID,
				Amount:        100,
				Description:   fmt.Sprintf("Referral Bonus - %s Added You", userauth.FullName),
				Type:          "Referral",
				LastBalance:   lastBalanceJoinee,
				Receipt:       receiptID,
				TransactionID: transactionID,
			}
			if createErr := tx.Create(&createJoineeWalletHistory).Error; createErr != nil {
				logger.Log.Error("Failed to create joinee wallet transaction",
					zap.Uint("walletID", joineeWallet.ID),
					zap.Error(createErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}

			logger.Log.Info("Referral reward processed",
				zap.Uint("referrerID", userID),
				zap.Uint("joineeID", joineeDetails.ID),
				zap.Uint("referralHistoryID", refRow.ID))
		}
	}

	tx.Commit()
	logger.Log.Info("Referrer check completed",
		zap.Uint("userID", userID),
		zap.Int("processedReferrals", len(referralHistory)))
}

func CheckForJoinee(c *gin.Context) {
	logger.Log.Info("Checking for joinee rewards")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	tx := config.DB.Begin()

	var userauth models.UserAuth
	if err := tx.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "Login First", "")
		return
	}

	var referralHistory models.ReferalHistory
	err := tx.First(&referralHistory, "joined_user_id = ? AND status = ?", userID, "Pending")
	if err.Error == nil {
		var order models.OrderItem
		if err := tx.First(&order, "created_at >= ? AND user_id = ? AND order_status = ?", referralHistory.CreatedAt, userID, "Delivered").Error; err == nil {
			var joineeWallet models.Wallet
			if err := tx.First(&joineeWallet, "user_id = ? ", userID).Error; err != nil {
				logger.Log.Error("Joinee wallet not found",
					zap.Uint("userID", userID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}

			lastBalance := joineeWallet.Balance
			joineeWallet.Balance += 100
			if saveErr := tx.Save(&joineeWallet).Error; saveErr != nil {
				logger.Log.Error("Failed to update joinee wallet",
					zap.Uint("walletID", joineeWallet.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}

			receiptID := "rcpt_" + uuid.New().String()
			rand.Seed(time.Now().UnixNano())
			transactionID := fmt.Sprintf("TXN-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
			createJoineeWalletHistory := models.WalletTransaction{
				UserID:        userID,
				WalletID:      joineeWallet.ID,
				Amount:        100,
				Description:   fmt.Sprintf("Referral Bonus - %s Added You", userauth.FullName),
				Type:          "Referral",
				Receipt:       receiptID,
				LastBalance:   lastBalance,
				TransactionID: strings.ToUpper(transactionID),
			}
			if createErr := tx.Create(&createJoineeWalletHistory).Error; createErr != nil {
				logger.Log.Error("Failed to create joinee wallet transaction",
					zap.Uint("walletID", joineeWallet.ID),
					zap.Error(createErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}

			var referrerAccountDetails models.ReferralAccount
			if err := tx.First(&referrerAccountDetails, referralHistory.ReferralID).Error; err != nil {
				logger.Log.Error("Referrer account not found",
					zap.Uint("referralID", referralHistory.ReferralID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusNotFound, "Account not found", "Something Went Wrong", "")
				return
			}

			var referrerDetails models.UserAuth
			if err := tx.First(&referrerDetails, referrerAccountDetails.UserID).Error; err != nil {
				logger.Log.Error("Referrer not found",
					zap.Uint("referrerID", referrerAccountDetails.UserID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Not Found", "Something Went Wrong", "")
				return
			}

			referralHistory.Status = "Complete"
			if saveErr := tx.Save(&referralHistory).Error; saveErr != nil {
				logger.Log.Error("Failed to update referral history status",
					zap.Uint("referralHistoryID", referralHistory.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral History Update Failed", "Something Went Wrong", "")
				return
			}

			referrerAccountDetails.Balance += 250
			if saveErr := tx.Save(&referrerAccountDetails).Error; saveErr != nil {
				logger.Log.Error("Failed to update referrer account balance",
					zap.Uint("referralID", referrerAccountDetails.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referral Account Details Update Failed", "Something Went Wrong", "")
				return
			}

			var referrerWallet models.Wallet
			if err := tx.First(&referrerWallet, "user_id = ? ", referrerDetails.ID).Error; err != nil {
				logger.Log.Error("Referrer wallet not found",
					zap.Uint("referrerID", referrerDetails.ID),
					zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Referrer Details Not Found", "Something Went Wrong", "")
				return
			}

			lastBalanceReferrer := referrerWallet.Balance
			referrerWallet.Balance += 250
			if saveErr := tx.Save(&referrerWallet).Error; saveErr != nil {
				logger.Log.Error("Failed to update referrer wallet",
					zap.Uint("walletID", referrerWallet.ID),
					zap.Error(saveErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Update Failed", "Something Went Wrong", "")
				return
			}

			createReferrerWalletHistory := models.WalletTransaction{
				UserID:        referrerDetails.ID,
				WalletID:      referrerWallet.ID,
				Amount:        250,
				Description:   fmt.Sprintf("Referral Bonus - %s Joined", userauth.FullName),
				Type:          "Referral",
				Receipt:       receiptID,
				LastBalance:   lastBalanceReferrer,
				TransactionID: transactionID,
			}
			if createErr := tx.Create(&createReferrerWalletHistory).Error; createErr != nil {
				logger.Log.Error("Failed to create referrer wallet transaction",
					zap.Uint("walletID", referrerWallet.ID),
					zap.Error(createErr))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Added Failed", "Something Went Wrong", "")
				return
			}

			logger.Log.Info("Joinee referral reward processed",
				zap.Uint("joineeID", userID),
				zap.Uint("referrerID", referrerDetails.ID),
				zap.Uint("referralHistoryID", referralHistory.ID))
		}
	}

	tx.Commit()
	logger.Log.Info("Joinee check completed",
		zap.Uint("userID", userID))
}
