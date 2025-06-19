# Order API Usage Examples

## Overview
Order API hỗ trợ cả **Guest Users** (chưa đăng nhập) và **Logged-in Users** (đã đăng nhập). Hệ thống tự động tạo đơn hàng từ giỏ hàng, quản lý thanh toán và xử lý các trạng thái đơn hàng.

---

## Authentication & Session Management

### For Guest Users (Chưa đăng nhập)
```bash
# Sử dụng X-Session-ID header (same session từ cart)
X-Session-ID: <uuid-session-id>
```

### For Logged-in Users (Đã đăng nhập)  
```bash
# Sử dụng JWT authentication
Authorization: Bearer <your-jwt-token>
X-CSRF-Token: <csrf-token>
```

---

## Order Status & Payment Status

### Order Status
- `pending` - Đang chờ xử lý
- `paid` - Đã thanh toán
- `shipped` - Đang giao hàng
- `delivered` - Đã giao hàng
- `cancelled` - Đã hủy

### Payment Status
- `unpaid` - Chưa thanh toán
- `paid` - Đã thanh toán
- `failed` - Thanh toán thất bại

### Payment Methods
- `cod` - Thanh toán khi nhận hàng
- `bank_transfer` - Chuyển khoản ngân hàng
- `momo` - Ví MoMo (sắp tích hợp)
- `zalopay` - ZaloPay (sắp tích hợp)
- `vnpay` - VNPay (sắp tích hợp)

---

## Order API Endpoints

### 1. Lấy danh sách phương thức thanh toán

```bash
GET /api/v1/orders/payment-methods
```

**Response:**
```json
{
  "success": true,
  "message": "Payment methods retrieved successfully",
  "data": [
    {
      "method": "cod",
      "method_text": "Thanh toán khi nhận hàng",
      "description": "Thanh toán bằng tiền mặt khi nhận hàng",
      "is_available": true
    },
    {
      "method": "bank_transfer",
      "method_text": "Chuyển khoản ngân hàng",
      "description": "Chuyển khoản qua ngân hàng với QR Code",
      "is_available": true,
      "extra": {
        "bank_info": {
          "bank_name": "Ngân hàng TMCP Á Châu (ACB)",
          "account_number": "1234567890",
          "account_name": "CONG TY TNHH E-COMMERCE"
        }
      }
    },
    {
      "method": "momo",
      "method_text": "Ví MoMo",
      "description": "Thanh toán qua ví điện tử MoMo",
      "is_available": false,
      "extra": {
        "note": "Tính năng sẽ được tích hợp sớm"
      }
    },
    {
      "method": "zalopay",
      "method_text": "ZaloPay",
      "description": "Thanh toán qua ví điện tử ZaloPay",
      "is_available": false,
      "extra": {
        "note": "Tính năng sẽ được tích hợp sớm"
      }
    },
    {
      "method": "vnpay",
      "method_text": "VNPay",
      "description": "Thanh toán qua cổng VNPay",
      "is_available": false,
      "extra": {
        "note": "Tính năng sẽ được tích hợp sớm"
      }
    }
  ]
}
```

### 2. Tính toán tóm tắt đơn hàng

```bash
POST /api/v1/orders/calculate
Content-Type: application/json
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000

{
  "customer_name": "Nguyễn Văn A",
  "customer_email": "nguyenvana@example.com",
  "customer_phone": "0901234567",
  "shipping_address": {
    "full_name": "Nguyễn Văn A",
    "phone": "0901234567",
    "address_line": "123 Đường ABC, Phường XYZ",
    "city": "Hồ Chí Minh",
    "district": "Quận 1",
    "ward": "Phường Bến Nghé",
    "postal_code": "70000"
  },
  "payment_method": "cod",
  "notes": "Giao hàng vào buổi chiều"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Order summary calculated successfully",
  "data": {
    "item_count": 2,
    "total_quantity": 3,
    "total_amount": 578000,
    "shipping_fee": 30000,
    "discount_amount": 0,
    "final_amount": 608000
  }
}
```

### 3. Tạo đơn hàng mới

#### Guest User
```bash
POST /api/v1/orders
Content-Type: application/json
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000

{
  "customer_name": "Nguyễn Văn A",
  "customer_email": "nguyenvana@example.com",
  "customer_phone": "0901234567",
  "shipping_address": {
    "full_name": "Nguyễn Văn A",
    "phone": "0901234567",
    "address_line": "123 Đường ABC, Phường XYZ",
    "city": "Hồ Chí Minh",
    "district": "Quận 1",
    "ward": "Phường Bến Nghé",
    "postal_code": "70000"
  },
  "payment_method": "cod",
  "notes": "Giao hàng vào buổi chiều"
}
```

