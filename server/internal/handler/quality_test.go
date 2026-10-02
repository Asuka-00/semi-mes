package handler_test

import (
	"net/http"
	"testing"
)

func TestQualitySpecFeedsRework(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, eqp := doJSON(http.MethodGet, "/api/v1/eqpEquipment", token, nil)
	tools := map[string]uint64{}
	for _, row := range eqp["data"].(map[string]any)["equipment"].([]any) {
		item := row.(map[string]any)
		tools[item["equipmentCode"].(string)] = uint64(item["id"].(float64))
	}
	lotID, _ := startCleanLot(t, token, "WO-QC-JUDGE")
	track := func(code string, value float64) {
		t.Helper()
		_, in := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools[code]})
		if in["code"] != float64(0) {
			t.Fatalf("track in %s: %+v", code, in)
		}
		body := map[string]any{"lotId": lotID, "qtyOut": 4}
		if value > 0 {
			body["measurements"] = []map[string]any{{"paramCode": "CD", "values": []float64{value}}}
		}
		_, out := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, body)
		if out["code"] != float64(0) {
			t.Fatalf("track out %s: %+v", code, out)
		}
	}
	track("WET-01", 0)
	track("PHOTO-01", 0)
	_, in := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["METRO-01"]})
	if in["code"] != float64(0) {
		t.Fatalf("metro in: %+v", in)
	}
	_, missing := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 4, "inspectionResult": "pass"})
	if missing["code"] == float64(0) {
		t.Fatalf("metro without samples should fail: %+v", missing)
	}
	_, out := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{
		"lotId": lotID, "qtyOut": 4, "inspectionResult": "pass",
		"measurements": []map[string]any{{"paramCode": "CD", "values": []float64{600}}},
	})
	if out["code"] != float64(0) {
		t.Fatalf("metro out: %+v", out)
	}
	_, decided := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID})
	lot := decided["data"].(map[string]any)["lot"].(map[string]any)
	if decided["code"] != float64(0) || lot["currentNodeKey"] != "photo" {
		t.Fatalf("spec fail should rework: %+v", decided)
	}
}

