package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	NoticeLotHold        = "lot_hold"
	NoticeReworkExceeded = "rework_exceeded"
	NoticePmDue          = "pm_due"
	NoticePmOverdue      = "pm_overdue"
	NoticeEqpDown        = "eqp_unscheduled_down"
)

// SysNotification is one in-app notice for a single user.
type SysNotification struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint64         `gorm:"column:user_id;index;not null" json:"userId"`
	Kind      string         `gorm:"column:kind;type:varchar(40);not null;index" json:"kind"`
	Params    string         `gorm:"column:params;type:text" json:"params"`
	RefType   string         `gorm:"column:ref_type;type:varchar(40)" json:"refType"`
	RefID     uint64         `gorm:"column:ref_id" json:"refId"`
	ReadAt    *time.Time     `gorm:"column:read_at" json:"readAt"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *SysNotification) TableName() string { return "sys_notification" }
