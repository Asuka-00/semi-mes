package dao

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// ErrDict is a duplicate code or a missing parent type.
var ErrDict = errors.New("dict rejected")

func ListDictTypes(db *gorm.DB, page, limit int, sort string, columns []Column) ([]model.SysDictType, int64, error) {
	q := applyColumns(db.Model(&model.SysDictType{}), columns, map[string]string{
		"type_code": "type_code", "type_name": "type_name", "status": "status", "created_at": "created_at",
	})
	return pageQuery[model.SysDictType](q, page, limit, orderClause(sort, map[string]string{
		"id": "id", "type_code": "type_code", "sort_order": "id", "created_at": "created_at",
	}, "id DESC"))
}

func SaveDictType(db *gorm.DB, row *model.SysDictType) error {
	row.TypeCode = strings.TrimSpace(row.TypeCode)
	row.TypeName = strings.TrimSpace(row.TypeName)
	if row.TypeCode == "" || row.TypeName == "" {
		return ErrDict
	}
	if row.Status == 0 {
		row.Status = 1
	}
	var existing model.SysDictType
	err := db.Where("type_code = ?", row.TypeCode).First(&existing).Error
	if err == nil && existing.ID != row.ID {
		return ErrDict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if row.ID == 0 {
		return db.Create(row).Error
	}
	return db.Model(&model.SysDictType{}).Where("id = ?", row.ID).Updates(map[string]any{
		"type_code": row.TypeCode, "type_name": row.TypeName, "remark": row.Remark, "status": row.Status,
	}).Error
}

func DeleteDictType(db *gorm.DB, id uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("type_id = ?", id).Delete(&model.SysDictItem{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.SysDictType{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrDict
		}
		return nil
	})
}

func ListDictItems(db *gorm.DB, page, limit int, sort string, columns []Column) ([]model.SysDictItem, int64, error) {
	q := applyColumns(db.Model(&model.SysDictItem{}), columns, map[string]string{
		"type_id": "type_id", "item_code": "item_code", "label_zh": "label_zh", "status": "status", "sort_order": "sort_order",
	})
	return pageQuery[model.SysDictItem](q, page, limit, orderClause(sort, map[string]string{
		"id": "id", "sort_order": "sort_order", "item_code": "item_code",
	}, "sort_order ASC, id ASC"))
}

func EnabledDictItems(db *gorm.DB, typeCode string) ([]model.SysDictItem, error) {
	var typ model.SysDictType
	if err := db.Where("type_code = ? AND status = 1", strings.TrimSpace(typeCode)).First(&typ).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []model.SysDictItem{}, nil
		}
		return nil, err
	}
	var rows []model.SysDictItem
	err := db.Where("type_id = ? AND status = 1", typ.ID).Order("sort_order ASC, id ASC").Find(&rows).Error
	if rows == nil {
		rows = []model.SysDictItem{}
	}
	return rows, err
}

