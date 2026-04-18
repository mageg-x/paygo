package user

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/middleware"
	"paygo/src/model"
	"paygo/src/service"

	"github.com/gin-gonic/gin"
)

// 商户Handler
type UserHandler struct {
	authSvc     *service.AuthService
	orderSvc    *service.OrderService
	settleSvc   *service.SettleService
	transferSvc *service.TransferService
	paymentSvc  *service.PaymentService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		authSvc:     service.NewAuthService(),
		orderSvc:    service.NewOrderService(),
		settleSvc:   service.NewSettleService(),
		transferSvc: service.NewTransferService(),
		paymentSvc:  service.NewPaymentService(),
	}
}

// 登录页面
func (h *UserHandler) LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "user/login.html", nil)
}

// 登录
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Type     string `json:"type"`
		UID      int    `json:"uid"`
		Key      string `json:"key"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var err error
	var token string
	var user *model.User
	ip := c.ClientIP()

	if req.Type == "key" {
		user, token, err = h.authSvc.UserKeyLogin(uint(req.UID), req.Key)
	} else {
		user, token, err = h.authSvc.UserLogin(uint(req.UID), req.Password)
	}

	if err != nil {
		log.Printf("[merchant_login_failed] ip=%s, uid=%d, error=%s", ip, req.UID, err.Error())
		// 业务逻辑错误返回 200 + code=1，而不是 401
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	log.Printf("[merchant_login_success] ip=%s, uid=%d", ip, user.UID)
	c.SetCookie("user_token", token, 86400*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "登录成功",
		"uid":   user.UID,
		"token": token,
	})
}

// 注册页面
func (h *UserHandler) RegPage(c *gin.Context) {
	c.HTML(http.StatusOK, "user/reg.html", nil)
}

// 注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Email      string `json:"email"`
		Phone      string `json:"phone"`
		Password   string `json:"password"`
		InviteCode string `json:"invite_code"`
		Code       string `json:"code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	ip := middleware.GetRealIP(c)

	// 注册验证码验证（按配置要求）
	verifyType := h.authSvc.GetConfig("user_verification") // 0=无, 1=邮箱, 2=手机
	switch verifyType {
	case "1":
		if req.Email == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "请填写邮箱"})
			return
		}
		if req.Code == "" || !h.authSvc.VerifyCode("reg", req.Email, req.Code) {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "验证码错误或已过期"})
			return
		}
	case "2":
		if req.Phone == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "请填写手机号"})
			return
		}
		if req.Code == "" || !h.authSvc.VerifyCode("reg", req.Phone, req.Code) {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "验证码错误或已过期"})
			return
		}
	default:
		// 配置未强制验证时，若传了验证码则做校验
		if req.Code != "" {
			target := req.Email
			if target == "" {
				target = req.Phone
			}
			if target != "" && !h.authSvc.VerifyCode("reg", target, req.Code) {
				c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "验证码错误或已过期"})
				return
			}
		}
	}

	user, err := h.authSvc.UserRegister(req.Email, req.Phone, req.Password, req.InviteCode, ip)
	if err != nil {
		log.Printf("[merchant_register_failed] ip=%s, error=%s", ip, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	log.Printf("[merchant_register_success] ip=%s, uid=%d", ip, user.UID)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "注册成功",
		"uid":  user.UID,
	})
}

// 注册 - 发送验证码
func (h *UserHandler) RegSendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	verifyType := h.authSvc.GetConfig("user_verification") // 0=无, 1=邮箱, 2=手机
	target := ""

	switch verifyType {
	case "1":
		target = strings.TrimSpace(req.Email)
		if target == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "请输入邮箱"})
			return
		}
	case "2":
		target = strings.TrimSpace(req.Phone)
		if target == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "请输入手机号"})
			return
		}
	default:
		if req.Email != "" {
			target = strings.TrimSpace(req.Email)
		} else {
			target = strings.TrimSpace(req.Phone)
		}
		if target == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "请填写邮箱或手机号"})
			return
		}
	}

	if _, err := h.authSvc.GenCode("reg", target); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "验证码已发送"})
}

// 登出
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie("user_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已退出"})
}

// 获取商户信息
func (h *UserHandler) Info(c *gin.Context) {
	uid := c.GetUint("uid")
	user, err := h.authSvc.GetUser(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"uid":        user.UID,
			"gid":        user.GID,
			"username":   user.Username,
			"email":      user.Email,
			"phone":      user.Phone,
			"qq":         user.Qq,
			"alipay_uid": user.AlipayUID,
			"wx_uid":     user.WxUID,
			"qq_uid":     user.QqUID,
			"money":      user.Money,
			"status":     user.Status,
			"cert":       user.Cert,
			"endtime":    user.Endtime,
		},
	})
}

