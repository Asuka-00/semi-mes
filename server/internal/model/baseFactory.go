package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseFactory struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	FactoryCode string         `gorm:"column:factory_code;type:text;not null" json:"factoryCode"`
	FactoryName string         `gorm:"column:factory_name;type:text;not null" json:"factoryName"`
	Address     string         `gorm:"column:address;type:text" json:"address"`
	Contact     string         `gorm:"column:contact;type:text" json:"contact"`
	Phone       string         `gorm:"column:phone;type:text" json:"phone"`
	Status      int            `gorm:"column:status;type:integer;not null" json:"status"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseFactory) TableName() string {
	return "base_factory"
}

// BaseFactoryColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseFactoryColumnNames = map[string]bool{
	"id":           true,
	"factory_code": true,
	"factory_name": true,
	"address":      true,
	"contact":      true,
	"phone":        true,
	"status":       true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
}
