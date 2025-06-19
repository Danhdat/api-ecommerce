package controllers

import (
	"net/http"
	"storelite/middleware"
	"storelite/models"
	"storelite/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderController struct {
	validator    *validator.Validate
	orderService *services.OrderService
}

func NewOrderController() *OrderController {
	return &OrderController{
		validator:    validator.New(),
		orderService: services.NewOrderService(),
	}
}

// getOrderIdentifier lấy thông tin để identify order (user ID hoặc session ID)
func (oc *OrderController) getOrderIdentifier(c *gin.Context) (*uint, string) {
	// Try to get user ID from JWT (if logged in)
	userID, exists := middleware.GetUserIDFromContext(c)
	if exists {
		return &userID, ""
	}

	// Get session ID for guest user
	sessionID := c.GetHeader("X-Session-ID")
	return nil, sessionID
}

// GetPaymentMethods lấy danh sách phương thức thanh toán
func (oc *OrderController) GetPaymentMethods(c *gin.Context) {
	methods := oc.orderService.GetPaymentMethods()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment methods retrieved successfully",
		Data:    methods,
	})
}

// CalculateOrderSummary tính toán tóm tắt đơn hàng
func (oc *OrderController) CalculateOrderSummary(c *gin.Context) {
	var req models.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := oc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	userID, sessionID := oc.getOrderIdentifier(c)

	// Get cart
	cartService := services.NewCartService()
	cart, err := cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get cart",
			Error:   err.Error(),
		})
		return
	}

	// Calculate summary
	summary, err := oc.orderService.CalculateOrderSummary(cart, req.ShippingAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to calculate order summary",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order summary calculated successfully",
		Data:    summary,
	})
}

// CreateOrder tạo đơn hàng mới
func (oc *OrderController) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := oc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	userID, sessionID := oc.getOrderIdentifier(c)

	// Create order
	order, err := oc.orderService.CreateOrder(userID, sessionID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to create order",
			Error:   err.Error(),
		})
		return
	}

	// Set session ID header for guest users
	if userID == nil {
		c.Header("X-Session-ID", sessionID)
	}

	//response := order.ToResponse()

	// Add payment instructions for specific methods
	var paymentInstructions interface{}
	switch req.PaymentMethod {
	case models.PaymentMethodBankTransfer:
		bankInfo, _ := oc.orderService.GetBankTransferInfo(order.OrderCode, order.FinalAmount)
		paymentInstructions = bankInfo
	case models.PaymentMethodCOD:
		paymentInstructions = map[string]interface{}{
			"message": "Thanh toán khi nhận hàng. Vui lòng chuẩn bị đúng số tiền khi nhận hàng.",
		}
	default:
		paymentInstructions = map[string]interface{}{
			"message": "Phương thức thanh toán này sẽ được tích hợp sớm.",
		}
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Order created successfully",
		Data: gin.H{
			"order":                order.ToResponse(),
			"payment_instructions": paymentInstructions,
		},
	})
}

// GetOrder lấy thông tin đơn hàng
func (oc *OrderController) GetOrder(c *gin.Context) {
	orderCode := c.Param("order_code")
	if orderCode == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Order code is required",
		})
		return
	}

	userID, sessionID := oc.getOrderIdentifier(c)

	order, err := oc.orderService.GetOrder(userID, sessionID, orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Order not found",
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
		Message: "Order retrieved successfully",
		Data:    order.ToResponse(),
	})
}

// GetBankTransferInfo lấy thông tin chuyển khoản
func (oc *OrderController) GetBankTransferInfo(c *gin.Context) {
	orderCode := c.Param("order_code")
	if orderCode == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Order code is required",
		})
		return
	}

	userID, sessionID := oc.getOrderIdentifier(c)

	// Get order to verify ownership and get amount
	order, err := oc.orderService.GetOrder(userID, sessionID, orderCode)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Order not found",
			Error:   err.Error(),
		})
		return
	}

	if order.PaymentMethod != models.PaymentMethodBankTransfer {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "This order does not use bank transfer payment method",
		})
		return
	}

	if order.PaymentStatus == models.PaymentStatusPaid {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "This order has already been paid",
		})
		return
	}

	bankInfo, err := oc.orderService.GetBankTransferInfo(orderCode, order.FinalAmount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get bank transfer info",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Bank transfer info retrieved successfully",
		Data:    bankInfo,
	})
}

// CancelOrder hủy đơn hàng
func (oc *OrderController) CancelOrder(c *gin.Context) {
	orderCode := c.Param("order_code")
	if orderCode == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Order code is required",
		})
		return
	}

	var req struct {
		Reason string `json:"reason" validate:"required,min=5,max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := oc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	userID, sessionID := oc.getOrderIdentifier(c)

	if err := oc.orderService.CancelOrder(userID, sessionID, orderCode, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to cancel order",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Order cancelled successfully",
	})
}

// ProcessPayment xử lý callback thanh toán (webhook)
func (oc *OrderController) ProcessPayment(c *gin.Context) {
	var req struct {
		OrderCode     string                 `json:"order_code" validate:"required"`
		TransactionID string                 `json:"transaction_id" validate:"required"`
		ResponseData  map[string]interface{} `json:"response_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := oc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	if err := oc.orderService.ProcessPayment(req.OrderCode, req.TransactionID, req.ResponseData); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to process payment",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Payment processed successfully",
	})
}

// ConfirmCODPayment xác nhận thanh toán COD (Admin only)
func (oc *OrderController) ConfirmCODPayment(c *gin.Context) {
	orderCode := c.Param("order_code")
	if orderCode == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Order code is required",
		})
		return
	}

	responseData := map[string]interface{}{
		"payment_method": "cod",
		"confirmed_by":   "admin",
		"confirmed_at":   time.Now(),
	}

	if err := oc.orderService.ProcessPayment(orderCode, "COD_"+orderCode, responseData); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Failed to confirm COD payment",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "COD payment confirmed successfully",
	})
}
