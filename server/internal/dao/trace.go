package dao

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

func (f flexTime) MarshalJSON() ([]byte, error) {
	if f.T == nil {
		return []byte("null"), nil
	}
	return json.Marshal(f.T.UTC().Format(time.RFC3339))
}

const traceMaxDepth = 20

// ErrTraceFilter means a reverse lookup was asked with no constraint.
var ErrTraceFilter = errors.New("trace filter required")

// TraceNode is one lot in the split/merge genealogy.
type TraceNode struct {
	LotID    uint64      `json:"lotId"`
	LotNo    string      `json:"lotNo"`
	Status   string      `json:"status"`
	Quantity int         `json:"quantity"`
	LinkType string      `json:"linkType"`
	LinkQty  int         `json:"linkQty"`
	Nodes    []TraceNode `json:"nodes"`
}

// TraceEvent is one lot history row plus the related lot number.
type TraceEvent struct {
	ID           uint64   `json:"id" gorm:"column:id"`
	EventType    string   `json:"eventType" gorm:"column:event_type"`
	FromNodeKey  string   `json:"fromNodeKey" gorm:"column:from_node_key"`
	ToNodeKey    string   `json:"toNodeKey" gorm:"column:to_node_key"`
	EdgeKey      string   `json:"edgeKey" gorm:"column:edge_key"`
	ReasonCode   string   `json:"reasonCode" gorm:"column:reason_code"`
	Reason       string   `json:"reason" gorm:"column:reason"`
	Quantity     int      `json:"quantity" gorm:"column:quantity"`
	RelatedLotID uint64   `json:"relatedLotID" gorm:"column:related_lot_id"`
	RelatedLotNo string   `json:"relatedLotNo" gorm:"column:related_lot_no"`
	CreatedAt    flexTime `json:"createdAt" gorm:"column:created_at"`
}

// TraceMove is one step move with equipment, operator, and recipe.
type TraceMove struct {
	ID               uint64   `json:"id" gorm:"column:id"`
	NodeKey          string   `json:"nodeKey" gorm:"column:node_key"`
	EquipmentCode    string   `json:"equipmentCode" gorm:"column:equipment_code"`
	EquipmentName    string   `json:"equipmentName" gorm:"column:equipment_name"`
	OperatorName     string   `json:"operatorName" gorm:"column:operator_name"`
	RecipeCode       string   `json:"recipeCode" gorm:"column:recipe_code"`
	Parameters       string   `json:"parameters" gorm:"column:parameters"`
	QtyIn            int      `json:"qtyIn" gorm:"column:qty_in"`
	QtyOut           int      `json:"qtyOut" gorm:"column:qty_out"`
	QtyScrap         int      `json:"qtyScrap" gorm:"column:qty_scrap"`
	State            string   `json:"state" gorm:"column:state"`
	TrackInAt        flexTime `json:"trackInAt" gorm:"column:track_in_at"`
	TrackOutAt       flexTime `json:"trackOutAt" gorm:"column:track_out_at"`
	InspectionResult string   `json:"inspectionResult" gorm:"column:inspection_result"`
	InspectionGrade  string   `json:"inspectionGrade" gorm:"column:inspection_grade"`
	DefectCode       string   `json:"defectCode" gorm:"column:defect_code"`
	FromNodeKey      string   `json:"fromNodeKey" gorm:"column:from_node_key"`
	ToNodeKey        string   `json:"toNodeKey" gorm:"column:to_node_key"`
	EdgeKey          string   `json:"edgeKey" gorm:"column:edge_key"`
	ResolveReason    string   `json:"resolveReason" gorm:"column:resolve_reason"`
}

