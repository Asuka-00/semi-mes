package dao

import (
	"errors"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
)

// ErrQc is a rejected quality action: missing samples, unknown code, or a disposition the lot cannot take.
var ErrQc = errors.New("quality rejected")

// SampleInput is the readings for one parameter.
type SampleInput struct {
	ParamCode string
	Values    []float64
}

// PlanInput creates or replaces a plan and its items.
type PlanInput struct {
	OperationID uint64
	ProductID   uint64
	PlanName    string
	Enabled     bool
	Items       []ItemInput
}

// ItemInput is one measurement parameter.
type ItemInput struct {
	ParamCode  string
	ParamName  string
	Unit       string
	Target     *float64
	LSL        *float64
	USL        *float64
	LCL        *float64
	UCL        *float64
	SampleSize int
	Required   bool
}

// PlanView is a plan plus items and the operation code.
type PlanView struct {
	model.QcInspectPlan
	OperationCode string                `json:"operationCode"`
	Items         []model.QcInspectItem `json:"items"`
}

// MeasureInput records readings for a lot against the matching plan.
type MeasureInput struct {
	LotID       uint64
	MoveID      uint64
	EquipmentID uint64
	OperationID uint64
	OperatorID  uint64
	Samples     []SampleInput
}

// DefectInput records a defect and applies its disposition.
type DefectInput struct {
	LotID       uint64
	EquipmentID uint64
	OperationID uint64
	NodeKey     string
	DefectCode  string
	Quantity    int
	Disposition string
	Note        string
	OperatorID  uint64
}

// ChartQuery filters an SPC chart.
type ChartQuery struct {
	ProductID   uint64
	OperationID uint64
	EquipmentID uint64
	ParamCode   string
	From        time.Time
	To          time.Time
}

// ChartPoint is one plotted statistic.
type ChartPoint struct {
	Index      int       `json:"index"`
	Value      float64   `json:"value"`
	At         time.Time `json:"at"`
	LotID      uint64    `json:"lotID"`
	Violations []int     `json:"violations"`
	OOS        bool      `json:"oos"`
}

// ChartView is an X-bar/R or I-MR chart.
type ChartView struct {
	ParamCode string             `json:"paramCode"`
	ChartType string             `json:"chartType"`
	Center    float64            `json:"center"`
	LCL       float64            `json:"lcl"`
	UCL       float64            `json:"ucl"`
	Points    []ChartPoint       `json:"points"`
	Events    []model.QcSpcEvent `json:"events"`
}

var (
	a2 = map[int]float64{2: 1.880, 3: 1.023, 4: 0.729, 5: 0.577, 6: 0.483, 7: 0.419, 8: 0.373, 9: 0.337, 10: 0.308}
)

// SaveInspectPlan creates or updates a plan and replaces its items.
func SaveInspectPlan(db *gorm.DB, id uint64, in PlanInput) (*model.QcInspectPlan, error) {
	if in.OperationID == 0 || strings.TrimSpace(in.PlanName) == "" || len(in.Items) == 0 {
		return nil, ErrQc
	}
	for _, item := range in.Items {
		if strings.TrimSpace(item.ParamCode) == "" {
			return nil, ErrQc
		}
	}
	var row *model.QcInspectPlan
	err := db.Transaction(func(tx *gorm.DB) error {
		if id == 0 {
			row = &model.QcInspectPlan{
				OperationID: in.OperationID, ProductID: in.ProductID, PlanName: strings.TrimSpace(in.PlanName), Enabled: in.Enabled,
			}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		} else {
			current := &model.QcInspectPlan{}
			if err := tx.First(current, id).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrWipNotFound
				}
				return err
			}
			if err := tx.Model(current).Updates(map[string]any{
				"operation_id": in.OperationID, "product_id": in.ProductID, "plan_name": strings.TrimSpace(in.PlanName), "enabled": in.Enabled,
			}).Error; err != nil {
				return err
			}
			row = current
		}
		if err := tx.Where("plan_id = ?", row.ID).Delete(&model.QcInspectItem{}).Error; err != nil {
			return err
		}
		for i, item := range in.Items {
			size := item.SampleSize
			if size <= 0 {
				size = 1
			}
			next := model.QcInspectItem{
				PlanID: row.ID, ParamCode: strings.TrimSpace(item.ParamCode), ParamName: item.ParamName, Unit: item.Unit,
				Target: item.Target, LSL: item.LSL, USL: item.USL, LCL: item.LCL, UCL: item.UCL,
				SampleSize: size, Required: item.Required, SortOrder: i,
			}
			if err := tx.Create(&next).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return row, err
}

// ListInspectPlans returns enabled and disabled plans.
func ListInspectPlans(db *gorm.DB) ([]PlanView, error) {
	var plans []model.QcInspectPlan
	if err := db.Order("id").Find(&plans).Error; err != nil {
		return nil, err
	}
	out := make([]PlanView, 0, len(plans))
	for _, plan := range plans {
		view, err := planView(db, &plan)
		if err != nil {
			return nil, err
		}
		out = append(out, *view)
	}
	return out, nil
}

// DeleteInspectPlan soft-deletes a plan and its items.
func DeleteInspectPlan(db *gorm.DB, id uint64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("plan_id = ?", id).Delete(&model.QcInspectItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.QcInspectPlan{}, id).Error
	})
}

