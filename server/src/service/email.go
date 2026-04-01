package service

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"paygo/src/config"
)

// 邮件服务
type EmailService struct{}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// 发送邮件
func (s *EmailService) Send(to, subject, body string) error {
	// 获取邮件配置
	smtpHost := config.Get("mail_smtp_host")
	smtpPort := config.Get("mail_smtp_port")
	username := config.Get("mail_username")
	password := config.Get("mail_password")

	if smtpHost == "" || username == "" || password == "" {
		return fmt.Errorf("邮件配置不完整")
	}

	port := 587 // 默认
	if smtpPort != "" {
		fmt.Sscanf(smtpPort, "%d", &port)
	}

	m := gomail.NewMessage()
	m.SetAddressHeader("From", username, "PayGo支付系统")
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(smtpHost, port, username, password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("发送失败: %v", err)
	}

	return nil
}

// 发送验证码邮件
func (s *EmailService) SendCode(email, code string) error {
	subject := "【PayGo支付】验证码"
	body := fmt.Sprintf(`
		<div style="max-width: 500px; margin: 0 auto; padding: 20px; font-family: Arial, sans-serif;">
			<div style="background: #007bff; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0;">
				<h2>PayGo支付系统</h2>
			</div>
			<div style="background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px;">
				<p style="font-size: 16px; color: #333;">您好！</p>
				<p style="font-size: 16px; color: #333;">您的验证码是：</p>
				<div style="background: #fff; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; border: 2px dashed #007bff;">
					<span style="font-size: 32px; font-weight: bold; color: #007bff; letter-spacing: 8px;">%s</span>
				</div>
				<p style="font-size: 14px; color: #666;">验证码将在5分钟后失效，请勿泄露给他人。</p>
				<hr style="margin: 20px 0; border: none; border-top: 1px solid #ddd;">
				<p style="font-size: 12px; color: #999;">此邮件由系统自动发送，请勿回复。</p>
			</div>
		</div>
	`, code)

	return s.Send(email, subject, body)
}

// 发送订单通知
func (s *EmailService) SendOrderNotify(email, tradeNo, amount string) error {
	subject := "【PayGo支付】订单通知"
	body := fmt.Sprintf(`
		<div style="max-width: 500px; margin: 0 auto; padding: 20px; font-family: Arial, sans-serif;">
			<div style="background: #28a745; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0;">
				<h2>订单支付成功</h2>
			</div>
			<div style="background: #f8f9fa; padding: 30px; border-radius: 0 0 8px 8px;">
				<p style="font-size: 16px; color: #333;">您好！</p>
				<p style="font-size: 16px; color: #333;">您的订单已完成支付：</p>
				<table style="width: 100%%; margin: 20px 0; border-collapse: collapse;">
					<tr>
						<td style="padding: 10px; border: 1px solid #ddd;">订单号</td>
						<td style="padding: 10px; border: 1px solid #ddd; font-family: monospace;">%s</td>
					</tr>
					<tr>
						<td style="padding: 10px; border: 1px solid #ddd;">支付金额</td>
						<td style="padding: 10px; border: 1px solid #ddd; color: #28a745; font-weight: bold;">%s</td>
					</tr>
				</table>
				<p style="font-size: 14px; color: #666;">感谢您的支持！</p>
			</div>
		</div>
	`, tradeNo, amount)

	return s.Send(email, subject, body)
}
