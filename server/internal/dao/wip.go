package dao

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
	"semi-mes/server/internal/routegraph"
)

// Business errors for work orders and lots.
var (
	ErrWipNotFound  = errors.New("wip not found")
	ErrWipState     = errors.New("wip bad state")
	ErrWipQty       = errors.New("wip quantity")
	ErrWipVersion   = errors.New("wip route version")
	ErrWipEquipment = errors.New("wip equipment")
	ErrWipInspect   = errors.New("wip inspection required")
	ErrWipTrack     = errors.New("wip track state")
)

// Column is one sponge-style list filter.
type Column struct {
	Name  string
	Exp   string
	Value string
}

// WorkOrderInput is the create and update payload.
type WorkOrderInput struct {
	OrderNo        string
	ProductID      uint64
	RouteVersionID uint64
	PlannedQty     int
	Priority       int
	DueDate        string
	Note           string
}

// WorkOrderView is a work order plus the names shown in the list.
type WorkOrderView struct {
	model.WipWorkOrder
	ProductCode string `json:"productCode"`
	ProductName string `json:"productName"`
	RouteCode   string `json:"routeCode"`
	RouteName   string `json:"routeName"`
	VersionNo   int    `json:"versionNo"`
	DueDateText string `json:"dueDateText"`
}

// ReleasedRoute is a published route version that a work order can bind.
type ReleasedRoute struct {
	VersionID   uint64 `json:"versionID"`
	RouteID     uint64 `json:"routeID"`
	VersionNo   int    `json:"versionNo"`
	RouteCode   string `json:"routeCode"`
	RouteName   string `json:"routeName"`
	ProductID   uint64 `json:"productID"`
	ProductCode string `json:"productCode"`
	ProductName string `json:"productName"`
}

// LotView adds the current node name and the work order number.
type LotView struct {
	model.WipLot
	OrderNo   string `json:"orderNo"`
	NodeName  string `json:"nodeName"`
	RouteCode string `json:"routeCode"`
	VersionNo int    `json:"versionNo"`
}

// LotLinkView is one genealogy edge with the other lot number.
type LotLinkView struct {
	model.WipLotLink
	ParentLotNo string `json:"parentLotNo"`
	ChildLotNo  string `json:"childLotNo"`
}

// AdvanceInput is the inspection context supplied when a lot leaves its step.
type AdvanceInput struct {
	InspectionResult string
	InspectionGrade  string
	DefectCode       string
}

// ListWorkOrders returns a page of work orders.
func ListWorkOrders(db *gorm.DB, page, limit int, sort string, columns []Column) ([]WorkOrderView, int64, error) {
	q := db.Table("wip_work_order o").
		Select("o.*, p.product_code, p.product_name, r.route_code, r.route_name, v.version_no").
		Joins("LEFT JOIN base_product p ON p.id = o.product_id").
		Joins("LEFT JOIN base_route_version v ON v.id = o.route_version_id").
		Joins("LEFT JOIN base_process_route r ON r.id = v.route_id").
		Where("o.deleted_at IS NULL")
	q = applyColumns(q, columns, map[string]string{
		"order_no": "o.order_no", "status": "o.status", "product_id": "o.product_id",
		"product_code": "p.product_code", "due_date": "o.due_date", "created_at": "o.created_at",
	})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []WorkOrderView
	err := q.Order(orderClause(sort, map[string]string{
		"id": "o.id", "order_no": "o.order_no", "status": "o.status", "due_date": "o.due_date", "created_at": "o.created_at",
	}, "o.id DESC")).
		Offset(page * limit).Limit(limit).Scan(&rows).Error
	for i := range rows {
		rows[i].DueDateText = formatDate(rows[i].DueDate)
	}
	return rows, total, err
}

