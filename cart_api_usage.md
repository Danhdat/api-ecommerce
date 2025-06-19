# Cart API Usage Examples

## Overview
Cart API hỗ trợ cả **Guest Users** (chưa đăng nhập) và **Logged-in Users** (đã đăng nhập). Hệ thống tự động quản lý giỏ hàng thông qua Session ID cho guest và User ID cho users đã đăng nhập.

---

## Authentication & Session Management

### For Guest Users (Chưa đăng nhập)
```bash
# Sử dụng X-Session-ID header
X-Session-ID: <uuid-session-id>
```

### For Logged-in Users (Đã đăng nhập)  
```bash
# Sử dụng JWT authentication
Authorization: Bearer <your-jwt-token>
X-CSRF-Token: <csrf-token>
```

### Auto Session ID Generation
Nếu guest user không có Session ID, server sẽ tự động tạo và trả về trong response header:
```bash
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

## Cart API Endpoints

### 1. Lấy thông tin giỏ hàng

#### Guest User
```bash
GET /api/v1/cart
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

#### Logged-in User
```bash
GET /api/v1/cart
Authorization: Bearer <jwt-token>
X-CSRF-Token: <csrf-token>
```

**Response:**
```json
{
  "success": true,
  "message": "Cart retrieved successfully",
  "data": {
    "id": 1,
    "user_id": null,
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "item_count": 2,
    "total_quantity": 3,
    "subtotal": 578000,
    "discount": 0,
    "total": 578000,
    "is_valid": true,
    "issues": [],
    "expires_at": "2025-06-18T16:30:00Z",
    "cart_items": [
      {
        "id": 1,
        "product_id": 1,
        "product_size_id": 2,
        "product": {
          "id": 1,
          "name": "Áo thun nam basic",
          "slug": "ao-thun-nam-basic",
          "price": 250000,
          "discount_price": 199000,
          "final_price": 199000,
          "thumbnail": "https://example.com/ao-thun.jpg",
          "category": {
            "id": 1,
            "name": "Thời trang nam"
          }
        },
        "product_size": {
          "id": 2,
          "size": "M",
          "stock": 30,
          "price": null,
          "final_price": 199000,
          "stock_status": "in_stock"
        },
        "quantity": 2,
        "price": 199000,
        "current_price": 199000,
        "subtotal": 398000,
        "is_available": true,
        "stock_status": "available",
        "message": "",
        "created_at": "2025-06-11T16:30:00Z",
        "updated_at": "2025-06-11T16:30:00Z"
      },
      {
        "id": 2,
        "product_id": 2,
        "product_size_id": 5,
        "product": {
          "id": 2,
          "name": "Áo polo nam cao cấp",
          "slug": "ao-polo-nam-cao-cap",
          "price": 350000,
          "discount_price": 299000,
          "final_price": 299000,
          "thumbnail": "https://example.com/polo.jpg"
        },
        "product_size": {
          "id": 5,
          "size": "L",
          "stock": 15,
          "price": 180000,
          "final_price": 180000,
          "stock_status": "in_stock"
        },
        "quantity": 1,
        "price": 180000,
        "current_price": 180000,
        "subtotal": 180000,
        "is_available": true,
        "stock_status": "available",
        "created_at": "2025-06-11T17:00:00Z",
        "updated_at": "2025-06-11T17:00:00Z"
      }
    ],
    "created_at": "2025-06-11T16:30:00Z",
    "updated_at": "2025-06-11T17:00:00Z"
  }
}
```

### 2. Thêm sản phẩm vào giỏ hàng

#### Guest User
```bash
POST /api/v1/cart/add
Content-Type: application/json
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000

{
  "product_id": 1,
  "product_size_id": 2,
  "quantity": 2
}
```

#### Logged-in User
```bash
POST /api/v1/cart/add
Content-Type: application/json
Authorization: Bearer <jwt-token>
X-CSRF-Token: <csrf-token>

{
  "product_id": 1,
  "product_size_id": 2,
  "quantity": 2
}
```

**Response thành công:**
```json
{
  "success": true,
  "message": "Item added to cart successfully",
  "data": {
    "id": 1,
    "item_count": 1,
    "total_quantity": 2,
    "subtotal": 398000,
    "total": 398000,
    "cart_items": [
      {
        "id": 1,
        "product_id": 1,
        "product_size_id": 2,
        "quantity": 2,
        "price": 199000,
        "subtotal": 398000,
        "is_available": true
      }
    ]
  }
}
```

**Response lỗi - Không đủ stock:**
```json
{
  "success": false,
  "message": "Failed to add item to cart",
  "error": "insufficient stock. Available: 1"
}
```

