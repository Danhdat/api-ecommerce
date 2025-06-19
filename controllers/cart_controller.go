package controllers

import (
	"net/http"
	"storelite/middleware"
	"storelite/models"
	"storelite/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CartController struct {
	validator   *validator.Validate
	cartService *services.CartService
}

func NewCartController() *CartController {
	return &CartController{
		validator:   validator.New(),
		cartService: services.NewCartService(),
	}
}

// getCartIdentifier lấy thông tin để identify cart (user ID hoặc session ID)
func (cc *CartController) getCartIdentifier(c *gin.Context) (*uint, string) {
	// Try to get user ID from JWT (if logged in)
	userID, exists := middleware.GetUserIDFromContext(c)
	if exists {
		return &userID, ""
	}

	// Get or create session ID for guest user
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = cc.cartService.GenerateSessionID()
		c.Header("X-Session-ID", sessionID)
	}

	return nil, sessionID
}

// GetCart lấy thông tin giỏ hàng
func (cc *CartController) GetCart(c *gin.Context) {
	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Cart retrieved successfully",
		Data:    cart.ToResponse(),
	})
}

// AddToCart thêm sản phẩm vào giỏ hàng
func (cc *CartController) AddToCart(c *gin.Context) {
	var req models.AddToCartRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.cartService.AddToCart(cart, req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to add item to cart",
			Error:   err.Error(),
		})
		return
	}

	// Reload cart with updated items
	cart, _ = cc.cartService.GetOrCreateCart(userID, sessionID)

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Item added to cart successfully",
		Data:    cart.ToResponse(),
	})
}

// UpdateCartItem cập nhật số lượng item trong cart
func (cc *CartController) UpdateCartItem(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid item ID",
		})
		return
	}

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.cartService.UpdateCartItem(cart.ID, uint(itemID), req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to update cart item",
			Error:   err.Error(),
		})
		return
	}

	// Reload cart with updated items
	cart, _ = cc.cartService.GetOrCreateCart(userID, sessionID)

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Cart item updated successfully",
		Data:    cart.ToResponse(),
	})
}

// RemoveFromCart xóa item khỏi cart
func (cc *CartController) RemoveFromCart(c *gin.Context) {
	itemID, err := strconv.ParseUint(c.Param("item_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid item ID",
		})
		return
	}

	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.cartService.RemoveFromCart(cart.ID, uint(itemID)); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to remove item from cart",
			Error:   err.Error(),
		})
		return
	}

	// Reload cart with updated items
	cart, _ = cc.cartService.GetOrCreateCart(userID, sessionID)

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Item removed from cart successfully",
		Data:    cart.ToResponse(),
	})
}

// ClearCart xóa tất cả items trong cart
func (cc *CartController) ClearCart(c *gin.Context) {
	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.cartService.ClearCart(cart.ID); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to clear cart",
			Error:   err.Error(),
		})
		return
	}

	// Reload cart
	cart, _ = cc.cartService.GetOrCreateCart(userID, sessionID)

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Cart cleared successfully",
		Data:    cart.ToResponse(),
	})
}

// ValidateCart kiểm tra tính hợp lệ của cart
func (cc *CartController) ValidateCart(c *gin.Context) {
	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	validation := cc.cartService.ValidateCart(cart)

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	if validation.IsValid {
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Cart is valid for checkout",
			Data: gin.H{
				"is_valid": true,
				"cart":     cart.ToResponse(),
			},
		})
	} else {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Cart validation failed",
			Data: gin.H{
				"is_valid": false,
				"issues":   validation.Issues,
				"cart":     cart.ToResponse(),
			},
		})
	}
}

// GetCartCount lấy số lượng items trong cart
func (cc *CartController) GetCartCount(c *gin.Context) {
	userID, sessionID := cc.getCartIdentifier(c)

	cart, err := cc.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	cartResponse := cart.ToResponse()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Cart count retrieved successfully",
		Data: gin.H{
			"item_count":     cartResponse.ItemCount,
			"total_quantity": cartResponse.TotalQuantity,
			"total":          cartResponse.Total,
		},
	})
}

// MergeGuestCart merge cart của guest khi user đăng nhập
func (cc *CartController) MergeGuestCart(c *gin.Context) {
	// This endpoint should be called after user login
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Session ID required",
		})
		return
	}

	if err := cc.cartService.MergeGuestCartToUser(sessionID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to merge guest cart",
			Error:   err.Error(),
		})
		return
	}

	// Get merged cart
	cart, _ := cc.cartService.GetOrCreateCart(&userID, "")

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Guest cart merged successfully",
		Data:    cart.ToResponse(),
	})
}
