package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	RealName  string         `gorm:"type:varchar(50)" json:"real_name"`
	Email     string         `gorm:"type:varchar(100)" json:"email"`
	Phone     string         `gorm:"type:varchar(20)" json:"phone"`
	Status    int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Roles     []Role         `gorm:"many2many:sys_user_role;" json:"roles,omitempty"`
}

func (User) TableName() string {
	return "sys_user"
}

type Role struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleCode    string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"role_code"`
	RoleName    string         `gorm:"type:varchar(50);not null" json:"role_name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	Status      int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Menus       []Menu         `gorm:"many2many:sys_role_menu;" json:"menus,omitempty"`
}

func (Role) TableName() string {
	return "sys_role"
}

type Menu struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ParentID       int64          `gorm:"default:0;not null" json:"parent_id"`
	MenuType       int8           `gorm:"type:tinyint;not null" json:"menu_type"`
	MenuName       string         `gorm:"type:varchar(50);not null" json:"menu_name"`
	PermissionCode string         `gorm:"type:varchar(100)" json:"permission_code"`
	RoutePath      string         `gorm:"type:varchar(200)" json:"route_path"`
	ComponentPath  string         `gorm:"type:varchar(200)" json:"component_path"`
	Icon           string         `gorm:"type:varchar(50)" json:"icon"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	Status         int8           `gorm:"type:tinyint;default:1;not null" json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Children       []Menu         `gorm:"-" json:"children,omitempty"`
}

func (Menu) TableName() string {
	return "sys_menu"
}

type UserRole struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	RoleID    int64     `gorm:"not null;index" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserRole) TableName() string {
	return "sys_user_role"
}

type RoleMenu struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	RoleID    int64     `gorm:"not null;index" json:"role_id"`
	MenuID    int64     `gorm:"not null;index" json:"menu_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (RoleMenu) TableName() string {
	return "sys_role_menu"
}
