package admin

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
	"paygo/src/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		log.Printf("[admin_login_failed] ip=%s, reason=parse params error, error=%s", c.ClientIP(), err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	ip := c.ClientIP()

	token, err := h.authSvc.AdminLogin(req.Username, req.Password)
	if err != nil {
		log.Printf("[admin_login_failed] ip=%s, username=%s, error=%s", ip, req.Username, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	log.Printf("[admin_login_success] ip=%s, username=%s", ip, req.Username)
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
	query.Offset(offset).Limit(pageSize).Order("addtime DESC, trade_no DESC").Find(&orders)

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

// AJAX: 结算列表
func (h *AdminHandler) AjaxSettleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	status := c.DefaultQuery("status", "-1")

	query := config.DB.Model(&model.Settle{})
	if status != "-1" {
		query = query.Where("status = ?", status)
	}

	var settles []model.Settle
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&settles)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  settles,
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
		Mod        string `json:"mod"`
		OldPwd     string `json:"old_pwd"`
		NewPwd     string `json:"new_pwd"`
		ConfirmPwd string `json:"confirm_pwd"`
		// 网站设置
		Sitename         string `json:"sitename"`
		Title            string `json:"title"`
		Localurl         string `json:"localurl"`
		Apiurl           string `json:"apiurl"`
		Email            string `json:"email"`
		Kfqq             string `json:"kfqq"`
		RegOpen          string `json:"reg_open"`
		SiteKeywords     string `json:"site_keywords"`
		SiteDescription  string `json:"site_description"`
		CdnUrl           string `json:"cdn_url"`
		UserVerification string `json:"user_verification"`
		// 支付设置
		TestOpen       string `json:"test_open"`
		PaySuccessPage string `json:"pay_success_page"`
		PayErrorPage   string `json:"pay_error_page"`
		PayMinMoney    string `json:"pay_min_money"`
		PayMaxMoney    string `json:"pay_max_money"`
		PayBlockGoods  string `json:"pay_block_goods"`
		PayFeeRate     string `json:"pay_fee_rate"`
		InviteCashback string `json:"invite_cashback"`
		QrcodeEnabled  string `json:"qrcode_enabled"`
		// 结算设置
		SettleMoney        string `json:"settle_money"`
		SettleCycle        string `json:"settle_cycle"`
		SettleAlipay       string `json:"settle_alipay"`
		SettleWxpay        string `json:"settle_wxpay"`
		SettleAutoTransfer string `json:"settle_auto_transfer"`
		// 转账设置
		TransferMin      string `json:"transfer_min"`
		TransferMax      string `json:"transfer_max"`
		TransferFee      string `json:"transfer_fee"`
		TransferAlipay   string `json:"transfer_alipay"`
		TransferWxpay    string `json:"transfer_wxpay"`
		TransferShowName string `json:"transfer_show_name"`
		// 快捷登录
		LoginAlipay string `json:"login_alipay"`
		LoginQq     string `json:"login_qq"`
		LoginWx     string `json:"login_wx"`
		// 通知设置
		NotifyEmail string `json:"notify_email"`
		EmailNotify string `json:"email_notify"`
		OrderNotify string `json:"order_notify"`
		// 实名认证
		CertificateRequired string `json:"certificate_required"`
		CertificateTypes    string `json:"certificate_types"`
		// IP类型
		IpType string `json:"ip_type"`
		// 代理设置
		ProxyEnabled string `json:"proxy_enabled"`
		ProxyHost    string `json:"proxy_host"`
		ProxyPort    string `json:"proxy_port"`
		ProxyUser    string `json:"proxy_user"`
		ProxyPass    string `json:"proxy_pass"`
		// 邮件设置
		MailSmtpHost string `json:"mail_smtp_host"`
		MailSmtpPort string `json:"mail_smtp_port"`
		MailUsername string `json:"mail_username"`
		MailPassword string `json:"mail_password"`
		MailFrom     string `json:"mail_from"`
		// 短信设置
		SmsEnabled  string `json:"sms_enabled"`
		SmsProvider string `json:"sms_provider"`
		SmsAppid    string `json:"sms_appid"`
		SmsAppkey   string `json:"sms_appkey"`
		// 公告
		GonggaoContent string `json:"gonggao_content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	ip := c.ClientIP()

	if req.Mod == "account" {
		// 管理员密码修改
		if req.NewPwd != req.ConfirmPwd {
			log.Printf("[admin_change_password_failed] ip=%s, reason=password mismatch", ip)
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "两次密码不一致"})
			return
		}

		ok, _ := h.authSvc.VerifyAdminPassword(req.OldPwd)
		if !ok {
			log.Printf("[admin_change_password_failed] ip=%s, reason=old password incorrect", ip)
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "原密码错误"})
			return
		}

		hashedPwd, err := h.authSvc.HashAdminPassword(req.NewPwd)
		if err != nil {
			log.Printf("[admin_change_password_failed] ip=%s, reason=hash new password failed, error=%s", ip, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
			return
		}

		err = h.authSvc.SaveConfig("admin_pwd", hashedPwd)
		if err != nil {
			log.Printf("[admin_change_password_failed] ip=%s, error=%s", ip, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "保存失败"})
			return
		}

		config.AppConfig.AdminPwd = hashedPwd
		log.Printf("[admin_change_password_success] ip=%s", ip)

		// 生成新token并返回给前端
		newToken := h.authSvc.GenAdminToken()
		// 同时设置cookie
		c.SetCookie("admin_token", newToken, 86400*30, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功", "token": newToken})
		return
	}

	// 根据 mod 确定要保存的字段
	cfgMap := make(map[string]string)

	switch req.Mod {
	case "site":
		cfgMap["sitename"] = req.Sitename
		cfgMap["title"] = req.Title
		cfgMap["localurl"] = req.Localurl
		cfgMap["apiurl"] = req.Apiurl
		cfgMap["email"] = req.Email
		cfgMap["kfqq"] = req.Kfqq
		cfgMap["reg_open"] = req.RegOpen
		cfgMap["site_keywords"] = req.SiteKeywords
		cfgMap["site_description"] = req.SiteDescription
		cfgMap["cdn_url"] = req.CdnUrl
		cfgMap["user_verification"] = req.UserVerification
	case "pay":
		cfgMap["test_open"] = req.TestOpen
		cfgMap["pay_success_page"] = req.PaySuccessPage
		cfgMap["pay_error_page"] = req.PayErrorPage
		cfgMap["pay_min_money"] = req.PayMinMoney
		cfgMap["pay_max_money"] = req.PayMaxMoney
		cfgMap["pay_block_goods"] = req.PayBlockGoods
		cfgMap["pay_fee_rate"] = req.PayFeeRate
		cfgMap["invite_cashback"] = req.InviteCashback
		cfgMap["qrcode_enabled"] = req.QrcodeEnabled
	case "settle":
		cfgMap["settle_money"] = req.SettleMoney
		cfgMap["settle_cycle"] = req.SettleCycle
		cfgMap["settle_alipay"] = req.SettleAlipay
		cfgMap["settle_wxpay"] = req.SettleWxpay
		cfgMap["settle_auto_transfer"] = req.SettleAutoTransfer
	case "transfer":
		cfgMap["transfer_min"] = req.TransferMin
		cfgMap["transfer_max"] = req.TransferMax
		cfgMap["transfer_fee"] = req.TransferFee
		cfgMap["transfer_alipay"] = req.TransferAlipay
		cfgMap["transfer_wxpay"] = req.TransferWxpay
		cfgMap["transfer_show_name"] = req.TransferShowName
	case "oauth":
		cfgMap["login_alipay"] = req.LoginAlipay
		cfgMap["login_qq"] = req.LoginQq
		cfgMap["login_wx"] = req.LoginWx
	case "notice":
		cfgMap["notify_email"] = req.NotifyEmail
		cfgMap["email_notify"] = req.EmailNotify
		cfgMap["order_notify"] = req.OrderNotify
	case "certificate":
		cfgMap["certificate_required"] = req.CertificateRequired
		cfgMap["certificate_types"] = req.CertificateTypes
	case "iptype":
		cfgMap["ip_type"] = req.IpType
	case "proxy":
		cfgMap["proxy_enabled"] = req.ProxyEnabled
		cfgMap["proxy_host"] = req.ProxyHost
		cfgMap["proxy_port"] = req.ProxyPort
		cfgMap["proxy_user"] = req.ProxyUser
		cfgMap["proxy_pass"] = req.ProxyPass
	case "mail":
		cfgMap["mail_smtp_host"] = req.MailSmtpHost
		cfgMap["mail_smtp_port"] = req.MailSmtpPort
		cfgMap["mail_username"] = req.MailUsername
		cfgMap["mail_password"] = req.MailPassword
		cfgMap["mail_from"] = req.MailFrom
	case "sms":
		cfgMap["sms_enabled"] = req.SmsEnabled
		cfgMap["sms_provider"] = req.SmsProvider
		cfgMap["sms_appid"] = req.SmsAppid
		cfgMap["sms_appkey"] = req.SmsAppkey
	case "gonggao":
		cfgMap["gonggao_content"] = req.GonggaoContent
	}

	for k, v := range cfgMap {
		if v != "" {
			h.authSvc.SaveConfig(k, v)
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

// AJAX: 转账列表
func (h *AdminHandler) AjaxTransferList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	status := c.DefaultQuery("status", "-1")
	search := c.DefaultQuery("search", "")

	query := config.DB.Model(&model.Transfer{})
	if status != "-1" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("biz_no LIKE ? OR account LIKE ? OR username LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var transfers []model.Transfer
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&transfers)

	// 获取商户名称
	type TransferWithUser struct {
		model.Transfer
		UserName string `json:"user_name"`
	}
	result := make([]TransferWithUser, len(transfers))
	for i, t := range transfers {
		result[i] = TransferWithUser{Transfer: t}
		var user model.User
		if err := config.DB.First(&user, t.UID).Error; err == nil {
			result[i].UserName = user.Username
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  result,
	})
}

// AJAX: 转账操作
func (h *AdminHandler) AjaxTransferOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		BizNo  string `json:"biz_no"`
		Status int    `json:"status"`
		Result string `json:"result"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "query":
		_, err := h.transferSvc.QueryTransfer(req.BizNo)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询成功"})
		return
	case "set_status":
		err := h.transferSvc.UpdateTransferStatus(req.BizNo, req.Status, req.Result)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return
	case "delete":
		result := config.DB.Where("biz_no = ?", req.BizNo).Delete(&model.Transfer{})
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return
	case "refund":
		err := h.transferSvc.RefundTransfer(req.BizNo)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "退回成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 通道列表
func (h *AdminHandler) AjaxChannelList(c *gin.Context) {
	var channels []model.Channel
	config.DB.Find(&channels)

	// 获取内置插件信息
	builtInPlugins := plugin.GetAllPluginsInfo()
	builtInMap := make(map[string]plugin.PluginInfo)
	for _, p := range builtInPlugins {
		builtInMap[p.Name] = p
	}

	// 构建响应数据
	type ChannelResponse struct {
		ID             uint              `json:"id"`
		Mode           int               `json:"mode"`
		Type           int               `json:"type"`
		Plugin         string            `json:"plugin"`
		Name           string            `json:"name"`
		Rate           float64           `json:"rate"`
		Status         int               `json:"status"`
		Paymethod      string            `json:"paymethod"`
		PaymethodNames string            `json:"paymethod_names"` // 支付方式名称列表
		Daytop         int               `json:"daytop"`
		Daystatus      int               `json:"daystatus"`
		Paymin         string            `json:"paymin"`
		Paymax         string            `json:"paymax"`
		Appwxmp        int               `json:"appwxmp"`
		Appwxa         int               `json:"appwxa"`
		Costrate       float64           `json:"costrate"`
		Config         string            `json:"config"`
		PluginShowname string            `json:"plugin_showname"`
		PluginSelect   map[string]string `json:"plugin_select"` // 插件支持的支付方式
	}

	result := make([]ChannelResponse, len(channels))
	for i, ch := range channels {
		result[i] = ChannelResponse{
			ID:        ch.ID,
			Mode:      ch.Mode,
			Type:      ch.Type,
			Plugin:    ch.Plugin,
			Name:      ch.Name,
			Rate:      ch.Rate,
			Status:    ch.Status,
			Paymethod: ch.Paymethod,
			Daytop:    ch.Daytop,
			Daystatus: ch.Daystatus,
			Paymin:    ch.Paymin,
			Paymax:    ch.Paymax,
			Appwxmp:   ch.Appwxmp,
			Appwxa:    ch.Appwxa,
			Costrate:  ch.Costrate,
			Config:    ch.Config,
		}

		// 获取插件信息和支付方式名称
		var plugin model.Plugin
		var selectMap map[string]string
		if err := config.DB.First(&plugin, "name = ?", ch.Plugin).Error; err == nil {
			result[i].PluginShowname = plugin.Showname
			// 如果数据库插件有 config，尝试解析 Select
			if plugin.Config != "" {
				// config 可能是 JSON 格式的 select 映射
				// 这里简化处理，selectMap 为空时会用内置插件的
			}
		}
		// 使用内置插件的 Select 映射
		if bp, ok := builtInMap[ch.Plugin]; ok {
			result[i].PluginShowname = bp.Showname
			selectMap = bp.Select
			result[i].PluginSelect = selectMap
		}

		// 转换支付方式编号为名称
		if ch.Paymethod != "" && selectMap != nil {
			names := make([]string, 0)
			codes := strings.Split(ch.Paymethod, ",")
			for _, code := range codes {
				code = strings.TrimSpace(code)
				if name, ok := selectMap[code]; ok {
					names = append(names, name)
				}
			}
			namesText := strings.Join(names, ",")
			result[i].PaymethodNames = namesText
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "",
		"data": result,
	})
}

// AJAX: 通道操作
func (h *AdminHandler) AjaxChannelOp(c *gin.Context) {
	var req struct {
		Action    string  `json:"action"`
		ID        uint    `json:"id"`
		Name      string  `json:"name"`
		Plugin    string  `json:"plugin"`
		Type      int     `json:"type"`
		Mode      int     `json:"mode"`
		Rate      float64 `json:"rate"`
		Costrate  float64 `json:"costrate"`
		Daytop    int     `json:"daytop"`
		Paymin    float64 `json:"paymin"`
		Paymax    float64 `json:"paymax"`
		Paymethod string  `json:"paymethod"`
		Status    int     `json:"status"`
		Config    string  `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[channel_op_params_error] err=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add", "edit":
		log.Printf("[channel_op] action=%s, name=%s, plugin=%s, type=%d", req.Action, req.Name, req.Plugin, req.Type)
		channel := model.Channel{
			Name:      req.Name,
			Plugin:    req.Plugin,
			Type:      req.Type,
			Mode:      req.Mode,
			Rate:      req.Rate,
			Costrate:  req.Costrate,
			Daytop:    req.Daytop,
			Paymin:    strconv.FormatFloat(req.Paymin, 'f', 2, 64),
			Paymax:    strconv.FormatFloat(req.Paymax, 'f', 2, 64),
			Paymethod: strings.TrimSpace(req.Paymethod),
			Status:    req.Status,
			Config:    req.Config,
		}

		var err error
		if req.Action == "edit" {
			err = config.DB.Model(&model.Channel{}).Where("id = ?", req.ID).Updates(channel).Error
		} else {
			err = config.DB.Create(&channel).Error
		}

		if err != nil {
			log.Printf("[channel_op_%s_failed] id=%d, name=%s, error=%s", req.Action, req.ID, req.Name, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功"})
		return

	case "delete":
		result := config.DB.Delete(&model.Channel{}, "id = ?", req.ID)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return

	case "set_status":
		config.DB.Model(&model.Channel{}).Where("id = ?", req.ID).Update("status", req.Status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return

	case "get":
		var ch model.Channel
		if err := config.DB.First(&ch, req.ID).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "通道不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": ch})
		return

	case "get_plugins":
		// 获取某类型支持的插件列表
		var plugins []model.Plugin
		// 根据支付类型筛选插件
		config.DB.Find(&plugins)
		filtered := make([]model.Plugin, 0)
		for _, p := range plugins {
			// 简单判断：插件的types字段包含对应类型
			// 1=支付宝, 2=微信, 3=QQ, 4=银行卡
			switch req.Type {
			case 1:
				if len(p.Types) > 0 && (p.Types[0] == '1' || p.Types[0] == 'a' || p.Types[0] == 'A') {
					filtered = append(filtered, p)
				}
			case 2:
				if len(p.Types) > 0 && (p.Types[0] == '2' || p.Types[0] == 'w' || p.Types[0] == 'W') {
					filtered = append(filtered, p)
				}
			default:
				filtered = append(filtered, p)
			}
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": filtered})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 插件列表 - 合并内置插件信息 + 数据库状态/配置
func (h *AdminHandler) AjaxPluginList(c *gin.Context) {
	// 1. 获取所有内置插件信息
	builtInPlugins := plugin.GetAllPluginsInfo()

	// 2. 从数据库获取已保存的插件状态和配置
	var dbPlugins []model.Plugin
	config.DB.Find(&dbPlugins)
	dbPluginMap := make(map[string]model.Plugin)
	for _, p := range dbPlugins {
		dbPluginMap[p.Name] = p
	}

	// 3. 合并数据
	type PluginResponse struct {
		Name       string                        `json:"name"`
		Showname   string                        `json:"showname"`
		Author     string                        `json:"author"`
		Link       string                        `json:"link"`
		Types      string                        `json:"types"`
		Transtypes string                        `json:"transtypes"`
		Status     int                           `json:"status"`
		Config     string                        `json:"config"`
		Note       string                        `json:"note"` // 内置插件有说明
		IsBuiltIn  bool                          `json:"is_builtin"`
		Select     map[string]string             `json:"select"` // 支付方式选择
		Inputs     map[string]plugin.InputConfig `json:"inputs"` // 配置字段定义
	}

	result := make([]PluginResponse, 0, len(builtInPlugins))
	for _, p := range builtInPlugins {
		dbPlugin, exists := dbPluginMap[p.Name]
		status := 1 // 默认启用
		cfg := "{}"
		if exists {
			status = dbPlugin.Status
			if dbPlugin.Config != "" {
				cfg = dbPlugin.Config
			}
		}
		result = append(result, PluginResponse{
			Name:       p.Name,
			Showname:   p.Showname,
			Author:     p.Author,
			Link:       p.Link,
			Types:      strings.Join(p.Types, ","),
			Transtypes: strings.Join(p.Transtypes, ","),
			Status:     status,
			Config:     cfg,
			Note:       p.Note,
			IsBuiltIn:  true,
			Select:     p.Select,
			Inputs:     p.Inputs,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "",
		"data": result,
	})
}

// AJAX: 插件操作
func (h *AdminHandler) AjaxPluginOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		Name   string `json:"name"`
		Status int    `json:"status"`
		Config string `json:"config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "refresh":
		builtInPlugins := plugin.GetAllPluginsInfo()
		synced := 0
		for _, p := range builtInPlugins {
			var existing model.Plugin
			err := config.DB.First(&existing, "name = ?", p.Name).Error
			if err != nil {
				newPlugin := model.Plugin{
					Name:       p.Name,
					Showname:   p.Showname,
					Author:     p.Author,
					Link:       p.Link,
					Types:      strings.Join(p.Types, ","),
					Transtypes: strings.Join(p.Transtypes, ","),
					Status:     1,
					Config:     "{}",
				}
				config.DB.Create(&newPlugin)
				synced++
			}
		}
		log.Printf("[plugin_refresh_success] synced=%d", synced)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "刷新成功"})
		return

	case "set_status":
		result := config.DB.Model(&model.Plugin{}).Where("name = ?", req.Name).Update("status", req.Status)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新失败"})
			return
		}
		log.Printf("[plugin_set_status] name=%s, status=%d", req.Name, req.Status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return

	case "get_config":
		var dbPlugin model.Plugin
		if err := config.DB.First(&dbPlugin, "name = ?", req.Name).Error; err != nil {
			builtIn := plugin.GetHandler(req.Name)
			if builtIn == nil {
				c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "插件不存在"})
				return
			}
			info := builtIn.GetInfo()
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"data": gin.H{
					"name":       req.Name,
					"showname":   info.Showname,
					"author":     info.Author,
					"types":      strings.Join(info.Types, ","),
					"transtypes": strings.Join(info.Transtypes, ","),
					"inputs":     info.Inputs,
					"config":     "{}",
				},
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": dbPlugin})
		return

	case "save_config":
		cfg := req.Config
		if cfg == "" {
			cfg = "{}"
		}

		// 在配置保存现场做校验，避免“测试通过但下单失败”的配置偏差
		if cfg != "{}" {
			p := plugin.GetHandler(req.Name)
			if p == nil {
				c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "插件不存在"})
				return
			}
			if tester, ok := p.(interface {
				TestConfig(config string) (bool, string)
			}); ok {
				success, msg := tester.TestConfig(cfg)
				if !success {
					c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "配置校验失败: " + msg})
					return
				}
			}
		}

		result := config.DB.Model(&model.Plugin{}).Where("name = ?", req.Name).Update("config", cfg)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存失败"})
			return
		}
		log.Printf("[plugin_save_config] name=%s", req.Name)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "配置已保存"})
		return

	case "test_config":
		p := plugin.GetHandler(req.Name)
		if p == nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "插件不存在"})
			return
		}

		tester, ok := p.(interface {
			TestConfig(config string) (bool, string)
		})
		if !ok {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "该插件不支持配置测试"})
			return
		}

		success, msg := tester.TestConfig(req.Config)
		if success {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": msg})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": msg})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 获取配置
