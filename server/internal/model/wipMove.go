package model

import (
	"time"

	"gorm.io/gorm"
)

// EqpEquipment is the equipment master. Module 4 extends this table; do not add a second one.
// Status is a SEMI E10-style state. Manufacturer is the vendor. ModelName is the model.
type EqpEquipment struct {
	ID             uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EquipmentCode  string         `gorm:"column:equipment_code;type:varchar(50);not null;uniqueIndex" json:"equipmentCode"`
	EquipmentName  string         `gorm:"column:equipment_name;type:varchar(100);not null" json:"equipmentName"`
	EquipmentGroup string         `gorm:"column:equipment_group;type:varchar(100);index" json:"equipmentGroup"`
	EquipmentType  string         `gorm:"column:equipment_type;type:varchar(50)" json:"equipmentType"`
	Status         string         `gorm:"column:status;type:varchar(30);not null" json:"status"`
	LineID         uint64         `gorm:"column:line_id" json:"lineID"`
	ModelName      string         `gorm:"column:model_name;type:varchar(50)" json:"modelName"`
	Manufacturer   string         `gorm:"column:manufacturer;type:varchar(100)" json:"manufacturer"`
	SerialNo       string         `gorm:"column:serial_no;type:varchar(80)" json:"serialNo"`
	Location       string         `gorm:"column:location;type:varchar(100)" json:"location"`
	ChamberCount   int            `gorm:"column:chamber_count" json:"chamberCount"`
	Capacity       int            `gorm:"column:capacity" json:"capacity"`
	InstallDate    *time.Time     `gorm:"column:install_date" json:"installDate"`
	ResumeState    string         `gorm:"column:resume_state;type:varchar(30)" json:"resumeState"`
	CreatedAt      *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName table name
func (m *EqpEquipment) TableName() string { return "eqp_equipment" }

// WipMove is one Track In / Track Out (or abort, or a non-operation pass).
type WipMove struct {
	ID               uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LotID            uint64     `gorm:"column:lot_id;not null;index" json:"lotID"`
	NodeKey          string     `gorm:"column:node_key;type:varchar(64);not null" json:"nodeKey"`
	OperationID      uint64     `gorm:"column:operation_id" json:"operationID"`
	EquipmentID      uint64     `gorm:"column:equipment_id" json:"equipmentID"`
	RecipeID         uint64     `gorm:"column:recipe_id" json:"recipeID"`
	OperatorID       uint64     `gorm:"column:operator_id" json:"operatorID"`
	QtyIn            int        `gorm:"column:qty_in" json:"qtyIn"`
	QtyOut           int        `gorm:"column:qty_out" json:"qtyOut"`
	QtyScrap         int        `gorm:"column:qty_scrap" json:"qtyScrap"`
	ScrapReasonCode  string     `gorm:"column:scrap_reason_code;type:varchar(50)" json:"scrapReasonCode"`
	InspectionResult string     `gorm:"column:inspection_result;type:varchar(50)" json:"inspectionResult"`
	InspectionGrade  string     `gorm:"column:inspection_grade;type:varchar(50)" json:"inspectionGrade"`
	DefectCode       string     `gorm:"column:defect_code;type:varchar(50)" json:"defectCode"`
	State            string     `gorm:"column:state;type:varchar(20);not null;index" json:"state"`
	TrackInAt        *time.Time `gorm:"column:track_in_at" json:"trackInAt"`
	TrackOutAt       *time.Time `gorm:"column:track_out_at" json:"trackOutAt"`
	QueueSeconds     int        `gorm:"column:queue_seconds" json:"queueSeconds"`
	ProcessSeconds   int        `gorm:"column:process_seconds" json:"processSeconds"`
	FromNodeKey      string     `gorm:"column:from_node_key;type:varchar(64)" json:"fromNodeKey"`
	ToNodeKey        string     `gorm:"column:to_node_key;type:varchar(64)" json:"toNodeKey"`
	EdgeKey          string     `gorm:"column:edge_key;type:varchar(64)" json:"edgeKey"`
	ResolveAction    string     `gorm:"column:resolve_action;type:varchar(20)" json:"resolveAction"`
	ResolveReason    string     `gorm:"column:resolve_reason;type:varchar(50)" json:"resolveReason"`
	AbortReason      string     `gorm:"column:abort_reason;type:varchar(255)" json:"abortReason"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName table name
func (m *WipMove) TableName() string { return "wip_move" }
