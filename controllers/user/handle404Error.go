package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Handle404Error(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", nil)
}