// TraceLotRef is the lot header on a trace report or a reverse-lookup hit.
type TraceLotRef struct {
	ID             uint64 `json:"id" gorm:"column:id"`
	LotNo          string `json:"lotNo" gorm:"column:lot_no"`
	Status         string `json:"status" gorm:"column:status"`
	Quantity       int    `json:"quantity" gorm:"column:quantity"`
	CurrentNodeKey string `json:"currentNodeKey" gorm:"column:current_node_key"`
	LotType        string `json:"lotType" gorm:"column:lot_type"`
	HoldReasonCode string `json:"holdReasonCode" gorm:"column:hold_reason_code"`
	HoldReason     string `json:"holdReason" gorm:"column:hold_reason"`
}

// TraceReport is the forward and backward history of one lot.
type TraceReport struct {
	Lot          TraceLotRef           `json:"lot"`
	Backward     []TraceNode           `json:"backward"`
	Forward      []TraceNode           `json:"forward"`
	History      []TraceEvent          `json:"history"`
	Moves        []TraceMove           `json:"moves"`
	Defects      []model.QcDefect      `json:"defects"`
	Measurements []model.QcMeasurement `json:"measurements"`
	Wafers       []any                 `json:"wafers"`
}

// TraceReverseQuery filters lots that passed a tool, recipe, operation, or time window.
type TraceReverseQuery struct {
	EquipmentID   uint64
	EquipmentCode string
	RecipeID      uint64
	OperationID   uint64
	NodeKey       string
	From          string
	ToExclusive   string
}

// LotTrace builds the genealogy and the step, hold, inspection, and defect history.
func LotTrace(db *gorm.DB, lotID uint64) (*TraceReport, error) {
	var lot model.WipLot
	if err := db.First(&lot, lotID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipNotFound
		}
		return nil, err
	}
	report := &TraceReport{
		Lot: TraceLotRef{
			ID: lot.ID, LotNo: lot.LotNo, Status: lot.Status, Quantity: lot.Quantity,
			CurrentNodeKey: lot.CurrentNodeKey, LotType: lot.LotType,
			HoldReasonCode: lot.HoldReasonCode, HoldReason: lot.HoldReason,
		},
		Backward: walkTrace(db, lot.ID, false, 0, map[uint64]struct{}{lot.ID: {}}),
		Forward:  walkTrace(db, lot.ID, true, 0, map[uint64]struct{}{lot.ID: {}}),
		Wafers:   []any{},
	}
	if report.Backward == nil {
		report.Backward = []TraceNode{}
	}
	if report.Forward == nil {
		report.Forward = []TraceNode{}
	}
	if err := db.Table("wip_lot_history h").
		Select("h.id, h.event_type, h.from_node_key, h.to_node_key, h.edge_key, h.reason_code, h.reason, h.quantity, h.related_lot_id, h.created_at, rl.lot_no AS related_lot_no").
		Joins("LEFT JOIN wip_lot rl ON rl.id = h.related_lot_id").
		Where("h.lot_id = ?", lot.ID).
		Order("h.id ASC").
		Scan(&report.History).Error; err != nil {
		return nil, err
	}
	if report.History == nil {
		report.History = []TraceEvent{}
	}
	if err := db.Table("wip_move m").
		Select(`m.id, m.node_key, m.qty_in, m.qty_out, m.qty_scrap, m.state, m.track_in_at, m.track_out_at,
			m.inspection_result, m.inspection_grade, m.defect_code, m.from_node_key, m.to_node_key, m.edge_key, m.resolve_reason,
			e.equipment_code, e.equipment_name, u.username AS operator_name, r.recipe_code, r.parameters`).
		Joins("LEFT JOIN eqp_equipment e ON e.id = m.equipment_id AND e.deleted_at IS NULL").
		Joins("LEFT JOIN sys_user u ON u.id = m.operator_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN base_recipe r ON r.id = m.recipe_id AND r.deleted_at IS NULL").
		Where("m.lot_id = ?", lot.ID).
		Order("m.id ASC").
		Scan(&report.Moves).Error; err != nil {
		return nil, err
	}
	if report.Moves == nil {
		report.Moves = []TraceMove{}
	}
	if err := db.Where("lot_id = ?", lot.ID).Order("id ASC").Find(&report.Defects).Error; err != nil {
		return nil, err
	}
	if report.Defects == nil {
		report.Defects = []model.QcDefect{}
	}
	if err := db.Where("lot_id = ?", lot.ID).Order("id ASC").Find(&report.Measurements).Error; err != nil {
		return nil, err
	}
	if report.Measurements == nil {
		report.Measurements = []model.QcMeasurement{}
	}
	return report, nil
}

