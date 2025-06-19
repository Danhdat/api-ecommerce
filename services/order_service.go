package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"storelite/config"
	"storelite/models"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"
)

type OrderService struct {
	db            *gorm.DB
	cartService   *CartService
	emailService  *EmailService
	orderCodeLock sync.Mutex
}

func NewOrderService() *OrderService {
	return &OrderService{
		db:           config.GetDB(),
		cartService:  NewCartService(),
		emailService: NewEmailService(),
	}
}

// GenerateOrderCode tạo mã đơn hàng theo quy tắc 00001-99999
func (os *OrderService) GenerateOrderCode() (string, error) {
	os.orderCodeLock.Lock()
	defer os.orderCodeLock.Unlock()

	// Get last order code
	var lastOrder models.Order
	err := os.db.Select("order_code").Order("id DESC").First(&lastOrder).Error

	var nextCode int = models.MinOrderCode

	if err != gorm.ErrRecordNotFound {
		if err != nil {
			return "", err
		}

		// Parse current code and increment
		currentCode, parseErr := strconv.Atoi(lastOrder.OrderCode)
		if parseErr != nil {
			return "", parseErr
		}

		nextCode = currentCode + 1
		if nextCode > models.MaxOrderCode {
			nextCode = models.MinOrderCode
		}
	}

	// Format with leading zeros
	orderCode := fmt.Sprintf("%05d", nextCode)

	// Check if code already exists (safety check)
	var existingOrder models.Order
	if err := os.db.Where("order_code = ?", orderCode).First(&existingOrder).Error; err == nil {
		// If exists, find next available code
		for i := nextCode; i <= models.MaxOrderCode; i++ {
			testCode := fmt.Sprintf("%05d", i)
			if err := os.db.Where("order_code = ?", testCode).First(&existingOrder).Error; err == gorm.ErrRecordNotFound {
				return testCode, nil
			}
		}
		// If all codes are taken, start from beginning
		for i := models.MinOrderCode; i < nextCode; i++ {
			testCode := fmt.Sprintf("%05d", i)
			if err := os.db.Where("order_code = ?", testCode).First(&existingOrder).Error; err == gorm.ErrRecordNotFound {
				return testCode, nil
			}
		}
		return "", errors.New("all order codes are taken")
	}

	return orderCode, nil
}

