package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

var Store = sessions.NewCookieStore([]byte("laptixsecretkey"))

func SendOtp(c *gin.Context) {

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
	}

	session, _ := Store.Get(c.Request, "session")
	email, exists := session.Values["email"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Session expired",
			"code":   400,
		})
		return
	}

	otp := utils.GenerateOTP(6)
	expiry := time.Now().Add(5 * time.Minute)

	var otpRecord models.Otp
	otpRecord.Email = email
	otpRecord.OTP = otp
	otpRecord.ExpireTime = expiry

	if err := config.DB.Create(&otpRecord).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to store OTP",
			"code":   500,
		})
		return
	}

	if err := utils.SendOTPToEmail(email, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to send OTP",
			"code":   500,
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/user/signup/verifyotp")
}

func VerifyOtp(c *gin.Context) {
	var otpInput struct {
		Email string `form:"email"`
		OTP   string `form:"otp"`
	}

	if err := c.ShouldBind(&otpInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var otpRecord models.Otp
	if err := config.DB.Where("email = ? AND otp = ?", otpInput.Email, otpInput.OTP).Order("created_at DESC").First(&otpRecord).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	if time.Now().After(otpRecord.ExpireTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP has expired"})
		return
	}

	session, _ := Store.Get(c.Request, "session")
	fullName, _ := session.Values["full_name"].(string)
	hashedPassword, _ := session.Values["password"].(string)

	user := models.UserAuth{
		FullName: fullName,
		Email:    otpInput.Email,
		Password: hashedPassword,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	session.Options.MaxAge = -1
	session.Save(c.Request, c.Writer)

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "OTP verified",
		"code":    http.StatusOK,
	})
}

func ResendOTP(c *gin.Context) {
	session, _ := Store.Get(c.Request, "session")
	email, exists := session.Values["email"].(string)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session expired. Please restart the signup process."})
		return
	}

	otp := utils.GenerateOTP(6)
	expiry := time.Now().Add(5 * time.Minute)

	var otpRecord models.Otp
	err := config.DB.Where("email = ?", email).First(&otpRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			otpRecord = models.Otp{
				Email:      email,
				OTP:        otp,
				ExpireTime: expiry,
			}
			if err := config.DB.Create(&otpRecord).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create OTP record"})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve OTP record"})
			return
		}
	} else {
		otpRecord.OTP = otp
		otpRecord.ExpireTime = expiry
		if err := config.DB.Save(&otpRecord).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update OTP record"})
			return
		}
	}

	if err := utils.SendOTPToEmail(email, otp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}
	c.JSON(http.StatusOK,gin.H{
		"status":"Status OK",
		"message":"OTP resend successfully",
		"code":http.StatusOK,
	})
}