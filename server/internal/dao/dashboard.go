package dao

import (
	"database/sql/driver"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

const dashboardHoldLimit = 8

// DashBucket is one chart group.
type DashBucket struct {
	Key   string `gorm:"column:bucket" json:"key"`
	Label string `gorm:"column:label" json:"label"`
	Count int    `json:"count"`
	Qty   int    `json:"qty"`
}

// DashHold is one held lot with how long it has been held.
type DashHold struct {
	ID             uint64     `gorm:"column:id" json:"id"`
	LotNo          string     `gorm:"column:lot_no" json:"lotNo"`
	ProductCode    string     `gorm:"column:product_code" json:"productCode"`
	NodeKey        string     `gorm:"column:current_node_key" json:"nodeKey"`
	NodeName       string     `gorm:"column:node_name" json:"nodeName"`
	Quantity       int        `gorm:"column:quantity" json:"quantity"`
	HoldReasonCode string     `gorm:"column:hold_reason_code" json:"holdReasonCode"`
	HoldReason     string     `gorm:"column:hold_reason" json:"holdReason"`
	HeldAt         *time.Time `gorm:"column:held_at" json:"heldAt"`
	HoldSeconds    int        `gorm:"-" json:"holdSeconds"`
}

// DashDay is one day of completed track-outs.
type DashDay struct {
	Day   string `json:"day"`
	Moves int    `json:"moves"`
	Scrap int    `json:"scrap"`
}

// DashEvent is a recent lot history row.
type DashEvent struct {
	ID        uint64     `gorm:"column:id" json:"id"`
	LotID     uint64     `gorm:"column:lot_id" json:"lotId"`
	LotNo     string     `gorm:"column:lot_no" json:"lotNo"`
	EventType string     `gorm:"column:event_type" json:"eventType"`
	Reason    string     `gorm:"column:reason" json:"reason"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"createdAt"`
}

// Dashboard is the home-page shop-floor summary.
// Sections the caller cannot see are empty and the matching flag is false.
type Dashboard struct {
	UpdatedAt      time.Time    `json:"updatedAt"`
	Wip            bool         `json:"wip"`
	Equipment      bool         `json:"equipment"`
	WipLots        int          `json:"wipLots"`
	WipQty         int          `json:"wipQty"`
	HoldLots       int          `json:"holdLots"`
	RunningLots    int          `json:"runningLots"`
	TodayMoves     int          `json:"todayMoves"`
	TodayCompleted int          `json:"todayCompleted"`
	TodayScrap     int          `json:"todayScrap"`
	ByStep         []DashBucket `json:"byStep"`
	ByProduct      []DashBucket `json:"byProduct"`
	ByStatus       []DashBucket `json:"byStatus"`
	Holds          []DashHold   `json:"holds"`
	Trend          []DashDay    `json:"trend"`
	Events         []DashEvent  `json:"events"`
	ByEquipment    []DashBucket `json:"byEquipment"`
	OverduePm      int          `json:"overduePm"`
	Notices        []NoticeView `json:"notices"`
	Unread         int          `json:"unread"`
}

type movePoint struct {
	At    time.Time `gorm:"column:track_out_at"`
	Scrap int       `gorm:"column:qty_scrap"`
}

type holdScan struct {
	ID             uint64     `gorm:"column:id"`
	LotNo          string     `gorm:"column:lot_no"`
	ProductCode    string     `gorm:"column:product_code"`
	NodeKey        string     `gorm:"column:current_node_key"`
	NodeName       string     `gorm:"column:node_name"`
	Quantity       int        `gorm:"column:quantity"`
	HoldReasonCode string     `gorm:"column:hold_reason_code"`
	HoldReason     string     `gorm:"column:hold_reason"`
	HeldRaw        flexTime   `gorm:"column:held_at"`
	ArrivedAt      *time.Time `gorm:"column:arrived_at"`
}

func (row holdScan) view() DashHold {
	return DashHold{
		ID: row.ID, LotNo: row.LotNo, ProductCode: row.ProductCode, NodeKey: row.NodeKey,
		NodeName: row.NodeName, Quantity: row.Quantity, HoldReasonCode: row.HoldReasonCode, HoldReason: row.HoldReason,
	}
}

// flexTime scans a timestamp that SQLite may return as text and MySQL/Postgres return as time.Time.
type flexTime struct {
	T *time.Time
}

func (f flexTime) Value() (driver.Value, error) {
	if f.T == nil {
		return nil, nil
	}
	return *f.T, nil
}

func (f *flexTime) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		f.T = nil
		return nil
	case time.Time:
		t := v
		f.T = &t
		return nil
	case string:
		f.T = parseStringTime(v)
		return nil
	case []byte:
		f.T = parseStringTime(string(v))
		return nil
	default:
		return fmt.Errorf("unsupported time %T", src)
	}
}