// FindInspectPlan prefers a product-specific plan over the operation-wide plan.
func FindInspectPlan(db *gorm.DB, operationID, productID uint64) (*PlanView, error) {
	if operationID == 0 {
		return nil, nil
	}
	var specific model.QcInspectPlan
	err := db.Where("enabled = ? AND operation_id = ? AND product_id = ?", true, operationID, productID).Order("id desc").First(&specific).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		return planView(db, &specific)
	}
	var general model.QcInspectPlan
	err = db.Where("enabled = ? AND operation_id = ? AND product_id = 0", true, operationID).Order("id desc").First(&general).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return planView(db, &general)
}

func planView(db *gorm.DB, plan *model.QcInspectPlan) (*PlanView, error) {
	view := &PlanView{QcInspectPlan: *plan, Items: []model.QcInspectItem{}}
	if err := db.Where("plan_id = ?", plan.ID).Order("sort_order, id").Find(&view.Items).Error; err != nil {
		return nil, err
	}
	var op model.BaseOperation
	if err := db.First(&op, plan.OperationID).Error; err == nil {
		view.OperationCode = op.OperationCode
	}
	return view, nil
}

// RecordMeasurements judges the samples, stores them, and runs SPC on the new points.
func RecordMeasurements(db *gorm.DB, in MeasureInput) (string, error) {
	if in.LotID == 0 {
		return "", ErrQc
	}
	var result string
	err := db.Transaction(func(tx *gorm.DB) error {
		judged, err := recordMeasurementsTx(tx, in)
		result = judged
		return err
	})
	return result, err
}

// recordMeasurementsTx is the same write, on a transaction the caller already opened.
func recordMeasurementsTx(tx *gorm.DB, in MeasureInput) (string, error) {
	lot, err := GetLot(tx, in.LotID)
	if err != nil {
		return "", err
	}
	opID := in.OperationID
	if opID == 0 {
		return "", ErrQc
	}
	plan, err := FindInspectPlan(tx, opID, lot.ProductID)
	if err != nil {
		return "", err
	}
	if plan == nil {
		return "", ErrQc
	}
	judged, rows, err := judgePlan(plan, in)
	if err != nil {
		return "", err
	}
	now := time.Now()
	for i := range rows {
		rows[i].LotID = lot.ID
		rows[i].MoveID = in.MoveID
		rows[i].PlanID = plan.ID
		rows[i].OperationID = opID
		rows[i].EquipmentID = in.EquipmentID
		rows[i].ProductID = lot.ProductID
		rows[i].OperatorID = in.OperatorID
		rows[i].MeasuredAt = now
		if err := tx.Create(&rows[i]).Error; err != nil {
			return "", err
		}
	}
	if err := tx.Create(&model.QcJudgement{LotID: lot.ID, OperationID: opID, Result: judged}).Error; err != nil {
		return "", err
	}
	lastByParam := map[string]model.QcMeasurement{}
	for _, row := range rows {
		lastByParam[row.ParamCode] = row
	}
	for _, row := range lastByParam {
		if err := evaluateParam(tx, lot, in.EquipmentID, opID, row); err != nil {
			return "", err
		}
	}
	return judged, nil
}