// CalculateOrderSummary tính toán tổng đơn hàng
func (os *OrderService) CalculateOrderSummary(cart *models.Cart, shippingAddress models.ShippingAddress) (*models.OrderSummary, error) {
	if len(cart.CartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	var totalAmount float64
	var totalQuantity int

	for _, item := range cart.CartItems {
		totalAmount += float64(item.Quantity) * item.Price
		totalQuantity += item.Quantity
	}

	// Calculate shipping fee (example logic - can be customized)
	shippingFee := os.calculateShippingFee(shippingAddress, totalAmount)

	// Calculate discount (example logic - can be customized)
	discountAmount := os.calculateDiscount(totalAmount)

	finalAmount := totalAmount + shippingFee - discountAmount

	return &models.OrderSummary{
		ItemCount:      len(cart.CartItems),
		TotalQuantity:  totalQuantity,
		TotalAmount:    totalAmount,
		ShippingFee:    shippingFee,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
	}, nil
}

// CreateOrder tạo đơn hàng từ cart
func (os *OrderService) CreateOrder(userID *uint, sessionID string, req models.CreateOrderRequest) (*models.Order, error) {
	// Get cart
	cart, err := os.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %v", err)
	}

	// Validate cart
	validation := os.cartService.ValidateCart(cart)
	if !validation.IsValid {
		return nil, fmt.Errorf("cart validation failed: %v", validation.Issues)
	}

	// Convert shipping address to JSONMap
	shippingAddrBytes, _ := json.Marshal(req.ShippingAddress)
	var shippingAddrMap models.JSONMap
	json.Unmarshal(shippingAddrBytes, &shippingAddrMap)

	// Calculate order summary
	summary, err := os.CalculateOrderSummary(cart, req.ShippingAddress)
	if err != nil {
		return nil, err
	}

	// Start transaction
	tx := os.db.Begin()

	// Generate order code
	orderCode, err := os.GenerateOrderCode()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to generate order code: %v", err)
	}

	// Create order
	order := models.Order{
		UserID:          userID,
		OrderCode:       orderCode,
		CustomerName:    req.CustomerName,
		CustomerEmail:   req.CustomerEmail,
		CustomerPhone:   req.CustomerPhone,
		ShippingAddress: shippingAddrMap,
		TotalAmount:     summary.TotalAmount,
		ShippingFee:     summary.ShippingFee,
		DiscountAmount:  summary.DiscountAmount,
		FinalAmount:     summary.FinalAmount,
		Status:          models.OrderStatusPending,
		PaymentMethod:   req.PaymentMethod,
		PaymentStatus:   models.PaymentStatusUnpaid,
		Notes:           req.Notes,
		SessionID:       sessionID,
		ExpiresAt:       time.Now().Add(models.OrderExpiryMinutes * time.Minute),
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	// Create order items and update stock
	for _, cartItem := range cart.CartItems {
		// Lock stock for concurrent access
		var productSize models.ProductSize
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", cartItem.ProductSizeID).
			First(&productSize).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("product size not found: %v", err)
		}

		// Check stock again
		if productSize.Stock < cartItem.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient stock for %s - %s. Available: %d, Required: %d",
				cartItem.Product.Name, productSize.Size, productSize.Stock, cartItem.Quantity)
		}

		// Update stock
		productSize.Stock -= cartItem.Quantity
		if err := tx.Save(&productSize).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update stock: %v", err)
		}

		// Update product total stock
		var product models.Product
		if err := tx.First(&product, cartItem.ProductID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("product not found: %v", err)
		}

		product.TotalStock -= cartItem.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update product total stock: %v", err)
		}

		// Create order item
		orderItem := models.OrderItem{
			OrderID:       order.ID,
			ProductID:     cartItem.ProductID,
			ProductSizeID: cartItem.ProductSizeID,
			ProductName:   cartItem.Product.Name,
			ProductSize:   productSize.Size,
			Quantity:      cartItem.Quantity,
			UnitPrice:     cartItem.Price,
			TotalPrice:    float64(cartItem.Quantity) * cartItem.Price,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create order item: %v", err)
		}
	}

	// Clear cart after successful order creation
	if err := os.cartService.ClearCart(cart.ID); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to clear cart: %v", err)
	}

	// Create initial payment record
	payment := models.Payment{
		OrderID:        order.ID,
		Amount:         order.FinalAmount,
		PaymentGateway: os.getPaymentGateway(req.PaymentMethod),
		Status:         models.PaymentRecordStatusPending,
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create payment record: %v", err)
	}

	tx.Commit()

	// Load complete order
	os.db.Preload("OrderItems").Preload("Payments").First(&order, order.ID)

	// Send confirmation email
	go os.sendOrderConfirmationEmail(&order)

	return &order, nil
}