// API: 商户资料中的 API 信息
func (h *UserHandler) AjaxProfileAPI(c *gin.Context) {
	uid := c.GetUint("uid")
	user, err := h.authSvc.GetUser(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"uid": user.UID,
			"key": user.Key,
		},
	})
}

// 商户中心首页
func (h *UserHandler) Index(c *gin.Context) {
	uid := c.GetUint("uid")

	user, _ := h.authSvc.GetUser(uid)

	// 统计数据
	now := time.Now()
	today := now.Format("2006-01-02")

	var todayMoney, totalMoney float64
	var todayCount, totalCount int64
	var settleMoney float64

	config.DB.Model(&model.Order{}).Where("uid = ? AND date = ? AND status = 1", uid, today).
		Select("COALESCE(SUM(money), 0)").Scan(&todayMoney)
	config.DB.Model(&model.Order{}).Where("uid = ? AND date = ? AND status = 1", uid, today).
		Count(&todayCount)

	config.DB.Model(&model.Order{}).Where("uid = ? AND status = 1", uid).
		Select("COALESCE(SUM(money), 0)").Scan(&totalMoney)
	config.DB.Model(&model.Order{}).Where("uid = ? AND status = 1", uid).
		Count(&totalCount)

	config.DB.Model(&model.Settle{}).Where("uid = ? AND status = 1", uid).
		Select("COALESCE(SUM(money), 0)").Scan(&settleMoney)

	// 最新订单
	var recentOrders []model.Order
	config.DB.Where("uid = ?", uid).Order("addtime DESC, trade_no DESC").Limit(10).Find(&recentOrders)

	// 公告
	var announces []model.Anounce
	config.DB.Where("status = 1").Order("sort DESC").Limit(5).Find(&announces)

	c.HTML(http.StatusOK, "user/index.html", gin.H{
		"user":          user,
		"today_money":   todayMoney,
		"today_count":   todayCount,
		"total_money":   totalMoney,
		"total_count":   totalCount,
		"settle_money":  settleMoney,
		"recent_orders": recentOrders,
		"announces":     announces,
	})
}

