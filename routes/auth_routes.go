package routes

import (
	"storelite/controllers"
	"storelite/middleware"
	"storelite/models"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup) {
	authController := controllers.NewAuthController()
	authMiddleware := middleware.NewAuthMiddleware()

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/recovery", authController.RequestRecovery)
		auth.POST("/recovery/verify", authController.VerifyRecovery)
	}

	// User routes
	users := router.Group("/users")
	users.Use(authMiddleware.RequireAuth())
	users.Use(authMiddleware.CSRFProtection())
	{
		users.GET("/", authMiddleware.RequireRole(models.RoleAdmin), authController.GetAllUsers)
		users.GET("/:id", authController.GetUserByID)
		users.GET("/profile", authController.GetProfile)
	}
}
