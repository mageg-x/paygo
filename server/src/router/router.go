package router

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopay/src/handler/admin"
	"gopay/src/handler/api"
	"gopay/src/handler/user"
	"gopay/src/middleware"
	"gopay/src/static"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())
	r.Use(middleware.CORS())

	fs := static.GetFileSystem()

	// ========== 静态文件 ==========
	r.GET("/static/*path", func(c *gin.Context) {
		path := c.Param("path")
		path = strings.TrimPrefix(path, "/")
		static.ServeFile(c.Writer, c.Request, fs, "static/"+path)
	})
	r.GET("/assets/*path", func(c *gin.Context) {
		path := c.Param("path")
		path = strings.TrimPrefix(path, "/")
		static.ServeFile(c.Writer, c.Request, fs, "assets/"+path)
	})
	r.GET("/snapshot/*path", func(c *gin.Context) {
		path := c.Param("path")
		path = strings.TrimPrefix(path, "/")
		static.ServeFile(c.Writer, c.Request, fs, "snapshot/"+path)
	})
	r.GET("/i-want.html", func(c *gin.Context) {
		static.ServeFile(c.Writer, c.Request, fs, "i-want.html")
	})
	r.GET("/uploads/*path", func(c *gin.Context) {
		rawPath := strings.TrimPrefix(c.Param("path"), "/")
		cleanRel := strings.TrimPrefix(filepath.Clean(rawPath), string(filepath.Separator))
		if cleanRel == "." || cleanRel == "" || strings.HasPrefix(cleanRel, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "无效文件路径"})
			return
		}
		fullPath := filepath.Join("uploads", cleanRel)
		base := filepath.Clean("uploads") + string(filepath.Separator)
		fullClean := filepath.Clean(fullPath)
		if fullClean != filepath.Clean("uploads") && !strings.HasPrefix(fullClean, base) {
			c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "无权访问该文件"})
			return
		}
		c.Header("X-Content-Type-Options", "nosniff")
		c.File(fullClean)
	})

	// ========== JSON API 全部在 /api/* ==========
	apiHandler := api.NewPayHandler()
	adminHandler := admin.NewAdminHandler()
	groupHandler := admin.NewGroupHandler()
	userHandler := user.NewUserHandler()

	api := r.Group("/api")
	{
		// 支付接口（公开）
		api.POST("/pay/submit", middleware.IPRateLimit(40, time.Minute), apiHandler.Submit)
		api.POST("/pay/cashier_submit", middleware.IPRateLimit(50, time.Minute), apiHandler.CashierSubmit)
		api.POST("/pay/create", middleware.IPRateLimit(40, time.Minute), apiHandler.Create)
		api.GET("/pay/query", middleware.IPRateLimit(120, time.Minute), apiHandler.Query)
		api.POST("/pay/query", middleware.IPRateLimit(120, time.Minute), apiHandler.Query)
		api.POST("/pay/refund", middleware.IPRateLimit(20, time.Minute), apiHandler.Refund)
		api.POST("/pay/test_notify_session", middleware.IPRateLimit(120, time.Minute), apiHandler.CreateTestNotifySession)
		api.GET("/pay/test_notify_session/:token", middleware.IPRateLimit(120, time.Minute), apiHandler.GetTestNotifySession)
		api.POST("/pay/test_notify/:token", apiHandler.TestNotifyCallback)
		api.POST("/pay/notify/:trade_no", apiHandler.Notify)
		api.GET("/pay/return/:trade_no", apiHandler.Return)
		api.GET("/pay/types", apiHandler.GetTypes)
		api.GET("/pay/channels", apiHandler.GetChannels)
		api.GET("/download/gopay/:target", apiHandler.Download)

		// 管理后台 API - 公开
		api.POST("/admin/login", middleware.IPRateLimit(10, time.Minute), adminHandler.Login)

		// 管理后台 API - 需要认证
		adminAuth := api.Group("/admin")
		adminAuth.Use(middleware.AdminAuth())
		{
			adminAuth.POST("/logout", adminHandler.Logout)
			adminAuth.GET("/users", adminHandler.AjaxUserList)
			adminAuth.POST("/user/op", adminHandler.AjaxUserOp)
			adminAuth.POST("/user/add", adminHandler.AddUser)
			adminAuth.POST("/user/update", adminHandler.UpdateUser)
			adminAuth.GET("/user/edit", adminHandler.EditUserPage)
			adminAuth.POST("/order/op", adminHandler.AjaxOrderOp)
			adminAuth.GET("/orders", adminHandler.AjaxOrderList)
			adminAuth.POST("/orders", adminHandler.AjaxOrderList)
			adminAuth.POST("/settle/op", adminHandler.AjaxSettleOp)
			adminAuth.GET("/settles", adminHandler.AjaxSettleList)
			adminAuth.POST("/set/save", adminHandler.SaveSettings)
			adminAuth.GET("/set/config", adminHandler.AjaxGetConfig)
			adminAuth.GET("/set/get", adminHandler.AjaxGetSettings)
			adminAuth.POST("/set/upload/wxkf", middleware.ConsoleOnly(), adminHandler.UploadWxkfQrcode)
			adminAuth.GET("/stats", adminHandler.AjaxStats)
			// 转账管理
			adminAuth.GET("/transfer", adminHandler.AjaxTransferList)
			adminAuth.POST("/transfer/op", adminHandler.AjaxTransferOp)
			// 通道管理
			adminAuth.GET("/channel", adminHandler.AjaxChannelList)
			adminAuth.POST("/channel/op", adminHandler.AjaxChannelOp)
			// 插件管理
			adminAuth.GET("/plugin", adminHandler.AjaxPluginList)
			adminAuth.POST("/plugin/op", adminHandler.AjaxPluginOp)
			// 邀请码管理
			adminAuth.GET("/invitecode", adminHandler.AjaxInviteCodeList)
			adminAuth.POST("/invitecode/generate", adminHandler.AjaxInviteCodeGenerate)
			adminAuth.POST("/invitecode/delete", adminHandler.AjaxInviteCodeDelete)
			// 用户组管理
			adminAuth.GET("/group", groupHandler.AjaxGroupList)
			adminAuth.POST("/group/op", groupHandler.AjaxGroupOp)
			// 风控管理
			adminAuth.GET("/risk", adminHandler.AjaxRiskList)
			adminAuth.POST("/risk/op", adminHandler.AjaxRiskOp)
			// 黑名单管理
			adminAuth.GET("/blacklist", adminHandler.AjaxBlacklistList)
			adminAuth.POST("/blacklist/op", adminHandler.AjaxBlacklistOp)
			// 域名授权管理
			adminAuth.GET("/domain", adminHandler.AjaxDomainList)
			adminAuth.POST("/domain/op", adminHandler.AjaxDomainOp)
			// 公告管理
			adminAuth.GET("/anounce", adminHandler.AjaxAnounceList)
			adminAuth.POST("/anounce/op", adminHandler.AjaxAnounceOp)
			adminAuth.GET("/announce", adminHandler.AjaxAnounceList)
			adminAuth.POST("/announce/op", adminHandler.AjaxAnounceOp)
			// 操作日志
			adminAuth.GET("/log", adminHandler.AjaxLogList)
			// SSO单点登录
			adminAuth.POST("/sso", adminHandler.AjaxSSOLogin)
			adminAuth.GET("/sso/recent", adminHandler.AjaxSSORecent)
			adminAuth.POST("/sso/recent/op", adminHandler.AjaxSSORecentOp)
			// 计划任务
			adminAuth.GET("/cron", adminHandler.AjaxCronList)
			adminAuth.POST("/cron/op", adminHandler.AjaxCronOp)
			// 支付方式
			adminAuth.GET("/paytype", adminHandler.AjaxPayTypeList)
			adminAuth.POST("/paytype/op", adminHandler.AjaxPayTypeOp)
			// 轮询配置管理
			adminAuth.GET("/roll", adminHandler.AjaxRollList)
			adminAuth.POST("/roll/op", adminHandler.AjaxRollOp)
			// 分账管理
			adminAuth.GET("/profit/order", adminHandler.AjaxProfitOrderList)
			adminAuth.GET("/profit/receiver", adminHandler.AjaxProfitReceiverList)
			adminAuth.POST("/profit/receiver/op", adminHandler.AjaxProfitReceiverOp)
			adminAuth.POST("/profit/do", adminHandler.AjaxProfitDo)
			// 批量转账
			adminAuth.GET("/transfer/batch", adminHandler.AjaxTransferBatchList)
			adminAuth.POST("/transfer/batch/create", adminHandler.AjaxTransferBatchCreate)
			// 格式化JSON
			adminAuth.POST("/format/json", adminHandler.FormatJson)
			// 数据清理
			adminAuth.GET("/clean/stats", adminHandler.AjaxCleanStats)
			adminAuth.POST("/clean/run", adminHandler.AjaxCleanRun)
			// 数据导出
			adminAuth.GET("/export/orders", adminHandler.AjaxExportOrders)
		}

		// 商户后台 API - 公开
		api.POST("/user/login", middleware.IPRateLimit(12, time.Minute), userHandler.Login)
		api.POST("/user/reg", middleware.IPRateLimit(8, time.Minute), userHandler.Register)
		api.POST("/user/reg/send", middleware.IPRateLimit(6, time.Minute), userHandler.RegSendCode)
		api.POST("/user/findpwd/send", middleware.IPRateLimit(6, time.Minute), userHandler.FindPwdSendCode)
		api.POST("/user/findpwd/reset", middleware.IPRateLimit(10, time.Minute), userHandler.FindPwdReset)

		// 商户后台 API - 需要认证
		userAuth := api.Group("/user")
		userAuth.Use(middleware.UserAuth())
		{
			userAuth.GET("/info", userHandler.Info)
			userAuth.GET("/profile/api", middleware.ConsoleOnly(), userHandler.AjaxProfileAPI)
			userAuth.GET("/stats", userHandler.AjaxStats)
			userAuth.POST("/logout", userHandler.Logout)
			userAuth.GET("/orders", userHandler.AjaxOrderList)
			userAuth.POST("/order/list", userHandler.AjaxOrderList)
			userAuth.POST("/order/op", userHandler.AjaxOrderOp)
			userAuth.POST("/settle/apply", userHandler.ApplySettle)
			userAuth.GET("/settles", userHandler.AjaxSettleList)
			userAuth.POST("/settle/list", userHandler.AjaxSettleList)
			userAuth.GET("/records", userHandler.AjaxRecordList)
			userAuth.POST("/record/list", userHandler.AjaxRecordList)
			userAuth.GET("/invite/records", userHandler.AjaxInviteRecords)
			userAuth.POST("/recharge/create", userHandler.AjaxRechargeCreate)
			userAuth.POST("/editinfo", userHandler.UpdateProfile)
			userAuth.POST("/certificate", userHandler.SubmitCertificate)
			userAuth.GET("/group/list", userHandler.AjaxGroupList)
			userAuth.POST("/group/buy", userHandler.AjaxGroupBuy)
			userAuth.GET("/group/transfer/list", userHandler.AjaxGroupTransferList)
			userAuth.POST("/group/transfer/create", userHandler.AjaxGroupTransferCreate)
		}
	}

	// ========== SPA Fallback ==========
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 路径返回 404
		if len(path) >= 5 && path[:5] == "/api" {
			c.JSON(404, gin.H{"code": 404, "msg": "API不存在"})
			return
		}
		// 其他所有路径返回 SPA
		static.ServeFile(c.Writer, c.Request, fs, "index.html")
	})

	return r
}
