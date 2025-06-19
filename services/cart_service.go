package services

import (
	"errors"
	"fmt"
	"storelite/config"
	"storelite/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartService struct {
	db *gorm.DB
}

func NewCartService() *CartService {
	return &CartService{
		db: config.GetDB(),
	}
}

// GetOrCreateCart lấy hoặc tạo cart cho user/session
func (cs *CartService) GetOrCreateCart(userID *uint, sessionID string) (*models.Cart, error) {
	var cart models.Cart

	// Query conditions
	query := cs.db.Preload("CartItems").
		Preload("CartItems.Product").
		Preload("CartItems.Product.Category").
		Preload("CartItems.ProductSize")

	if userID != nil {
		// Logged in user
		err := query.Where("user_id = ?", *userID).First(&cart).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		if err == gorm.ErrRecordNotFound {
			// Create new cart for user
			cart = models.Cart{
				UserID:    userID,
				SessionID: sessionID,
				ExpiresAt: cart.GetExpiryTime(false),
			}
			if err := cs.db.Create(&cart).Error; err != nil {
				return nil, err
			}
		}
	} else {
		// Guest user
		err := query.Where("session_id = ? AND user_id IS NULL", sessionID).First(&cart).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		if err == gorm.ErrRecordNotFound {
			// Create new cart for guest
			cart = models.Cart{
				SessionID: sessionID,
				ExpiresAt: cart.GetExpiryTime(true),
			}
			if err := cs.db.Create(&cart).Error; err != nil {
				return nil, err
			}
		}
	}

	// Clean expired cart items
	if cart.IsExpired() {
		cs.CleanExpiredCart(&cart)
	}

	// Reload cart with items
	cs.db.Preload("CartItems").
		Preload("CartItems.Product").
		Preload("CartItems.Product.Category").
		Preload("CartItems.ProductSize").
		First(&cart, cart.ID)

	return &cart, nil
}

// AddToCart thêm sản phẩm vào giỏ hàng
func (cs *CartService) AddToCart(cart *models.Cart, req models.AddToCartRequest) error {
	// Validate product and size
	var product models.Product
	if err := cs.db.Preload("ProductSizes").First(&product, req.ProductID).Error; err != nil {
		return errors.New("product not found")
	}

	if !product.IsActive {
		return errors.New("product is not active")
	}

	var productSize models.ProductSize
	if err := cs.db.Where("id = ? AND product_id = ?", req.ProductSizeID, req.ProductID).
		First(&productSize).Error; err != nil {
		return errors.New("product size not found")
	}

	if !productSize.IsActive {
		return errors.New("product size is not active")
	}

	// Check stock availability
	if productSize.Stock < req.Quantity {
		return fmt.Errorf("insufficient stock. Available: %d", productSize.Stock)
	}

	// Check if item already exists in cart
	var existingItem models.CartItem
	err := cs.db.Where("cart_id = ? AND product_id = ? AND product_size_id = ?",
		cart.ID, req.ProductID, req.ProductSizeID).First(&existingItem).Error

	// Calculate price
	price := product.Price
	if productSize.Price != nil {
		price = *productSize.Price
	} else if product.DiscountPrice != nil && *product.DiscountPrice > 0 {
		price = *product.DiscountPrice
	}

	if err == gorm.ErrRecordNotFound {
		// Check cart items limit
		var itemCount int64
		cs.db.Model(&models.CartItem{}).Where("cart_id = ?", cart.ID).Count(&itemCount)
		if itemCount >= models.MaxCartItems {
			return fmt.Errorf("cart items limit exceeded (%d items)", models.MaxCartItems)
		}

		// Create new cart item
		cartItem := models.CartItem{
			CartID:        cart.ID,
			ProductID:     req.ProductID,
			ProductSizeID: req.ProductSizeID,
			Quantity:      req.Quantity,
			Price:         price,
		}

		if err := cs.db.Create(&cartItem).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		// Update existing item
		newQuantity := existingItem.Quantity + req.Quantity
		if newQuantity > models.MaxItemQuantity {
			return fmt.Errorf("item quantity limit exceeded (%d items)", models.MaxItemQuantity)
		}

		if productSize.Stock < newQuantity {
			return fmt.Errorf("insufficient stock for total quantity. Available: %d", productSize.Stock)
		}

		existingItem.Quantity = newQuantity
		existingItem.Price = price // Update to current price
		if err := cs.db.Save(&existingItem).Error; err != nil {
			return err
		}
	}

	// Update cart expiry
	isGuest := cart.UserID == nil
	cart.ExpiresAt = cart.GetExpiryTime(isGuest)
	cs.db.Save(cart)

	return nil
}

