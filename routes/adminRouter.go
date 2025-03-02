package routes

import (
	controllers "github.com/anfastk/E-Commerce-Website/controllers/admin"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/gin-gonic/gin"
)

var RoleAdmin = "Admin"

func AdminRoutes(r *gin.Engine) {
	admin := r.Group("/admin")
	admin.Use(middleware.NoCacheMiddleware())
	{
		admin.GET("/login", controllers.ShowLoginPage)
		admin.POST("/login", controllers.AdminLoginHandler)
		admin.GET("/settings",controllers.ShowSettings)
		admin.POST("/logout",controllers.AdminLogoutHandler)
	}
	// Admin Product Managemant
	product := r.Group("/admin/products")
	product.Use(middleware.AuthMiddleware(RoleAdmin))
	product.Use(middleware.NoCacheMiddleware())
	{
		product.GET("/", controllers.ShowProductsAdmin)
		product.GET("/main/add", controllers.ShowAddMainProduct)
		product.POST("main/add", controllers.AddMainProductDetails)
		product.GET("/main/details", controllers.ShowMainProductDetails)
		product.POST("/main/submit-description", controllers.AddProductDescription)
		product.POST("/main/submit-offer", controllers.AddProductOffer)
		product.GET("/variants/add/:id", controllers.ShowProductVariant)
		product.POST("/variants/add", controllers.AddProductVariants)
		product.GET("/variant/details", controllers.ShowMutiProductVariantDetails)
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
		product.PATCH("/variant/update/description/:id", controllers.UpdateProductDescription)
		product.PATCH("/variant/update/specification/:id", controllers.UpdateProductSpecification)

	}
	// Admin User Managemant
	adminUser := r.Group("/admin/users")
	adminUser.Use(middleware.AuthMiddleware(RoleAdmin))
	adminUser.Use(middleware.NoCacheMiddleware())
	{
		adminUser.GET("/", controllers.ListUsers)
		adminUser.POST("/:id/block", controllers.BlockUser)
		adminUser.POST("/:id/delete", controllers.DeleteUser)
	}
	// Admin Category Managemant
	category := r.Group("/admin/category")
	category.Use(middleware.AuthMiddleware(RoleAdmin))
	category.Use(middleware.NoCacheMiddleware())
	{
		category.GET("/", controllers.ListCategory)
		category.PATCH("/:id/edit", controllers.EditCategory)
		category.POST("/add", controllers.AddCategory)
		category.POST("/:id/delete", controllers.DeleteCategory)
	}
	OrderList:=r.Group("/admin/orderlist")
	OrderList.Use(middleware.AuthMiddleware(RoleAdmin))
	OrderList.Use(middleware.NoCacheMiddleware())
	{
		OrderList.GET("/", controllers.ShowOrderManagent)
		OrderList.GET("/details/:id", controllers.ShowOrderDetailManagement)
		OrderList.PATCH("/details/status/update",controllers.ChangeOrderStatus)
		OrderList.POST("/details/return/request",controllers.ApproveReturn)
	}
}
