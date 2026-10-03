package dao

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
	"semi-mes/server/internal/routegraph"
)

// ScrapReasons are the reason codes accepted on Track Out.
var ScrapReasons = map[string]bool{"BROKEN": true, "PARTICLE": true, "SCRATCH": true, "OTHER": true}

// TrackInInput checks a waiting lot into an equipment.
type TrackInInput struct {
	LotID       uint64
	EquipmentID uint64
	RecipeID    uint64
	OperatorID  uint64
}

// TrackOutInput closes the open Track In and moves the lot.
type TrackOutInput struct {
	LotID            uint64
	QtyOut           int
	QtyScrap         int
	ScrapReasonCode  string
	InspectionResult string
	InspectionGrade  string
	DefectCode       string
	Measurements     []SampleInput
	OperatorID       uint64
}

// PassInput leaves a start or decision node without equipment.
type PassInput struct {
	LotID            uint64
	InspectionResult string
	InspectionGrade  string
	DefectCode       string
	OperatorID       uint64
}

// MoveView is one history row for the WIP page.
type MoveView struct {
	model.WipMove
	LotNo         string `json:"lotNo"`
	EquipmentCode string `json:"equipmentCode"`
	EquipmentName string `json:"equipmentName"`
	OperatorName  string `json:"operatorName"`
	NodeName      string `json:"nodeName"`
	ProductCode   string `json:"productCode"`
}

// StationView is what the workstation shows for one lot.
type StationView struct {
	Lot                LotView              `json:"lot"`
	NodeType           string               `json:"nodeType"`
	NodeName           string               `json:"nodeName"`
	EquipmentGroup     string               `json:"equipmentGroup"`
	OperationID        uint64               `json:"operationID"`
	RecipeID           uint64               `json:"recipeID"`
	InspectionRequired bool                 `json:"inspectionRequired"`
	InspectPlan        *PlanView            `json:"inspectPlan"`
	LatestResult       string               `json:"latestResult"`
	OpenMove           *model.WipMove       `json:"openMove"`
	Equipment          []model.EqpEquipment `json:"equipment"`
}

// StatusCount is one bucket on the WIP overview.
type StatusCount struct {
	Key   string `gorm:"column:bucket" json:"key"`
	Count int    `json:"count"`
	Qty   int    `json:"qty"`
}

// Overview is the shop-floor summary.
type Overview struct {
	ByStatus  []StatusCount `json:"byStatus"`
	ByNode    []StatusCount `json:"byNode"`
	ByProduct []StatusCount `json:"byProduct"`
	Holds     []LotView     `json:"holds"`
}

// ListEquipment returns the stub equipment master, optionally one group.
func ListEquipment(db *gorm.DB, group string) ([]model.EqpEquipment, error) {
	q := db.Model(&model.EqpEquipment{}).Order("equipment_code")
	if group != "" {
		q = q.Where("equipment_group = ?", group)
	}
	var rows []model.EqpEquipment
	err := q.Find(&rows).Error
	return rows, err
}

// GetStation loads a lot and the equipment that may track it in.
func GetStation(db *gorm.DB, lotNo string) (*StationView, error) {
	lot, err := findLotByNo(db, lotNo)
	if err != nil {
		return nil, err
	}
	graph, err := LoadRouteGraph(db, lot.RouteVersionID)
	if err != nil {
		return nil, err
	}
	node, ok := findNode(graph, lot.CurrentNodeKey)
	if !ok {
		return nil, ErrWipVersion
	}
	view := &StationView{
		NodeType: node.Type, NodeName: node.Name, EquipmentGroup: node.EquipmentGroup,
		OperationID: node.OperationID, RecipeID: node.RecipeID,
		InspectionRequired: inspectionRequired(graph, node.Key),
	}
	rows, _, err := ListLots(db, 0, 1, "-id", []Column{{Name: "lot_no", Exp: "=", Value: lot.LotNo}})
	if err != nil {
		return nil, err
	}
	if len(rows) > 0 {
		view.Lot = rows[0]
	}
	if node.Type == routegraph.NodeOperation {
		eqp, err := allowedEquipment(db, node.EquipmentGroup, node.OperationID, node.RecipeID)
		if err != nil {
			return nil, err
		}
		view.Equipment = eqp
	}
	if view.Equipment == nil {
		view.Equipment = []model.EqpEquipment{}
	}
	if node.OperationID != 0 {
		plan, err := FindInspectPlan(db, node.OperationID, lot.ProductID)
		if err != nil {
			return nil, err
		}
		view.InspectPlan = plan
	}
	judged, err := LatestJudgement(db, lot.ID)
	if err != nil {
		return nil, err
	}
	view.LatestResult = judged
	open, err := openMove(db, lot.ID)
	if err != nil {
		return nil, err
	}
	view.OpenMove = open
	return view, nil
}