// 订单列表
func (h *UserHandler) OrderList(c *gin.Context) {
	uid := c.GetUint("uid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	status := c.DefaultQuery("status", "-1")
	tradeNo := c.Query("trade_no")

	var orders []model.Order
	var total int64

	query := config.DB.Model(&model.Order{}).Where("uid = ?", uid)
	if status != "-1" {
		query = query.Where("status = ?", status)
	}
	if tradeNo != "" {
		query = query.Where("trade_no LIKE ?", "%"+tradeNo+"%")
	}

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Order("addtime DESC, trade_no DESC").Find(&orders)

	c.HTML(http.StatusOK, "user/order.html", gin.H{
		"orders": orders,
		"total":  total,
		"page":   page,
	})
}

// 结算记录
func (h *UserHandler) SettleList(c *gin.Context) {
	uid := c.GetUint("uid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var settles []model.Settle
	var total int64

	config.DB.Model(&model.Settle{}).Where("uid = ?", uid).Count(&total)
	config.DB.Where("uid = ?", uid).Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").Find(&settles)

	c.HTML(http.StatusOK, "user/settle.html", gin.H{
		"settles": settles,
		"total":   total,
		"page":    page,
	})
}

// 资金记录
func (h *UserHandler) RecordList(c *gin.Context) {
	uid := c.GetUint("uid")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var records []model.Record
	var total int64

	config.DB.Model(&model.Record{}).Where("uid = ?", uid).Count(&total)
	config.DB.Where("uid = ?", uid).Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").Find(&records)

	c.HTML(http.StatusOK, "user/record.html", gin.H{
		"records": records,
		"total":   total,
		"page":    page,
	})
}

// 申请结算
func (h *UserHandler) ApplySettle(c *gin.Context) {
	uid := c.GetUint("uid")
	var req struct {
		Account  string  `json:"account"`
		Username string  `json:"username"`
		Money    float64 `json:"money"`
		Type     int     `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Account = c.PostForm("account")
		req.Username = c.PostForm("username")
		req.Money, _ = strconv.ParseFloat(c.PostForm("money"), 10)
		req.Type, _ = strconv.Atoi(c.PostForm("type"))
	}

	settle, err := h.settleSvc.ApplySettle(uid, req.Account, req.Username, req.Money, req.Type)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "申请成功", "settle": settle})
}

// 资料编辑页面
func (h *UserHandler) EditInfo(c *gin.Context) {
	uid := c.GetUint("uid")
	user, _ := h.authSvc.GetUser(uid)

	c.HTML(http.StatusOK, "user/editinfo.html", gin.H{
		"user": user,
	})
}

// 更新资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	uid := c.GetUint("uid")

	var req struct {
		Username  string `json:"username"`
		Phone     string `json:"phone"`
		Qq        string `json:"qq"`
		AlipayUID string `json:"alipay_uid"`
		WxUID     string `json:"wx_uid"`
		QqUID     string `json:"qq_uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Username = c.PostForm("username")
		req.Phone = c.PostForm("phone")
		req.Qq = c.PostForm("qq")
		req.AlipayUID = c.PostForm("alipay_uid")
		req.WxUID = c.PostForm("wx_uid")
		req.QqUID = c.PostForm("qq_uid")
	}

	data := map[string]interface{}{}
	if req.Username != "" {
		data["username"] = req.Username
	}
	if req.Phone != "" {
		data["phone"] = req.Phone
	}
	if req.Qq != "" {
		data["qq"] = req.Qq
	}
	if req.AlipayUID != "" {
		data["alipay_uid"] = req.AlipayUID
	}
	if req.WxUID != "" {
		data["wx_uid"] = req.WxUID
	}
	if req.QqUID != "" {
		data["qq_uid"] = req.QqUID
	}

	err := h.authSvc.UpdateUser(uid, data)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
}

// 实名认证页面
func (h *UserHandler) CertificatePage(c *gin.Context) {
	uid := c.GetUint("uid")
	user, _ := h.authSvc.GetUser(uid)

	c.HTML(http.StatusOK, "user/certificate.html", gin.H{
		"user": user,
	})
}

// 提交实名认证
func (h *UserHandler) SubmitCertificate(c *gin.Context) {
	uid := c.GetUint("uid")
	var req struct {
		Certname string `json:"certname"`
		Certno   string `json:"certno"`
		Certtype int    `json:"certtype"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Certname = c.PostForm("certname")
		req.Certno = c.PostForm("certno")
		req.Certtype, _ = strconv.Atoi(c.PostForm("certtype"))
	}

	data := map[string]interface{}{
		"cert":     1, // 待审核
		"certname": req.Certname,
		"certno":   req.Certno,
		"certtype": req.Certtype,
		"certtime": time.Now(),
	}

	err := h.authSvc.UpdateUser(uid, data)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "提交失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "提交成功"})
}

// API: 订单列表
func (h *UserHandler) AjaxOrderList(c *gin.Context) {
	uid := c.GetUint("uid")
	page := userIntParam(c, "page", 1)
	pageSize := userIntParam(c, "limit", 20)
	status := userStringParam(c, "status")
	tradeNo := userStringParam(c, "trade_no")
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	orders, total, _ := h.orderSvc.GetUserOrders(uid, -1, page, pageSize, tradeNo)

	if status != "" && status != "-1" {
		s, _ := strconv.Atoi(status)
		orders, total, _ = h.orderSvc.GetUserOrders(uid, s, page, pageSize, tradeNo)
	}

	respData := enrichOrdersWithTypeName(orders)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  respData,
	})
}

// API: 商户首页统计
func (h *UserHandler) AjaxStats(c *gin.Context) {
	uid := c.GetUint("uid")
	now := time.Now()
	loc := now.Location()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	var todayMoney, totalMoney float64
	var todayCount, totalCount int64
	paidLikeStatuses := []int{model.OrderStatusPaid, model.OrderStatusRefunded}

	config.DB.Model(&model.Order{}).
		Where("uid = ? AND addtime >= ? AND addtime < ?", uid, todayStart, tomorrowStart).
		Count(&todayCount)
	config.DB.Model(&model.Order{}).
		Where("uid = ?", uid).
		Count(&totalCount)

	config.DB.Model(&model.Order{}).
		Where("uid = ? AND addtime >= ? AND addtime < ? AND status IN ?", uid, todayStart, tomorrowStart, paidLikeStatuses).
		Select("COALESCE(SUM(money), 0)").Scan(&todayMoney)
	config.DB.Model(&model.Order{}).
		Where("uid = ? AND status IN ?", uid, paidLikeStatuses).
		Select("COALESCE(SUM(money), 0)").Scan(&totalMoney)

	var recentOrders []model.Order
	config.DB.Where("uid = ?", uid).
		Order("addtime DESC, trade_no DESC").
		Limit(10).
		Find(&recentOrders)
	recentOrdersWithType := enrichOrdersWithTypeName(recentOrders)

	var announces []model.Anounce
	config.DB.Where("status = 1").
		Order("sort DESC").
		Limit(5).
		Find(&announces)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"today_money":   todayMoney,
			"today_count":   todayCount,
			"total_money":   totalMoney,
			"total_count":   totalCount,
			"recent_orders": recentOrdersWithType,
			"announces":     announces,
		},
	})
}

// API: 结算列表
func (h *UserHandler) AjaxSettleList(c *gin.Context) {
	uid := c.GetUint("uid")
	page := userIntParam(c, "page", 1)
	pageSize := userIntParam(c, "limit", 20)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	settles, total, _ := h.settleSvc.GetUserSettles(uid, page, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  settles,
	})
}

// API: 资金记录
func (h *UserHandler) AjaxRecordList(c *gin.Context) {
	uid := c.GetUint("uid")
	page := userIntParam(c, "page", 1)
	pageSize := userIntParam(c, "limit", 20)
	action := userIntParam(c, "action", 0)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	records, total, _ := h.transferSvc.GetUserRecords(uid, action, page, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  records,
	})
}

// API: 邀请记录
func (h *UserHandler) AjaxInviteRecords(c *gin.Context) {
	uid := c.GetUint("uid")
	page := userIntParam(c, "page", 1)
	pageSize := userIntParam(c, "limit", 20)
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	query := config.DB.Model(&model.Record{}).
		Where("uid = ? AND action = ?", uid, 7)
	query.Count(&total)

	var records []model.Record
	query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  records,
	})
}

// API: 创建余额充值订单
func (h *UserHandler) AjaxRechargeCreate(c *gin.Context) {
	uid := c.GetUint("uid")
	var req struct {
		Type      int     `json:"type"`
		Money     float64 `json:"money"`
		NotifyURL string  `json:"notify_url"`
		ReturnURL string  `json:"return_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}
	if req.Type <= 0 || req.Money <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	result, err := h.paymentSvc.SubmitRechargePayment(service.SubmitParams{
		UID:        uid,
		OutTradeNo: fmt.Sprintf("RECHARGE_%d_%d", uid, time.Now().UnixNano()),
		Type:       req.Type,
		Name:       "余额充值",
		Money:      req.Money,
		NotifyURL:  req.NotifyURL,
		ReturnURL:  req.ReturnURL,
		Param:      fmt.Sprintf("recharge|%d", uid),
		IP:         c.ClientIP(),
		Device:     "pc",
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     0,
		"trade_no": result["trade_no"],
		"result":   result["result"],
	})
}

// API: 订单操作（商户）
func (h *UserHandler) AjaxOrderOp(c *gin.Context) {
	uid := c.GetUint("uid")

	var req struct {
		Action  string  `json:"action" binding:"required"`
		TradeNo string  `json:"trade_no" binding:"required"`
		Money   float64 `json:"money"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	order, err := h.orderSvc.GetOrder(req.TradeNo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if order.UID != uid {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "订单不属于当前商户"})
		return
	}

	switch req.Action {
	case "notify":
		if err := h.orderSvc.RetryNotify(req.TradeNo); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已触发重新通知"})
		return
	case "refund":
		if req.Money <= 0 {
			available := order.Money - order.Refundmoney
			if available <= 0 {
				c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "可退款金额为0"})
				return
			}
			req.Money = available
		}
		if err := h.orderSvc.Refund(req.TradeNo, req.Money); err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "退款成功"})
		return
	default:
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
		return
	}
}

// 找回密码 - 发送验证码
func (h *UserHandler) FindPwdSendCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	// 检查邮箱是否存在
	var user model.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "该邮箱未注册"})
		return
	}

	// 发送验证码（邮件）
	if _, err := h.authSvc.GenCode("findpwd", req.Email); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "验证码已发送"})
}

// 找回密码 - 重置密码
func (h *UserHandler) FindPwdReset(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	// 验证验证码
	if !h.authSvc.VerifyCode("findpwd", req.Email, req.Code) {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "验证码错误或已过期"})
		return
	}

	// 获取用户
	var user model.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}

	// 更新密码（bcrypt）
	pwdHash, err := h.authSvc.HashUserPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密码处理失败"})
		return
	}
	if err := config.DB.Model(&user).Update("pwd", pwdHash).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "密码更新失败"})
		return
	}

	log.Printf("[reset_password_success] email=%s", req.Email)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "密码重置成功"})
}

// AJAX: 获取用户组列表(商户端)
func (h *UserHandler) AjaxGroupList(c *gin.Context) {
	var groups []model.Group
	if err := config.DB.Order("sort ASC").Find(&groups).Error; err != nil {
		log.Printf("[group_list_failed] uid=%d, error=%s", c.GetInt("uid"), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取用户组列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": groups})
}

// AJAX: 购买用户组（余额支付）
func (h *UserHandler) AjaxGroupBuy(c *gin.Context) {
	uid := c.GetUint("uid")
	var req struct {
		GroupID uint `json:"group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	var user model.User
	if err := config.DB.Where("uid = ?", uid).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}

	var group model.Group
	if err := config.DB.Where("gid = ? AND isbuy = 1", req.GroupID).First(&group).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户组不存在或不可购买"})
		return
	}

	if group.Price <= 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "该用户组未设置售价"})
		return
	}
	if user.Money < group.Price {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "余额不足"})
		return
	}

	tx := config.DB.Begin()
	oldMoney := user.Money
	newMoney := user.Money - group.Price
	updates := map[string]interface{}{
		"gid":   group.GID,
		"money": newMoney,
	}
	if group.Expire > 0 {
		base := time.Now()
		if user.Endtime.After(base) && user.GID == group.GID {
			base = user.Endtime
		}
		updates["endtime"] = base.AddDate(0, group.Expire, 0)
	}

	if err := tx.Model(&model.User{}).Where("uid = ?", uid).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "购买失败"})
		return
	}

	tradeNo := fmt.Sprintf("GB%d%d", uid, time.Now().Unix())
	record := &model.Record{
		UID:      uid,
		Action:   9,
		Money:    -group.Price,
		Oldmoney: oldMoney,
		Newmoney: newMoney,
		Type:     "group_buy",
		TradeNo:  tradeNo,
		Date:     time.Now(),
	}
	if err := tx.Create(record).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "购买失败"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "购买失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "购买成功",
		"data": gin.H{
			"gid":     group.GID,
			"name":    group.Name,
			"balance": newMoney,
		},
	})
}

