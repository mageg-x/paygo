package service

import (
	"crypto/hmac"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/middleware"
	"paygo/src/model"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// 管理员登录
func (s *AuthService) AdminLogin(username, password string) (string, error) {
	cfg := config.AppConfig
	if username != cfg.AdminUser {
		log.Printf("[admin_login_failed] username=%s, reason=invalid credentials", username)
		return "", errors.New("用户名或密码错误")
	}

	ok, needUpgrade := s.VerifyAdminPassword(password)
	if !ok {
		log.Printf("[admin_login_failed] username=%s, reason=invalid credentials", username)
		return "", errors.New("用户名或密码错误")
	}

	if needUpgrade {
		hashedPwd, err := s.HashAdminPassword(password)
		if err != nil {
			log.Printf("[admin_password_migrate_failed] username=%s, reason=hash failed, error=%s", username, err.Error())
			return "", errors.New("密码校验失败")
		}
		if err := s.SaveConfig("admin_pwd", hashedPwd); err != nil {
			log.Printf("[admin_password_migrate_failed] username=%s, reason=save failed, error=%s", username, err.Error())
			return "", errors.New("密码校验失败")
		}
		cfg.AdminPwd = hashedPwd
		log.Printf("[admin_password_migrated] username=%s", username)
	}

	token := s.genAdminToken(username, cfg.AdminPwd)
	return token, nil
}

func (s *AuthService) genAdminToken(username, password string) string {
	hash := md5.Sum([]byte(username + password + password + config.AppConfig.SysKey))
	return hex.EncodeToString(hash[:])
}

func (s *AuthService) GenAdminToken() string {
	cfg := config.AppConfig
	return s.genAdminToken(cfg.AdminUser, cfg.AdminPwd)
}

func isBcryptHash(v string) bool {
	v = strings.TrimSpace(v)
	return strings.HasPrefix(v, "$2a$") || strings.HasPrefix(v, "$2b$") || strings.HasPrefix(v, "$2y$")
}

func (s *AuthService) hashPassword(raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", errors.New("密码不能为空")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *AuthService) HashUserPassword(raw string) (string, error) {
	return s.hashPassword(raw)
}

func (s *AuthService) HashAdminPassword(raw string) (string, error) {
	return s.hashPassword(raw)
}

func (s *AuthService) VerifyUserPassword(stored, raw, userKey string) (ok bool, needUpgrade bool) {
	stored = strings.TrimSpace(stored)
	if stored == "" {
		return false, false
	}
	if isBcryptHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(raw)) == nil, false
	}

	legacy := md5.Sum([]byte(raw + userKey))
	legacyHex := hex.EncodeToString(legacy[:])
	if strings.EqualFold(stored, legacyHex) {
		return true, true
	}
	return false, false
}

func (s *AuthService) VerifyAdminPassword(input string) (ok bool, needUpgrade bool) {
	stored := strings.TrimSpace(config.AppConfig.AdminPwd)
	if stored == "" {
		return false, false
	}
	if isBcryptHash(stored) {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(input)) == nil, false
	}
	if input == stored {
		return true, true
	}
	return false, false
}