func (h *AdminHandler) AjaxGetConfig(c *gin.Context) {
	var configs []model.Config
	config.DB.Find(&configs)

	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.K] = cfg.V
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": configMap,
	})
}

// AJAX: 根据key数组获取配置
func (h *AdminHandler) AjaxGetSettings(c *gin.Context) {
	keysStr := c.Query("keys")
	if keysStr == "" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "keys参数不能为空"})
		return
	}

	keys := strings.Split(keysStr, ",")
	var configs []model.Config
	config.DB.Where("k IN ?", keys).Find(&configs)

	configMap := make(map[string]string)
	for _, cfg := range configs {
		configMap[cfg.K] = cfg.V
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": configMap,
	})
}

// AJAX: 上传微信客服二维码（base64图片数据）
func (h *AdminHandler) UploadWxkfQrcode(c *gin.Context) {
	var req struct {
		Data string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	raw := strings.TrimSpace(req.Data)
	if raw == "" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "图片不能为空"})
		return
	}

	commaIdx := strings.Index(raw, ",")
	if commaIdx <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "图片格式错误"})
		return
	}
	prefix := raw[:commaIdx]
	body := raw[commaIdx+1:]

	ext := ".png"
	if strings.Contains(prefix, "image/jpeg") || strings.Contains(prefix, "image/jpg") {
		ext = ".jpg"
	} else if strings.Contains(prefix, "image/webp") {
		ext = ".webp"
	} else if !strings.Contains(prefix, "image/png") {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "仅支持PNG/JPG/WEBP图片"})
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "图片解码失败"})
		return
	}
	if len(decoded) > 2*1024*1024 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "图片大小不能超过2MB"})
		return
	}

	dir := "uploads/wxkf"
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "创建目录失败"})
		return
	}

	filename := fmt.Sprintf("wxkf_%d%s", time.Now().UnixNano(), ext)
	path := fmt.Sprintf("%s/%s", dir, filename)
	if err := os.WriteFile(path, decoded, 0644); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存图片失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"path": "/" + path,
		},
	})
}

