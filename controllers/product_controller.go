package controllers

import (
	"fmt"
	"net/http"
	"storelite/config"
	"storelite/models"
	"storelite/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ProductController struct {
	validator *validator.Validate
}

func NewProductController() *ProductController {
	return &ProductController{
		validator: validator.New(),
	}
}

// CreateProduct tạo sản phẩm mới với sizes (Admin only)
func (pc *ProductController) CreateProduct(c *gin.Context) {
	var req models.ProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := pc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Validate category exists
	var category models.Category
	if err := config.GetDB().Where("id = ? AND is_active = ?", req.CategoryID, true).First(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Category not found or inactive",
		})
		return
	}

	// Validate discount price
	if req.DiscountPrice != nil && *req.DiscountPrice >= req.Price {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Discount price must be less than original price",
		})
		return
	}

	// Validate sizes
	if len(req.Sizes) == 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "At least one size is required",
		})
		return
	}

	// Check for duplicate sizes
	sizeMap := make(map[string]bool)
	for _, size := range req.Sizes {
		sizeKey := strings.ToUpper(strings.TrimSpace(size.Size))
		if sizeMap[sizeKey] {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Duplicate size: %s", size.Size),
			})
			return
		}
		sizeMap[sizeKey] = true
	}

	// Sanitize input
	req.Name = utils.SanitizeString(req.Name)
	req.Description = utils.SanitizeString(req.Description)

	// Generate slug
	slug := utils.GenerateSlug(req.Name)

	// Check if slug already exists
	var existingProduct models.Product
	if err := config.GetDB().Where("slug = ?", slug).First(&existingProduct).Error; err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Message: "Product with this name already exists",
		})
		return
	}

	// Set default values
	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Validate and sanitize images
	var validImages models.StringArray
	for _, img := range req.Images {
		sanitizedImg := utils.SanitizeString(img)
		if sanitizedImg != "" {
			validImages = append(validImages, sanitizedImg)
		}
	}

	// Calculate total stock
	totalStock := 0
	for _, size := range req.Sizes {
		totalStock += size.Stock
	}

	// Start transaction
	tx := config.GetDB().Begin()

	product := models.Product{
		CategoryID:    req.CategoryID,
		Name:          req.Name,
		Slug:          slug,
		Price:         req.Price,
		DiscountPrice: req.DiscountPrice,
		Description:   req.Description,
		TotalStock:    totalStock,
		Thumbnail:     req.Thumbnail,
		Images:        validImages,
		IsFeatured:    isFeatured,
		IsActive:      isActive,
	}

	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create product",
			Error:   err.Error(),
		})
		return
	}

	// Create product sizes
	for _, sizeReq := range req.Sizes {
		sizeIsActive := true
		if sizeReq.IsActive != nil {
			sizeIsActive = *sizeReq.IsActive
		}

		productSize := models.ProductSize{
			ProductID: product.ID,
			Size:      utils.SanitizeString(sizeReq.Size),
			Stock:     sizeReq.Stock,
			Price:     sizeReq.Price,
			IsActive:  sizeIsActive,
		}

		if err := tx.Create(&productSize).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to create product size",
				Error:   err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// Load complete product with relationships
	config.GetDB().Preload("Category").Preload("ProductSizes").First(&product, product.ID)

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "Product created successfully",
		Data:    product.ToResponse(),
	})
}