// ListReleasedRoutes returns versions a new work order may bind.
func ListReleasedRoutes(db *gorm.DB) ([]ReleasedRoute, error) {
	var rows []ReleasedRoute
	err := db.Table("base_route_version v").
		Select("v.id AS version_id, v.route_id, v.version_no, r.route_code, r.route_name, r.product_id, p.product_code, p.product_name").
		Joins("JOIN base_process_route r ON r.id = v.route_id").
		Joins("JOIN base_product p ON p.id = r.product_id").
		Where("v.state = ? AND v.deleted_at IS NULL", routegraph.StateReleased).
		Order("r.route_code, v.version_no desc").
		Scan(&rows).Error
	return rows, err
}

// CreateWorkOrder inserts a created order bound to a released route version of the same product.
func CreateWorkOrder(db *gorm.DB, in WorkOrderInput) (*model.WipWorkOrder, error) {
	if strings.TrimSpace(in.OrderNo) == "" {
		number, err := Allocate(db, model.RuleWorkOrder, time.Now())
		if err != nil {
			return nil, err
		}
		in.OrderNo = number
	}
	if err := validateOrderInput(db, in, 0); err != nil {
		return nil, err
	}
	due, err := parseDate(in.DueDate)
	if err != nil {
		return nil, ErrWipState
	}
	row := &model.WipWorkOrder{
		OrderNo: in.OrderNo, ProductID: in.ProductID, RouteVersionID: in.RouteVersionID,
		PlannedQty: in.PlannedQty, Priority: normalizePriority(in.Priority), DueDate: due,
		Status: model.OrderCreated, Note: in.Note,
	}
	if err := db.Create(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}

// UpdateWorkOrder edits a created order. Later statuses only accept a note.
func UpdateWorkOrder(db *gorm.DB, id uint64, in WorkOrderInput) error {
	row, err := getOrder(db, id)
	if err != nil {
		return err
	}
	if row.Status != model.OrderCreated {
		return db.Model(row).Update("note", in.Note).Error
	}
	if err := validateOrderInput(db, in, id); err != nil {
		return err
	}
	due, err := parseDate(in.DueDate)
	if err != nil {
		return ErrWipState
	}
	return db.Model(row).Updates(map[string]any{
		"order_no": in.OrderNo, "product_id": in.ProductID, "route_version_id": in.RouteVersionID,
		"planned_qty": in.PlannedQty, "priority": normalizePriority(in.Priority), "due_date": due, "note": in.Note,
	}).Error
}

// DeleteWorkOrder removes an order that has not been released.
func DeleteWorkOrder(db *gorm.DB, id uint64) error {
	row, err := getOrder(db, id)
	if err != nil {
		return err
	}
	if row.Status != model.OrderCreated {
		return ErrWipState
	}
	return db.Delete(row).Error
}

// ReleaseWorkOrder moves created to released.
func ReleaseWorkOrder(db *gorm.DB, id uint64) (*model.WipWorkOrder, error) {
	row, err := getOrder(db, id)
	if err != nil {
		return nil, err
	}
	if row.Status != model.OrderCreated {
		return nil, ErrWipState
	}
	if _, err := releasedVersion(db, row.RouteVersionID, row.ProductID); err != nil {
		return nil, err
	}
	row.Status = model.OrderReleased
	if err := db.Model(row).Update("status", row.Status).Error; err != nil {
		return nil, err
	}
	return row, nil
}

// CloseWorkOrder moves completed to closed.
func CloseWorkOrder(db *gorm.DB, id uint64) (*model.WipWorkOrder, error) {
	row, err := getOrder(db, id)
	if err != nil {
		return nil, err
	}
	if row.Status != model.OrderCompleted {
		return nil, ErrWipState
	}
	row.Status = model.OrderClosed
	if err := db.Model(row).Update("status", row.Status).Error; err != nil {
		return nil, err
	}
	return row, nil
}

// StartLot creates the next lot number and places it on the route start node.
func StartLot(db *gorm.DB, orderID uint64, qty int, lotType string) (*model.WipLot, error) {
	if qty <= 0 {
		return nil, ErrWipQty
	}
	if lotType != model.LotTypeProduction && lotType != model.LotTypeEngineering {
		return nil, ErrWipState
	}
	var created *model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		order, err := getOrder(tx, orderID)
		if err != nil {
			return err
		}
		if order.Status != model.OrderReleased && order.Status != model.OrderInProgress {
			return ErrWipState
		}
		if order.ReleasedQty+qty > order.PlannedQty {
			return ErrWipQty
		}
		if _, err := releasedVersion(tx, order.RouteVersionID, order.ProductID); err != nil {
			return err
		}
		graph, err := LoadRouteGraph(tx, order.RouteVersionID)
		if err != nil {
			return err
		}
		start := ""
		for _, n := range graph.Nodes {
			if n.Type == routegraph.NodeStart {
				start = n.Key
				break
			}
		}
		if start == "" {
			return ErrWipVersion
		}
		now := time.Now()
		lotNo, err := Allocate(tx, model.RuleLot, now)
		if err != nil {
			return err
		}
		order.NextLotSeq++
		lot := &model.WipLot{
			LotNo: lotNo, OrderID: order.ID,
			ProductID: order.ProductID, RouteVersionID: order.RouteVersionID, CurrentNodeKey: start,
			Quantity: qty, Priority: order.Priority, LotType: lotType, Status: model.LotWaiting, ReworkJSON: "{}",
			ArrivedAt: &now,
		}
		if err := tx.Create(lot).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: model.EventStart, ToNodeKey: start, Quantity: qty,
		}).Error; err != nil {
			return err
		}
		if err := createLotWafers(tx, lot, now); err != nil {
			return err
		}
		status := model.OrderInProgress
		if err := tx.Model(order).Updates(map[string]any{
			"next_lot_seq": order.NextLotSeq, "released_qty": order.ReleasedQty + qty, "status": status,
		}).Error; err != nil {
			return err
		}
		created = lot
		return nil
	})
	return created, err
}