#### Logged-in User
```bash
POST /api/v1/orders
Content-Type: application/json
Authorization: Bearer <jwt-token>
X-CSRF-Token: <csrf-token>

{
  "customer_name": "Nguyễn Văn A",
  "customer_email": "nguyenvana@example.com",
  "customer_phone": "0901234567",
  "shipping_address": {
    "full_name": "Nguyễn Văn A",
    "phone": "0901234567",
    "address_line": "123 Đường ABC, Phường XYZ",
    "city": "Hồ Chí Minh",
    "district": "Quận 1",
    "ward": "Phường Bến Nghé",
    "postal_code": "70000"
  },
  "payment_method": "bank_transfer",
  "notes": ""
}
```

**Response thành công:**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "order": {
      "id": 1,
      "user_id": null,
      "order_code": "00001",
      "customer_name": "Nguyễn Văn A",
      "customer_email": "nguyenvana@example.com",
      "customer_phone": "0901234567",
      "shipping_address": {
        "full_name": "Nguyễn Văn A",
        "phone": "0901234567",
        "address_line": "123 Đường ABC, Phường XYZ",
        "city": "Hồ Chí Minh",
        "district": "Quận 1",
        "ward": "Phường Bến Nghé",
        "postal_code": "70000"
      },
      "total_amount": 578000,
      "shipping_fee": 30000,
      "discount_amount": 0,
      "final_amount": 608000,
      "status": "pending",
      "status_text": "Đang chờ xử lý",
      "payment_method": "cod",
      "payment_method_text": "Thanh toán khi nhận hàng",
      "payment_status": "unpaid",
      "payment_status_text": "Chưa thanh toán",
      "notes": "Giao hàng vào buổi chiều",
      "expires_at": "2025-06-11T19:00:00Z",
      "order_items": [
        {
          "id": 1,
          "product_id": 1,
          "product_size_id": 2,
          "product_name": "Áo thun nam basic",
          "product_size": "M",
          "quantity": 2,
          "unit_price": 199000,
          "total_price": 398000
        },
        {
          "id": 2,
          "product_id": 2,
          "product_size_id": 5,
          "product_name": "Áo polo nam cao cấp",
          "product_size": "L",
          "quantity": 1,
          "unit_price": 180000,
          "total_price": 180000
        }
      ],
      "created_at": "2025-06-11T18:30:00Z",
      "updated_at": "2025-06-11T18:30:00Z"
    },
    "payment_instructions": {
      "message": "Thanh toán khi nhận hàng. Vui lòng chuẩn bị đúng số tiền khi nhận hàng."
    }
  }
}
```

**Response cho Bank Transfer:**
```json
{
  "success": true,
  "message": "Order created successfully",
  "data": {
    "order": {
      "order_code": "00002",
      "payment_method": "bank_transfer",
      "final_amount": 608000
    },
    "payment_instructions": {
      "bank_name": "Ngân hàng TMCP Á Châu (ACB)",
      "account_number": "1234567890",
      "account_name": "CONG TY TNHH E-COMMERCE",
      "amount": 608000,
      "transfer_note": "DH 00002",
      "qr_code_url": "https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=Bank:1234567890,Amount:608000,Note:DH 00002"
    }
  }
}
```

**Response lỗi - Cart trống:**
```json
{
  "success": false,
  "message": "Failed to create order",
  "error": "cart validation failed: [Giỏ hàng trống]"
}
```

**Response lỗi - Không đủ stock:**
```json
{
  "success": false,
  "message": "Failed to create order",
  "error": "insufficient stock for Áo thun nam basic - M. Available: 1, Required: 2"
}
```

### 4. Xem thông tin đơn hàng

```bash
GET /api/v1/orders/00001
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "message": "Order retrieved successfully",
  "data": {
    "id": 1,
    "user_id": null,
    "order_code": "00001",
    "customer_name": "Nguyễn Văn A",
    "customer_email": "nguyenvana@example.com",
    "customer_phone": "0901234567",
    "shipping_address": {
      "full_name": "Nguyễn Văn A",
      "phone": "0901234567",
      "address_line": "123 Đường ABC, Phường XYZ",
      "city": "Hồ Chí Minh",
      "district": "Quận 1",
      "ward": "Phường Bến Nghé"
    },
    "total_amount": 578000,
    "shipping_fee": 30000,
    "discount_amount": 0,
    "final_amount": 608000,
    "status": "pending",
    "status_text": "Đang chờ xử lý",
    "payment_method": "cod",
    "payment_method_text": "Thanh toán khi nhận hàng",
    "payment_status": "unpaid",
    "payment_status_text": "Chưa thanh toán",
    "expires_at": "2025-06-11T19:00:00Z",
    "order_items": [
      {
        "id": 1,
        "product_id": 1,
        "product_size_id": 2,
        "product_name": "Áo thun nam basic",
        "product_size": "M",
        "quantity": 2,
        "unit_price": 199000,
        "total_price": 398000
      }
    ],
    "created_at": "2025-06-11T18:30:00Z",
    "updated_at": "2025-06-11T18:30:00Z"
  }
}
```

### 5. Lấy thông tin chuyển khoản ngân hàng

```bash
GET /api/v1/orders/00002/bank-transfer
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "message": "Bank transfer info retrieved successfully",
  "data": {
    "bank_name": "Ngân hàng TMCP Á Châu (ACB)",
    "account_number": "1234567890",
    "account_name": "CONG TY TNHH E-COMMERCE",
    "amount": 608000,
    "transfer_note": "DH 00002",
    "qr_code_url": "https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=Bank:1234567890,Amount:608000,Note:DH 00002"
  }
}
```

**Response lỗi - Sai payment method:**
```json
{
  "success": false,
  "message": "This order does not use bank transfer payment method"
}
```

**Response lỗi - Đã thanh toán:**
```json
{
  "success": false,
  "message": "This order has already been paid"
}
```

### 6. Hủy đơn hàng

```bash
POST /api/v1/orders/00001/cancel
Content-Type: application/json
X-Session-ID: 550e8400-e29b-41d4-a716-446655440000

