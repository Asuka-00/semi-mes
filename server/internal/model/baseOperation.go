package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseOperation struct {
	ID            uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RouteID       int            `gorm:"column:route_id;type:int(11);not null" json:"routeID"`
	OperationCode string         `gorm:"column:operation_code;type:text;not null" json:"operationCode"`
	OperationName string         `gorm:"column:operation_name;type:text;not null" json:"operationName"`
	OperationType string         `gorm:"column:operation_type;type:text" json:"operationType"`
	Sequence      int            `gorm:"column:sequence;type:int(11);not null" json:"sequence"`
	StandardTime  int            `gorm:"column:standard_time;type:int(11);not null" json:"standardTime"`
	Description   string         `gorm:"column:description;type:text" json:"description"`
	Status        int            `gorm:"column:status;type:int(11);not null" json:"status"`
	CreatedAt     *time.Time     `gorm:"column:created_at;type:datetime" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at;type:datetime" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseOperation) TableName() string {
	return "base_operation"
}

// BaseOperationColumnNames Whitelist for custom query fields to prevent sql injection attacks
var BaseOperationColumnNames = map[string]bool{
	"id":             true,
	"route_id":       true,
	"operation_code": true,
	"operation_name": true,
	"operation_type": true,
	"sequence":       true,
	"standard_time":  true,
	"description":    true,
	"status":         true,
	"created_at":     true,
	"updated_at":     true,
	"deleted_at":     true,
}
