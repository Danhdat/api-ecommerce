package controllers

import (
	"net/http"
	"storelite/config"
	"storelite/middleware"
	"storelite/models"
	"storelite/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ReviewController struct {
	validator *validator.Validate
}

func NewReviewController() *ReviewController {
	return &ReviewController{
		validator: validator.New(),
	}
}

// CreateReview tạo đánh giá mới (Authenticated users only)
func (rc *ReviewController) CreateReview(c *gin.Context) {
	var req models.ReviewRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := rc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from JWT
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	// Check if product exists and is active
	var product models.Product
	if err := config.GetDB().Where("id = ? AND is_active = ?", req.ProductID, true).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Product not found or inactive",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Check if user already reviewed this product
	var existingReview models.Review
	if err := config.GetDB().Where("product_id = ? AND user_id = ?", req.ProductID, userID).First(&existingReview).Error; err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Message: "You have already reviewed this product",
		})
		return
	}

	// Sanitize comment
	req.Comment = utils.SanitizeString(req.Comment)

	review := models.Review{
		ProductID: req.ProductID,
		UserID:    userID,
		Comment:   req.Comment,
		Rating:    req.Rating,
		IsActive:  true,
	}

	// Start transaction
	tx := config.GetDB().Begin()

	if err := tx.Create(&review).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create review",
			Error:   err.Error(),
		})
		return
	}

	// Update product rating statistics
	if err := rc.updateProductRatingStats(tx, req.ProductID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update product statistics",
			Error:   err.Error(),
		})
		return
	}

	tx.Commit()

	// Load user for response
	config.GetDB().Preload("User").First(&review, review.ID)

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Review created successfully",
		Data:    review.ToResponse(),
	})
}

// UpdateReview cập nhật đánh giá (User can only update their own reviews)
func (rc *ReviewController) UpdateReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid review ID",
		})
		return
	}

	var req models.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := rc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Get user ID from JWT
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	// Get user role
	userRole, _ := middleware.GetUserRoleFromContext(c)

	var review models.Review
	if err := config.GetDB().First(&review, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Review not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Check if user owns this review or is admin
	if review.UserID != userID && userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, APIResponse{
			Success: false,
			Message: "You can only update your own reviews",
		})
		return
	}

	// Sanitize comment
	req.Comment = utils.SanitizeString(req.Comment)

	// Update review
	oldRating := review.Rating
	review.Comment = req.Comment
	review.Rating = req.Rating

	// Start transaction
	tx := config.GetDB().Begin()

	if err := tx.Save(&review).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update review",
			Error:   err.Error(),
		})
		return
	}

	// Update product rating statistics if rating changed
	if oldRating != req.Rating {
		if err := rc.updateProductRatingStats(tx, review.ProductID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to update product statistics",
				Error:   err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// Load user for response
	config.GetDB().Preload("User").First(&review, review.ID)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Review updated successfully",
		Data:    review.ToResponse(),
	})
}

// DeleteReview xóa đánh giá (User can only delete their own reviews, Admin can delete any)
func (rc *ReviewController) DeleteReview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid review ID",
		})
		return
	}

	// Get user ID and role from JWT
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	userRole, _ := middleware.GetUserRoleFromContext(c)

	var review models.Review
	if err := config.GetDB().First(&review, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Review not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Database error",
			Error:   err.Error(),
		})
		return
	}

	// Check if user owns this review or is admin
	if review.UserID != userID && userRole != models.RoleAdmin {
		c.JSON(http.StatusForbidden, APIResponse{
			Success: false,
			Message: "You can only delete your own reviews",
		})
		return
	}

	productID := review.ProductID

	// Start transaction
	tx := config.GetDB().Begin()

	if err := tx.Delete(&review).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete review",
			Error:   err.Error(),
		})
		return
	}

	// Update product rating statistics
	if err := rc.updateProductRatingStats(tx, productID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update product statistics",
			Error:   err.Error(),
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Review deleted successfully",
	})
}