{
  "reason": "Thay đổi ý định mua hàng"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Order cancelled successfully"
}
```

**Response lỗi - Không thể hủy:**
```json
{
  "success": false,
  "message": "Failed to cancel order",
  "error": "order cannot be cancelled. Current status: shipped"
}
```

### 7. Webhook xử lý thanh toán

```bash
POST /api/v1/webhook/payment
Content-Type: application/json

{
  "order_code": "00002",
  "transaction_id": "TXN_20250611_001",
  "response_data": {
    "gateway": "bank_transfer",
    "status": "success",
    "amount": 608000,
    "payment_time": "2025-06-11T19:15:00Z",
    "reference_code": "REF_123456"
  }
}
```

**Response:**
```json
{
  "success": true,
  "message": "Payment processed successfully"
}
```

### 8. Xác nhận thanh toán COD (Admin only)

```bash
POST /api/v1/admin/orders/00001/confirm-cod
Authorization: Bearer <admin-jwt-token>
X-CSRF-Token: <csrf-token>
```

**Response:**
```json
{
  "success": true,
  "message": "COD payment confirmed successfully"
}
```

---

## Workflow Examples

### Workflow 1: Guest User Order with COD

```bash
# 1. Thêm sản phẩm vào cart
POST /api/v1/cart/add
X-Session-ID: guest-session-uuid
{
  "product_id": 1,
  "product_size_id": 2,
  "quantity": 2
}

# 2. Lấy payment methods
GET /api/v1/orders/payment-methods

# 3. Tính toán order summary
POST /api/v1/orders/calculate
X-Session-ID: guest-session-uuid
{
  "customer_name": "Nguyễn Văn A",
  "customer_email": "nguyenvana@example.com",
  "customer_phone": "0901234567",
  "shipping_address": {...},
  "payment_method": "cod"
}

# 4. Tạo đơn hàng
POST /api/v1/orders
X-Session-ID: guest-session-uuid
{...}

# 5. Xem đơn hàng đã tạo
GET /api/v1/orders/00001
X-Session-ID: guest-session-uuid
```

### Workflow 2: Bank Transfer Payment

```bash
# 1. Tạo đơn hàng với bank transfer
POST /api/v1/orders
{
  "payment_method": "bank_transfer",
  ...
}

# 2. Lấy QR code và thông tin chuyển khoản
GET /api/v1/orders/00002/bank-transfer

# 3. Customer chuyển khoản theo thông tin

# 4. Admin xác nhận thanh toán (webhook hoặc manual)
POST /api/v1/webhook/payment
{
  "order_code": "00002",
  "transaction_id": "BANK_TXN_001",
  "response_data": {...}
}
```

### Workflow 3: Order Cancellation

```bash
# 1. Kiểm tra đơn hàng
GET /api/v1/orders/00001

# 2. Hủy đơn hàng với lý do
POST /api/v1/orders/00001/cancel
{
  "reason": "Khách hàng thay đổi ý định"
}

