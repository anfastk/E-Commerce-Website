package routes

import (
	controllers "github.com/anfastk/E-Commerce-Website/controllers/admin"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/gin-gonic/gin"
)

var RoleAdmin = "Admin"

func AdminRoutes(r *gin.Engine) {
	admin := r.Group("/admin")
	{
		admin.GET("/login", controllers.ShowLoginPage)
		admin.POST("/login", controllers.AdminLoginHandler)
		admin.GET("/settings", controllers.ShowSettings)
		admin.POST("/logout", controllers.AdminLogoutHandler)
	}
	// Admin Product Managemant
	product := r.Group("/admin/products")
	product.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		product.GET("/", controllers.ShowProductsAdmin)
		product.GET("/main/add", controllers.ShowAddMainProduct)
		product.POST("main/add", controllers.AddMainProductDetails)
		product.GET("/main/details", controllers.ShowMainProductDetails)
		product.POST("/main/submit-description", controllers.AddProductDescription)
		product.POST("/main/submit-offer", controllers.AddProductOffer)
		product.POST("/main/edit/offer", controllers.UpdateProductOffer)
		product.POST("/main/delete/offer", controllers.DeleteProductOffer)
		product.GET("/variants/add/:id", controllers.ShowProductVariant)
		product.POST("/variants/add", controllers.AddProductVariants)
		product.GET("/variant/detail", controllers.ShowSingleProductVariantDetail)
		product.POST("/variant/submit-specification", controllers.AddProductSpecification)
		product.POST("/variant/delete/:id", controllers.DeleteProductVariant)
		product.POST("/variant/image/change/", controllers.ReplaceVariantProductImage)
		product.GET("/variant/detail/update/:id", controllers.ShowEditProductVariant)
		product.PATCH("/variant/detail/update/:id", controllers.EditProductVariant)
		product.GET("/main/details/update/:id", controllers.ShowEditMainProduct)
		product.POST("/main/image/change", controllers.ReplaceMainProductImage)
		product.PATCH("/main/details/update/:id", controllers.EditMainProduct)
		product.POST("/main/details/delete/:id", controllers.DeleteMainProduct)
		product.DELETE("/variant/specification/delete/:id", controllers.DeleteSpecification)
		product.DELETE("/variant/description/delete/:id", controllers.DeleteDescription)
		product.PATCH("/variant/update/specification/:id", controllers.UpdateProductSpecification)
	}
	// Admin User Managemant
	adminUser := r.Group("/admin/users")
	adminUser.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		adminUser.GET("/", controllers.ListUsers)
		adminUser.POST("/:id/block", controllers.BlockUser)
		adminUser.POST("/:id/delete", controllers.DeleteUser)
	}
	// Admin Category Managemant
	category := r.Group("/admin/category")
	category.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		category.GET("/", controllers.ListCategory)
		category.PATCH("/:id/edit", controllers.EditCategory)
		category.POST("/add", controllers.AddCategory)
		category.POST("/:id/delete", controllers.DeleteCategory)
		category.GET("/details/:id", controllers.ShowCategoryDetails)
		category.POST("/add/offer", controllers.AddCategoryOffer)
		category.PATCH("/offer/edit", controllers.UpdateCategoryOffer)
		category.POST("/delete/offer", controllers.DeleteCategoryOffer)
	}
	OrderList := r.Group("/admin/orderlist")
	OrderList.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		OrderList.GET("/", controllers.ShowOrderManagent)
		OrderList.GET("/details/:id", controllers.ShowOrderDetailManagement)
		OrderList.PATCH("/details/status/update", controllers.ChangeOrderStatus)
		OrderList.POST("/details/return/request", controllers.ApproveReturn)
	}

	coupon := r.Group("/admin/coupon")
	coupon.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		coupon.GET("/", controllers.ShowCoupon)
		coupon.POST("/add", controllers.AddCoupon)
		coupon.POST("/delete/:id", controllers.DeleteCoupon)
		coupon.GET("/details/:id", controllers.CouponDetails)
		coupon.POST("/details/edit/:id", controllers.UpdateCoupon)
	}

	sales := r.Group("/sales")
	sales.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		sales.GET("/", controllers.GetSalesDashboard)
		sales.GET("/filter", controllers.GetSalesData)
		sales.GET("/recent-orders", controllers.GetRecentOrdersUnfiltered) // New endpoint
		sales.GET("/download/report", controllers.DownloadSalesReport)
	}

	wallet := r.Group("/admin/wallet")
	wallet.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		wallet.GET("/management", controllers.ShowWalletManagement)
		wallet.GET("/management/details/:id", controllers.ShowTransactionDetails)
	}

	adminDashboard := r.Group("/admin/dashboard")
	adminDashboard.Use(middleware.AuthMiddleware(RoleAdmin))
	{
		adminDashboard.GET("/", controllers.DashboardHandler)
		adminDashboard.GET("/stats", controllers.StatsHandler)
		adminDashboard.GET("/orders", controllers.OrdersHandler)
		adminDashboard.GET("/charts", controllers.ChartsHandler)
	}
}
