package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/i18n"
	"semi-mes/server/internal/model"
	"semi-mes/server/internal/routegraph"
)

type wipListBody struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	Sort    string `json:"sort"`
	Columns []struct {
		Name  string `json:"name"`
		Exp   string `json:"exp"`
		Value any    `json:"value"`
	} `json:"columns"`
}

type workOrderBody struct {
	OrderNo        string `json:"orderNo"`
	ProductID      uint64 `json:"productID"`
	RouteVersionID uint64 `json:"routeVersionID"`
	PlannedQty     int    `json:"plannedQty"`
	Priority       int    `json:"priority"`
	DueDate        string `json:"dueDate"`
	Note           string `json:"note"`
}

type startLotBody struct {
	Quantity int    `json:"quantity"`
	LotType  string `json:"lotType"`
}

type holdBody struct {
	ReasonCode string `json:"reasonCode"`
	Reason     string `json:"reason"`
}

type splitBody struct {
	Quantities []int `json:"quantities"`
}

type mergeBody struct {
	TargetID  uint64   `json:"targetId"`
	SourceIDs []uint64 `json:"sourceIds"`
}

type advanceBody struct {
	InspectionResult string `json:"inspectionResult"`
	InspectionGrade  string `json:"inspectionGrade"`
	DefectCode       string `json:"defectCode"`
}

func (b workOrderBody) input() dao.WorkOrderInput {
	return dao.WorkOrderInput{
		OrderNo: b.OrderNo, ProductID: b.ProductID, RouteVersionID: b.RouteVersionID,
		PlannedQty: b.PlannedQty, Priority: b.Priority, DueDate: b.DueDate, Note: b.Note,
	}
}

func (b wipListBody) query() (int, int, string, []dao.Column) {
	limit := b.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 2000 {
		limit = 2000
	}
	page := b.Page
	if page < 0 {
		page = 0
	}
	cols := make([]dao.Column, 0, len(b.Columns))
	for _, col := range b.Columns {
		cols = append(cols, dao.Column{Name: col.Name, Exp: col.Exp, Value: stringify(col.Value)})
	}
	return page, limit, b.Sort, cols
}

func stringify(v any) string {
	switch n := v.(type) {
	case string:
		return n
	case float64:
		return strconv.FormatInt(int64(n), 10)
	case []any:
		parts := make([]string, 0, len(n))
		for _, item := range n {
			if text := stringify(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, ",")
	default:
		return ""
	}
}

// ListWorkOrders returns a page of work orders.
func ListWorkOrders(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListWorkOrders(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if rows == nil {
		rows = []dao.WorkOrderView{}
	}
	response.Success(c, gin.H{"wipWorkOrders": rows, "total": total})
}

// ReleasedRoutes lists published route versions.
func ReleasedRoutes(c *gin.Context) {
	rows, err := dao.ListReleasedRoutes(database.GetDB())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if rows == nil {
		rows = []dao.ReleasedRoute{}
	}
	response.Success(c, gin.H{"versions": rows})
}

// CreateWorkOrder inserts a created order.
func CreateWorkOrder(c *gin.Context) {
	body := workOrderBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.CreateWorkOrder(database.GetDB(), body.input())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID})
}

// UpdateWorkOrder edits an order.
func UpdateWorkOrder(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := workOrderBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.UpdateWorkOrder(database.GetDB(), id, body.input()); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// DeleteWorkOrder removes a created order.
func DeleteWorkOrder(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.DeleteWorkOrder(database.GetDB(), id); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

// ReleaseWorkOrder publishes an order so lots can start.
func ReleaseWorkOrder(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.ReleaseWorkOrder(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID, "status": row.Status})
}

// CloseWorkOrder closes a completed order.
func CloseWorkOrder(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.CloseWorkOrder(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID, "status": row.Status})
}

