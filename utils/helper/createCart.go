package helper

import (
	"errors"
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreateCart(c *gin.Context, userID uint) {
	var cart models.Cart
	if err := config.DB.First(&cart, "user_id = ?", userID).Error; err != nil {
		cart = models.Cart{
			UserID: userID,
		}
		if createErr := config.DB.Create(&cart).Error; createErr != nil {
			logger.Log.Error("Failed to create cart", zap.Uint("userID", userID), zap.Error(createErr))
			config.DB.Rollback()
			err := errors.New("Cart creation Failed")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		logger.Log.Info("Created new cart", zap.Uint("cartID", cart.ID))
	}
}
