package model

import (
	"time"

	"gorm.io/gorm"
)

// EqpCapability says which operation and optional recipe an equipment can run.
// RecipeID 0 means any recipe of that operation. No rows means the equipment group alone is enough.
type EqpCapability struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EquipmentID uint64         `gorm:"column:equipment_id;not null;index" json:"equipmentID"`
	OperationID uint64         `gorm:"column:operation_id;not null;index" json:"operationID"`
	RecipeID    uint64         `gorm:"column:recipe_id" json:"recipeID"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *EqpCapability) TableName() string { return "eqp_capability" }

// EqpStateLog is one SEMI E10-style state interval. The open row has no EndedAt.
type EqpStateLog struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EquipmentID     uint64     `gorm:"column:equipment_id;not null;index" json:"equipmentID"`
	FromState       string     `gorm:"column:from_state;type:varchar(30)" json:"fromState"`
	ToState         string     `gorm:"column:to_state;type:varchar(30);not null" json:"toState"`
	ReasonCode      string     `gorm:"column:reason_code;type:varchar(50)" json:"reasonCode"`
	Reason          string     `gorm:"column:reason;type:varchar(255)" json:"reason"`
	OperatorID      uint64     `gorm:"column:operator_id" json:"operatorID"`
	StartedAt       time.Time  `gorm:"column:started_at;not null" json:"startedAt"`
	EndedAt         *time.Time `gorm:"column:ended_at" json:"endedAt"`
	DurationSeconds int        `gorm:"column:duration_seconds" json:"durationSeconds"`
	CreatedAt       *time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *EqpStateLog) TableName() string { return "eqp_state_log" }

// EqpPmPlan is a time-based and/or lot-count PM plan for one equipment or a whole group.
type EqpPmPlan struct {
	ID              uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EquipmentID     uint64         `gorm:"column:equipment_id;index" json:"equipmentID"`
	EquipmentGroup  string         `gorm:"column:equipment_group;type:varchar(100)" json:"equipmentGroup"`
	PlanName        string         `gorm:"column:plan_name;type:varchar(100);not null" json:"planName"`
	TriggerType     string         `gorm:"column:trigger_type;type:varchar(20);not null" json:"triggerType"`
	IntervalDays    int            `gorm:"column:interval_days" json:"intervalDays"`
	IntervalCount   int            `gorm:"column:interval_count" json:"intervalCount"`
	ChecklistJSON   string         `gorm:"column:checklist_json;type:text" json:"checklistJson"`
	Enabled         bool           `gorm:"column:enabled" json:"enabled"`
	BlockTrackIn    bool           `gorm:"column:block_track_in" json:"blockTrackIn"`
	LastCompletedAt *time.Time     `gorm:"column:last_completed_at" json:"lastCompletedAt"`
	NextDueAt       *time.Time     `gorm:"column:next_due_at" json:"nextDueAt"`
	LotsSince       int            `gorm:"column:lots_since" json:"lotsSince"`
	CreatedAt       *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *EqpPmPlan) TableName() string { return "eqp_pm_plan" }

// EqpPmTask is one generated or in-progress PM execution.
type EqpPmTask struct {
	ID            uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PlanID        uint64         `gorm:"column:plan_id;not null;index" json:"planID"`
	EquipmentID   uint64         `gorm:"column:equipment_id;not null;index" json:"equipmentID"`
	TaskNo        string         `gorm:"column:task_no;type:varchar(40);index" json:"taskNo"`
	Status        string         `gorm:"column:status;type:varchar(20);not null;index" json:"status"`
	DueAt         *time.Time     `gorm:"column:due_at" json:"dueAt"`
	StartedAt     *time.Time     `gorm:"column:started_at" json:"startedAt"`
	FinishedAt    *time.Time     `gorm:"column:finished_at" json:"finishedAt"`
	Result        string         `gorm:"column:result;type:varchar(20)" json:"result"`
	Note          string         `gorm:"column:note;type:varchar(255)" json:"note"`
	ChecklistJSON string         `gorm:"column:checklist_json;type:text" json:"checklistJson"`
	OperatorID    uint64         `gorm:"column:operator_id" json:"operatorID"`
	CreatedAt     *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *EqpPmTask) TableName() string { return "eqp_pm_task" }

const (
	PmTime  = "time"
	PmCount = "count"
	PmBoth  = "both"

	PmDue        = "due"
	PmOverdue    = "overdue"
	PmInProgress = "in_progress"
	PmDone       = "done"
	PmCancelled  = "cancelled"
)
