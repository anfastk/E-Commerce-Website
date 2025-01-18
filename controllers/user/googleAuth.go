package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
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
	
	session, _ := Store.Get(c.Request, "session")
	session.Values["oauth_state"] = state
	session.Save(c.Request, c.Writer)

	url := config.GoogleOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	session, _ := Store.Get(c.Request, "session")
	expectedState, ok := session.Values["oauth_state"].(string)
	if !ok || c.Query("state") != expectedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
		return
	}

	code := c.Query("code")
	token, err := config.GoogleOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	client := config.GoogleOAuthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read user info"})
		return
	}

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(userData, &googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}

	var user models.UserAuth
	result := config.DB.Where("email = ?", googleUser.Email).First(&user)
	
	if result.Error != nil {
		user = models.UserAuth{
			FullName:      googleUser.Name,
			Email:         googleUser.Email,
			Password:      "", 
			GoogleID:      googleUser.Email,
			ProfilePic:  googleUser.Picture,
			IsVerified:    googleUser.VerifiedEmail,
		}
		
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	session.Values["user_id"] = user.ID
	session.Values["email"] = user.Email
	session.Values["is_logged_in"] = true
	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusTemporaryRedirect, "/")
}