// ListLots returns a page of lots.
func ListLots(db *gorm.DB, page, limit int, sort string, columns []Column) ([]LotView, int64, error) {
	q := db.Table("wip_lot l").
		Select("l.*, o.order_no, n.name AS node_name, r.route_code, v.version_no").
		Joins("LEFT JOIN wip_work_order o ON o.id = l.order_id").
		Joins("LEFT JOIN base_route_version v ON v.id = l.route_version_id").
		Joins("LEFT JOIN base_process_route r ON r.id = v.route_id").
		Joins("LEFT JOIN base_route_node n ON n.version_id = l.route_version_id AND n.node_key = l.current_node_key AND n.deleted_at IS NULL").
		Joins("LEFT JOIN base_product p ON p.id = l.product_id").
		Where("l.deleted_at IS NULL")
	q = applyColumns(q, columns, map[string]string{
		"lot_no": "l.lot_no", "status": "l.status", "order_id": "l.order_id",
		"product_id": "l.product_id", "product_code": "p.product_code", "created_at": "l.created_at",
		"hold_reason_code": "l.hold_reason_code",
	})
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []LotView
	err := q.Order(orderClause(sort, map[string]string{
		"id": "l.id", "lot_no": "l.lot_no", "status": "l.status", "created_at": "l.created_at",
	}, "l.id DESC")).
		Offset(page * limit).Limit(limit).Scan(&rows).Error
	return rows, total, err
}

// GetLot loads one lot.
func GetLot(db *gorm.DB, id uint64) (*model.WipLot, error) {
	var row model.WipLot
	err := db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWipNotFound
	}
	return &row, err
}

// LotHistory returns events newest last.
func LotHistory(db *gorm.DB, lotID uint64) ([]model.WipLotHistory, error) {
	var rows []model.WipLotHistory
	err := db.Where("lot_id = ?", lotID).Order("id asc").Find(&rows).Error
	return rows, err
}

// LotLinks returns split and merge edges touching this lot.
func LotLinks(db *gorm.DB, lotID uint64) ([]LotLinkView, error) {
	var rows []LotLinkView
	err := db.Table("wip_lot_link k").
		Select("k.*, p.lot_no AS parent_lot_no, c.lot_no AS child_lot_no").
		Joins("JOIN wip_lot p ON p.id = k.parent_lot_id").
		Joins("JOIN wip_lot c ON c.id = k.child_lot_id").
		Where("k.parent_lot_id = ? OR k.child_lot_id = ?", lotID, lotID).
		Order("k.id asc").Scan(&rows).Error
	return rows, err
}

