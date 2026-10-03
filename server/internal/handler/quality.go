package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/config"
	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/model"
)

type itemBody struct {
	ParamCode  string   `json:"paramCode"`
	ParamName  string   `json:"paramName"`
	Unit       string   `json:"unit"`
	Target     *float64 `json:"target"`
	LSL        *float64 `json:"lsl"`
	USL        *float64 `json:"usl"`
	LCL        *float64 `json:"lcl"`
	UCL        *float64 `json:"ucl"`
	SampleSize int      `json:"sampleSize"`
	Required   bool     `json:"required"`
}

type planBody struct {
	OperationID uint64     `json:"operationID"`
	ProductID   uint64     `json:"productID"`
	PlanName    string     `json:"planName"`
	Enabled     bool       `json:"enabled"`
	Items       []itemBody `json:"items"`
}

type measureBody struct {
	LotID       uint64       `json:"lotId"`
	EquipmentID uint64       `json:"equipmentId"`
	OperationID uint64       `json:"operationId"`
	Samples     []sampleBody `json:"samples"`
}

type defectCodeBody struct {
	DefectCode string `json:"defectCode"`
	DefectName string `json:"defectName"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Status     int    `json:"status"`
}

type defectBody struct {
	LotID       uint64 `json:"lotId"`
	EquipmentID uint64 `json:"equipmentId"`
	OperationID uint64 `json:"operationId"`
	NodeKey     string `json:"nodeKey"`
	DefectCode  string `json:"defectCode"`
	Quantity    int    `json:"quantity"`
	Disposition string `json:"disposition"`
	ReasonCode  string `json:"reasonCode"`
	Note        string `json:"note"`
}

type policyBody struct {
	ParamCode   string `json:"paramCode"`
	OperationID uint64 `json:"operationID"`
	OnOOC       string `json:"onOOC"`
	OnOOS       string `json:"onOOS"`
}

type limitBody struct {
	ParamCode   string   `json:"paramCode"`
	OperationID uint64   `json:"operationID"`
	ProductID   uint64   `json:"productID"`
	EquipmentID uint64   `json:"equipmentID"`
	ChartType   string   `json:"chartType"`
	Center      *float64 `json:"center"`
	LCL         *float64 `json:"lcl"`
	UCL         *float64 `json:"ucl"`
	UseManual   bool     `json:"useManual"`
}

func qcPathID(c *gin.Context) uint64 {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return id
}

func planInput(body planBody) dao.PlanInput {
	items := make([]dao.ItemInput, 0, len(body.Items))
	for _, item := range body.Items {
		items = append(items, dao.ItemInput{
			ParamCode: item.ParamCode, ParamName: item.ParamName, Unit: item.Unit,
			Target: item.Target, LSL: item.LSL, USL: item.USL, LCL: item.LCL, UCL: item.UCL,
			SampleSize: item.SampleSize, Required: item.Required,
		})
	}
	return dao.PlanInput{OperationID: body.OperationID, ProductID: body.ProductID, PlanName: body.PlanName, Enabled: body.Enabled, Items: items}
}

// ListInspectPlans returns measurement plans.
func ListInspectPlans(c *gin.Context) {
	rows, _, err := dao.ListInspectPlans(database.GetDB(), 0, 500, nil)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"plans": rows})
}

// ListInspectPlanPage returns a filtered page of measurement plans.
func ListInspectPlanPage(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	page, limit, _, cols := body.query()
	rows, total, err := dao.ListInspectPlans(database.GetDB(), page, limit, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"plans": rows, "total": total})
}

// SaveInspectPlan creates or replaces a plan.
func SaveInspectPlan(c *gin.Context) {
	body := planBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.SaveInspectPlan(database.GetDB(), qcPathID(c), planInput(body))
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

// DeleteInspectPlan soft-deletes a plan.
func DeleteInspectPlan(c *gin.Context) {
	if err := dao.DeleteInspectPlan(database.GetDB(), qcPathID(c)); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": qcPathID(c)})
}

// RecordMeasurement stores readings for a waiting or running lot.
func RecordMeasurement(c *gin.Context) {
	body := measureBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	samples := make([]dao.SampleInput, 0, len(body.Samples))
	for _, sample := range body.Samples {
		samples = append(samples, dao.SampleInput{ParamCode: sample.ParamCode, Values: sample.Values})
	}
	result, err := dao.RecordMeasurements(database.GetDB(), dao.MeasureInput{
		LotID: body.LotID, EquipmentID: body.EquipmentID, OperationID: body.OperationID, OperatorID: op, Samples: samples,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"result": result})
}

// ListDefectCodes returns the defect master.
func ListDefectCodes(c *gin.Context) {
	rows, _, err := dao.ListDefectCodes(database.GetDB(), 0, 500, nil)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"codes": rows})
}

// ListDefectCodePage returns a filtered page of defect codes.
func ListDefectCodePage(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	page, limit, _, cols := body.query()
	rows, total, err := dao.ListDefectCodes(database.GetDB(), page, limit, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"codes": rows, "total": total})
}

// SaveDefectCode creates or updates a defect code.
func SaveDefectCode(c *gin.Context) {
	body := defectCodeBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if body.Status == 0 {
		body.Status = 1
	}
	row, err := dao.SaveDefectCode(database.GetDB(), qcPathID(c), body.DefectCode, body.DefectName, body.Category, body.Severity, body.Status)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

// ListDefects returns recent defect records.
func ListDefects(c *gin.Context) {
	lotID, _ := strconv.ParseUint(c.Query("lotId"), 10, 64)
	rows, _, err := dao.ListDefects(database.GetDB(), lotID, 0, 100, nil)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"defects": rows})
}

// ListDefectPage returns a filtered page of defect records.
func ListDefectPage(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	page, limit, _, cols := body.query()
	rows, total, err := dao.ListDefects(database.GetDB(), 0, page, limit, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"defects": rows, "total": total})
}

// RecordDefect stores a defect and applies the disposition.
func RecordDefect(c *gin.Context) {
	body := defectBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	row, err := dao.RecordDefect(database.GetDB(), dao.DefectInput{
		LotID: body.LotID, EquipmentID: body.EquipmentID, OperationID: body.OperationID, NodeKey: body.NodeKey,
		DefectCode: body.DefectCode, Quantity: body.Quantity, Disposition: body.Disposition, ReasonCode: body.ReasonCode, Note: body.Note, OperatorID: op,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

// DefectPareto sums defect quantities.
func DefectPareto(c *gin.Context) {
	productID, _ := strconv.ParseUint(c.Query("productId"), 10, 64)
	operationID, _ := strconv.ParseUint(c.Query("operationId"), 10, 64)
	from, to := parseRange(c.Query("from"), c.Query("to"))
	rows, err := dao.DefectPareto(database.GetDB(), productID, operationID, from, to)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"rows": rows})
}

// SpcChart returns an I-MR or X-bar chart.
func SpcChart(c *gin.Context) {
	if spcDisabled(c) {
		return
	}
	productID, _ := strconv.ParseUint(c.Query("productId"), 10, 64)
	operationID, _ := strconv.ParseUint(c.Query("operationId"), 10, 64)
	equipmentID, _ := strconv.ParseUint(c.Query("equipmentId"), 10, 64)
	from, to := parseRange(c.Query("from"), c.Query("to"))
	view, err := dao.LoadChart(database.GetDB(), dao.ChartQuery{
		ProductID: productID, OperationID: operationID, EquipmentID: equipmentID,
		ParamCode: c.Query("param"), From: from, To: to,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, view)
}

// ListSpcPolicies returns OOC reactions.
func ListSpcPolicies(c *gin.Context) {
	if spcDisabled(c) {
		return
	}
	rows, err := dao.ListSpcPolicies(database.GetDB())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"policies": rows})
}

// SaveSpcPolicy upserts a reaction.
func SaveSpcPolicy(c *gin.Context) {
	if spcDisabled(c) {
		return
	}
	body := policyBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.SaveSpcPolicy(database.GetDB(), model.QcSpcPolicy{
		ParamCode: body.ParamCode, OperationID: body.OperationID, OnOOC: body.OnOOC, OnOOS: body.OnOOS,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

// SaveSpcLimit upserts a manual control limit.
func SaveSpcLimit(c *gin.Context) {
	if spcDisabled(c) {
		return
	}
	body := limitBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row, err := dao.SaveSpcLimit(database.GetDB(), model.QcSpcLimit{
		ParamCode: body.ParamCode, OperationID: body.OperationID, ProductID: body.ProductID, EquipmentID: body.EquipmentID,
		ChartType: body.ChartType, Center: body.Center, LCL: body.LCL, UCL: body.UCL, UseManual: body.UseManual,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, row)
}

func spcDisabled(c *gin.Context) bool {
	if config.SPCEnabled() {
		return false
	}
	fail(c, 40003, "error.auth.forbidden")
	return true
}

func parseRange(from, to string) (time.Time, time.Time) {
	return parseTime(from), parseTime(to)
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