// UpdateCartItem cập nhật số lượng item trong cart
func (cs *CartService) UpdateCartItem(cartID uint, itemID uint, quantity int) error {
	var cartItem models.CartItem
	if err := cs.db.Preload("ProductSize").
		Where("id = ? AND cart_id = ?", itemID, cartID).
		First(&cartItem).Error; err != nil {
		return errors.New("cart item not found")
	}

	if quantity == 0 {
		// Remove item
		return cs.db.Delete(&cartItem).Error
	}

	if quantity > models.MaxItemQuantity {
		return fmt.Errorf("item quantity limit exceeded (%d items)", models.MaxItemQuantity)
	}

	// Check stock
	if cartItem.ProductSize.Stock < quantity {
		return fmt.Errorf("insufficient stock. Available: %d", cartItem.ProductSize.Stock)
	}

	cartItem.Quantity = quantity
	return cs.db.Save(&cartItem).Error
}

// RemoveFromCart xóa item khỏi cart
func (cs *CartService) RemoveFromCart(cartID uint, itemID uint) error {
	return cs.db.Where("id = ? AND cart_id = ?", itemID, cartID).Delete(&models.CartItem{}).Error
}

// ClearCart xóa tất cả items trong cart
func (cs *CartService) ClearCart(cartID uint) error {
	return cs.db.Where("cart_id = ?", cartID).Delete(&models.CartItem{}).Error
}

// ValidateCart kiểm tra tính hợp lệ của cart trước khi checkout
func (cs *CartService) ValidateCart(cart *models.Cart) models.CartValidationResult {
	var issues []string

	// Check cart expiry
	if cart.IsExpired() {
		issues = append(issues, "Giỏ hàng đã hết hạn")
		return models.CartValidationResult{
			IsValid: false,
			Issues:  issues,
		}
	}

	// Check if cart is empty
	if len(cart.CartItems) == 0 {
		issues = append(issues, "Giỏ hàng trống")
		return models.CartValidationResult{
			IsValid: false,
			Issues:  issues,
		}
	}

	// Validate each item
	for _, item := range cart.CartItems {
		// Reload current stock and product status
		var currentProduct models.Product
		var currentSize models.ProductSize

		cs.db.First(&currentProduct, item.ProductID)
		cs.db.First(&currentSize, item.ProductSizeID)

		if !currentProduct.IsActive {
			issues = append(issues, fmt.Sprintf("%s không còn hoạt động", currentProduct.Name))
			continue
		}

		if !currentSize.IsActive {
			issues = append(issues, fmt.Sprintf("%s - Size %s không còn hoạt động",
				currentProduct.Name, currentSize.Size))
			continue
		}

		if currentSize.Stock == 0 {
			issues = append(issues, fmt.Sprintf("%s - Size %s đã hết hàng",
				currentProduct.Name, currentSize.Size))
			continue
		}

		if currentSize.Stock < item.Quantity {
			issues = append(issues, fmt.Sprintf("%s - Size %s chỉ còn %d sản phẩm (yêu cầu %d)",
				currentProduct.Name, currentSize.Size, currentSize.Stock, item.Quantity))
			continue
		}
	}

	return models.CartValidationResult{
		IsValid: len(issues) == 0,
		Issues:  issues,
	}
}

// CleanExpiredCart xóa cart đã hết hạn
func (cs *CartService) CleanExpiredCart(cart *models.Cart) error {
	// Clear all items
	if err := cs.ClearCart(cart.ID); err != nil {
		return err
	}

	// Update expiry time
	isGuest := cart.UserID == nil
	cart.ExpiresAt = cart.GetExpiryTime(isGuest)
	return cs.db.Save(cart).Error
}

// CleanExpiredCarts dọn dẹp tất cả carts đã hết hạn (có thể chạy bằng cron job)
func (cs *CartService) CleanExpiredCarts() error {
	// Delete expired cart items
	if err := cs.db.Exec(`
		DELETE FROM cart_items 
		WHERE cart_id IN (
			SELECT id FROM carts WHERE expires_at < ?
		)
	`, time.Now()).Error; err != nil {
		return err
	}

	// Delete expired carts
	return cs.db.Where("expires_at < ?", time.Now()).Delete(&models.Cart{}).Error
}

// MergeGuestCartToUser merge cart của guest vào user khi đăng nhập
func (cs *CartService) MergeGuestCartToUser(sessionID string, userID uint) error {
	// Get guest cart
	var guestCart models.Cart
	if err := cs.db.Preload("CartItems").
		Where("session_id = ? AND user_id IS NULL", sessionID).
		First(&guestCart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No guest cart to merge
		}
		return err
	}

	// Get or create user cart
	userCart, err := cs.GetOrCreateCart(&userID, sessionID)
	if err != nil {
		return err
	}

	// Merge items
	for _, guestItem := range guestCart.CartItems {
		req := models.AddToCartRequest{
			ProductID:     guestItem.ProductID,
			ProductSizeID: guestItem.ProductSizeID,
			Quantity:      guestItem.Quantity,
		}

		// Try to add to user cart (will merge with existing items)
		cs.AddToCart(userCart, req)
	}

	// Delete guest cart
	cs.db.Where("cart_id = ?", guestCart.ID).Delete(&models.CartItem{})
	cs.db.Delete(&guestCart)

	return nil
}

// GenerateSessionID tạo session ID mới cho guest user
func (cs *CartService) GenerateSessionID() string {
	return uuid.New().String()
}