// GetOrder lấy thông tin đơn hàng
func (os *OrderService) GetOrder(userID *uint, sessionID string, orderCode string) (*models.Order, error) {
	var order models.Order
	query := os.db.Preload("OrderItems").Preload("Payments")

	if userID != nil {
		query = query.Where("order_code = ? AND user_id = ?", orderCode, *userID)
	} else {
		query = query.Where("order_code = ? AND session_id = ? AND user_id IS NULL", orderCode, sessionID)
	}

	if err := query.First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

// GetPaymentMethods lấy danh sách phương thức thanh toán
func (os *OrderService) GetPaymentMethods() []models.PaymentMethodInfo {
	return []models.PaymentMethodInfo{
		{
			Method:      models.PaymentMethodCOD,
			MethodText:  "Thanh toán khi nhận hàng",
			Description: "Thanh toán bằng tiền mặt khi nhận hàng",
			IsAvailable: true,
		},
		{
			Method:      models.PaymentMethodBankTransfer,
			MethodText:  "Chuyển khoản ngân hàng",
			Description: "Chuyển khoản qua ngân hàng với QR Code",
			IsAvailable: true,
			Extra: map[string]interface{}{
				"bank_info": os.getBankTransferInfo(),
			},
		},
		{
			Method:      models.PaymentMethodMomo,
			MethodText:  "Ví MoMo",
			Description: "Thanh toán qua ví điện tử MoMo",
			IsAvailable: false, // Will be enabled when integrated
			Extra: map[string]interface{}{
				"note": "Tính năng sẽ được tích hợp sớm",
			},
		},
		{
			Method:      models.PaymentMethodZaloPay,
			MethodText:  "ZaloPay",
			Description: "Thanh toán qua ví điện tử ZaloPay",
			IsAvailable: false, // Will be enabled when integrated
			Extra: map[string]interface{}{
				"note": "Tính năng sẽ được tích hợp sớm",
			},
		},
		{
			Method:      models.PaymentMethodVNPay,
			MethodText:  "VNPay",
			Description: "Thanh toán qua cổng VNPay",
			IsAvailable: false, // Will be enabled when integrated
			Extra: map[string]interface{}{
				"note": "Tính năng sẽ được tích hợp sớm",
			},
		},
	}
}

// CancelOrder hủy đơn hàng
func (os *OrderService) CancelOrder(userID *uint, sessionID string, orderCode string, reason string) error {
	order, err := os.GetOrder(userID, sessionID, orderCode)
	if err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	if !order.CanBeCancelled() {
		return fmt.Errorf("order cannot be cancelled. Current status: %s", order.Status)
	}

	tx := os.db.Begin()

	// Update order status
	order.Status = models.OrderStatusCancelled
	order.Notes = fmt.Sprintf("%s\nLý do hủy: %s", order.Notes, reason)

	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order: %v", err)
	}

	// Restore stock
	for _, item := range order.OrderItems {
		// Update product size stock
		if err := tx.Model(&models.ProductSize{}).
			Where("id = ?", item.ProductSizeID).
			Update("stock", gorm.Expr("stock + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to restore product size stock: %v", err)
		}

		// Update product total stock
		if err := tx.Model(&models.Product{}).
			Where("id = ?", item.ProductID).
			Update("total_stock", gorm.Expr("total_stock + ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to restore product total stock: %v", err)
		}
	}

	// Update payment status
	if err := tx.Model(&models.Payment{}).
		Where("order_id = ?", order.ID).
		Update("status", models.PaymentRecordStatusCancelled).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update payment status: %v", err)
	}

	tx.Commit()

	// Send cancellation email
	go os.sendOrderCancellationEmail(order)

	return nil
}

// ProcessPayment xử lý thanh toán
func (os *OrderService) ProcessPayment(orderCode string, transactionID string, responseData map[string]interface{}) error {
	var order models.Order
	if err := os.db.Preload("Payments").Where("order_code = ?", orderCode).First(&order).Error; err != nil {
		return fmt.Errorf("order not found: %v", err)
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		return nil // Already paid
	}

	tx := os.db.Begin()

	// Update order payment status
	order.PaymentStatus = models.PaymentStatusPaid
	order.Status = models.OrderStatusPaid
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update order: %v", err)
	}

	// Update payment record
	now := time.Now()
	responseDataJSON := models.JSONMap(responseData)

	if err := tx.Model(&models.Payment{}).
		Where("order_id = ?", order.ID).
		Updates(map[string]interface{}{
			"transaction_id": transactionID,
			"status":         models.PaymentRecordStatusCompleted,
			"payment_date":   &now,
			"response_data":  responseDataJSON,
		}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update payment: %v", err)
	}

	tx.Commit()

	// Send payment confirmation email
	go os.sendPaymentConfirmationEmail(&order)

	return nil
}

