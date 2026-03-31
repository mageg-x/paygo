package admin

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/service"

	"github.com/gin-gonic/gin"
)

// 管理员Handler
type AdminHandler struct {
	authSvc     *service.AuthService
	orderSvc    *service.OrderService
	settleSvc   *service.SettleService
	transferSvc *service.TransferService
	userSvc     *service.AuthService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		authSvc:     service.NewAuthService(),
		orderSvc:    service.NewOrderService(),
		settleSvc:   service.NewSettleService(),
		transferSvc: service.NewTransferService(),
		userSvc:     service.NewAuthService(),
	}
}

// 登录页面
func (h *AdminHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin/login.html", nil)
}

// 登录处理
func (h *AdminHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[管理员登录失败] IP: %s, 错误: 参数解析失败: %s", c.ClientIP(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	ip := c.ClientIP()

	token, err := h.authSvc.AdminLogin(req.Username, req.Password)
	if err != nil {
		log.Printf("[管理员登录失败] IP: %s, 用户名: %s, 错误: %s", ip, req.Username, err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	log.Printf("[管理员登录成功] IP: %s, 用户名: %s", ip, req.Username)
	c.SetCookie("admin_token", token, 86400*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "登录成功", "token": token})
}

// 登出
func (h *AdminHandler) Logout(c *gin.Context) {
	c.SetCookie("admin_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已退出"})
}

// 首页
func (h *AdminHandler) Index(c *gin.Context) {
	// 统计信息
	var orderCount, userCount int64
	var todayMoney float64

	config.DB.Model(&model.Order{}).Count(&orderCount)
	config.DB.Model(&model.User{}).Count(&userCount)

	today := time.Now().Format("2006-01-02")
	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", today).
		Select("COALESCE(SUM(money), 0)").Scan(&todayMoney)

	c.HTML(http.StatusOK, "admin/index.html", gin.H{
		"order_count": orderCount,
		"user_count":  userCount,
		"today_money": todayMoney,
	})
}

// 商户列表
func (h *AdminHandler) UserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	config.DB.Model(&model.User{}).Count(&total)
	config.DB.Offset(offset).Limit(pageSize).Order("uid DESC").Find(&users)

	c.HTML(http.StatusOK, "admin/ulist.html", gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// 订单列表
func (h *AdminHandler) OrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	offset := (page - 1) * pageSize

	status := c.DefaultQuery("status", "-1")
	tradeNo := c.Query("trade_no")

	var orders []model.Order
	var total int64

	query := config.DB.Model(&model.Order{})
	if status != "-1" {
		query = query.Where("status = ?", status)
	}
	if tradeNo != "" {
		query = query.Where("trade_no LIKE ?", "%"+tradeNo+"%")
	}

	query.Count(&total)
	query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders)

	c.HTML(http.StatusOK, "admin/order.html", gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
	})
}

// 结算列表
func (h *AdminHandler) SettleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var settles []model.Settle
	var total int64

	config.DB.Model(&model.Settle{}).Count(&total)
	config.DB.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&settles)

	c.HTML(http.StatusOK, "admin/settle.html", gin.H{
		"settles": settles,
		"total":   total,
		"page":    page,
	})
}

// 转账列表
func (h *AdminHandler) TransferList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var transfers []model.Transfer
	var total int64

	config.DB.Model(&model.Transfer{}).Count(&total)
	config.DB.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&transfers)

	c.HTML(http.StatusOK, "admin/transfer.html", gin.H{
		"transfers": transfers,
		"total":     total,
		"page":      page,
	})
}

// 系统设置页面
func (h *AdminHandler) Settings(c *gin.Context) {
	mod := c.DefaultQuery("mod", "site")

	// 加载所有配置
	var configs []model.Config
	config.DB.Find(&configs)

	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.K] = cfg.V
	}

	c.HTML(http.StatusOK, "admin/set.html", gin.H{
		"mod":          mod,
		"paygo/config": configMap,
	})
}

