package dao

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// ErrDuplicate means an active row already uses the same username or role code.
// Soft-deleted rows are ignored, so the key can be used again.
var ErrDuplicate = errors.New("duplicate active key")

func rejectActiveUser(db *gorm.DB, username string, except uint64) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil
	}
	var n int64
	q := db.Model(&model.SysUser{}).Where("username = ?", username)
	if except > 0 {
		q = q.Where("id <> ?", except)
	}
	if err := q.Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return ErrDuplicate
	}
	return nil
}

func rejectActiveRole(db *gorm.DB, roleCode string, except uint64) error {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return nil
	}
	var n int64
	q := db.Model(&model.SysRole{}).Where("role_code = ?", roleCode)
	if except > 0 {
		q = q.Where("id <> ?", except)
	}
	if err := q.Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return ErrDuplicate
	}
	return nil
}