func judgePlan(plan *PlanView, in MeasureInput) (string, []model.QcMeasurement, error) {
	byCode := map[string][]float64{}
	for _, sample := range in.Samples {
		byCode[strings.TrimSpace(sample.ParamCode)] = sample.Values
	}
	overall := model.SpecPass
	var rows []model.QcMeasurement
	for _, item := range plan.Items {
		values := byCode[item.ParamCode]
		if item.Required && len(values) < item.SampleSize {
			return "", nil, ErrQc
		}
		if len(values) == 0 {
			continue
		}
		for i, value := range values {
			spec := specOf(item, value)
			if item.Required && spec == model.SpecFail {
				overall = model.SpecFail
			}
			rows = append(rows, model.QcMeasurement{
				ItemID: item.ID, ParamCode: item.ParamCode, SampleNo: i + 1, Value: value, SpecResult: spec,
			})
		}
	}
	if len(rows) == 0 {
		return "", nil, ErrQc
	}
	return overall, rows, nil
}

func specOf(item model.QcInspectItem, value float64) string {
	if item.LSL != nil && value < *item.LSL {
		return model.SpecFail
	}
	if item.USL != nil && value > *item.USL {
		return model.SpecFail
	}
	if item.LSL == nil && item.USL == nil {
		return "na"
	}
	return model.SpecPass
}

// LatestJudgement returns the newest spec result for a lot, or empty when none exists.
func LatestJudgement(db *gorm.DB, lotID uint64) (string, error) {
	var row model.QcJudgement
	err := db.Where("lot_id = ?", lotID).Order("id desc").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return row.Result, nil
}

func evaluateParam(tx *gorm.DB, lot *model.WipLot, equipmentID, operationID uint64, newest model.QcMeasurement) error {
	var history []model.QcMeasurement
	q := tx.Where("param_code = ? AND operation_id = ?", newest.ParamCode, operationID).Order("id")
	if equipmentID > 0 {
		q = q.Where("equipment_id = ?", equipmentID)
	}
	if err := q.Find(&history).Error; err != nil {
		return err
	}
	manual, err := manualLimit(tx, newest.ParamCode, operationID, equipmentID)
	if err != nil {
		return err
	}
	points, _, _, _, _ := buildChart(history, manual)
	if len(points) == 0 {
		return nil
	}
	last := points[len(points)-1]
	oos := newest.SpecResult == model.SpecFail
	var rules []int
	for _, rule := range last.Violations {
		rules = append(rules, rule)
	}
	if len(rules) == 0 && !oos {
		return nil
	}
	policy, err := findPolicy(tx, newest.ParamCode, operationID)
	if err != nil {
		return err
	}
	reaction := model.ReactNone
	kind := model.SpcOOC
	ruleNo := 0
	if len(rules) > 0 {
		ruleNo = rules[0]
		reaction = policy.OnOOC
	}
	if oos {
		kind = model.SpcOOS
		if policy.OnOOS != "" && policy.OnOOS != model.ReactNone {
			reaction = policy.OnOOS
		}
	}
	event := model.QcSpcEvent{
		ParamCode: newest.ParamCode, OperationID: operationID, EquipmentID: equipmentID, ProductID: lot.ProductID,
		LotID: lot.ID, MeasurementID: newest.ID, Rule: ruleNo, Kind: kind, Value: newest.Value, Reaction: reaction,
	}
	if err := tx.Create(&event).Error; err != nil {
		return err
	}
	return applyReaction(tx, lot, equipmentID, reaction)
}

func findPolicy(db *gorm.DB, param string, operationID uint64) (model.QcSpcPolicy, error) {
	var row model.QcSpcPolicy
	err := db.Where("param_code = ? AND operation_id = ?", param, operationID).Order("id desc").First(&row).Error
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return row, err
	}
	err = db.Where("param_code = ? AND operation_id = 0", param).Order("id desc").First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.QcSpcPolicy{OnOOC: model.ReactNone, OnOOS: model.ReactNone}, nil
	}
	return row, err
}