// AJAX: 邀请码列表
func (h *AdminHandler) AjaxInviteCodeList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	search := c.Query("search")

	query := config.DB.Model(&model.InviteCode{})
	if search != "" {
		query = query.Where("code LIKE ?", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var list []model.InviteCode
	query.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  list,
	})
}

// AJAX: 生成邀请码
func (h *AdminHandler) AjaxInviteCodeGenerate(c *gin.Context) {
	var req struct {
		Num int `json:"num"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	if req.Num <= 0 || req.Num > 100 {
		req.Num = 1
	}

	codes := make([]string, 0, req.Num)
	for i := 0; i < req.Num; i++ {
		created := false
		for attempt := 0; attempt < 8; attempt++ {
			code, err := generateInviteCode()
			if err != nil {
				log.Printf("[invite_code_generate_failed] reason=random failed, error=%s", err.Error())
				break
			}
			var exists int64
			config.DB.Model(&model.InviteCode{}).Where("code = ?", code).Count(&exists)
			if exists > 0 {
				continue
			}

			invite := &model.InviteCode{
				Code:    code,
				Addtime: time.Now(),
				Status:  0,
			}
			if err := config.DB.Create(invite).Error; err == nil {
				codes = append(codes, code)
				created = true
				break
			}
		}
		if !created {
			log.Printf("[invite_code_generate_failed] reason=exhausted retries, index=%d", i)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "成功生成 " + strconv.Itoa(len(codes)) + " 个邀请码",
		"codes": codes,
	})
}

// AJAX: 删除邀请码
func (h *AdminHandler) AjaxInviteCodeDelete(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	config.DB.Delete(&model.InviteCode{}, req.ID)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
}

// 生成邀请码
func generateInviteCode() (string, error) {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	result := make([]byte, 8)
	randomBytes := make([]byte, len(result))
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	for i := range result {
		result[i] = chars[int(randomBytes[i])%len(chars)]
	}
	return string(result), nil
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
	page := adminIntParam(c, "page", 1)
	pageSize := adminIntParam(c, "limit", 20)
	status := adminStringParam(c, "status")
	uid := adminIntParam(c, "uid", 0)
	payType := adminIntParam(c, "type", 0)
	tradeNo := adminStringParam(c, "trade_no")
	startDate := adminStringParam(c, "start_date")
	endDate := adminStringParam(c, "end_date")

	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	query := config.DB.Model(&model.Order{})
	if status != "" && status != "-1" {
		query = query.Where("status = ?", status)
	}
	if uid > 0 {
		query = query.Where("uid = ?", uid)
	}
	if payType > 0 {
		query = query.Where("type = ?", payType)
	}
	if tradeNo != "" {
		query = query.Where("(trade_no LIKE ? OR out_trade_no LIKE ?)", "%"+tradeNo+"%", "%"+tradeNo+"%")
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("addtime >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("addtime < ?", t.AddDate(0, 0, 1))
		}
	}

	var orders []model.Order
	var total int64
	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("addtime DESC").Find(&orders)

	// 组装支付类型名称，避免前端显示“未知”
	typeNameMap := make(map[int]string)
	var payTypes []model.PayType
	if err := config.DB.Find(&payTypes).Error; err == nil {
		for _, pt := range payTypes {
			name := strings.TrimSpace(pt.Showname)
			if name == "" {
				name = strings.TrimSpace(pt.Name)
			}
			if name != "" {
				typeNameMap[int(pt.ID)] = name
			}
		}
	}

	channelPluginMap := make(map[int]string)
	channelIDs := make([]int, 0, len(orders))
	channelIDSeen := make(map[int]struct{}, len(orders))
	for _, o := range orders {
		if o.Channel <= 0 {
			continue
		}
		if _, ok := channelIDSeen[o.Channel]; ok {
			continue
		}
		channelIDSeen[o.Channel] = struct{}{}
		channelIDs = append(channelIDs, o.Channel)
	}
	if len(channelIDs) > 0 {
		var channels []model.Channel
		if err := config.DB.Where("id IN ?", channelIDs).Find(&channels).Error; err == nil {
			for _, ch := range channels {
				channelPluginMap[int(ch.ID)] = strings.TrimSpace(ch.Plugin)
			}
		}
	}

	fallbackTypeNameByPlugin := func(pluginName string) string {
		switch strings.ToLower(strings.TrimSpace(pluginName)) {
		case "alipay":
			return "支付宝"
		case "wxpay":
			return "微信支付"
		default:
			return ""
		}
	}

	type orderWithTypeName struct {
		model.Order
		Typename string `json:"typename"`
	}

	respData := make([]orderWithTypeName, 0, len(orders))
	for _, o := range orders {
		typeName := typeNameMap[o.Type]
		if typeName == "" {
			typeName = fallbackTypeNameByPlugin(channelPluginMap[o.Channel])
		}
		if typeName == "" {
			typeName = fmt.Sprintf("类型%d", o.Type)
		}
		respData = append(respData, orderWithTypeName{
			Order:    o,
			Typename: typeName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  respData,
	})
}

// AJAX: 订单操作
func (h *AdminHandler) AjaxOrderOp(c *gin.Context) {
	var req struct {
		Action  string  `json:"action"`
		TradeNo string  `json:"trade_no"`
		Money   float64 `json:"money"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var err error
	successMsg := "操作成功"
	switch req.Action {
	case "refund":
		if req.Money <= 0 {
			order, e := h.orderSvc.GetOrder(req.TradeNo)
			if e != nil {
				err = e
				break
			}
			available := order.Money - order.Refundmoney
			if available <= 0 {
				err = fmt.Errorf("可退款金额为0")
				break
			}
			req.Money = available
		}
		err = h.orderSvc.Refund(req.TradeNo, req.Money)
	case "freeze":
		err = h.orderSvc.Freeze(req.TradeNo)
	case "unfreeze":
		err = h.orderSvc.Unfreeze(req.TradeNo)
	case "notify":
		err = h.orderSvc.RetryNotify(req.TradeNo)
	case "refresh":
		outcome, e := service.RefreshOrderStatus(req.TradeNo)
		if e != nil {
			err = e
			break
		}
		statusText := strings.TrimSpace(outcome.Status)
		if statusText == "" {
			statusText = "UNKNOWN"
		}
		if outcome.Filled {
			successMsg = fmt.Sprintf("刷新完成：上游已支付，订单已更新（%s）", statusText)
		} else {
			successMsg = fmt.Sprintf("刷新完成：上游状态 %s（未支付）", statusText)
		}
	}

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": successMsg})
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
	var req struct {
		Action string  `json:"action"`
		UID    uint    `json:"uid"`
		Status int     `json:"status"`
		Money  float64 `json:"money"`
		Type   string  `json:"type"`
		Key    string  `json:"key"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "reset_key":
		if req.UID == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "无效的商户ID"})
			return
		}
		newKey := generateAPIKey()
		result := config.DB.Model(&model.User{}).Where("uid = ?", req.UID).Update("key", newKey)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "重置失败"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "商户不存在"})
			return
		}
		log.Printf("[admin_reset_key] uid=%d, new_key=%s", req.UID, newKey)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密钥已重置（旧密码登录将失效）", "key": newKey})
		return
	case "set_key":
		if req.UID == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "无效的商户ID"})
			return
		}
		newKey := strings.TrimSpace(req.Key)
		if newKey == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密钥不能为空"})
			return
		}
		if len(newKey) < 8 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密钥长度至少8位"})
			return
		}
		result := config.DB.Model(&model.User{}).Where("uid = ?", req.UID).Update("key", newKey)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "修改失败"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "商户不存在"})
			return
		}
		log.Printf("[admin_set_key] uid=%d", req.UID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密钥已修改（旧密码登录将失效）", "key": newKey})
		return
	case "set_status":
		config.DB.Model(&model.User{}).Where("uid = ?", req.UID).Update("status", req.Status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return
	case "recharge":
		err := h.transferSvc.AdminChangeMoney(req.UID, req.Money, req.Type, "管理员操作")
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
		return
	case "delete":
		if req.UID <= 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "无效的商户ID"})
			return
		}
		result := config.DB.Delete(&model.User{}, "uid = ?", req.UID)
		if result.Error != nil {
			log.Printf("[admin_delete_user_failed] uid=%d, error=%s", req.UID, result.Error.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		log.Printf("[admin_delete_user_success] uid=%d", req.UID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// 添加商户
func (h *AdminHandler) AddUser(c *gin.Context) {
	var req struct {
		GID      int    `json:"gid"`
		Phone    string `json:"phone"`
		Email    string `json:"email"`
		Pwd      string `json:"pwd"`
		QQ       string `json:"qq"`
		URL      string `json:"url"`
		SettleID int    `json:"settle_id"`
		Account  string `json:"account"`
		Username string `json:"username"`
		Mode     int    `json:"mode"`
		Pay      int    `json:"pay"`
		Settle   int    `json:"settle"`
		Status   int    `json:"status"`
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
		pwdHash, err := h.authSvc.HashUserPassword(req.Pwd)
		if err != nil {
			log.Printf("[add_user_failed] uid=%d, reason=password hash failed, error=%s", user.UID, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密码处理失败"})
			return
		}
		if err := config.DB.Model(user).Update("pwd", pwdHash).Error; err != nil {
			log.Printf("[add_user_failed] uid=%d, reason=save password failed, error=%s", user.UID, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "设置密码失败"})
			return
		}
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
		"code":   0,
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
		pwdHash, err := h.authSvc.HashUserPassword(req.Pwd)
		if err != nil {
			log.Printf("[update_user_failed] uid=%d, reason=password hash failed, error=%s", req.UID, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密码处理失败"})
			return
		}
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
	randomBytes := make([]byte, len(result))
	if _, err := rand.Read(randomBytes); err != nil {
		now := time.Now().UnixNano()
		for i := range result {
			randomBytes[i] = byte((now >> (uint(i%8) * 8)) & 0xff)
		}
	}
	for i := range result {
		result[i] = chars[int(randomBytes[i])%len(chars)]
	}
	return string(result)
}

// AJAX: 结算操作
func (h *AdminHandler) AjaxSettleOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		ID     uint   `json:"id"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var err error
	switch req.Action {
	case "approve":
		err = h.settleSvc.ApproveSettle(req.ID)
	case "reject":
		err = h.settleSvc.RejectSettle(req.ID, req.Reason)
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
	loc := now.Location()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tomorrowStart := todayStart.AddDate(0, 0, 1)
	yesterdayStart := todayStart.AddDate(0, 0, -1)

	var todayOrderMoney, yesterdayOrderMoney float64
	var todayOrderCount, yesterdayOrderCount int64
	var userCount int64
	paidLikeStatuses := []int{model.OrderStatusPaid, model.OrderStatusRefunded}

	// 订单数：统计当天/昨日的全部订单；交易额：统计已支付+已退款订单金额
	config.DB.Model(&model.Order{}).Where("addtime >= ? AND addtime < ?", todayStart, tomorrowStart).
		Count(&todayOrderCount)
	config.DB.Model(&model.Order{}).Where("addtime >= ? AND addtime < ? AND status IN ?", todayStart, tomorrowStart, paidLikeStatuses).
		Select("COALESCE(SUM(money), 0)").Scan(&todayOrderMoney)

	config.DB.Model(&model.Order{}).Where("addtime >= ? AND addtime < ?", yesterdayStart, todayStart).
		Count(&yesterdayOrderCount)
	config.DB.Model(&model.Order{}).Where("addtime >= ? AND addtime < ? AND status IN ?", yesterdayStart, todayStart, paidLikeStatuses).
		Select("COALESCE(SUM(money), 0)").Scan(&yesterdayOrderMoney)

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

// AJAX: 风控记录列表
func (h *AdminHandler) AjaxRiskList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	uid := c.DefaultQuery("uid", "")
	uidInt, _ := strconv.Atoi(uid)

	query := config.DB.Model(&model.Risk{})
	if uidInt > 0 {
		query = query.Where("uid = ?", uidInt)
	}

	var total int64
	query.Count(&total)

	var list []model.Risk
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list)

	// 获取商户名称
	type RiskWithUser struct {
		model.Risk
		UserName string `json:"user_name"`
	}
	result := make([]RiskWithUser, len(list))
	for i, r := range list {
		result[i] = RiskWithUser{Risk: r}
		var user model.User
		if err := config.DB.First(&user, r.UID).Error; err == nil {
			result[i].UserName = user.Username
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  result,
	})
}

// AJAX: 风控操作
func (h *AdminHandler) AjaxRiskOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		ID     uint   `json:"id"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "set_status":
		config.DB.Model(&model.Risk{}).Where("id = ?", req.ID).Update("status", req.Status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 黑名单列表
func (h *AdminHandler) AjaxBlacklistList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	blackType := c.DefaultQuery("type", "")

	query := config.DB.Model(&model.Blacklist{})
	if blackType != "" {
		query = query.Where("type = ?", blackType)
	}

	var total int64
	query.Count(&total)

	var list []model.Blacklist
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  list,
	})
}

// AJAX: 黑名单操作
func (h *AdminHandler) AjaxBlacklistOp(c *gin.Context) {
	var req struct {
		Action  string `json:"action"`
		ID      uint   `json:"id"`
		Type    int    `json:"type"`
		Content string `json:"content"`
		Remark  string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add":
		black := &model.Blacklist{
			Type:    req.Type,
			Content: req.Content,
			Remark:  req.Remark,
			Addtime: time.Now(),
		}
		if err := config.DB.Create(black).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "添加失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功"})
		return

	case "delete":
		config.DB.Delete(&model.Blacklist{}, req.ID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 域名授权列表
func (h *AdminHandler) AjaxDomainList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	uid := c.DefaultQuery("uid", "")
	uidInt, _ := strconv.Atoi(uid)

	query := config.DB.Model(&model.Domain{})
	if uidInt > 0 {
		query = query.Where("uid = ?", uidInt)
	}

	var total int64
	query.Count(&total)

	var list []model.Domain
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list)

	// 获取商户名称
	type DomainWithUser struct {
		model.Domain
		UserName string `json:"user_name"`
	}
	result := make([]DomainWithUser, len(list))
	for i, d := range list {
		result[i] = DomainWithUser{Domain: d}
		var user model.User
		if err := config.DB.First(&user, d.UID).Error; err == nil {
			result[i].UserName = user.Username
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  result,
	})
}

// AJAX: 域名授权操作
func (h *AdminHandler) AjaxDomainOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		ID     uint   `json:"id"`
		UID    uint   `json:"uid"`
		Domain string `json:"domain"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add":
		domain := &model.Domain{
			UID:     req.UID,
			Domain:  req.Domain,
			Status:  1,
			Addtime: func() *time.Time { t := time.Now(); return &t }(),
		}
		if err := config.DB.Create(domain).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "添加失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功"})
		return

	case "delete":
		config.DB.Delete(&model.Domain{}, req.ID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return

	case "set_status":
		config.DB.Model(&model.Domain{}).Where("id = ?", req.ID).Update("status", req.Status)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 公告列表
func (h *AdminHandler) AjaxAnounceList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}

	var total int64
	config.DB.Model(&model.Anounce{}).Count(&total)

	var list []model.Anounce
	config.DB.Offset((page - 1) * pageSize).Limit(pageSize).Order("sort DESC, id DESC").Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  list,
	})
}

