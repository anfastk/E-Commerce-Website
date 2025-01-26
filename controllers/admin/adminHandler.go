package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

var RoleAdmin = "Admin"

func ShowLoginPage(c *gin.Context) {
	tokenString, err := c.Cookie("jwtTokensAdmin")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			c.Redirect(http.StatusSeeOther, "/admin/users")
			return
		}
	}

	c.HTML(http.StatusOK,"adminLogin.html", nil)
}

type AdminInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AdminLoginHandler(c *gin.Context) {
	var admin models.AdminModel
	var input AdminInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Binding the data",
			"code":   400,
		})
		return
	}

	if input.Email == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Email and Password are required",
			"code":   400,
		})
		return
	}

	if err := config.DB.Where("email = ?", input.Email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "Unauthorized",
				"error":  "Invalid Email",
				"code":   401,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal Server Error",
				"error":  "Database query failed",
				"code":   500,
			})
		}
		return
	}

	if admin.Password != input.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "Unauthorized",
			"error":  "Invalid Password",
			"code":   401,
		})
		return
	}

	token, err := middleware.GenerateJWT(admin.ID, admin.Email, RoleAdmin)
	if err != nil {
		fmt.Println("Error for generating JWT tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to generate JWT tokens",
			"code":   "500",
		})
		return
	}

	c.SetCookie("jwtTokensAdmin", token, int((time.Hour * 3).Seconds()), "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
		"code":    http.StatusOK,
	})
}

func ShowSettings(c * gin.Context){
	c.HTML(http.StatusOK,"settings.html",nil)
}

func AdminLogoutHandler(c * gin.Context){

	c.SetCookie("jwtTokensAdmin", "",-1, "/", "", false, true)


	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Admin logged out successfully",
		"code":    http.StatusOK,
	})
}