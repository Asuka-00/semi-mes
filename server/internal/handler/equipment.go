package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
)

type equipmentBody struct {
	EquipmentCode  string `json:"equipmentCode"`
	EquipmentName  string `json:"equipmentName"`
	EquipmentGroup string `json:"equipmentGroup"`
	EquipmentType  string `json:"equipmentType"`
	Status         string `json:"status"`
	LineID         uint64 `json:"lineID"`
	ModelName      string `json:"modelName"`
	Manufacturer   string `json:"manufacturer"`
	SerialNo       string `json:"serialNo"`
	Location       string `json:"location"`
	ChamberCount   int    `json:"chamberCount"`
	Capacity       int    `json:"capacity"`
	InstallDate    string `json:"installDate"`
	Capabilities   []struct {
		OperationID uint64 `json:"operationID"`
		RecipeID    uint64 `json:"recipeID"`
	} `json:"capabilities"`
}

func (b equipmentBody) input() dao.EquipmentInput {
	caps := make([]dao.CapabilityInput, 0, len(b.Capabilities))
	for _, cap := range b.Capabilities {
		caps = append(caps, dao.CapabilityInput{OperationID: cap.OperationID, RecipeID: cap.RecipeID})
	}
	return dao.EquipmentInput{
		EquipmentCode: b.EquipmentCode, EquipmentName: b.EquipmentName, EquipmentGroup: b.EquipmentGroup,
		EquipmentType: b.EquipmentType, Status: b.Status, LineID: b.LineID, ModelName: b.ModelName,
		Manufacturer: b.Manufacturer, SerialNo: b.SerialNo, Location: b.Location, ChamberCount: b.ChamberCount,
		Capacity: b.Capacity, InstallDate: b.InstallDate, Capabilities: caps,
	}
}

type stateBody struct {
	ToState    string `json:"toState"`
	ReasonCode string `json:"reasonCode"`
	Reason     string `json:"reason"`
}

type pmPlanBody struct {
	EquipmentID    uint64   `json:"equipmentID"`
	EquipmentGroup string   `json:"equipmentGroup"`
	PlanName       string   `json:"planName"`
	TriggerType    string   `json:"triggerType"`
	IntervalDays   int      `json:"intervalDays"`
	IntervalCount  int      `json:"intervalCount"`
	Checklist      []string `json:"checklist"`
	Enabled        *bool    `json:"enabled"`
	BlockTrackIn   bool     `json:"blockTrackIn"`
	NextDueAt      string   `json:"nextDueAt"`
}

func (b pmPlanBody) input() dao.PmPlanInput {
	enabled := true
	if b.Enabled != nil {
		enabled = *b.Enabled
	}
	return dao.PmPlanInput{
		EquipmentID: b.EquipmentID, EquipmentGroup: b.EquipmentGroup, PlanName: b.PlanName, TriggerType: b.TriggerType,
		IntervalDays: b.IntervalDays, IntervalCount: b.IntervalCount, Checklist: b.Checklist, Enabled: enabled,
		BlockTrackIn: b.BlockTrackIn, NextDueAt: b.NextDueAt,
	}
}

type pmCompleteBody struct {
	Items  []dao.CheckItem `json:"items"`
	Result string          `json:"result"`
	Note   string          `json:"note"`
}

type pmListBody struct {
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	Sort        string `json:"sort"`
	Status      string `json:"status"`
	EquipmentID uint64 `json:"equipmentID"`
	Columns     []struct {
		Name  string `json:"name"`
		Exp   string `json:"exp"`
		Value any    `json:"value"`
	} `json:"columns"`
}

func CreateEquipment(c *gin.Context) {
	body := equipmentBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.CreateEquipment(database.GetDB(), body.input())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID})
}

func UpdateEquipment(c *gin.Context) {
	id := eqpPathID(c)
	if id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := equipmentBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.UpdateEquipment(database.GetDB(), id, body.input())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

func DeleteEquipment(c *gin.Context) {
	id := eqpPathID(c)
	if id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.DeleteEquipment(database.GetDB(), id); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func ListEquipmentPage(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListEquipmentPage(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"eqpEquipments": rows, "total": total})
}

func GetEquipment(c *gin.Context) {
	id := eqpPathID(c)
	if id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	detail, err := dao.GetEquipmentDetail(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, detail)
}

func ChangeEquipmentState(c *gin.Context) {
	id := eqpPathID(c)
	body := stateBody{}
	if id == 0 || c.ShouldBindJSON(&body) != nil || body.ToState == "" {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	row, err := dao.ChangeEquipmentState(database.GetDB(), id, body.ToState, body.ReasonCode, body.Reason, op)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

func EquipmentBoard(c *gin.Context) {
	board, err := dao.LoadEquipmentBoard(database.GetDB())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, board)
}

func SavePmPlan(c *gin.Context) {
	body := pmPlanBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	row, err := dao.SavePmPlan(database.GetDB(), id, body.input())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID, "plan": row})
}

func ListPmPlans(c *gin.Context) {
	rows, _, err := dao.ListPmPlans(database.GetDB(), 0, 500, nil)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"plans": rows})
}

func ListPmPlanPage(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	page, limit, _, cols := body.query()
	rows, total, err := dao.ListPmPlans(database.GetDB(), page, limit, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"plans": rows, "total": total})
}

func DeletePmPlan(c *gin.Context) {
	id := eqpPathID(c)
	if id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.DeletePmPlan(database.GetDB(), id); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func ListPmTasks(c *gin.Context) {
	body := pmListBody{}
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	cols := make([]dao.Column, 0, len(body.Columns)+2)
	hasStatus := false
	hasEqp := false
	for _, col := range body.Columns {
		if col.Name == "status" {
			hasStatus = true
		}
		if col.Name == "equipment_id" {
			hasEqp = true
		}
		cols = append(cols, dao.Column{Name: col.Name, Exp: col.Exp, Value: stringify(col.Value)})
	}
	if body.Status != "" && !hasStatus {
		cols = append(cols, dao.Column{Name: "status", Exp: "=", Value: body.Status})
	}
	if body.EquipmentID > 0 && !hasEqp {
		cols = append(cols, dao.Column{Name: "equipment_id", Exp: "=", Value: strconv.FormatUint(body.EquipmentID, 10)})
	}
	page, limit := body.Page, body.Limit
	if page < 0 {
		page = 0
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 2000 {
		limit = 2000
	}
	rows, total, err := dao.ListPmTasks(database.GetDB(), page, limit, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"tasks": rows, "total": total})
}

func GeneratePmTasks(c *gin.Context) {
	if err := dao.EnsurePmTasks(database.GetDB()); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"at": time.Now()})
}

func StartPmTask(c *gin.Context) {
	id := eqpPathID(c)
	if id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	row, err := dao.StartPm(database.GetDB(), id, op)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

func CompletePmTask(c *gin.Context) {
	id := eqpPathID(c)
	body := pmCompleteBody{}
	if id == 0 || c.ShouldBindJSON(&body) != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	row, err := dao.CompletePm(database.GetDB(), id, op, body.Items, body.Result, body.Note)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

func eqpPathID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return id
}
