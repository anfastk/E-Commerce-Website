package controllers

import (
	"errors"
	"net/http"
	"time" 

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
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
			c.Redirect(http.StatusSeeOther, "/admin/dashboard")
			return
		}
	}

	c.HTML(http.StatusOK, "adminLogin.html", nil)
}

type AdminInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AdminLoginHandler(c *gin.Context) {
	var admin models.AdminModel
	var input AdminInput

	if err := c.ShouldBind(&input); err != nil {
		logger.Log.Error("Failed TO Binding The Data,Invalid Data Entered", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Failed TO Binding The Data", "Invalid Data Entered", "")
		return
	}

	if input.Email == "" || input.Password == "" {
        logger.Log.Error("Email and Password are required")
		helper.RespondWithError(c, http.StatusBadRequest, "Email and Password are required", "Email and Password are required", "")
		return
	}

	if err := config.DB.Where("email = ?", input.Email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
            logger.Log.Error("Invalid Email", zap.Error(err))
			helper.RespondWithError(c, http.StatusBadRequest, "Invalid Email", "Invalid Email", "")
		} else {
            logger.Log.Error("Database query failed", zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Database query failed", "Something Went Wrong", "")
		}
		return
	}

	if admin.Password != input.Password {
        logger.Log.Error("Invalid Password")
		helper.RespondWithError(c, http.StatusUnauthorized, "Invalid Password", "Invalid Password", "")
		return
	}

	token, err := middleware.GenerateJWT(admin.ID, admin.Email, RoleAdmin)
	if err != nil {
        logger.Log.Error("Failed to generate JWT tokens", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to generate JWT tokens", "Failed to generate JWT tokens", "")
		return
	}

	c.SetCookie("jwtTokensAdmin", token, int((time.Hour * 6).Seconds()), "/", "", false, true)

    logger.Log.Info("Admin Logined successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
		"code":    http.StatusOK,
	})
}

func ShowSettings(c *gin.Context) {
    logger.Log.Info("Admin Settings Open successfully")
	c.HTML(http.StatusOK, "settings.html", nil)
}

func AdminLogoutHandler(c *gin.Context) {
	c.SetCookie("jwtTokensAdmin", "", -1, "/", "", false, true)
    logger.Log.Info("Admin Logout successfully")
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Admin logged out successfully",
		"code":    http.StatusOK,
	})
}
