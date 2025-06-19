package services

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer *gomail.Dialer
	from   string
}

func NewEmailService() *EmailService {
	env := godotenv.Load()
	if env != nil {
		log.Fatal("Lỗi khi đọc file .env")
	}
	host := getEnv("SMTP_HOST", "smtp.gmail.com")
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	username := getEnv("SMTP_USERNAME", "")
	password := getEnv("SMTP_PASSWORD", "")
	from := getEnv("SMTP_FROM", username)

	dialer := gomail.NewDialer(host, port, username, password)

	return &EmailService{
		dialer: dialer,
		from:   from,
	}
}

func (es *EmailService) SendRecoveryEmail(to, code, fullname string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Khôi phục tài khoản - E-commerce")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Khôi phục tài khoản</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; padding: 30px; border-radius: 10px;">
        <h2 style="color: #333; text-align: center;">Khôi phục tài khoản</h2>
        
        <p>Xin chào <strong>%s</strong>,</p>
        
        <p>Chúng tôi nhận được yêu cầu khôi phục tài khoản của bạn. Tài khoản của bạn đã bị khóa do nhập sai mật khẩu quá nhiều lần.</p>
        
        <div style="background-color: white; padding: 20px; border-radius: 5px; margin: 20px 0; text-align: center;">
            <p style="margin: 0; font-size: 16px;">Mã khôi phục của bạn là:</p>
            <h3 style="color: #007bff; font-size: 24px; letter-spacing: 2px; margin: 10px 0;">%s</h3>
        </div>
        
        <p><strong>Lưu ý quan trọng:</strong></p>
        <ul>
            <li>Mã khôi phục này có hiệu lực trong <strong>30 phút</strong></li>
            <li>Chỉ sử dụng mã này một lần duy nhất</li>
            <li>Không chia sẻ mã này với bất kỳ ai</li>
            <li>Nếu bạn không yêu cầu khôi phục, vui lòng bỏ qua email này</li>
        </ul>
        
        <p>Để khôi phục tài khoản, vui lòng sử dụng API endpoint:</p>
        <code style="background-color: #f1f1f1; padding: 5px; border-radius: 3px;">POST /api/v1/auth/recovery/verify</code>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        
        <p style="color: #666; font-size: 14px; text-align: center;">
            Email này được gửi tự động, vui lòng không trả lời.<br>
            © 2025 E-commerce API. All rights reserved.
        </p>
    </div>
</body>
</html>
	`, fullname, code)

	m.SetBody("text/html", body)

	return es.dialer.DialAndSend(m)
}

func (es *EmailService) SendAccountLockedEmail(to, fullname string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", es.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Tài khoản bị khóa - E-commerce")

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Tài khoản bị khóa</title>
</head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #fff3cd; padding: 30px; border-radius: 10px; border-left: 5px solid #ffc107;">
        <h2 style="color: #856404; text-align: center;">⚠️ Tài khoản bị khóa</h2>
        
        <p>Xin chào <strong>%s</strong>,</p>
        
        <p>Tài khoản của bạn đã bị khóa do nhập sai mật khẩu <strong>5 lần liên tiếp</strong>.</p>
        
        <div style="background-color: #f8d7da; padding: 15px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #dc3545;">
            <p style="margin: 0; color: #721c24;">
                <strong>Lý do khóa:</strong> Nhập sai mật khẩu quá nhiều lần<br>
                <strong>Thời gian:</strong> %s<br>
                <strong>Trạng thái:</strong> Tài khoản bị vô hiệu hóa
            </p>
        </div>
        
        <p><strong>Để khôi phục tài khoản:</strong></p>
        <ol>
            <li>Gửi yêu cầu khôi phục qua API: <code>POST /api/v1/auth/recovery</code></li>
            <li>Nhận mã khôi phục qua email</li>
            <li>Sử dụng mã để kích hoạt lại tài khoản</li>
        </ol>
        
        <p style="color: #856404; background-color: #fff3cd; padding: 10px; border-radius: 5px;">
            <strong>Bảo mật:</strong> Nếu bạn không thực hiện các lần đăng nhập này, 
            tài khoản của bạn có thể đang bị tấn công. Vui lòng thay đổi mật khẩu ngay sau khi khôi phục.
        </p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        
        <p style="color: #666; font-size: 14px; text-align: center;">
            Email này được gửi tự động, vui lòng không trả lời.<br>
            © 2025 E-commerce API. All rights reserved.
        </p>
    </div>
</body>
</html>
	`, fullname, getCurrentTime())

	m.SetBody("text/html", body)

	return es.dialer.DialAndSend(m)
}

func getCurrentTime() string {
	return fmt.Sprintf("%s (UTC+7)",
		time.Now().In(time.FixedZone("ICT", 7*60*60)).Format("2006-01-02 15:04:05"))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