// HoldLot freezes a waiting lot.
func HoldLot(db *gorm.DB, id uint64, reasonCode, reason string) (*model.WipLot, error) {
	if !AcceptCode(db, "hold", reasonCode, nil) {
		return nil, ErrWipState
	}
	var updated *model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		lot, err := GetLot(tx, id)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting {
			return ErrWipState
		}
		lot.Status = model.LotHold
		lot.HoldReasonCode = reasonCode
		lot.HoldReason = reason
		if err := tx.Model(lot).Updates(map[string]any{
			"status": lot.Status, "hold_reason_code": reasonCode, "hold_reason": reason,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: model.EventHold, FromNodeKey: lot.CurrentNodeKey, ToNodeKey: lot.CurrentNodeKey,
			ReasonCode: reasonCode, Reason: reason, Quantity: lot.Quantity,
		}).Error; err != nil {
			return err
		}
		if err := notifyLot(tx, lot, model.NoticeLotHold); err != nil {
			return err
		}
		updated = lot
		return nil
	})
	return updated, err
}

// ReleaseHoldLot returns a held lot to waiting without moving it.
func ReleaseHoldLot(db *gorm.DB, id uint64, reasonCode, reason string) (*model.WipLot, error) {
	var updated *model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		lot, err := GetLot(tx, id)
		if err != nil {
			return err
		}
		if lot.Status != model.LotHold {
			return ErrWipState
		}
		if strings.TrimSpace(reasonCode) != "" && !AcceptCode(tx, "release", reasonCode, nil) {
			return ErrWipState
		}
		if err := tx.Model(lot).Updates(map[string]any{
			"status": model.LotWaiting, "hold_reason_code": "", "hold_reason": "",
		}).Error; err != nil {
			return err
		}
		lot.Status = model.LotWaiting
		lot.HoldReasonCode = ""
		lot.HoldReason = ""
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: model.EventRelease, FromNodeKey: lot.CurrentNodeKey, ToNodeKey: lot.CurrentNodeKey,
			ReasonCode: reasonCode, Reason: reason, Quantity: lot.Quantity,
		}).Error; err != nil {
			return err
		}
		updated = lot
		return nil
	})
	return updated, err
}

// SplitLot keeps the remainder on the parent and opens a child lot for each quantity.
func SplitLot(db *gorm.DB, id uint64, quantities []int) (*model.WipLot, []model.WipLot, error) {
	return splitLot(db, id, quantities, nil)
}

func SplitLotByWafer(db *gorm.DB, id uint64, groups [][]uint64) (*model.WipLot, []model.WipLot, error) {
	quantities := make([]int, len(groups))
	for i, group := range groups {
		quantities[i] = len(group)
	}
	return splitLot(db, id, quantities, groups)
}

