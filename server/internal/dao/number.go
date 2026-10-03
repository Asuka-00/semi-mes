package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"semi-mes/server/internal/model"
)

// ErrNumber means the rule is missing, disabled, or a sequence could not be reserved.
var ErrNumber = errors.New("number rule")

var errNumberRetry = errors.New("number retry")

// Allocate reserves the next number for a rule. It is safe to call from several requests on SQLite, MySQL, and PostgreSQL.
func Allocate(db *gorm.DB, ruleCode string, now time.Time) (string, error) {
	rule, err := enabledRule(db, ruleCode)
	if err != nil {
		return "", err
	}
	period := periodKey(rule.ResetPeriod, now)
	var last error
	for attempt := 0; attempt < 16; attempt++ {
		if attempt > 0 {
			// Immediate retries livelock when several SQLite writers deadlock on the same sequence row.
			delay := time.Duration(attempt) * 5 * time.Millisecond
			if delay > 50*time.Millisecond {
				delay = 50 * time.Millisecond
			}
			time.Sleep(delay)
		}
		issued := 0
		err = allocateTx(db, func(tx *gorm.DB) error {
			var seq model.SysNumberSeq
			q := tx
			if rowLock(tx) {
				q = tx.Clauses(clause.Locking{Strength: "UPDATE"})
			}
			err := q.Where("rule_code = ? AND period_key = ?", rule.RuleCode, period).First(&seq).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				seq = model.SysNumberSeq{RuleCode: rule.RuleCode, PeriodKey: period, Current: 1, UpdatedAt: &now}
				if err = tx.Create(&seq).Error; err != nil {
					return errNumberRetry
				}
				issued = 1
				return nil
			}
			if err != nil {
				return err
			}
			next := seq.Current + 1
			res := tx.Model(&model.SysNumberSeq{}).Where("id = ? AND current = ?", seq.ID, seq.Current).Update("current", next)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 1 {
				return errNumberRetry
			}
			issued = next
			return nil
		})
		if err == nil {
			return formatRule(*rule, now, issued), nil
		}
		if !errors.Is(err, errNumberRetry) && !isUniqueConflict(err) {
			return "", err
		}
		last = err
	}
	if last == nil {
		last = ErrNumber
	}
	return "", fmt.Errorf("%w: %v", ErrNumber, last)
}

// PreviewNumber shows the next code without consuming a sequence.
func PreviewNumber(db *gorm.DB, ruleCode string, now time.Time) (string, error) {
	rule, err := enabledRule(db, ruleCode)
	if err != nil {
		return "", err
	}
	var seq model.SysNumberSeq
	_ = db.Where("rule_code = ? AND period_key = ?", rule.RuleCode, periodKey(rule.ResetPeriod, now)).First(&seq).Error
	return formatRule(*rule, now, seq.Current+1), nil
}

func enabledRule(db *gorm.DB, ruleCode string) (*model.SysNumberRule, error) {
	var rule model.SysNumberRule
	err := db.Where("rule_code = ? AND status = 1", strings.TrimSpace(ruleCode)).First(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNumber
		}
		return nil, err
	}
	if rule.SeqLength < 1 {
		rule.SeqLength = 3
	}
	if rule.SeqLength > 12 {
		rule.SeqLength = 12
	}
	return &rule, nil
}

func periodKey(reset string, now time.Time) string {
	switch strings.ToLower(strings.TrimSpace(reset)) {
	case "daily":
		return now.Format("20060102")
	case "monthly":
		return now.Format("200601")
	case "yearly":
		return now.Format("2006")
	default:
		return "*"
	}
}

func formatRule(rule model.SysNumberRule, now time.Time, seq int) string {
	parts := make([]string, 0, 3)
	if prefix := strings.TrimSpace(rule.Prefix); prefix != "" {
		parts = append(parts, prefix)
	}
	if date := datePart(rule.DatePart, now); date != "" {
		parts = append(parts, date)
	}
	parts = append(parts, fmt.Sprintf("%0*d", rule.SeqLength, seq))
	sep := rule.Separator
	if sep == "" {
		return strings.Join(parts, "")
	}
	return strings.Join(parts, sep)
}

func datePart(part string, now time.Time) string {
	switch strings.TrimSpace(part) {
	case "", "none":
		return ""
	case "yyyy":
		return now.Format("2006")
	case "yyyyMM":
		return now.Format("200601")
	default:
		return now.Format("20060102")
	}
}

// allocateTx reserves the sequence inside one transaction.
// SQLite deferred transactions deadlock when two connections read the row and then both try to write it.
// BEGIN IMMEDIATE takes the write lock before that read, so the other writer waits instead of deadlocking.
func allocateTx(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	if db != nil && db.Dialector != nil && db.Dialector.Name() == "sqlite" && !inTx(db) {
		return immediateSQLite(db, fn)
	}
	return db.Transaction(fn)
}

func inTx(db *gorm.DB) bool {
	if db == nil || db.Statement == nil || db.Statement.ConnPool == nil {
		return false
	}
	_, ok := db.Statement.ConnPool.(gorm.Tx)
	return ok
}

func immediateSQLite(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err = conn.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		return err
	}
	sess := db.Session(&gorm.Session{Context: ctx, NewDB: true, SkipDefaultTransaction: true})
	sess.Statement.ConnPool = conn
	if err = fn(sess); err != nil {
		_, _ = conn.ExecContext(ctx, "ROLLBACK")
		return err
	}
	if _, err = conn.ExecContext(ctx, "COMMIT"); err != nil {
		_, _ = conn.ExecContext(ctx, "ROLLBACK")
		return err
	}
	return nil
}

func rowLock(db *gorm.DB) bool {
	if db == nil || db.Dialector == nil {
		return false
	}
	switch db.Dialector.Name() {
	case "mysql", "postgres", "postgresql":
		return true
	default:
		return false
	}
}

func isUniqueConflict(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "unique") || strings.Contains(text, "duplicate") || strings.Contains(text, "locked") || strings.Contains(text, "busy")
}
