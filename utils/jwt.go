package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Role      int    `json:"role"`
	CSRFToken string `json:"csrf_token"`
	jwt.RegisteredClaims
}

// Thời gian mặc định JWT có hiệu lực (24 giờ)
const (
	DefaultTokenExpiry = 24 * time.Hour // 24 hours
	MaxFailedLogins    = 5
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Lỗi khi đọc file .env")
	}
}

// Lấy secret key để ký JWT từ biến môi trường JWT_SECRET
// Nếu không có biến môi trường JWT_SECRET thì tạo một key ngẫu nhiên
func GetJWTSecret() string {
	loadEnv()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your-super-secret-jwt-key-change-this-in-production"
	}
	return secret
}

// Lấy thời gian hết hạn của token từ biến môi trường JWT_EXPIRY_HOURS
// Nếu không có biến môi trường này thì mặc định là 24 giờ
func GetTokenExpiry() time.Duration {
	loadEnv()
	expiryStr := os.Getenv("JWT_EXPIRY_HOURS")
	if expiryStr == "" {
		return DefaultTokenExpiry
	}

	hours, err := strconv.Atoi(expiryStr)
	if err != nil {
		return DefaultTokenExpiry
	}

	return time.Duration(hours) * time.Hour
}

// Tạo JWT token và CSRF token
// Lấy thời gian hết hạn từ GetTokenExpiry()
// Tạo CSRF token ngẫu nhiên 32 byte
// Đóng gói các claims vào JWT
// Ký token bằng thuật toán HS256 với secret key
// Trả về: token string, CSRF token, thời gian hết hạn
func GenerateToken(userID uint, email string, role int) (string, string, time.Time, error) {
	expiry := GetTokenExpiry()
	expiresAt := time.Now().Add(expiry)

	// Generate CSRF token
	csrfToken, err := GenerateRandomString(32)
	if err != nil {
		return "", "", time.Time{}, err
	}

	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		CSRFToken: csrfToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ecommerce-api",
			Audience:  []string{"ecommerce-web", "ecommerce-mobile"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(GetJWTSecret()))
	if err != nil {
		return "", "", time.Time{}, err
	}

	return tokenString, csrfToken, expiresAt, nil
}

// Xác thực JWT
// Kiểm tra chữ ký có đúng phương pháp HMAC không
// Dùng secret key để xác thực
// Kiểm tra token còn hiệu lực (chưa hết hạn)
// Trả về claims nếu hợp lệ
func ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(GetJWTSecret()), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token is expired")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Tạo chuỗi ngẫu nhiên an toàn (dùng cho CSRF token và recovery code)
// Sử dụng crypto/rand để đảm bảo tính bảo mật
// Mã hóa kết quả thành chuỗi hex
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Tạo mã khôi phục (16 byte, khi mã hóa hex sẽ thành 32 ký tự)
func GenerateRecoveryCode() (string, error) {
	return GenerateRandomString(16) // 32 character hex string
}
