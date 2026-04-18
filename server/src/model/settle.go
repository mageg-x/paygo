package model

import (
	"time"
)

// 结算状态常量
const (
	SettleStatusPending   = 0 // 待处理
	SettleStatusCompleted = 1 // 已完成
	SettleStatusProcessing = 2 // 正在结算
	SettleStatusFailed    = 3 // 结算失败
)

// 结算记录表
type Settle struct {
	ID              uint      `gorm:"primaryKey;column:id" json:"id"`
	UID             uint      `gorm:"column:uid" json:"uid"`
	Batch           string    `gorm:"column:batch" json:"batch"`
	Auto            int       `gorm:"column:auto" json:"auto"`
	Type            int       `gorm:"column:type" json:"type"`
	Account         string    `gorm:"column:account" json:"account"`
	Username        string    `gorm:"column:username" json:"username"`
	Money           float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Realmoney       float64   `gorm:"column:realmoney;type:decimal(10,2)" json:"realmoney"`
	Addtime         time.Time `gorm:"column:addtime" json:"addtime"`
	Endtime         time.Time `gorm:"column:endtime" json:"endtime"`
	Status          int       `gorm:"column:status" json:"status"`
	TransferStatus  int       `gorm:"column:transfer_status" json:"transfer_status"`
	TransferNo      string    `gorm:"column:transfer_no" json:"transfer_no"`
	TransferResult  string    `gorm:"column:transfer_result" json:"transfer_result"`
	TransferDate    time.Time `gorm:"column:transfer_date" json:"transfer_date"`
	Result          string    `gorm:"column:result" json:"result"`
}

func (Settle) TableName() string {
	return "settle"
}

// 批量结算批次表
type Batch struct {
	Batch   string    `gorm:"primaryKey;column:batch" json:"batch"`
	Allmoney float64  `gorm:"column:allmoney;type:decimal(10,2)" json:"allmoney"`
	Count   int       `gorm:"column:count" json:"count"`
	Time    time.Time `gorm:"column:time" json:"time"`
	Status  int       `gorm:"column:status" json:"status"`
}

func (Batch) TableName() string {
	return "batch"
}

// 转账记录表
type Transfer struct {
	BizNo        string    `gorm:"primaryKey;column:biz_no" json:"biz_no"`
	PayOrderNo   string    `gorm:"column:pay_order_no" json:"pay_order_no"`
	UID          uint      `gorm:"column:uid" json:"uid"`
	Type         string    `gorm:"column:type" json:"type"`
	Channel      int       `gorm:"column:channel" json:"channel"`
	Account      string    `gorm:"column:account" json:"account"`
	Username     string    `gorm:"column:username" json:"username"`
	Money        float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Costmoney    float64   `gorm:"column:costmoney;type:decimal(10,2)" json:"costmoney"`
	Paytime      time.Time `gorm:"column:paytime" json:"paytime"`
	Status       int       `gorm:"column:status" json:"status"`
	API          int       `gorm:"column:api" json:"api"`
	Desc         string    `gorm:"column:desc" json:"desc"`
	Result       string    `gorm:"column:result" json:"result"`
}

func (Transfer) TableName() string {
	return "transfer"
}

// 分账接收人表(支付宝)
type PsReceiver struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Channel   int       `gorm:"column:channel" json:"channel"`
	UID       uint      `gorm:"column:uid" json:"uid"`
	Account   string    `gorm:"column:account" json:"account"`
	Name      string    `gorm:"column:name" json:"name"`
	Rate      string    `gorm:"column:rate" json:"rate"`
	Minmoney  string    `gorm:"column:minmoney" json:"minmoney"`
	Status    int       `gorm:"column:status" json:"status"`
	Addtime   time.Time `gorm:"column:addtime" json:"addtime"`
}

func (PsReceiver) TableName() string {
	return "psreceiver"
}

// 分账接收人表2(银行卡)
type PsReceiver2 struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Channel   int       `gorm:"column:channel" json:"channel"`
	UID       uint      `gorm:"column:uid" json:"uid"`
	BankType  int       `gorm:"column:bank_type" json:"bank_type"`
	CardID    string    `gorm:"column:card_id" json:"card_id"`
	CardName  string    `gorm:"column:card_name" json:"card_name"`
	TelNo     string    `gorm:"column:tel_no" json:"tel_no"`
	CertID    string    `gorm:"column:cert_id" json:"cert_id"`
	BankCode  string    `gorm:"column:bank_code" json:"bank_code"`
	ProvCode  string    `gorm:"column:prov_code" json:"prov_code"`
	AreaCode  string    `gorm:"column:area_code" json:"area_code"`
	Settleid  string    `gorm:"column:settleid" json:"settleid"`
	Rate      string    `gorm:"column:rate" json:"rate"`
	Minmoney  string    `gorm:"column:minmoney" json:"minmoney"`
	Status    int       `gorm:"column:status" json:"status"`
	Addtime   time.Time `gorm:"column:addtime" json:"addtime"`
}

func (PsReceiver2) TableName() string {
	return "psreceiver2"
}

// 分账订单表
type PsOrder struct {
	ID          uint      `gorm:"primaryKey;column:id" json:"id"`
	RID         int       `gorm:"column:rid" json:"rid"`
	TradeNo     string    `gorm:"column:trade_no" json:"trade_no"`
	ApiTradeNo  string    `gorm:"column:api_trade_no" json:"api_trade_no"`
	SettleNo    string    `gorm:"column:settle_no" json:"settle_no"`
	Money       float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Status      int       `gorm:"column:status" json:"status"`
	Result      string    `gorm:"column:result" json:"result"`
	Addtime     time.Time `gorm:"column:addtime" json:"addtime"`
}

func (PsOrder) TableName() string {
	return "psorder"
}
