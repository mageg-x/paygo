package middleware

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"paygo/src/config"
	"paygo/src/model"

	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		ip := c.ClientIP()
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[%s] %s %s %d %v", ip, method, path, status, latency)
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Recover 异常恢复中间件
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": -1,
					"msg":  "服务器内部错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// GetRealIP 获取真实IP
func GetRealIP(c *gin.Context) string {
	// 优先从 X-Forwarded-For 获取
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// X-Real-IP
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 默认ClientIP
	return c.ClientIP()
}

// GetClientIPCity 获取IP归属地
func GetClientIPCity(ip string) string {
	// TODO: 实现IP归属地查询
	// 可以使用 ip2region 库
	return "未知"
}

// 管理员认证中间件
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Admin-Token")
		if token == "" {
			cookie, err := c.Cookie("admin_token")
			if err != nil || cookie == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未登录"})
				c.Abort()
				return
			}
			token = cookie
		}

		// 验证token：重新生成token比较
		cfg := config.AppConfig
		expectedToken := generateAdminToken(cfg.AdminUser, cfg.AdminPwd, cfg.SysKey)

		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "登录已过期"})
			c.Abort()
			return
		}

		c.Set("admin_user", cfg.AdminUser)
		c.Next()
	}
}

// 生成管理员token
func generateAdminToken(username, password, sysKey string) string {
	hash := md5.Sum([]byte(username + password + password + sysKey))
	return fmt.Sprintf("%x", hash)
}

// 商户认证中间件
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("User-Token")
		if token == "" {
			cookie, err := c.Cookie("user_token")
			if err != nil || cookie == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "未登录"})
				c.Abort()
				return
			}
			token = cookie
		}

		// TODO: 解密token获取用户信息
		// token格式: uid\t session \t expiretime (加密)

		// 临时实现：使用uid作为简单验证
		var user model.User
		result := config.DB.First(&user, token)
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}

		if user.Status != 0 {
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "msg": "账号已被禁用"})
			c.Abort()
			return
		}

		c.Set("uid", user.UID)
		c.Set("user", &user)
		c.Next()
	}
}