// TrackIn checks a waiting operation lot into an allowed equipment.
func TrackIn(db *gorm.DB, in TrackInInput) (*model.WipLot, *model.WipMove, error) {
	var lot *model.WipLot
	var move *model.WipMove
	err := db.Transaction(func(tx *gorm.DB) error {
		row, err := GetLot(tx, in.LotID)
		if err != nil {
			return err
		}
		if row.Status != model.LotWaiting || row.Quantity <= 0 {
			return ErrWipTrack
		}
		graph, err := LoadRouteGraph(tx, row.RouteVersionID)
		if err != nil {
			return err
		}
		node, ok := findNode(graph, row.CurrentNodeKey)
		if !ok || node.Type != routegraph.NodeOperation {
			return ErrWipTrack
		}
		eqp, err := loadEquipment(tx, in.EquipmentID)
		if err != nil {
			return err
		}
		recipeID, err := resolveRecipe(tx, node, in.RecipeID)
		if err != nil {
			return err
		}
		if err := TrackInGate(tx, eqp, node, recipeID); err != nil {
			return err
		}
		now := time.Now()
		queue := 0
		if row.ArrivedAt != nil {
			queue = int(now.Sub(*row.ArrivedAt).Seconds())
			if queue < 0 {
				queue = 0
			}
		}
		move = &model.WipMove{
			LotID: row.ID, NodeKey: node.Key, OperationID: node.OperationID, EquipmentID: eqp.ID, RecipeID: recipeID,
			OperatorID: in.OperatorID, QtyIn: row.Quantity, State: model.MoveOpen, TrackInAt: &now, QueueSeconds: queue,
			FromNodeKey: node.Key,
		}
		if err := tx.Create(move).Error; err != nil {
			return err
		}
		row.Status = model.LotRunning
		if err := tx.Model(row).Update("status", row.Status).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: row.ID, EventType: model.EventTrackIn, FromNodeKey: node.Key, ToNodeKey: node.Key,
			Quantity: row.Quantity, RelatedLotID: eqp.ID,
		}).Error; err != nil {
			return err
		}
		if err := occupyEquipment(tx, eqp, in.OperatorID); err != nil {
			return err
		}
		lot = row
		return nil
	})
	return lot, move, err
}

// AbortTrackIn cancels an open check-in and leaves the lot waiting on the same node.
func AbortTrackIn(db *gorm.DB, lotID, operatorID uint64, reason string) (*model.WipLot, error) {
	var lot *model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		row, err := GetLot(tx, lotID)
		if err != nil {
			return err
		}
		if row.Status != model.LotRunning {
			return ErrWipTrack
		}
		move, err := openMove(tx, row.ID)
		if err != nil || move == nil {
			return ErrWipTrack
		}
		now := time.Now()
		if err := tx.Model(move).Updates(map[string]any{
			"state": model.MoveAborted, "track_out_at": now, "abort_reason": reason, "operator_id": operatorID,
			"process_seconds": secondsSince(move.TrackInAt, now),
		}).Error; err != nil {
			return err
		}
		row.Status = model.LotWaiting
		if err := tx.Model(row).Update("status", row.Status).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: row.ID, EventType: model.EventAbort, FromNodeKey: row.CurrentNodeKey, ToNodeKey: row.CurrentNodeKey,
			Reason: reason, Quantity: row.Quantity,
		}).Error; err != nil {
			return err
		}
		if err := releaseEquipment(tx, move.EquipmentID, operatorID); err != nil {
			return err
		}
		lot = row
		return nil
	})
	return lot, err
}

