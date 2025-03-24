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
		auth.GET("/signup", controllers.ShowSignup)
		auth.POST("/signup", controllers.SignUp)
		auth.GET("/signup/otp", controllers.SendOtp)
		auth.GET("/signup/verifyotp", controllers.ShowOtpVerifyPage)
		auth.POST("/signup/verifyotp", controllers.VerifyOtp)
		auth.POST("/signup/otp/resend", controllers.ResendOTP)
		auth.GET("/login", controllers.ShowLogin)
		auth.POST("/login", controllers.UserLoginHandler)
		auth.GET("/forgot/password", controllers.ForgotPasswordEmail)
		auth.POST("/forgot/password", controllers.ForgotUserEmail)
		auth.POST("/reset/password", controllers.PasswordReset)
		auth.POST("/logout", middleware.AuthMiddleware(RoleUser), controllers.UserLogoutHandler)
	}
	r.GET("/", controllers.UserHome)
	r.GET("/products", controllers.ShowProducts)
	r.GET("/products/details/:id", controllers.ShowProductDetail)
	r.GET("/products/filter", controllers.FilterProducts)
	r.POST("/checkout/payment/verify", middleware.AuthMiddleware(RoleUser), controllers.VerifyRazorpayPayment)
	r.POST("/order/failed", middleware.AuthMiddleware(RoleUser), controllers.PaymentFailureHandler)
	r.GET("/contactUs", controllers.ShowContactUs)

	userProfile := r.Group("/profile")
	userProfile.Use(middleware.AuthMiddleware(RoleUser))
	{
		userProfile.GET("/", controllers.ProfileDetails)
		userProfile.PATCH("/", controllers.ProfileUpdate)
		userProfile.PATCH("/avathar/update", controllers.ProfileImageUpdate)
		userProfile.GET("/manage/address", controllers.ManageAddress)
		userProfile.GET("/add/address", controllers.ShowAddAddress)
		userProfile.POST("/add/address", controllers.AddAddress)
		userProfile.GET("/edit/address/:id", controllers.ShowEditAddress)
		userProfile.PATCH("/edit/address", controllers.EditAddress)
		userProfile.POST("/address/:id/default", controllers.SetAsDefaultAddress)
		userProfile.DELETE("/delete/address/:id", controllers.DeleteAddress)
		userProfile.GET("/settings", controllers.Settings)
		userProfile.GET("/change/password", controllers.ShowChangePassword)
		userProfile.POST("/change/password", controllers.ChangePassword)
		userProfile.GET("/order/details", controllers.OrderDetails)
		userProfile.GET("/order/details/track/:id", controllers.TrackingPage)
		userProfile.POST("/order/details/track/:id/cancel", controllers.CancelSpecificOrder)
		userProfile.POST("/order/details/track/:id/cancel/all", controllers.CancelAllOrderItems)
		userProfile.POST("/order/details/track/pay/now", controllers.PayNow)
		userProfile.POST("/order/details/track/item/return", controllers.ReturnOrder)
		userProfile.POST("/order/details/track/pay/now/verify", controllers.VerifyPayNowRazorpayPayment)
		userProfile.GET("/order/history", controllers.OrderHistory)
		userProfile.GET("/order/history/data", controllers.OrderHistoryData)
		userProfile.GET("/wallet", controllers.WalletHandler)
		userProfile.GET("/referral", controllers.ShowReferralPage)
		userProfile.POST("/referral/add", controllers.AddReferral)
		userProfile.POST("/wallet/add/amount", controllers.AddMoneyTOWalltet)
		userProfile.POST("/wallet/add/amount/verify", controllers.VerifyAddTOWalletRazorpayPayment)
		userProfile.POST("/wallet/send/gift/card", controllers.SendGiftCard)
		userProfile.GET("/order/details/track/invoices/:id", controllers.DownloadInvoice)
	}

	cart := r.Group("/cart")
	cart.Use(middleware.AuthMiddleware(RoleUser))
	{
		cart.GET("/", controllers.ShowCart)
		cart.POST("/total", controllers.ShowCartTotal)
		cart.POST("/add/:id", controllers.AddToCart)
		cart.POST("/update/quantity/:id", controllers.CartItemUpdate)
		cart.POST("/delete/:id", controllers.DeleteCartItems)

	}

	checkout := r.Group("/checkout")
	checkout.Use(middleware.AuthMiddleware(RoleUser))
	{
		checkout.POST("/", controllers.ShowCheckoutPage)
		checkout.POST("/addresses", controllers.ShippingAddress)
		checkout.POST("/payment", controllers.PaymentPage)
		checkout.POST("/payment/proceed", controllers.ProceedToPayment)
		checkout.POST("/check/coupon", controllers.CheckCoupon)
		checkout.GET("/check/wallet/balance", controllers.FetchWalletBalance)
		checkout.POST("/redeem/gift/code", controllers.RedeemGiftCard)
	}

	wishlist := r.Group("/wishlist")
	wishlist.Use(middleware.AuthMiddleware(RoleUser))
	{
		wishlist.GET("/", controllers.ShowWishlist)
		wishlist.POST("/add/:id", controllers.AddToWishlist)
		wishlist.POST("/products/move/cart/:id", controllers.WishlistTOCart)
		wishlist.POST("/all/products/move/cart", controllers.WishlistAllTOCart)
		wishlist.POST("/remove/:id", controllers.RemoveFromWishlist)
	}

	r.NoRoute(controllers.Handle404Error)
}
