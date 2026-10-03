package dao

import (
	"errors"
	"regexp"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

var pageKeyPattern = regexp.MustCompile(`^[A-Za-z0-9:_-]{1,80}$`)

// GetUserPref returns the saved payload. A missing row is an empty payload.
func GetUserPref(db *gorm.DB, userID uint64, pageKey string) (string, error) {
	if !pageKeyPattern.MatchString(pageKey) {
		return "", ErrWipState
	}
	var row model.SysUserPref
	err := db.Where("user_id = ? AND page_key = ?", userID, pageKey).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return row.Payload, nil
}

// SaveUserPref upserts the page payload for one user.
func SaveUserPref(db *gorm.DB, userID uint64, pageKey, payload string) error {
	if !pageKeyPattern.MatchString(pageKey) || !utf8.ValidString(payload) || len(payload) > 20000 {
		return ErrWipState
	}
	var row model.SysUserPref
	err := db.Where("user_id = ? AND page_key = ?", userID, pageKey).First(&row).Error
	now := time.Now()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = model.SysUserPref{UserID: userID, PageKey: pageKey, Payload: payload, UpdatedAt: &now}
		return db.Create(&row).Error
	}
	if err != nil {
		return err
	}
	return db.Model(&row).Updates(map[string]any{"payload": payload, "updated_at": now}).Error
}
