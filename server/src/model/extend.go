package model

import (
	"time"
)

// 分账记录表
type PsRecord struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	UID        uint      `gorm:"column:uid" json:"uid"`
	PsOrderID  int       `gorm:"column:ps_order_id" json:"ps_order_id"`
	TradeNo    string    `gorm:"column:trade_no" json:"trade_no"`
	Money      float64   `gorm:"column:money;type:decimal(10,2)" json:"money"`
	Rate       string    `gorm:"column:rate" json:"rate"`
	Fee        float64   `gorm:"column:fee;type:decimal(10,2)" json:"fee"`
	Status     int       `gorm:"column:status" json:"status"`
	Result     string    `gorm:"column:result" json:"result"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PsRecord) TableName() string {
	return "ps_record"
}

// 代理商表
type Agent struct {
	ID           uint      `gorm:"primaryKey;column:id" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	Password     string    `gorm:"column:password" json:"password"`
	Realname     string    `gorm:"column:realname" json:"realname"`
	Phone        string    `gorm:"column:phone" json:"phone"`
	Email        string    `gorm:"column:email" json:"email"`
	Level        int       `gorm:"column:level" json:"level"`
	ParentID     uint      `gorm:"column:parent_id" json:"parent_id"`
	Commission   string    `gorm:"column:commission" json:"commission"`
	TotalIncome  float64   `gorm:"column:total_income;type:decimal(10,2)" json:"total_income"`
	Balance      float64   `gorm:"column:balance;type:decimal(10,2)" json:"balance"`
	Status       int       `gorm:"column:status" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Agent) TableName() string {
	return "agent"
}

// 客服表
type Kefu struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Account   string    `gorm:"column:account" json:"account"`
	Password  string    `gorm:"column:password" json:"password"`
	Type      int       `gorm:"column:type" json:"type"`         // 1=QQ, 2=微信, 3=手机, 4=邮件
	Avatar    string    `gorm:"column:avatar" json:"avatar"`
	Qrcode    string    `gorm:"column:qrcode" json:"qrcode"`
	Link      string    `gorm:"column:link" json:"link"`
	Status    int       `gorm:"column:status" json:"status"`
	Sort      int       `gorm:"column:sort" json:"sort"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Kefu) TableName() string {
	return "kefu"
}

// 邮件队列表
type MailQueue struct {
	ID         uint      `gorm:"primaryKey;column:id" json:"id"`
	ToEmail    string    `gorm:"column:to_email" json:"to_email"`
	Subject    string    `gorm:"column:subject" json:"subject"`
	Body       string    `gorm:"column:body" json:"body"`
	Status     int       `gorm:"column:status" json:"status"`   // 0=待发送, 1=已发送, 2=失败
	Retry      int       `gorm:"column:retry" json:"retry"`
	SendAt     time.Time `gorm:"column:send_at" json:"send_at"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (MailQueue) TableName() string {
	return "mail_queue"
}

// 用户组转让记录表
type UserGroupTransfer struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	FromUID   uint      `gorm:"column:from_uid" json:"from_uid"`
	ToUID     uint      `gorm:"column:to_uid" json:"to_uid"`
	GroupID   uint      `gorm:"column:group_id" json:"group_id"`
	Price     float64   `gorm:"column:price;type:decimal(10,2)" json:"price"`
	Status    int       `gorm:"column:status" json:"status"`   // 0=待处理, 1=已完成, 2=已拒绝
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (UserGroupTransfer) TableName() string {
	return "user_group_transfer"
}