# 3. Kiểm tra status đã cập nhật
GET /api/v1/orders/00001
# Status: "cancelled", stock được hoàn trả
```

---

## Error Responses

### Validation Errors

**Invalid shipping address:**
```json
{
  "success": false,
  "message": "Validation failed",
  "error": "Field 'shipping_address.city' is required"
}
```

**Invalid payment method:**
```json
{
  "success": false,
  "message": "Validation failed",
  "error": "Field 'payment_method' is invalid. Must be one of: cod, bank_transfer, momo, zalopay, vnpay"
}
```

### Business Logic Errors

**Empty cart:**
```json
{
  "success": false,
  "message": "Failed to create order",
  "error": "cart validation failed: [Giỏ hàng trống]"
}
```

**Insufficient stock:**
```json
{
  "success": false,
  "message": "Failed to create order",
  "error": "insufficient stock for Áo thun nam basic - M. Available: 1, Required: 2"
}
```

**Order not found:**
```json
{
  "success": false,
  "message": "Order not found",
  "error": "record not found"
}
```

**Order expired:**
```json
{
  "success": false,
  "message": "Failed to create order",
  "error": "cart validation failed: [Giỏ hàng đã hết hạn]"
}
```

---

## Order Code System

### Format
- **5 digits**: 00001, 00002, ..., 99999
- **Auto increment**: Tự động tăng từ đơn hàng trước
- **Reset cycle**: Khi đến 99999 → reset về 00001
- **Thread safe**: Sử dụng mutex để tránh duplicate

### Examples
```bash
# First order
Order Code: 00001

# After 50 orders  
Order Code: 00051

# At limit
Order Code: 99999

# Reset cycle
Order Code: 00001
```

---

## Shipping Fee Calculation

### Logic
- **Free shipping**: Đơn hàng ≥ 500,000 VNĐ
- **Base fee**: 30,000 VNĐ
- **Remote areas**: +20,000 VNĐ (Cà Mau, An Giang, Kiên Giang, Hà Giang)

### Examples
```bash
# Order 300,000 VNĐ - TP.HCM
Shipping: 30,000 VNĐ

# Order 300,000 VNĐ - Cà Mau  
Shipping: 50,000 VNĐ

# Order 600,000 VNĐ - anywhere
Shipping: 0 VNĐ (Free)
```

---

## Discount Calculation

### Logic
- **5% discount**: Đơn hàng ≥ 1,000,000 VNĐ
- **No minimum**: Không giảm giá cho đơn < 1,000,000 VNĐ

### Examples
```bash
# Order 800,000 VNĐ
Discount: 0 VNĐ

# Order 1,200,000 VNĐ
Discount: 60,000 VNĐ (5%)
```

---

## Order Expiry

### Rules
- **Expiry time**: 30 phút sau khi tạo đơn
- **Auto cleanup**: Chạy mỗi 15 phút
- **Stock restore**: Tự động hoàn trả stock khi hủy

### Process
1. Đơn hàng tạo → expires_at = now + 30 minutes
2. Nếu chưa thanh toán sau 30 phút → tự động hủy
3. Stock được hoàn trả về inventory
4. Email thông báo hủy đơn

---

## Best Practices

### 1. Order Tracking
```javascript
// Lưu order code để tracking
const orderCode = response.data.order.order_code;
localStorage.setItem('last_order', orderCode);

// Check order status periodically
const checkOrderStatus = async (orderCode) => {
  const response = await fetch(`/api/v1/orders/${orderCode}`);
  return response.data;
};
```

### 2. Payment Method Selection
```javascript
// Get available payment methods
const paymentMethods = await fetch('/api/v1/orders/payment-methods');
const availableMethods = paymentMethods.data.filter(m => m.is_available);

// Show only available methods to user
```

### 3. Real-time Stock Check
```javascript
// Calculate order before creation
const summary = await fetch('/api/v1/orders/calculate', {
  method: 'POST',
  body: JSON.stringify(orderData)
});

if (summary.success) {
  // Proceed with order creation
  const order = await createOrder(orderData);
}
```

### 4. Bank Transfer Flow
```javascript
// After creating bank transfer order
if (order.payment_method === 'bank_transfer') {
  // Get QR code and bank info
  const bankInfo = await fetch(`/api/v1/orders/${orderCode}/bank-transfer`);
  
  // Display QR code and instructions
  showBankTransferInstructions(bankInfo.data);
  
  // Set up payment confirmation polling
  startPaymentStatusPolling(orderCode);
}
```

---

## Notes

1. **Session Persistence**: Guest users cần lưu session ID để access orders
2. **Stock Locking**: Hệ thống sử dụng database locking để tránh overselling
3. **Email Notifications**: Tự động gửi email cho các sự kiện quan trọng
4. **Webhook Security**: Production cần implement signature verification
5. **Order Expiry**: Đơn hàng tự động hủy sau 30 phút nếu chưa thanh toán
6. **Concurrent Orders**: Multiple users có thể đặt cùng sản phẩm an toàn
7. **Payment Gateway**: Ready để tích hợp MoMo, ZaloPay, VNPay