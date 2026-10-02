package model

import (
	"time"

	"gorm.io/gorm"
)

type SysMenu struct {
	ID             uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParentID       int            `gorm:"column:parent_id;type:integer;not null" json:"parentID"`
	MenuType       int            `gorm:"column:menu_type;type:integer;not null" json:"menuType"`
	MenuName       string         `gorm:"column:menu_name;type:text;not null" json:"menuName"`
	PermissionCode string         `gorm:"column:permission_code;type:text" json:"permissionCode"`
	RouteName      string         `gorm:"column:route_name;type:text" json:"routeName"`
	RoutePath      string         `gorm:"column:route_path;type:text" json:"routePath"`
	ComponentPath  string         `gorm:"column:component_path;type:text" json:"componentPath"`
	Icon           string         `gorm:"column:icon;type:text" json:"icon"`
	SortOrder      int            `gorm:"column:sort_order;type:integer;not null" json:"sortOrder"`
	Status         int            `gorm:"column:status;type:integer;not null" json:"status"`
	CreatedAt      *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *SysMenu) TableName() string {
	return "sys_menu"
}

// SysMenuColumnNames Whitelist for custom query fields to prevent sql injection attacks
var SysMenuColumnNames = map[string]bool{
	"id":              true,
	"parent_id":       true,
	"menu_type":       true,
	"menu_name":       true,
	"permission_code": true,
	"route_name":      true,
	"route_path":      true,
	"component_path":  true,
	"icon":            true,
	"sort_order":      true,
	"status":          true,
	"created_at":      true,
	"updated_at":      true,
	"deleted_at":      true,
}
