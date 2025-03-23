package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ShowContactUs(c *gin.Context){
	c.HTML(http.StatusOK,"contactUs.html",nil)
}