// FindLotByNo returns the lot id for a lot number.
func FindLotByNo(db *gorm.DB, lotNo string) (uint64, error) {
	var lot model.WipLot
	err := db.Where("lot_no = ?", strings.TrimSpace(lotNo)).First(&lot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrWipNotFound
	}
	return lot.ID, err
}

// ReverseTrace lists lots processed by the given equipment, recipe, operation, or time window.
func ReverseTrace(db *gorm.DB, query TraceReverseQuery) ([]TraceLotRef, error) {
	if query.EquipmentID == 0 && strings.TrimSpace(query.EquipmentCode) == "" && query.RecipeID == 0 &&
		query.OperationID == 0 && strings.TrimSpace(query.NodeKey) == "" && query.From == "" && query.ToExclusive == "" {
		return nil, ErrTraceFilter
	}
	q := db.Table("wip_move m").
		Select("DISTINCT l.id, l.lot_no, l.status, l.quantity, l.current_node_key, l.lot_type, l.hold_reason_code, l.hold_reason").
		Joins("JOIN wip_lot l ON l.id = m.lot_id AND l.deleted_at IS NULL").
		Joins("LEFT JOIN eqp_equipment e ON e.id = m.equipment_id")
	if query.EquipmentID > 0 {
		q = q.Where("m.equipment_id = ?", query.EquipmentID)
	}
	if code := strings.TrimSpace(query.EquipmentCode); code != "" {
		q = q.Where("e.equipment_code = ?", code)
	}
	if query.RecipeID > 0 {
		q = q.Where("m.recipe_id = ?", query.RecipeID)
	}
	if query.OperationID > 0 {
		q = q.Where("m.operation_id = ?", query.OperationID)
	}
	if key := strings.TrimSpace(query.NodeKey); key != "" {
		q = q.Where("m.node_key = ?", key)
	}
	if query.From != "" {
		q = q.Where("m.track_in_at >= ?", query.From)
	}
	if query.ToExclusive != "" {
		q = q.Where("m.track_in_at < ?", query.ToExclusive)
	}
	var rows []TraceLotRef
	err := q.Order("l.id DESC").Limit(200).Scan(&rows).Error
	if rows == nil {
		rows = []TraceLotRef{}
	}
	return rows, err
}

func walkTrace(db *gorm.DB, lotID uint64, forward bool, depth int, seen map[uint64]struct{}) []TraceNode {
	if depth >= traceMaxDepth {
		return []TraceNode{}
	}
	var links []model.WipLotLink
	if forward {
		_ = db.Where("parent_lot_id = ?", lotID).Order("id ASC").Find(&links).Error
	} else {
		_ = db.Where("child_lot_id = ?", lotID).Order("id ASC").Find(&links).Error
	}
	out := make([]TraceNode, 0, len(links))
	for _, link := range links {
		nextID := link.ChildLotID
		if !forward {
			nextID = link.ParentLotID
		}
		if _, ok := seen[nextID]; ok {
			continue
		}
		var lot model.WipLot
		if err := db.First(&lot, nextID).Error; err != nil {
			continue
		}
		childSeen := make(map[uint64]struct{}, len(seen)+1)
		for id := range seen {
			childSeen[id] = struct{}{}
		}
		childSeen[nextID] = struct{}{}
		out = append(out, TraceNode{
			LotID: lot.ID, LotNo: lot.LotNo, Status: lot.Status, Quantity: lot.Quantity,
			LinkType: link.LinkType, LinkQty: link.Quantity,
			Nodes: walkTrace(db, nextID, forward, depth+1, childSeen),
		})
	}
	return out
}