// StartLot releases quantity from a work order onto the route start node.
func StartLot(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := startLotBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	lot, err := dao.StartLot(database.GetDB(), id, body.Quantity, body.LotType)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": lot.ID, "lotNo": lot.LotNo, "currentNodeKey": lot.CurrentNodeKey})
}

// ListLots returns a page of lots.
func ListLots(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListLots(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if rows == nil {
		rows = []dao.LotView{}
	}
	response.Success(c, gin.H{"wipLots": rows, "total": total})
}

// GetLot returns the lot, its history, genealogy, and the route graph.
func GetLot(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	lot, err := dao.GetLot(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	history, err := dao.LotHistory(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	links, err := dao.LotLinks(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	graph, err := dao.LoadRouteGraph(database.GetDB(), lot.RouteVersionID)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if history == nil {
		history = []model.WipLotHistory{}
	}
	if links == nil {
		links = []dao.LotLinkView{}
	}
	moves, err := dao.ListMovesByLot(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if moves == nil {
		moves = []dao.MoveView{}
	}
	response.Success(c, gin.H{
		"lot": lot, "history": history, "links": links, "moves": moves,
		"nodes": nodesDTO(graph), "edges": edgesDTO(graph),
	})
}

// HoldLot freezes a waiting lot.
func HoldLot(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := holdBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	lot, err := dao.HoldLot(database.GetDB(), id, body.ReasonCode, body.Reason)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": lot.ID, "status": lot.Status})
}

// ReleaseHoldLot clears a hold and leaves the lot on the same node.
func ReleaseHoldLot(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := holdBody{}
	_ = c.ShouldBindJSON(&body)
	lot, err := dao.ReleaseHoldLot(database.GetDB(), id, body.ReasonCode, body.Reason)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": lot.ID, "status": lot.Status})
}

// SplitLot opens child lots and keeps the remainder on the parent.
func SplitLot(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := splitBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	parent, children, err := dao.SplitLot(database.GetDB(), id, body.Quantities)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": parent, "children": children})
}

// MergeLots folds source lots into the target.
func MergeLots(c *gin.Context) {
	body := mergeBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.TargetID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	lot, err := dao.MergeLots(database.GetDB(), body.TargetID, body.SourceIDs)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot})
}

// AdvanceLot moves a waiting lot with the route resolver.
func AdvanceLot(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := advanceBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	lot, result, err := dao.AdvanceLot(database.GetDB(), id, dao.AdvanceInput{
		InspectionResult: body.InspectionResult, InspectionGrade: body.InspectionGrade, DefectCode: body.DefectCode,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot, "result": result})
}

func writeWipErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, dao.ErrWipNotFound):
		c.JSON(200, gin.H{"code": 40004, "msg": i18n.T(c, "error.wip.not_found"), "data": struct{}{}})
	case errors.Is(err, dao.ErrWipState):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.wip.bad_state"), "data": struct{}{}})
	case errors.Is(err, dao.ErrWipQty):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.wip.qty"), "data": struct{}{}})
	case errors.Is(err, dao.ErrWipVersion):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.wip.version"), "data": struct{}{}})
	case errors.Is(err, dao.ErrWipEquipment):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.wip.equipment"), "data": struct{}{}})
	case errors.Is(err, dao.ErrEqpState):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.eqp.state"), "data": struct{}{}})
	case errors.Is(err, dao.ErrEqpPM):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.eqp.pm"), "data": struct{}{}})
	case errors.Is(err, dao.ErrWipInspect):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.wip.inspect"), "data": struct{}{}})
	case errors.Is(err, dao.ErrQc):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.qc.rejected"), "data": struct{}{}})
	case errors.Is(err, dao.ErrWipTrack):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.wip.track"), "data": struct{}{}})
	case errors.Is(err, routegraph.ErrNoPath):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.route.no_path"), "data": struct{}{}})
	default:
		c.JSON(200, gin.H{"code": 50000, "msg": i18n.T(c, "error.server.internal"), "data": struct{}{}})
	}
}
