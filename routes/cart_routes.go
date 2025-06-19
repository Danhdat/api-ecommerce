package routes

import (
	"storelite/controllers"

	"github.com/gin-gonic/gin"
)

func SetupCartRoutes(router *gin.RouterGroup) {
	cartController := controllers.NewCartController()

	// Cart routes (no authentication required - support both guest and logged-in users)
	cart := router.Group("/cart")
	{
		cart.GET("", cartController.GetCart)
		cart.POST("/add", cartController.AddToCart)
		cart.PUT("/items/:item_id", cartController.UpdateCartItem)
		cart.DELETE("/items/:item_id", cartController.RemoveFromCart)
		cart.DELETE("/clear", cartController.ClearCart)
		cart.GET("/validate", cartController.ValidateCart)
		cart.GET("/count", cartController.GetCartCount)
		cart.POST("/merge", cartController.MergeGuestCart) // For merging guest cart when user logs in
	}
}
