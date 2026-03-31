package admin

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"
	"paygo/src/plugin"
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
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
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
		Mod    string `json:"mod"`
		OldPwd string `json:"old_pwd"`
		NewPwd string `json:"new_pwd"`
		ConfirmPwd string `json:"confirm_pwd"`
		// 网站设置
		Sitename string `json:"sitename"`
		Title string `json:"title"`
		Localurl string `json:"localurl"`
		Apiurl string `json:"apiurl"`
		Email string `json:"email"`
		Kfqq string `json:"kfqq"`
		RegOpen string `json:"reg_open"`
		// 支付设置
		TestOpen string `json:"test_open"`
		PaySuccessPage string `json:"pay_success_page"`
		PayErrorPage string `json:"pay_error_page"`
		// 结算设置
		SettleMoney string `json:"settle_money"`
		SettleCycle string `json:"settle_cycle"`
		SettleAlipay string `json:"settle_alipay"`
		SettleWxpay string `json:"settle_wxpay"`
		// 转账设置
		TransferMin string `json:"transfer_min"`
		TransferMax string `json:"transfer_max"`
		TransferFee string `json:"transfer_fee"`
		TransferAlipay string `json:"transfer_alipay"`
		TransferWxpay string `json:"transfer_wxpay"`
		// 快捷登录
		LoginAlipay string `json:"login_alipay"`
		LoginQq string `json:"login_qq"`
		LoginWx string `json:"login_wx"`
		// 通知设置
		NotifyEmail string `json:"notify_email"`
		EmailNotify string `json:"email_notify"`
		OrderNotify string `json:"order_notify"`
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
	case "pay":
		cfgMap["test_open"] = req.TestOpen
		cfgMap["pay_success_page"] = req.PaySuccessPage
		cfgMap["pay_error_page"] = req.PayErrorPage
	case "settle":
		cfgMap["settle_money"] = req.SettleMoney
		cfgMap["settle_cycle"] = req.SettleCycle
		cfgMap["settle_alipay"] = req.SettleAlipay
		cfgMap["settle_wxpay"] = req.SettleWxpay
	case "transfer":
		cfgMap["transfer_min"] = req.TransferMin
		cfgMap["transfer_max"] = req.TransferMax
		cfgMap["transfer_fee"] = req.TransferFee
		cfgMap["transfer_alipay"] = req.TransferAlipay
		cfgMap["transfer_wxpay"] = req.TransferWxpay
	case "oauth":
		cfgMap["login_alipay"] = req.LoginAlipay
		cfgMap["login_qq"] = req.LoginQq
		cfgMap["login_wx"] = req.LoginWx
	case "notice":
		cfgMap["notify_email"] = req.NotifyEmail
		cfgMap["email_notify"] = req.EmailNotify
		cfgMap["order_notify"] = req.OrderNotify
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
		ID              uint    `json:"id"`
		Mode            int     `json:"mode"`
		Type            int     `json:"type"`
		Plugin          string  `json:"plugin"`
		Name            string  `json:"name"`
		Rate            float64 `json:"rate"`
		Status          int     `json:"status"`
		Apptype         string  `json:"apptype"`
		Daytop          int     `json:"daytop"`
		Daystatus       int     `json:"daystatus"`
		Paymin          string  `json:"paymin"`
		Paymax          string  `json:"paymax"`
		Appwxmp         int     `json:"appwxmp"`
		Appwxa          int     `json:"appwxa"`
		Costrate        float64 `json:"costrate"`
		Config          string  `json:"config"`
		PluginShowname   string  `json:"plugin_showname"`
	}

	result := make([]ChannelResponse, len(channels))
	for i, ch := range channels {
		result[i] = ChannelResponse{
			ID:       ch.ID,
			Mode:     ch.Mode,
			Type:     ch.Type,
			Plugin:   ch.Plugin,
			Name:     ch.Name,
			Rate:     ch.Rate,
			Status:   ch.Status,
			Apptype:  ch.Apptype,
			Daytop:   ch.Daytop,
			Daystatus: ch.Daystatus,
			Paymin:   ch.Paymin,
			Paymax:   ch.Paymax,
			Appwxmp:  ch.Appwxmp,
			Appwxa:   ch.Appwxa,
			Costrate: ch.Costrate,
			Config:   ch.Config,
		}
		// 获取插件显示名
		var plugin model.Plugin
		if err := config.DB.First(&plugin, "name = ?", ch.Plugin).Error; err == nil {
			result[i].PluginShowname = plugin.Showname
		} else if bp, ok := builtInMap[ch.Plugin]; ok {
			result[i].PluginShowname = bp.Showname
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
		Action   string  `json:"action"`
		ID       uint    `json:"id"`
		Name     string  `json:"name"`
		Plugin   string  `json:"plugin"`
		Type     int     `json:"type"`
		Mode     int     `json:"mode"`
		Rate     float64 `json:"rate"`
		Costrate float64 `json:"costrate"`
		Daytop   int     `json:"daytop"`
		Paymin   float64 `json:"paymin"`
		Paymax   float64 `json:"paymax"`
		Apptype  string  `json:"apptype"`
		Status   int     `json:"status"`
		Config   string  `json:"config"`
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
			Name:     req.Name,
			Plugin:   req.Plugin,
			Type:     req.Type,
			Mode:     req.Mode,
			Rate:     req.Rate,
			Costrate: req.Costrate,
			Daytop:   req.Daytop,
			Paymin:   strconv.FormatFloat(req.Paymin, 'f', 2, 64),
			Paymax:   strconv.FormatFloat(req.Paymax, 'f', 2, 64),
			Apptype:  req.Apptype,
			Status:   req.Status,
			Config:   req.Config,
		}

		var err error
		if req.Action == "edit" {
			err = config.DB.Model(&model.Channel{}).Where("id = ?", req.ID).Updates(channel).Error
		} else {
			err = config.DB.Create(&channel).Error
		}

		if err != nil {
			log.Printf("[channel_op_%s_failed] error=%s", req.Action, err.Error())
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
		Name       string `json:"name"`
		Showname   string `json:"showname"`
		Author     string `json:"author"`
		Link       string `json:"link"`
		Types      string `json:"types"`
		Transtypes string `json:"transtypes"`
		Status     int    `json:"status"`
		Config     string `json:"config"`
		Note       string `json:"note"` // 内置插件有说明
		IsBuiltIn  bool   `json:"is_builtin"`
	}

	result := make([]PluginResponse, 0, len(builtInPlugins))
	for _, p := range builtInPlugins {
		dbPlugin, exists := dbPluginMap[p.Name]
		status := 1  // 默认启用
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
		log.Printf("[插件刷新] 同步了 %d 个新插件", synced)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "刷新成功"})
		return

	case "set_status":
		result := config.DB.Model(&model.Plugin{}).Where("name = ?", req.Name).Update("status", req.Status)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新失败"})
			return
		}
		log.Printf("[插件状态更新] name=%s, status=%d", req.Name, req.Status)
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
					"types":     strings.Join(info.Types, ","),
					"transtypes": strings.Join(info.Transtypes, ","),
					"inputs":     info.Inputs,
					"config":    "{}",
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
		result := config.DB.Model(&model.Plugin{}).Where("name = ?", req.Name).Update("config", cfg)
		if result.Error != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "保存失败"})
			return
		}
		log.Printf("[插件配置保存] name=%s", req.Name)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "配置已保存"})
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
	switch req.Action {
	case "refund":
		err = h.orderSvc.Refund(req.TradeNo, req.Money)
	case "freeze":
		err = h.orderSvc.Freeze(req.TradeNo)
	case "unfreeze":
		err = h.orderSvc.Unfreeze(req.TradeNo)
	case "notify":
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
	var req struct {
		Action string `json:"action"`
		UID    uint   `json:"uid"`
		Status int    `json:"status"`
		Money  float64 `json:"money"`
		Type   string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	switch req.Action {
	case "reset_key":
		newKey := generateAPIKey()
		config.DB.Model(&model.User{}).Where("uid = ?", req.UID).Update("key", newKey)
		log.Printf("[admin_reset_key] uid=%d, new_key=%s", req.UID, newKey)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密钥已重置", "key": newKey})
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
