package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseProductionLine struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	WorkshopID  int            `gorm:"column:workshop_id;type:int(11);not null" json:"workshopID"`
	LineCode    string         `gorm:"column:line_code;type:text;not null" json:"lineCode"`
	LineName    string         `gorm:"column:line_name;type:text;not null" json:"lineName"`
	Capacity    int            `gorm:"column:capacity;type:int(11);not null" json:"capacity"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Status      int            `gorm:"column:status;type:int(11);not null" json:"status"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:datetime" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:datetime" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseProductionLine) TableName() string {
	return "base_production_line"
}

// BaseProductionLineColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseProductionLineColumnNames = map[string]bool{
	"id":          true,
	"workshop_id": true,
	"line_code":   true,
	"line_name":   true,
	"capacity":    true,
	"description": true,
	"status":      true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
}
