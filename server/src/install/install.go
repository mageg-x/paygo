package install

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"paygo/src/config"
	"paygo/src/model"

	"github.com/gin-gonic/gin"
)

// InstallHandler 安装向导处理
type InstallHandler struct{}

// NewInstallHandler 创建安装处理器
func NewInstallHandler() *InstallHandler {
	return &InstallHandler{}
}

// CheckInstallStatus 检查安装状态
func (h *InstallHandler) CheckInstallStatus(c *gin.Context) {
	// 检查是否已安装（通过检查config表是否有admin_user配置）
	var cfg model.Config
	result := config.DB.Where("k = ?", "admin_user").First(&cfg)

	if result.RowsAffected > 0 && cfg.V != "" {
		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"msg":    "已安装",
			"status": 1,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"msg":    "未安装",
		"status": 0,
	})
}

// DoInstall 执行安装
func (h *InstallHandler) DoInstall(c *gin.Context) {
	var req struct {
		DbPath    string `json:"db_path"`
		AdminUser string `json:"admin_user"`
		AdminPwd  string `json:"admin_pwd"`
		SysKey    string `json:"sys_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	// 如果提供了数据库路径，初始化数据库
	if req.DbPath != "" {
		if err := initDatabase(req.DbPath); err != nil {
			log.Printf("[install] init database failed: %v", err)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "数据库初始化失败: " + err.Error()})
			return
		}
	}

	// 保存管理员配置
	if req.AdminUser != "" {
		config.Set("admin_user", req.AdminUser)
	}
	if req.AdminPwd != "" {
		config.Set("admin_pwd", req.AdminPwd)
	}
	if req.SysKey != "" {
		config.Set("sys_key", req.SysKey)
	}

	// 创建默认用户组
	createDefaultGroup()

	// 创建默认支付类型
	createDefaultPayTypes()

	log.Printf("[install] install completed, admin_user=%s", req.AdminUser)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "安装成功"})
}

// initDatabase 初始化数据库
func initDatabase(dbPath string) error {
	// 确保目录存在
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 重新初始化数据库连接
	config.LoadConfig(dbPath, config.AppConfig.Port)
	config.InitDB()

	return nil
}

// createDefaultGroup 创建默认用户组
func createDefaultGroup() {
	var count int64
	config.DB.Model(&model.Group{}).Count(&count)
	if count == 0 {
		groups := []model.Group{
			{Name: "普通商户", SettleRate: "0.5", Sort: 1},
			{Name: "VIP商户", SettleRate: "0.3", Sort: 2},
			{Name: "钻石商户", SettleRate: "0.2", Sort: 3},
		}
		for _, g := range groups {
			config.DB.Create(&g)
		}
		log.Println("[install] default groups created")
	}
}

// createDefaultPayTypes 创建默认支付类型
func createDefaultPayTypes() {
	var count int64
	config.DB.Model(&model.PayType{}).Count(&count)
	if count == 0 {
		types := []model.PayType{
			{Name: "alipay", Showname: "支付宝", Status: 1},
			{Name: "wechatpay", Showname: "微信支付", Status: 1},
		}
		for _, t := range types {
			config.DB.Create(&t)
		}
		log.Println("[install] default pay types created")
	}
}