**Response lỗi - Sản phẩm không hoạt động:**
```json
{
  "success": false,
  "message": "Failed to add item to cart",
  "error": "product is not active"
}
```

### 3. Cập nhật số lượng sản phẩm trong giỏ

```bash
PUT /api/v1/cart/items/1
Content-Type: application/json
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000

{
  "quantity": 3
}
```

**Response:**
```json
{
  "success": true,
  "message": "Cart item updated successfully",
  "data": {
    "id": 1,
    "item_count": 1,
    "total_quantity": 3,
    "subtotal": 597000,
    "total": 597000,
    "cart_items": [
      {
        "id": 1,
        "quantity": 3,
        "subtotal": 597000
      }
    ]
  }
}
```

**Xóa sản phẩm bằng cách set quantity = 0:**
```bash
PUT /api/v1/cart/items/1
Content-Type: application/json

{
  "quantity": 0
}
```

### 4. Xóa sản phẩm khỏi giỏ hàng

```bash
DELETE /api/v1/cart/items/1
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "message": "Item removed from cart successfully",
  "data": {
    "id": 1,
    "item_count": 0,
    "total_quantity": 0,
    "subtotal": 0,
    "total": 0,
    "cart_items": []
  }
}
```

### 5. Xóa tất cả sản phẩm trong giỏ

```bash
DELETE /api/v1/cart/clear
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "message": "Cart cleared successfully",
  "data": {
    "id": 1,
    "item_count": 0,
    "total_quantity": 0,
    "subtotal": 0,
    "total": 0,
    "cart_items": []
  }
}
```

### 6. Validate giỏ hàng trước khi checkout

```bash
GET /api/v1/cart/validate
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response hợp lệ:**
```json
{
  "success": true,
  "message": "Cart is valid for checkout",
  "data": {
    "is_valid": true,
    "cart": {
      "id": 1,
      "item_count": 2,
      "total": 578000,
      "is_valid": true,
      "issues": []
    }
  }
}
```

**Response không hợp lệ:**
```json
{
  "success": false,
  "message": "Cart validation failed",
  "data": {
    "is_valid": false,
    "issues": [
      "Áo thun nam basic - Size M: Chỉ còn 1 sản phẩm (yêu cầu 2)",
      "Áo polo nam cao cấp - Size L không còn hoạt động"
    ],
    "cart": {
      "id": 1,
      "item_count": 2,
      "is_valid": false,
      "cart_items": [
        {
          "id": 1,
          "is_available": false,
          "stock_status": "insufficient_stock",
          "message": "Chỉ còn 1 sản phẩm"
        },
        {
          "id": 2,
          "is_available": false,
          "stock_status": "size_inactive",
          "message": "Size không còn hoạt động"
        }
      ]
    }
  }
}
```

### 7. Lấy số lượng items trong giỏ

```bash
GET /api/v1/cart/count
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "message": "Cart count retrieved successfully",
  "data": {
    "item_count": 2,
    "total_quantity": 3,
    "total": 578000
  }
}
```

### 8. Merge guest cart khi đăng nhập

```bash
POST /api/v1/cart/merge
Content-Type: application/json
Authorization: Bearer <jwt-token>
X-CSRF-Token: <csrf-token>
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "message": "Guest cart merged successfully",
  "data": {
    "id": 2,
    "user_id": 123,
    "session_id": "",
    "item_count": 3,
    "total_quantity": 5,
    "subtotal": 897000,
    "total": 897000,
    "cart_items": [
      {
        "id": 3,
        "product_id": 1,
        "quantity": 3,
        "subtotal": 597000
      },
      {
        "id": 4,
        "product_id": 2,
        "quantity": 2,
        "subtotal": 300000
      }
    ]
  }
}
```

---

## Workflow Examples

### Workflow 1: Guest User Shopping

```bash
# 1. Lấy giỏ hàng (tự động tạo session)
GET /api/v1/cart
# Response header: X-Session-ID: new-uuid

# 2. Thêm sản phẩm vào giỏ
POST /api/v1/cart/add
X-Session-ID: new-uuid
{
  "product_id": 1,
  "product_size_id": 2,
  "quantity": 2
}

# 3. Cập nhật số lượng
PUT /api/v1/cart/items/1
X-Session-ID: new-uuid
{
  "quantity": 3
}

# 4. Validate trước checkout
GET /api/v1/cart/validate
X-Session-ID: new-uuid

# 5. Proceed to checkout...
```

### Workflow 2: Guest User Login & Merge Cart

```bash
# 1. Guest user có cart với session ID
GET /api/v1/cart
X-Session-ID: guest-session-uuid

# 2. User đăng nhập
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "password"
}

# 3. Merge guest cart vào user cart
POST /api/v1/cart/merge
Authorization: Bearer jwt-token
X-CSRF-Token: csrf-token
X-Session-ID: guest-session-uuid

