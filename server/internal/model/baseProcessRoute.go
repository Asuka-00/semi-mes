package model

import (
	"time"
)

type BaseProcessRoute struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProductID   int        `gorm:"column:product_id;type:int(11);not null" json:"productID"`
	RouteCode   string     `gorm:"column:route_code;type:text;not null" json:"routeCode"`
	RouteName   string     `gorm:"column:route_name;type:text;not null" json:"routeName"`
	Version     string     `gorm:"column:version;type:text" json:"version"`
	IsDefault   int        `gorm:"column:is_default;type:int(11);not null" json:"isDefault"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Status      int        `gorm:"column:status;type:int(11);not null" json:"status"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:datetime" json:"createdAt"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:datetime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime" json:"deletedAt"`
}

// TableName table name
func (m *BaseProcessRoute) TableName() string {
	return "base_process_route"
}

// BaseProcessRouteColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseProcessRouteColumnNames = map[string]bool{
	"id":          true,
	"product_id":  true,
	"route_code":  true,
	"route_name":  true,
	"version":     true,
	"is_default":  true,
	"description": true,
	"status":      true,
	"created_at":  true,
	"updated_at":  true,
	"deleted_at":  true,
}


