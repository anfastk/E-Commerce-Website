package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserId uint
	Email  string `json:"username"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

var JwtSecretKey = []byte(os.Getenv("SECRETKEY"))

func JwtTocken(c *gin.Context, userId uint, email string, role string) {
	tokenString, err := GenerateJWT(userId, email, role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Failed to generate jwt tocken",
			"code":   400,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Status": "OK",
		"token":  tokenString,
		"code":   "200",
	})
}

func GenerateJWT(userId uint, email string, role string) (string, error) {
	claims := Claims{
		UserId: uint(userId),
		Email:  email,
		Role:   role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtSecretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("jwtTokens" + requiredRole)

		if err != nil || tokenString == "" {
			if requiredRole == "Admin" || requiredRole == "User" {
				redirectPath := fmt.Sprintf("/%s/login", strings.ToLower(requiredRole))
				c.Redirect(http.StatusSeeOther, redirectPath)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  "Unauthorized",
					"message": "Can't find cookie",
					"code":    401,
				})
			}
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtSecretKey, nil
		})
		var userDetails models.UserAuth
		if err:=config.DB.First(&userDetails,"id=? AND is_blocked = ?",claims.UserId,false).Error;err!=nil {
			c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
		}

		if err != nil || !token.Valid {
			if requiredRole == "Admin" || requiredRole == "User" {
				redirectPath := fmt.Sprintf("/%s/login", strings.ToLower(requiredRole))
				c.Redirect(http.StatusSeeOther, redirectPath)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"status":  "Unauthorized",
					"message": "Invalid or expired JWT Token.",
					"code":    401,
				})
			}
			c.Abort()
			return
		}

		if claims.Role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"status": "Forbidden",
				"error":  "Insufficient permissions",
				"code":   403,
			})
			c.Abort()
			return
		}

		c.Set("userid", claims.UserId)
		c.Next()
	}
}

func NoCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}
