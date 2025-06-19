package models

import (
	"time"

	"gorm.io/gorm"
)

type Review struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id" gorm:"not null;index" validate:"required"`
	UserID    uint           `json:"user_id" gorm:"not null;index" validate:"required"`
	Comment   string         `json:"comment" gorm:"type:text" validate:"required,min=10,max=1000"`
	Rating    int            `json:"rating" gorm:"not null;check:rating >= 1 AND rating <= 5" validate:"required,min=1,max=5"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	User    User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type ReviewRequest struct {
	ProductID uint   `json:"product_id" validate:"required"`
	Comment   string `json:"comment" validate:"required,min=10,max=1000"`
	Rating    int    `json:"rating" validate:"required,min=1,max=5"`
}

type ReviewResponse struct {
	ID        uint         `json:"id"`
	ProductID uint         `json:"product_id"`
	UserID    uint         `json:"user_id"`
	User      UserResponse `json:"user"`
	Comment   string       `json:"comment"`
	Rating    int          `json:"rating"`
	IsActive  bool         `json:"is_active"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type ReviewListParams struct {
	ProductID *uint  `form:"product_id"`
	UserID    *uint  `form:"user_id"`
	Rating    *int   `form:"rating"`
	IsActive  *bool  `form:"is_active"`
	Sort      string `form:"sort"` // newest, oldest, rating_high, rating_low
	Page      int    `form:"page"`
	Limit     int    `form:"limit"`
}

type ReviewStats struct {
	TotalReviews    int64            `json:"total_reviews"`
	AverageRating   float64          `json:"average_rating"`
	RatingBreakdown map[string]int64 `json:"rating_breakdown"`
}

func (r *Review) ToResponse() ReviewResponse {
	return ReviewResponse{
		ID:        r.ID,
		ProductID: r.ProductID,
		UserID:    r.UserID,
		User:      r.User.ToResponse(),
		Comment:   r.Comment,
		Rating:    r.Rating,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

func GetRatingText(rating int) string {
	switch rating {
	case 1:
		return "Rất tệ"
	case 2:
		return "Tệ"
	case 3:
		return "Trung bình"
	case 4:
		return "Tốt"
	case 5:
		return "Rất tốt"
	default:
		return "Không xác định"
	}
}