func SaveDictItem(db *gorm.DB, row *model.SysDictItem) error {
	row.ItemCode = strings.TrimSpace(row.ItemCode)
	if row.TypeID == 0 || row.ItemCode == "" {
		return ErrDict
	}
	var typ model.SysDictType
	if err := db.First(&typ, row.TypeID).Error; err != nil {
		return ErrDict
	}
	if row.Status == 0 {
		row.Status = 1
	}
	var existing model.SysDictItem
	err := db.Where("type_id = ? AND item_code = ?", row.TypeID, row.ItemCode).First(&existing).Error
	if err == nil && existing.ID != row.ID {
		return ErrDict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if row.ID == 0 {
		return db.Create(row).Error
	}
	return db.Model(&model.SysDictItem{}).Where("id = ?", row.ID).Updates(map[string]any{
		"type_id": row.TypeID, "item_code": row.ItemCode, "label_zh": row.LabelZh, "label_en": row.LabelEn,
		"color": row.Color, "tag": row.Tag, "sort_order": row.SortOrder, "status": row.Status,
	}).Error
}

func ListReasonCodes(db *gorm.DB, page, limit int, sort string, columns []Column) ([]model.MesReasonCode, int64, error) {
	q := applyColumns(db.Model(&model.MesReasonCode{}), columns, map[string]string{
		"category": "category", "reason_code": "reason_code", "name_zh": "name_zh", "status": "status",
	})
	return pageQuery[model.MesReasonCode](q, page, limit, orderClause(sort, map[string]string{
		"id": "id", "sort_order": "sort_order", "reason_code": "reason_code", "category": "category",
	}, "category ASC, sort_order ASC, id ASC"))
}

func EnabledReasons(db *gorm.DB, category string) ([]model.MesReasonCode, error) {
	var rows []model.MesReasonCode
	err := db.Where("category = ? AND status = 1", strings.TrimSpace(category)).Order("sort_order ASC, id ASC").Find(&rows).Error
	if rows == nil {
		rows = []model.MesReasonCode{}
	}
	return rows, err
}

func SaveReasonCode(db *gorm.DB, row *model.MesReasonCode) error {
	row.Category = strings.TrimSpace(row.Category)
	row.ReasonCode = strings.TrimSpace(row.ReasonCode)
	if row.Category == "" || row.ReasonCode == "" {
		return ErrDict
	}
	if row.Status == 0 {
		row.Status = 1
	}
	var existing model.MesReasonCode
	err := db.Where("category = ? AND reason_code = ?", row.Category, row.ReasonCode).First(&existing).Error
	if err == nil && existing.ID != row.ID {
		return ErrDict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if row.ID == 0 {
		return db.Create(row).Error
	}
	return db.Model(&model.MesReasonCode{}).Where("id = ?", row.ID).Updates(map[string]any{
		"category": row.Category, "reason_code": row.ReasonCode, "name_zh": row.NameZh, "name_en": row.NameEn,
		"sort_order": row.SortOrder, "status": row.Status,
	}).Error
}

// AcceptCode reports whether a reason is allowed. An empty catalog falls back to the hard-coded map.
func AcceptCode(db *gorm.DB, category, code string, fallback map[string]bool) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	var total int64
	if err := db.Model(&model.MesReasonCode{}).Where("category = ? AND status = 1", category).Count(&total).Error; err != nil {
		return fallback[code]
	}
	if total == 0 {
		if fallback == nil {
			return true
		}
		return fallback[code]
	}
	var hit int64
	if err := db.Model(&model.MesReasonCode{}).Where("category = ? AND reason_code = ? AND status = 1", category, code).Count(&hit).Error; err != nil {
		return false
	}
	return hit > 0
}

func ListNumberRules(db *gorm.DB, page, limit int, sort string, columns []Column) ([]model.SysNumberRule, int64, error) {
	q := applyColumns(db.Model(&model.SysNumberRule{}), columns, map[string]string{
		"rule_code": "rule_code", "rule_name": "rule_name", "status": "status",
	})
	return pageQuery[model.SysNumberRule](q, page, limit, orderClause(sort, map[string]string{
		"id": "id", "rule_code": "rule_code",
	}, "id ASC"))
}

func SaveNumberRule(db *gorm.DB, row *model.SysNumberRule) error {
	row.RuleCode = strings.TrimSpace(row.RuleCode)
	row.RuleName = strings.TrimSpace(row.RuleName)
	if row.RuleCode == "" || row.RuleName == "" {
		return ErrDict
	}
	if row.SeqLength < 1 {
		row.SeqLength = 3
	}
	if row.SeqLength > 12 {
		row.SeqLength = 12
	}
	switch strings.ToLower(row.ResetPeriod) {
	case "daily", "monthly", "yearly", "never":
	default:
		row.ResetPeriod = "never"
	}
	switch row.DatePart {
	case "", "none", "yyyy", "yyyyMM", "yyyyMMdd":
	default:
		return ErrDict
	}
	if row.Status == 0 {
		row.Status = 1
	}
	var existing model.SysNumberRule
	err := db.Where("rule_code = ?", row.RuleCode).First(&existing).Error
	if err == nil && existing.ID != row.ID {
		return ErrDict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if row.ID == 0 {
		return db.Create(row).Error
	}
	return db.Model(&model.SysNumberRule{}).Where("id = ?", row.ID).Updates(map[string]any{
		"rule_code": row.RuleCode, "rule_name": row.RuleName, "prefix": row.Prefix, "date_part": row.DatePart,
		"seq_length": row.SeqLength, "reset_period": row.ResetPeriod, "separator": row.Separator, "status": row.Status,
	}).Error
}

func pageQuery[T any](q *gorm.DB, page, limit int, order string) ([]T, int64, error) {
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []T
	err := q.Order(order).Offset(page * limit).Limit(limit).Find(&rows).Error
	if rows == nil {
		rows = []T{}
	}
	return rows, total, err
}
