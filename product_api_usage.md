# Product API Usage Examples

## Authentication Headers
Đối với các endpoint cần authentication, sử dụng headers:
```bash
Authorization: Bearer <your-jwt-token>
X-CSRF-Token: <csrf-token-from-login>
```

---

## Categories API

### 1. Lấy danh sách categories (Public)
```bash
GET /api/v1/categories?is_active=true&search=&page=1&limit=20
```

**Response:**
```json
{
  "success": true,
  "message": "Categories retrieved successfully",
  "data": {
    "categories": [
      {
        "id": 1,
        "name": "Thời trang nam",
        "slug": "thoi-trang-nam",
        "description": "Quần áo thời trang dành cho nam giới",
        "thumbnail": "https://example.com/fashion-men.jpg",
        "is_active": true,
        "product_count": 25,
        "created_at": "2025-06-11T10:00:00Z",
        "updated_at": "2025-06-11T10:00:00Z"
      },
      {
        "id": 2,
        "name": "Điện tử",
        "slug": "dien-tu",
        "description": "Sản phẩm điện tử, công nghệ",
        "thumbnail": "https://example.com/electronics.jpg",
        "is_active": true,
        "product_count": 18,
        "created_at": "2025-06-11T10:00:00Z",
        "updated_at": "2025-06-11T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 2,
      "total_pages": 1
    }
  }
}
```

### 2. Tạo category mới (Admin only)
```bash
POST /api/v1/admin/categories
Content-Type: application/json
Authorization: Bearer <admin-token>
X-CSRF-Token: <csrf-token>

{
  "name": "Giày dép",
  "description": "Giày dép thời trang nam nữ",
  "thumbnail": "https://example.com/shoes.jpg",
  "is_active": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Category created successfully",
  "data": {
    "id": 3,
    "name": "Giày dép",
    "slug": "giay-dep",
    "description": "Giày dép thời trang nam nữ",
    "thumbnail": "https://example.com/shoes.jpg",
    "is_active": true,
    "product_count": 0,
    "created_at": "2025-06-11T15:30:00Z",
    "updated_at": "2025-06-11T15:30:00Z"
  }
}
```

### 3. Lấy category theo slug (Public)
```bash
GET /api/v1/categories/slug/thoi-trang-nam
```

### 4. Cập nhật category (Admin only)
```bash
PUT /api/v1/admin/categories/1
Content-Type: application/json
Authorization: Bearer <admin-token>
X-CSRF-Token: <csrf-token>

{
  "name": "Thời trang nam cao cấp",
  "description": "Thời trang nam chất lượng cao, thiết kế hiện đại",
  "thumbnail": "https://example.com/premium-fashion.jpg",
  "is_active": true
}
```

### 5. Xóa category (Admin only)
```bash
DELETE /api/v1/admin/categories/3
Authorization: Bearer <admin-token>
X-CSRF-Token: <csrf-token>
```

---

## Products API

### 1. Lấy danh sách sản phẩm với filter (Public)
```bash
GET /api/v1/products?category_id=1&search=áo&min_price=100000&max_price=500000&is_featured=true&sort=price_asc&page=1&limit=10&in_stock=true
```

