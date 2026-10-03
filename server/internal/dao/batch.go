package dao

import (
	"errors"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// ErrBatch is an unknown resource or action.
var ErrBatch = errors.New("batch rejected")

// BatchFailure is one id the batch action could not apply.
type BatchFailure struct {
	ID     uint64 `json:"id"`
	Reason string `json:"reason"`
}

// BatchResult lists ids that succeeded and ids that did not.
type BatchResult struct {
	OK     []uint64       `json:"ok"`
	Failed []BatchFailure `json:"failed"`
}

// RunBatch applies delete, enable, disable, hold, or release to the given ids.
func RunBatch(db *gorm.DB, resource, action string, ids []uint64, reasonCode, reason string) (*BatchResult, error) {
	if len(ids) == 0 || len(ids) > 200 {
		return nil, ErrBatch
	}
	switch resource {
	case "baseFactory":
		return batchStatus(db, func() any { return &model.BaseFactory{} }, action, ids, nil)
	case "baseWorkshop":
		return batchStatus(db, func() any { return &model.BaseWorkshop{} }, action, ids, nil)
	case "baseProductionLine":
		return batchStatus(db, func() any { return &model.BaseProductionLine{} }, action, ids, nil)
	case "baseProduct":
		return batchStatus(db, func() any { return &model.BaseProduct{} }, action, ids, nil)
	case "baseProcessRoute":
		return batchStatus(db, func() any { return &model.BaseProcessRoute{} }, action, ids, nil)
	case "baseOperation":
		return batchStatus(db, func() any { return &model.BaseOperation{} }, action, ids, nil)
	case "baseRecipe":
		return batchStatus(db, func() any { return &model.BaseRecipe{} }, action, ids, nil)
	case "sysUser":
		return batchStatus(db, func() any { return &model.SysUser{} }, action, ids, map[uint64]struct{}{1: {}})
	case "sysRole":
		return batchStatus(db, func() any { return &model.SysRole{} }, action, ids, map[uint64]struct{}{1: {}})
	case "sysMenu":
		return batchStatus(db, func() any { return &model.SysMenu{} }, action, ids, nil)
	case "sysDictType":
		return batchStatus(db, func() any { return &model.SysDictType{} }, action, ids, nil)
	case "sysDictItem":
		return batchStatus(db, func() any { return &model.SysDictItem{} }, action, ids, nil)
	case "sysReasonCode":
		return batchStatus(db, func() any { return &model.MesReasonCode{} }, action, ids, nil)
	case "sysNumberRule":
		return batchStatus(db, func() any { return &model.SysNumberRule{} }, action, ids, nil)
	case "qcDefectCode":
		return batchStatus(db, func() any { return &model.QcDefectCode{} }, action, ids, nil)
	case "eqpPmPlan":
		return batchFlag(db, func() any { return &model.EqpPmPlan{} }, action, ids, DeletePmPlan)
	case "qcInspectPlan":
		return batchFlag(db, func() any { return &model.QcInspectPlan{} }, action, ids, DeleteInspectPlan)
	case "wipWorkOrder":
		return batchWorkOrders(db, action, ids)
	case "eqpEquipment":
		return batchEquipment(db, action, ids)
	case "wipLot":
		return batchLots(db, action, ids, reasonCode, reason)
	default:
		return nil, ErrBatch
	}
}

func newBatchResult() *BatchResult {
	return &BatchResult{OK: []uint64{}, Failed: []BatchFailure{}}
}

func batchStatus(db *gorm.DB, newRow func() any, action string, ids []uint64, protect map[uint64]struct{}) (*BatchResult, error) {
	if action != "delete" && action != "enable" && action != "disable" {
		return nil, ErrBatch
	}
	out := newBatchResult()
	for _, id := range ids {
		if _, skip := protect[id]; skip {
			out.Failed = append(out.Failed, BatchFailure{ID: id, Reason: "protected"})
			continue
		}
		row := newRow()
		var res *gorm.DB
		switch action {
		case "delete":
			res = db.Where("id = ?", id).Delete(row)
		case "enable":
			res = db.Model(row).Where("id = ?", id).Update("status", 1)
		default:
			res = db.Model(row).Where("id = ?", id).Update("status", 2)
		}
		if res.Error != nil || res.RowsAffected == 0 {
			out.Failed = append(out.Failed, BatchFailure{ID: id, Reason: "rejected"})
			continue
		}
		out.OK = append(out.OK, id)
	}
	return out, nil
}

func batchFlag(db *gorm.DB, newRow func() any, action string, ids []uint64, remove func(*gorm.DB, uint64) error) (*BatchResult, error) {
	if action != "delete" && action != "enable" && action != "disable" {
		return nil, ErrBatch
	}
	out := newBatchResult()
	for _, id := range ids {
		var err error
		switch action {
		case "delete":
			err = remove(db, id)
		case "enable":
			err = db.Model(newRow()).Where("id = ?", id).Update("enabled", true).Error
		default:
			err = db.Model(newRow()).Where("id = ?", id).Update("enabled", false).Error
		}
		if err != nil {
			out.Failed = append(out.Failed, BatchFailure{ID: id, Reason: "rejected"})
			continue
		}
		out.OK = append(out.OK, id)
	}
	return out, nil
}

func batchWorkOrders(db *gorm.DB, action string, ids []uint64) (*BatchResult, error) {
	if action != "delete" {
		return nil, ErrBatch
	}
	out := newBatchResult()
	for _, id := range ids {
		if err := DeleteWorkOrder(db, id); err != nil {
			out.Failed = append(out.Failed, BatchFailure{ID: id, Reason: "rejected"})
			continue
		}
		out.OK = append(out.OK, id)
	}
	return out, nil
}

func batchEquipment(db *gorm.DB, action string, ids []uint64) (*BatchResult, error) {
	if action != "delete" {
		return nil, ErrBatch
	}
	out := newBatchResult()
	for _, id := range ids {
		if err := DeleteEquipment(db, id); err != nil {
			out.Failed = append(out.Failed, BatchFailure{ID: id, Reason: "rejected"})
			continue
		}
		out.OK = append(out.OK, id)
	}
	return out, nil
}

func batchLots(db *gorm.DB, action string, ids []uint64, reasonCode, reason string) (*BatchResult, error) {
	if action != "hold" && action != "release" {
		return nil, ErrBatch
	}
	out := newBatchResult()
	for _, id := range ids {
		var err error
		if action == "hold" {
			_, err = HoldLot(db, id, reasonCode, reason)
		} else {
			_, err = ReleaseHoldLot(db, id, reasonCode, reason)
		}
		if err != nil {
			out.Failed = append(out.Failed, BatchFailure{ID: id, Reason: "rejected"})
			continue
		}
		out.OK = append(out.OK, id)
	}
	return out, nil
}