// 保存设置
func (h *AdminHandler) SaveSettings(c *gin.Context) {
	var req struct {
		Mod            string `json:"mod"`
		OldPwd         string `json:"old_pwd"`
		NewPwd         string `json:"new_pwd"`
		ConfirmPwd     string `json:"confirm_pwd"`
		Sitename       string `json:"sitename"`
		Localurl       string `json:"localurl"`
		Apiurl         string `json:"apiurl"`
		Kfqq           string `json:"kfqq"`
		RegOpen        string `json:"reg_open"`
		SettleMoney    string `json:"settle_money"`
		SettleAlipay   string `json:"settle_alipay"`
		SettleWxpay    string `json:"settle_wxpay"`
		TransferAlipay string `json:"transfer_alipay"`
		TransferWxpay  string `json:"transfer_wxpay"`
		LoginAlipay    string `json:"login_alipay"`
		LoginQq        string `json:"login_qq"`
		LoginWx        string `json:"login_wx"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	ip := c.ClientIP()

	if req.Mod == "account" {
		// 管理员密码修改
		if req.NewPwd != req.ConfirmPwd {
			log.Printf("[管理员修改密码失败] IP: %s, 错误: 两次密码不一致", ip)
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "两次密码不一致"})
			return
		}

		cfg := config.AppConfig
		if req.OldPwd != cfg.AdminPwd {
			log.Printf("[管理员修改密码失败] IP: %s, 错误: 原密码错误", ip)
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "原密码错误"})
			return
		}

		err := h.authSvc.SaveConfig("admin_pwd", req.NewPwd)
		if err != nil {
			log.Printf("[管理员修改密码失败] IP: %s, 错误: %s", ip, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
			return
		}

		cfg.AdminPwd = req.NewPwd
		log.Printf("[管理员修改密码成功] IP: %s", ip)

		// 生成新token并返回给前端
		newToken := generateAdminToken(cfg.AdminUser, cfg.AdminPwd, cfg.SysKey)
		// 同时设置cookie
		c.SetCookie("admin_token", newToken, 86400*30, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功", "token": newToken})
		return
	} else {
		// 其他配置保存
		cfgMap := map[string]string{
			"sitename":        req.Sitename,
			"localurl":        req.Localurl,
			"apiurl":          req.Apiurl,
			"kfqq":            req.Kfqq,
			"reg_open":        req.RegOpen,
			"settle_money":    req.SettleMoney,
			"settle_alipay":   req.SettleAlipay,
			"settle_wxpay":    req.SettleWxpay,
			"transfer_alipay": req.TransferAlipay,
			"transfer_wxpay":  req.TransferWxpay,
			"login_alipay":    req.LoginAlipay,
			"login_qq":        req.LoginQq,
			"login_wx":        req.LoginWx,
		}

		for k, v := range cfgMap {
			if v != "" {
				h.authSvc.SaveConfig(k, v)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功"})
}

// 通道管理
func (h *AdminHandler) ChannelList(c *gin.Context) {
	var channels []model.Channel
	config.DB.Find(&channels)

	c.HTML(http.StatusOK, "admin/channel.html", gin.H{
		"channels": channels,
	})
}

// 插件管理
func (h *AdminHandler) PluginList(c *gin.Context) {
	var plugins []model.Plugin
	config.DB.Find(&plugins)

	c.HTML(http.StatusOK, "admin/plugin.html", gin.H{
		"plugins": plugins,
	})
}

// AJAX: 获取订单列表
func (h *AdminHandler) AjaxOrderList(c *gin.Context) {
	page, _ := strconv.Atoi(c.PostForm("page"))
	pageSize, _ := strconv.Atoi(c.PostForm("limit"))
	status := c.PostForm("status")

	query := config.DB.Model(&model.Order{})
	if status != "" && status != "-1" {
		query = query.Where("status = ?", status)
	}

	var orders []model.Order
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  orders,
	})
}

// AJAX: 订单操作
func (h *AdminHandler) AjaxOrderOp(c *gin.Context) {
	action := c.PostForm("action")
	tradeNo := c.PostForm("trade_no")

	var err error
	switch action {
	case "refund":
		moneyStr := c.PostForm("money")
		money, _ := strconv.ParseFloat(moneyStr, 10)
		err = h.orderSvc.Refund(tradeNo, money)
	case "freeze":
		err = h.orderSvc.Freeze(tradeNo)
	case "unfreeze":
		err = h.orderSvc.Unfreeze(tradeNo)
	case "notify":
		// 重新通知
		err = nil
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 获取商户列表
func (h *AdminHandler) AjaxUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var users []model.User
	var total int64

	config.DB.Model(&model.User{}).Count(&total)
	config.DB.Offset(offset).Limit(pageSize).Order("uid DESC").Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  users,
	})
}

// AJAX: 商户操作
func (h *AdminHandler) AjaxUserOp(c *gin.Context) {
	action := c.PostForm("action")
	uid, _ := strconv.Atoi(c.PostForm("uid"))

	switch action {
	case "reset_key":
		// 重置密钥
		newKey := generateAPIKey()
		config.DB.Model(&model.User{}).Where("uid = ?", uid).Update("key", newKey)
		log.Printf("[admin_reset_key] uid=%d, new_key=%s", uid, newKey)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密钥已重置", "key": newKey})
		return
	case "set_status":
		status, _ := strconv.Atoi(c.PostForm("status"))
		config.DB.Model(&model.User{}).Where("uid = ?", uid).Update("status", status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return
	case "recharge":
		money, _ := strconv.ParseFloat(c.PostForm("money"), 10)
		typ := c.PostForm("type")
		err := h.transferSvc.AdminChangeMoney(uint(uid), money, typ, "管理员操作")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
		return
	case "delete":
		// 删除商户
		if uid <= 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "无效的商户ID"})
			return
		}
		result := config.DB.Delete(&model.User{}, "uid = ?", uid)
		if result.Error != nil {
			log.Printf("[admin_delete_user_failed] uid=%d, error=%s", uid, result.Error.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		log.Printf("[admin_delete_user_success] uid=%d", uid)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// 添加商户
func (h *AdminHandler) AddUser(c *gin.Context) {
	var req struct {
		GID      int     `json:"gid"`
		Phone    string  `json:"phone"`
		Email    string  `json:"email"`
		Pwd      string  `json:"pwd"`
		QQ       string  `json:"qq"`
		URL      string  `json:"url"`
		SettleID int     `json:"settle_id"`
		Account  string  `json:"account"`
		Username string  `json:"username"`
		Mode     int      `json:"mode"`
		Pay      int      `json:"pay"`
		Settle   int      `json:"settle"`
		Status   int      `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[add_user_failed] reason=invalid params, error=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	// 验证必填项
	if req.Account == "" || req.Username == "" {
		log.Printf("[add_user_failed] reason=account or username empty")
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "结算账号和姓名不能为空"})
		return
	}

	// 检查手机号是否已存在
	if req.Phone != "" {
		var count int64
		config.DB.Model(&model.User{}).Where("phone = ?", req.Phone).Count(&count)
		if count > 0 {
			log.Printf("[add_user_failed] phone=%s, reason=phone already exists", req.Phone)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "手机号已存在"})
			return
		}
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		var count int64
		config.DB.Model(&model.User{}).Where("email = ?", req.Email).Count(&count)
		if count > 0 {
			log.Printf("[add_user_failed] email=%s, reason=email already exists", req.Email)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "邮箱已存在"})
			return
		}
	}

	// 生成密钥
	key := generateAPIKey()

	// 默认用户组
	if req.GID == 0 {
		req.GID = 1
	}

	user := &model.User{
		GID:      uint(req.GID),
		Key:      key,
		SettleID: req.SettleID,
		Account:  req.Account,
		Username: req.Username,
		Money:    0,
		URL:      req.URL,
		Email:    req.Email,
		Qq:       req.QQ,
		Phone:    req.Phone,
		Mode:     req.Mode,
		Cert:     0,
		Pay:      req.Pay,
		Settle:   req.Settle,
		Status:   req.Status,
		Addtime:  time.Now(),
		Lasttime: time.Now(),
	}

	if err := config.DB.Create(user).Error; err != nil {
		log.Printf("[add_user_failed] reason=create user failed, error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "添加商户失败"})
		return
	}

	// 如果设置了密码
	if req.Pwd != "" {
		pwdHash := md5Hash(req.Pwd + key)
		config.DB.Model(user).Update("pwd", pwdHash)
	}

	log.Printf("[add_user_success] uid=%d, account=%s, username=%s", user.UID, req.Account, req.Username)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功", "uid": user.UID, "key": key})
}