// 商户登录
func (s *AuthService) UserLogin(uid uint, pwd string) (*model.User, string, error) {
	var user model.User
	result := config.DB.First(&user, uid)
	if result.Error != nil {
		log.Printf("[user_login_failed] uid=%d, reason=user not found, error=%s", uid, result.Error.Error())
		return nil, "", errors.New("用户不存在")
	}

	if user.Status != 1 {
		log.Printf("[user_login_failed] uid=%d, reason=user disabled, status=%d", uid, user.Status)
		return nil, "", errors.New("账号已被禁用")
	}

	// 验证密码（兼容旧版MD5，登录后自动升级为bcrypt）
	ok, needUpgrade := s.VerifyUserPassword(user.Pwd, pwd, user.Key)
	if !ok {
		log.Printf("[user_login_failed] uid=%d, reason=invalid password")
		return nil, "", errors.New("密码错误")
	}
	if needUpgrade {
		newPwd, err := s.HashUserPassword(pwd)
		if err != nil {
			log.Printf("[user_login_failed] uid=%d, reason=upgrade hash failed, error=%s", uid, err.Error())
			return nil, "", errors.New("密码校验失败")
		}
		if err := config.DB.Model(&model.User{}).Where("uid = ?", user.UID).Update("pwd", newPwd).Error; err != nil {
			log.Printf("[user_login_failed] uid=%d, reason=upgrade save failed, error=%s", uid, err.Error())
			return nil, "", errors.New("密码校验失败")
		}
		user.Pwd = newPwd
		log.Printf("[user_password_migrated] uid=%d", uid)
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

	if user.Status != 1 {
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
	ts := time.Now().Unix()
	payload := fmt.Sprintf("%d.%d", uid, ts)
	mac := hmac.New(sha256.New, []byte(config.AppConfig.SysKey+"|"+key))
	mac.Write([]byte(payload))
	return fmt.Sprintf("%s.%s", payload, hex.EncodeToString(mac.Sum(nil)))
}

func (s *AuthService) GenUserToken(uid uint, key string) string {
	return s.genUserToken(uid, key)
}

// 商户注册
func (s *AuthService) UserRegister(email, phone, password, inviteCode string, ip string) (*model.User, error) {
	// 检查注册是否开放
	regOpen := s.GetConfig("reg_open")
	if regOpen != "1" && regOpen != "2" {
		log.Printf("[user_register_failed] email=%s, phone=%s, reason=registration closed")
		return nil, errors.New("注册已关闭")
	}

	// 仅邀请注册模式必须有邀请码
	if regOpen == "2" && inviteCode == "" {
		log.Printf("[user_register_failed] email=%s, phone=%s, reason=invite code required")
		return nil, errors.New("邀请码不能为空")
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
		err := config.DB.Where("code = ? AND status = 0", inviteCode).First(&invite).Error
		if err != nil {
			log.Printf("[user_register_failed] invite_code=%s, reason=invalid or used invite code")
			return nil, errors.New("邀请码无效或已使用")
		}
		upid = *invite.UID
		// 标记邀请码已使用
		config.DB.Model(&invite).Updates(map[string]interface{}{
			"status": 1,
			"uid":    0, // 临时，保存后会被替换
		})
	}

	// 生成密钥
	key := s.genAPIKey()

	// 密码哈希
	pwdStr, err := s.HashUserPassword(password)
	if err != nil {
		log.Printf("[user_register_failed] email=%s, phone=%s, reason=password hash failed, error=%s", email, phone, err.Error())
		return nil, errors.New("密码处理失败")
	}

	// 检查是否需要审核
	userReview := s.GetConfig("user_review")
	// user_review=1 表示需要审核，pay=2 表示待审核
	// user_review=0 表示不需要审核，pay=1 表示正常
	payStatus := 1
	if userReview == "1" {
		payStatus = 2 // 待审核
	}

	// 获取默认用户组
	defaultGID := uint(1)
	if dg := s.GetConfig("default_group"); dg != "" {
		if parsed, err := strconv.ParseUint(dg, 10, 32); err == nil {
			defaultGID = uint(parsed)
		}
	}

	user := &model.User{
		GID:      defaultGID, // 默认用户组
		Upid:     upid,
		Key:      key,
		Pwd:      pwdStr,
		Account:  email,
		Email:    email,
		Phone:    phone,
		Money:    0,
		Cert:     0,
		Pay:      payStatus, // 1=正常, 2=待审核
		Settle:   1,
		Keylogin: 1,
		Apply:    1,
		Status:   1, // 账户状态：1正常
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

	// 更新邀请码关联的UID
	if inviteCode != "" {
		config.DB.Model(&model.InviteCode{}).Where("code = ?", inviteCode).Update("uid", user.UID)
	}

	log.Printf("[user_register_success] uid=%d, email=%s, phone=%s, ip=%s, pay_status=%d", user.UID, email, phone, ip, payStatus)
	return user, nil
}

func (s *AuthService) genAPIKey() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 32)
	randomBytes := make([]byte, len(result))
	if _, err := crand.Read(randomBytes); err != nil {
		fallback := md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		for i := range result {
			randomBytes[i] = fallback[i%len(fallback)]
		}
	}
	for i := range result {
		result[i] = chars[int(randomBytes[i])%len(chars)]
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
	to = strings.TrimSpace(to)
	if to == "" {
		return "", errors.New("接收地址不能为空")
	}

	now := time.Now().Unix()

	// 发送频控：60秒内仅允许发送一次
	var last model.RegCode
	if err := config.DB.Where("scene = ? AND `to` = ?", scene, to).Order("time DESC").First(&last).Error; err == nil {
		if now-int64(last.Time) < 60 {
			return "", errors.New("发送过于频繁，请稍后重试")
		}
	}

	// 生成6位数字验证码（加密随机数）
	n, err := crand.Int(crand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", errors.New("验证码生成失败")
	}
	code := fmt.Sprintf("%06d", n.Int64())

	targetType := 0
	if isEmailAddress(to) {
		targetType = 1
	} else if isPhoneNumber(to) {
		targetType = 2
	} else {
		return "", errors.New("接收地址格式错误")
	}

	// 存储验证码
	regcode := &model.RegCode{
		Scene:  scene,
		Type:   targetType,
		To:     to,
		Code:   code,
		Time:   int(now),
		Status: 0,
	}

	if err := config.DB.Create(regcode).Error; err != nil {
		log.Printf("[gen_code_failed] scene=%s, to=%s, reason=save code failed, error=%s", scene, to, err.Error())
		return "", err
	}

	// 发送验证码；若发送失败，删除刚写入的验证码记录，避免无效验证码残留
	if err := s.sendCode(to, code); err != nil {
		config.DB.Delete(&model.RegCode{}, regcode.ID)
		log.Printf("[gen_code_failed] scene=%s, to=%s, reason=send failed, error=%s", scene, to, err.Error())
		return "", err
	}

	log.Printf("[gen_code_success] scene=%s, to=%s", scene, to)
	return code, nil
}

// 获取配置
func (s *AuthService) GetConfig(k string) string {
	var cfg model.Config
	result := config.DB.Where("k = ?", k).Limit(1).Find(&cfg)
	if result.Error != nil || result.RowsAffected == 0 {
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
	result := config.DB.Where("k = ?", k).First(&cfg)
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
	city := middleware.GetClientIPCity(ip)
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

func (s *AuthService) sendCode(to, code string) error {
	if isEmailAddress(to) {
		emailSvc := NewEmailService()
		return emailSvc.SendCode(to, code)
	}

	if isPhoneNumber(to) {
		smsSvc := GetSmsService()
		if smsSvc == nil {
			return errors.New("短信服务未启用")
		}
		return smsSvc.SendCode(to, code)
	}

	return errors.New("不支持的接收地址")
}

func isEmailAddress(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func isPhoneNumber(s string) bool {
	phoneRegexp := regexp.MustCompile(`^\d{6,20}$`)
	return phoneRegexp.MatchString(s)
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
