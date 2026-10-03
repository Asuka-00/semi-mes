package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
	"semi-mes/server/internal/routegraph"
)

var (
	ErrEqpState = errors.New("equipment state")
	ErrEqpPM    = errors.New("equipment pm")
)

// EqpReasonCodes are the reason codes accepted on a state change.
var EqpReasonCodes = map[string]bool{
	"PROD_START": true, "PROD_END": true, "ENG_SETUP": true, "ENG_DONE": true,
	"PM_START": true, "PM_DONE": true, "BREAKDOWN": true, "REPAIR_DONE": true,
	"NO_WIP": true, "SHIFT_END": true, "SHIFT_START": true, "OTHER": true,
}

var eqpTransitions = map[string]map[string]bool{
	model.EqpStandby:         {model.EqpEngineering: true, model.EqpScheduledDown: true, model.EqpUnscheduledDown: true, model.EqpNonScheduled: true, model.EqpProductive: true},
	model.EqpProductive:      {model.EqpStandby: true, model.EqpEngineering: true, model.EqpUnscheduledDown: true, model.EqpScheduledDown: true},
	model.EqpEngineering:     {model.EqpStandby: true, model.EqpScheduledDown: true, model.EqpUnscheduledDown: true, model.EqpNonScheduled: true, model.EqpProductive: true},
	model.EqpScheduledDown:   {model.EqpStandby: true, model.EqpEngineering: true, model.EqpUnscheduledDown: true},
	model.EqpUnscheduledDown: {model.EqpStandby: true, model.EqpEngineering: true, model.EqpScheduledDown: true},
	model.EqpNonScheduled:    {model.EqpStandby: true, model.EqpEngineering: true, model.EqpScheduledDown: true},
}

// CapabilityInput is one operation/recipe the equipment can run.
type CapabilityInput struct {
	OperationID uint64
	RecipeID    uint64
}

// EquipmentInput creates or updates the master row and replaces capabilities.
type EquipmentInput struct {
	EquipmentCode  string
	EquipmentName  string
	EquipmentGroup string
	EquipmentType  string
	Status         string
	LineID         uint64
	ModelName      string
	Manufacturer   string
	SerialNo       string
	Location       string
	ChamberCount   int
	Capacity       int
	InstallDate    string
	Capabilities   []CapabilityInput
}

// CheckItem is one PM checklist line.
type CheckItem struct {
	Name   string `json:"name"`
	Result string `json:"result"`
	Note   string `json:"note"`
}

// EquipmentDetail is the equipment page payload.
type EquipmentDetail struct {
	Equipment    model.EqpEquipment  `json:"equipment"`
	Capabilities []CapabilityView    `json:"capabilities"`
	StateLogs    []model.EqpStateLog `json:"stateLogs"`
	PmTasks      []PmTaskView        `json:"pmTasks"`
	RecentLots   []RecentLot         `json:"recentLots"`
	OpenLots     int                 `json:"openLots"`
}

// CapabilityView joins the operation code for the ledger.
type CapabilityView struct {
	model.EqpCapability
	OperationCode string `json:"operationCode"`
	OperationName string `json:"operationName"`
	RecipeCode    string `json:"recipeCode"`
}

// RecentLot is a move that used this equipment.
type RecentLot struct {
	LotNo      string     `json:"lotNo"`
	NodeKey    string     `json:"nodeKey"`
	State      string     `json:"state"`
	QtyIn      int        `json:"qtyIn"`
	QtyOut     int        `json:"qtyOut"`
	TrackInAt  *time.Time `json:"trackInAt"`
	TrackOutAt *time.Time `json:"trackOutAt"`
}

// UtilRow is one equipment line on the status board.
type UtilRow struct {
	ID             uint64  `json:"id"`
	EquipmentCode  string  `json:"equipmentCode"`
	EquipmentName  string  `json:"equipmentName"`
	EquipmentGroup string  `json:"equipmentGroup"`
	Status         string  `json:"status"`
	OpenLots       int     `json:"openLots"`
	Capacity       int     `json:"capacity"`
	Utilization    float64 `json:"utilization"`
}

// EquipmentBoard summarizes states and recent utilization.
type EquipmentBoard struct {
	ByState []struct {
		Key   string `json:"key"`
		Count int    `json:"count"`
	} `json:"byState"`
	Tools   []UtilRow `json:"tools"`
	Overdue int       `json:"overdue"`
}

