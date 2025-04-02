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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var RoleUser = "User"

func ShowSignup(c *gin.Context) {
	logger.Log.Info("Showing signup page")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already logged in, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		logger.Log.Warn("Invalid token found", zap.Error(err))
	}

	logger.Log.Info("Signup page loaded")
	c.HTML(http.StatusSeeOther, "signup.html", nil)
}

func HashPassword(password string) (string, error) {
	logger.Log.Debug("Hashing password")
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Log.Error("Failed to hash password", zap.Error(err))
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	logger.Log.Debug("Checking password hash")
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		logger.Log.Warn("Password hash check failed", zap.Error(err))
		return false
	}
	return true
}

func SignUp(c *gin.Context) {
	logger.Log.Info("Processing signup")

	var userInput struct {
		FullName        string `form:"full_name"`
		Email           string `form:"email"`
		Password        string `form:"password"`
		ConfirmPassword string `form:"confirm_password"`
	}

	if err := c.ShouldBind(&userInput); err != nil {
		logger.Log.Error("Failed to bind signup data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Binding the data", "")
		return
	}

	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userInput.Email).First(&existingUser).Error; err == nil {
		logger.Log.Warn("Email already exists",
			zap.String("email", userInput.Email))
		helper.RespondWithError(c, http.StatusConflict, "Email address already exists", "Email address already exists", "")
		return
	}

	if userInput.Password != userInput.ConfirmPassword {
		logger.Log.Warn("Passwords do not match",
			zap.String("email", userInput.Email))
		helper.RespondWithError(c, http.StatusBadRequest, "Passwords do not match", "Passwords do not match", "")
		return
	}

	hashedPassword, err := HashPassword(userInput.Password)
	if err != nil {
		logger.Log.Error("Failed to hash password",
			zap.String("email", userInput.Email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to process password", "Failed to process password", "")
		return
	}

	session, err := Store.Get(c.Request, "session")
	if err != nil {
		logger.Log.Error("Failed to create session",
			zap.String("email", userInput.Email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create session", "Failed to create session", "")
		return
	}

	session.Values["full_name"] = userInput.FullName
	session.Values["email"] = userInput.Email
	session.Values["password"] = hashedPassword
	if err := session.Save(c.Request, c.Writer); err != nil {
		logger.Log.Error("Failed to save session",
			zap.String("email", userInput.Email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save session", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Signup successful, redirecting to OTP verification",
		zap.String("email", userInput.Email))
	c.Redirect(http.StatusSeeOther, "/auth/signup/otp")
}

func ShowOtpVerifyPage(c *gin.Context) {
	logger.Log.Info("Showing OTP verification page")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already logged in, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		logger.Log.Warn("Invalid token found", zap.Error(err))
	}

	session, _ := Store.Get(c.Request, "session")
	email, exists := session.Values["email"].(string)
	if !exists {
		logger.Log.Error("Session expired or email not found")
		helper.RespondWithError(c, http.StatusBadRequest, "Session expired", "Session expired", "")
		return
	}

	logger.Log.Info("OTP verification page loaded",
		zap.String("email", email))
	c.HTML(http.StatusOK, "OTPEmailVerification.html", gin.H{
		"email":   email,
		"message": "OTP sent successfully",
	})
}

func ShowLogin(c *gin.Context) {
	logger.Log.Info("Showing login page")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already logged in, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		logger.Log.Warn("Invalid token found", zap.Error(err))
	}

	logger.Log.Info("Login page loaded")
	c.HTML(http.StatusSeeOther, "login.html", nil)
}

type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func UserLoginHandler(c *gin.Context) {
	logger.Log.Info("Processing user login")

	var user models.UserAuth
	var input UserInput

	if err := c.ShouldBind(&input); err != nil {
		logger.Log.Error("Failed to bind login data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Binding the data", "Binding the data", "")
		return
	}

	if input.Email == "" || input.Password == "" {
		logger.Log.Warn("Missing email or password")
		helper.RespondWithError(c, http.StatusBadRequest, "Email and Password are required", "Email and Password are required", "")
		return
	}

	if err := config.DB.Unscoped().Where("email = ?", input.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Warn("Invalid email",
				zap.String("email", input.Email))
			helper.RespondWithError(c, http.StatusUnauthorized, "Invalid Email", "Invalid Email", "")
			return
		}
		logger.Log.Error("Database error during login",
			zap.String("email", input.Email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Database error", "Something Went Wrong", "")
		return
	}

	if user.IsDeleted {
		logger.Log.Warn("Attempt to login with deleted account",
			zap.String("email", input.Email))
		helper.RespondWithError(c, http.StatusUnauthorized, "Your Account Is Deleted", "Your Account Is Deleted", "")
		return
	}

	if !CheckPasswordHash(input.Password, user.Password) {
		logger.Log.Warn("Invalid password",
			zap.String("email", input.Email))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid password", "Invalid password", "")
		return
	}

	if user.Status == "Blocked" {
		logger.Log.Warn("Attempt to login with blocked account",
			zap.String("email", input.Email))
		helper.RespondWithError(c, http.StatusUnauthorized, "Your Account Is Blocked", "Your Account Is Blocked", "")
		return
	}

	token, err := middleware.GenerateJWT(user.ID, user.Email, RoleUser)
	if err != nil {
		logger.Log.Error("Failed to generate JWT",
			zap.String("email", input.Email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to generate JWT tokens", "Failed to generate JWT tokens", "")
		return
	}

	c.SetCookie("jwtTokensUser", token, int((time.Hour * 1).Seconds()), "/", "", false, true)
	logger.Log.Info("User login successful",
		zap.String("email", input.Email),
		zap.Uint("userID", user.ID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login successful",
		"token":   token,
		"code":    200,
	})
}

func UserLogoutHandler(c *gin.Context) {
	logger.Log.Info("Processing user logout")

	tokenString, _ := c.Cookie("jwtTokensUser")
	if tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})
		if err == nil && token.Valid {
			logger.Log.Info("User logged out",
				zap.String("email", claims.Email))
		}
	}

	c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
	logger.Log.Info("Logout completed")
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "User logged out successfully",
		"code":    http.StatusOK,
	})
}

func ForgotPasswordEmail(c *gin.Context) {
	logger.Log.Info("Showing forgot password email page")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already logged in, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		logger.Log.Warn("Invalid token found", zap.Error(err))
	}

	logger.Log.Info("Forgot password email page loaded")
	c.HTML(http.StatusSeeOther, "forgotPasswordEmail.html", nil)
}

