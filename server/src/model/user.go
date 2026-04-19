package model

import (
	"time"
)

// 商户表
type User struct {
	UID          uint      `gorm:"primaryKey;column:uid" json:"uid"`
	GID          uint      `gorm:"column:gid" json:"gid"`
	Upid         uint      `gorm:"column:upid" json:"upid"`
	Key          string    `gorm:"column:key" json:"key"`
	Pwd          string    `gorm:"column:pwd" json:"-"`
	Account      string    `gorm:"column:account" json:"account"`
	Username     string    `gorm:"column:username" json:"username"`
	Codename     string    `gorm:"column:codename" json:"codename"`
	SettleID     int       `gorm:"column:settle_id" json:"settle_id"`
	AlipayUID    string    `gorm:"column:alipay_uid" json:"alipay_uid"`
	QqUID        string    `gorm:"column:qq_uid" json:"qq_uid"`
	WxUID        string    `gorm:"column:wx_uid" json:"wx_uid"`
	Money        float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Email        string    `gorm:"column:email" json:"email"`
	Phone        string    `gorm:"column:phone" json:"phone"`
	Qq           string    `gorm:"column:qq" json:"qq"`
	URL          string    `gorm:"column:url" json:"url"`
	Cert         int       `gorm:"column:cert" json:"cert"`
	Certtype     int       `gorm:"column:certtype" json:"certtype"`
	Certmethod   int       `gorm:"column:certmethod" json:"certmethod"`
	Certno       string    `gorm:"column:certno" json:"certno"`
	Certname     string    `gorm:"column:certname" json:"certname"`
	Certtime     time.Time `gorm:"column:certtime" json:"certtime"`
	Certtoken    string    `gorm:"column:certtoken" json:"certtoken"`
	Certcorpno   string    `gorm:"column:certcorpno" json:"certcorpno"`
	Certcorpname string    `gorm:"column:certcorpname" json:"certcorpname"`
	Addtime      time.Time `gorm:"column:addtime" json:"addtime"`
	Lasttime     time.Time `gorm:"column:lasttime" json:"lasttime"`
	Endtime      time.Time `gorm:"column:endtime" json:"endtime"`
	Level        int       `gorm:"column:level" json:"level"`
	Pay          int       `gorm:"column:pay" json:"pay"`
	Settle       int       `gorm:"column:settle" json:"settle"`
	Keylogin     int       `gorm:"column:keylogin" json:"keylogin"`
	Apply        int       `gorm:"column:apply" json:"apply"`
	Mode         int       `gorm:"column:mode" json:"mode"`
	Status       int       `gorm:"column:status" json:"status"`
	Refund       int       `gorm:"column:refund" json:"refund"`
	Transfer     int       `gorm:"column:transfer" json:"transfer"`
	Keytype      int       `gorm:"column:keytype" json:"keytype"`
	Publickey    string    `gorm:"column:publickey" json:"publickey"`
	Channelinfo  string    `gorm:"column:channelinfo" json:"channelinfo"`
	Ordername    string    `gorm:"column:ordername" json:"ordername"`
	Msgconfig    string    `gorm:"column:msgconfig" json:"msgconfig"`
}

func (User) TableName() string {
	return "user"
}

// 用户组
type Group struct {
	GID        uint    `gorm:"primaryKey;column:gid" json:"gid"`
	Name       string  `gorm:"column:name" json:"name"`
	Info       string  `gorm:"column:info" json:"info"`
	Isbuy      int     `gorm:"column:isbuy" json:"isbuy"`
	Price      float64 `gorm:"column:price;type:decimal(10,2)" json:"price"`
	Sort       int     `gorm:"column:sort" json:"sort"`
	Expire     int     `gorm:"column:expire" json:"expire"`
	SettleOpen int     `gorm:"column:settle_open" json:"settle_open"`
	SettleType int     `gorm:"column:settle_type" json:"settle_type"`
	SettleRate string  `gorm:"column:settle_rate" json:"settle_rate"`
	Config     string  `gorm:"column:config" json:"gopay/config"`
	Settings   string  `gorm:"column:settings" json:"settings"`
}

func (Group) TableName() string {
	return "group"
}

// 资金变动记录
type Record struct {
	ID       uint      `gorm:"primaryKey;column:id" json:"id"`
	UID      uint      `gorm:"column:uid" json:"uid"`
	Action   int       `gorm:"column:action" json:"action"`
	Money    float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Oldmoney float64   `gorm:"column:oldmoney;type:decimal(10,2)" json:"oldmoney"`
	Newmoney float64   `gorm:"column:newmoney;type:decimal(10,2)" json:"newmoney"`
	Type     string    `gorm:"column:type" json:"type"`
	TradeNo  string    `gorm:"column:trade_no" json:"trade_no"`
	Date     time.Time `gorm:"column:date" json:"date"`
}

func (Record) TableName() string {
	return "record"
}

// 操作日志
type Log struct {
	ID   uint      `gorm:"primaryKey;column:id" json:"id"`
	UID  uint      `gorm:"column:uid" json:"uid"`
	Type string    `gorm:"column:type" json:"type"`
	Date time.Time `gorm:"column:date" json:"date"`
	IP   string    `gorm:"column:ip" json:"ip"`
	City string    `gorm:"column:city" json:"city"`
	Data string    `gorm:"column:data" json:"data"`
}

func (Log) TableName() string {
	return "log"
}