// UpdateProduct cập nhật sản phẩm với sizes (Admin only)
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := pc.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	var product models.Product
	if err := config.GetDB().First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Product not found",
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

	// Validate category exists
	var category models.Category
	if err := config.GetDB().Where("id = ? AND is_active = ?", req.CategoryID, true).First(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Category not found or inactive",
		})
		return
	}

	// Validate discount price
	if req.DiscountPrice != nil && *req.DiscountPrice >= req.Price {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Discount price must be less than original price",
		})
		return
	}

	// Validate sizes
	if len(req.Sizes) == 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "At least one size is required",
		})
		return
	}

	// Check for duplicate sizes
	sizeMap := make(map[string]bool)
	for _, size := range req.Sizes {
		sizeKey := strings.ToUpper(strings.TrimSpace(size.Size))
		if sizeMap[sizeKey] {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Duplicate size: %s", size.Size),
			})
			return
		}
		sizeMap[sizeKey] = true
	}

	// Sanitize input
	req.Name = utils.SanitizeString(req.Name)
	req.Description = utils.SanitizeString(req.Description)

	// Generate new slug if name changed
	newSlug := utils.GenerateSlug(req.Name)
	if newSlug != product.Slug {
		// Check if new slug already exists
		var existingProduct models.Product
		if err := config.GetDB().Where("slug = ? AND id != ?", newSlug, id).First(&existingProduct).Error; err == nil {
			c.JSON(http.StatusConflict, APIResponse{
				Success: false,
				Message: "Product with this name already exists",
			})
			return
		}
		product.Slug = newSlug
	}

	// Validate and sanitize images
	var validImages models.StringArray
	for _, img := range req.Images {
		sanitizedImg := utils.SanitizeString(img)
		if sanitizedImg != "" {
			validImages = append(validImages, sanitizedImg)
		}
	}

	// Calculate total stock
	totalStock := 0
	for _, size := range req.Sizes {
		totalStock += size.Stock
	}

	// Start transaction
	tx := config.GetDB().Begin()

	// Update product fields
	product.CategoryID = req.CategoryID
	product.Name = req.Name
	product.Price = req.Price
	product.DiscountPrice = req.DiscountPrice
	product.Description = req.Description
	product.TotalStock = totalStock
	product.Thumbnail = req.Thumbnail
	product.Images = validImages

	if req.IsFeatured != nil {
		product.IsFeatured = *req.IsFeatured
	}
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update product",
			Error:   err.Error(),
		})
		return
	}

	// Delete existing product sizes
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductSize{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete existing sizes",
			Error:   err.Error(),
		})
		return
	}

	// Create new product sizes
	for _, sizeReq := range req.Sizes {
		sizeIsActive := true
		if sizeReq.IsActive != nil {
			sizeIsActive = *sizeReq.IsActive
		}

		productSize := models.ProductSize{
			ProductID: product.ID,
			Size:      utils.SanitizeString(sizeReq.Size),
			Stock:     sizeReq.Stock,
			Price:     sizeReq.Price,
			IsActive:  sizeIsActive,
		}

		if err := tx.Create(&productSize).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, APIResponse{
				Success: false,
				Message: "Failed to create product size",
				Error:   err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// Load complete product with relationships
	config.GetDB().Preload("Category").Preload("ProductSizes").First(&product, product.ID)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Product updated successfully",
		Data:    product.ToResponse(),
	})
}

// DeleteProduct xóa sản phẩm (Admin only)
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	// Check if product has reviews
	var reviewCount int64
	config.GetDB().Model(&models.Review{}).Where("product_id = ?", id).Count(&reviewCount)

	if reviewCount > 0 {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Cannot delete product that has reviews",
			Data: gin.H{
				"review_count": reviewCount,
			},
		})
		return
	}

	if err := config.GetDB().Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to delete product",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Product deleted successfully",
	})
}

// GetProducts lấy danh sách sản phẩm với filter và search (Public)
func (pc *ProductController) GetProducts(c *gin.Context) {
	var params models.ProductListParams

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
		params.Limit = 20
	}
	if params.Sort == "" {
		params.Sort = "newest"
	}

	offset := (params.Page - 1) * params.Limit

	query := config.GetDB().Model(&models.Product{}).
		Preload("Category").
		Preload("ProductSizes", "is_active = ?", true)

	// Apply filters
	if params.CategoryID != nil {
		query = query.Where("category_id = ?", *params.CategoryID)
	}

	if params.Search != "" {
		searchTerm := "%" + utils.SanitizeString(params.Search) + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	if params.MinPrice != nil {
		query = query.Where("price >= ?", *params.MinPrice)
	}

	if params.MaxPrice != nil {
		query = query.Where("price <= ?", *params.MaxPrice)
	}

	if params.IsFeatured != nil {
		query = query.Where("is_featured = ?", *params.IsFeatured)
	}

	if params.IsActive != nil {
		query = query.Where("is_active = ?", *params.IsActive)
	} else {
		// Default: only show active products for public
		query = query.Where("is_active = ?", true)
	}

	if params.InStock != nil && *params.InStock {
		query = query.Where("total_stock > 0")
	}

	// Apply sorting
	switch params.Sort {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	case "name_asc":
		query = query.Order("name ASC")
	case "name_desc":
		query = query.Order("name DESC")
	case "newest":
		query = query.Order("created_at DESC")
	case "oldest":
		query = query.Order("created_at ASC")
	case "rating":
		query = query.Order("average_rating DESC")
	case "popular":
		query = query.Order("view_count DESC")
	default:
		query = query.Order("created_at ASC")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get products
	var products []models.Product
	err := query.Limit(params.Limit).
		Offset(offset).
		Find(&products).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to fetch products",
			Error:   err.Error(),
		})
		return
	}

	// Convert to response format
	var responses []models.ProductResponse
	for _, product := range products {
		responses = append(responses, product.ToResponse())
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Products retrieved successfully",
		Data: gin.H{
			"products": responses,
			"pagination": gin.H{
				"page":        params.Page,
				"limit":       params.Limit,
				"total":       total,
				"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
			},
			"filters": gin.H{
				"category_id": params.CategoryID,
				"search":      params.Search,
				"min_price":   params.MinPrice,
				"max_price":   params.MaxPrice,
				"is_featured": params.IsFeatured,
				"in_stock":    params.InStock,
				"sort":        params.Sort,
			},
		},
	})
}

