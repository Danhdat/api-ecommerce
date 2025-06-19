package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:255;uniqueIndex" validate:"required,min=2,max=255"`
	Slug        string         `json:"slug" gorm:"uniqueIndex;not null;size:255" validate:"required,min=2,max=255"`
	Description string         `json:"description" gorm:"type:text"`
	Thumbnail   string         `json:"thumbnail" gorm:"size:500"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Products []Product `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

type CategoryRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=255"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail" validate:"omitempty,url"`
	IsActive    *bool  `json:"is_active"`
}

type CategoryResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Thumbnail    string    `json:"thumbnail"`
	IsActive     bool      `json:"is_active"`
	ProductCount int64     `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		Thumbnail:   c.Thumbnail,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
