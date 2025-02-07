package helper

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, status int, error string, message string, redirect string) {
	c.JSON(status, gin.H{
		"status":   http.StatusText(status),
		"error":    error,
		"message":  message,
		"code":     status,
		"redirect": redirect,
	})
}
