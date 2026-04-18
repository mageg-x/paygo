package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"paygo/src/config"
	"paygo/src/model"

	"github.com/gin-gonic/gin"
)

type ipRateBucket struct {
	windowStart int64
	count       int
}

var (
	ipRateMu      sync.Mutex
	ipRateBuckets = make(map[string]*ipRateBucket)
	ipRateLastGC  int64
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
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		allowOrigin := ""
		if origin != "" && isAllowedCORSOrigin(origin, c.Request.Host) {
			allowOrigin = origin
		}

		if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")

		if c.Request.Method == "OPTIONS" {
			if origin != "" && allowOrigin == "" {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func normalizeOrigin(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return ""
	}
	scheme := strings.ToLower(strings.TrimSpace(u.Scheme))
	if scheme != "http" && scheme != "https" {
		return ""
	}
	host := strings.TrimSpace(u.Host)
	if host == "" {
		return ""
	}
	return scheme + "://" + strings.ToLower(host)
}

func isAllowedCORSOrigin(origin string, reqHost string) bool {
	normalizedOrigin := normalizeOrigin(origin)
	if normalizedOrigin == "" {
		return false
	}
	originURL, err := url.Parse(normalizedOrigin)
	if err != nil || originURL == nil {
		return false
	}

	originHost := strings.ToLower(strings.TrimSpace(originURL.Host))
	requestHost := strings.ToLower(strings.TrimSpace(reqHost))
	if requestHost != "" && originHost == requestHost {
		return true
	}

	originHostname := strings.ToLower(strings.TrimSpace(originURL.Hostname()))
	reqHostname := normalizeHost(reqHost)
	if isLoopbackHost(originHostname) && isLoopbackHost(reqHostname) {
		return true
	}

	// 允许配置白名单，逗号分隔，示例：
	// cors_allow_origins=https://a.example.com,http://localhost:3000
	allowList := strings.TrimSpace(config.Get("cors_allow_origins"))
	if allowList == "" {
		return false
	}
	for _, item := range strings.Split(allowList, ",") {
		candidate := normalizeOrigin(item)
		if candidate != "" && strings.EqualFold(candidate, normalizedOrigin) {
			return true
		}
	}
	return false
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

// IPRateLimit 基于客户端IP+路由的固定窗口限流。
func IPRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	windowSec := int64(window.Seconds())
	if windowSec <= 0 {
		windowSec = 60
	}

	return func(c *gin.Context) {
		route := strings.TrimSpace(c.FullPath())
		if route == "" {
			route = strings.TrimSpace(c.Request.URL.Path)
		}
		key := c.ClientIP() + "|" + c.Request.Method + "|" + route
		now := time.Now().Unix()

		ipRateMu.Lock()
		if shouldRateLimit(key, now, windowSec, limit) {
			ipRateMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 1, "msg": "请求过于频繁，请稍后重试"})
			c.Abort()
			return
		}
		maybeCleanupRateBuckets(now, windowSec)
		ipRateMu.Unlock()

		c.Next()
	}
}

func shouldRateLimit(key string, now int64, windowSec int64, limit int) bool {
	b, ok := ipRateBuckets[key]
	if !ok {
		ipRateBuckets[key] = &ipRateBucket{windowStart: now, count: 1}
		return false
	}
	if now-b.windowStart >= windowSec {
		b.windowStart = now
		b.count = 1
		return false
	}
	b.count++
	return b.count > limit
}

func maybeCleanupRateBuckets(now int64, windowSec int64) {
	if now-ipRateLastGC < 120 {
		return
	}
	ipRateLastGC = now
	expireBefore := now - windowSec*2
	for k, b := range ipRateBuckets {
		if b == nil || b.windowStart < expireBefore {
			delete(ipRateBuckets, k)
		}
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
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return "未知"
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "未知"
	}

	if parsed.IsLoopback() {
		return "本地回环"
	}

	if parsed.IsLinkLocalUnicast() || parsed.IsLinkLocalMulticast() {
		return "链路本地"
	}

	if isPrivateIP(parsed) {
		return "内网地址"
	}

	return "公网IP"
}

func isPrivateIP(ip net.IP) bool {
	// IPv4 私网网段
	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10", // CGNAT
	}

	for _, cidr := range privateCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	// IPv6 本地地址
	if strings.HasPrefix(ip.String(), "fc") || strings.HasPrefix(ip.String(), "fd") {
		return true
	}

	return false
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

		if !IsValidAdminToken(token) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "登录已过期"})
			c.Abort()
			return
		}

		cfg := config.AppConfig
		c.Set("admin_user", cfg.AdminUser)
		c.Next()
	}
}

const adminTokenTTL = 30 * 24 * time.Hour

