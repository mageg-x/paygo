package model

import (
	"time"
)

// 支付类型表
type PayType struct {
	ID       uint   `gorm:"primaryKey;column:id" json:"id"`
	Name     string `gorm:"column:name" json:"name"`
	Device   int    `gorm:"column:device" json:"device"`
	Showname string `gorm:"column:showname" json:"showname"`
	Status   int    `gorm:"column:status" json:"status"`
}

func (PayType) TableName() string {
	return "type"
}

// 插件表
type Plugin struct {
	Name       string `gorm:"primaryKey;column:name" json:"name"`
	Showname   string `gorm:"column:showname" json:"showname"`
	Author     string `gorm:"column:author" json:"author"`
	Link       string `gorm:"column:link" json:"link"`
	Types      string `gorm:"column:types" json:"types"`
	Transtypes string `gorm:"column:transtypes" json:"transtypes"`
	Status     int    `gorm:"column:status;default:1" json:"status"`     // 0=禁用, 1=启用
	Config     string `gorm:"column:config" json:"config"`                 // 插件私有配置JSON
}

func (Plugin) TableName() string {
	return "plugin"
}

// 支付通道表
type Channel struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	Mode       int       `gorm:"column:mode" json:"mode"`
	Type       int       `gorm:"column:type" json:"type"`
	Plugin     string    `gorm:"column:plugin" json:"plugin"`
	Name       string    `gorm:"column:name" json:"name"`
	Rate       float64   `gorm:"column:rate;type:decimal(5,2)" json:"rate"`
	Status     int       `gorm:"column:status" json:"status"`
	Apptype    string    `gorm:"column:apptype" json:"apptype"`
	Daytop     int       `gorm:"column:daytop" json:"daytop"`
	Daystatus  int       `gorm:"column:daystatus" json:"daystatus"`
	Paymin     string    `gorm:"column:paymin" json:"paymin"`
	Paymax     string    `gorm:"column:paymax" json:"paymax"`
	Appwxmp    int       `gorm:"column:appwxmp" json:"appwxmp"`
	Appwxa     int       `gorm:"column:appwxa" json:"appwxa"`
	Costrate   float64   `gorm:"column:costrate;type:decimal(5,2)" json:"costrate"`
	Config     string    `gorm:"column:config" json:"config"`
}

func (Channel) TableName() string {
	return "channel"
}

// 轮询配置表
type Roll struct {
	ID     uint    `gorm:"primaryKey;column:id" json:"id"`
	Type   int     `gorm:"column:type" json:"type"`
	Name   string  `gorm:"column:name" json:"name"`
	Kind   int     `gorm:"column:kind" json:"kind"`
	Info   string  `gorm:"column:info" json:"info"`
	Status int     `gorm:"column:status" json:"status"`
	Index  int     `gorm:"column:index" json:"index"`
}

func (Roll) TableName() string {
	return "roll"
}

// 子通道表
type SubChannel struct {
	ID       uint      `gorm:"primaryKey;column:id" json:"id"`
	Channel  int       `gorm:"column:channel" json:"channel"`
	UID      uint      `gorm:"column:uid" json:"uid"`
	Name     string    `gorm:"column:name" json:"name"`
	Status   int       `gorm:"column:status" json:"status"`
	Info     string    `gorm:"column:info" json:"info"`
	Addtime  time.Time `gorm:"column:addtime" json:"addtime"`
	Usetime  time.Time `gorm:"column:usetime" json:"usetime"`
	ApplyID  int       `gorm:"column:apply_id" json:"apply_id"`
}

func (SubChannel) TableName() string {
	return "subchannel"
}

// 微信配置表
type Weixin struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	Type         int       `gorm:"column:type" json:"type"`
	Name         string    `gorm:"column:name" json:"name"`
	Status       int       `gorm:"column:status" json:"status"`
	Appid        string    `gorm:"column:appid" json:"appid"`
	Appsecret    string    `gorm:"column:appsecret" json:"appsecret"`
	AccessToken  string    `gorm:"column:access_token" json:"access_token"`
	Addtime      time.Time `gorm:"column:addtime" json:"addtime"`
	Updatetime   time.Time `gorm:"column:updatetime" json:"updatetime"`
	Expiretime   time.Time `gorm:"column:expiretime" json:"expiretime"`
}

func (Weixin) TableName() string {
	return "weixin"
}
