package middleware

import (
	"storelite/services"

	"github.com/gin-gonic/gin"
)

type CartMiddleware struct {
	cartService *services.CartService
}

func NewCartMiddleware() *CartMiddleware {
	return &CartMiddleware{
		cartService: services.NewCartService(),
	}
}

// EnsureSessionID đảm bảo có session ID cho guest user
func (cm *CartMiddleware) EnsureSessionID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if user is authenticated
		if _, exists := GetUserIDFromContext(c); exists {
			c.Next()
			return
		}

		// Check if session ID exists
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			// Generate new session ID
			sessionID = cm.cartService.GenerateSessionID()
			c.Header("X-Session-ID", sessionID)
		}

		c.Next()
	}
}

// CleanExpiredCartsMiddleware middleware để dọn dẹp carts hết hạn
// Có thể chạy định kỳ hoặc trước một số operations
func (cm *CartMiddleware) CleanExpiredCartsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clean expired carts (can be run periodically)
		go cm.cartService.CleanExpiredCarts()
		c.Next()
	}
}