// TrackOut records output and scrap, then uses the route resolver to leave the step.
func TrackOut(db *gorm.DB, in TrackOutInput) (*model.WipLot, routegraph.Result, error) {
	var lot *model.WipLot
	var result routegraph.Result
	err := db.Transaction(func(tx *gorm.DB) error {
		row, err := GetLot(tx, in.LotID)
		if err != nil {
			return err
		}
		if row.Status != model.LotRunning {
			return ErrWipTrack
		}
		move, err := openMove(tx, row.ID)
		if err != nil || move == nil {
			return ErrWipTrack
		}
		if in.QtyOut < 0 || in.QtyScrap < 0 || in.QtyOut+in.QtyScrap != move.QtyIn {
			return ErrWipQty
		}
		if in.QtyScrap > 0 && !AcceptCode(tx, "scrap", in.ScrapReasonCode, ScrapReasons) {
			return ErrWipQty
		}
		graph, err := LoadRouteGraph(tx, row.RouteVersionID)
		if err != nil {
			return err
		}
		node, ok := findNode(graph, row.CurrentNodeKey)
		if !ok {
			return ErrWipVersion
		}
		plan, err := FindInspectPlan(tx, node.OperationID, row.ProductID)
		if err != nil {
			return err
		}
		if plan != nil {
			judged, err := recordMeasurementsTx(tx, MeasureInput{
				LotID: row.ID, MoveID: move.ID, EquipmentID: move.EquipmentID, OperationID: node.OperationID,
				OperatorID: in.OperatorID, Samples: in.Measurements,
			})
			if err != nil {
				return err
			}
			in.InspectionResult = judged
		}
		if inspectionRequired(graph, row.CurrentNodeKey) && strings.TrimSpace(in.InspectionResult) == "" {
			return ErrWipInspect
		}
		row.Quantity = in.QtyOut
		now := time.Now()
		if in.QtyOut == 0 {
			row.Status = model.LotScrapped
			if err := tx.Model(row).Updates(map[string]any{"quantity": 0, "status": row.Status}).Error; err != nil {
				return err
			}
			if err := closeMove(tx, move, in, now, row.CurrentNodeKey, routegraph.Result{Action: "scrap", Reason: in.ScrapReasonCode}); err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: row.ID, EventType: model.EventScrap, FromNodeKey: row.CurrentNodeKey, ToNodeKey: row.CurrentNodeKey,
				ReasonCode: in.ScrapReasonCode, Quantity: in.QtyScrap,
			}).Error; err != nil {
				return err
			}
			if err := releaseEquipment(tx, move.EquipmentID, in.OperatorID); err != nil {
				return err
			}
			if err := bumpPmLots(tx, move.EquipmentID); err != nil {
				return err
			}
			if err := EnsurePmTasks(tx); err != nil {
				return err
			}
			lot = row
			result = routegraph.Result{Action: "scrap", Reason: in.ScrapReasonCode}
			return refreshOrder(tx, row.OrderID)
		}
		result, err = moveLot(tx, row, AdvanceInput{
			InspectionResult: in.InspectionResult, InspectionGrade: in.InspectionGrade, DefectCode: in.DefectCode,
		}, model.EventTrackOut)
		if err != nil {
			return err
		}
		if err := closeMove(tx, move, in, now, row.CurrentNodeKey, result); err != nil {
			return err
		}
		if err := releaseEquipment(tx, move.EquipmentID, in.OperatorID); err != nil {
			return err
		}
		if err := bumpPmLots(tx, move.EquipmentID); err != nil {
			return err
		}
		if err := EnsurePmTasks(tx); err != nil {
			return err
		}
		lot = row
		return nil
	})
	return lot, result, err
}

// PassNode leaves a start or decision node. Operation steps must Track In first.
func PassNode(db *gorm.DB, in PassInput) (*model.WipLot, routegraph.Result, error) {
	var lot *model.WipLot
	var result routegraph.Result
	err := db.Transaction(func(tx *gorm.DB) error {
		row, err := GetLot(tx, in.LotID)
		if err != nil {
			return err
		}
		if row.Status != model.LotWaiting {
			return ErrWipTrack
		}
		graph, err := LoadRouteGraph(tx, row.RouteVersionID)
		if err != nil {
			return err
		}
		node, ok := findNode(graph, row.CurrentNodeKey)
		if !ok || node.Type == routegraph.NodeOperation || node.Type == routegraph.NodeEnd {
			return ErrWipTrack
		}
		if strings.TrimSpace(in.InspectionResult) == "" {
			judged, err := LatestJudgement(tx, row.ID)
			if err != nil {
				return err
			}
			if judged != "" {
				in.InspectionResult = judged
			}
		}
		if inspectionRequired(graph, node.Key) && strings.TrimSpace(in.InspectionResult) == "" {
			return ErrWipInspect
		}
		now := time.Now()
		queue := 0
		if row.ArrivedAt != nil {
			queue = int(now.Sub(*row.ArrivedAt).Seconds())
			if queue < 0 {
				queue = 0
			}
		}
		result, err = moveLot(tx, row, AdvanceInput{
			InspectionResult: in.InspectionResult, InspectionGrade: in.InspectionGrade, DefectCode: in.DefectCode,
		}, model.EventPass)
		if err != nil {
			return err
		}
		move := &model.WipMove{
			LotID: row.ID, NodeKey: node.Key, OperatorID: in.OperatorID, QtyIn: row.Quantity, QtyOut: row.Quantity,
			State: model.MoveCompleted, TrackInAt: &now, TrackOutAt: &now, QueueSeconds: queue,
			InspectionResult: in.InspectionResult, InspectionGrade: in.InspectionGrade, DefectCode: in.DefectCode,
			FromNodeKey: node.Key, ToNodeKey: row.CurrentNodeKey, EdgeKey: result.EdgeKey,
			ResolveAction: result.Action, ResolveReason: result.Reason,
		}
		if result.Action == routegraph.ActionHold {
			move.ToNodeKey = node.Key
			move.QtyOut = row.Quantity
		}
		if err := tx.Create(move).Error; err != nil {
			return err
		}
		lot = row
		return nil
	})
	return lot, result, err
}

