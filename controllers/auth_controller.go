package controllers

import (
	"fmt"
	"net/http"
	"reflect"
	"storelite/config"
	"storelite/models"
	"storelite/services"
	"storelite/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	validator    *validator.Validate
	emailService *services.EmailService
}

func NewAuthController() *AuthController {
	return &AuthController{
		validator:    validator.New(),
		emailService: services.NewEmailService(),
	}
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Register xử lý đăng ký tài khoản mới
func (ac *AuthController) Register(c *gin.Context) {
	var req models.RegisterRequest

	// Bind JSON data
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Validate request
	if err := ac.validator.Struct(&req); err != nil {
		errors := err.(validator.ValidationErrors)
		errorMessages := make([]string, 0)
		for _, e := range errors {
			field, _ := reflect.TypeOf(req).FieldByName(e.StructField()) //reflect
			customMessage := field.Tag.Get("message")
			if customMessage != "" {
				errorMessages = append(errorMessages, customMessage)
			} else {
				// Message mặc định
				switch e.Tag() {
				case "required":
					errorMessages = append(errorMessages, fmt.Sprintf("%s là bắt buộc", e.Field()))
				case "email":
					errorMessages = append(errorMessages, "Email không đúng định dạng")
				case "min":
					errorMessages = append(errorMessages, fmt.Sprintf("%s phải có ít nhất %s ký tự", e.Field(), e.Param()))
				case "eqfield":
					errorMessages = append(errorMessages, fmt.Sprintf("%s phải giống với %s", e.Field(), e.Param()))
				default:
					errorMessages = append(errorMessages, fmt.Sprintf("Lỗi validation cho %s", e.Field()))
				}
			}
		}
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   strings.Join(errorMessages, ", "),
		})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := config.GetDB().Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Message: "Email already registered",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to hash password",
		})
		return
	}

	// Parse birthday if provided
	var birthday *time.Time
	if req.Birthday != "" {
		if parsedBirthday, err := time.Parse("2006-01-02", req.Birthday); err == nil {
			birthday = &parsedBirthday
		} else {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid birthday format. Use YYYY-MM-DD",
			})
			return
		}
	}

	// Create new user
	user := models.User{
		Fullname: req.Fullname,
		Email:    req.Email,
		Password: string(hashedPassword),
		Address:  req.Address,
		Phone:    req.Phone,
		Birthday: birthday,
		Role:     models.RoleUser, // Default role is user
		IsActive: true,            // Default is active
	}

	// Save to database
	if err := config.GetDB().Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user.ToResponse(),
	})
}