func ForgotUserEmail(c *gin.Context) {
	logger.Log.Info("Processing forgot password email")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already logged in, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		logger.Log.Warn("Invalid token found", zap.Error(err))
	}

	var userInput struct {
		Email string `json:"email" form:"email"`
	}

	if err := c.ShouldBind(&userInput); err != nil {
		logger.Log.Error("Invalid request payload", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Invalid request", "")
		return
	}

	if userInput.Email == "" {
		logger.Log.Warn("Email not provided")
		helper.RespondWithError(c, http.StatusBadRequest, "Please enter your email address", "Please enter your email address", "")
		return
	}

	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userInput.Email).First(&existingUser).Error; err != nil {
		logger.Log.Warn("User not found for forgot password",
			zap.String("email", userInput.Email),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Email not found", "Email not found", "")
		return
	}

	logger.Log.Info("Forgot password email verified, showing reset page",
		zap.String("email", userInput.Email))
	c.HTML(http.StatusOK, "resetPassword.html", gin.H{
		"status": "OK",
		"Email":  userInput.Email,
		"code":   http.StatusOK,
	})
}

func PasswordReset(c *gin.Context) {
	logger.Log.Info("Processing password reset")

	tokenString, err := c.Cookie("jwtTokensUser")
	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.JwtSecretKey, nil
		})

		if err == nil && token.Valid {
			logger.Log.Info("User already logged in, redirecting to home",
				zap.String("email", claims.Email))
			c.Redirect(http.StatusSeeOther, "/")
			return
		}
		logger.Log.Warn("Invalid token found", zap.Error(err))
	}

	userEmail := c.PostForm("email")
	password := c.PostForm("password")
	conformPassword := c.PostForm("conform_password")

	if password == "" || conformPassword == "" {
		logger.Log.Warn("Missing password fields",
			zap.String("email", userEmail))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Invalid input data", "")
		return
	}

	if password != conformPassword {
		logger.Log.Warn("Passwords do not match",
			zap.String("email", userEmail))
		helper.RespondWithError(c, http.StatusBadRequest, "Password not match", "Password not match", "")
		return
	}

	var existingUser models.UserAuth
	if err := config.DB.Where("email = ?", userEmail).First(&existingUser).Error; err != nil {
		logger.Log.Error("User not found for password reset",
			zap.String("email", userEmail),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	session, err := Store.Get(c.Request, "session")
	if err != nil {
		logger.Log.Error("Failed to create session for password reset",
			zap.String("email", userEmail),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create session", "Failed to create session", "")
		return
	}

	hashedPassword, err := HashPassword(conformPassword)
	if err != nil {
		logger.Log.Error("Failed to hash new password",
			zap.String("email", userEmail),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to process password", "Something Went Wrong", "")
		return
	}

	session.Values["email"] = userEmail
	session.Values["password"] = hashedPassword
	if err := session.Save(c.Request, c.Writer); err != nil {
		logger.Log.Error("Failed to save session for password reset",
			zap.String("email", userEmail),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save session", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Password reset initiated, redirecting to OTP",
		zap.String("email", userEmail))
	c.Redirect(http.StatusSeeOther, "/auth/signup/otp")
}
