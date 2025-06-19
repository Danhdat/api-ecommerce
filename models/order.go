package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Order status constants
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusDelivered = "delivered"
	OrderStatusCancelled = "cancelled"
)

// Payment status constants
const (
	PaymentStatusUnpaid = "unpaid"
	PaymentStatusPaid   = "paid"
	PaymentStatusFailed = "failed"
)

// Payment method constants
const (
	PaymentMethodCOD          = "cod"           // Cash on delivery
	PaymentMethodBankTransfer = "bank_transfer" // Bank transfer
	PaymentMethodMomo         = "momo"          // Momo wallet
	PaymentMethodZaloPay      = "zalopay"       // ZaloPay
	PaymentMethodVNPay        = "vnpay"         // VNPay
)

// JSONMap for storing JSON data
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = JSONMap{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return errors.New("cannot scan into JSONMap")
	}
}

type Order struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          *uint          `json:"user_id" gorm:"index"` // NULL for guest users
	OrderCode       string         `json:"order_code" gorm:"uniqueIndex;size:5;not null"`
	CustomerName    string         `json:"customer_name" gorm:"not null;size:255"`
	CustomerEmail   string         `json:"customer_email" gorm:"not null;size:255"`
	CustomerPhone   string         `json:"customer_phone" gorm:"not null;size:20"`
	ShippingAddress JSONMap        `json:"shipping_address" gorm:"type:jsonb;not null"`
	TotalAmount     float64        `json:"total_amount" gorm:"not null;type:decimal(12,2)"`
	ShippingFee     float64        `json:"shipping_fee" gorm:"type:decimal(10,2);default:0"`
	DiscountAmount  float64        `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	FinalAmount     float64        `json:"final_amount" gorm:"not null;type:decimal(12,2)"`
	Status          string         `json:"status" gorm:"not null;size:20;default:'pending'"`
	PaymentMethod   string         `json:"payment_method" gorm:"not null;size:50"`
	PaymentStatus   string         `json:"payment_status" gorm:"not null;size:20;default:'unpaid'"`
	Notes           string         `json:"notes" gorm:"type:text"`
	SessionID       string         `json:"session_id" gorm:"index;size:255"` // For guest orders
	ExpiresAt       time.Time      `json:"expires_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`
	Payments   []Payment   `json:"payments,omitempty" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	OrderID       uint           `json:"order_id" gorm:"not null;index"`
	ProductID     uint           `json:"product_id" gorm:"not null;index"`
	ProductSizeID uint           `json:"product_size_id" gorm:"not null;index"`
	ProductName   string         `json:"product_name" gorm:"not null;size:255"` // Snapshot
	ProductSize   string         `json:"product_size" gorm:"not null;size:50"`  // Snapshot
	Quantity      int            `json:"quantity" gorm:"not null"`
	UnitPrice     float64        `json:"unit_price" gorm:"not null;type:decimal(10,2)"`
	TotalPrice    float64        `json:"total_price" gorm:"not null;type:decimal(12,2)"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Order           Order       `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Product         Product     `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	ProductSizeInfo ProductSize `json:"product_size_info,omitempty" gorm:"foreignKey:ProductSizeID"`
}

// Request/Response structs
type ShippingAddress struct {
	FullName    string `json:"full_name" validate:"required,min=2,max=255"`
	Phone       string `json:"phone" validate:"required,min=10,max=20"`
	AddressLine string `json:"address_line" validate:"required,min=10,max=500"`
	City        string `json:"city" validate:"required,min=2,max=100"`
	District    string `json:"district" validate:"required,min=2,max=100"`
	Ward        string `json:"ward" validate:"required,min=2,max=100"`
	PostalCode  string `json:"postal_code"`
}

type CreateOrderRequest struct {
	CustomerName    string          `json:"customer_name" validate:"required,min=2,max=255"`
	CustomerEmail   string          `json:"customer_email" validate:"required,email"`
	CustomerPhone   string          `json:"customer_phone" validate:"required,min=10,max=20"`
	ShippingAddress ShippingAddress `json:"shipping_address" validate:"required"`
	PaymentMethod   string          `json:"payment_method" validate:"required,oneof=cod bank_transfer momo zalopay vnpay"`
	Notes           string          `json:"notes"`
}

