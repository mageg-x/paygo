package service

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// 管理员登录
func (s *AuthService) AdminLogin(username, password string) (string, error) {
	cfg := config.AppConfig
	if username != cfg.AdminUser || password != cfg.AdminPwd {
		log.Printf("[admin_login_failed] username=%s, reason=invalid credentials", username)
		return "", errors.New("用户名或密码错误")
	}

	token := s.genAdminToken(username, password)
	return token, nil
}

func (s *AuthService) genAdminToken(username, password string) string {
	hash := md5.Sum([]byte(username + password + password + config.AppConfig.SysKey))
	return hex.EncodeToString(hash[:])
}

// 商户登录
func (s *AuthService) UserLogin(uid uint, pwd string) (*model.User, string, error) {
	var user model.User
	result := config.DB.First(&user, uid)
	if result.Error != nil {
		log.Printf("[user_login_failed] uid=%d, reason=user not found, error=%s", uid, result.Error.Error())
		return nil, "", errors.New("用户不存在")
	}

	if user.Status != 0 {
		log.Printf("[user_login_failed] uid=%d, reason=user disabled, status=%d", uid, user.Status)
		return nil, "", errors.New("账号已被禁用")
	}

	// 验证密码
	pwdHash := md5.Sum([]byte(pwd + user.Key))
	pwdStr := hex.EncodeToString(pwdHash[:])
	if pwdStr != user.Pwd {
		log.Printf("[user_login_failed] uid=%d, reason=invalid password")
		return nil, "", errors.New("密码错误")
	}

	// 更新最后登录时间
	config.DB.Model(&user).Update("lasttime", time.Now())

	// 生成token
	token := s.genUserToken(user.UID, user.Key)

	return &user, token, nil
}

// 商户密钥登录
func (s *AuthService) UserKeyLogin(uid uint, key string) (*model.User, string, error) {
	var user model.User
	result := config.DB.First(&user, uid)
	if result.Error != nil {
		log.Printf("[user_key_login_failed] uid=%d, reason=user not found, error=%s", uid, result.Error.Error())
		return nil, "", errors.New("用户不存在")
	}

	if user.Status != 0 {
		log.Printf("[user_key_login_failed] uid=%d, reason=user disabled, status=%d", uid, user.Status)
		return nil, "", errors.New("账号已被禁用")
	}

	if key != user.Key {
		log.Printf("[user_key_login_failed] uid=%d, reason=invalid key")
		return nil, "", errors.New("密钥错误")
	}

	// 更新最后登录时间
	config.DB.Model(&user).Update("lasttime", time.Now())

	token := s.genUserToken(user.UID, user.Key)

	return &user, token, nil
}

func (s *AuthService) genUserToken(uid uint, key string) string {
	hash := md5.Sum([]byte(fmt.Sprintf("%d%s%d", uid, key, time.Now().Unix())))
	return fmt.Sprintf("%d_%s", uid, hex.EncodeToString(hash[:]))
}

// 商户注册
func (s *AuthService) UserRegister(email, phone, password, inviteCode string, ip string) (*model.User, error) {
	// 检查注册是否开放
	var regOpen model.Config
	config.DB.First(&regOpen, "reg_open")
	if regOpen.V != "1" && regOpen.V != "2" {
		log.Printf("[user_register_failed] email=%s, phone=%s, reason=registration closed")
		return nil, errors.New("注册已关闭")
	}

	// 检查邮箱/手机是否已存在
	if email != "" {
		var count int64
		config.DB.Model(&model.User{}).Where("email = ?", email).Count(&count)
		if count > 0 {
			log.Printf("[user_register_failed] email=%s, reason=email already registered")
			return nil, errors.New("邮箱已被注册")
		}
	}

	if phone != "" {
		var count int64
		config.DB.Model(&model.User{}).Where("phone = ?", phone).Count(&count)
		if count > 0 {
			log.Printf("[user_register_failed] phone=%s, reason=phone already registered")
			return nil, errors.New("手机号已被注册")
		}
	}

	// 处理邀请码
	var upid uint
	if inviteCode != "" {
		var invite model.InviteCode
		result := config.DB.Where("code = ? AND status = 0", inviteCode).First(&invite)
		if result.Error == nil {
			upid = *invite.UID
		}
	}

	// 生成密钥
	key := s.genAPIKey()

	// 密码哈希
	pwdHash := md5.Sum([]byte(password + key))
	pwdStr := hex.EncodeToString(pwdHash[:])

	user := &model.User{
		GID:      1, // 默认用户组
		Upid:     upid,
		Key:      key,
		Pwd:      pwdStr,
		Account:  email,
		Email:    email,
		Phone:    phone,
		Money:    0,
		Cert:     0,
		Pay:      1,
		Settle:   1,
		Keylogin: 1,
		Apply:    1,
		Status:   0,
		Refund:   1,
		Transfer: 0,
		Keytype:  0,
		Addtime:  time.Now(),
		Lasttime: time.Now(),
	}

	result := config.DB.Create(user)
	if result.Error != nil {
		log.Printf("[user_register_failed] email=%s, phone=%s, reason=create user failed, error=%s", email, phone, result.Error.Error())
		return nil, errors.New("创建用户失败")
	}

	log.Printf("[user_register_success] uid=%d, email=%s, phone=%s, ip=%s", user.UID, email, phone, ip)
	return user, nil
}