// AJAX: 公告操作
func (h *AdminHandler) AjaxAnounceOp(c *gin.Context) {
	var req struct {
		Action  string `json:"action"`
		ID      uint   `json:"id"`
		Content string `json:"content"`
		Color   string `json:"color"`
		Sort    int    `json:"sort"`
		Status  int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add":
		anounce := &model.Anounce{
			Content: req.Content,
			Color:   req.Color,
			Sort:    req.Sort,
			Status:  req.Status,
			Addtime: time.Now(),
		}
		if err := config.DB.Create(anounce).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "添加失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功"})
		return

	case "edit":
		updates := map[string]interface{}{
			"content": req.Content,
			"color":   req.Color,
			"sort":    req.Sort,
			"status":  req.Status,
		}
		if err := config.DB.Model(&model.Anounce{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
		return

	case "delete":
		config.DB.Delete(&model.Anounce{}, req.ID)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

// AJAX: 操作日志列表
func (h *AdminHandler) AjaxLogList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	uid := c.DefaultQuery("uid", "")
	uidInt, _ := strconv.Atoi(uid)

	query := config.DB.Model(&model.Log{})
	if uidInt > 0 {
		query = query.Where("uid = ?", uidInt)
	}

	var total int64
	query.Count(&total)

	var list []model.Log
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list)

	// 获取商户名称
	type LogWithUser struct {
		model.Log
		UserName string `json:"user_name"`
	}
	result := make([]LogWithUser, len(list))
	for i, l := range list {
		result[i] = LogWithUser{Log: l}
		var user model.User
		if err := config.DB.First(&user, l.UID).Error; err == nil {
			result[i].UserName = user.Username
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  result,
	})
}

