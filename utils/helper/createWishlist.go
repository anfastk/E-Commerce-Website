package helper

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateWishlist(c *gin.Context, userID uint) {
	var wishlist models.Wishlist
	if err := config.DB.First(&wishlist, "user_id = ?", userID).Error; err != nil {
		wishlist = models.Wishlist{UserID: userID}
		if createErr := config.DB.Create(&wishlist).Error; createErr != nil {
			logger.Log.Error("Failed to create wishlist",
				zap.Uint("userID", userID),
				zap.Error(createErr))
			RespondWithError(c, http.StatusInternalServerError, "Failed to create wishlist", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Wishlist created",
			zap.Uint("userID", userID),
			zap.Uint("wishlistID", wishlist.ID))
	}
}
