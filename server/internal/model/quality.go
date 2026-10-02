package model

import (
	"time"

	"gorm.io/gorm"
)

// QcInspectPlan is a measurement plan for one operation, optionally one product.
// ProductID 0 applies to every product that does not have its own plan.
type QcInspectPlan struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OperationID uint64         `gorm:"column:operation_id;not null;index" json:"operationID"`
	ProductID   uint64         `gorm:"column:product_id;index" json:"productID"`
	PlanName    string         `gorm:"column:plan_name;type:varchar(100);not null" json:"planName"`
	Enabled     bool           `gorm:"column:enabled" json:"enabled"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcInspectPlan) TableName() string { return "qc_inspect_plan" }

// QcInspectItem is one parameter on a plan. Nil limits mean that side is not checked.
type QcInspectItem struct {
	ID         uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PlanID     uint64         `gorm:"column:plan_id;not null;index" json:"planID"`
	ParamCode  string         `gorm:"column:param_code;type:varchar(50);not null" json:"paramCode"`
	ParamName  string         `gorm:"column:param_name;type:varchar(100)" json:"paramName"`
	Unit       string         `gorm:"column:unit;type:varchar(20)" json:"unit"`
	Target     *float64       `gorm:"column:target" json:"target"`
	LSL        *float64       `gorm:"column:lsl" json:"lsl"`
	USL        *float64       `gorm:"column:usl" json:"usl"`
	LCL        *float64       `gorm:"column:lcl" json:"lcl"`
	UCL        *float64       `gorm:"column:ucl" json:"ucl"`
	SampleSize int            `gorm:"column:sample_size" json:"sampleSize"`
	Required   bool           `gorm:"column:required" json:"required"`
	SortOrder  int            `gorm:"column:sort_order" json:"sortOrder"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcInspectItem) TableName() string { return "qc_inspect_item" }

// QcMeasurement is one sample reading.
type QcMeasurement struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LotID       uint64         `gorm:"column:lot_id;not null;index" json:"lotID"`
	MoveID      uint64         `gorm:"column:move_id;index" json:"moveID"`
	PlanID      uint64         `gorm:"column:plan_id;index" json:"planID"`
	ItemID      uint64         `gorm:"column:item_id;index" json:"itemID"`
	OperationID uint64         `gorm:"column:operation_id;index" json:"operationID"`
	EquipmentID uint64         `gorm:"column:equipment_id;index" json:"equipmentID"`
	ProductID   uint64         `gorm:"column:product_id;index" json:"productID"`
	ParamCode   string         `gorm:"column:param_code;type:varchar(50);not null;index" json:"paramCode"`
	SampleNo    int            `gorm:"column:sample_no" json:"sampleNo"`
	Value       float64        `gorm:"column:value" json:"value"`
	SpecResult  string         `gorm:"column:spec_result;type:varchar(10)" json:"specResult"`
	OperatorID  uint64         `gorm:"column:operator_id" json:"operatorID"`
	MeasuredAt  time.Time      `gorm:"column:measured_at;index" json:"measuredAt"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcMeasurement) TableName() string { return "qc_measurement" }

// QcJudgement is the latest spec result used by the route resolver.
type QcJudgement struct {
	ID          uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LotID       uint64     `gorm:"column:lot_id;not null;index" json:"lotID"`
	OperationID uint64     `gorm:"column:operation_id" json:"operationID"`
	Result      string     `gorm:"column:result;type:varchar(10);not null" json:"result"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *QcJudgement) TableName() string { return "qc_judgement" }

