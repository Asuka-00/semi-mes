package dao

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// NoticeView is a notice with both languages filled in.
type NoticeView struct {
	ID        uint64     `json:"id"`
	Kind      string     `json:"kind"`
	TitleZh   string     `json:"titleZh"`
	TitleEn   string     `json:"titleEn"`
	BodyZh    string     `json:"bodyZh"`
	BodyEn    string     `json:"bodyEn"`
	Read      bool       `json:"read"`
	CreatedAt *time.Time `json:"createdAt"`
	RefType   string     `json:"refType"`
	RefID     uint64     `json:"refId"`
}

func notifyUsers(tx *gorm.DB, kind string, params map[string]string, refType string, refID uint64) error {
	if tx == nil {
		return nil
	}
	var ids []uint64
	if err := tx.Model(&model.SysUser{}).Where("status = ?", 1).Pluck("id", &ids).Error; err != nil {
		return err
	}
	raw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	for _, id := range ids {
		row := model.SysNotification{UserID: id, Kind: kind, Params: string(raw), RefType: refType, RefID: refID}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func notifyLot(tx *gorm.DB, lot *model.WipLot, kind string) error {
	if lot == nil {
		return nil
	}
	return notifyUsers(tx, kind, map[string]string{
		"lotNo":      lot.LotNo,
		"reason":     lot.HoldReason,
		"reasonCode": lot.HoldReasonCode,
	}, "lot", lot.ID)
}

// ListNotices returns the newest notices for one user.
func ListNotices(db *gorm.DB, userID uint64, limit int) ([]NoticeView, int, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var rows []model.SysNotification
	if err := db.Where("user_id = ?", userID).Order("id desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	var unread int64
	if err := db.Model(&model.SysNotification{}).Where("user_id = ? AND read_at IS NULL", userID).Count(&unread).Error; err != nil {
		return nil, 0, err
	}
	out := make([]NoticeView, 0, len(rows))
	for _, row := range rows {
		params := map[string]string{}
		_ = json.Unmarshal([]byte(row.Params), &params)
		zhTitle, enTitle, zhBody, enBody := noticeText(row.Kind, params)
		out = append(out, NoticeView{
			ID: row.ID, Kind: row.Kind, TitleZh: zhTitle, TitleEn: enTitle, BodyZh: zhBody, BodyEn: enBody,
			Read: row.ReadAt != nil, CreatedAt: row.CreatedAt, RefType: row.RefType, RefID: row.RefID,
		})
	}
	return out, int(unread), nil
}

// MarkNoticeRead marks one notice read for its owner.
func MarkNoticeRead(db *gorm.DB, userID, id uint64) error {
	now := time.Now()
	res := db.Model(&model.SysNotification{}).Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).Update("read_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var n int64
		if err := db.Model(&model.SysNotification{}).Where("id = ? AND user_id = ?", id, userID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return ErrWipNotFound
		}
	}
	return nil
}

// MarkAllNoticesRead marks every unread notice for the user.
func MarkAllNoticesRead(db *gorm.DB, userID uint64) error {
	return db.Model(&model.SysNotification{}).Where("user_id = ? AND read_at IS NULL", userID).Update("read_at", time.Now()).Error
}

func noticeText(kind string, params map[string]string) (string, string, string, string) {
	lot := params["lotNo"]
	reason := params["reason"]
	if reason == "" {
		reason = params["reasonCode"]
	}
	eqp := params["eqpCode"]
	plan := params["planName"]
	switch kind {
	case model.NoticeReworkExceeded:
		return "返工超限", "Rework limit",
			fmt.Sprintf("批次 %s 返工次数已达上限，已 Hold。%s", lot, reason),
			fmt.Sprintf("Lot %s exceeded the rework limit and was held. %s", lot, reason)
	case model.NoticeLotHold:
		return "批次 Hold", "Lot hold",
			fmt.Sprintf("批次 %s 已 Hold。%s", lot, reason),
			fmt.Sprintf("Lot %s is on hold. %s", lot, reason)
	case model.NoticePmDue:
		return "PM 到期", "PM due",
			fmt.Sprintf("设备 %s 的保养「%s」已到期。", eqp, plan),
			fmt.Sprintf("PM \"%s\" is due on %s.", plan, eqp)
	case model.NoticePmOverdue:
		return "PM 超期", "PM overdue",
			fmt.Sprintf("设备 %s 的保养「%s」已超期。", eqp, plan),
			fmt.Sprintf("PM \"%s\" is overdue on %s.", plan, eqp)
	case model.NoticeEqpDown:
		return "设备停机", "Equipment down",
			fmt.Sprintf("设备 %s 非计划停机。%s", eqp, reason),
			fmt.Sprintf("%s is unscheduled down. %s", eqp, reason)
	default:
		return kind, kind, kind, kind
	}
}