func TestQualityRulesReactionDefectsPermissions(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, chart := doJSON(http.MethodGet, "/api/v1/qcSpc/chart?param=CD", token, nil)
	if chart["code"] != float64(0) {
		t.Fatalf("chart: %+v", chart)
	}
	points := chart["data"].(map[string]any)["points"].([]any)
	events := chart["data"].(map[string]any)["events"].([]any)
	sawRule1 := false
	for _, raw := range points {
		point := raw.(map[string]any)
		violations, _ := point["violations"].([]any)
		for _, rule := range violations {
			if rule == float64(1) {
				sawRule1 = true
			}
		}
	}
	if !sawRule1 || len(events) == 0 {
		t.Fatalf("expected an OOC point: %+v", chart["data"])
	}

	_, plans := doJSON(http.MethodGet, "/api/v1/qcInspectPlan", token, nil)
	var operationID uint64
	for _, raw := range plans["data"].(map[string]any)["plans"].([]any) {
		plan := raw.(map[string]any)
		operationID = uint64(plan["operationID"].(float64))
	}
	if operationID == 0 {
		t.Fatal("missing inspect plan")
	}

	ruleEqp := createEntity(t, token, "/api/v1/eqpEquipment", map[string]any{
		"equipmentCode": "EQP-RULE", "equipmentName": "规则台", "equipmentGroup": "METRO", "status": "standby", "capacity": 1,
	})
	lotID, _ := startCleanLot(t, token, "WO-QC-RULE")
	for i := 0; i < 9; i++ {
		_, measured := doJSON(http.MethodPost, "/api/v1/qcMeasurement", token, map[string]any{
			"lotId": lotID, "equipmentId": mustNum(ruleEqp), "operationId": operationID,
			"samples": []map[string]any{{"paramCode": "CD", "values": []float64{510}}},
		})
		if measured["code"] != float64(0) {
			t.Fatalf("sample %d: %+v", i, measured)
		}
	}
	_, detail := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotID), token, nil)
	if detail["data"].(map[string]any)["lot"].(map[string]any)["status"] != "hold" {
		t.Fatalf("rule 2 should hold: %+v", detail["data"].(map[string]any)["lot"])
	}

	_, _ = doJSON(http.MethodPut, "/api/v1/qcSpc/policy", token, map[string]any{
		"paramCode": "CD", "operationID": operationID, "onOOC": "hold_and_engineering", "onOOS": "hold_and_engineering",
	})
	t.Cleanup(func() {
		_, _ = doJSON(http.MethodPut, "/api/v1/qcSpc/policy", token, map[string]any{
			"paramCode": "CD", "operationID": operationID, "onOOC": "hold_lot", "onOOS": "hold_lot",
		})
	})
	reactEqp := createEntity(t, token, "/api/v1/eqpEquipment", map[string]any{
		"equipmentCode": "EQP-REACT", "equipmentName": "反应台", "equipmentGroup": "METRO", "status": "standby", "capacity": 1,
	})
	lotReact, _ := startCleanLot(t, token, "WO-QC-REACT")
	_, reacted := doJSON(http.MethodPost, "/api/v1/qcMeasurement", token, map[string]any{
		"lotId": lotReact, "equipmentId": mustNum(reactEqp), "operationId": operationID,
		"samples": []map[string]any{{"paramCode": "CD", "values": []float64{700}}},
	})
	if reacted["code"] != float64(0) {
		t.Fatalf("reaction measure: %+v", reacted)
	}
	_, held := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotReact), token, nil)
	if held["data"].(map[string]any)["lot"].(map[string]any)["status"] != "hold" {
		t.Fatalf("oos hold: %+v", held["data"].(map[string]any)["lot"])
	}
	_, eqp := doJSON(http.MethodGet, "/api/v1/eqpEquipment/"+reactEqp, token, nil)
	if eqp["data"].(map[string]any)["equipment"].(map[string]any)["status"] != "engineering" {
		t.Fatalf("equipment reaction: %+v", eqp["data"].(map[string]any)["equipment"])
	}

	lotScrap, _ := startCleanLot(t, token, "WO-QC-SCRAP")
	_, scrap := doJSON(http.MethodPost, "/api/v1/qcDefect", token, map[string]any{
		"lotId": lotScrap, "defectCode": "PARTICLE", "quantity": 1, "disposition": "scrap", "note": "颗粒",
	})
	if scrap["code"] != float64(0) {
		t.Fatalf("scrap: %+v", scrap)
	}
	_, after := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotScrap), token, nil)
	lot := after["data"].(map[string]any)["lot"].(map[string]any)
	history := after["data"].(map[string]any)["history"].([]any)
	if int(lot["quantity"].(float64)) != 3 || !hasEvent(history, "scrap") {
		t.Fatalf("scrap effect: %+v", after["data"])
	}
	_, rework := doJSON(http.MethodPost, "/api/v1/qcDefect", token, map[string]any{
		"lotId": lotScrap, "defectCode": "PATTERN", "quantity": 1, "disposition": "rework", "note": "图形",
	})
	if rework["code"] != float64(0) {
		t.Fatalf("rework: %+v", rework)
	}
	_, reworkLot := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotScrap), token, nil)
	reworkRow := reworkLot["data"].(map[string]any)["lot"].(map[string]any)
	if reworkRow["currentNodeKey"] != "clean" || reworkRow["status"] != "waiting" || !hasEvent(reworkLot["data"].(map[string]any)["history"].([]any), "rework") {
		t.Fatalf("rework stays on the node: %+v", reworkRow)
	}
	lotHold, _ := startCleanLot(t, token, "WO-QC-HOLD")
	_, heldDefect := doJSON(http.MethodPost, "/api/v1/qcDefect", token, map[string]any{
		"lotId": lotHold, "defectCode": "SCRATCH", "quantity": 1, "disposition": "hold", "note": "划伤",
	})
	if heldDefect["code"] != float64(0) {
		t.Fatalf("hold defect: %+v", heldDefect)
	}
	_, holdLot := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotHold), token, nil)
	if holdLot["data"].(map[string]any)["lot"].(map[string]any)["status"] != "hold" {
		t.Fatalf("defect hold: %+v", holdLot["data"].(map[string]any)["lot"])
	}

	_, pareto := doJSON(http.MethodGet, "/api/v1/qcDefect/pareto", token, nil)
	if pareto["code"] != float64(0) || len(pareto["data"].(map[string]any)["rows"].([]any)) == 0 {
		t.Fatalf("pareto: %+v", pareto)
	}
	_, user := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "viewerqc", "password": "viewer123", "realName": "Viewer", "status": 1,
	})
	userID := toID(user["data"].(map[string]any)["id"])
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{4}})
	viewer := login(t, "viewerqc", "viewer123")
	_, forbidden := doJSON(http.MethodPost, "/api/v1/qcDefect", viewer, map[string]any{
		"lotId": lotHold, "defectCode": "PARTICLE", "quantity": 1, "disposition": "use_as_is",
	})
	if forbidden["code"] != float64(40003) {
		t.Fatalf("viewer defect: %+v", forbidden)
	}
	_, allowed := doJSON(http.MethodGet, "/api/v1/qcSpc/chart?param=CD", viewer, nil)
	if allowed["code"] != float64(0) {
		t.Fatalf("viewer chart: %+v", allowed)
	}
}

func hasEvent(history []any, event string) bool {
	for _, raw := range history {
		if raw.(map[string]any)["eventType"] == event {
			return true
		}
	}
	return false
}

func uintToID(id uint64) string {
	return strconvFormat(id)
}
