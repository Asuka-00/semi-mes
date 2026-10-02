package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseRouteVersion is one immutable snapshot of a process route after release.
type BaseRouteVersion struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RouteID   uint64         `gorm:"column:route_id;not null;index" json:"routeID"`
	VersionNo int            `gorm:"column:version_no;not null" json:"versionNo"`
	State     string         `gorm:"column:state;type:varchar(20);not null" json:"state"`
	Note      string         `gorm:"column:note;type:text" json:"note"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseRouteVersion) TableName() string { return "base_route_version" }

// BaseRouteNode is a vertex on a route version.
type BaseRouteNode struct {
	ID             uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	VersionID      uint64         `gorm:"column:version_id;not null;index" json:"versionID"`
	NodeKey        string         `gorm:"column:node_key;type:varchar(64);not null" json:"nodeKey"`
	NodeType       string         `gorm:"column:node_type;type:varchar(20);not null" json:"nodeType"`
	Name           string         `gorm:"column:name;type:varchar(100)" json:"name"`
	OperationID    uint64         `gorm:"column:operation_id" json:"operationID"`
	RecipeID       uint64         `gorm:"column:recipe_id" json:"recipeID"`
	EquipmentGroup string         `gorm:"column:equipment_group;type:varchar(100)" json:"equipmentGroup"`
	PosX           float64        `gorm:"column:pos_x" json:"posX"`
	PosY           float64        `gorm:"column:pos_y" json:"posY"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseRouteNode) TableName() string { return "base_route_node" }

// BaseRouteEdge is a directed link, optionally conditional or a rework loop.
type BaseRouteEdge struct {
	ID            uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	VersionID     uint64         `gorm:"column:version_id;not null;index" json:"versionID"`
	EdgeKey       string         `gorm:"column:edge_key;type:varchar(64);not null" json:"edgeKey"`
	FromKey       string         `gorm:"column:from_key;type:varchar(64);not null" json:"fromKey"`
	ToKey         string         `gorm:"column:to_key;type:varchar(64);not null" json:"toKey"`
	EdgeKind      string         `gorm:"column:edge_kind;type:varchar(20);not null" json:"edgeKind"`
	IsDefault     int            `gorm:"column:is_default" json:"isDefault"`
	Priority      int            `gorm:"column:priority" json:"priority"`
	ConditionJSON string         `gorm:"column:condition_json;type:text" json:"conditionJson"`
	MaxRework     int            `gorm:"column:max_rework" json:"maxRework"`
	OnExceed      string         `gorm:"column:on_exceed;type:varchar(20)" json:"onExceed"`
	Label         string         `gorm:"column:label;type:varchar(100)" json:"label"`
	CreatedAt     *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *BaseRouteEdge) TableName() string { return "base_route_edge" }