func parseStringTime(raw string) *time.Time {
	if raw == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return &parsed
		}
	}
	return nil
}

// ShopDashboard aggregates WIP, equipment, and the caller's notices.
// wip and equipment gate those sections so a user without the permission does not receive the numbers.
func ShopDashboard(db *gorm.DB, userID uint64, wip, equipment bool) (*Dashboard, error) {
	now := time.Now()
	out := &Dashboard{
		UpdatedAt: now, Wip: wip, Equipment: equipment,
		ByStep: []DashBucket{}, ByProduct: []DashBucket{}, ByStatus: []DashBucket{},
		Holds: []DashHold{}, Trend: []DashDay{}, Events: []DashEvent{}, ByEquipment: []DashBucket{},
		Notices: []NoticeView{},
	}
	if wip {
		if err := fillWip(db, out, now); err != nil {
			return nil, err
		}
	}
	if equipment {
		if err := fillEquipment(db, out); err != nil {
			return nil, err
		}
	}
	notices, unread, err := ListNotices(db, userID, 6)
	if err != nil {
		return nil, err
	}
	if notices != nil {
		out.Notices = notices
	}
	out.Unread = unread
	return out, nil
}

func fillWip(db *gorm.DB, out *Dashboard, now time.Time) error {
	active := []string{model.LotWaiting, model.LotRunning, model.LotHold}
	if err := db.Table("wip_lot").
		Select("status AS bucket, COUNT(*) AS count, COALESCE(SUM(quantity),0) AS qty").
		Where("deleted_at IS NULL AND status <> ?", model.LotMerged).
		Group("status").Scan(&out.ByStatus).Error; err != nil {
		return err
	}
	for _, row := range out.ByStatus {
		switch row.Key {
		case model.LotWaiting, model.LotRunning, model.LotHold:
			out.WipLots += row.Count
			out.WipQty += row.Qty
		}
		if row.Key == model.LotHold {
			out.HoldLots = row.Count
		}
		if row.Key == model.LotRunning {
			out.RunningLots = row.Count
		}
	}
	if err := db.Table("wip_lot l").
		Select("l.current_node_key AS bucket, COALESCE(MAX(n.name), l.current_node_key) AS label, COUNT(*) AS count, COALESCE(SUM(l.quantity),0) AS qty").
		Joins("LEFT JOIN base_route_node n ON n.version_id = l.route_version_id AND n.node_key = l.current_node_key AND n.deleted_at IS NULL").
		Where("l.deleted_at IS NULL AND l.status IN ?", active).
		Group("l.current_node_key").
		Order("count DESC").
		Scan(&out.ByStep).Error; err != nil {
		return err
	}
	if err := db.Table("wip_lot l").
		Select("COALESCE(MAX(p.product_code), '') AS bucket, COALESCE(MAX(p.product_name), MAX(p.product_code), '') AS label, COUNT(*) AS count, COALESCE(SUM(l.quantity),0) AS qty").
		Joins("LEFT JOIN base_product p ON p.id = l.product_id AND p.deleted_at IS NULL").
		Where("l.deleted_at IS NULL AND l.status IN ?", active).
		Group("p.product_code").
		Order("count DESC").
		Scan(&out.ByProduct).Error; err != nil {
		return err
	}
	var holds []holdScan
	if err := db.Table("wip_lot l").
		Select(`l.id, l.lot_no, COALESCE(p.product_code, '') AS product_code, l.current_node_key,
			COALESCE(n.name, l.current_node_key) AS node_name, l.quantity, l.hold_reason_code, l.hold_reason,
			h.held_at AS held_at, l.arrived_at AS arrived_at`).
		Joins("LEFT JOIN base_product p ON p.id = l.product_id AND p.deleted_at IS NULL").
		Joins("LEFT JOIN base_route_node n ON n.version_id = l.route_version_id AND n.node_key = l.current_node_key AND n.deleted_at IS NULL").
		Joins(`LEFT JOIN (
			SELECT lot_id, MAX(created_at) AS held_at FROM wip_lot_history WHERE event_type = ? GROUP BY lot_id
		) h ON h.lot_id = l.id`, model.EventHold).
		Where("l.deleted_at IS NULL AND l.status = ?", model.LotHold).
		Scan(&holds).Error; err != nil {
		return err
	}
	ranked := make([]DashHold, 0, len(holds))
	for _, row := range holds {
		item := row.view()
		item.HeldAt = row.HeldRaw.T
		if item.HeldAt == nil {
			item.HeldAt = row.ArrivedAt
		}
		if item.HeldAt != nil {
			sec := int(now.Sub(*item.HeldAt).Seconds())
			if sec < 0 {
				sec = 0
			}
			item.HoldSeconds = sec
		}
		ranked = append(ranked, item)
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].HoldSeconds == ranked[j].HoldSeconds {
			return ranked[i].ID < ranked[j].ID
		}
		return ranked[i].HoldSeconds > ranked[j].HoldSeconds
	})
	if len(ranked) > dashboardHoldLimit {
		ranked = ranked[:dashboardHoldLimit]
	}
	out.Holds = ranked

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	trendStart := start.AddDate(0, 0, -6)
	var points []movePoint
	if err := db.Table("wip_move").
		Select("track_out_at, qty_scrap").
		Where("state = ? AND track_out_at >= ? AND track_out_at < ?", model.MoveCompleted, trendStart, end).
		Scan(&points).Error; err != nil {
		return err
	}
	days := make([]DashDay, 7)
	index := map[string]int{}
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, -6+i)
		key := day.Format("2006-01-02")
		days[i] = DashDay{Day: key}
		index[key] = i
	}
	for _, point := range points {
		key := point.At.In(now.Location()).Format("2006-01-02")
		slot, ok := index[key]
		if !ok {
			continue
		}
		days[slot].Moves++
		days[slot].Scrap += point.Scrap
		if !point.At.Before(start) && point.At.Before(end) {
			out.TodayMoves++
			out.TodayScrap += point.Scrap
		}
	}
	out.Trend = days
	var completed []uint64
	if err := db.Table("wip_lot_history").
		Where("event_type = ? AND created_at >= ? AND created_at < ?", model.EventComplete, start, end).
		Distinct("lot_id").
		Pluck("lot_id", &completed).Error; err != nil {
		return err
	}
	out.TodayCompleted = len(completed)
	var events []DashEvent
	if err := db.Table("wip_lot_history h").
		Select("h.id, h.lot_id, l.lot_no, h.event_type, h.reason, h.created_at").
		Joins("JOIN wip_lot l ON l.id = h.lot_id AND l.deleted_at IS NULL").
		Order("h.id DESC").Limit(8).Scan(&events).Error; err != nil {
		return err
	}
	if events != nil {
		out.Events = events
	}
	if out.ByStatus == nil {
		out.ByStatus = []DashBucket{}
	}
	if out.ByStep == nil {
		out.ByStep = []DashBucket{}
	}
	if out.ByProduct == nil {
		out.ByProduct = []DashBucket{}
	}
	if out.Holds == nil {
		out.Holds = []DashHold{}
	}
	return nil
}

func fillEquipment(db *gorm.DB, out *Dashboard) error {
	if err := db.Table("eqp_equipment").
		Select("status AS bucket, status AS label, COUNT(*) AS count").
		Where("deleted_at IS NULL").
		Group("status").
		Scan(&out.ByEquipment).Error; err != nil {
		return err
	}
	var overdue int64
	if err := db.Model(&model.EqpPmTask{}).Where("status = ?", model.PmOverdue).Count(&overdue).Error; err != nil {
		return err
	}
	out.OverduePm = int(overdue)
	if out.ByEquipment == nil {
		out.ByEquipment = []DashBucket{}
	}
	return nil
}
