package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    *uint          `json:"user_id" gorm:"index"`             // NULL cho guest user
	SessionID string         `json:"session_id" gorm:"index;size:255"` // UUID cho guest user
	ExpiresAt time.Time      `json:"expires_at" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	CartItems []CartItem `json:"cart_items,omitempty" gorm:"foreignKey:CartID"`
}

type CartItem struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	CartID        uint           `json:"cart_id" gorm:"not null;index" validate:"required"`
	ProductID     uint           `json:"product_id" gorm:"not null;index" validate:"required"`
	ProductSizeID uint           `json:"product_size_id" gorm:"not null;index" validate:"required"`
	Quantity      int            `json:"quantity" gorm:"not null" validate:"required,min=1"`
	Price         float64        `json:"price" gorm:"not null;type:decimal(10,2)"` // Giá tại thời điểm thêm vào cart
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Cart        Cart        `json:"cart,omitempty" gorm:"foreignKey:CartID"`
	Product     Product     `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ProductSize ProductSize `json:"product_size,omitempty" gorm:"foreignKey:ProductSizeID"`
}

// Request/Response structs
type AddToCartRequest struct {
	ProductID     uint `json:"product_id" validate:"required"`
	ProductSizeID uint `json:"product_size_id" validate:"required"`
	Quantity      int  `json:"quantity" validate:"required,min=1,max=100"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=0,max=100"`
}

type CartItemResponse struct {
	ID            uint                `json:"id"`
	ProductID     uint                `json:"product_id"`
	ProductSizeID uint                `json:"product_size_id"`
	Product       ProductResponse     `json:"product"`
	ProductSize   ProductSizeResponse `json:"product_size"`
	Quantity      int                 `json:"quantity"`
	Price         float64             `json:"price"`
	CurrentPrice  float64             `json:"current_price"` // Giá hiện tại của sản phẩm
	Subtotal      float64             `json:"subtotal"`
	IsAvailable   bool                `json:"is_available"`
	StockStatus   string              `json:"stock_status"`
	Message       string              `json:"message,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type CartResponse struct {
	ID            uint               `json:"id"`
	UserID        *uint              `json:"user_id"`
	SessionID     string             `json:"session_id"`
	ItemCount     int                `json:"item_count"`
	TotalQuantity int                `json:"total_quantity"`
	Subtotal      float64            `json:"subtotal"`
	Discount      float64            `json:"discount"`
	Total         float64            `json:"total"`
	IsValid       bool               `json:"is_valid"`
	Issues        []string           `json:"issues,omitempty"`
	ExpiresAt     time.Time          `json:"expires_at"`
	CartItems     []CartItemResponse `json:"cart_items"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

type CartValidationResult struct {
	IsValid bool     `json:"is_valid"`
	Issues  []string `json:"issues"`
}

// Constants
const (
	CartExpiryHours      = 24 * 7 // 7 days
	GuestCartExpiryHours = 24     // 1 day for guest
	MaxCartItems         = 50
	MaxItemQuantity      = 100
)

// Helper methods
func (c *Cart) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

func (c *Cart) GetExpiryTime(isGuest bool) time.Time {
	if isGuest {
		return time.Now().Add(GuestCartExpiryHours * time.Hour)
	}
	return time.Now().Add(CartExpiryHours * time.Hour)
}

func (ci *CartItem) ToResponse(product Product, productSize ProductSize) CartItemResponse {
	// Calculate current price
	currentPrice := product.Price
	if productSize.Price != nil {
		currentPrice = *productSize.Price
	} else if product.DiscountPrice != nil && *product.DiscountPrice > 0 {
		currentPrice = *product.DiscountPrice
	}

	// Check availability
	isAvailable := product.IsActive && productSize.IsActive && productSize.Stock >= ci.Quantity

	stockStatus := "available"
	message := ""

	if !product.IsActive {
		stockStatus = "product_inactive"
		message = "Sản phẩm không còn hoạt động"
		isAvailable = false
	} else if !productSize.IsActive {
		stockStatus = "size_inactive"
		message = "Size không còn hoạt động"
		isAvailable = false
	} else if productSize.Stock == 0 {
		stockStatus = "out_of_stock"
		message = "Hết hàng"
		isAvailable = false
	} else if productSize.Stock < ci.Quantity {
		stockStatus = "insufficient_stock"
		message = fmt.Sprintf("Chỉ còn %d sản phẩm", productSize.Stock)
		isAvailable = false
	}

	return CartItemResponse{
		ID:            ci.ID,
		ProductID:     ci.ProductID,
		ProductSizeID: ci.ProductSizeID,
		Product:       product.ToResponse(),
		ProductSize:   productSize.ToResponse(product.Price, product.DiscountPrice),
		Quantity:      ci.Quantity,
		Price:         ci.Price,
		CurrentPrice:  currentPrice,
		Subtotal:      float64(ci.Quantity) * ci.Price,
		IsAvailable:   isAvailable,
		StockStatus:   stockStatus,
		Message:       message,
		CreatedAt:     ci.CreatedAt,
		UpdatedAt:     ci.UpdatedAt,
	}
}

func (c *Cart) ToResponse() CartResponse {
	var items []CartItemResponse
	var subtotal float64
	var totalQuantity int
	var issues []string
	isValid := true

	for _, item := range c.CartItems {
		itemResponse := item.ToResponse(item.Product, item.ProductSize)
		items = append(items, itemResponse)

		totalQuantity += item.Quantity
		subtotal += itemResponse.Subtotal

		if !itemResponse.IsAvailable {
			isValid = false
			issues = append(issues, fmt.Sprintf("%s - %s: %s",
				item.Product.Name, item.ProductSize.Size, itemResponse.Message))
		}
	}

	// Check cart expiry
	if c.IsExpired() {
		isValid = false
		issues = append(issues, "Giỏ hàng đã hết hạn")
	}

	// For now, no discount logic - can be extended later
	discount := 0.0
	total := subtotal - discount

	return CartResponse{
		ID:            c.ID,
		UserID:        c.UserID,
		SessionID:     c.SessionID,
		ItemCount:     len(items),
		TotalQuantity: totalQuantity,
		Subtotal:      subtotal,
		Discount:      discount,
		Total:         total,
		IsValid:       isValid,
		Issues:        issues,
		ExpiresAt:     c.ExpiresAt,
		CartItems:     items,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}
