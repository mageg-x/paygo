package user

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"paygo/src/config"
	"paygo/src/middleware"
	"paygo/src/model"
	"paygo/src/service"
)

// 商户Handler
type UserHandler struct {
	authSvc     *service.AuthService
	orderSvc    *service.OrderService
	settleSvc   *service.SettleService
	transferSvc *service.TransferService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		authSvc:     service.NewAuthService(),
		orderSvc:    service.NewOrderService(),
		settleSvc:   service.NewSettleService(),
		transferSvc: service.NewTransferService(),
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
		log.Printf("[商户登录失败] IP: %s, UID: %d, 错误: %s", ip, req.UID, err.Error())
		// 业务逻辑错误返回 200 + code=1，而不是 401
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	log.Printf("[商户登录成功] IP: %s, UID: %d", ip, user.UID)
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
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	ip := middleware.GetRealIP(c)

	// 验证码验证（TODO: 实际实现）
	// code := c.PostForm("code")
	// if !h.authSvc.VerifyCode("reg", email, code) {
	//     c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "验证码错误"})
	//     return
	// }

	user, err := h.authSvc.UserRegister(req.Email, req.Phone, req.Password, req.InviteCode, ip)
	if err != nil {
		log.Printf("[商户注册失败] IP: %s, 错误: %s", ip, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	log.Printf("[商户注册成功] IP: %s, UID: %d", ip, user.UID)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "注册成功",
		"uid":  user.UID,
	})
}

// 登出
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie("user_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "已退出"})
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
	config.DB.Where("uid = ?", uid).Order("id DESC").Limit(10).Find(&recentOrders)

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
	query.Offset((page-1)*pageSize).Limit(pageSize).Order("id DESC").Find(&orders)

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
	config.DB.Where("uid = ?", uid).Offset((page-1)*pageSize).Limit(pageSize).
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
	config.DB.Where("uid = ?", uid).Offset((page-1)*pageSize).Limit(pageSize).
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
	account := c.PostForm("account")
	username := c.PostForm("username")
	moneyStr := c.PostForm("money")
	settleType, _ := strconv.Atoi(c.PostForm("type"))

	money, _ := strconv.ParseFloat(moneyStr, 10)

	settle, err := h.settleSvc.ApplySettle(uid, account, username, money, settleType)
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

	username := c.PostForm("username")
	phone := c.PostForm("phone")
	qq := c.PostForm("qq")

	data := map[string]interface{}{}
	if username != "" {
		data["username"] = username
	}
	if phone != "" {
		data["phone"] = phone
	}
	if qq != "" {
		data["qq"] = qq
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
	certname := c.PostForm("certname")
	certno := c.PostForm("certno")
	certtype, _ := strconv.Atoi(c.PostForm("certtype"))

	data := map[string]interface{}{
		"cert":      1, // 待审核
		"certname":  certname,
		"certno":    certno,
		"certtype":  certtype,
		"certtime":  time.Now(),
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
	page, _ := strconv.Atoi(c.PostForm("page"))
	pageSize, _ := strconv.Atoi(c.PostForm("limit"))
	status := c.PostForm("status")

	orders, total, _ := h.orderSvc.GetUserOrders(uid, -1, page, pageSize)

	if status != "" && status != "-1" {
		s, _ := strconv.Atoi(status)
		orders, total, _ = h.orderSvc.GetUserOrders(uid, s, page, pageSize)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  orders,
	})
}

// API: 结算列表
func (h *UserHandler) AjaxSettleList(c *gin.Context) {
	uid := c.GetUint("uid")
	page, _ := strconv.Atoi(c.PostForm("page"))
	pageSize, _ := strconv.Atoi(c.PostForm("limit"))

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
	page, _ := strconv.Atoi(c.PostForm("page"))
	pageSize, _ := strconv.Atoi(c.PostForm("limit"))
	action, _ := strconv.Atoi(c.PostForm("action"))

	records, total, _ := h.transferSvc.GetUserRecords(uid, action, page, pageSize)

	c.JSON(http.StatusOK, gin.H{
		"code":  0,
		"msg":   "",
		"count": total,
		"data":  records,
	})
}
