package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var RoleUser = "User"

func ShowSignup(c *gin.Context) {
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
	c.HTML(http.StatusSeeOther, "signup.html", nil)
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SignUp(c *gin.Context) {
	//	var user models.UserAuth
	var userInput struct {
		FullName        string `form:"full_name" `
		Email           string `form:"email" `
		Password        string `form:"password" `
		ConfirmPassword string `form:"confirm_password" `
	}

	if err := c.ShouldBind(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  err.Error(),
			"code":   400,
		})
		return
	}

	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userInput.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status": "Conflict",
			"error":  "Email address already exists",
			"code":   409,
		})
		return
	}

	if userInput.Password != userInput.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Passwords do not match",
			"code":   400,
		})
		return
	}
	hashedPassword, err := HashPassword(userInput.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to process password",
			"code":   500,
		})
		return
	}

	session, err := Store.Get(c.Request, "session")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to create session",
			"code":   500,
		})
		return
	}
	session.Values["full_name"] = userInput.FullName
	session.Values["email"] = userInput.Email
	session.Values["password"] = hashedPassword
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusSeeOther, "/user/signup/otp")

}

func ShowOtpVerifyPage(c *gin.Context) {
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
	c.HTML(http.StatusOK, "OTPEmailVerification.html", gin.H{
		"email":   email,
		"message": "OTP sent successfully",
	})
}

func ShowLogin(c *gin.Context) {

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
	c.HTML(http.StatusSeeOther, "login.html", nil)
}

type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UserLoginHandler(c *gin.Context) {
	var user models.UserAuth
	var input UserInput

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

	if err := config.DB.Unscoped().Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "Unauthorized",
				"error":  "Invalid Email",
				"code":   401,
			})
			return
		}
	}
	if user.IsDeleted{
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "Unauthorized",
			"error":  "Your Account Is Deleted",
			"code":   401,
		})
		return
	}
	if !CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid password",
			"code":   400,
		})
		return
	}

	if user.Status == "Blocked" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "Unauthorized",
			"error":  "Your Account Is Blocked",
			"code":   403,
		})
		return
	}

	token, err := middleware.GenerateJWT(user.ID, user.Email, RoleUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to generate JWT tokens",
			"code":   "500",
		})
		return
	}
	c.SetCookie("jwtTokensUser", token, int((time.Hour * 1).Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
		"code":    200,
	})
}

func UserLogoutHandler(c *gin.Context) {

	c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "User logged out successfully",
		"code":    http.StatusOK,
	})
}