func applyReaction(tx *gorm.DB, lot *model.WipLot, equipmentID uint64, reaction string) error {
	if reaction == "" || reaction == model.ReactNone {
		return nil
	}
	hold := reaction == model.ReactHold || reaction == model.ReactHoldEngineering || reaction == model.ReactHoldDown
	eng := reaction == model.ReactEngineering || reaction == model.ReactHoldEngineering
	down := reaction == model.ReactDown || reaction == model.ReactHoldDown
	if hold && lot.Status == model.LotWaiting {
		lot.Status = model.LotHold
		lot.HoldReasonCode = "OOC"
		lot.HoldReason = "SPC"
		if err := tx.Model(lot).Updates(map[string]any{
			"status": lot.Status, "hold_reason_code": lot.HoldReasonCode, "hold_reason": lot.HoldReason,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&model.WipLotHistory{
			LotID: lot.ID, EventType: model.EventHold, FromNodeKey: lot.CurrentNodeKey, ToNodeKey: lot.CurrentNodeKey,
			ReasonCode: "OOC", Reason: "SPC", Quantity: lot.Quantity,
		}).Error; err != nil {
			return err
		}
	}
	if equipmentID == 0 || (!eng && !down) {
		return nil
	}
	eqp, err := loadEquipment(tx, equipmentID)
	if err != nil {
		return err
	}
	target := model.EqpEngineering
	if down {
		target = model.EqpUnscheduledDown
	}
	if NormalizeEqpState(eqp.Status) == target {
		return nil
	}
	return applyState(tx, eqp, target, "OOC", reaction, 0, true)
}

// LoadChart builds the chart for the filter. Limits are manual when configured, otherwise from the points.
func LoadChart(db *gorm.DB, q ChartQuery) (*ChartView, error) {
	if strings.TrimSpace(q.ParamCode) == "" {
		return nil, ErrQc
	}
	query := db.Where("param_code = ?", q.ParamCode).Order("id")
	if q.OperationID > 0 {
		query = query.Where("operation_id = ?", q.OperationID)
	}
	if q.EquipmentID > 0 {
		query = query.Where("equipment_id = ?", q.EquipmentID)
	}
	if q.ProductID > 0 {
		query = query.Where("product_id = ?", q.ProductID)
	}
	if !q.From.IsZero() {
		query = query.Where("measured_at >= ?", q.From)
	}
	if !q.To.IsZero() {
		query = query.Where("measured_at <= ?", q.To)
	}
	var rows []model.QcMeasurement
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	manual, err := manualLimit(db, q.ParamCode, q.OperationID, q.EquipmentID)
	if err != nil {
		return nil, err
	}
	points, chartType, center, lcl, ucl := buildChart(rows, manual)
	events := []model.QcSpcEvent{}
	ev := db.Where("param_code = ?", q.ParamCode).Order("id desc").Limit(50)
	if q.OperationID > 0 {
		ev = ev.Where("operation_id = ?", q.OperationID)
	}
	if err := ev.Find(&events).Error; err != nil {
		return nil, err
	}
	if points == nil {
		points = []ChartPoint{}
	}
	return &ChartView{ParamCode: q.ParamCode, ChartType: chartType, Center: center, LCL: lcl, UCL: ucl, Points: points, Events: events}, nil
}

func buildChart(rows []model.QcMeasurement, manual *model.QcSpcLimit) ([]ChartPoint, string, float64, float64, float64) {
	if len(rows) == 0 {
		return nil, "imr", 0, 0, 0
	}
	sample := 1
	if rows[0].ItemID > 0 {
		// subgroup size is inferred from consecutive sample numbers when present
		maxSample := 1
		for _, row := range rows {
			if row.SampleNo > maxSample {
				maxSample = row.SampleNo
			}
		}
		if maxSample > 1 {
			sample = maxSample
		}
	}
	chartType := "imr"
	var values []float64
	var stamps []time.Time
	var lots []uint64
	var oos []bool
	if sample == 1 {
		for _, row := range rows {
			values = append(values, row.Value)
			stamps = append(stamps, row.MeasuredAt)
			lots = append(lots, row.LotID)
			oos = append(oos, row.SpecResult == model.SpecFail)
		}
	} else {
		chartType = "xbar"
		for i := 0; i+sample <= len(rows); i += sample {
			sum := 0.0
			failed := false
			for j := 0; j < sample; j++ {
				sum += rows[i+j].Value
				if rows[i+j].SpecResult == model.SpecFail {
					failed = true
				}
			}
			values = append(values, sum/float64(sample))
			stamps = append(stamps, rows[i+sample-1].MeasuredAt)
			lots = append(lots, rows[i].LotID)
			oos = append(oos, failed)
		}
	}
	center, lcl, ucl := limitsFor(values, sample, chartType)
	if manual != nil && manual.UseManual && manual.Center != nil && manual.LCL != nil && manual.UCL != nil {
		center, lcl, ucl = *manual.Center, *manual.LCL, *manual.UCL
		if manual.ChartType != "" {
			chartType = manual.ChartType
		}
	}
	points := make([]ChartPoint, len(values))
	for i, value := range values {
		points[i] = ChartPoint{
			Index: i + 1, Value: value, At: stamps[i], LotID: lots[i],
			Violations: rulesAt(values, i, center, lcl, ucl),
			OOS:        oos[i],
		}
	}
	return points, chartType, center, lcl, ucl
}

func limitsFor(values []float64, sample int, chartType string) (float64, float64, float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	mean := average(values)
	if chartType == "xbar" && sample >= 2 && sample <= 10 && len(values) >= 2 {
		ranges := make([]float64, 0, len(values))
		// ranges are not available from means alone; use a moving range of the subgroup means as a stand-in when raw subgroups were already reduced
		for i := 1; i < len(values); i++ {
			ranges = append(ranges, math.Abs(values[i]-values[i-1]))
		}
		rbar := average(ranges)
		factor := a2[sample]
		return mean, mean - factor*rbar, mean + factor*rbar
	}
	if len(values) < 2 {
		return mean, mean, mean
	}
	mr := 0.0
	for i := 1; i < len(values); i++ {
		mr += math.Abs(values[i] - values[i-1])
	}
	mrBar := mr / float64(len(values)-1)
	// I-MR individuals: E2 = 2.660
	return mean, mean - 2.660*mrBar, mean + 2.660*mrBar
}

func rulesAt(values []float64, index int, center, lcl, ucl float64) []int {
	var rules []int
	value := values[index]
	if value > ucl || value < lcl {
		rules = append(rules, 1)
	}
	if index >= 8 && sameSide(values[index-8:index+1], center) {
		rules = append(rules, 2)
	}
	if index >= 5 && monotonic(values[index-5:index+1]) {
		rules = append(rules, 3)
	}
	if index >= 13 && alternating(values[index-13:index+1]) {
		rules = append(rules, 4)
	}
	return rules
}

func sameSide(values []float64, center float64) bool {
	above := values[0] > center
	if values[0] == center {
		return false
	}
	for _, value := range values {
		if above && value <= center {
			return false
		}
		if !above && value >= center {
			return false
		}
	}
	return true
}

func monotonic(values []float64) bool {
	up, down := true, true
	for i := 1; i < len(values); i++ {
		if values[i] <= values[i-1] {
			up = false
		}
		if values[i] >= values[i-1] {
			down = false
		}
	}
	return up || down
}

func alternating(values []float64) bool {
	if len(values) < 2 {
		return false
	}
	prev := values[1] - values[0]
	if prev == 0 {
		return false
	}
	for i := 2; i < len(values); i++ {
		delta := values[i] - values[i-1]
		if delta == 0 || (prev > 0 && delta > 0) || (prev < 0 && delta < 0) {
			return false
		}
		prev = delta
	}
	return true
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// SaveDefectCode creates or updates a defect code.
func SaveDefectCode(db *gorm.DB, id uint64, code, name, category, severity string, status int) (*model.QcDefectCode, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, ErrQc
	}
	if severity == "" {
		severity = "minor"
	}
	if id == 0 {
		row := &model.QcDefectCode{DefectCode: code, DefectName: name, Category: category, Severity: severity, Status: status}
		if row.Status == 0 {
			row.Status = 1
		}
		return row, db.Create(row).Error
	}
	var row model.QcDefectCode
	if err := db.First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWipNotFound
		}
		return nil, err
	}
	if err := db.Model(&row).Updates(map[string]any{
		"defect_code": code, "defect_name": name, "category": category, "severity": severity, "status": status,
	}).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// ListDefectCodes returns the master.
