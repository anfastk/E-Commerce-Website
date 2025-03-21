package middleware

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			if c.Writer.Status() == http.StatusNotFound {
				c.HTML(http.StatusNotFound, "404.html", nil)
			} else {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			}
		}
	}
}

func DBRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.IsConfigErr && config.ConfigErr != nil {
			logger.Log.Error("Error Found In Config", zap.Error(config.ConfigErr))
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": config.ConfigErr,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
