package admin

import (
	"log"
	"net/http"
	"strconv"

	"paygo/src/config"
	"paygo/src/model"

	"github.com/gin-gonic/gin"
)

// 用户组Handler
type GroupHandler struct{}

func NewGroupHandler() *GroupHandler {
	return &GroupHandler{}
}

// AJAX: 用户组列表
func (h *GroupHandler) AjaxGroupList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var groups []model.Group
	var total int64

	config.DB.Model(&model.Group{}).Count(&total)
	if err := config.DB.Offset(offset).Limit(pageSize).Order("sort ASC, gid ASC").Find(&groups).Error; err != nil {
		log.Printf("[用户组列表查询失败] error=%s", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "获取列表失败"})
		return
	}

	// 获取默认分组ID
	defaultGroupID := uint(1)
	if dg := config.Get("default_group"); dg != "" {
		if parsed, err := strconv.ParseUint(dg, 10, 32); err == nil {
			defaultGroupID = uint(parsed)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":        0,
		"msg":         "",
		"count":       total,
		"data":        groups,
		"default_gid": defaultGroupID,
	})
}

// AJAX: 用户组操作
func (h *GroupHandler) AjaxGroupOp(c *gin.Context) {
	var req struct {
		Action     string  `json:"action"`
		Gid        uint    `json:"gid"`
		Name       string  `json:"name"`
		Info       string  `json:"info"`
		Sort       int     `json:"sort"`
		Isbuy      int     `json:"isbuy"`
		Price      float64 `json:"price"`
		Expire     int     `json:"expire"`
		SettleOpen int     `json:"settle_open"`
		SettleType int     `json:"settle_type"`
		SettleRate any     `json:"settle_rate"` // 兼容数字/字符串
		Settings   string  `json:"settings"`
		Config     string  `json:"config"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[用户组操作参数错误] error=%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	if req.Action == "" {
		log.Printf("[用户组操作参数错误] action为空")
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误：action不能为空"})
		return
	}

	settleRate, err := parseSettleRate(req.SettleRate)
	if err != nil {
		log.Printf("[用户组操作参数错误] settle_rate=%v, error=%s", req.SettleRate, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误：结算费率格式错误"})
		return
	}

	switch req.Action {
	case "add":
		group := model.Group{
			Name:       req.Name,
			Info:       req.Info,
			Sort:       req.Sort,
			Isbuy:      req.Isbuy,
			Price:      req.Price,
			Expire:     req.Expire,
			SettleOpen: req.SettleOpen,
			SettleType: req.SettleType,
			SettleRate: strconv.FormatFloat(settleRate, 'f', 2, 64),
			Settings:   req.Settings,
			Config:     req.Config,
		}
		if err := config.DB.Create(&group).Error; err != nil {
			log.Printf("[用户组添加失败] name=%s, error=%s", req.Name, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "添加失败"})
			return
		}
		log.Printf("[用户组添加成功] gid=%d, name=%s", group.GID, req.Name)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "添加成功"})
		return

	case "edit":
		updates := map[string]interface{}{
			"name":        req.Name,
			"info":        req.Info,
			"sort":        req.Sort,
			"isbuy":       req.Isbuy,
			"price":       req.Price,
			"expire":      req.Expire,
			"settle_open": req.SettleOpen,
			"settle_type": req.SettleType,
			"settle_rate": strconv.FormatFloat(settleRate, 'f', 2, 64),
			"settings":    req.Settings,
			"config":      req.Config,
		}
		if err := config.DB.Model(&model.Group{}).Where("gid = ?", req.Gid).Updates(updates).Error; err != nil {
			log.Printf("[用户组更新失败] gid=%d, error=%s", req.Gid, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "更新失败"})
			return
		}
		log.Printf("[用户组更新成功] gid=%d, name=%s", req.Gid, req.Name)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功"})
		return

	case "delete":
		if req.Gid == 1 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "不能删除默认用户组"})
			return
		}
		// 检查是否有商户使用该用户组
		var count int64
		config.DB.Model(&model.User{}).Where("gid = ?", req.Gid).Count(&count)
		if count > 0 {
			log.Printf("[用户组删除失败] gid=%d, reason=该用户组下有商户, count=%d", req.Gid, count)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "该用户组下有商户，无法删除"})
			return
		}
		if err := config.DB.Delete(&model.Group{}, "gid = ?", req.Gid).Error; err != nil {
			log.Printf("[用户组删除失败] gid=%d, error=%s", req.Gid, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "删除失败"})
			return
		}
		log.Printf("[用户组删除成功] gid=%d", req.Gid)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功"})
		return

	case "set_default":
		// 检查用户组是否存在
		var group model.Group
		if err := config.DB.First(&group, req.Gid).Error; err != nil {
			log.Printf("[设置默认用户组失败] gid=%d, error=用户组不存在", req.Gid)
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户组不存在"})
			return
		}
		// 更新配置
		if err := config.Set("default_group", strconv.FormatUint(uint64(req.Gid), 10)); err != nil {
			log.Printf("[设置默认用户组失败] gid=%d, error=%s", req.Gid, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "设置失败"})
			return
		}
		log.Printf("[设置默认用户组成功] gid=%d, name=%s", req.Gid, group.Name)
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "设置成功"})
		return

	case "get":
		var group model.Group
		if err := config.DB.First(&group, req.Gid).Error; err != nil {
			log.Printf("[获取用户组失败] gid=%d, error=%s", req.Gid, err.Error())
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "用户组不存在"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": group})
		return

	default:
		log.Printf("[用户组操作未知动作] action=%s", req.Action)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "未知操作"})
	}
}

func parseSettleRate(v any) (float64, error) {
	switch t := v.(type) {
	case nil:
		return 0, nil
	case float64:
		return t, nil
	case float32:
		return float64(t), nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case string:
		if t == "" {
			return 0, nil
		}
		return strconv.ParseFloat(t, 64)
	default:
		return 0, strconv.ErrSyntax
	}
}
