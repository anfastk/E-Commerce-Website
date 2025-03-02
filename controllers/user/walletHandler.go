package controllers

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func CreateWallet(c *gin.Context,userID uint){
	var wallet models.Wallet
	if err:=config.DB.First(&wallet,"user_id = ?",userID).Error;err!=nil {
		userWallet := models.Wallet{
			UserID:  userID,
			Balance: 0,
		}
		if createErr := config.DB.Create(&userWallet).Error; createErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Creation Failed", "Something Went Wrong", "")
			return
		}
	}
}