// ListMoves returns a page of move history.
func ListMoves(db *gorm.DB, page, limit int, sort string, columns []Column) ([]MoveView, int64, error) {
	allow := map[string]string{
		"lot_no": "l.lot_no", "state": "m.state", "node_key": "m.node_key", "equipment_code": "e.equipment_code",
		"equipment_id": "m.equipment_id", "product_code": "p.product_code",
		"track_out_at": "m.track_out_at", "created_at": "m.created_at", "id": "m.id",
	}
	q := db.Table("wip_move m").
		Select("m.*, l.lot_no, e.equipment_code, e.equipment_name, u.username AS operator_name, n.name AS node_name, p.product_code").
		Joins("JOIN wip_lot l ON l.id = m.lot_id").
		Joins("LEFT JOIN eqp_equipment e ON e.id = m.equipment_id").
		Joins("LEFT JOIN sys_user u ON u.id = m.operator_id").
		Joins("LEFT JOIN base_route_node n ON n.version_id = l.route_version_id AND n.node_key = m.node_key AND n.deleted_at IS NULL").
		Joins("LEFT JOIN base_product p ON p.id = l.product_id")
	q = applyColumns(q, columns, allow)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []MoveView
	err := q.Order(orderClause(sort, allow, "m.id DESC")).Offset(page * limit).Limit(limit).Scan(&rows).Error
	return rows, total, err
}

// ListMovesByLot returns every move of one lot, oldest first.
func ListMovesByLot(db *gorm.DB, lotID uint64) ([]MoveView, error) {
	var rows []MoveView
	err := db.Table("wip_move m").
		Select("m.*, l.lot_no, e.equipment_code, e.equipment_name, u.username AS operator_name, n.name AS node_name").
		Joins("JOIN wip_lot l ON l.id = m.lot_id").
		Joins("LEFT JOIN eqp_equipment e ON e.id = m.equipment_id").
		Joins("LEFT JOIN sys_user u ON u.id = m.operator_id").
		Joins("LEFT JOIN base_route_node n ON n.version_id = l.route_version_id AND n.node_key = m.node_key AND n.deleted_at IS NULL").
		Where("m.lot_id = ?", lotID).Order("m.id ASC").Scan(&rows).Error
	return rows, err
}

