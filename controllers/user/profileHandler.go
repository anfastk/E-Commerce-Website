package controllers

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
)

func Settings(c *gin.Context){
	userID := c.MustGet("userid").(uint)
	
	var userDetails models.UserAuth
	if err:=config.DB.First(&userDetails,"id=? AND is_blocked = ?",userID,false).Error;err!=nil {
		c.Redirect(http.StatusSeeOther,"/user/login")
		return
	}
	
	c.HTML(http.StatusOK,"profileSettings.html",gin.H{
		"User":userDetails,
	})
}