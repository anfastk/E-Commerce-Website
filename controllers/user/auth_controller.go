package controllers

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
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
		helper.RespondWithError(c, http.StatusBadRequest, "Session expired", "Session expired", "")
		return
	}

	otp := utils.GenerateOTP(6)
	expiry := time.Now().Add(5 * time.Minute)

	var otpRecord models.Otp
	otpRecord.Email = email
	otpRecord.OTP = otp
	otpRecord.ExpireTime = expiry

	if err := config.DB.Create(&otpRecord).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to store OTP", "Failed to store OTP", "")
		return
	}

	if err := utils.SendOTPToEmail(email, otp); err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to send OTP", "Failed to send OTP", "")
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
		helper.RespondWithError(c, http.StatusBadRequest, err.Error(), err.Error(), "")
		return
	}

	var otpRecord models.Otp
	if err := config.DB.Where("email = ? AND otp = ?", otpInput.Email, otpInput.OTP).Order("created_at DESC").First(&otpRecord).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid OTP", "Invalid OTP", "")
		return
	}

	if time.Now().After(otpRecord.ExpireTime) {
		helper.RespondWithError(c, http.StatusBadRequest, "OTP has expired", "OTP has expired", "")
		return
	}

	session, _ := Store.Get(c.Request, "session")
	var userAuth models.UserAuth
	if err := config.DB.First(&userAuth, "email = ?", otpInput.Email).Error; err != nil {
		session, _ := Store.Get(c.Request, "session")
		fullName, _ := session.Values["full_name"].(string)
		hashedPassword, _ := session.Values["password"].(string)

		userAuth := models.UserAuth{
			FullName: fullName,
			Email:    otpInput.Email,
			Password: hashedPassword,
			ProfilePic: os.Getenv("DEFAULT_PROFILE_PIC"),
		}

		if err := config.DB.Create(&userAuth).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create user", "Failed to create user", "")
			return
		}
	} else {
		session, _ := Store.Get(c.Request, "session")
		hashedPassword, _ := session.Values["password"].(string)

		userAuth.Password = hashedPassword

		if err := config.DB.Model(&userAuth).
			Where("id = ?", userAuth.ID).
			Updates(map[string]interface{}{"password": hashedPassword}).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to change password", "Failed to change password", "")
			return
		}
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
		helper.RespondWithError(c, http.StatusBadRequest, "Session expired", "Session expired. Please restart the signup process.", "")
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
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create OTP record", "Failed to create OTP record", "")
				return
			}
		} else {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve OTP record", "Failed to retrieve OTP record", "")
			return
		}
	} else {
		otpRecord.OTP = otp
		otpRecord.ExpireTime = expiry
		if err := config.DB.Save(&otpRecord).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update OTP record", "Failed to update OTP record", "")
			return
		}
	}

	if err := utils.SendOTPToEmail(email, otp); err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to send OTP", "Failed to send OTP", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "OTP resend successfully",
		"code":    http.StatusOK,
	})
}
