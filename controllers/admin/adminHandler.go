package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
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
			c.Redirect(http.StatusSeeOther, "/admin/products")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data")
		return
	}

	if input.Email == "" || input.Password == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Email and Password are required")
		return
	}

	if err := config.DB.Where("email = ?", input.Email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.RespondWithError(c, http.StatusUnauthorized, "Invalid Email")
		} else {
			helper.RespondWithError(c, http.StatusInternalServerError, "Database query failed")
		}
		return
	}

	if admin.Password != input.Password {
		helper.RespondWithError(c, http.StatusUnauthorized, "Invalid Password")
		return
	}

	token, err := middleware.GenerateJWT(admin.ID, admin.Email, RoleAdmin)
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to generate JWT tokens")
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