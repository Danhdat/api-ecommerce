package models

import (
	"time"

	"gorm.io/gorm"
)

type ProductSize struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id" gorm:"not null;index" validate:"required"`
	Size      string         `json:"size" gorm:"not null;size:50" validate:"required,min=1,max=50"`
	Stock     int            `json:"stock" gorm:"not null;default:0" validate:"min=0"`
	Price     *float64       `json:"price" gorm:"type:decimal(10,2)" validate:"omitempty,min=0"` // Giá riêng cho size này (nếu có)
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}

type ProductSizeRequest struct {
	Size     string   `json:"size" validate:"required,min=1,max=50"`
	Stock    int      `json:"stock" validate:"min=0"`
	Price    *float64 `json:"price" validate:"omitempty,min=0"`
	IsActive *bool    `json:"is_active"`
}

type ProductSizeResponse struct {
	ID          uint      `json:"id"`
	ProductID   uint      `json:"product_id"`
	Size        string    `json:"size"`
	Stock       int       `json:"stock"`
	Price       *float64  `json:"price"`
	FinalPrice  float64   `json:"final_price"` // Giá cuối cùng (size price hoặc product price)
	IsActive    bool      `json:"is_active"`
	StockStatus string    `json:"stock_status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Unique constraint for product_id + size
func (ProductSize) TableName() string {
	return "product_sizes"
}

func (ps *ProductSize) ToResponse(productPrice float64, productDiscountPrice *float64) ProductSizeResponse {
	// Calculate final price
	finalPrice := productPrice
	if ps.Price != nil {
		finalPrice = *ps.Price
	} else if productDiscountPrice != nil && *productDiscountPrice > 0 {
		finalPrice = *productDiscountPrice
	}

	// Determine stock status
	stockStatus := "in_stock"
	if ps.Stock == 0 {
		stockStatus = "out_of_stock"
	} else if ps.Stock <= 5 {
		stockStatus = "low_stock"
	}

	return ProductSizeResponse{
		ID:          ps.ID,
		ProductID:   ps.ProductID,
		Size:        ps.Size,
		Stock:       ps.Stock,
		Price:       ps.Price,
		FinalPrice:  finalPrice,
		IsActive:    ps.IsActive,
		StockStatus: stockStatus,
		CreatedAt:   ps.CreatedAt,
		UpdatedAt:   ps.UpdatedAt,
	}
}
