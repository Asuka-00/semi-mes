package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RuleCarrier = "carrier"
	RuleWafer   = "wafer"

	CarrierEmpty    = "empty"
	CarrierInUse    = "in_use"
	CarrierCleaning = "cleaning"
	CarrierDown     = "down"

	BindBound   = "bound"
	BindUnbound = "unbound"

	WaferActive    = "active"
	WaferScrapped  = "scrapped"
	WaferCompleted = "completed"
)

// WipCarrier is a FOUP, cassette, or other wafer container.
type WipCarrier struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CarrierNo   string         `gorm:"column:carrier_no;type:varchar(40);not null;uniqueIndex" json:"carrierNo"`
	CarrierType string         `gorm:"column:carrier_type;type:varchar(20);not null;index" json:"carrierType"`
	Capacity    int            `gorm:"column:capacity;not null" json:"capacity"`
	Status      string         `gorm:"column:status;type:varchar(20);not null;index" json:"status"`
	Location    string         `gorm:"column:location;type:varchar(80)" json:"location"`
	CleanCount  int            `gorm:"column:clean_count" json:"cleanCount"`
	CleanLimit  int            `gorm:"column:clean_limit" json:"cleanLimit"`
	Note        string         `gorm:"column:note;type:varchar(255)" json:"note"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *WipCarrier) TableName() string { return "wip_carrier" }

// WipCarrierBind is the current or previous lot assignment of a carrier.
type WipCarrierBind struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CarrierID uint64     `gorm:"column:carrier_id;not null;index" json:"carrierID"`
	LotID     uint64     `gorm:"column:lot_id;not null;index" json:"lotID"`
	Status    string     `gorm:"column:status;type:varchar(20);not null;index" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *WipCarrierBind) TableName() string { return "wip_carrier_bind" }

// WipWafer is one wafer inside a lot.
type WipWafer struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	WaferNo   string     `gorm:"column:wafer_no;type:varchar(40);not null;uniqueIndex" json:"waferNo"`
	LotID     uint64     `gorm:"column:lot_id;not null;index" json:"lotID"`
	CarrierID uint64     `gorm:"column:carrier_id;index" json:"carrierID"`
	Slot      int        `gorm:"column:slot" json:"slot"`
	Status    string     `gorm:"column:status;type:varchar(20);not null;index" json:"status"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *WipWafer) TableName() string { return "wip_wafer" }

// WipWaferHistory records slot, carrier, scrap, and lot-move events for one wafer.
type WipWaferHistory struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	WaferID     uint64     `gorm:"column:wafer_id;not null;index" json:"waferID"`
	LotID       uint64     `gorm:"column:lot_id;not null;index" json:"lotID"`
	MoveID      uint64     `gorm:"column:move_id;index" json:"moveID"`
	EventType   string     `gorm:"column:event_type;type:varchar(20);not null" json:"eventType"`
	FromNodeKey string     `gorm:"column:from_node_key;type:varchar(64)" json:"fromNodeKey"`
	ToNodeKey   string     `gorm:"column:to_node_key;type:varchar(64)" json:"toNodeKey"`
	FromSlot    int        `gorm:"column:from_slot" json:"fromSlot"`
	ToSlot      int        `gorm:"column:to_slot" json:"toSlot"`
	CarrierID   uint64     `gorm:"column:carrier_id" json:"carrierID"`
	ReasonCode  string     `gorm:"column:reason_code;type:varchar(50)" json:"reasonCode"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *WipWaferHistory) TableName() string { return "wip_wafer_history" }
