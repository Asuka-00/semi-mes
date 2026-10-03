package model

import "time"

// SysAuditLog is one recorded operation. Rows are append-only.
type SysAuditLog struct {
	ID         uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID     uint64     `gorm:"column:user_id;index" json:"userId"`
	Username   string     `gorm:"column:username;type:varchar(64);index" json:"username"`
	IP         string     `gorm:"column:ip;type:varchar(64)" json:"ip"`
	Module     string     `gorm:"column:module;type:varchar(40);index" json:"module"`
	Action     string     `gorm:"column:action;type:varchar(40);index" json:"action"`
	EntityType string     `gorm:"column:entity_type;type:varchar(40);index" json:"entityType"`
	EntityID   uint64     `gorm:"column:entity_id;index" json:"entityId"`
	EntityCode string     `gorm:"column:entity_code;type:varchar(80);index" json:"entityCode"`
	Summary    string     `gorm:"column:summary;type:varchar(255)" json:"summary"`
	BeforeJSON string     `gorm:"column:before_json;type:text" json:"beforeJson"`
	AfterJSON  string     `gorm:"column:after_json;type:text" json:"afterJson"`
	DiffJSON   string     `gorm:"column:diff_json;type:text" json:"diffJson"`
	CreatedAt  *time.Time `gorm:"column:created_at;index" json:"createdAt"`
}

func (m *SysAuditLog) TableName() string { return "sys_audit_log" }
