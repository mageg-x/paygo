package model

import (
	"time"
)

// 系统配置表
type Config struct {
	K string `gorm:"primaryKey;column:k" json:"k"`
	V string `gorm:"column:v" json:"v"`
}

func (Config) TableName() string {
	return "config"
}

// 缓存表
type Cache struct {
	K      string `gorm:"primaryKey;column:k" json:"k"`
	V      string `gorm:"column:v" json:"v"`
	Expire int    `gorm:"column:expire" json:"expire"`
}

func (Cache) TableName() string {
	return "cache"
}

// 公告表
type Anounce struct {
	ID      uint      `gorm:"primaryKey;column:id" json:"id"`
	Content string    `gorm:"column:content" json:"content"`
	Color   string    `gorm:"column:color" json:"color"`
	Sort    int       `gorm:"column:sort" json:"sort"`
	Addtime time.Time `gorm:"column:addtime" json:"addtime"`
	Status  int       `gorm:"column:status" json:"status"`
}

func (Anounce) TableName() string {
	return "anounce"
}

// 注册码表
type RegCode struct {
	ID      uint      `gorm:"primaryKey;column:id" json:"id"`
	UID     uint      `gorm:"column:uid" json:"uid"`
	Scene   string    `gorm:"column:scene" json:"scene"`
	Type    int       `gorm:"column:type" json:"type"`
	Code    string    `gorm:"column:code" json:"code"`
	To      string    `gorm:"column:to" json:"to"`
	Time    int       `gorm:"column:time" json:"time"`
	IP      string    `gorm:"column:ip" json:"ip"`
	Status  int       `gorm:"column:status" json:"status"`
	Errcount int      `gorm:"column:errcount" json:"errcount"`
}

func (RegCode) TableName() string {
	return "regcode"
}

// 邀请码表
type InviteCode struct {
	ID      uint       `gorm:"primaryKey;column:id" json:"id"`
	Code    string     `gorm:"column:code" json:"code"`
	Addtime time.Time  `gorm:"column:addtime" json:"addtime"`
	Usetime *time.Time `gorm:"column:usetime" json:"usetime"`
	UID     *uint      `gorm:"column:uid" json:"uid"`
	Status  int        `gorm:"column:status" json:"status"`
}

func (InviteCode) TableName() string {
	return "invitecode"
}

// 风控记录表
type Risk struct {
	ID      uint      `gorm:"primaryKey;column:id" json:"id"`
	UID     uint      `gorm:"column:uid" json:"uid"`
	Type    int       `gorm:"column:type" json:"type"`
	URL     string    `gorm:"column:url" json:"url"`
	Content string    `gorm:"column:content" json:"content"`
	Date    time.Time `gorm:"column:date" json:"date"`
	Status  int       `gorm:"column:status" json:"status"`
}

func (Risk) TableName() string {
	return "risk"
}

// 域名表
type Domain struct {
	ID      uint       `gorm:"primaryKey;column:id" json:"id"`
	UID     uint       `gorm:"column:uid" json:"uid"`
	Domain  string     `gorm:"column:domain" json:"domain"`
	Status  int        `gorm:"column:status" json:"status"`
	Addtime *time.Time `gorm:"column:addtime" json:"addtime"`
	Endtime *time.Time `gorm:"column:endtime" json:"endtime"`
}

func (Domain) TableName() string {
	return "domain"
}

// 黑名单表
type Blacklist struct {
	ID      uint       `gorm:"primaryKey;column:id" json:"id"`
	Type    int        `gorm:"column:type" json:"type"`
	Content string     `gorm:"column:content" json:"content"`
	Addtime time.Time  `gorm:"column:addtime" json:"addtime"`
	Endtime *time.Time `gorm:"column:endtime" json:"endtime"`
	Remark  string     `gorm:"column:remark" json:"remark"`
}

func (Blacklist) TableName() string {
	return "blacklist"
}
