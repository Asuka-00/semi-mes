package model

import (
	"time"
)

type SysRoleMenu struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RoleID    int        `gorm:"column:role_id;type:int(11);not null" json:"roleID"`
	MenuID    int        `gorm:"column:menu_id;type:int(11);not null" json:"menuID"`
	CreatedAt *time.Time `gorm:"column:created_at;type:datetime" json:"createdAt"`
	UpdatedAt *time.Time `gorm:"column:updated_at;type:datetime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName table name
func (m *SysRoleMenu) TableName() string {
	return "sys_role_menu"
}

// SysRoleMenuColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SysRoleMenuColumnNames = map[string]bool{
	"id":         true,
	"role_id":    true,
	"menu_id":    true,
	"created_at": true,
	"updated_at": true,
	"deleted_at": true,
}