// WIPOverview groups active and held lots.
func WIPOverview(db *gorm.DB) (*Overview, error) {
	out := &Overview{}
	if err := db.Table("wip_lot").Select("status AS bucket, COUNT(*) AS count, COALESCE(SUM(quantity),0) AS qty").
		Where("deleted_at IS NULL AND status <> ?", model.LotMerged).Group("status").Scan(&out.ByStatus).Error; err != nil {
		return nil, err
	}
	if err := db.Table("wip_lot l").
		Select("l.current_node_key AS bucket, COUNT(*) AS count, COALESCE(SUM(l.quantity),0) AS qty").
		Where("l.deleted_at IS NULL AND l.status IN ?", []string{model.LotWaiting, model.LotRunning, model.LotHold}).
		Group("l.current_node_key").Scan(&out.ByNode).Error; err != nil {
		return nil, err
	}
	if err := db.Table("wip_lot l").
		Select("p.product_code AS bucket, COUNT(*) AS count, COALESCE(SUM(l.quantity),0) AS qty").
		Joins("LEFT JOIN base_product p ON p.id = l.product_id").
		Where("l.deleted_at IS NULL AND l.status IN ?", []string{model.LotWaiting, model.LotRunning, model.LotHold}).
		Group("p.product_code").Scan(&out.ByProduct).Error; err != nil {
		return nil, err
	}
	holds, _, err := ListLots(db, 0, 50, "-id", []Column{{Name: "status", Exp: "=", Value: model.LotHold}})
	if err != nil {
		return nil, err
	}
	out.Holds = holds
	if out.ByStatus == nil {
		out.ByStatus = []StatusCount{}
	}
	if out.ByNode == nil {
		out.ByNode = []StatusCount{}
	}
	if out.ByProduct == nil {
		out.ByProduct = []StatusCount{}
	}
	if out.Holds == nil {
		out.Holds = []LotView{}
	}
	return out, nil
}

func closeMove(tx *gorm.DB, move *model.WipMove, in TrackOutInput, now time.Time, to string, result routegraph.Result) error {
	return tx.Model(move).Updates(map[string]any{
		"qty_out": in.QtyOut, "qty_scrap": in.QtyScrap, "scrap_reason_code": in.ScrapReasonCode,
		"inspection_result": in.InspectionResult, "inspection_grade": in.InspectionGrade, "defect_code": in.DefectCode,
		"state": model.MoveCompleted, "track_out_at": now, "process_seconds": secondsSince(move.TrackInAt, now),
		"to_node_key": to, "edge_key": result.EdgeKey, "resolve_action": result.Action, "resolve_reason": result.Reason,
		"operator_id": in.OperatorID,
	}).Error
}

func findLotByNo(db *gorm.DB, lotNo string) (*model.WipLot, error) {
	var row model.WipLot
	err := db.Where("lot_no = ?", strings.TrimSpace(lotNo)).First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipNotFound
		}
		return nil, err
	}
	return &row, nil
}

func findNode(g routegraph.Graph, key string) (routegraph.Node, bool) {
	for _, n := range g.Nodes {
		if n.Key == key {
			return n, true
		}
	}
	return routegraph.Node{}, false
}

func inspectionRequired(g routegraph.Graph, key string) bool {
	for _, e := range g.Edges {
		if e.From != key || e.IsDefault || e.Condition == nil {
			continue
		}
		preds := append([]routegraph.Predicate{}, e.Condition.All...)
		preds = append(preds, e.Condition.Any...)
		for _, p := range preds {
			if strings.HasPrefix(p.Field, "inspection.") || strings.HasPrefix(p.Field, "defect.") {
				return true
			}
		}
	}
	return false
}

func allowedEquipment(db *gorm.DB, group string, operationID, recipeID uint64) ([]model.EqpEquipment, error) {
	q := db.Model(&model.EqpEquipment{})
	if group != "" {
		q = q.Where("equipment_group = ?", group)
	}
	var rows []model.EqpEquipment
	if err := q.Order("equipment_code").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]model.EqpEquipment, 0, len(rows))
	for i := range rows {
		if canTrackIn(db, &rows[i], group, operationID, recipeID) == nil {
			out = append(out, rows[i])
		}
	}
	return out, nil
}

func loadEquipment(db *gorm.DB, id uint64) (*model.EqpEquipment, error) {
	var row model.EqpEquipment
	err := db.First(&row, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipEquipment
		}
		return nil, err
	}
	return &row, nil
}

func resolveRecipe(db *gorm.DB, node routegraph.Node, recipeID uint64) (uint64, error) {
	if recipeID == 0 {
		return node.RecipeID, nil
	}
	var recipe model.BaseRecipe
	if err := db.First(&recipe, recipeID).Error; err != nil {
		return 0, ErrWipState
	}
	if node.OperationID > 0 && uint64(recipe.OperationID) != node.OperationID {
		return 0, ErrWipState
	}
	return recipeID, nil
}

func openMove(db *gorm.DB, lotID uint64) (*model.WipMove, error) {
	var row model.WipMove
	err := db.Where("lot_id = ? AND state = ?", lotID, model.MoveOpen).Order("id desc").First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func secondsSince(from *time.Time, to time.Time) int {
	if from == nil {
		return 0
	}
	sec := int(to.Sub(*from).Seconds())
	if sec < 0 {
		return 0
	}
	return sec
}
