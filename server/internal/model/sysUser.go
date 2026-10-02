package model

import (
	"time"

	"gorm.io/gorm"
)

type SysUser struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username  string         `gorm:"column:username;type:text;not null" json:"username"`
	Password  string         `gorm:"column:password;type:text;not null" json:"password"`
	RealName  string         `gorm:"column:real_name;type:text" json:"realName"`
	Email     string         `gorm:"column:email;type:text" json:"email"`
	Phone     string         `gorm:"column:phone;type:text" json:"phone"`
	Status    int            `gorm:"column:status;type:integer;not null" json:"status"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *SysUser) TableName() string {
	return "sys_user"
}

// SysUserColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SysUserColumnNames = map[string]bool{
	"id":         true,
	"username":   true,
	"password":   true,
	"real_name":  true,
	"email":      true,
	"phone":      true,
	"status":     true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
}