func splitLot(db *gorm.DB, id uint64, quantities []int, groups [][]uint64) (*model.WipLot, []model.WipLot, error) {
	if len(quantities) == 0 {
		return nil, nil, ErrWipQty
	}
	sum := 0
	for _, q := range quantities {
		if q <= 0 {
			return nil, nil, ErrWipQty
		}
		sum += q
	}
	var parent *model.WipLot
	var children []model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		lot, err := GetLot(tx, id)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting {
			return ErrWipState
		}
		if sum >= lot.Quantity {
			return ErrWipQty
		}
		order, err := getOrder(tx, lot.OrderID)
		if err != nil {
			return err
		}
		lot.Quantity -= sum
		if err := tx.Model(lot).Update("quantity", lot.Quantity).Error; err != nil {
			return err
		}
		for i, q := range quantities {
			childNo, err := Allocate(tx, model.RuleLot, time.Now())
			if err != nil {
				return err
			}
			order.NextLotSeq++
			child := model.WipLot{
				LotNo: childNo, OrderID: lot.OrderID,
				ProductID: lot.ProductID, RouteVersionID: lot.RouteVersionID, CurrentNodeKey: lot.CurrentNodeKey,
				Quantity: q, Priority: lot.Priority, LotType: lot.LotType, Status: model.LotWaiting, ReworkJSON: lot.ReworkJSON,
			}
			if err := tx.Create(&child).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotLink{
				ParentLotID: lot.ID, ChildLotID: child.ID, LinkType: model.LinkSplit, Quantity: q,
			}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: lot.ID, EventType: model.EventSplit, FromNodeKey: lot.CurrentNodeKey, ToNodeKey: lot.CurrentNodeKey,
				Quantity: q, RelatedLotID: child.ID,
			}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: child.ID, EventType: model.EventSplit, ToNodeKey: child.CurrentNodeKey, Quantity: q, RelatedLotID: lot.ID,
			}).Error; err != nil {
				return err
			}
			var ids []uint64
			if groups != nil {
				ids = groups[i]
			}
			if err := moveSplitWafers(tx, lot, &child, q, ids); err != nil {
				return err
			}
			children = append(children, child)
		}
		if err := finishSplitSlots(tx, lot.ID); err != nil {
			return err
		}
		if err := tx.Model(order).Update("next_lot_seq", order.NextLotSeq).Error; err != nil {
			return err
		}
		parent = lot
		return nil
	})
	return parent, children, err
}

// MergeLots adds source quantities into the target and marks sources merged.
func MergeLots(db *gorm.DB, targetID uint64, sourceIDs []uint64) (*model.WipLot, error) {
	if len(sourceIDs) == 0 {
		return nil, ErrWipQty
	}
	var target *model.WipLot
	err := db.Transaction(func(tx *gorm.DB) error {
		lot, err := GetLot(tx, targetID)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting || lot.Quantity <= 0 {
			return ErrWipState
		}
		for _, sid := range sourceIDs {
			if sid == targetID {
				return ErrWipState
			}
			src, err := GetLot(tx, sid)
			if err != nil {
				return err
			}
			if src.Status != model.LotWaiting || src.OrderID != lot.OrderID || src.RouteVersionID != lot.RouteVersionID || src.CurrentNodeKey != lot.CurrentNodeKey || src.Quantity <= 0 {
				return ErrWipState
			}
			moved := src.Quantity
			if err := tx.Model(src).Updates(map[string]any{"status": model.LotMerged, "quantity": 0}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotLink{
				ParentLotID: src.ID, ChildLotID: lot.ID, LinkType: model.LinkMerge, Quantity: moved,
			}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: src.ID, EventType: model.EventMerge, FromNodeKey: src.CurrentNodeKey, Quantity: moved, RelatedLotID: lot.ID,
			}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: lot.ID, EventType: model.EventMerge, ToNodeKey: lot.CurrentNodeKey, Quantity: moved, RelatedLotID: src.ID,
			}).Error; err != nil {
				return err
			}
			if err := moveMergeWafers(tx, src.ID, lot.ID); err != nil {
				return err
			}
			lot.Quantity += moved
		}
		if err := tx.Model(lot).Update("quantity", lot.Quantity).Error; err != nil {
			return err
		}
		target = lot
		return nil
	})
	return target, err
}

// AdvanceLot asks the route resolver for the next node and records the move, hold, or completion.
func AdvanceLot(db *gorm.DB, id uint64, in AdvanceInput) (*model.WipLot, routegraph.Result, error) {
	var updated *model.WipLot
	var result routegraph.Result
	err := db.Transaction(func(tx *gorm.DB) error {
		lot, err := GetLot(tx, id)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting {
			return ErrWipState
		}
		result, err = moveLot(tx, lot, in, model.EventAdvance)
		if err != nil {
			return err
		}
		updated = lot
		return nil
	})
	return updated, result, err
}

