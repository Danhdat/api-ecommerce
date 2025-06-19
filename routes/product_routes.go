package routes

import (
	"storelite/controllers"
	"storelite/middleware"
	"storelite/models"

	"github.com/gin-gonic/gin"
)

func SetupProductRoutes(router *gin.RouterGroup) {
	categoryController := controllers.NewCategoryController()
	productController := controllers.NewProductController()
	reviewController := controllers.NewReviewController()
	authMiddleware := middleware.NewAuthMiddleware()

	// Public Category routes
	categories := router.Group("/categories")
	{
		categories.GET("", categoryController.GetCategories)
		categories.GET("/:id", categoryController.GetCategoryByID)
		categories.GET("/slug/:slug", categoryController.GetCategoryBySlug)
	}

	// Admin Category routes
	adminCategories := router.Group("/admin/categories")
	adminCategories.Use(authMiddleware.RequireAuth())
	adminCategories.Use(authMiddleware.CSRFProtection())
	adminCategories.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		adminCategories.POST("", categoryController.CreateCategory)
		adminCategories.PUT("/:id", categoryController.UpdateCategory)
		adminCategories.DELETE("/:id", categoryController.DeleteCategory)
	}

	// Public Product routes
	products := router.Group("/products")
	{
		products.GET("", productController.GetProducts)
		products.GET("/featured", productController.GetFeaturedProducts)
		products.GET("/search", productController.SearchProducts)
		products.GET("/:id", productController.GetProductByID)
		products.GET("/slug/:slug", productController.GetProductBySlug)
	}

	// Admin Product routes
	adminProducts := router.Group("/admin/products")
	adminProducts.Use(authMiddleware.RequireAuth())
	adminProducts.Use(authMiddleware.CSRFProtection())
	adminProducts.Use(authMiddleware.RequireRole(models.RoleAdmin))
	{
		adminProducts.POST("", productController.CreateProduct)
		adminProducts.PUT("/:id", productController.UpdateProduct)
		adminProducts.DELETE("/:id", productController.DeleteProduct)
	}

	// Public Review routes
	reviews := router.Group("/reviews")
	{
		reviews.GET("/product/:product_id", reviewController.GetProductReviews)
	}

	// Authenticated Review routes
	authReviews := router.Group("/reviews")
	authReviews.Use(authMiddleware.RequireAuth())
	authReviews.Use(authMiddleware.CSRFProtection())
	{
		authReviews.POST("", reviewController.CreateReview)
		authReviews.PUT("/:id", reviewController.UpdateReview)
		authReviews.DELETE("/:id", reviewController.DeleteReview)
		authReviews.GET("/my-reviews", reviewController.GetUserReviews)
	}
}
