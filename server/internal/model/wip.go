package model

import (
	"time"

	"gorm.io/gorm"
)

// Work order and lot status values.
const (
	OrderCreated    = "created"
	OrderReleased   = "released"
	OrderInProgress = "in_progress"
	OrderCompleted  = "completed"
	OrderClosed     = "closed"

	LotWaiting   = "waiting"
	LotRunning   = "running"
	LotHold      = "hold"
	LotCompleted = "completed"
	LotScrapped  = "scrapped"
	LotMerged    = "merged"

	LotTypeProduction  = "production"
	LotTypeEngineering = "engineering"

	LinkSplit = "split"
	LinkMerge = "merge"

	EventStart    = "start"
	EventAdvance  = "advance"
	EventHold     = "hold"
	EventRelease  = "release"
	EventSplit    = "split"
	EventMerge    = "merge"
	EventComplete = "complete"
	EventTrackIn  = "track_in"
	EventTrackOut = "track_out"
	EventAbort    = "abort"
	EventPass     = "pass"
	EventScrap    = "scrap"

	MoveOpen      = "open"
	MoveCompleted = "completed"
	MoveAborted   = "aborted"

	EqpIdle = "idle"
	EqpDown = "down"
)

// WipWorkOrder releases lots onto one published route version.
type WipWorkOrder struct {
	ID             uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrderNo        string         `gorm:"column:order_no;type:varchar(40);not null;uniqueIndex" json:"orderNo"`
	ProductID      uint64         `gorm:"column:product_id;not null;index" json:"productID"`
	RouteVersionID uint64         `gorm:"column:route_version_id;not null;index" json:"routeVersionID"`
	PlannedQty     int            `gorm:"column:planned_qty;not null" json:"plannedQty"`
	ReleasedQty    int            `gorm:"column:released_qty" json:"releasedQty"`
	CompletedQty   int            `gorm:"column:completed_qty" json:"completedQty"`
	NextLotSeq     int            `gorm:"column:next_lot_seq" json:"nextLotSeq"`
	Priority       int            `gorm:"column:priority" json:"priority"`
	DueDate        *time.Time     `gorm:"column:due_date" json:"dueDate"`
	Status         string         `gorm:"column:status;type:varchar(20);not null;index" json:"status"`
	Note           string         `gorm:"column:note;type:text" json:"note"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *WipWorkOrder) TableName() string { return "wip_work_order" }

// WipLot is one physical batch bound to a route version and its current node.
type WipLot struct {
	ID             uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LotNo          string         `gorm:"column:lot_no;type:varchar(50);not null;uniqueIndex" json:"lotNo"`
	OrderID        uint64         `gorm:"column:order_id;not null;index" json:"orderID"`
	ProductID      uint64         `gorm:"column:product_id;not null" json:"productID"`
	RouteVersionID uint64         `gorm:"column:route_version_id;not null;index" json:"routeVersionID"`
	CurrentNodeKey string         `gorm:"column:current_node_key;type:varchar(64);not null" json:"currentNodeKey"`
	Quantity       int            `gorm:"column:quantity;not null" json:"quantity"`
	Priority       int            `gorm:"column:priority" json:"priority"`
	LotType        string         `gorm:"column:lot_type;type:varchar(20);not null" json:"lotType"`
	Status         string         `gorm:"column:status;type:varchar(20);not null;index" json:"status"`
	HoldReasonCode string         `gorm:"column:hold_reason_code;type:varchar(50)" json:"holdReasonCode"`
	HoldReason     string         `gorm:"column:hold_reason;type:varchar(255)" json:"holdReason"`
	ReworkJSON     string         `gorm:"column:rework_json;type:text" json:"reworkJson"`
	ArrivedAt      *time.Time     `gorm:"column:arrived_at" json:"arrivedAt"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *WipLot) TableName() string { return "wip_lot" }

// WipLotHistory records start, advance, hold, release, split, and merge.
type WipLotHistory struct {
	ID           uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LotID        uint64     `gorm:"column:lot_id;not null;index" json:"lotID"`
	EventType    string     `gorm:"column:event_type;type:varchar(20);not null" json:"eventType"`
	FromNodeKey  string     `gorm:"column:from_node_key;type:varchar(64)" json:"fromNodeKey"`
	ToNodeKey    string     `gorm:"column:to_node_key;type:varchar(64)" json:"toNodeKey"`
	EdgeKey      string     `gorm:"column:edge_key;type:varchar(64)" json:"edgeKey"`
	ReasonCode   string     `gorm:"column:reason_code;type:varchar(50)" json:"reasonCode"`
	Reason       string     `gorm:"column:reason;type:varchar(255)" json:"reason"`
	Quantity     int        `gorm:"column:quantity" json:"quantity"`
	RelatedLotID uint64     `gorm:"column:related_lot_id" json:"relatedLotID"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"createdAt"`
}

// TableName table name
func (m *WipLotHistory) TableName() string { return "wip_lot_history" }

// WipLotLink is the parent-child record for split and merge.
type WipLotLink struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParentLotID uint64     `gorm:"column:parent_lot_id;not null;index" json:"parentLotID"`
	ChildLotID  uint64     `gorm:"column:child_lot_id;not null;index" json:"childLotID"`
	LinkType    string     `gorm:"column:link_type;type:varchar(20);not null" json:"linkType"`
	Quantity    int        `gorm:"column:quantity" json:"quantity"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"createdAt"`
}

// TableName table name
func (m *WipLotLink) TableName() string { return "wip_lot_link" }
