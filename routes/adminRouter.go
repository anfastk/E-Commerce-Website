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
		product.POST("/variant/delete/:id",controllers.DeleteProductVariant)
		product.POST("/variant/image/delete/:id",controllers.DeleteVariantImage)
		product.GET("/variant/detail/update/:id",controllers.ShowEditProductVariant)
		product.PATCH("/variant/detail/update/:id",controllers.EditProductVariant)
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
}