type OrderItemResponse struct {
	ID            uint    `json:"id"`
	ProductID     uint    `json:"product_id"`
	ProductSizeID uint    `json:"product_size_id"`
	ProductName   string  `json:"product_name"`
	ProductSize   string  `json:"product_size"`
	Quantity      int     `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	TotalPrice    float64 `json:"total_price"`
}

type OrderResponse struct {
	ID                uint                `json:"id"`
	UserID            *uint               `json:"user_id"`
	OrderCode         string              `json:"order_code"`
	CustomerName      string              `json:"customer_name"`
	CustomerEmail     string              `json:"customer_email"`
	CustomerPhone     string              `json:"customer_phone"`
	ShippingAddress   ShippingAddress     `json:"shipping_address"`
	TotalAmount       float64             `json:"total_amount"`
	ShippingFee       float64             `json:"shipping_fee"`
	DiscountAmount    float64             `json:"discount_amount"`
	FinalAmount       float64             `json:"final_amount"`
	Status            string              `json:"status"`
	StatusText        string              `json:"status_text"`
	PaymentMethod     string              `json:"payment_method"`
	PaymentMethodText string              `json:"payment_method_text"`
	PaymentStatus     string              `json:"payment_status"`
	PaymentStatusText string              `json:"payment_status_text"`
	Notes             string              `json:"notes"`
	ExpiresAt         time.Time           `json:"expires_at"`
	OrderItems        []OrderItemResponse `json:"order_items"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

type OrderSummary struct {
	ItemCount      int     `json:"item_count"`
	TotalQuantity  int     `json:"total_quantity"`
	TotalAmount    float64 `json:"total_amount"`
	ShippingFee    float64 `json:"shipping_fee"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
}

// Constants
const (
	OrderExpiryMinutes = 30 // 30 minutes for payment
	MaxOrderCode       = 99999
	MinOrderCode       = 1
)

// Helper methods
func (o *Order) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusPaid
}

func (o *Order) ToResponse() OrderResponse {
	var shippingAddr ShippingAddress
	shippingAddrBytes, _ := json.Marshal(o.ShippingAddress)
	json.Unmarshal(shippingAddrBytes, &shippingAddr)

	var items []OrderItemResponse
	for _, item := range o.OrderItems {
		items = append(items, OrderItemResponse{
			ID:            item.ID,
			ProductID:     item.ProductID,
			ProductSizeID: item.ProductSizeID,
			ProductName:   item.ProductName,
			ProductSize:   item.ProductSize,
			Quantity:      item.Quantity,
			UnitPrice:     item.UnitPrice,
			TotalPrice:    item.TotalPrice,
		})
	}

	return OrderResponse{
		ID:                o.ID,
		UserID:            o.UserID,
		OrderCode:         o.OrderCode,
		CustomerName:      o.CustomerName,
		CustomerEmail:     o.CustomerEmail,
		CustomerPhone:     o.CustomerPhone,
		ShippingAddress:   shippingAddr,
		TotalAmount:       o.TotalAmount,
		ShippingFee:       o.ShippingFee,
		DiscountAmount:    o.DiscountAmount,
		FinalAmount:       o.FinalAmount,
		Status:            o.Status,
		StatusText:        GetOrderStatusText(o.Status),
		PaymentMethod:     o.PaymentMethod,
		PaymentMethodText: GetPaymentMethodText(o.PaymentMethod),
		PaymentStatus:     o.PaymentStatus,
		PaymentStatusText: GetPaymentStatusText(o.PaymentStatus),
		Notes:             o.Notes,
		ExpiresAt:         o.ExpiresAt,
		OrderItems:        items,
		CreatedAt:         o.CreatedAt,
		UpdatedAt:         o.UpdatedAt,
	}
}

// Helper functions for text display
func GetOrderStatusText(status string) string {
	switch status {
	case OrderStatusPending:
		return "Đang chờ xử lý"
	case OrderStatusPaid:
		return "Đã thanh toán"
	case OrderStatusShipped:
		return "Đang giao hàng"
	case OrderStatusDelivered:
		return "Đã giao hàng"
	case OrderStatusCancelled:
		return "Đã hủy"
	default:
		return "Không xác định"
	}
}

func GetPaymentMethodText(method string) string {
	switch method {
	case PaymentMethodCOD:
		return "Thanh toán khi nhận hàng"
	case PaymentMethodBankTransfer:
		return "Chuyển khoản ngân hàng"
	case PaymentMethodMomo:
		return "Ví MoMo"
	case PaymentMethodZaloPay:
		return "ZaloPay"
	case PaymentMethodVNPay:
		return "VNPay"
	default:
		return "Không xác định"
	}
}

func GetPaymentStatusText(status string) string {
	switch status {
	case PaymentStatusUnpaid:
		return "Chưa thanh toán"
	case PaymentStatusPaid:
		return "Đã thanh toán"
	case PaymentStatusFailed:
		return "Thanh toán thất bại"
	default:
		return "Không xác định"
	}
}
