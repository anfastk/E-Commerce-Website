package helper

import (
	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func FetchUserID(c *gin.Context) uint {
	tokenString, err := c.Cookie("jwtTokensUser")
	var userID uint
 
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.GetJwtKey(), nil
		})

		if err == nil && token.Valid && claims.Role == "User" {
			userID = claims.UserId

			var user models.UserAuth
			if err := config.DB.First(&user, userID).Error; err != nil || user.IsBlocked || user.IsDeleted {
				c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
			}
		}
	}
	return userID
}