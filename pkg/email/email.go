package email

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/cheel98/flashcard-backend/internal/config"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

// EmailService 邮件服务
type EmailService struct {
	config *config.EmailConfig
	logger *zap.Logger
}

// NewEmailService 创建新的邮件服务
func NewEmailService(cfg *config.Config, logger *zap.Logger) *EmailService {
	return &EmailService{
		config: &cfg.Email,
		logger: logger,
	}
}

// SendCaptcha 发送验证码邮件
func (e *EmailService) SendCaptcha(toEmail, captcha string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.config.FromEmail)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "验证码 - Flashcard App")

	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>验证码</h2>
			<p>您的验证码是：<strong>%s</strong></p>
			<p>验证码有效期为5分钟，请及时使用。</p>
			<p>如果您没有请求此验证码，请忽略此邮件。</p>
			<br>
			<p>此邮件由系统自动发送，请勿回复。</p>
		</body>
		</html>
	`, captcha)

	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.config.SMTPHost, e.config.SMTPPort, e.config.SMTPUsername, e.config.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		e.logger.Error("发送邮件失败",
			zap.String("to", toEmail),
			zap.Error(err))
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	e.logger.Info("验证码邮件发送成功", zap.String("to", toEmail))
	return nil
}

// GenerateCaptcha 生成6位数字验证码
func (e *EmailService) GenerateCaptcha() (string, error) {
	const digits = "0123456789"
	const length = 6

	captcha := make([]byte, length)
	for i := range captcha {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			e.logger.Error("生成验证码失败", zap.Error(err))
			return "", fmt.Errorf("生成验证码失败: %w", err)
		}
		captcha[i] = digits[n.Int64()]
	}

	return string(captcha), nil
}