// moveLot resolves the next node and updates the lot. The caller owns the transaction and the status check.
func moveLot(tx *gorm.DB, lot *model.WipLot, in AdvanceInput, eventName string) (routegraph.Result, error) {
	graph, err := LoadRouteGraph(tx, lot.RouteVersionID)
	if err != nil {
		return routegraph.Result{}, err
	}
	var product model.BaseProduct
	if err := tx.First(&product, lot.ProductID).Error; err != nil {
		return routegraph.Result{}, err
	}
	ctx := routegraph.LotContext{
		InspectionResult: in.InspectionResult,
		InspectionGrade:  in.InspectionGrade,
		DefectCode:       in.DefectCode,
		ProductCode:      product.ProductCode,
		Priority:         float64(lot.Priority),
		LotType:          lot.LotType,
		ReworkCounts:     decodeRework(lot.ReworkJSON),
	}
	result, err := routegraph.Resolve(graph, lot.CurrentNodeKey, ctx)
	if err != nil {
		if errors.Is(err, routegraph.ErrNodeNotFound) || errors.Is(err, routegraph.ErrNoPath) {
			return routegraph.Result{}, ErrWipVersion
		}
		return routegraph.Result{}, err
	}
	return applyResolved(tx, lot, graph, lot.CurrentNodeKey, ctx, result, eventName, "")
}

// reworkLot moves the lot along the route rework edge, or holds it when the max count is reached.
// A lot sitting on the operation just before a decision uses that decision's rework edge.
func reworkLot(tx *gorm.DB, lot *model.WipLot, defectCode, note string) error {
	graph, err := LoadRouteGraph(tx, lot.RouteVersionID)
	if err != nil {
		return err
	}
	var product model.BaseProduct
	if err := tx.First(&product, lot.ProductID).Error; err != nil {
		return err
	}
	ctx := routegraph.LotContext{
		InspectionResult: "fail",
		DefectCode:       defectCode,
		ProductCode:      product.ProductCode,
		Priority:         float64(lot.Priority),
		LotType:          lot.LotType,
		ReworkCounts:     decodeRework(lot.ReworkJSON),
	}
	result, ok := pickRework(graph, lot.CurrentNodeKey, ctx)
	if !ok {
		return ErrQc
	}
	_, err = applyResolved(tx, lot, graph, lot.CurrentNodeKey, ctx, result, "rework", note)
	return err
}

func pickRework(graph routegraph.Graph, from string, ctx routegraph.LotContext) (routegraph.Result, bool) {
	if result, err := routegraph.Resolve(graph, from, ctx); err == nil && reworkResult(graph, result) {
		return result, true
	}
	decision, ok := defaultDecision(graph, from)
	if !ok {
		return routegraph.Result{}, false
	}
	hopped, err := routegraph.Resolve(graph, decision, ctx)
	if err != nil || !reworkResult(graph, hopped) {
		return routegraph.Result{}, false
	}
	return hopped, true
}

func defaultDecision(graph routegraph.Graph, from string) (string, bool) {
	var to string
	n := 0
	for _, edge := range graph.Edges {
		if edge.From == from && edge.IsDefault {
			to = edge.To
			n++
		}
	}
	if n != 1 {
		return "", false
	}
	for _, node := range graph.Nodes {
		if node.Key == to && node.Type == routegraph.NodeDecision {
			return to, true
		}
	}
	return "", false
}

func reworkResult(graph routegraph.Graph, result routegraph.Result) bool {
	if result.Action == routegraph.ActionHold && result.Reason == routegraph.ReasonReworkExceeded && result.EdgeKey != "" {
		return true
	}
	if result.Action != routegraph.ActionMove {
		return false
	}
	for _, edge := range graph.Edges {
		if edge.Key == result.EdgeKey && edge.Kind == routegraph.EdgeRework {
			return true
		}
	}
	return false
}