**Response:**
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": {
    "products": [
      {
        "id": 1,
        "category_id": 1,
        "category": {
          "id": 1,
          "name": "Thời trang nam",
          "slug": "thoi-trang-nam"
        },
        "name": "Áo thun nam basic",
        "slug": "ao-thun-nam-basic",
        "price": 250000,
        "discount_price": 199000,
        "final_price": 199000,
        "discount_rate": 20.4,
        "description": "Áo thun nam cotton 100%, co giãn thoải mái",
        "total_stock": 85,
        "stock_status": "in_stock",
        "thumbnail": "https://example.com/ao-thun-basic.jpg",
        "images": [
          "https://example.com/ao-thun-1.jpg",
          "https://example.com/ao-thun-2.jpg",
          "https://example.com/ao-thun-3.jpg"
        ],
        "is_featured": true,
        "is_active": true,
        "view_count": 1250,
        "average_rating": 4.5,
        "review_count": 23,
        "sizes": [
          {
            "id": 1,
            "product_id": 1,
            "size": "S",
            "stock": 20,
            "price": null,
            "final_price": 199000,
            "is_active": true,
            "stock_status": "in_stock",
            "created_at": "2025-06-11T10:00:00Z",
            "updated_at": "2025-06-11T10:00:00Z"
          },
          {
            "id": 2,
            "product_id": 1,
            "size": "M",
            "stock": 30,
            "price": null,
            "final_price": 199000,
            "is_active": true,
            "stock_status": "in_stock",
            "created_at": "2025-06-11T10:00:00Z",
            "updated_at": "2025-06-11T10:00:00Z"
          },
          {
            "id": 3,
            "product_id": 1,
            "size": "L",
            "stock": 25,
            "price": 209000,
            "final_price": 209000,
            "is_active": true,
            "stock_status": "in_stock",
            "created_at": "2025-06-11T10:00:00Z",
            "updated_at": "2025-06-11T10:00:00Z"
          },
          {
            "id": 4,
            "product_id": 1,
            "size": "XL",
            "stock": 10,
            "price": 219000,
            "final_price": 219000,
            "is_active": true,
            "stock_status": "low_stock",
            "created_at": "2025-06-11T10:00:00Z",
            "updated_at": "2025-06-11T10:00:00Z"
          }
        ],
        "created_at": "2025-06-11T10:00:00Z",
        "updated_at": "2025-06-11T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "total_pages": 1
    },
    "filters": {
      "category_id": 1,
      "search": "áo",
      "min_price": 100000,
      "max_price": 500000,
      "is_featured": true,
      "in_stock": true,
      "sort": "price_asc"
    }
  }
}
```

### 2. Tạo sản phẩm mới với sizes (Admin only)
```bash
POST /api/v1/admin/products
Content-Type: application/json
Authorization: Bearer <admin-token>
X-CSRF-Token: <csrf-token>

