package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GoogleUserInfo struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func InitiateGoogleAuth(c *gin.Context) {
	logger.Log.Info("Initiating Google OAuth")

	state := uuid.New().String()
	logger.Log.Debug("Generated OAuth state", zap.String("state", state))

	c.SetCookie("oauth_state", state, 600, "/", "", false, true)
	url := config.GoogleOAuthConfig.AuthCodeURL(state)

	logger.Log.Info("Redirecting to Google OAuth URL", zap.String("url", url))
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	logger.Log.Info("Handling Google OAuth callback")

	state, err := c.Cookie("oauth_state")
	if err != nil || c.Query("state") != state {
		logger.Log.Warn("Invalid OAuth state",
			zap.Error(err),
			zap.String("cookieState", state),
			zap.String("queryState", c.Query("state")))
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=Invalid+OAuth+state")
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", false, true)
	logger.Log.Debug("Cleared OAuth state cookie")

	code := c.Query("code")
	token, err := config.GoogleOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		logger.Log.Error("Failed to exchange OAuth token", zap.String("code", code), zap.Error(err))
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=Failed+to+exchange+token")
		return
	}

	client := config.GoogleOAuthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Log.Error("Failed to get user info from Google", zap.Error(err))
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login?error=Failed+to+get+user+info")
		return
	}
	defer resp.Body.Close()

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("Failed to read user info response", zap.Error(err))
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login?error=Failed+to+read+user+info")
		return
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(userData, &googleUser); err != nil {
		logger.Log.Error("Failed to parse Google user info",
			zap.String("userData", string(userData)),
			zap.Error(err))
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login?error=Failed+to+parse+user+info")
		return
	}

	logger.Log.Debug("Retrieved Google user info",
		zap.String("email", googleUser.Email),
		zap.String("name", googleUser.Name))

	var user models.UserAuth
	result := config.DB.Unscoped().Where("email = ?", googleUser.Email).First(&user)

	if result.Error != nil {
		cloudinaryURL, uploadErr := utils.UploadImageToCloudinary(nil, nil, config.InitCloudinary(), "ProfilePicture", googleUser.Picture)
		if uploadErr != nil {
			logger.Log.Warn("Failed to upload profile picture to Cloudinary",
				zap.String("email", googleUser.Email),
				zap.Error(uploadErr))
			cloudinaryURL = "" // Fallback to empty string if upload fails
		}

		referralCode := helper.GenerateReferralCode()
		user = models.UserAuth{
			FullName:     strings.ToUpper(googleUser.Name),
			Email:        googleUser.Email,
			Password:     "",
			GoogleID:     googleUser.Email,
			ProfilePic:   cloudinaryURL,
			IsVerified:   googleUser.VerifiedEmail,
			Status:       "Active",
			ReferralCode: strings.ToUpper(referralCode),
		}

		if err := config.DB.Create(&user).Error; err != nil {
			logger.Log.Error("Failed to create new user",
				zap.String("email", googleUser.Email),
				zap.Error(err))
			c.Redirect(http.StatusTemporaryRedirect, "/auth/login?error=Failed+to+create+user")
			return
		}
		logger.Log.Info("Created new user from Google OAuth",
			zap.String("email", user.Email),
			zap.Uint("userID", user.ID))
	}

	if user.IsBlocked || user.IsDeleted {
		logger.Log.Warn("User account blocked or deleted",
			zap.String("email", user.Email),
			zap.Bool("isBlocked", user.IsBlocked),
			zap.Bool("isDeleted", user.IsDeleted))
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login?error=Your+Account+Is+Blocked+Or+Deleted")
		return
	}

	jwtToken, err := middleware.GenerateJWT(user.ID, user.Email, RoleUser)
	if err != nil {
		logger.Log.Error("Failed to generate JWT token",
			zap.String("email", user.Email),
			zap.Error(err))
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login?error=Failed+to+generate+JWT+token")
		return
	}

	c.SetCookie("jwtTokensUser", jwtToken, int((time.Hour * 1).Seconds()), "/", "", false, true)
	logger.Log.Info("Google OAuth login successful",
		zap.String("email", user.Email),
		zap.Uint("userID", user.ID))
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
