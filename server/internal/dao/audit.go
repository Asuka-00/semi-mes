package dao

import (
	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// ListAudit returns a page of audit rows. Unknown filter and sort fields are ignored.
func ListAudit(db *gorm.DB, page, limit int, sort string, columns []Column) ([]model.SysAuditLog, int64, error) {
	q := db.Model(&model.SysAuditLog{})
	q = applyColumns(q, columns, map[string]string{
		"username":    "username",
		"module":      "module",
		"action":      "action",
		"entity_type": "entity_type",
		"entity_code": "entity_code",
		"entity_id":   "entity_id",
		"user_id":     "user_id",
		"created_at":  "created_at",
	})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.SysAuditLog
	err := q.Order(orderClause(sort, map[string]string{
		"id": "id", "created_at": "created_at", "username": "username", "module": "module", "action": "action",
	}, "id DESC")).Offset(page * limit).Limit(limit).Find(&rows).Error
	if rows == nil {
		rows = []model.SysAuditLog{}
	}
	return rows, total, err
}