// GetProductByID lấy thông tin sản phẩm theo ID (Public)
func (pc *ProductController) GetProductByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid product ID",
		})
		return
	}

	var product models.Product
	if err := config.GetDB().
		Preload("Category").
		Preload("ProductSizes", "is_active = ?", true).
		First(&product, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Product not found",
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

	// Increment view count
	go func() {
		config.GetDB().Model(&product).Update("view_count", gorm.Expr("view_count + 1"))
	}()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Product retrieved successfully",
		Data:    product.ToResponse(),
	})
}

// GetProductBySlug lấy thông tin sản phẩm theo slug (Public)
func (pc *ProductController) GetProductBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var product models.Product
	if err := config.GetDB().
		Preload("Category").
		Preload("ProductSizes", "is_active = ?", true).
		Where("slug = ?", slug).
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "Product not found",
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

	// Increment view count
	go func() {
		config.GetDB().Model(&product).Update("view_count", gorm.Expr("view_count + 1"))
	}()

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Product retrieved successfully",
		Data:    product.ToResponse(),
	})
}

// GetFeaturedProducts lấy danh sách sản phẩm nổi bật (Public)
func (pc *ProductController) GetFeaturedProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var products []models.Product
	err := config.GetDB().
		Preload("Category").
		Preload("ProductSizes", "is_active = ?", true).
		Where("is_featured = ? AND is_active = ?", true, true).
		Order("created_at DESC").
		Limit(limit).
		Find(&products).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to fetch featured products",
			Error:   err.Error(),
		})
		return
	}

	var responses []models.ProductResponse
	for _, product := range products {
		responses = append(responses, product.ToResponse())
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Featured products retrieved successfully",
		Data:    responses,
	})
}

// SearchProducts tìm kiếm sản phẩm nâng cao (Public)
func (pc *ProductController) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Search query is required",
		})
		return
	}

	// Sanitize search query
	query = utils.SanitizeString(query)
	searchTerms := strings.Split(strings.ToLower(query), " ")

	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build search query
	db := config.GetDB().Model(&models.Product{}).Preload("Category")

	// Search in name, description
	searchQuery := "("
	args := []interface{}{}

	for i, term := range searchTerms {
		if i > 0 {
			searchQuery += " AND "
		}
		searchQuery += "(name ILIKE ? OR description ILIKE ?)"
		args = append(args, "%"+term+"%", "%"+term+"%")
	}
	searchQuery += ")"

	db = db.Where(searchQuery, args...)
	db = db.Where("is_active = ?", true)

	// Get total count
	var total int64
	db.Count(&total)

	// Get products with relevance scoring
	var products []models.Product
	err := db.Preload("Category").
		Preload("ProductSizes", "is_active = ?", true).
		Order("view_count DESC, average_rating DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&products).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to search products",
			Error:   err.Error(),
		})
		return
	}

	var responses []models.ProductResponse
	for _, product := range products {
		responses = append(responses, product.ToResponse())
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d products for '%s'", total, query),
		Data: gin.H{
			"products": responses,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": (total + int64(limit) - 1) / int64(limit),
			},
			"search_query": query,
		},
	})
}
