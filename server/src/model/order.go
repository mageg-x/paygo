package model

import (
	"time"
)

// 订单状态常量
const (
	OrderStatusPending  = 0 // 待支付
	OrderStatusPaid     = 1 // 已支付
	OrderStatusRefunded = 2 // 已退款
	OrderStatusFrozen   = 3 // 已冻结
	OrderStatusPreauth  = 4 // 预授权
)

// 订单表
type Order struct {
	TradeNo     string    `gorm:"primaryKey;column:trade_no" json:"trade_no"`
	OutTradeNo  string    `gorm:"column:out_trade_no" json:"out_trade_no"`
	ApiTradeNo  string    `gorm:"column:api_trade_no" json:"api_trade_no"`
	UID         uint      `gorm:"column:uid" json:"uid"`
	Tid         int       `gorm:"column:tid" json:"tid"`
	Type        int       `gorm:"column:type" json:"type"`
	Channel     int       `gorm:"column:channel" json:"channel"`
	Name        string    `gorm:"column:name" json:"name"`
	Money       float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Realmoney   float64   `gorm:"column:realmoney;type:decimal(10,2)" json:"realmoney"`
	Getmoney    float64   `gorm:"column:getmoney;type:decimal(10,2)" json:"getmoney"`
	Profitmoney float64   `gorm:"column:profitmoney;type:decimal(10,2)" json:"profitmoney"`
	Refundmoney float64   `gorm:"column:refundmoney;type:decimal(10,2)" json:"refundmoney"`
	NotifyURL   string    `gorm:"column:notify_url" json:"notify_url"`
	ReturnURL   string    `gorm:"column:return_url" json:"return_url"`
	Param       string    `gorm:"column:param" json:"param"`
	Addtime     time.Time `gorm:"column:addtime;index:idx_order_status_addtime,priority:2" json:"addtime"`
	Endtime     time.Time `gorm:"column:endtime" json:"endtime"`
	Date        string    `gorm:"column:date" json:"date"`
	Domain      string    `gorm:"column:domain" json:"domain"`
	Domain2     string    `gorm:"column:domain2" json:"domain2"`
	IP          string    `gorm:"column:ip" json:"ip"`
	Buyer       string    `gorm:"column:buyer" json:"buyer"`
	Status      int       `gorm:"column:status;index:idx_order_status_addtime,priority:1" json:"status"`
	Notify      int       `gorm:"column:notify" json:"notify"`
	Notifytime  time.Time `gorm:"column:notifytime" json:"notifytime"`
	Invite      uint      `gorm:"column:invite" json:"invite"`
	Invitemoney float64   `gorm:"column:invitemoney;type:decimal(10,2)" json:"invitemoney"`
	Combine     int       `gorm:"column:combine" json:"combine"`
	Profits     int       `gorm:"column:profits" json:"profits"`
	Profits2    int       `gorm:"column:profits2" json:"profits2"`
	Settle      int       `gorm:"column:settle" json:"settle"`
	Subchannel  int       `gorm:"column:subchannel" json:"subchannel"`
	Payurl      string    `gorm:"column:payurl" json:"payurl"`
	Ext         string    `gorm:"column:ext" json:"ext"`
	Version     int       `gorm:"column:version" json:"version"`
}

func (Order) TableName() string {
	return "order"
}

// 退款订单表
type RefundOrder struct {
	RefundNo    string    `gorm:"primaryKey;column:refund_no" json:"refund_no"`
	OutRefundNo string    `gorm:"column:out_refund_no" json:"out_refund_no"`
	TradeNo     string    `gorm:"column:trade_no" json:"trade_no"`
	UID         uint      `gorm:"column:uid" json:"uid"`
	Money       float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Reducemoney float64   `gorm:"column:reducemoney;type:decimal(10,2)" json:"reducemoney"`
	Status      int       `gorm:"column:status" json:"status"`
	Addtime     time.Time `gorm:"column:addtime" json:"addtime"`
	Endtime     time.Time `gorm:"column:endtime" json:"endtime"`
}

func (RefundOrder) TableName() string {
	return "refundorder"
}
