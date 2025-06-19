# E-commerce API with Golang, Gin & GORM

API thương mại điện tử được xây dựng bằng Golang với framework Gin và GORM ORM, sử dụng PostgreSQL làm cơ sở dữ liệu.

## Cấu trúc dự án

```
ecommerce-api/
├── config/
│   └── database.go          # Cấu hình kết nối database
├── controllers/
│   └── auth_controller.go   # Controller xử lý authentication
├── models/
│   └── user.go             # Model User
├── routes/
│   ├── auth_routes.go      # Routes cho authentication
│   └── routes.go           # Setup routes chính
├── main.go                 # File main
├── go.mod                  # Go modules
├── .env.example           # File mẫu environment variables
└── README.md              # Tài liệu hướng dẫn
```

## Tính năng hiện tại

### User Model
- **ID**: Primary key tự động tăng
- **Fullname**: Họ tên đầy đủ (bắt buộc)
- **Email**: Email duy nhất (bắt buộc)
- **Password**: Mật khẩu được hash bằng bcrypt
- **Address**: Địa chỉ
- **Phone**: Số điện thoại
- **Birthday**: Ngày sinh
- **Role**: Vai trò (0: Admin, 1: User, 2: VIP)
- **IsActive**: Trạng thái hoạt động
- **Timestamps**: CreatedAt, UpdatedAt, DeletedAt

### API Endpoints

#### Authentication
- `POST /api/v1/auth/register` - Đăng ký tài khoản mới

#### Users
- `GET /api/v1/users` - Lấy danh sách tất cả users
- `GET /api/v1/users/:id` - Lấy thông tin user theo ID

#### Health Check
- `GET /health` - Kiểm tra trạng thái API

## Cài đặt và chạy dự án

### Yêu cầu hệ thống
- Go 1.21+
- PostgreSQL 12+

### Bước 1: Clone dự án
```bash
git clone <repository-url>
cd ecommerce-api
```

### Bước 2: Cài đặt dependencies
```bash
go mod tidy
```

### Bước 3: Cấu hình database
1. Tạo database PostgreSQL:
```sql
CREATE DATABASE ecommerce_db;
```

2. Sao chép file `.env.example` thành `.env` và cập nhật thông tin:
```bash
cp .env.example .env
```

3. Chỉnh sửa file `.env` với thông tin database của bạn:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=ecommerce_db
DB_SSLMODE=disable
```

### Bước 4: Chạy ứng dụng
```bash
go run main.go
```

API sẽ chạy tại `http://localhost:8080`

## Sử dụng API

### Đăng ký tài khoản mới

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "fullname": "Nguyễn Văn A",
  "email": "nguyenvana@example.com",
  "password": "password123",
  "confirm_password": "password123",
  "address": "123 Đường ABC, Quận 1, TP.HCM",
  "phone": "0901234567",
  "birthday": "1990-01-15"
}
```

**Response thành công:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "fullname": "Nguyễn Văn A",
    "email": "nguyenvana@example.com",
    "address": "123 Đường ABC, Quận 1, TP.HCM",
    "phone": "0901234567",
    "birthday": "1990-01-15T00:00:00Z",
    "role": 1,
    "is_active": true
  }
}
```

### Lấy danh sách users

**Endpoint:** `GET /api/v1/users`

**Response:**
```json
{
  "success": true,
  "message": "Users fetched successfully",
  "data": [
    {
      "id": 1,
      "fullname": "Nguyễn Văn A",
      "email": "nguyenvana@example.com",
      "address": "123 Đường ABC, Quận 1, TP.HCM",
      "phone": "0901234567",
      "birthday": "1990-01-15T00:00:00Z",
      "role": 1,
      "is_active": true
    }
  ]
}
```

## Validation Rules

### Đăng ký tài khoản
- **fullname**: Bắt buộc, 2-255 ký tự
- **email**: Bắt buộc, định dạng email hợp lệ, duy nhất
- **password**: Bắt buộc, tối thiểu 6 ký tự
- **confirm_password**: Bắt buộc, phải giống password
- **phone**: Tùy chọn, 10-20 ký tự nếu có
- **birthday**: Tùy chọn, định dạng YYYY-MM-DD

## Roles (Vai trò)
- **0**: Admin - Quản trị viên
- **1**: User - Người dùng thường (default)
- **2**: VIP - Người dùng VIP

## Security Features
- Mật khẩu được hash bằng bcrypt
- Email validation và unique constraint
- Input validation với go-playground/validator
- CORS middleware
- Soft delete với GORM

## Tính năng sắp tới
- [ ] Đăng nhập (login)
- [ ] JWT Authentication
- [ ] Quản lý sản phẩm (Product)
- [ ] Quản lý danh mục (Category)
- [ ] Giỏ hàng (Cart)
- [ ] Đặt hàng (Order)
- [ ] Thanh toán (Payment)

## Development

### Thêm model mới
1. Tạo model trong thư mục `models/`
2. Thêm model vào function `autoMigrate()` trong `main.go`
3. Tạo controller tương ứng trong `controllers/`
4. Setup routes trong `routes/`

### Testing
```bash
# Test đăng ký tài khoản
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "fullname": "Test User",
    "email": "test@example.com",
    "password": "password123",
    "confirm_password": "password123"
  }'
```