func (s *AuthService) genAPIKey() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 32)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// 验证验证码
func (s *AuthService) VerifyCode(scene, to, code string) bool {
	var regcode model.RegCode
	result := config.DB.Where("scene = ? AND `to` = ? AND code = ? AND status = 0", scene, to, code).
		Order("time DESC").First(&regcode)

	if result.Error != nil {
		return false
	}

	// 检查是否过期（10分钟）
	if time.Now().Unix()-int64(regcode.Time) > 600 {
		return false
	}

	// 标记为已使用
	config.DB.Model(&regcode).Update("status", 1)

	return true
}

// 生成验证码
func (s *AuthService) GenCode(scene, to string) (string, error) {
	// 生成6位数字验证码
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 存储验证码
	regcode := &model.RegCode{
		Scene:  scene,
		To:     to,
		Code:   code,
		Time:   int(time.Now().Unix()),
		Status: 0,
	}

	result := config.DB.Create(regcode)
	if result.Error != nil {
		log.Printf("[gen_code_failed] scene=%s, to=%s, error=%s", scene, to, result.Error.Error())
		return "", result.Error
	}

	// TODO: 发送验证码（邮件/短信）

	return code, nil
}

// 获取配置
func (s *AuthService) GetConfig(k string) string {
	var cfg model.Config
	result := config.DB.First(&cfg, k)
	if result.Error != nil {
		return ""
	}
	return cfg.V
}

// 获取多个配置
func (s *AuthService) GetConfigs(keys []string) map[string]string {
	var cfgs []model.Config
	config.DB.Where("k IN ?", keys).Find(&cfgs)

	result := make(map[string]string)
	for _, c := range cfgs {
		result[c.K] = c.V
	}
	return result
}

// 保存配置
func (s *AuthService) SaveConfig(k, v string) error {
	var cfg model.Config
	result := config.DB.First(&cfg, k)
	if result.Error != nil {
		// 不存在则创建
		cfg = model.Config{K: k, V: v}
		return config.DB.Create(&cfg).Error
	}

	cfg.V = v
	return config.DB.Save(&cfg).Error
}

// 记录日志
func (s *AuthService) AddLog(uid uint, logType, data, ip string) {
	city := ""
	log := &model.Log{
		UID:  uid,
		Type: logType,
		Date: time.Now(),
		IP:   ip,
		City: city,
		Data: data,
	}
	config.DB.Create(log)
}

// 验证回调签名
func (s *AuthService) VerifySign(params map[string]string, sign string, key string) bool {
	// 构造签名字符串
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	// 按字典序排序
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	// 拼接
	var str string
	for _, k := range keys {
		if params[k] != "" && k != "sign" && k != "sign_type" {
			str += k + "=" + params[k] + "&"
		}
	}
	str += "key=" + key

	// MD5
	hash := md5.Sum([]byte(str))
	md5Str := strings.ToLower(hex.EncodeToString(hash[:]))

	return md5Str == sign
}

// 生成签名
func (s *AuthService) MakeSign(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	// 按字典序排序
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	var str string
	for _, k := range keys {
		if params[k] != "" && k != "sign" && k != "sign_type" {
			str += k + "=" + params[k] + "&"
		}
	}
	str += "key=" + key

	hash := md5.Sum([]byte(str))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

// 获取商户信息
func (s *AuthService) GetUser(uid uint) (*model.User, error) {
	var user model.User
	result := config.DB.First(&user, uid)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// 更新商户信息
func (s *AuthService) UpdateUser(uid uint, data map[string]interface{}) error {
	return config.DB.Model(&model.User{}).Where("uid = ?", uid).Updates(data).Error
}

// 获取商户配置
func (s *AuthService) GetUserSettings(uid uint) (map[string]string, error) {
	var user model.User
	result := config.DB.First(&user, uid)
	if result.Error != nil {
		return nil, result.Error
	}

	settings := make(map[string]string)
	if user.Channelinfo != "" {
		json.Unmarshal([]byte(user.Channelinfo), &settings)
	}

	return settings, nil
}
