package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/sessions"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Store = sessions.NewCookieStore([]byte("laptixsecretkey"))

func SendOtp(c *gin.Context) {
	logger.Log.Info("Requested to send OTP")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already authenticated, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
	}

	session, _ := Store.Get(c.Request, "session")
	email, exists := session.Values["email"].(string)
	if !exists {
		logger.Log.Warn("No email in session, redirecting to login")
		c.Redirect(http.StatusUnauthorized, "/auth/login")
		return
	}

	otp := utils.GenerateOTP(6)
	expiry := time.Now().Add(5 * time.Minute)

	var otpRecord models.Otp
	otpRecord.Email = email
	otpRecord.OTP = otp
	otpRecord.ExpireTime = expiry

	if err := config.DB.Create(&otpRecord).Error; err != nil {
		logger.Log.Error("Failed to store OTP in database",
			zap.String("email", email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to store OTP", "Something Went Wrong,Please Try Again ", "")
		return
	}

	if err := utils.SendOTPToEmail(email, otp); err != nil {
		logger.Log.Error("Failed to send OTP email",
			zap.String("email", email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to send OTP", "Failed To Send OTP , Please Try Again ", "")
		return
	}

	logger.Log.Info("OTP sent successfully",
		zap.String("email", email),
		zap.String("otp", otp))
	c.Redirect(http.StatusSeeOther, "/auth/signup/verifyotp")
}

func VerifyOtp(c *gin.Context) {
	logger.Log.Info("Requested to verify OTP")

	var otpInput struct {
		Email string `form:"email"`
		OTP   string `form:"otp"`
	}

	if err := c.ShouldBind(&otpInput); err != nil {
		logger.Log.Error("Failed to bind OTP input",
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, err.Error(), err.Error(), "")
		return
	}

	var otpRecord models.Otp
	if err := config.DB.Where("email = ? AND otp = ?", otpInput.Email, otpInput.OTP).Order("created_at DESC").First(&otpRecord).Error; err != nil {
		logger.Log.Warn("Invalid OTP provided",
			zap.String("email", otpInput.Email),
			zap.String("otp", otpInput.OTP),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid OTP", "Invalid OTP", "")
		return
	}

	if time.Now().After(otpRecord.ExpireTime) {
		logger.Log.Warn("OTP has expired",
			zap.String("email", otpInput.Email),
			zap.String("otp", otpInput.OTP))
		helper.RespondWithError(c, http.StatusBadRequest, "OTP has expired", "OTP has expired", "")
		return
	}

	session, _ := Store.Get(c.Request, "session")
	var userAuth models.UserAuth
	if err := config.DB.First(&userAuth, "email = ?", otpInput.Email).Error; err != nil {
		fullName, _ := session.Values["full_name"].(string)
		hashedPassword, _ := session.Values["password"].(string)

		referralCode := helper.GenerateReferralCode()

		userAuth = models.UserAuth{
			FullName:     fullName,
			Email:        otpInput.Email,
			Password:     hashedPassword,
			ProfilePic:   os.Getenv("DEFAULT_PROFILE_PIC"),
			ReferralCode: strings.ToUpper(referralCode),
		}

		if err := config.DB.Create(&userAuth).Error; err != nil {
			logger.Log.Error("Failed to create new user",
				zap.String("email", otpInput.Email),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create user", "Something Went Wrong,Please Try Again ", "")
			return
		}
		logger.Log.Info("New user created successfully",
			zap.String("email", userAuth.Email),
			zap.String("fullName", userAuth.FullName))
	} else {
		hashedPassword, _ := session.Values["password"].(string)
		userAuth.Password = hashedPassword

		if err := config.DB.Model(&userAuth).
			Where("id = ?", userAuth.ID).
			Updates(map[string]interface{}{"password": hashedPassword}).Error; err != nil {
			logger.Log.Error("Failed to update user password",
				zap.String("email", userAuth.Email),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to change password", "Something Went Wrong , Please Try Again ", "")
			return
		}
		logger.Log.Info("User password updated successfully",
			zap.String("email", userAuth.Email))
	}

	session.Options.MaxAge = -1
	if err := session.Save(c.Request, c.Writer); err != nil {
		logger.Log.Warn("Failed to clear session",
			zap.String("email", otpInput.Email),
			zap.Error(err))
	}

	logger.Log.Info("OTP verified successfully",
		zap.String("email", otpInput.Email))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "OTP verified",
		"code":    http.StatusOK,
	})
}

func ResendOTP(c *gin.Context) {
	logger.Log.Info("Requested to resend OTP")

	session, _ := Store.Get(c.Request, "session")
	email, exists := session.Values["email"].(string)
	if !exists {
		logger.Log.Warn("No email in session for OTP resend")
		c.Redirect(http.StatusUnauthorized, "/auth/signup")
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
				logger.Log.Error("Failed to create new OTP record",
					zap.String("email", email),
					zap.Error(err))
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create OTP record", "Something Went Wrong , Please Try Again ", "")
				return
			}
		} else {
			logger.Log.Error("Failed to retrieve existing OTP record",
				zap.String("email", email),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to retrieve OTP record", "Something Went Wrong , Please Try Again ", "")
			return
		}
	} else {
		otpRecord.OTP = otp
		otpRecord.ExpireTime = expiry
		if err := config.DB.Save(&otpRecord).Error; err != nil {
			logger.Log.Error("Failed to update OTP record",
				zap.String("email", email),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update OTP record", "Something Went Wrong , Please Try Again", "")
			return
		}
	}

	if err := utils.SendOTPToEmail(email, otp); err != nil {
		logger.Log.Error("Failed to resend OTP email",
			zap.String("email", email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to send OTP", "Something Went Wrong , Please Try Again", "")
		return
	}

	logger.Log.Info("OTP resent successfully",
		zap.String("email", email),
		zap.String("otp", otp))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Status OK",
		"message": "OTP resend successfully",
		"code":    http.StatusOK,
	})
}
