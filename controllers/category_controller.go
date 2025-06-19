package controllers

import (
	"log"
	"net/http"
	"storelite/config"
	"storelite/models"
	"storelite/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type CategoryController struct {
	validator *validator.Validate
}

func NewCategoryController() *CategoryController {
	return &CategoryController{
		validator: validator.New(),
	}
}

// CreateCategory tạo danh mục mới (Admin only)
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var req models.CategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Sanitize input
	req.Name = utils.SanitizeString(req.Name)
	req.Description = utils.SanitizeString(req.Description)

	// Generate slug
	slug := utils.GenerateSlug(req.Name)

	// Check if slug already exists
	var existingCategory models.Category
	if err := config.GetDB().Where("slug = ?", slug).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Message: "Category with this name already exists",
		})
		return
	}

	// Set default value for IsActive if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	category := models.Category{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Thumbnail:   req.Thumbnail,
		IsActive:    isActive,
	}

	if err := config.GetDB().Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Category created successfully",
		Data:    category.ToResponse(),
	})
}

// UpdateCategory cập nhật danh mục (Admin only)
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid category ID",
		})
		return
	}

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := cc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	var category models.Category
	if err := config.GetDB().First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Category not found",
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

	// Sanitize input
	req.Name = utils.SanitizeString(req.Name)
	req.Description = utils.SanitizeString(req.Description)

	// Generate new slug if name changed
	newSlug := utils.GenerateSlug(req.Name)
	if newSlug != category.Slug {
		// Check if new slug already exists
		var existingCategory models.Category
		if err := config.GetDB().Where("slug = ? AND id != ?", newSlug, id).First(&existingCategory).Error; err == nil {
			c.JSON(http.StatusConflict, APIResponse{
				Success: false,
				Message: "Category with this name already exists",
			})
			return
		}
		category.Slug = newSlug
	}

	// Update fields
	category.Name = req.Name
	category.Description = req.Description
	category.Thumbnail = req.Thumbnail
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := config.GetDB().Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Category updated successfully",
		Data:    category.ToResponse(),
	})
}

// DeleteCategory xóa danh mục (Admin only)
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid category ID",
		})
		return
	}

	// Check if category exists
	var category models.Category
	if err := config.GetDB().First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Category not found",
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

	// Check if category has products
	var productCount int64
	config.GetDB().Model(&models.Product{}).Where("category_id = ?", id).Count(&productCount)
	log.Printf("Next 1")
	if productCount > 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Cannot delete category that has products",
			Data: gin.H{
				"product_count": productCount,
			},
		})
		return
	}

	if err := config.GetDB().Delete(&models.Category{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete category",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Category deleted successfully",
	})
}

// GetCategories lấy danh sách danh mục (Public)
func (cc *CategoryController) GetCategories(c *gin.Context) {
	isActive := c.Query("is_active")
	search := c.Query("search")

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	query := config.GetDB().Model(&models.Category{})

	// Apply filters
	if isActive != "" {
		if isActive == "true" {
			query = query.Where("is_active = ?", true)
		} else if isActive == "false" {
			query = query.Where("is_active = ?", false)
		}
	}

	if search != "" {
		searchTerm := "%" + utils.SanitizeString(search) + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get categories with product count
	var categories []models.Category
	err := query.Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Find(&categories).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to fetch categories",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response format and add product count
	var responses []models.CategoryResponse
	for _, category := range categories {
		response := category.ToResponse()

		// Count active products in category
		var productCount int64
		config.GetDB().Model(&models.Product{}).
			Where("category_id = ? AND is_active = ?", category.ID, true).
			Count(&productCount)
		response.ProductCount = productCount

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Categories retrieved successfully",
		Data: gin.H{
			"categories": responses,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// GetCategoryByID lấy thông tin danh mục theo ID (Public)
func (cc *CategoryController) GetCategoryByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid category ID",
		})
		return
	}

	var category models.Category
	if err := config.GetDB().First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Category not found",
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

	response := category.ToResponse()

	// Count active products in category
	var productCount int64
	config.GetDB().Model(&models.Product{}).
		Where("category_id = ? AND is_active = ?", category.ID, true).
		Count(&productCount)
	response.ProductCount = productCount

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Category retrieved successfully",
		Data:    response,
	})
}

// GetCategoryBySlug lấy thông tin danh mục theo slug (Public)
func (cc *CategoryController) GetCategoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var category models.Category
	if err := config.GetDB().Where("slug = ?", slug).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Category not found",
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

	response := category.ToResponse()

	// Count active products in category
	var productCount int64
	config.GetDB().Model(&models.Product{}).
		Where("category_id = ? AND is_active = ?", category.ID, true).
		Count(&productCount)
	response.ProductCount = productCount

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Category retrieved successfully",
		Data:    response,
	})
}