func applyResolved(tx *gorm.DB, lot *model.WipLot, graph routegraph.Graph, from string, ctx routegraph.LotContext, result routegraph.Result, eventName, note string) (routegraph.Result, error) {
	switch result.Action {
	case routegraph.ActionHold:
		lot.Status = model.LotHold
		lot.HoldReasonCode = "REWORK_LIMIT"
		lot.HoldReason = result.Reason
		if err := tx.Model(lot).Updates(map[string]any{
			"status": lot.Status, "hold_reason_code": lot.HoldReasonCode, "hold_reason": lot.HoldReason, "quantity": lot.Quantity,
		}).Error; err != nil {
			return result, err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: model.EventHold, FromNodeKey: from, ToNodeKey: from, EdgeKey: result.EdgeKey,
			ReasonCode: lot.HoldReasonCode, Reason: result.Reason, Quantity: lot.Quantity,
		}).Error; err != nil {
			return result, err
		}
		kind := model.NoticeLotHold
		if result.Reason == routegraph.ReasonReworkExceeded {
			kind = model.NoticeReworkExceeded
		}
		return result, notifyLot(tx, lot, kind)
	case routegraph.ActionEnd:
		return result, completeLot(tx, lot, from, result)
	case routegraph.ActionMove:
		counts := ctx.ReworkCounts
		if counts == nil {
			counts = map[string]int{}
		}
		for _, e := range graph.Edges {
			if e.Key == result.EdgeKey && e.Kind == routegraph.EdgeRework {
				counts[e.Key]++
				break
			}
		}
		now := time.Now()
		lot.ReworkJSON = encodeRework(counts)
		lot.CurrentNodeKey = result.NextNodeKey
		lot.ArrivedAt = &now
		event := eventName
		if result.NextNodeType == routegraph.NodeEnd {
			event = model.EventComplete
			lot.Status = model.LotCompleted
		} else if lot.Status != model.LotWaiting {
			lot.Status = model.LotWaiting
		}
		if err := tx.Model(lot).Updates(map[string]any{
			"current_node_key": lot.CurrentNodeKey, "rework_json": lot.ReworkJSON, "status": lot.Status,
			"quantity": lot.Quantity, "arrived_at": now,
		}).Error; err != nil {
			return result, err
		}
		reason := result.Reason
		if note != "" {
			reason = note
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: event, FromNodeKey: from, ToNodeKey: lot.CurrentNodeKey, EdgeKey: result.EdgeKey,
			Reason: reason, Quantity: lot.Quantity,
		}).Error; err != nil {
			return result, err
		}
		if lot.Status == model.LotCompleted {
			return result, refreshOrder(tx, lot.OrderID)
		}
		return result, nil
	default:
		return result, ErrWipVersion
	}
}

func completeLot(tx *gorm.DB, lot *model.WipLot, from string, result routegraph.Result) error {
	lot.Status = model.LotCompleted
	if err := tx.Model(lot).Updates(map[string]any{"status": lot.Status, "quantity": lot.Quantity}).Error; err != nil {
		return err
	}
	if err := tx.Create(&model.WipLotHistory{
		LotID: lot.ID, EventType: model.EventComplete, FromNodeKey: from, ToNodeKey: from, Reason: result.Reason, Quantity: lot.Quantity,
	}).Error; err != nil {
		return err
	}
	if err := completeLotWafers(tx, lot.ID); err != nil {
		return err
	}
	return refreshOrder(tx, lot.OrderID)
}

func refreshOrder(tx *gorm.DB, orderID uint64) error {
	order, err := getOrder(tx, orderID)
	if err != nil {
		return err
	}
	if order.Status == model.OrderClosed {
		return nil
	}
	var completed int
	if err := tx.Model(&model.WipLot{}).Where("order_id = ? AND status = ?", orderID, model.LotCompleted).
		Select("COALESCE(SUM(quantity),0)").Scan(&completed).Error; err != nil {
		return err
	}
	var active int64
	if err := tx.Model(&model.WipLot{}).Where("order_id = ? AND status IN ?", orderID, []string{model.LotWaiting, model.LotRunning, model.LotHold}).
		Count(&active).Error; err != nil {
		return err
	}
	status := model.OrderInProgress
	if active == 0 && order.ReleasedQty >= order.PlannedQty && order.ReleasedQty > 0 {
		status = model.OrderCompleted
	}
	return tx.Model(order).Updates(map[string]any{"completed_qty": completed, "status": status}).Error
}