// Login xử lý đăng nhập
func (ac *AuthController) Login(c *gin.Context) {
	var req models.LoginRequest

	// Bind JSON data
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	// Validate request
	if err := ac.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Get IP and User Agent for logging
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Find user by email
	var user models.User
	if err := config.GetDB().Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Log failed attempt
		ac.logLoginAttempt(req.Email, ipAddress, userAgent, false, "Email not found")

		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	// Check if account is active
	if !user.IsActive {
		// Log failed attempt
		ac.logLoginAttempt(req.Email, ipAddress, userAgent, false, "Account is locked")

		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "Tài khoản đã bị khóa do nhập sai mật khẩu quá " +
				fmt.Sprintf("%d", utils.MaxFailedLogins) + " lần. Vui lòng kiểm tra email để nhận mã khôi phục.",
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Increment failed login count
		user.FailedLoginCount++

		// Check if exceeded max failed attempts
		if user.FailedLoginCount >= utils.MaxFailedLogins {
			user.IsActive = false

			// Send account locked email
			go ac.emailService.SendAccountLockedEmail(user.Email, user.Fullname)
		}

		// Update user in database
		config.GetDB().Save(&user)

		// Log failed attempt
		failReason := fmt.Sprintf("Invalid password (attempt %d/%d)",
			user.FailedLoginCount, utils.MaxFailedLogins)
		ac.logLoginAttempt(req.Email, ipAddress, userAgent, false, failReason)

		// Return appropriate message
		if user.FailedLoginCount >= utils.MaxFailedLogins {
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: "Tài khoản đã bị khóa do nhập sai mật khẩu quá " +
					fmt.Sprintf("%d", utils.MaxFailedLogins) + " lần. Mã khôi phục đã được gửi đến email của bạn.",
			})
		} else {
			remainingAttempts := utils.MaxFailedLogins - user.FailedLoginCount
			c.JSON(http.StatusUnauthorized, APIResponse{
				Success: false,
				Message: fmt.Sprintf("Mật khẩu không đúng. Còn %d lần thử.", remainingAttempts),
			})
		}
		return
	}

	// Login successful - reset failed count and update last login
	lastLoginAt := user.LastLoginAt
	user.FailedLoginCount = 0
	user.LastLoginAt = &time.Time{}
	*user.LastLoginAt = time.Now()

	if err := config.GetDB().Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update user login info",
		})
		return
	}

	// Generate JWT token
	token, csrfToken, expiresAt, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to generate token",
		})
		return
	}

	// Log successful attempt
	ac.logLoginAttempt(req.Email, ipAddress, userAgent, true, "Login successful")

	// Calculate login count (for demo, you might want to store this in a separate table)
	var loginCount int64
	config.GetDB().Model(&models.LoginAttempt{}).
		Where("email = ? AND is_success = ?", user.Email, true).
		Count(&loginCount)

	// Return success response
	response := models.LoginResponse{
		User:      user.ToResponse(),
		Token:     token,
		TokenType: "Bearer",
		ExpiresAt: expiresAt,
		ExpiresIn: int64(utils.GetTokenExpiry().Seconds()),
		CSRFToken: csrfToken,
		LoginInfo: models.LoginInfoData{
			LoginTime:   time.Now(),
			LastLoginAt: lastLoginAt,
			LoginCount:  int(loginCount) + 1,
		},
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// logLoginAttempt ghi log các lần đăng nhập
func (ac *AuthController) logLoginAttempt(email, ipAddress, userAgent string, isSuccess bool, failReason string) {
	attempt := models.LoginAttempt{
		Email:       email,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		IsSuccess:   isSuccess,
		FailReason:  failReason,
		AttemptedAt: time.Now(),
	}

	config.GetDB().Create(&attempt)
}

// RequestRecovery xử lý yêu cầu khôi phục tài khoản
func (ac *AuthController) RequestRecovery(c *gin.Context) {
	var req models.RecoveryRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}
	if err := ac.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}
	// Find user
	var user models.User
	if err := config.GetDB().Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Message: "Nếu email tồn tại, mã khôi phục đã được gửi đến email của bạn.",
		})
		return
	}

	// Generate recovery code
	code, err := utils.GenerateRecoveryCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to generate recovery code",
		})
		return
	}
	// Save recovery code to database
	recoveryCode := models.RecoveryCode{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(30 * time.Minute), // 30 minutes expiry
	}
	if err := config.GetDB().Create(&recoveryCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to save recovery code",
		})
		return
	}
	// Send recovery email
	go ac.emailService.SendRecoveryEmail(user.Email, code, user.Fullname)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Mã khôi phục đã được gửi đến email của bạn. Mã có hiệu lực trong 30 phút.",
	})
}

// VerifyRecovery xử lý xác minh mã khôi phục
func (ac *AuthController) VerifyRecovery(c *gin.Context) {
	var req models.RecoveryVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Invalid request data",
			Error:   err.Error(),
		})
		return
	}

	if err := ac.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   err.Error(),
		})
		return
	}

	// Find recovery code
	var recoveryCode models.RecoveryCode
	if err := config.GetDB().Preload("User").
		Where("code = ? AND is_used = ? AND expires_at > ?",
			req.Code, false, time.Now()).
		First(&recoveryCode).Error; err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Mã khôi phục không hợp lệ hoặc đã hết hạn.",
		})
		return
	}

	// Mark recovery code as used
	recoveryCode.IsUsed = true
	if err := config.GetDB().Save(&recoveryCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to update recovery code",
		})
		return
	}

	// Reactivate user account and reset failed login count
	user := recoveryCode.User
	user.IsActive = true
	user.FailedLoginCount = 0

	if err := config.GetDB().Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to reactivate account",
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Tài khoản đã được kích hoạt lại thành công. Bạn có thể đăng nhập ngay bây giờ.",
		Data: gin.H{
			"user_email": user.Email,
			"is_active":  user.IsActive,
		},
	})
}

// GetProfile lấy thông tin profile của user hiện tại
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, APIResponse{
			Success: false,
			Message: "User ID not found in context",
		})
		return
	}

	var user models.User
	if err := config.GetDB().First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "User not found",
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

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    user.ToResponse(),
	})
}

// GetUserByID lấy thông tin user theo ID
func (ac *AuthController) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := config.GetDB().First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, APIResponse{
				Success: false,
				Message: "User not found",
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

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "User found",
		Data:    user.ToResponse(),
	})
}

// GetAllUsers lấy danh sách tất cả users
func (ac *AuthController) GetAllUsers(c *gin.Context) {
	var users []models.User

	if err := config.GetDB().Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: "Failed to fetch users",
			Error:   err.Error(),
		})
		return
	}

	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "Users fetched successfully",
		Data:    userResponses,
	})
}
