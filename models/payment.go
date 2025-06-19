package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment gateway constants
const (
	PaymentGatewayInternal     = "internal"      // For COD
	PaymentGatewayBankTransfer = "bank_transfer" // Bank transfer
	PaymentGatewayMomo         = "momo"          // MoMo
	PaymentGatewayZaloPay      = "zalopay"       // ZaloPay
	PaymentGatewayVNPay        = "vnpay"         // VNPay
)

// Payment status for payments table
const (
	PaymentRecordStatusPending   = "pending"
	PaymentRecordStatusCompleted = "completed"
	PaymentRecordStatusFailed    = "failed"
	PaymentRecordStatusCancelled = "cancelled"
)

type Payment struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	OrderID        uint           `json:"order_id" gorm:"not null;index"`
	TransactionID  string         `json:"transaction_id" gorm:"index;size:255"` // From payment gateway
	Amount         float64        `json:"amount" gorm:"not null;type:decimal(12,2)"`
	PaymentGateway string         `json:"payment_gateway" gorm:"not null;size:50"`
	Status         string         `json:"status" gorm:"not null;size:20;default:'pending'"`
	PaymentDate    *time.Time     `json:"payment_date"`
	ResponseData   JSONMap        `json:"response_data" gorm:"type:jsonb"` // Response from payment gateway
	Notes          string         `json:"notes" gorm:"type:text"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Order Order `json:"order,omitempty" gorm:"foreignKey:OrderID"`
}

type CreatePaymentRequest struct {
	OrderID        uint    `json:"order_id" validate:"required"`
	Amount         float64 `json:"amount" validate:"required,min=0"`
	PaymentGateway string  `json:"payment_gateway" validate:"required"`
	TransactionID  string  `json:"transaction_id"`
}

type PaymentResponse struct {
	ID             uint       `json:"id"`
	OrderID        uint       `json:"order_id"`
	OrderCode      string     `json:"order_code"`
	TransactionID  string     `json:"transaction_id"`
	Amount         float64    `json:"amount"`
	PaymentGateway string     `json:"payment_gateway"`
	GatewayText    string     `json:"gateway_text"`
	Status         string     `json:"status"`
	StatusText     string     `json:"status_text"`
	PaymentDate    *time.Time `json:"payment_date"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type BankTransferInfo struct {
	BankName      string  `json:"bank_name"`
	AccountNumber string  `json:"account_number"`
	AccountName   string  `json:"account_name"`
	Amount        float64 `json:"amount"`
	TransferNote  string  `json:"transfer_note"`
	QRCodeURL     string  `json:"qr_code_url"`
}

type PaymentMethodInfo struct {
	Method      string                 `json:"method"`
	MethodText  string                 `json:"method_text"`
	Description string                 `json:"description"`
	IsAvailable bool                   `json:"is_available"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:             p.ID,
		OrderID:        p.OrderID,
		OrderCode:      p.Order.OrderCode,
		TransactionID:  p.TransactionID,
		Amount:         p.Amount,
		PaymentGateway: p.PaymentGateway,
		GatewayText:    GetPaymentGatewayText(p.PaymentGateway),
		Status:         p.Status,
		StatusText:     GetPaymentRecordStatusText(p.Status),
		PaymentDate:    p.PaymentDate,
		Notes:          p.Notes,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func GetPaymentGatewayText(gateway string) string {
	switch gateway {
	case PaymentGatewayInternal:
		return "Hệ thống nội bộ"
	case PaymentGatewayBankTransfer:
		return "Chuyển khoản ngân hàng"
	case PaymentGatewayMomo:
		return "Ví MoMo"
	case PaymentGatewayZaloPay:
		return "ZaloPay"
	case PaymentGatewayVNPay:
		return "VNPay"
	default:
		return "Không xác định"
	}
}

func GetPaymentRecordStatusText(status string) string {
	switch status {
	case PaymentRecordStatusPending:
		return "Đang chờ xử lý"
	case PaymentRecordStatusCompleted:
		return "Hoàn thành"
	case PaymentRecordStatusFailed:
		return "Thất bại"
	case PaymentRecordStatusCancelled:
		return "Đã hủy"
	default:
		return "Không xác định"
	}
}