// AJAX: SSO单点登录（管理员登录商户账号）
func (h *AdminHandler) AjaxSSOLogin(c *gin.Context) {
	var req struct {
		UID uint `json:"uid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var user model.User
	if err := config.DB.First(&user, req.UID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "商户不存在"})
		return
	}

	// 生成商户token
	token := h.authSvc.GenUserToken(user.UID, user.Key)

	log.Printf("[sso_login_success] uid=%d", req.UID)
	h.saveSSORecent(user)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "登录成功",
		"token": token,
		"uid":   user.UID,
	})
}

// AJAX: SSO最近登录记录
func (h *AdminHandler) AjaxSSORecent(c *gin.Context) {
	list := h.loadSSORecent()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "",
		"data": list,
	})
}

// AJAX: SSO最近登录记录操作
func (h *AdminHandler) AjaxSSORecentOp(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required"` // clear, remove
		UID    uint   `json:"uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "clear":
		_ = config.Set("admin_sso_recent", "[]")
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已清空"})
		return
	case "remove":
		if req.UID == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "无效的商户ID"})
			return
		}
		list := h.loadSSORecent()
		filtered := make([]map[string]interface{}, 0, len(list))
		for _, it := range list {
			if uidFloat, ok := it["uid"].(float64); ok && uint(uidFloat) == req.UID {
				continue
			}
			filtered = append(filtered, it)
		}
		b, _ := json.Marshal(filtered)
		_ = config.Set("admin_sso_recent", string(b))
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已删除"})
		return
	default:
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
		return
	}
}

func (h *AdminHandler) loadSSORecent() []map[string]interface{} {
	raw := strings.TrimSpace(config.Get("admin_sso_recent"))
	if raw == "" {
		return make([]map[string]interface{}, 0)
	}
	var list []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		log.Printf("[sso_recent_parse_failed] raw=%s, error=%s", raw, err.Error())
		return make([]map[string]interface{}, 0)
	}
	return list
}

func (h *AdminHandler) saveSSORecent(user model.User) {
	list := h.loadSSORecent()
	next := make([]map[string]interface{}, 0, 10)

	item := map[string]interface{}{
		"uid":      user.UID,
		"username": user.Username,
		"account":  user.Account,
		"time":     time.Now().Format("2006-01-02 15:04:05"),
	}
	next = append(next, item)

	for _, it := range list {
		uidFloat, ok := it["uid"].(float64)
		if ok && uint(uidFloat) == user.UID {
			continue
		}
		next = append(next, it)
		if len(next) >= 10 {
			break
		}
	}

	b, _ := json.Marshal(next)
	if err := config.Set("admin_sso_recent", string(b)); err != nil {
		log.Printf("[sso_recent_save_failed] uid=%d, error=%s", user.UID, err.Error())
	}
}

// AJAX: 获取计划任务列表
func (h *AdminHandler) AjaxCronList(c *gin.Context) {
	cronSvc := service.GetCronService()
	tasks := cronSvc.ListTasks()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "",
		"data": tasks,
	})
}

// AJAX: 计划任务操作
func (h *AdminHandler) AjaxCronOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"` // get, set, run
		Name   string `json:"name"`
		Enable *bool  `json:"enable"`
		Spec   string `json:"spec"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	cronSvc := service.GetCronService()

	switch req.Action {
	case "get":
		// 获取任务详情
		tasks := cronSvc.ListTasks()
		for _, t := range tasks {
			if t["name"] == req.Name {
				c.JSON(http.StatusOK, gin.H{"code": 0, "data": t})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "任务不存在"})

	case "set":
		// 设置任务开关
		key := fmt.Sprintf("cron_%s", req.Name)
		if req.Enable != nil {
			if *req.Enable {
				config.Set(key, "1")
				// 重新添加任务
				spec := config.Get(key + "_spec")
				if spec != "" {
					cronSvc.AddTask(req.Name, spec, getCronTask(req.Name))
				}
			} else {
				config.Set(key, "0")
				cronSvc.RemoveTask(req.Name)
			}
		}
		// 更新执行周期
		if req.Spec != "" {
			config.Set(key+"_spec", req.Spec)
			cronSvc.RemoveTask(req.Name)
			if config.Get(key) == "1" {
				cronSvc.AddTask(req.Name, req.Spec, getCronTask(req.Name))
			}
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "设置成功"})

	case "run":
		// 立即执行一次
		go func() {
			getCronTask(req.Name)()
		}()
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "任务已触发"})

	default:
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}

// 获取Cron任务函数
func getCronTask(name string) func() {
	switch name {
	case "auto_settle":
		return func() { service.AutoSettleTask() }
	case "retry_notify":
		return func() { service.RetryNotifyTask() }
	case "order_query":
		return func() { service.OrderQueryTask() }
	case "risk_check":
		return func() { service.RiskCheckTask() }
	case "cleanup":
		return func() { service.CleanupTask() }
	default:
		return func() {}
	}
}

// AJAX: 获取支付类型列表
func (h *AdminHandler) AjaxPayTypeList(c *gin.Context) {
	var types []model.PayType
	config.DB.Find(&types)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": types})
}

// AJAX: 支付类型操作
func (h *AdminHandler) AjaxPayTypeOp(c *gin.Context) {
	var req struct {
		Action   string `json:"action"` // add, edit, delete
		ID       uint   `json:"id"`
		Name     string `json:"name"`
		Showname string `json:"showname"`
		Status   int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[paytype_op_params_error] err=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add", "edit":
		log.Printf("[paytype_op] action=%s, name=%s, showname=%s", req.Action, req.Name, req.Showname)
		pt := model.PayType{
			Name:     req.Name,
			Showname: req.Showname,
			Status:   req.Status,
		}
		var err error
		if req.Action == "edit" {
			err = config.DB.Model(&model.PayType{}).Where("id = ?", req.ID).Updates(pt).Error
		} else {
			err = config.DB.Create(&pt).Error
		}
		if err != nil {
			log.Printf("[paytype_op_%s_failed] error=%s", req.Action, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功"})
		return

	case "delete":
		result := config.DB.Delete(&model.PayType{}, "id = ?", req.ID)
		if result.Error != nil {
			log.Printf("[paytype_op_delete_failed] id=%d, error=%s", req.ID, result.Error.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return

	case "set_status":
		if err := config.DB.Model(&model.PayType{}).Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
			log.Printf("[paytype_op_set_status_failed] id=%d, status=%d, error=%s", req.ID, req.Status, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "状态更新失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return

	default:
		log.Printf("[paytype_op_unknown_action] action=%s", req.Action)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
	}
}

// AJAX: 获取轮询配置列表
func (h *AdminHandler) AjaxRollList(c *gin.Context) {
	var rolls []model.Roll
	if err := config.DB.Find(&rolls).Error; err != nil {
		log.Printf("[roll_list_failed] error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rolls})
}

// AJAX: 轮询配置操作
func (h *AdminHandler) AjaxRollOp(c *gin.Context) {
	var req struct {
		Action string `json:"action"` // add, edit, delete
		ID     uint   `json:"id"`
		Type   int    `json:"type"`
		Name   string `json:"name"`
		Kind   int    `json:"kind"`
		Info   string `json:"info"`
		Status int    `json:"status"`
		Index  int    `json:"index"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[roll_op_params_error] err=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add", "edit":
		log.Printf("[roll_op] action=%s, name=%s, type=%d", req.Action, req.Name, req.Type)
		roll := model.Roll{
			Type:   req.Type,
			Name:   req.Name,
			Kind:   req.Kind,
			Info:   req.Info,
			Status: req.Status,
			Index:  req.Index,
		}
		var err error
		if req.Action == "edit" {
			err = config.DB.Model(&model.Roll{}).Where("id = ?", req.ID).Updates(roll).Error
		} else {
			err = config.DB.Create(&roll).Error
		}
		if err != nil {
			log.Printf("[roll_op_%s_failed] error=%s", req.Action, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功"})
		return

	case "delete":
		result := config.DB.Delete(&model.Roll{}, "id = ?", req.ID)
		if result.Error != nil {
			log.Printf("[roll_op_delete_failed] id=%d, error=%s", req.ID, result.Error.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return

	case "set_status":
		if err := config.DB.Model(&model.Roll{}).Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
			log.Printf("[roll_op_set_status_failed] id=%d, status=%d, error=%s", req.ID, req.Status, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "状态更新失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return

	default:
		log.Printf("[roll_op_unknown_action] action=%s", req.Action)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
	}
}

// AJAX: 获取分账订单列表
func (h *AdminHandler) AjaxProfitOrderList(c *gin.Context) {
	var orders []model.PsOrder
	if err := config.DB.Order("id DESC").Find(&orders).Error; err != nil {
		log.Printf("[profit_order_list_failed] error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": orders})
}

// AJAX: 获取分账接收方列表
func (h *AdminHandler) AjaxProfitReceiverList(c *gin.Context) {
	var receivers []model.PsReceiver
	if err := config.DB.Order("id DESC").Find(&receivers).Error; err != nil {
		log.Printf("[profit_receiver_list_failed] error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": receivers})
}

// AJAX: 分账接收方操作
func (h *AdminHandler) AjaxProfitReceiverOp(c *gin.Context) {
	var req struct {
		Action  string `json:"action"` // add, edit, delete
		ID      uint   `json:"id"`
		UID     uint   `json:"uid"`
		Name    string `json:"name"`
		Account string `json:"account"`
		Rate    string `json:"rate"`
		Status  int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[profit_receiver_op_params_error] err=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "add", "edit":
		log.Printf("[profit_receiver_op] action=%s, name=%s, account=%s", req.Action, req.Name, req.Account)
		r := model.PsReceiver{
			UID:     req.UID,
			Name:    req.Name,
			Account: req.Account,
			Rate:    req.Rate,
			Status:  req.Status,
		}
		var err error
		if req.Action == "edit" {
			err = config.DB.Model(&model.PsReceiver{}).Where("id = ?", req.ID).Updates(r).Error
		} else {
			err = config.DB.Create(&r).Error
		}
		if err != nil {
			log.Printf("[profit_receiver_op_%s_failed] error=%s", req.Action, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存成功"})
		return

	case "delete":
		result := config.DB.Delete(&model.PsReceiver{}, "id = ?", req.ID)
		if result.Error != nil {
			log.Printf("[profit_receiver_op_delete_failed] id=%d, error=%s", req.ID, result.Error.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return

	case "set_status":
		if err := config.DB.Model(&model.PsReceiver{}).Where("id = ?", req.ID).Update("status", req.Status).Error; err != nil {
			log.Printf("[profit_receiver_op_set_status_failed] id=%d, status=%d, error=%s", req.ID, req.Status, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "状态更新失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态已更新"})
		return

	default:
		log.Printf("[profit_receiver_op_unknown_action] action=%s", req.Action)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
	}
}

// AJAX: 执行分账
func (h *AdminHandler) AjaxProfitDo(c *gin.Context) {
	var req struct {
		PsNo    string `json:"ps_no"`
		TradeNo string `json:"trade_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[profit_do_params_error] err=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	tradeNo := strings.TrimSpace(req.PsNo)
	if tradeNo == "" {
		tradeNo = strings.TrimSpace(req.TradeNo)
	}
	if tradeNo == "" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "缺少订单号"})
		return
	}

	log.Printf("[profit_do] trade_no=%s", tradeNo)
	profitSvc := service.NewProfitService()
	if err := profitSvc.ProcessProfitSharing(tradeNo); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "分账执行完成"})
}

// AJAX: 批量转账记录列表
func (h *AdminHandler) AjaxTransferBatchList(c *gin.Context) {
	var batches []model.Batch
	if err := config.DB.Order("time DESC").Find(&batches).Error; err != nil {
		log.Printf("[transfer_batch_list_failed] error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取列表失败"})
		return
	}

	list := make([]gin.H, 0, len(batches))
	for _, b := range batches {
		var successCount int64
		var failedCount int64
		config.DB.Model(&model.Transfer{}).Where("pay_order_no = ? AND status = 1", b.Batch).Count(&successCount)
		config.DB.Model(&model.Transfer{}).Where("pay_order_no = ? AND status = 2", b.Batch).Count(&failedCount)

		list = append(list, gin.H{
			"batch_no":   b.Batch,
			"filename":   "-",
			"total":      b.Count,
			"success":    successCount,
			"failed":     failedCount,
			"amount":     b.Allmoney,
			"status":     b.Status,
			"created_at": b.Time.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

// AJAX: 创建批量转账
func (h *AdminHandler) AjaxTransferBatchCreate(c *gin.Context) {
	var req struct {
		Filename string `json:"filename"`
		Data     string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[transfer_batch_create_params_error] err=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	log.Printf("[transfer_batch_create] filename=%s, data_len=%d", req.Filename, len(req.Data))
	var rows []struct {
		UID     uint    `json:"uid"`
		Name    string  `json:"name"`
		Account string  `json:"account"`
		Amount  float64 `json:"amount"`
		Remark  string  `json:"remark"`
	}
	if err := json.Unmarshal([]byte(req.Data), &rows); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "批量数据格式错误"})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "批量数据不能为空"})
		return
	}

	batchNo := fmt.Sprintf("B%s%03d", time.Now().Format("20060102150405"), time.Now().UnixNano()%1000)
	now := time.Now()

	tx := config.DB.Begin()
	var totalAmount float64
	validCount := 0

	for i, row := range rows {
		if row.UID == 0 || row.Account == "" || row.Amount <= 0 {
			continue
		}
		transfer := model.Transfer{
			BizNo:      fmt.Sprintf("%s%04d", batchNo, i+1),
			PayOrderNo: batchNo,
			UID:        row.UID,
			Type:       "batch",
			Channel:    0,
			Account:    row.Account,
			Username:   row.Name,
			Money:      row.Amount,
			Costmoney:  0,
			Paytime:    now,
			Status:     0,
			API:        0,
			Desc:       row.Remark,
			Result:     req.Filename,
		}
		if err := tx.Create(&transfer).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "创建转账任务失败"})
			return
		}
		totalAmount += row.Amount
		validCount++
	}

	if validCount == 0 {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "没有有效的转账记录"})
		return
	}

	batch := model.Batch{
		Batch:    batchNo,
		Allmoney: totalAmount,
		Count:    validCount,
		Time:     now,
		Status:   0,
	}
	if err := tx.Create(&batch).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "创建批次失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "创建批次失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "批量转账任务已创建",
		"data": gin.H{
			"batch_no": batchNo,
			"total":    validCount,
			"amount":   totalAmount,
		},
	})
}

// 上传证书文件
func (h *AdminHandler) UploadCert(c *gin.Context) {
	file, err := c.FormFile("cert")
	if err != nil {
		log.Printf("[upload_cert_failed] reason=get file failed, error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "请选择要上传的文件"})
		return
	}

	// 限制文件大小 5MB
	if file.Size > 5*1024*1024 {
		log.Printf("[upload_cert_failed] reason=file too large, size=%d", file.Size)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "文件大小不能超过5MB"})
		return
	}

	// 检查文件类型（只允许 .pem, .cert, .key, .p12, .pfx）
	ext := strings.ToLower(file.Filename)
	if !strings.HasSuffix(ext, ".pem") && !strings.HasSuffix(ext, ".cert") &&
		!strings.HasSuffix(ext, ".key") && !strings.HasSuffix(ext, ".p12") &&
		!strings.HasSuffix(ext, ".pfx") {
		log.Printf("[upload_cert_failed] reason=invalid file type, filename=%s", file.Filename)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "只允许上传 .pem, .cert, .key, .p12, .pfx 格式的证书文件"})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("cert_%d_%s", time.Now().UnixNano(), file.Filename)
	uploadPath := fmt.Sprintf("certs/%s", filename)

	// 确保目录存在
	certDir := "certs"
	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Printf("[upload_cert_failed] reason=create dir failed, error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "上传失败"})
		return
	}

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		log.Printf("[upload_cert_failed] reason=save file failed, error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存文件失败"})
		return
	}

	log.Printf("[upload_cert_success] filename=%s, size=%d", filename, file.Size)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "上传成功",
		"data": map[string]string{
			"path": uploadPath,
			"name": filename,
		},
	})
}

// 格式化JSON
func (h *AdminHandler) FormatJson(c *gin.Context) {
	var req struct {
		Json string `json:"json"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var data interface{}
	if err := json.Unmarshal([]byte(req.Json), &data); err != nil {
		log.Printf("[format_json_failed] error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "JSON格式错误: " + err.Error()})
		return
	}

	formatted, _ := json.MarshalIndent(data, "", "  ")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": string(formatted)})
}

// AJAX: 数据清理统计
func (h *AdminHandler) AjaxCleanStats(c *gin.Context) {
	orderTimeout := adminIntParam(c, "order_timeout", 24)
	maxRetry := adminIntParam(c, "max_retry", 10)
	logDays := adminIntParam(c, "log_days", 30)

	if orderTimeout <= 0 {
		orderTimeout = 24
	}
	if maxRetry <= 0 {
		maxRetry = 10
	}
	if logDays <= 0 {
		logDays = 30
	}

	orderDeadline := time.Now().Add(-time.Duration(orderTimeout) * time.Hour)
	notifyDeadline := time.Now().Add(-time.Duration(maxRetry) * 5 * time.Minute)
	logDeadline := time.Now().AddDate(0, 0, -logDays)

	var orderCount int64
	var failedNotifyCount int64
	var logCount int64
	var cacheBytes int64

	config.DB.Model(&model.Order{}).
		Where("status = ? AND addtime < ?", model.OrderStatusPending, orderDeadline).
		Count(&orderCount)
	config.DB.Model(&model.Order{}).
		Where("status = ? AND notify = ? AND notifytime < ?", model.OrderStatusPaid, 2, notifyDeadline).
		Count(&failedNotifyCount)
	config.DB.Model(&model.Log{}).Where("date < ?", logDeadline).Count(&logCount)
	config.DB.Model(&model.Cache{}).Select("COALESCE(SUM(LENGTH(v)),0)").Scan(&cacheBytes)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"order_count":         orderCount,
			"failed_notify_count": failedNotifyCount,
			"log_count":           logCount,
			"cache_size_bytes":    cacheBytes,
			"cache_size":          formatBytes(cacheBytes),
		},
	})
}

// AJAX: 执行数据清理
func (h *AdminHandler) AjaxCleanRun(c *gin.Context) {
	var req struct {
		Action       string `json:"action" binding:"required"`
		OrderTimeout int    `json:"order_timeout"`
		MaxRetry     int    `json:"max_retry"`
		LogDays      int    `json:"log_days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var affected int64
	switch req.Action {
	case "orders":
		if req.OrderTimeout <= 0 {
			req.OrderTimeout = 24
		}
		deadline := time.Now().Add(-time.Duration(req.OrderTimeout) * time.Hour)
		result := config.DB.Where("status = ? AND addtime < ?", model.OrderStatusPending, deadline).Delete(&model.Order{})
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "清理失败"})
			return
		}
		affected = result.RowsAffected
	case "failed_notifies":
		if req.MaxRetry <= 0 {
			req.MaxRetry = 10
		}
		deadline := time.Now().Add(-time.Duration(req.MaxRetry) * 5 * time.Minute)
		result := config.DB.Model(&model.Order{}).
			Where("status = ? AND notify = ? AND notifytime < ?", model.OrderStatusPaid, 2, deadline).
			Update("notify", 0)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "清理失败"})
			return
		}
		affected = result.RowsAffected
	case "logs":
		if req.LogDays <= 0 {
			req.LogDays = 30
		}
		deadline := time.Now().AddDate(0, 0, -req.LogDays)
		result := config.DB.Where("date < ?", deadline).Delete(&model.Log{})
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "清理失败"})
			return
		}
		affected = result.RowsAffected
	case "cache":
		result := config.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Cache{})
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "清理失败"})
			return
		}
		affected = result.RowsAffected
	default:
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "清理完成", "data": gin.H{"count": affected}})
}

// AJAX: 导出订单数据
func (h *AdminHandler) AjaxExportOrders(c *gin.Context) {
	limit := adminIntParam(c, "limit", 100000)
	if limit <= 0 {
		limit = 100000
	}
	if limit > 100000 {
		limit = 100000
	}

	status := adminStringParam(c, "status")
	uid := adminIntParam(c, "uid", 0)
	payType := adminIntParam(c, "type", 0)
	startDate := adminStringParam(c, "start_date")
	endDate := adminStringParam(c, "end_date")

	query := config.DB.Model(&model.Order{})
	if status != "" && status != "-1" {
		query = query.Where("status = ?", status)
	}
	if uid > 0 {
		query = query.Where("uid = ?", uid)
	}
	if payType > 0 {
		query = query.Where("type = ?", payType)
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("addtime >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("addtime < ?", t.AddDate(0, 0, 1))
		}
	}

	var total int64
	query.Count(&total)
	if total > int64(limit) {
		total = int64(limit)
	}

	var orders []model.Order
	query.Order("addtime DESC").Limit(limit).Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"count": total,
		"data":  orders,
	})
}

func adminStringParam(c *gin.Context, key string) string {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		return v
	}
	return strings.TrimSpace(c.PostForm(key))
}

func adminIntParam(c *gin.Context, key string, defaultValue int) int {
	value := adminStringParam(c, key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func formatBytes(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
}