// PmPlanInput edits a PM plan.
type PmPlanInput struct {
	EquipmentID    uint64
	EquipmentGroup string
	PlanName       string
	TriggerType    string
	IntervalDays   int
	IntervalCount  int
	Checklist      []string
	Enabled        bool
	BlockTrackIn   bool
	NextDueAt      string
}

// PmTaskView is a task plus the names operators need.
type PmTaskView struct {
	model.EqpPmTask
	PlanName       string `json:"planName"`
	EquipmentCode  string `json:"equipmentCode"`
	EquipmentName  string `json:"equipmentName"`
	EquipmentGroup string `json:"equipmentGroup"`
}

func NormalizeEqpState(state string) string {
	switch state {
	case model.EqpIdle, "":
		return model.EqpStandby
	case model.EqpDown:
		return model.EqpUnscheduledDown
	default:
		return state
	}
}

func eqpCapacity(eqp *model.EqpEquipment) int {
	if eqp.Capacity > 0 {
		return eqp.Capacity
	}
	if eqp.ChamberCount > 0 {
		return eqp.ChamberCount
	}
	return 1
}

// CreateEquipment inserts the master, capabilities, and the opening state log.
func CreateEquipment(db *gorm.DB, in EquipmentInput) (*model.EqpEquipment, error) {
	row, err := equipmentFromInput(in, true)
	if err != nil {
		return nil, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.EqpEquipment{}).Where("equipment_code = ?", row.EquipmentCode).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrEqpState
		}
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		if err := replaceCapabilities(tx, row.ID, in.Capabilities); err != nil {
			return err
		}
		return applyState(tx, row, row.Status, "SHIFT_START", "", 0, true)
	})
	return row, err
}

// UpdateEquipment replaces master fields and capabilities. Status changes go through ChangeEquipmentState.
func UpdateEquipment(db *gorm.DB, id uint64, in EquipmentInput) (*model.EqpEquipment, error) {
	row, err := loadEquipment(db, id)
	if err != nil {
		return nil, err
	}
	next, err := equipmentFromInput(in, false)
	if err != nil {
		return nil, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.EqpEquipment{}).Where("equipment_code = ? AND id <> ?", next.EquipmentCode, id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrEqpState
		}
		if err := tx.Model(row).Updates(map[string]any{
			"equipment_code": next.EquipmentCode, "equipment_name": next.EquipmentName, "equipment_group": next.EquipmentGroup,
			"equipment_type": next.EquipmentType, "line_id": next.LineID, "model_name": next.ModelName, "manufacturer": next.Manufacturer,
			"serial_no": next.SerialNo, "location": next.Location, "chamber_count": next.ChamberCount, "capacity": next.Capacity,
			"install_date": next.InstallDate,
		}).Error; err != nil {
			return err
		}
		return replaceCapabilities(tx, id, in.Capabilities)
	})
	if err != nil {
		return nil, err
	}
	return loadEquipment(db, id)
}

// DeleteEquipment soft-deletes a tool that is not running lots.
func DeleteEquipment(db *gorm.DB, id uint64) error {
	eqp, err := loadEquipment(db, id)
	if err != nil {
		return err
	}
	open, err := openLotCount(db, id)
	if err != nil {
		return err
	}
	if open > 0 || NormalizeEqpState(eqp.Status) == model.EqpProductive {
		return ErrEqpState
	}
	return db.Delete(eqp).Error
}