# 4. Tiếp tục với user cart
GET /api/v1/cart
Authorization: Bearer jwt-token
```

### Workflow 3: Logged-in User Shopping

```bash
# 1. Lấy giỏ hàng user
GET /api/v1/cart
Authorization: Bearer jwt-token
X-CSRF-Token: csrf-token

# 2. Thêm sản phẩm
POST /api/v1/cart/add
Authorization: Bearer jwt-token
X-CSRF-Token: csrf-token
{
  "product_id": 2,
  "product_size_id": 5,
  "quantity": 1
}

# 3. Validate và checkout
GET /api/v1/cart/validate
Authorization: Bearer jwt-token
```

---

## Error Responses

### Validation Errors

**Invalid product:**
```json
{
  "success": false,
  "message": "Failed to add item to cart",
  "error": "product not found"
}
```

**Invalid size:**
```json
{
  "success": false,
  "message": "Failed to add item to cart", 
  "error": "product size not found"
}
```

**Insufficient stock:**
```json
{
  "success": false,
  "message": "Failed to add item to cart",
  "error": "insufficient stock. Available: 5"
}
```

**Quantity limit exceeded:**
```json
{
  "success": false,
  "message": "Failed to add item to cart",
  "error": "item quantity limit exceeded (100 items)"
}
```

**Cart items limit:**
```json
{
  "success": false,
  "message": "Failed to add item to cart",
  "error": "cart items limit exceeded (50 items)"
}
```

### Cart Item Not Found
```json
{
  "success": false,
  "message": "Failed to update cart item",
  "error": "cart item not found"
}
```

### Expired Cart
```json
{
  "success": false,
  "message": "Cart validation failed",
  "data": {
    "is_valid": false,
    "issues": ["Giỏ hàng đã hết hạn"]
  }
}
```

---

## Stock Status Types

| Status | Description |
|--------|-------------|
| `available` | Sản phẩm có sẵn, đủ stock |
| `product_inactive` | Sản phẩm không còn hoạt động |
| `size_inactive` | Size không còn hoạt động |
| `out_of_stock` | Hết hàng |
| `insufficient_stock` | Không đủ stock theo quantity yêu cầu |

---

## Price Logic

### Price Priority (cao xuống thấp):
1. **Size-specific price** (ProductSize.Price)
2. **Product discount price** (Product.DiscountPrice) 
3. **Product regular price** (Product.Price)

### Price Tracking:
- **price**: Giá tại thời điểm add vào cart
- **current_price**: Giá hiện tại của sản phẩm
- **final_price**: Giá cuối cùng tính theo logic priority

---

## Cart Expiry

### Expiry Time:
- **Guest Cart**: 24 giờ
- **User Cart**: 7 ngày (168 giờ)

### Auto Cleanup:
- Hệ thống tự động dọn dẹp expired carts mỗi 6 giờ
- Khi cart expired, tất cả items sẽ bị xóa
- Cart sẽ được reset với thời gian expiry mới

---

## Best Practices

### 1. Session Management
```javascript
// Frontend nên lưu Session ID trong localStorage/sessionStorage
const sessionId = localStorage.getItem('cart_session_id');
if (!sessionId) {
  // Gọi API để lấy session ID mới
  // Lưu session ID từ response header
}
```

### 2. Real-time Validation
```javascript
// Validate cart trước khi redirect đến checkout
const validateResponse = await fetch('/api/v1/cart/validate');
if (!validateResponse.data.is_valid) {
  // Hiển thị issues cho user
  // Cho phép user fix các vấn đề
}
```

### 3. Cart Merge Flow
```javascript
// Sau khi user login thành công
if (guestSessionId) {
  await fetch('/api/v1/cart/merge', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${jwtToken}`,
      'X-CSRF-Token': csrfToken,
      'X-Session-ID': guestSessionId
    }
  });
  // Clear guest session
  localStorage.removeItem('cart_session_id');
}
```

### 4. Error Handling
```javascript
// Handle stock errors gracefully
try {
  await addToCart(productId, sizeId, quantity);
} catch (error) {
  if (error.message.includes('insufficient stock')) {
    // Show available stock to user
    // Suggest reducing quantity
  }
}
```

---

## Notes

1. **Session ID**: Được tự động tạo nếu không có, trả về trong header
2. **Merge Logic**: Items trùng sẽ được cộng quantity khi merge
3. **Price Updates**: Cart hiển thị cả giá cũ và giá mới để user biết
4. **Stock Changes**: Real-time check stock khi validate
5. **Concurrent Access**: Database constraints đảm bảo consistency
6. **Performance**: Auto cleanup expired carts để optimize database