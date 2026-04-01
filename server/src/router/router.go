package router

import (
	"strings"

	"paygo/src/handler/admin"
	"paygo/src/handler/api"
	"paygo/src/handler/user"
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

	// ========== JSON API 全部在 /api/* ==========
	apiHandler := api.NewPayHandler()
	adminHandler := admin.NewAdminHandler()
	userHandler := user.NewUserHandler()

	api := r.Group("/api")
	{
		// 支付接口（公开）
		api.POST("/pay/submit", apiHandler.Submit)
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
			adminAuth.POST("/orders", adminHandler.AjaxOrderList)
			adminAuth.POST("/settle/op", adminHandler.AjaxSettleOp)
			adminAuth.GET("/settles", adminHandler.AjaxSettleList)
			adminAuth.POST("/set/save", adminHandler.SaveSettings)
			adminAuth.GET("/set/config", adminHandler.AjaxGetConfig)
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
		}

		// 商户后台 API - 公开
		api.POST("/user/login", userHandler.Login)
		api.POST("/user/reg", userHandler.Register)

		// 商户后台 API - 需要认证
		userAuth := api.Group("/user")
		userAuth.Use(middleware.UserAuth())
		{
			userAuth.GET("/info", userHandler.Info)
			userAuth.POST("/logout", userHandler.Logout)
			userAuth.POST("/order/list", userHandler.AjaxOrderList)
			userAuth.POST("/settle/apply", userHandler.ApplySettle)
			userAuth.POST("/settle/list", userHandler.AjaxSettleList)
			userAuth.POST("/record/list", userHandler.AjaxRecordList)
			userAuth.POST("/editinfo", userHandler.UpdateProfile)
			userAuth.POST("/certificate", userHandler.SubmitCertificate)
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
