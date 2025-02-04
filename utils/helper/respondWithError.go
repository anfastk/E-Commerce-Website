package helper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"status":  http.StatusText(status),
		"message": message,
		"code":    status,
	})
}
