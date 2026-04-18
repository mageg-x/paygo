package router

import (
	"strings"

	"paygo/src/handler/admin"
	"paygo/src/handler/api"
	"paygo/src/handler/user"
	"paygo/src/install"
	"paygo/src/middleware"
	"paygo/src/static"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	// 中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recover())
	r.Use(middleware.CORS())

	fs := static.GetFileSystem()

	// ========== 安装向导 ==========
	installHandler := install.NewInstallHandler()
	r.GET("/install/status", installHandler.CheckInstallStatus)
	r.POST("/install/do", installHandler.DoInstall)

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
	r.GET("/uploads/*path", func(c *gin.Context) {
		path := c.Param("path")
		path = strings.TrimPrefix(path, "/")
		c.File("uploads/" + path)
	})

	// ========== JSON API 全部在 /api/* ==========
	apiHandler := api.NewPayHandler()
	adminHandler := admin.NewAdminHandler()
	groupHandler := admin.NewGroupHandler()
	userHandler := user.NewUserHandler()

	api := r.Group("/api")
	{
		// 支付接口（公开）
		api.POST("/pay/submit", apiHandler.Submit)
		api.POST("/pay/cashier_submit", apiHandler.CashierSubmit)
		api.POST("/pay/create", apiHandler.Create)
		api.GET("/pay/query", apiHandler.Query)
		api.POST("/pay/query", apiHandler.Query)
		api.POST("/pay/refund", apiHandler.Refund)
		api.POST("/pay/notify/:trade_no", apiHandler.Notify)
		api.GET("/pay/return/:trade_no", apiHandler.Return)
		api.GET("/pay/types", apiHandler.GetTypes)
		api.GET("/pay/channels", apiHandler.GetChannels)

		// 管理后台 API - 公开
		api.POST("/admin/login", adminHandler.Login)

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
			adminAuth.POST("/set/upload/wxkf", adminHandler.UploadWxkfQrcode)
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
			// 支付类型管理
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
			// 上传证书
			adminAuth.POST("/upload/cert", adminHandler.UploadCert)
			// 格式化JSON
			adminAuth.POST("/format/json", adminHandler.FormatJson)
			// 数据清理
			adminAuth.GET("/clean/stats", adminHandler.AjaxCleanStats)
			adminAuth.POST("/clean/run", adminHandler.AjaxCleanRun)
			// 数据导出
			adminAuth.GET("/export/orders", adminHandler.AjaxExportOrders)
		}

		// 商户后台 API - 公开
		api.POST("/user/login", userHandler.Login)
		api.POST("/user/reg", userHandler.Register)
		api.POST("/user/reg/send", userHandler.RegSendCode)
		api.POST("/user/findpwd/send", userHandler.FindPwdSendCode)
		api.POST("/user/findpwd/reset", userHandler.FindPwdReset)

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
