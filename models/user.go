package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleAdmin = 0
	RoleUser  = 1
	RoleVIP   = 2
)

type User struct {
	ID               uint           `json:"id" gorm:"primaryKey"`
	Fullname         string         `json:"fullname" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	Email            string         `json:"email" gorm:"uniqueIndex;not null;size:255" validate:"required,email"`
	Password         string         `json:"-" gorm:"not null;size:255" validate:"required,min=6"`
	Address          string         `json:"address" gorm:"size:500"`
	Phone            string         `json:"phone" gorm:"size:20" validate:"omitempty,min=10,max=20"`
	Birthday         *time.Time     `json:"birthday"`
	Role             int            `json:"role" gorm:"default:1;check:role IN (0,1,2)"`
	IsActive         bool           `json:"is_active" gorm:"default:true"`
	FailedLoginCount int            `json:"-" gorm:"default:0"`
	LastLoginAt      *time.Time     `json:"last_login_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

type RegisterRequest struct {
	Fullname        string `json:"fullname" validate:"required,min=2,max=255"`
	Email           string `json:"email" validate:"required,email" message:""`
	Password        string `json:"password" validate:"required,min=6" message:""`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password" message:""`
	Address         string `json:"address"`
	Phone           string `json:"phone" validate:"omitempty,min=10,max=20"`
	Birthday        string `json:"birthday"` // Format: YYYY-MM-DD
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User      UserResponse  `json:"user"`
	Token     string        `json:"token"`
	TokenType string        `json:"token_type"`
	ExpiresAt time.Time     `json:"expires_at"`
	ExpiresIn int64         `json:"expires_in"` // seconds
	CSRFToken string        `json:"csrf_token"`
	LoginInfo LoginInfoData `json:"login_info"`
}

type LoginInfoData struct {
	LoginTime   time.Time  `json:"login_time"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LoginCount  int        `json:"login_count"`
}

type RecoveryCode struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	Code      string         `json:"code" gorm:"uniqueIndex;not null;size:255"`
	IsUsed    bool           `json:"is_used" gorm:"default:false"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	User      User           `json:"-" gorm:"foreignKey:UserID"`
}

type LoginAttempt struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Email       string         `json:"email" gorm:"not null;size:255"`
	IPAddress   string         `json:"ip_address" gorm:"size:45"`
	UserAgent   string         `json:"user_agent" gorm:"size:500"`
	IsSuccess   bool           `json:"is_success" gorm:"default:false"`
	FailReason  string         `json:"fail_reason" gorm:"size:255"`
	AttemptedAt time.Time      `json:"attempted_at"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type RecoveryRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type RecoveryVerifyRequest struct {
	Code string `json:"code" validate:"required"`
}

type UserResponse struct {
	ID       uint       `json:"id"`
	Fullname string     `json:"fullname"`
	Email    string     `json:"email"`
	Address  string     `json:"address"`
	Phone    string     `json:"phone"`
	Birthday *time.Time `json:"birthday"`
	Role     int        `json:"role"`
	IsActive bool       `json:"is_active"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Fullname: u.Fullname,
		Email:    u.Email,
		Address:  u.Address,
		Phone:    u.Phone,
		Birthday: u.Birthday,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}

func GetRoleName(role int) string {
	switch role {
	case RoleAdmin:
		return "Admin"
	case RoleUser:
		return "User"
	case RoleVIP:
		return "VIP"
	default:
		return "Unknown"
	}
}