func ListDefectCodes(db *gorm.DB) ([]model.QcDefectCode, error) {
	var rows []model.QcDefectCode
	err := db.Order("defect_code").Find(&rows).Error
	if rows == nil {
		rows = []model.QcDefectCode{}
	}
	return rows, err
}

// RecordDefect stores the defect and applies scrap, hold, rework, or use-as-is.
func RecordDefect(db *gorm.DB, in DefectInput) (*model.QcDefect, error) {
	if in.LotID == 0 || in.Quantity <= 0 || strings.TrimSpace(in.DefectCode) == "" {
		return nil, ErrQc
	}
	switch in.Disposition {
	case model.DispositionRework, model.DispositionScrap, model.DispositionUseAsIs, model.DispositionHold:
	default:
		return nil, ErrQc
	}
	var saved *model.QcDefect
	err := db.Transaction(func(tx *gorm.DB) error {
		var code model.QcDefectCode
		if err := tx.Where("defect_code = ?", in.DefectCode).First(&code).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrQc
			}
			return err
		}
		lot, err := GetLot(tx, in.LotID)
		if err != nil {
			return err
		}
		if lot.Status != model.LotWaiting && lot.Status != model.LotHold {
			return ErrQc
		}
		node := in.NodeKey
		if node == "" {
			node = lot.CurrentNodeKey
		}
		row := &model.QcDefect{
			LotID: lot.ID, NodeKey: node, OperationID: in.OperationID, EquipmentID: in.EquipmentID, DefectCode: in.DefectCode,
			Quantity: in.Quantity, Disposition: in.Disposition, Note: in.Note, OperatorID: in.OperatorID,
		}
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		switch in.Disposition {
		case model.DispositionScrap:
			if in.Quantity > lot.Quantity {
				return ErrQc
			}
			lot.Quantity -= in.Quantity
			status := lot.Status
			if lot.Quantity == 0 {
				status = model.LotScrapped
				lot.Status = status
			}
			if err := tx.Model(lot).Updates(map[string]any{"quantity": lot.Quantity, "status": status}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: lot.ID, EventType: model.EventScrap, FromNodeKey: node, ToNodeKey: node,
				ReasonCode: in.DefectCode, Quantity: in.Quantity,
			}).Error; err != nil {
				return err
			}
		case model.DispositionHold:
			lot.Status = model.LotHold
			lot.HoldReasonCode = in.DefectCode
			lot.HoldReason = in.Note
			if err := tx.Model(lot).Updates(map[string]any{
				"status": lot.Status, "hold_reason_code": lot.HoldReasonCode, "hold_reason": lot.HoldReason,
			}).Error; err != nil {
				return err
			}
			if err := tx.Create(&model.WipLotHistory{
				LotID: lot.ID, EventType: model.EventHold, FromNodeKey: node, ToNodeKey: node,
				ReasonCode: in.DefectCode, Reason: in.Note, Quantity: in.Quantity,
			}).Error; err != nil {
				return err
			}
		case model.DispositionRework:
			if err := tx.Create(&model.WipLotHistory{
				LotID: lot.ID, EventType: "rework", FromNodeKey: node, ToNodeKey: node,
				ReasonCode: in.DefectCode, Reason: in.Note, Quantity: in.Quantity,
			}).Error; err != nil {
				return err
			}
		default:
			if err := tx.Create(&model.WipLotHistory{
				LotID: lot.ID, EventType: "use_as_is", FromNodeKey: node, ToNodeKey: node,
				ReasonCode: in.DefectCode, Reason: in.Note, Quantity: in.Quantity,
			}).Error; err != nil {
				return err
			}
		}
		saved = row
		return nil
	})
	return saved, err
}