{
  "category_id": 1,
  "name": "Áo polo nam cao cấp",
  "price": 350000,
  "discount_price": 299000,
  "description": "Áo polo nam chất liệu pique cao cấp, form regular fit",
  "thumbnail": "https://example.com/polo-thumbnail.jpg",
  "images": [
    "https://example.com/polo-1.jpg",
    "https://example.com/polo-2.jpg",
    "https://example.com/polo-3.jpg",
    "https://example.com/polo-4.jpg"
  ],
  "is_featured": true,
  "is_active": true,
  "sizes": [
    {
      "size": "S",
      "stock": 15,
      "price": null,
      "is_active": true
    },
    {
      "size": "M",
      "stock": 25,
      "price": null,
      "is_active": true
    },
    {
      "size": "L", 
      "stock": 20,
      "price": 309000,
      "is_active": true
    },
    {
      "size": "XL",
      "stock": 12,
      "price": 319000,
      "is_active": true
    },
    {
      "size": "XXL",
      "stock": 8,
      "price": 329000,
      "is_active": true
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": 2,
    "category_id": 1,
    "category": {
      "id": 1,
      "name": "Thời trang nam",
      "slug": "thoi-trang-nam"
    },
    "name": "Áo polo nam cao cấp",
    "slug": "ao-polo-nam-cao-cap",
    "price": 350000,
    "discount_price": 299000,
    "final_price": 299000,
    "discount_rate": 14.57,
    "description": "Áo polo nam chất liệu pique cao cấp, form regular fit",
    "total_stock": 80,
    "stock_status": "in_stock",
    "thumbnail": "https://example.com/polo-thumbnail.jpg",
    "images": [
      "https://example.com/polo-1.jpg",
      "https://example.com/polo-2.jpg",
      "https://example.com/polo-3.jpg",
      "https://example.com/polo-4.jpg"
    ],
    "is_featured": true,
    "is_active": true,
    "view_count": 0,
    "average_rating": 0,
    "review_count": 0,
    "sizes": [
      {
        "id": 5,
        "product_id": 2,
        "size": "S",
        "stock": 15,
        "price": null,
        "final_price": 299000,
        "is_active": true,
        "stock_status": "in_stock"
      },
      {
        "id": 6,
        "product_id": 2,
        "size": "M",
        "stock": 25,
        "price": null,
        "final_price": 299000,
        "is_active": true,
        "stock_status": "in_stock"
      },
      {
        "id": 7,
        "product_id": 2,
        "size": "L",
        "stock": 20,
        "price": 309000,
        "final_price": 309000,
        "is_active": true,
        "stock_status": "in_stock"
      },
      {
        "id": 8,
        "product_id": 2,
        "size": "XL",
        "stock": 12,
        "price": 319000,
        "final_price": 319000,
        "is_active": true,
        "stock_status": "in_stock"
      },
      {
        "id": 9,
        "product_id": 2,
        "size": "XXL",
        "stock": 8,
        "price": 329000,
        "final_price": 329000,
        "is_active": true,
        "stock_status": "low_stock"
      }
    ],
    "created_at": "2025-06-11T16:00:00Z",
    "updated_at": "2025-06-11T16:00:00Z"
  }
}
```

### 3. Cập nhật sản phẩm (Admin only)
```bash
PUT /api/v1/admin/products/2
Content-Type: application/json
Authorization: Bearer <admin-token>
X-CSRF-Token: <csrf-token>

{
  "category_id": 1,
  "name": "Áo polo nam premium",
  "price": 380000,
  "discount_price": 319000,
  "description": "Áo polo nam chất liệu pique premium, form slim fit hiện đại",
  "thumbnail": "https://example.com/polo-premium.jpg",
  "images": [
    "https://example.com/polo-premium-1.jpg",
    "https://example.com/polo-premium-2.jpg"
  ],
  "is_featured": true,
  "is_active": true,
  "sizes": [
    {
      "size": "S",
      "stock": 20,
      "price": null,
      "is_active": true
    },
    {
      "size": "M",
      "stock": 30,
      "price": null,
      "is_active": true
    },
    {
      "size": "L",
      "stock": 25,
      "price": 329000,
      "is_active": true
    },
    {
      "size": "XL",
      "stock": 15,
      "price": 339000,
      "is_active": true
    }
  ]
}
```

### 4. Lấy sản phẩm theo ID (Public)
```bash
GET /api/v1/products/1
```

### 5. Lấy sản phẩm theo slug (Public)
```bash
GET /api/v1/products/slug/ao-thun-nam-basic
```

### 6. Lấy sản phẩm nổi bật (Public)
```bash
GET /api/v1/products/featured?limit=8
```

### 7. Tìm kiếm sản phẩm (Public)
```bash
GET /api/v1/products/search?q=áo%20thun&page=1&limit=20
```

**Response:**
```json
{
  "success": true,
  "message": "Found 3 products for 'áo thun'",
  "data": {
    "products": [
      {
        "id": 1,
        "name": "Áo thun nam basic",
        "slug": "ao-thun-nam-basic",
        "price": 250000,
        "final_price": 199000,
        "total_stock": 85,
        "sizes": [...]
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 3,
      "total_pages": 1
    },
    "search_query": "áo thun"
  }
}
```

### 8. Xóa sản phẩm (Admin only)
```bash
DELETE /api/v1/admin/products/2
Authorization: Bearer <admin-token>
X-CSRF-Token: <csrf-token>
```

---

## Reviews API

### 1. Tạo đánh giá mới (Authenticated users only)
```bash
POST /api/v1/reviews
Content-Type: application/json
Authorization: Bearer <user-token>
X-CSRF-Token: <csrf-token>

{
  "product_id": 1,
  "comment": "Sản phẩm rất tốt, chất lượng vượt mong đợi. Áo mặc rất thoải mái, chất liệu cotton mềm mịn. Size M vừa vặn với tôi. Sẽ mua thêm các màu khác.",
  "rating": 5
}
```

**Response:**
```json
{
  "success": true,
  "message": "Review created successfully",
  "data": {
    "id": 1,
    "product_id": 1,
    "user_id": 2,
    "user": {
      "id": 2,
      "fullname": "Nguyễn Văn A",
      "email": "nguyenvana@example.com"
    },
    "comment": "Sản phẩm rất tốt, chất lượng vượt mong đợi. Áo mặc rất thoải mái, chất liệu cotton mềm mịn. Size M vừa vặn với tôi. Sẽ mua thêm các màu khác.",
    "rating": 5,
    "is_active": true,
    "created_at": "2025-06-11T17:30:00Z",
    "updated_at": "2025-06-11T17:30:00Z"
  }
}
```

### 2. Lấy đánh giá của sản phẩm (Public)
```bash
GET /api/v1/reviews/product/1?rating=5&sort=newest&page=1&limit=10
```

**Response:**
```json
{
  "success": true,
  "message": "Reviews retrieved successfully",
  "data": {
    "reviews": [
      {
        "id": 1,
        "product_id": 1,
        "user_id": 2,
        "user": {
          "id": 2,
          "fullname": "Nguyễn Văn A",
          "email": "nguyenvana@example.com"
        },
        "comment": "Sản phẩm rất tốt, chất lượng vượt mong đợi...",
        "rating": 5,
        "is_active": true,
        "created_at": "2025-06-11T17:30:00Z",
        "updated_at": "2025-06-11T17:30:00Z"
      },
      {
        "id": 2,
        "product_id": 1,
        "user_id": 3,
        "user": {
          "id": 3,
          "fullname": "Trần Thị B",
          "email": "tranthib@example.com"
        },
        "comment": "Đóng gói cẩn thận, giao hàng nhanh. Áo đẹp, phù hợp làm quà tặng.",
        "rating": 5,
        "is_active": true,
        "created_at": "2025-06-11T16:45:00Z",
        "updated_at": "2025-06-11T16:45:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 23,
      "total_pages": 3
    },
    "statistics": {
      "total_reviews": 23,
      "average_rating": 4.5,
      "rating_breakdown": {
        "5": 15,
        "4": 6,
        "3": 2,
        "2": 0,
        "1": 0
      }
    }
  }
}
```

### 3. Lấy đánh giá của user hiện tại (Authenticated)
```bash
GET /api/v1/reviews/my-reviews?page=1&limit=10&sort=newest
Authorization: Bearer <user-token>
X-CSRF-Token: <csrf-token>
```

**Response:**
```json
{
  "success": true,
  "message": "User reviews retrieved successfully",
  "data": {
    "reviews": [
      {
        "id": 1,
        "product": {
          "id": 1,
          "name": "Áo thun nam basic",
          "slug": "ao-thun-nam-basic",
          "price": 250000,
          "final_price": 199000,
          "thumbnail": "https://example.com/ao-thun-basic.jpg"
        },
        "comment": "Sản phẩm rất tốt, chất lượng vượt mong đợi...",
        "rating": 5,
        "is_active": true,
        "created_at": "2025-06-11T17:30:00Z",
        "updated_at": "2025-06-11T17:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

### 4. Cập nhật đánh giá (User own review or Admin)
```bash
PUT /api/v1/reviews/1
Content-Type: application/json
Authorization: Bearer <user-token>
X-CSRF-Token: <csrf-token>

{
  "product_id": 1,
  "comment": "Sau 2 tuần sử dụng, sản phẩm vẫn giữ được form dáng và màu sắc. Rất hài lòng với chất lượng. Khuyên mọi người nên mua.",
  "rating": 5
}
```

### 5. Xóa đánh giá (User own review or Admin)
```bash
DELETE /api/v1/reviews/1
Authorization: Bearer <user-token>
X-CSRF-Token: <csrf-token>
```

---

## Sort Options

### Products Sort
- `newest` - Mới nhất (mặc định)
- `oldest` - Cũ nhất  
- `price_asc` - Giá tăng dần
- `price_desc` - Giá giảm dần
- `name_asc` - Tên A-Z
- `name_desc` - Tên Z-A
- `rating` - Đánh giá cao nhất
- `popular` - Phổ biến nhất (lượt xem)

### Reviews Sort
- `newest` - Mới nhất (mặc định)
- `oldest` - Cũ nhất
- `rating_high` - Rating cao xuống thấp
- `rating_low` - Rating thấp lên cao

---

## Error Responses

### Validation Error
```json
{
  "success": false,
  "message": "Validation failed",
  "error": "Field 'name' is required"
}
```

### Product Not Found
```json
{
  "success": false,
  "message": "Product not found"
}
```

### Duplicate Size Error
```json
{
  "success": false,
  "message": "Duplicate size: M"
}
```

### Already Reviewed Error  
```json
{
  "success": false,
  "message": "You have already reviewed this product"
}
```

### Permission Denied
```json
{
  "success": false,
  "message": "You can only update your own reviews"
}
```

---

## Stock Status

- `in_stock` - Còn hàng (stock > 10)
- `low_stock` - Sắp hết hàng (stock 1-10)
- `out_of_stock` - Hết hàng (stock = 0)

---

## Notes

1. **Size Management**: Mỗi sản phẩm phải có ít nhất 1 size
2. **Price Override**: Size có thể có giá riêng khác với product price
3. **Total Stock**: Tự động tính từ tổng stock của tất cả sizes
4. **Unique Constraint**: Không được trùng size trong cùng 1 product
5. **Auto Migration**: Database tự động tạo constraint unique (product_id + size)
6. **View Count**: Tự động tăng khi user xem chi tiết sản phẩm
7. **Rating Stats**: Tự động cập nhật khi có review mới/sửa/xóa