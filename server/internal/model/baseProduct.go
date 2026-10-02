package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseProduct struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProductCode string         `gorm:"column:product_code;type:text;not null" json:"productCode"`
	ProductName string         `gorm:"column:product_name;type:text;not null" json:"productName"`
	ProductType string         `gorm:"column:product_type;type:text" json:"productType"`
	Version     string         `gorm:"column:version;type:text" json:"version"`
	Description string         `gorm:"column:description;type:text" json:"description"`
	Status      int            `gorm:"column:status;type:integer;not null" json:"status"`
	CreatedAt   *time.Time     `gorm:"column:created_at;type:timestamp" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at;type:timestamp" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseProduct) TableName() string {
	return "base_product"
}

// BaseProductColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseProductColumnNames = map[string]bool{
	"id":           true,
	"product_code": true,
	"product_name": true,
	"product_type": true,
	"version":      true,
	"description":  true,
	"status":       true,
	"created_at":   true,
	"updated_at":   true,
	"deleted_at":   true,
}
