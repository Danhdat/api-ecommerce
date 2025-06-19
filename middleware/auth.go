package middleware

import (
	"net/http"
	"storelite/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := utils.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("csrf_token", claims.CSRFToken)

		c.Next()
	}
}

func (am *AuthMiddleware) RequireRole(allowedRoles ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User role not found in context",
			})
			c.Abort()
			return
		}

		role, ok := userRole.(int)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid user role",
			})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Insufficient permissions",
		})
		c.Abort()
	}
}

func (am *AuthMiddleware) CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for GET requests
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// Skip CSRF for public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "CSRF token required",
			})
			c.Abort()
			return
		}

		// Get CSRF token from JWT claims
		contextCSRF, exists := c.Get("csrf_token")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "CSRF token not found in context",
			})
			c.Abort()
			return
		}

		if csrfToken != contextCSRF {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Invalid CSRF token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/recovery",
		"/api/v1/auth/recovery/verify",
		"/health",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// Helper function to get user ID from context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// Helper function to get user email from context
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", false
	}

	userEmail, ok := email.(string)
	return userEmail, ok
}

// Helper function to get user role from context
func GetUserRoleFromContext(c *gin.Context) (int, bool) {
	role, exists := c.Get("user_role")
	if !exists {
		return 0, false
	}

	userRole, ok := role.(int)
	return userRole, ok
}