// AJAX: 用户组转让记录
func (h *UserHandler) AjaxGroupTransferList(c *gin.Context) {
	uid := c.GetInt("uid")
	var transfers []model.UserGroupTransfer
	if err := config.DB.Where("from_uid = ? OR to_uid = ?", uid, uid).Order("id DESC").Find(&transfers).Error; err != nil {
		log.Printf("[group_transfer_list_failed] uid=%d, error=%s", uid, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取转让记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": transfers})
}

// AJAX: 创建用户组转让
func (h *UserHandler) AjaxGroupTransferCreate(c *gin.Context) {
	uid := c.GetInt("uid")
	var req struct {
		TargetUID int     `json:"target_uid" binding:"required"`
		GroupID   uint    `json:"group_id" binding:"required"`
		Price     float64 `json:"price"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[group_transfer_params_error] uid=%d, error=%s", uid, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	log.Printf("[group_transfer] from_uid=%d, to_uid=%d, group_id=%d, price=%.2f", uid, req.TargetUID, req.GroupID, req.Price)

	// 创建转让记录
	transfer := model.UserGroupTransfer{
		FromUID: uint(uid),
		ToUID:   uint(req.TargetUID),
		GroupID: req.GroupID,
		Price:   req.Price,
		Status:  0,
	}
	if err := config.DB.Create(&transfer).Error; err != nil {
		log.Printf("[group_transfer_create_failed] from_uid=%d, to_uid=%d, error=%s", uid, req.TargetUID, err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "创建转让记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "转让请求已提交"})
}

func userStringParam(c *gin.Context, key string) string {
	if v := strings.TrimSpace(c.Query(key)); v != "" {
		return v
	}
	return strings.TrimSpace(c.PostForm(key))
}

func userIntParam(c *gin.Context, key string, defaultValue int) int {
	value := userStringParam(c, key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

type userOrderWithTypeName struct {
	model.Order
	Typename string `json:"typename"`
}

func enrichOrdersWithTypeName(orders []model.Order) []userOrderWithTypeName {
	if len(orders) == 0 {
		return make([]userOrderWithTypeName, 0)
	}

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

	respData := make([]userOrderWithTypeName, 0, len(orders))
	for _, o := range orders {
		typeName := typeNameMap[o.Type]
		if typeName == "" {
			typeName = fallbackTypeNameByPlugin(channelPluginMap[o.Channel])
		}
		if typeName == "" {
			typeName = fmt.Sprintf("类型%d", o.Type)
		}
		respData = append(respData, userOrderWithTypeName{
			Order:    o,
			Typename: typeName,
		})
	}

	return respData
}
