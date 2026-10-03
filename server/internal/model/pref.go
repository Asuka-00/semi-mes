package model

import "time"

// SysUserPref stores one user's list settings for a page.
type SysUserPref struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint64     `gorm:"column:user_id;uniqueIndex:uk_user_page;not null" json:"userId"`
	PageKey   string     `gorm:"column:page_key;type:varchar(80);uniqueIndex:uk_user_page;not null" json:"pageKey"`
	Payload   string     `gorm:"column:payload;type:text" json:"payload"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *SysUserPref) TableName() string { return "sys_user_pref" }
