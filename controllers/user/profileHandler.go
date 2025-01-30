package controllers

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
)

func ProfileSettings(c *gin.Context){
	userID := c.MustGet("userid").(uint)
	
	var userDetails models.UserAuth
	if err:=config.DB.First(&userDetails,userID).Error;err!=nil {
		c.JSON(http.StatusNotFound,gin.H{
			"status":"Not Found",
			"error":"User not found",
			"code":http.StatusNotFound,
		})
	}
	
	c.HTML(http.StatusOK,"profileSettings.html",gin.H{
		"User":userDetails,
	})
}