// GetBankTransferInfo lấy thông tin chuyển khoản
func (os *OrderService) GetBankTransferInfo(orderCode string, amount float64) (*models.BankTransferInfo, error) {
	bankInfo := os.getBankTransferInfo()

	transferNote := fmt.Sprintf("DH %s", orderCode)
	qrCodeURL := os.generateQRCode(bankInfo["account_number"].(string), amount, transferNote)

	return &models.BankTransferInfo{
		BankName:      bankInfo["bank_name"].(string),
		AccountNumber: bankInfo["account_number"].(string),
		AccountName:   bankInfo["account_name"].(string),
		Amount:        amount,
		TransferNote:  transferNote,
		QRCodeURL:     qrCodeURL,
	}, nil
}

// CleanExpiredOrders dọn dẹp đơn hàng hết hạn
func (os *OrderService) CleanExpiredOrders() error {
	var expiredOrders []models.Order
	if err := os.db.Preload("OrderItems").
		Where("expires_at < ? AND status = ? AND payment_status = ?",
			time.Now(), models.OrderStatusPending, models.PaymentStatusUnpaid).
		Find(&expiredOrders).Error; err != nil {
		return err
	}

	for _, order := range expiredOrders {
		// Cancel expired order
		os.CancelOrder(order.UserID, order.SessionID, order.OrderCode, "Đơn hàng hết hạn thanh toán")
	}

	return nil
}

// Helper methods
func (os *OrderService) calculateShippingFee(address models.ShippingAddress, totalAmount float64) float64 {
	// Example shipping fee calculation
	// Free shipping for orders over 500,000 VND
	if totalAmount >= 500000 {
		return 0
	}

	// Base shipping fee
	baseShippingFee := 30000.0

	// Additional fee for remote areas (example logic)
	remoteCities := []string{"Cà Mau", "An Giang", "Kiên Giang", "Hà Giang"}
	for _, city := range remoteCities {
		if address.City == city {
			baseShippingFee += 20000
			break
		}
	}

	return baseShippingFee
}

func (os *OrderService) calculateDiscount(totalAmount float64) float64 {
	// Example discount calculation
	// 5% discount for orders over 1,000,000 VND
	if totalAmount >= 1000000 {
		return totalAmount * 0.05
	}

	return 0
}

func (os *OrderService) getPaymentGateway(paymentMethod string) string {
	switch paymentMethod {
	case models.PaymentMethodCOD:
		return models.PaymentGatewayInternal
	case models.PaymentMethodBankTransfer:
		return models.PaymentGatewayBankTransfer
	case models.PaymentMethodMomo:
		return models.PaymentGatewayMomo
	case models.PaymentMethodZaloPay:
		return models.PaymentGatewayZaloPay
	case models.PaymentMethodVNPay:
		return models.PaymentGatewayVNPay
	default:
		return models.PaymentGatewayInternal
	}
}

func (os *OrderService) getBankTransferInfo() map[string]interface{} {
	return map[string]interface{}{
		"bank_name":      "Ngân hàng TMCP Á Châu (ACB)",
		"account_number": "1234567890",
		"account_name":   "CONG TY TNHH E-COMMERCE",
	}
}

func (os *OrderService) generateQRCode(accountNumber string, amount float64, transferNote string) string {
	// Generate QR code URL for bank transfer
	// This is a placeholder - in production, integrate with actual QR code generation service
	return fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=Bank:%s,Amount:%.0f,Note:%s",
		accountNumber, amount, transferNote)
}

func (os *OrderService) sendOrderConfirmationEmail(order *models.Order) {
	// Send order confirmation email
	// Implementation depends on email service
	subject := fmt.Sprintf("Xác nhận đơn hàng #%s", order.OrderCode)
	// Email content would include order details, payment instructions, etc.
	_ = subject // Placeholder
}

func (os *OrderService) sendPaymentConfirmationEmail(order *models.Order) {
	// Send payment confirmation email
	subject := fmt.Sprintf("Thanh toán thành công đơn hàng #%s", order.OrderCode)
	// Email content would include payment confirmation, shipping info, etc.
	_ = subject // Placeholder
}

func (os *OrderService) sendOrderCancellationEmail(order *models.Order) {
	// Send order cancellation email
	subject := fmt.Sprintf("Đơn hàng #%s đã được hủy", order.OrderCode)
	// Email content would include cancellation reason, refund info, etc.
	_ = subject // Placeholder
}