// QcDefectCode is the defect master.
type QcDefectCode struct {
	ID         uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DefectCode string         `gorm:"column:defect_code;type:varchar(50);not null;uniqueIndex" json:"defectCode"`
	DefectName string         `gorm:"column:defect_name;type:varchar(100);not null" json:"defectName"`
	Category   string         `gorm:"column:category;type:varchar(50)" json:"category"`
	Severity   string         `gorm:"column:severity;type:varchar(20)" json:"severity"`
	Status     int            `gorm:"column:status" json:"status"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcDefectCode) TableName() string { return "qc_defect_code" }

// QcDefect is one defect record and its disposition.
type QcDefect struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LotID       uint64         `gorm:"column:lot_id;not null;index" json:"lotID"`
	NodeKey     string         `gorm:"column:node_key;type:varchar(64)" json:"nodeKey"`
	OperationID uint64         `gorm:"column:operation_id;index" json:"operationID"`
	EquipmentID uint64         `gorm:"column:equipment_id;index" json:"equipmentID"`
	MoveID      uint64         `gorm:"column:move_id" json:"moveID"`
	DefectCode  string         `gorm:"column:defect_code;type:varchar(50);not null;index" json:"defectCode"`
	Quantity    int            `gorm:"column:quantity" json:"quantity"`
	Disposition string         `gorm:"column:disposition;type:varchar(20);not null" json:"disposition"`
	Note        string         `gorm:"column:note;type:varchar(255)" json:"note"`
	OperatorID  uint64         `gorm:"column:operator_id" json:"operatorID"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcDefect) TableName() string { return "qc_defect" }

// QcSpcPolicy configures what happens on OOC or OOS for a parameter.
type QcSpcPolicy struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParamCode   string         `gorm:"column:param_code;type:varchar(50);index" json:"paramCode"`
	OperationID uint64         `gorm:"column:operation_id;index" json:"operationID"`
	OnOOC       string         `gorm:"column:on_ooc;type:varchar(30)" json:"onOOC"`
	OnOOS       string         `gorm:"column:on_oos;type:varchar(30)" json:"onOOS"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcSpcPolicy) TableName() string { return "qc_spc_policy" }

// QcSpcLimit is an optional manual control limit. Zero equipment or product means any.
type QcSpcLimit struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParamCode   string         `gorm:"column:param_code;type:varchar(50);index" json:"paramCode"`
	OperationID uint64         `gorm:"column:operation_id;index" json:"operationID"`
	ProductID   uint64         `gorm:"column:product_id" json:"productID"`
	EquipmentID uint64         `gorm:"column:equipment_id" json:"equipmentID"`
	ChartType   string         `gorm:"column:chart_type;type:varchar(10)" json:"chartType"`
	Center      *float64       `gorm:"column:center" json:"center"`
	LCL         *float64       `gorm:"column:lcl" json:"lcl"`
	UCL         *float64       `gorm:"column:ucl" json:"ucl"`
	UseManual   bool           `gorm:"column:use_manual" json:"useManual"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *QcSpcLimit) TableName() string { return "qc_spc_limit" }

// QcSpcEvent is one OOC or OOS violation.
type QcSpcEvent struct {
	ID            uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParamCode     string     `gorm:"column:param_code;type:varchar(50);index" json:"paramCode"`
	OperationID   uint64     `gorm:"column:operation_id;index" json:"operationID"`
	EquipmentID   uint64     `gorm:"column:equipment_id;index" json:"equipmentID"`
	ProductID     uint64     `gorm:"column:product_id" json:"productID"`
	LotID         uint64     `gorm:"column:lot_id;index" json:"lotID"`
	MeasurementID uint64     `gorm:"column:measurement_id;index" json:"measurementID"`
	Rule          int        `gorm:"column:rule_no" json:"rule"`
	Kind          string     `gorm:"column:kind;type:varchar(10)" json:"kind"`
	Value         float64    `gorm:"column:value" json:"value"`
	Reaction      string     `gorm:"column:reaction;type:varchar(30)" json:"reaction"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *QcSpcEvent) TableName() string { return "qc_spc_event" }

const (
	DispositionRework  = "rework"
	DispositionScrap   = "scrap"
	DispositionUseAsIs = "use_as_is"
	DispositionHold    = "hold"

	SpecPass = "pass"
	SpecFail = "fail"

	SpcOOC = "ooc"
	SpcOOS = "oos"

	ReactNone            = "none"
	ReactHold            = "hold_lot"
	ReactEngineering     = "eqp_engineering"
	ReactDown            = "eqp_down"
	ReactHoldEngineering = "hold_and_engineering"
	ReactHoldDown        = "hold_and_down"
)
