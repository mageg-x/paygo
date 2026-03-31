package plugin

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// 插件接口
type Plugin interface {
	// 获取插件信息
	GetInfo() PluginInfo

	// 支付提交
	Submit(params map[string]interface{}) (SubmitResult, error)

	// 移动端提交
	Mapi(params map[string]interface{}) (SubmitResult, error)

	// 异步回调
	Notify(tradeNo string, c *gin.Context) (NotifyResult, error)

	// 同步回调
	Return(tradeNo string, c *gin.Context) (ReturnResult, error)

	// 支付成功页面
	OK(tradeNo string) (string, error)

	// 退款
	Refund(params map[string]interface{}) (RefundResult, error)

	// 转账
	Transfer(params map[string]interface{}) (TransferResult, error)

	// 转账查询
	TransferQuery(params map[string]interface{}) (TransferQueryResult, error)
}

// 插件信息
type PluginInfo struct {
	Name       string                 // 插件名称
	Showname   string                 // 显示名称
	Author     string                 // 作者
	Link       string                 // 链接
	Types      []string               // 支持的支付类型
	Transtypes []string               // 支持的转账类型
	Inputs     map[string]InputConfig // 配置项
	Select     map[string]string      // 支付方式选择
	Note       string                 // 说明
}

// 输入配置
type InputConfig struct {
	Name    string            // 配置名称
	Type    string            // input, textarea, select
	Options map[string]string // select选项
	Note    string            // 说明
}

// 提交结果
type SubmitResult struct {
	Type string      // jump, qrcode, html, error, page, scheme, jsapi, app, scan
	URL  string      // 跳转URL或二维码内容
	Page string      // 页面模板
	Data interface{} // 扩展数据
	Msg  string      // 错误信息
}

// 回调结果
type NotifyResult struct {
	Success    bool    // 是否成功
	TradeNo    string  // 平台订单号
	APITradeNo string  // API订单号
	Amount     float64 // 金额
	Buyer      string  // 买家
	Message    string  // 消息
}

// 同步回调结果
type ReturnResult struct {
	Success bool
	TradeNo string
	Message string
	URL     string // 跳转URL
}

// 退款结果
type RefundResult struct {
	Code    int     // 0成功, -1失败
	TradeNo string  // 订单号
	Fee     float64 // 退款金额
	Time    string  // 退款时间
	Buyer   string  // 买家
	ErrCode string  // 错误码
	ErrMsg  string  // 错误信息
}

// 转账结果
type TransferResult struct {
	Code    int    // 0成功, -1失败
	OrderID string // 转账订单号
	PayDate string // 支付时间
	ErrCode string // 错误码
	ErrMsg  string // 错误信息
}

// 转账查询结果
type TransferQueryResult struct {
	Code    int     // 0成功, -1失败
	Status  int     // 状态: 0=处理中, 1=成功, 2=失败
	Amount  float64 // 金额
	PayDate string  // 支付时间
	ErrMsg  string  // 错误信息
}

// 插件注册表
var pluginRegistry = make(map[string]func() Plugin)

// 注册插件
func Register(name string, factory func() Plugin) {
	pluginRegistry[name] = factory
}

// 获取插件处理器
func GetHandler(name string) Plugin {
	factory, ok := pluginRegistry[name]
	if !ok {
		return nil
	}
	return factory()
}

// 获取所有插件
func GetAllPlugins() []string {
	names := make([]string, 0, len(pluginRegistry))
	for name := range pluginRegistry {
		names = append(names, name)
	}
	return names
}

// 基础插件实现
type BasePlugin struct {
	info PluginInfo
}

func (p *BasePlugin) GetInfo() PluginInfo {
	return p.info
}

// 辅助函数
func Strval(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case int:
		return string(rune(v.(int)))
	case int64:
		return strconv.FormatInt(v.(int64), 10)
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', 2, 64)
	default:
		return ""
	}
}

func Atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func Atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
