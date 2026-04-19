package service

import (
	"fmt"

	"gopay/src/config"

	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/dysmsapi-20170525/v2/client"
	"github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

// 短信服务
type SmsService struct {
	client *client.Client
}

func NewSmsService() (*SmsService, error) {
	accessKeyId := config.Get("sms_access_key_id")
	accessKeySecret := config.Get("sms_access_key_secret")

	if accessKeyId == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("短信配置不完整")
	}

	conf := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}

	c, err := client.NewClient(conf)
	if err != nil {
		return nil, err
	}

	return &SmsService{
		client: c,
	}, nil
}

// Send 发送短信
func (s *SmsService) Send(phone string, templateCode string, params map[string]string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("短信服务未初始化")
	}

	signName := config.Get("sms_sign_name")
	if signName == "" {
		signName = "GoPay支付"
	}

	templateParam := ""
	for k, v := range params {
		templateParam += fmt.Sprintf(`{"%s":"%s"}`, k, v)
	}

	sendSmsRequest := &client.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(templateParam),
	}

	runtime := &service.RuntimeOptions{}
	resp, err := s.client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		return err
	}
	if *resp.Body.Code != "OK" {
		return fmt.Errorf("短信发送失败: %s - %s", *resp.Body.Code, *resp.Body.Message)
	}
	return nil
}

// SendCode 发送验证码
func (s *SmsService) SendCode(phone, code string) error {
	templateCode := config.Get("sms_template_code")
	if templateCode == "" {
		return fmt.Errorf("未配置短信模板")
	}
	return s.Send(phone, templateCode, map[string]string{
		"code": code,
	})
}

// SendOrderNotify 发送订单通知
func (s *SmsService) SendOrderNotify(phone, tradeNo, amount string) error {
	templateCode := config.Get("sms_order_template_code")
	if templateCode == "" {
		return fmt.Errorf("未配置订单通知模板")
	}
	return s.Send(phone, templateCode, map[string]string{
		"trade_no": tradeNo,
		"amount":   amount,
	})
}

// 全局短信服务实例
var globalSmsService *SmsService

// InitSmsService 初始化短信服务
func InitSmsService() error {
	if config.Get("sms_enabled") != "1" {
		fmt.Println("[SMS] 短信服务未启用")
		return nil
	}

	var err error
	globalSmsService, err = NewSmsService()
	if err != nil {
		fmt.Printf("[SMS] 初始化失败: %v\n", err)
		return err
	}
	fmt.Println("[SMS] 短信服务初始化成功")
	return nil
}

// GetSmsService 获取短信服务
func GetSmsService() *SmsService {
	return globalSmsService
}
