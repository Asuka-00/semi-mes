package model

import (
	"time"

	"gorm.io/gorm"
)

type SysUserRole struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    int            `gorm:"column:user_id;type:integer;not null" json:"userID"`
	RoleID    int            `gorm:"column:role_id;type:integer;not null" json:"roleID"`
	CreatedAt *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *SysUserRole) TableName() string {
	return "sys_user_role"
}

// SysUserRoleColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SysUserRoleColumnNames = map[string]bool{
	"id":         true,
	"user_id":    true,
	"role_id":    true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
}
