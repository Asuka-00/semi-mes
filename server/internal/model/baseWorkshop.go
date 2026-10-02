package model

import (
	"time"
)

type BaseWorkshop struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FactoryID    int        `gorm:"column:factory_id;type:int(11);not null" json:"factoryID"`
	WorkshopCode string     `gorm:"column:workshop_code;type:text;not null" json:"workshopCode"`
	WorkshopName string     `gorm:"column:workshop_name;type:text;not null" json:"workshopName"`
	WorkshopType string     `gorm:"column:workshop_type;type:text" json:"workshopType"`
	Description  string     `gorm:"column:description;type:text" json:"description"`
	Status       int        `gorm:"column:status;type:int(11);not null" json:"status"`
	CreatedAt    *time.Time `gorm:"column:created_at;type:datetime" json:"createdAt"`
	UpdatedAt    *time.Time `gorm:"column:updated_at;type:datetime" json:"updatedAt"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName table name
func (m *BaseWorkshop) TableName() string {
	return "base_workshop"
}

// BaseWorkshopColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseWorkshopColumnNames = map[string]bool{
	"id":            true,
	"factory_id":    true,
	"workshop_code": true,
	"workshop_name": true,
	"workshop_type": true,
	"description":   true,
	"status":        true,
	"created_at":    true,
	"updated_at":    true,
	"deleted_at":    true,
}