// GetProductReviews lấy danh sách đánh giá của sản phẩm (Public)
func (rc *ReviewController) GetProductReviews(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	var params models.ReviewListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	// Set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 10
	}
	if params.Sort == "" {
		params.Sort = "newest"
	}

	offset := (params.Page - 1) * params.Limit

	query := config.GetDB().Model(&models.Review{}).
		Preload("User").
		Where("product_id = ?", productID)

	// Apply filters
	if params.Rating != nil {
		query = query.Where("rating = ?", *params.Rating)
	}

	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	} else {
		// Default: only show active reviews
		query = query.Where("is_active = ?", true)
	}

	// Apply sorting
	switch params.Sort {
	case "newest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	case "rating_high":
		query = query.Order("rating DESC, created_at DESC")
	case "rating_low":
		query = query.Order("rating ASC, created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get reviews
	var reviews []models.Review
	err = query.Limit(params.Limit).
		Offset(offset).
		Find(&reviews).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to fetch reviews",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response format
	var responses []models.ReviewResponse
	for _, review := range reviews {
		responses = append(responses, review.ToResponse())
	}

	// Get review statistics
	stats, err := rc.getReviewStats(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to get review statistics",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Reviews retrieved successfully",
		Data: gin.H{
			"reviews": responses,
			"pagination": gin.H{
				"page":        params.Page,
				"limit":       params.Limit,
				"total":       total,
				"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
			},
			"statistics": stats,
		},
	})
}

// GetUserReviews lấy danh sách đánh giá của user (Authenticated user)
func (rc *ReviewController) GetUserReviews(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "User ID not found",
		})
		return
	}

	var params models.ReviewListParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	// Set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 10
	}
	if params.Sort == "" {
		params.Sort = "newest"
	}

	offset := (params.Page - 1) * params.Limit

	query := config.GetDB().Model(&models.Review{}).
		Preload("Product").
		Preload("Product.Category").
		Where("user_id = ?", userID)

	// Apply filters
	if params.ProductID != nil {
		query = query.Where("product_id = ?", *params.ProductID)
	}

	if params.Rating != nil {
		query = query.Where("rating = ?", *params.Rating)
	}

	// Apply sorting
	switch params.Sort {
	case "newest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	case "rating_high":
		query = query.Order("rating DESC, created_at DESC")
	case "rating_low":
		query = query.Order("rating ASC, created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get reviews
	var reviews []models.Review
	err := query.Limit(params.Limit).
		Offset(offset).
		Find(&reviews).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to fetch user reviews",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response format with product info
	var responses []gin.H
	for _, review := range reviews {
		reviewResp := review.ToResponse()
		productResp := review.Product.ToResponse()

		responses = append(responses, gin.H{
			"id":         reviewResp.ID,
			"product":    productResp,
			"comment":    reviewResp.Comment,
			"rating":     reviewResp.Rating,
			"is_active":  reviewResp.IsActive,
			"created_at": reviewResp.CreatedAt,
			"updated_at": reviewResp.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User reviews retrieved successfully",
		Data: gin.H{
			"reviews": responses,
			"pagination": gin.H{
				"page":        params.Page,
				"limit":       params.Limit,
				"total":       total,
				"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
			},
		},
	})
}

// updateProductRatingStats cập nhật thống kê rating của sản phẩm
func (rc *ReviewController) updateProductRatingStats(tx *gorm.DB, productID uint) error {
	var stats struct {
		Count   int64   `json:"count"`
		Average float64 `json:"average"`
	}

	// Calculate new statistics
	err := tx.Model(&models.Review{}).
		Where("product_id = ? AND is_active = ?", productID, true).
		Select("COUNT(*) as count, AVG(rating) as average").
		Scan(&stats).Error

	if err != nil {
		return err
	}

	// Update product
	return tx.Model(&models.Product{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"review_count":   stats.Count,
			"average_rating": stats.Average,
		}).Error
}

// getReviewStats lấy thống kê đánh giá của sản phẩm
func (rc *ReviewController) getReviewStats(productID uint) (models.ReviewStats, error) {
	var stats models.ReviewStats

	// Get total reviews and average rating
	err := config.GetDB().Model(&models.Review{}).
		Where("product_id = ? AND is_active = ?", productID, true).
		Select("COUNT(*) as total_reviews, AVG(rating) as average_rating").
		Scan(&stats).Error

	if err != nil {
		return stats, err
	}

	// Get rating breakdown
	rows, err := config.GetDB().Model(&models.Review{}).
		Where("product_id = ? AND is_active = ?", productID, true).
		Select("rating, COUNT(*) as count").
		Group("rating").
		Order("rating DESC").
		Rows()

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	stats.RatingBreakdown = make(map[string]int64)
	for i := 1; i <= 5; i++ {
		stats.RatingBreakdown[strconv.Itoa(i)] = 0
	}

	for rows.Next() {
		var rating int
		var count int64
		if err := rows.Scan(&rating, &count); err != nil {
			continue
		}
		stats.RatingBreakdown[strconv.Itoa(rating)] = count
	}

	return stats, nil
}
