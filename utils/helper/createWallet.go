package helper

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateWallet(c *gin.Context, userID uint) {
	logger.Log.Info("Creating wallet", zap.Uint("userID", userID))

	var wallet models.Wallet
	if err := config.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
		userWallet := models.Wallet{
			UserID:  userID,
			Balance: 0,
		}
		if createErr := config.DB.Create(&userWallet).Error; createErr != nil {
			logger.Log.Error("Failed to create wallet",
				zap.Uint("userID", userID),
				zap.Error(createErr))
			RespondWithError(c, http.StatusInternalServerError, "Wallet Creation Failed", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Wallet created successfully",
			zap.Uint("userID", userID),
			zap.Uint("walletID", userWallet.ID))
	} else {
		logger.Log.Debug("Wallet already exists",
			zap.Uint("userID", userID),
			zap.Uint("walletID", wallet.ID))
	}
}