// GenerateAdminToken 生成管理员Token，格式: username.ts.hmac(hex)
func GenerateAdminToken(username, password, sysKey string) string {
	ts := time.Now().Unix()
	payload := username + "." + strconv.FormatInt(ts, 10)
	mac := hmac.New(sha256.New, []byte(sysKey+"|"+password))
	mac.Write([]byte(payload))
	return payload + "." + hex.EncodeToString(mac.Sum(nil))
}

func verifyAdminToken(token, username, password, sysKey string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}
	if parts[0] != username {
		return false
	}

	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || ts <= 0 {
		return false
	}
	now := time.Now().Unix()
	if now-ts > int64(adminTokenTTL.Seconds()) || ts-now > 300 {
		return false
	}

	payload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(sysKey+"|"+password))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(strings.TrimSpace(parts[2]))), []byte(expected))
}

// IsValidAdminToken 校验管理员token是否有效
func IsValidAdminToken(token string) bool {
	cfg := config.AppConfig
	return verifyAdminToken(token, cfg.AdminUser, cfg.AdminPwd, cfg.SysKey)
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

		// token格式：uid.ts.hmac(hex)
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}

		uid, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}
		ts, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}
		now := time.Now().Unix()
		// token 有效期30天
		if ts <= 0 || now-ts > 30*24*3600 || ts-now > 300 {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}

		var user model.User
		result := config.DB.First(&user, uint(uid))
		if result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}

		if user.Status != 1 {
			c.JSON(http.StatusForbidden, gin.H{"code": -1, "msg": "账号已被禁用"})
			c.Abort()
			return
		}
		payload := parts[0] + "." + parts[1]
		mac := hmac.New(sha256.New, []byte(config.AppConfig.SysKey+"|"+user.Key))
		mac.Write([]byte(payload))
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(strings.ToLower(strings.TrimSpace(parts[2]))), []byte(expected)) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": -1, "msg": "登录已过期"})
			c.Abort()
			return
		}

		c.Set("uid", user.UID)
		c.Set("user", &user)
		c.Next()
	}
}

func isSameHostURL(raw, host string) (bool, *url.URL) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, nil
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return false, nil
	}
	if u.Host == "" {
		return false, u
	}
	reqHost := normalizeHost(host)
	srcHost := normalizeHost(u.Host)
	if reqHost == "" || srcHost == "" {
		return false, u
	}
	if strings.EqualFold(reqHost, srcHost) {
		return true, u
	}
	// 兼容本地开发常见场景：localhost 与 127.0.0.1 / ::1
	if isLoopbackHost(reqHost) && isLoopbackHost(srcHost) {
		return true, u
	}
	return false, u
}

func normalizeHost(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse("//" + raw)
	if err != nil || u == nil {
		return strings.ToLower(strings.Trim(raw, "[]"))
	}
	h := strings.TrimSpace(u.Hostname())
	if h == "" {
		return strings.ToLower(strings.Trim(raw, "[]"))
	}
	return strings.ToLower(h)
}

func isLoopbackHost(h string) bool {
	h = normalizeHost(h)
	if h == "" {
		return false
	}
	if h == "localhost" {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

// ConsoleOnly 仅允许后台控制台页面发起的请求访问。
// 用于限制敏感接口（如 API 密钥读取）不被商户业务系统直接调用。
func ConsoleOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := strings.TrimSpace(c.Request.Host)
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		referer := strings.TrimSpace(c.GetHeader("Referer"))
		secFetchSite := strings.ToLower(strings.TrimSpace(c.GetHeader("Sec-Fetch-Site")))

		if secFetchSite == "cross-site" {
			log.Printf("[console_only_denied] path=%s, reason=cross-site request, origin=%s, referer=%s", c.Request.URL.Path, origin, referer)
			c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅允许后台页面访问"})
			c.Abort()
			return
		}
		// 浏览器同源请求在部分环境下可能不带 Origin/Referer，允许通过。
		if origin == "" && referer == "" && secFetchSite == "same-origin" {
			c.Next()
			return
		}

		okOrigin := false
		if origin != "" {
			if same, _ := isSameHostURL(origin, host); same {
				okOrigin = true
			}
		}

		okReferer := false
		if referer != "" {
			if same, u := isSameHostURL(referer, host); same && u != nil {
				p := u.EscapedPath()
				if strings.HasPrefix(p, "/user") || strings.HasPrefix(p, "/admin") {
					okReferer = true
				}
			}
		}

		if !okOrigin && !okReferer {
			log.Printf("[console_only_denied] path=%s, reason=origin/referer not console, origin=%s, referer=%s", c.Request.URL.Path, origin, referer)
			c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "仅允许后台页面访问"})
			c.Abort()
			return
		}

		c.Next()
	}
}