// ListEquipmentPage returns a filtered page of the ledger.
func ListEquipmentPage(db *gorm.DB, page, limit int, sort string, columns []Column) ([]model.EqpEquipment, int64, error) {
	allow := map[string]string{
		"equipment_code": "equipment_code", "equipment_group": "equipment_group", "status": "status", "equipment_name": "equipment_name",
		"id": "id", "created_at": "created_at",
	}
	q := applyColumns(db.Model(&model.EqpEquipment{}), columns, allow)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.EqpEquipment
	err := q.Order(orderClause(sort, allow, "equipment_code")).Offset(page * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// GetEquipmentDetail loads state history, PM, and recent lots.
func GetEquipmentDetail(db *gorm.DB, id uint64) (*EquipmentDetail, error) {
	eqp, err := loadEquipment(db, id)
	if err != nil {
		return nil, err
	}
	detail := &EquipmentDetail{Equipment: *eqp, Capabilities: []CapabilityView{}, StateLogs: []model.EqpStateLog{}, PmTasks: []PmTaskView{}, RecentLots: []RecentLot{}}
	if err := db.Table("eqp_capability c").
		Select("c.*, o.operation_code, o.operation_name, r.recipe_code").
		Joins("LEFT JOIN base_operation o ON o.id = c.operation_id").
		Joins("LEFT JOIN base_recipe r ON r.id = c.recipe_id AND r.deleted_at IS NULL").
		Where("c.equipment_id = ? AND c.deleted_at IS NULL", id).
		Order("c.id").Scan(&detail.Capabilities).Error; err != nil {
		return nil, err
	}
	if err := db.Where("equipment_id = ?", id).Order("id desc").Limit(30).Find(&detail.StateLogs).Error; err != nil {
		return nil, err
	}
	tasks, _, err := ListPmTasks(db, 0, 20, []Column{{Name: "equipment_id", Exp: "=", Value: fmt.Sprintf("%d", id)}})
	if err != nil {
		return nil, err
	}
	detail.PmTasks = tasks
	if err := db.Table("wip_move m").
		Select("l.lot_no, m.node_key, m.state, m.qty_in, m.qty_out, m.track_in_at, m.track_out_at").
		Joins("JOIN wip_lot l ON l.id = m.lot_id").
		Where("m.equipment_id = ?", id).Order("m.id desc").Limit(10).Scan(&detail.RecentLots).Error; err != nil {
		return nil, err
	}
	detail.OpenLots, err = openLotCount(db, id)
	return detail, err
}

// ChangeEquipmentState applies an operator transition. Productive is reserved for Track In.
func ChangeEquipmentState(db *gorm.DB, id uint64, to, reasonCode, reason string, operatorID uint64) (*model.EqpEquipment, error) {
	var updated *model.EqpEquipment
	err := db.Transaction(func(tx *gorm.DB) error {
		eqp, err := loadEquipment(tx, id)
		if err != nil {
			return err
		}
		if err := applyState(tx, eqp, to, reasonCode, reason, operatorID, false); err != nil {
			return err
		}
		updated = eqp
		return nil
	})
	return updated, err
}

// LoadEquipmentBoard returns counts, overdue PM, and 24h utilization.
func LoadEquipmentBoard(db *gorm.DB) (*EquipmentBoard, error) {
	var tools []model.EqpEquipment
	if err := db.Order("equipment_code").Find(&tools).Error; err != nil {
		return nil, err
	}
	board := &EquipmentBoard{Tools: []UtilRow{}}
	counts := map[string]int{}
	now := time.Now()
	for _, tool := range tools {
		state := NormalizeEqpState(tool.Status)
		counts[state]++
		open, err := openLotCount(db, tool.ID)
		if err != nil {
			return nil, err
		}
		util, err := utilization(db, tool.ID, now)
		if err != nil {
			return nil, err
		}
		board.Tools = append(board.Tools, UtilRow{
			ID: tool.ID, EquipmentCode: tool.EquipmentCode, EquipmentName: tool.EquipmentName,
			EquipmentGroup: tool.EquipmentGroup, Status: state, OpenLots: open, Capacity: eqpCapacity(&tool), Utilization: util,
		})
	}
	order := []string{model.EqpProductive, model.EqpStandby, model.EqpEngineering, model.EqpScheduledDown, model.EqpUnscheduledDown, model.EqpNonScheduled}
	for _, key := range order {
		if counts[key] == 0 {
			continue
		}
		board.ByState = append(board.ByState, struct {
			Key   string `json:"key"`
			Count int    `json:"count"`
		}{Key: key, Count: counts[key]})
	}
	if board.ByState == nil {
		board.ByState = []struct {
			Key   string `json:"key"`
			Count int    `json:"count"`
		}{}
	}
	var overdue int64
	if err := db.Model(&model.EqpPmTask{}).Where("status = ?", model.PmOverdue).Count(&overdue).Error; err != nil {
		return nil, err
	}
	board.Overdue = int(overdue)
	return board, nil
}

func utilization(db *gorm.DB, equipmentID uint64, now time.Time) (float64, error) {
	window := now.Add(-24 * time.Hour)
	var logs []model.EqpStateLog
	if err := db.Where("equipment_id = ? AND to_state = ?", equipmentID, model.EqpProductive).Find(&logs).Error; err != nil {
		return 0, err
	}
	seconds := 0
	for _, log := range logs {
		start := log.StartedAt
		end := now
		if log.EndedAt != nil {
			end = *log.EndedAt
		}
		if end.Before(window) || !start.Before(end) {
			continue
		}
		if start.Before(window) {
			start = window
		}
		if end.After(now) {
			end = now
		}
		seconds += int(end.Sub(start).Seconds())
	}
	util := float64(seconds) / 86400
	if util > 1 {
		util = 1
	}
	if util < 0 {
		util = 0
	}
	return util, nil
}

// SavePmPlan creates or updates a plan.
func SavePmPlan(db *gorm.DB, id uint64, in PmPlanInput) (*model.EqpPmPlan, error) {
	if strings.TrimSpace(in.PlanName) == "" {
		return nil, ErrEqpPM
	}
	if in.TriggerType != model.PmTime && in.TriggerType != model.PmCount && in.TriggerType != model.PmBoth {
		return nil, ErrEqpPM
	}
	if in.EquipmentID == 0 && strings.TrimSpace(in.EquipmentGroup) == "" {
		return nil, ErrEqpPM
	}
	raw, _ := json.Marshal(in.Checklist)
	var due *time.Time
	if strings.TrimSpace(in.NextDueAt) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(in.NextDueAt))
		if err != nil {
			return nil, ErrEqpPM
		}
		due = &parsed
	}
	if id == 0 {
		row := &model.EqpPmPlan{
			EquipmentID: in.EquipmentID, EquipmentGroup: in.EquipmentGroup, PlanName: strings.TrimSpace(in.PlanName),
			TriggerType: in.TriggerType, IntervalDays: in.IntervalDays, IntervalCount: in.IntervalCount,
			ChecklistJSON: string(raw), Enabled: in.Enabled, BlockTrackIn: in.BlockTrackIn, NextDueAt: due,
		}
		if err := db.Create(row).Error; err != nil {
			return nil, err
		}
		return row, nil
	}
	var row model.EqpPmPlan
	if err := db.First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipNotFound
		}
		return nil, err
	}
	if err := db.Model(&row).Updates(map[string]any{
		"equipment_id": in.EquipmentID, "equipment_group": in.EquipmentGroup, "plan_name": strings.TrimSpace(in.PlanName),
		"trigger_type": in.TriggerType, "interval_days": in.IntervalDays, "interval_count": in.IntervalCount,
		"checklist_json": string(raw), "enabled": in.Enabled, "block_track_in": in.BlockTrackIn, "next_due_at": due,
	}).Error; err != nil {
		return nil, err
	}
	if err := db.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// ListPmPlans returns a filtered page of plans. limit <= 0 returns up to 500 rows.
