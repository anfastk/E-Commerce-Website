package routes

import (
	controllers "github.com/anfastk/E-Commerce-Website/controllers/user"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/gin-gonic/gin"
)

var RoleUser = "User"

func UserRouter(r *gin.Engine) {

	auth := r.Group("/auth")
	{
		auth.GET("/google/login", controllers.InitiateGoogleAuth)
		auth.GET("/google/callback", controllers.HandleGoogleCallback)
	}
	r.GET("/products", middleware.NoCacheMiddleware(), controllers.ShowProducts)
	r.GET("/products/details/:id",middleware.NoCacheMiddleware(),controllers.ShowProductDetail)

	userSignup := r.Group("/user/signup")
	userSignup.Use(middleware.NoCacheMiddleware())
	{
		userSignup.GET("/", controllers.ShowSignup)
		userSignup.POST("/", controllers.SignUp)
		userSignup.GET("/otp", controllers.SendOtp)
		userSignup.POST("/verifyotp", controllers.VerifyOtp)
		userSignup.GET("/otp/resend", controllers.ResendOTP)
	}
	userLogin := r.Group("/user/login")
	userLogin.Use(middleware.NoCacheMiddleware())
	{
		userLogin.GET("/", controllers.ShowLogin)
	}
}