// 编辑商户页面
func (h *AdminHandler) EditUserPage(c *gin.Context) {
	uid, _ := strconv.Atoi(c.Query("uid"))
	if uid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效的商户ID"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "商户不存在"})
		return
	}

	// 获取用户组列表
	var groups []model.Group
	config.DB.Find(&groups)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"user":   user,
		"groups": groups,
	})
}

// 更新商户
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	var req struct {
		UID      uint    `json:"uid"`
		GID      int     `json:"gid"`
		Phone    string  `json:"phone"`
		Email    string  `json:"email"`
		Pwd      string  `json:"pwd"`
		QQ       string  `json:"qq"`
		URL      string  `json:"url"`
		SettleID int     `json:"settle_id"`
		Account  string  `json:"account"`
		Username string  `json:"username"`
		Mode     int     `json:"mode"`
		Pay      int     `json:"pay"`
		Settle   int     `json:"settle"`
		Status   int     `json:"status"`
		Money    float64 `json:"money"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[update_user_failed] reason=invalid params, error=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, req.UID).Error; err != nil {
		log.Printf("[update_user_failed] uid=%d, reason=user not found", req.UID)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "商户不存在"})
		return
	}

	// 更新字段
	updates := map[string]interface{}{
		"gid":       req.GID,
		"settle_id": req.SettleID,
		"account":   req.Account,
		"username":  req.Username,
		"money":     req.Money,
		"url":       req.URL,
		"email":     req.Email,
		"qq":        req.QQ,
		"phone":     req.Phone,
		"mode":      req.Mode,
		"pay":       req.Pay,
		"settle":    req.Settle,
		"status":    req.Status,
	}

	// 如果设置了新密码
	if req.Pwd != "" {
		pwdHash := md5Hash(req.Pwd + user.Key)
		updates["pwd"] = pwdHash
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		log.Printf("[update_user_failed] uid=%d, error=%s", req.UID, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新失败"})
		return
	}

	log.Printf("[update_user_success] uid=%d", req.UID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// 辅助函数：生成随机密钥
func generateAPIKey() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 32)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

// 辅助函数：MD5哈希
func md5Hash(s string) string {
	h := md5.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}

// 辅助函数：生成管理员token
func generateAdminToken(username, password, sysKey string) string {
	hash := md5.Sum([]byte(username + password + password + sysKey))
	return fmt.Sprintf("%x", hash)
}

// AJAX: 结算操作
func (h *AdminHandler) AjaxSettleOp(c *gin.Context) {
	action := c.PostForm("action")
	id, _ := strconv.Atoi(c.PostForm("id"))

	var err error
	switch action {
	case "approve":
		err = h.settleSvc.ApproveSettle(uint(id))
	case "reject":
		reason := c.PostForm("reason")
		err = h.settleSvc.RejectSettle(uint(id), reason)
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// 统计API
func (h *AdminHandler) AjaxStats(c *gin.Context) {
	now := time.Now()
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	var todayOrderMoney, yesterdayOrderMoney float64
	var todayOrderCount, yesterdayOrderCount int64
	var userCount int64

	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", today).
		Select("COALESCE(SUM(money), 0)").Scan(&todayOrderMoney)
	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", today).
		Count(&todayOrderCount)

	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", yesterday).
		Select("COALESCE(SUM(money), 0)").Scan(&yesterdayOrderMoney)
	config.DB.Model(&model.Order{}).Where("date = ? AND status = 1", yesterday).
		Count(&yesterdayOrderCount)

	config.DB.Model(&model.User{}).Count(&userCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"today_order_money":     todayOrderMoney,
			"today_order_count":     todayOrderCount,
			"yesterday_order_money": yesterdayOrderMoney,
			"yesterday_order_count": yesterdayOrderCount,
			"user_count":            userCount,
		},
	})
}
