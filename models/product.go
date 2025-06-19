package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringArray for handling JSON array of strings
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return errors.New("cannot scan into StringArray")
	}
}

type Product struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	CategoryID    uint           `json:"category_id" gorm:"not null;index" validate:"required"`
	Name          string         `json:"name" gorm:"not null;size:255;index" validate:"required,min=2,max=255"`
	Slug          string         `json:"slug" gorm:"uniqueIndex;not null;size:255" validate:"required,min=2,max=255"`
	Price         float64        `json:"price" gorm:"not null;type:decimal(10,2)" validate:"required,min=0"`
	DiscountPrice *float64       `json:"discount_price" gorm:"type:decimal(10,2)" validate:"omitempty,min=0"`
	Description   string         `json:"description" gorm:"type:text"`
	TotalStock    int            `json:"total_stock" gorm:"default:0"` // Tổng stock của tất cả sizes
	Thumbnail     string         `json:"thumbnail" gorm:"size:500"`
	Images        StringArray    `json:"images" gorm:"type:jsonb"`
	IsFeatured    bool           `json:"is_featured" gorm:"default:false"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	ViewCount     int64          `json:"view_count" gorm:"default:0"`
	AverageRating float64        `json:"average_rating" gorm:"type:decimal(3,2);default:0"`
	ReviewCount   int64          `json:"review_count" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Category     Category      `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	ProductSizes []ProductSize `json:"product_sizes,omitempty" gorm:"foreignKey:ProductID"`
	Reviews      []Review      `json:"reviews,omitempty" gorm:"foreignKey:ProductID"`
}

type ProductRequest struct {
	CategoryID    uint                 `json:"category_id" validate:"required"`
	Name          string               `json:"name" validate:"required,min=2,max=255"`
	Price         float64              `json:"price" validate:"required,min=0"`
	DiscountPrice *float64             `json:"discount_price" validate:"omitempty,min=0"`
	Description   string               `json:"description"`
	Thumbnail     string               `json:"thumbnail" validate:"omitempty,url"`
	Images        StringArray          `json:"images"`
	IsFeatured    *bool                `json:"is_featured"`
	IsActive      *bool                `json:"is_active"`
	Sizes         []ProductSizeRequest `json:"sizes" validate:"required,min=1,dive"`
}

type ProductResponse struct {
	ID            uint                  `json:"id"`
	CategoryID    uint                  `json:"category_id"`
	Category      CategoryResponse      `json:"category"`
	Name          string                `json:"name"`
	Slug          string                `json:"slug"`
	Price         float64               `json:"price"`
	DiscountPrice *float64              `json:"discount_price"`
	FinalPrice    float64               `json:"final_price"`
	DiscountRate  float64               `json:"discount_rate"`
	Description   string                `json:"description"`
	TotalStock    int                   `json:"total_stock"`
	StockStatus   string                `json:"stock_status"`
	Thumbnail     string                `json:"thumbnail"`
	Images        StringArray           `json:"images"`
	IsFeatured    bool                  `json:"is_featured"`
	IsActive      bool                  `json:"is_active"`
	ViewCount     int64                 `json:"view_count"`
	AverageRating float64               `json:"average_rating"`
	ReviewCount   int64                 `json:"review_count"`
	Sizes         []ProductSizeResponse `json:"sizes"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
}

type ProductListParams struct {
	CategoryID *uint    `form:"category_id"`
	Search     string   `form:"search"`
	MinPrice   *float64 `form:"min_price"`
	MaxPrice   *float64 `form:"max_price"`
	IsFeatured *bool    `form:"is_featured"`
	IsActive   *bool    `form:"is_active"`
	Sort       string   `form:"sort"` // price_asc, price_desc, name_asc, name_desc, newest, oldest, rating, popular
	Page       int      `form:"page"`
	Limit      int      `form:"limit"`
	InStock    *bool    `form:"in_stock"`
}

func (p *Product) ToResponse() ProductResponse {
	finalPrice := p.Price
	discountRate := 0.0

	if p.DiscountPrice != nil && *p.DiscountPrice > 0 && *p.DiscountPrice < p.Price {
		finalPrice = *p.DiscountPrice
		discountRate = ((p.Price - *p.DiscountPrice) / p.Price) * 100
	}

	stockStatus := "in_stock"
	if p.TotalStock == 0 {
		stockStatus = "out_of_stock"
	} else if p.TotalStock <= 10 {
		stockStatus = "low_stock"
	}

	// Convert ProductSizes to ProductSizeResponse
	var sizeResponses []ProductSizeResponse
	for _, size := range p.ProductSizes {
		sizeResponses = append(sizeResponses, size.ToResponse(p.Price, p.DiscountPrice))
	}

	return ProductResponse{
		ID:            p.ID,
		CategoryID:    p.CategoryID,
		Category:      p.Category.ToResponse(),
		Name:          p.Name,
		Slug:          p.Slug,
		Price:         p.Price,
		DiscountPrice: p.DiscountPrice,
		FinalPrice:    finalPrice,
		DiscountRate:  discountRate,
		Description:   p.Description,
		TotalStock:    p.TotalStock,
		StockStatus:   stockStatus,
		Thumbnail:     p.Thumbnail,
		Images:        p.Images,
		IsFeatured:    p.IsFeatured,
		IsActive:      p.IsActive,
		ViewCount:     p.ViewCount,
		AverageRating: p.AverageRating,
		ReviewCount:   p.ReviewCount,
		Sizes:         sizeResponses,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func (p *Product) UpdateRatingStats() error {
	// This will be called after review creation/update/deletion
	// Implementation in product service
	return nil
}