// DefectView is a defect row plus the lot number.
type DefectView struct {
	model.QcDefect
	LotNo string `json:"lotNo"`
}

// ListDefects returns recent defect records.
func ListDefects(db *gorm.DB, lotID uint64) ([]DefectView, error) {
	q := db.Table("qc_defect d").
		Select("d.*, l.lot_no").
		Joins("JOIN wip_lot l ON l.id = d.lot_id").
		Where("d.deleted_at IS NULL").
		Order("d.id desc").Limit(100)
	if lotID > 0 {
		q = q.Where("d.lot_id = ?", lotID)
	}
	var rows []DefectView
	err := q.Scan(&rows).Error
	if rows == nil {
		rows = []DefectView{}
	}
	return rows, err
}

func manualLimit(db *gorm.DB, param string, operationID, equipmentID uint64) (*model.QcSpcLimit, error) {
	var limit model.QcSpcLimit
	err := db.Where("param_code = ? AND use_manual = ? AND (operation_id = 0 OR operation_id = ?) AND (equipment_id = 0 OR equipment_id = ?)",
		param, true, operationID, equipmentID).Order("id desc").First(&limit).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &limit, nil
}

// ParetoRow is one bar on the defect Pareto.
type ParetoRow struct {
	DefectCode string `json:"defectCode"`
	DefectName string `json:"defectName"`
	Severity   string `json:"severity"`
	Quantity   int    `json:"quantity"`
}

