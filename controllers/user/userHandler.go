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
	var userInput struct {
		FullName        string `form:"full_name"`
		Email           string `form:"email"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm_password"`
	}

	if err := c.ShouldBind(&userInput); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userInput.Email).First(&existingUser).Error; err == nil {
		helper.RespondWithError(c, http.StatusConflict, "Email address already exists")
		return
	}

	if userInput.Password != userInput.ConfirmPassword {
		helper.RespondWithError(c, http.StatusBadRequest, "Passwords do not match")
		return
	}

	hashedPassword, err := HashPassword(userInput.Password)
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to process password")
		return
	}

	session, err := Store.Get(c.Request, "session")
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create session")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Session expired")
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
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data")
		return
	}

	if input.Email == "" || input.Password == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Email and Password are required")
		return
	}

	if err := config.DB.Unscoped().Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helper.RespondWithError(c, http.StatusUnauthorized, "Invalid Email")
			return
		}
	}
	if user.IsDeleted {
		helper.RespondWithError(c, http.StatusUnauthorized, "Your Account Is Deleted")
		return
	}
	if !CheckPasswordHash(input.Password, user.Password) {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid password")
		return
	}

	if user.Status == "Blocked" {
		helper.RespondWithError(c, http.StatusUnauthorized, "Your Account Is Blocked")
		return
	}

	token, err := middleware.GenerateJWT(user.ID, user.Email, RoleUser)
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to generate JWT tokens")
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

func ForgotPasswordEmail(c *gin.Context) {
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
	c.HTML(http.StatusSeeOther, "forgotPasswordEmail.html", nil)
}

func ForgotUserEmail(c *gin.Context) {
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
	userEmail := c.PostForm("email")
	if userEmail == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Enter email")
		return
	}
	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userEmail).First(&existingUser).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "User not found")
		return
	}
	c.HTML(http.StatusOK, "resetPassword.html", gin.H{
		"status": "OK",
		"Email":  userEmail,
		"code":   http.StatusOK,
	})
}

func PasswordReset(c *gin.Context) {
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
	userEmail := c.PostForm("email")
	password := c.PostForm("password")
	conformPassword := c.PostForm("conform_password")
	if password == "" || conformPassword == "" {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	if password != conformPassword {
		helper.RespondWithError(c, http.StatusBadRequest, "Password not match")
		return
	}

	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userEmail).First(&existingUser).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "User not found")
		return
	}

	session, err := Store.Get(c.Request, "session")
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create session")
		return
	}
	hashedPassword, err := HashPassword(conformPassword)
	session.Values["email"] = userEmail
	session.Values["password"] = hashedPassword
	session.Save(c.Request, c.Writer)
	c.Redirect(http.StatusSeeOther, "/user/signup/otp")
}