package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RuleWorkOrder = "work_order"
	RuleLot       = "lot"
)

// SysDictType groups configurable code lists such as lot status.
type SysDictType struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TypeCode  string         `gorm:"column:type_code;type:varchar(40);not null;uniqueIndex" json:"typeCode"`
	TypeName  string         `gorm:"column:type_name;type:varchar(80);not null" json:"typeName"`
	Remark    string         `gorm:"column:remark;type:varchar(255)" json:"remark"`
	Status    int            `gorm:"column:status;index" json:"status"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *SysDictType) TableName() string { return "sys_dict_type" }

// SysDictItem is one enabled code inside a dictionary type.
type SysDictItem struct {
	ID        uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TypeID    uint64         `gorm:"column:type_id;not null;index;uniqueIndex:uk_dict_item" json:"typeID"`
	ItemCode  string         `gorm:"column:item_code;type:varchar(40);not null;uniqueIndex:uk_dict_item" json:"itemCode"`
	LabelZh   string         `gorm:"column:label_zh;type:varchar(80)" json:"labelZh"`
	LabelEn   string         `gorm:"column:label_en;type:varchar(80)" json:"labelEn"`
	Color     string         `gorm:"column:color;type:varchar(20)" json:"color"`
	Tag       string         `gorm:"column:tag;type:varchar(20)" json:"tag"`
	SortOrder int            `gorm:"column:sort_order;index" json:"sortOrder"`
	Status    int            `gorm:"column:status;index" json:"status"`
	CreatedAt *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *SysDictItem) TableName() string { return "sys_dict_item" }

// MesReasonCode is a reason selected by hold, release, scrap, rework, or equipment forms.
type MesReasonCode struct {
	ID         uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Category   string         `gorm:"column:category;type:varchar(30);not null;index;uniqueIndex:uk_reason" json:"category"`
	ReasonCode string         `gorm:"column:reason_code;type:varchar(50);not null;uniqueIndex:uk_reason" json:"reasonCode"`
	NameZh     string         `gorm:"column:name_zh;type:varchar(80)" json:"nameZh"`
	NameEn     string         `gorm:"column:name_en;type:varchar(80)" json:"nameEn"`
	SortOrder  int            `gorm:"column:sort_order" json:"sortOrder"`
	Status     int            `gorm:"column:status;index" json:"status"`
	CreatedAt  *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *MesReasonCode) TableName() string { return "mes_reason_code" }

// SysNumberRule configures prefix, date part, width, and reset period.
type SysNumberRule struct {
	ID          uint64         `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RuleCode    string         `gorm:"column:rule_code;type:varchar(40);not null;uniqueIndex" json:"ruleCode"`
	RuleName    string         `gorm:"column:rule_name;type:varchar(80);not null" json:"ruleName"`
	Prefix      string         `gorm:"column:prefix;type:varchar(20)" json:"prefix"`
	DatePart    string         `gorm:"column:date_part;type:varchar(20)" json:"datePart"`
	SeqLength   int            `gorm:"column:seq_length" json:"seqLength"`
	ResetPeriod string         `gorm:"column:reset_period;type:varchar(20)" json:"resetPeriod"`
	Separator   string         `gorm:"column:separator;type:varchar(5)" json:"separator"`
	Status      int            `gorm:"column:status" json:"status"`
	CreatedAt   *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (m *SysNumberRule) TableName() string { return "sys_number_rule" }

// SysNumberSeq stores the last issued sequence for one rule and reset bucket.
type SysNumberSeq struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	RuleCode  string     `gorm:"column:rule_code;type:varchar(40);not null;uniqueIndex:uk_number_seq" json:"ruleCode"`
	PeriodKey string     `gorm:"column:period_key;type:varchar(20);not null;uniqueIndex:uk_number_seq" json:"periodKey"`
	Current   int        `gorm:"column:current" json:"current"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *SysNumberSeq) TableName() string { return "sys_number_seq" }