// DefectPareto sums quantities by defect code.
func DefectPareto(db *gorm.DB, productID, operationID uint64, from, to time.Time) ([]ParetoRow, error) {
	q := db.Table("qc_defect d").
		Select("d.defect_code, c.defect_name, c.severity, SUM(d.quantity) AS quantity").
		Joins("LEFT JOIN qc_defect_code c ON c.defect_code = d.defect_code AND c.deleted_at IS NULL").
		Joins("JOIN wip_lot l ON l.id = d.lot_id").
		Where("d.deleted_at IS NULL")
	if productID > 0 {
		q = q.Where("l.product_id = ?", productID)
	}
	if operationID > 0 {
		q = q.Where("d.operation_id = ?", operationID)
	}
	if !from.IsZero() {
		q = q.Where("d.created_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("d.created_at <= ?", to)
	}
	var rows []ParetoRow
	err := q.Group("d.defect_code, c.defect_name, c.severity").Order("quantity desc").Scan(&rows).Error
	if rows == nil {
		rows = []ParetoRow{}
	}
	return rows, err
}

// SaveSpcPolicy upserts the reaction for a parameter.
func SaveSpcPolicy(db *gorm.DB, in model.QcSpcPolicy) (*model.QcSpcPolicy, error) {
	if strings.TrimSpace(in.ParamCode) == "" {
		return nil, ErrQc
	}
	if in.OnOOC == "" {
		in.OnOOC = model.ReactNone
	}
	if in.OnOOS == "" {
		in.OnOOS = model.ReactNone
	}
	var row model.QcSpcPolicy
	err := db.Where("param_code = ? AND operation_id = ?", in.ParamCode, in.OperationID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = in
		return &row, db.Create(&row).Error
	}
	if err != nil {
		return nil, err
	}
	if err := db.Model(&row).Updates(map[string]any{"on_ooc": in.OnOOC, "on_oos": in.OnOOS}).Error; err != nil {
		return nil, err
	}
	row.OnOOC = in.OnOOC
	row.OnOOS = in.OnOOS
	return &row, nil
}

// SaveSpcLimit upserts a manual control limit for a parameter and operation.
func SaveSpcLimit(db *gorm.DB, in model.QcSpcLimit) (*model.QcSpcLimit, error) {
	if strings.TrimSpace(in.ParamCode) == "" || in.Center == nil || in.LCL == nil || in.UCL == nil {
		return nil, ErrQc
	}
	var row model.QcSpcLimit
	err := db.Where("param_code = ? AND operation_id = ? AND equipment_id = ? AND product_id = ?",
		in.ParamCode, in.OperationID, in.EquipmentID, in.ProductID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row = in
		if row.ChartType == "" {
			row.ChartType = "imr"
		}
		return &row, db.Create(&row).Error
	}
	if err != nil {
		return nil, err
	}
	if err := db.Model(&row).Updates(map[string]any{
		"chart_type": in.ChartType, "center": in.Center, "lcl": in.LCL, "ucl": in.UCL, "use_manual": in.UseManual,
	}).Error; err != nil {
		return nil, err
	}
	row.ChartType = in.ChartType
	row.Center, row.LCL, row.UCL = in.Center, in.LCL, in.UCL
	row.UseManual = in.UseManual
	return &row, nil
}

// ListSpcPolicies returns configured reactions.
func ListSpcPolicies(db *gorm.DB) ([]model.QcSpcPolicy, error) {
	var rows []model.QcSpcPolicy
	err := db.Order("param_code").Find(&rows).Error
	if rows == nil {
		rows = []model.QcSpcPolicy{}
	}
	return rows, err
}
