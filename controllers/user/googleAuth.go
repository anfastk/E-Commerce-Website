package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GoogleUserInfo struct {
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func InitiateGoogleAuth(c *gin.Context) {
	state := uuid.New().String()

	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	state, err := c.Cookie("oauth_state")
	if err != nil || c.Query("state") != state {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=Invalid+OAuth+state")
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	code := c.Query("code")
	token, err := config.GoogleOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=Failed+to+exchange+token")
		return
	}

	client := config.GoogleOAuthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/user/login?error=Failed+to+get+user+info")
		return
	}
	defer resp.Body.Close()

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/user/login?error=Failed+to+read+user+info")
		return
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(userData, &googleUser); err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/user/login?error=Failed+to+parse+user+info")
		return
	}

	var user models.UserAuth
	result := config.DB.Where("email = ?", googleUser.Email).First(&user)

	if result.Error != nil {
		user = models.UserAuth{
			FullName:   googleUser.Name,
			Email:      googleUser.Email,
			Password:   "",
			GoogleID:   googleUser.Email,
			ProfilePic: googleUser.Picture,
			IsVerified: googleUser.VerifiedEmail,
			Status:     "Active", // Default status for new users
		}

		if err := config.DB.Create(&user).Error; err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/user/login?error=Failed+to+create+user")
			return
		}
	}

	if user.Status == "Blocked" {
		c.Redirect(http.StatusTemporaryRedirect, "/user/login?error=Your+account+is+blocked")
		return
	}

	jwtToken, err := middleware.GenerateJWT(user.ID, user.Email, RoleUser)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/user/login?error=Failed+to+generate+JWT+token")
		return
	}

	c.SetCookie("jwtTokensUser", jwtToken, int((time.Hour * 1).Seconds()), "/", "", false, true)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}
