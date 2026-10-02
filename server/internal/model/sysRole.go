package model

import (
	"time"

	"gorm.io/gorm"
)

type SysRole struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleCode    string         `gorm:"column:role_code;type:text;not null" json:"roleCode"`
	RoleName    string         `gorm:"column:role_name;type:text;not null" json:"roleName"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Status      int            `gorm:"column:status;type:integer;not null" json:"status"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *SysRole) TableName() string {
	return "sys_role"
}

// SysRoleColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SysRoleColumnNames = map[string]bool{
	"id":          true,
	"role_code":   true,
	"role_name":   true,
	"description": true,
	"status":      true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
}
