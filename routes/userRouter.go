package routes

import (
	controllers "github.com/anfastk/E-Commerce-Website/controllers/user"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/gin-gonic/gin"
)

var RoleUser = "User"

func UserRouter(r *gin.Engine) {

	auth := r.Group("/auth")
	auth.Use(middleware.NoCacheMiddleware())
	{
		auth.GET("/google/login", controllers.InitiateGoogleAuth)
		auth.GET("/google/callback", controllers.HandleGoogleCallback)
	}
	r.GET("/", middleware.NoCacheMiddleware(), controllers.UserHome)
	r.GET("/products", middleware.NoCacheMiddleware(), controllers.ShowProducts)
	r.GET("/products/details/:id", middleware.NoCacheMiddleware(), controllers.ShowProductDetail)

	user := r.Group("/user")
	user.Use(middleware.NoCacheMiddleware())
	{
		user.GET("/signup", controllers.ShowSignup)
		user.POST("/signup", controllers.SignUp)
		user.GET("/signup/otp", controllers.SendOtp)
		user.GET("/signup/verifyotp", controllers.ShowOtpVerifyPage)
		user.POST("/signup/verifyotp", controllers.VerifyOtp)
		user.POST("/signup/otp/resend", controllers.ResendOTP)
		user.GET("/login", controllers.ShowLogin)
		user.POST("/login", controllers.UserLoginHandler)
		user.GET("/forgot/password", controllers.ForgotPasswordEmail)
		user.POST("/forgot/password", controllers.ForgotUserEmail)
		user.POST("/reset/password", controllers.PasswordReset)
		user.POST("/logout", middleware.AuthMiddleware(RoleUser), controllers.UserLogoutHandler)
	}

	userProfile := r.Group("/profile")
	userProfile.Use(middleware.NoCacheMiddleware())
	userProfile.Use(middleware.AuthMiddleware(RoleUser))
	{
		userProfile.GET("/", controllers.ProfileDetails)
		userProfile.PATCH("/", controllers.ProfileUpdate)
		userProfile.GET("/manage/address", controllers.ManageAddress)
		userProfile.GET("/add/address", controllers.ShowAddAddress)
		userProfile.POST("/add/address", controllers.AddAddress)
		userProfile.GET("/edit/address/:id", controllers.ShowEditAddress)
		userProfile.PATCH("/edit/address", controllers.EditAddress)
		userProfile.POST("/address/:id/default",controllers.SetAsDefaultAddress)
		userProfile.POST("/delete/address/:id", controllers.DeleteAddress)
		userProfile.GET("/settings", controllers.Settings)
		userProfile.GET("/change/password", controllers.ShowChangePassword)
		userProfile.POST("/change/password", controllers.ChangePassword)
	}

	cart := r.Group("/cart")
	cart.Use(middleware.NoCacheMiddleware())
	cart.Use(middleware.AuthMiddleware(RoleUser))
	{
		cart.GET("/", controllers.ShowCart)
		cart.GET("/total", controllers.ShowCartTotal)
		cart.POST("/add/:id", controllers.AddToCart)
		cart.POST("/update/quantity/:id", controllers.CartItemUpdate)
		cart.POST("/delete/:id", controllers.DeleteCartItems)
		cart.GET("/checkout", controllers.ShowCheckoutPage)
	}

}