func validateOrderInput(db *gorm.DB, in WorkOrderInput, selfID uint64) error {
	if strings.TrimSpace(in.OrderNo) == "" || len(in.OrderNo) > 40 || in.PlannedQty <= 0 {
		return ErrWipQty
	}
	var product model.BaseProduct
	if err := db.First(&product, in.ProductID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrWipNotFound
		}
		return err
	}
	if _, err := releasedVersion(db, in.RouteVersionID, in.ProductID); err != nil {
		return err
	}
	var existing model.WipWorkOrder
	err := db.Where("order_no = ?", in.OrderNo).First(&existing).Error
	if err == nil && existing.ID != selfID {
		return ErrWipState
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func releasedVersion(db *gorm.DB, versionID, productID uint64) (*model.BaseRouteVersion, error) {
	var ver model.BaseRouteVersion
	if err := db.First(&ver, versionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipVersion
		}
		return nil, err
	}
	if ver.State != routegraph.StateReleased {
		return nil, ErrWipVersion
	}
	var route model.BaseProcessRoute
	if err := db.First(&route, ver.RouteID).Error; err != nil {
		return nil, err
	}
	if uint64(route.ProductID) != productID {
		return nil, ErrWipVersion
	}
	return &ver, nil
}

func getOrder(db *gorm.DB, id uint64) (*model.WipWorkOrder, error) {
	var row model.WipWorkOrder
	err := db.First(&row, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrWipNotFound
	}
	return &row, err
}

func normalizePriority(p int) int {
	if p < 1 || p > 4 {
		return 2
	}
	return p
}

func parseDate(s string) (*time.Time, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func formatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func decodeRework(raw string) map[string]int {
	out := map[string]int{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	if out == nil {
		return map[string]int{}
	}
	return out
}

func encodeRework(m map[string]int) string {
	if m == nil {
		return "{}"
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func applyColumns(q *gorm.DB, columns []Column, allow map[string]string) *gorm.DB {
	for _, col := range columns {
		field, ok := allow[col.Name]
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(col.Exp)) {
		case "like":
			if strings.TrimSpace(col.Value) == "" {
				continue
			}
			value := col.Value
			if !strings.Contains(value, "%") {
				value = "%" + value + "%"
			}
			q = q.Where(field+" LIKE ?", value)
		case "in":
			vals := splitFilter(col.Value)
			if len(vals) == 0 {
				continue
			}
			q = q.Where(field+" IN ?", vals)
		case "gte", ">=":
			if col.Value == "" {
				continue
			}
			q = q.Where(field+" >= ?", col.Value)
		case "lte", "<=":
			if col.Value == "" {
				continue
			}
			q = q.Where(field+" <= ?", col.Value)
		case "gt", ">":
			if col.Value == "" {
				continue
			}
			q = q.Where(field+" > ?", col.Value)
		case "lt", "<":
			if col.Value == "" {
				continue
			}
			q = q.Where(field+" < ?", col.Value)
		case "neq", "!=":
			if col.Value == "" {
				continue
			}
			q = q.Where(field+" <> ?", col.Value)
		case "isnull":
			q = q.Where(field + " IS NULL")
		default:
			if col.Value == "" {
				continue
			}
			q = q.Where(field+" = ?", col.Value)
		}
	}
	return q
}

// pullColumn removes the first column with name and returns its value.
func pullColumn(columns []Column, name string) ([]Column, string, bool) {
	rest := make([]Column, 0, len(columns))
	value := ""
	found := false
	for _, col := range columns {
		if !found && col.Name == name {
			value = col.Value
			found = true
			continue
		}
		rest = append(rest, col)
	}
	return rest, value, found
}

func splitFilter(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func orderClause(sort string, allow map[string]string, fallback string) string {
	desc := strings.HasPrefix(sort, "-")
	name := strings.TrimPrefix(sort, "-")
	field, ok := allow[name]
	if !ok {
		return fallback
	}
	if desc {
		return field + " DESC"
	}
	return field + " ASC"
}