func ListPmPlans(db *gorm.DB, page, limit int, columns []Column) ([]model.EqpPmPlan, int64, error) {
	rest, enabled, hasEnabled := pullColumn(columns, "enabled")
	q := applyColumns(db.Model(&model.EqpPmPlan{}), rest, map[string]string{
		"plan_name": "plan_name", "equipment_id": "equipment_id", "equipment_group": "equipment_group",
		"trigger_type": "trigger_type", "created_at": "created_at",
	})
	if hasEnabled {
		q = q.Where("enabled = ?", enabled == "1" || strings.EqualFold(enabled, "true"))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 || limit > 2000 {
		limit = 500
	}
	var rows []model.EqpPmPlan
	err := q.Order("id desc").Offset(page * limit).Limit(limit).Find(&rows).Error
	if rows == nil {
		rows = []model.EqpPmPlan{}
	}
	return rows, total, err
}

// DeletePmPlan soft-deletes a plan that has no in-progress task.
func DeletePmPlan(db *gorm.DB, id uint64) error {
	var open int64
	if err := db.Model(&model.EqpPmTask{}).Where("plan_id = ? AND status = ?", id, model.PmInProgress).Count(&open).Error; err != nil {
		return err
	}
	if open > 0 {
		return ErrEqpPM
	}
	return db.Delete(&model.EqpPmPlan{}, id).Error
}

// ListPmTasks returns tasks. It refreshes due tasks first. Filters use the shared whitelist.
func ListPmTasks(db *gorm.DB, page, limit int, columns []Column) ([]PmTaskView, int64, error) {
	if err := EnsurePmTasks(db); err != nil {
		return nil, 0, err
	}
	q := db.Table("eqp_pm_task t").
		Select("t.*, p.plan_name, e.equipment_code, e.equipment_name, e.equipment_group").
		Joins("JOIN eqp_pm_plan p ON p.id = t.plan_id AND p.deleted_at IS NULL").
		Joins("JOIN eqp_equipment e ON e.id = t.equipment_id AND e.deleted_at IS NULL").
		Where("t.deleted_at IS NULL")
	q = applyColumns(q, columns, map[string]string{
		"status": "t.status", "equipment_id": "t.equipment_id", "task_no": "t.task_no",
		"equipment_code": "e.equipment_code", "plan_name": "p.plan_name", "due_at": "t.due_at", "created_at": "t.created_at",
	})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 2000 {
		limit = 2000
	}
	var rows []PmTaskView
	err := q.Order("t.id desc").Offset(page * limit).Limit(limit).Scan(&rows).Error
	if rows == nil {
		rows = []PmTaskView{}
	}
	return rows, total, err
}

// EnsurePmTasks opens a due or overdue task when a plan's time or lot count is reached.
func EnsurePmTasks(db *gorm.DB) error {
	var plans []model.EqpPmPlan
	if err := db.Where("enabled = ?", true).Find(&plans).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, plan := range plans {
		targets, err := planTargets(db, &plan)
		if err != nil {
			return err
		}
		for _, eqp := range targets {
			if !planDue(&plan, now) {
				continue
			}
			var open int64
			if err := db.Model(&model.EqpPmTask{}).
				Where("plan_id = ? AND equipment_id = ? AND status IN ?", plan.ID, eqp.ID, []string{model.PmDue, model.PmOverdue, model.PmInProgress}).
				Count(&open).Error; err != nil {
				return err
			}
			if open > 0 {
				if err := markOverdue(db, plan.ID, eqp.ID, now); err != nil {
					return err
				}
				continue
			}
			dueAt := now
			if plan.NextDueAt != nil {
				dueAt = *plan.NextDueAt
			}
			status := model.PmDue
			if dueAt.Before(now) {
				status = model.PmOverdue
			}
			task := model.EqpPmTask{
				PlanID: plan.ID, EquipmentID: eqp.ID, TaskNo: fmt.Sprintf("PM-%d-%d", plan.ID, eqp.ID),
				Status: status, DueAt: &dueAt, ChecklistJSON: plan.ChecklistJSON,
			}
			if err := db.Create(&task).Error; err != nil {
				return err
			}
			kind := model.NoticePmDue
			if status == model.PmOverdue {
				kind = model.NoticePmOverdue
			}
			if err := notifyUsers(db, kind, map[string]string{
				"eqpCode": eqp.EquipmentCode, "planName": plan.PlanName,
			}, "pm_task", task.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func planDue(plan *model.EqpPmPlan, now time.Time) bool {
	timeDue := plan.NextDueAt != nil && !plan.NextDueAt.After(now)
	countDue := plan.IntervalCount > 0 && plan.LotsSince >= plan.IntervalCount
	switch plan.TriggerType {
	case model.PmTime:
		return timeDue
	case model.PmCount:
		return countDue
	default:
		return timeDue || countDue
	}
}

func markOverdue(db *gorm.DB, planID, equipmentID uint64, now time.Time) error {
	res := db.Model(&model.EqpPmTask{}).
		Where("plan_id = ? AND equipment_id = ? AND status = ? AND due_at < ?", planID, equipmentID, model.PmDue, now).
		Update("status", model.PmOverdue)
	if res.Error != nil || res.RowsAffected == 0 {
		return res.Error
	}
	var eqp model.EqpEquipment
	if err := db.First(&eqp, equipmentID).Error; err != nil {
		return err
	}
	var plan model.EqpPmPlan
	_ = db.First(&plan, planID).Error
	return notifyUsers(db, model.NoticePmOverdue, map[string]string{
		"eqpCode": eqp.EquipmentCode, "planName": plan.PlanName,
	}, "equipment", eqp.ID)
}

func planTargets(db *gorm.DB, plan *model.EqpPmPlan) ([]model.EqpEquipment, error) {
	q := db.Model(&model.EqpEquipment{})
	if plan.EquipmentID > 0 {
		q = q.Where("id = ?", plan.EquipmentID)
	} else {
		q = q.Where("equipment_group = ?", plan.EquipmentGroup)
	}
	var rows []model.EqpEquipment
	err := q.Find(&rows).Error
	return rows, err
}

// StartPm moves the equipment to scheduled down and opens the task.
func StartPm(db *gorm.DB, taskID, operatorID uint64) (*model.EqpPmTask, error) {
	var task *model.EqpPmTask
	err := db.Transaction(func(tx *gorm.DB) error {
		row, err := loadTask(tx, taskID)
		if err != nil {
			return err
		}
		if row.Status != model.PmDue && row.Status != model.PmOverdue {
			return ErrEqpPM
		}
		open, err := openLotCount(tx, row.EquipmentID)
		if err != nil {
			return err
		}
		if open > 0 {
			return ErrEqpPM
		}
		eqp, err := loadEquipment(tx, row.EquipmentID)
		if err != nil {
			return err
		}
		state := NormalizeEqpState(eqp.Status)
		if state == model.EqpStandby || state == model.EqpEngineering {
			if err := tx.Model(eqp).Update("resume_state", state).Error; err != nil {
				return err
			}
			eqp.ResumeState = state
		}
		if state != model.EqpScheduledDown {
			if err := applyState(tx, eqp, model.EqpScheduledDown, "PM_START", "", operatorID, true); err != nil {
				return err
			}
		}
		now := time.Now()
		if err := tx.Model(row).Updates(map[string]any{"status": model.PmInProgress, "started_at": now, "operator_id": operatorID}).Error; err != nil {
			return err
		}
		row.Status = model.PmInProgress
		row.StartedAt = &now
		task = row
		return nil
	})
	return task, err
}

// CompletePm records the checklist and returns the equipment to standby when no other PM is running.
func CompletePm(db *gorm.DB, taskID, operatorID uint64, items []CheckItem, result, note string) (*model.EqpPmTask, error) {
	if result != "pass" && result != "fail" {
		return nil, ErrEqpPM
	}
	var task *model.EqpPmTask
	err := db.Transaction(func(tx *gorm.DB) error {
		row, err := loadTask(tx, taskID)
		if err != nil {
			return err
		}
		if row.Status != model.PmInProgress {
			return ErrEqpPM
		}
		var plan model.EqpPmPlan
		if err := tx.First(&plan, row.PlanID).Error; err != nil {
			return err
		}
		if err := checklistComplete(plan.ChecklistJSON, items); err != nil {
			return err
		}
		raw, _ := json.Marshal(items)
		now := time.Now()
		if err := tx.Model(row).Updates(map[string]any{
			"status": model.PmDone, "finished_at": now, "result": result, "note": note,
			"checklist_json": string(raw), "operator_id": operatorID,
		}).Error; err != nil {
			return err
		}
		next := map[string]any{"lots_since": 0, "last_completed_at": now}
		if plan.IntervalDays > 0 {
			due := now.Add(time.Duration(plan.IntervalDays) * 24 * time.Hour)
			next["next_due_at"] = due
		}
		if err := tx.Model(&plan).Updates(next).Error; err != nil {
			return err
		}
		var still int64
		if err := tx.Model(&model.EqpPmTask{}).Where("equipment_id = ? AND status = ? AND id <> ?", row.EquipmentID, model.PmInProgress, row.ID).Count(&still).Error; err != nil {
			return err
		}
		eqp, err := loadEquipment(tx, row.EquipmentID)
		if err != nil {
			return err
		}
		if still == 0 && NormalizeEqpState(eqp.Status) == model.EqpScheduledDown {
			back := eqp.ResumeState
			if back != model.EqpStandby && back != model.EqpEngineering {
				back = model.EqpStandby
			}
			if err := applyState(tx, eqp, back, "PM_DONE", result, operatorID, true); err != nil {
				return err
			}
		}
		row.Status = model.PmDone
		row.Result = result
		task = row
		return nil
	})
	return task, err
}

func checklistComplete(planJSON string, items []CheckItem) error {
	var expected []CheckItem
	if strings.TrimSpace(planJSON) != "" && planJSON != "null" {
		if err := json.Unmarshal([]byte(planJSON), &expected); err != nil {
			var names []string
			if err2 := json.Unmarshal([]byte(planJSON), &names); err2 != nil {
				return ErrEqpPM
			}
			for _, name := range names {
				expected = append(expected, CheckItem{Name: name})
			}
		}
	}
	if len(expected) == 0 {
		return nil
	}
	got := map[string]string{}
	for _, item := range items {
		got[item.Name] = item.Result
	}
	for _, item := range expected {
		name := item.Name
		if name == "" {
			continue
		}
		result := got[name]
		if result != "pass" && result != "fail" {
			return ErrEqpPM
		}
	}
	return nil
}

func loadTask(db *gorm.DB, id uint64) (*model.EqpPmTask, error) {
	var row model.EqpPmTask
	if err := db.First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipNotFound
		}
		return nil, err
	}
	return &row, nil
}

func canTrackIn(db *gorm.DB, eqp *model.EqpEquipment, group string, operationID, recipeID uint64) error {
	state := NormalizeEqpState(eqp.Status)
	if state != model.EqpStandby && state != model.EqpEngineering && state != model.EqpProductive {
		return ErrWipEquipment
	}
	if group != "" && eqp.EquipmentGroup != group {
		return ErrWipEquipment
	}
	open, err := openLotCount(db, eqp.ID)
	if err != nil {
		return err
	}
	if open >= eqpCapacity(eqp) {
		return ErrWipEquipment
	}
	if err := capabilityAllows(db, eqp.ID, operationID, recipeID); err != nil {
		return err
	}
	return overdueBlocks(db, eqp)
}

func capabilityAllows(db *gorm.DB, equipmentID, operationID, recipeID uint64) error {
	var rows []model.EqpCapability
	if err := db.Where("equipment_id = ?", equipmentID).Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	for _, row := range rows {
		if row.OperationID != operationID {
			continue
		}
		if row.RecipeID == 0 || row.RecipeID == recipeID {
			return nil
		}
	}
	return ErrWipEquipment
}

func overdueBlocks(db *gorm.DB, eqp *model.EqpEquipment) error {
	var count int64
	err := db.Table("eqp_pm_task t").
		Joins("JOIN eqp_pm_plan p ON p.id = t.plan_id AND p.deleted_at IS NULL").
		Where("t.deleted_at IS NULL AND t.status IN ? AND t.due_at < ? AND p.block_track_in = ?", []string{model.PmDue, model.PmOverdue}, time.Now(), true).
		Where("t.equipment_id = ? AND (p.equipment_id = ? OR (p.equipment_id = 0 AND p.equipment_group = ?))", eqp.ID, eqp.ID, eqp.EquipmentGroup).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrWipEquipment
	}
	return nil
}

func occupyEquipment(tx *gorm.DB, eqp *model.EqpEquipment, operatorID uint64) error {
	state := NormalizeEqpState(eqp.Status)
	if state == model.EqpProductive {
		return nil
	}
	if err := tx.Model(eqp).Update("resume_state", state).Error; err != nil {
		return err
	}
	eqp.ResumeState = state
	return applyState(tx, eqp, model.EqpProductive, "PROD_START", "", operatorID, true)
}

func releaseEquipment(tx *gorm.DB, equipmentID, operatorID uint64) error {
	if equipmentID == 0 {
		return nil
	}
	open, err := openLotCount(tx, equipmentID)
	if err != nil || open > 0 {
		return err
	}
	eqp, err := loadEquipment(tx, equipmentID)
	if err != nil {
		return err
	}
	if NormalizeEqpState(eqp.Status) != model.EqpProductive {
		return nil
	}
	to := eqp.ResumeState
	if to != model.EqpStandby && to != model.EqpEngineering {
		to = model.EqpStandby
	}
	return applyState(tx, eqp, to, "PROD_END", "", operatorID, true)
}

func bumpPmLots(tx *gorm.DB, equipmentID uint64) error {
	eqp, err := loadEquipment(tx, equipmentID)
	if err != nil {
		return err
	}
	var plans []model.EqpPmPlan
	if err := tx.Where("enabled = ? AND (equipment_id = ? OR (equipment_id = 0 AND equipment_group = ?))", true, eqp.ID, eqp.EquipmentGroup).Find(&plans).Error; err != nil {
		return err
	}
	for _, plan := range plans {
		if plan.TriggerType == model.PmTime {
			continue
		}
		if err := tx.Model(&plan).Update("lots_since", gorm.Expr("lots_since + ?", 1)).Error; err != nil {
			return err
		}
	}
	return nil
}

func applyState(tx *gorm.DB, eqp *model.EqpEquipment, to, reasonCode, reason string, operatorID uint64, system bool) error {
	to = NormalizeEqpState(to)
	from := NormalizeEqpState(eqp.Status)
	if from == to && eqp.ID > 0 {
		// Opening log on create uses the same state the row was inserted with.
		var existing int64
		if err := tx.Model(&model.EqpStateLog{}).Where("equipment_id = ?", eqp.ID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrEqpState
		}
	}
	if !system && to == model.EqpProductive {
		return ErrEqpState
	}
	if from != to {
		if !eqpTransitions[from][to] {
			return ErrEqpState
		}
	}
	if !system && !AcceptCode(tx, "eqp", reasonCode, EqpReasonCodes) {
		return ErrEqpState
	}
	if system && reasonCode == "" {
		return ErrEqpState
	}
	if !system && from == model.EqpProductive && to != model.EqpUnscheduledDown {
		open, err := openLotCount(tx, eqp.ID)
		if err != nil {
			return err
		}
		if open > 0 {
			return ErrEqpState
		}
	}
	now := time.Now()
	var openLog model.EqpStateLog
	err := tx.Where("equipment_id = ? AND ended_at IS NULL", eqp.ID).Order("id desc").First(&openLog).Error
	if err == nil {
		seconds := int(now.Sub(openLog.StartedAt).Seconds())
		if seconds < 0 {
			seconds = 0
		}
		if err := tx.Model(&openLog).Updates(map[string]any{"ended_at": now, "duration_seconds": seconds}).Error; err != nil {
			return err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	log := model.EqpStateLog{
		EquipmentID: eqp.ID, FromState: eqp.Status, ToState: to, ReasonCode: reasonCode, Reason: reason,
		OperatorID: operatorID, StartedAt: now,
	}
	if err := tx.Create(&log).Error; err != nil {
		return err
	}
	if err := tx.Model(eqp).Update("status", to).Error; err != nil {
		return err
	}
	eqp.Status = to
	if to == model.EqpUnscheduledDown && from != to {
		return notifyUsers(tx, model.NoticeEqpDown, map[string]string{
			"eqpCode": eqp.EquipmentCode, "reason": reason, "reasonCode": reasonCode,
		}, "equipment", eqp.ID)
	}
	return nil
}

func openLotCount(db *gorm.DB, equipmentID uint64) (int, error) {
	var count int64
	err := db.Model(&model.WipMove{}).Where("equipment_id = ? AND state = ?", equipmentID, model.MoveOpen).Count(&count).Error
	return int(count), err
}

func replaceCapabilities(tx *gorm.DB, equipmentID uint64, caps []CapabilityInput) error {
	if err := tx.Where("equipment_id = ?", equipmentID).Delete(&model.EqpCapability{}).Error; err != nil {
		return err
	}
	for _, cap := range caps {
		if cap.OperationID == 0 {
			continue
		}
		row := model.EqpCapability{EquipmentID: equipmentID, OperationID: cap.OperationID, RecipeID: cap.RecipeID}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func equipmentFromInput(in EquipmentInput, withStatus bool) (*model.EqpEquipment, error) {
	code := strings.TrimSpace(in.EquipmentCode)
	name := strings.TrimSpace(in.EquipmentName)
	if code == "" || name == "" || len(code) > 50 {
		return nil, ErrEqpState
	}
	status := NormalizeEqpState(in.Status)
	if status == model.EqpProductive {
		return nil, ErrEqpState
	}
	if _, ok := eqpTransitions[status]; !ok && status != model.EqpStandby {
		return nil, ErrEqpState
	}
	if !withStatus {
		status = ""
	}
	capacity := in.Capacity
	if capacity <= 0 {
		capacity = 1
	}
	chambers := in.ChamberCount
	if chambers <= 0 {
		chambers = capacity
	}
	var install *time.Time
	if strings.TrimSpace(in.InstallDate) != "" {
		parsed, err := time.Parse("2006-01-02", strings.TrimSpace(in.InstallDate))
		if err != nil {
			return nil, ErrEqpState
		}
		install = &parsed
	}
	row := &model.EqpEquipment{
		EquipmentCode: code, EquipmentName: name, EquipmentGroup: strings.TrimSpace(in.EquipmentGroup),
		EquipmentType: strings.TrimSpace(in.EquipmentType), Status: status, LineID: in.LineID,
		ModelName: in.ModelName, Manufacturer: in.Manufacturer, SerialNo: in.SerialNo, Location: in.Location,
		ChamberCount: chambers, Capacity: capacity, InstallDate: install, ResumeState: model.EqpStandby,
	}
	if row.Status == "" {
		row.Status = model.EqpStandby
	}
	return row, nil
}

// TrackInGate is used by Track In before the lot is checked in.
func TrackInGate(db *gorm.DB, eqp *model.EqpEquipment, node routegraph.Node, recipeID uint64) error {
	return canTrackIn(db, eqp, node.EquipmentGroup, node.OperationID, recipeID)
}
