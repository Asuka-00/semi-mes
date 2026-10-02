package handler_test

import (
	"net/http"
	"testing"
)

func TestTrackInOut(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, eqp := doJSON(http.MethodGet, "/api/v1/eqpEquipment", token, nil)
	if eqp["code"] != float64(0) {
		t.Fatalf("equipment: %+v", eqp)
	}
	tools := map[string]uint64{}
	for _, row := range eqp["data"].(map[string]any)["equipment"].([]any) {
		item := row.(map[string]any)
		tools[item["equipmentCode"].(string)] = uint64(item["id"].(float64))
	}
	if tools["WET-01"] == 0 || tools["PHOTO-01"] == 0 || tools["PHOTO-09"] == 0 {
		t.Fatalf("missing tools: %+v", tools)
	}

	_, versions := doJSON(http.MethodGet, "/api/v1/wipWorkOrder/releasedVersions", token, nil)
	version := versions["data"].(map[string]any)["versions"].([]any)[0].(map[string]any)
	orderID := createEntity(t, token, "/api/v1/wipWorkOrder", map[string]any{
		"orderNo": "WO-TRACK", "productID": uint64(version["productID"].(float64)),
		"routeVersionID": uint64(version["versionID"].(float64)), "plannedQty": 20, "priority": 2,
	})
	_, _ = doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/release", token, map[string]any{})
	_, started := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 10, "lotType": "production"})
	lotID := uint64(started["data"].(map[string]any)["id"].(float64))
	lotIDStr := toID(started["data"].(map[string]any)["id"])
	lotNo := started["data"].(map[string]any)["lotNo"].(string)

	_, badIn := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["WET-01"]})
	if badIn["code"] == float64(0) {
		t.Fatalf("track in at start should fail: %+v", badIn)
	}
	_, passed := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID})
	if passed["code"] != float64(0) || passed["data"].(map[string]any)["lot"].(map[string]any)["currentNodeKey"] != "clean" {
		t.Fatalf("pass start: %+v", passed)
	}
	_, wrong := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["PHOTO-01"]})
	if wrong["code"] == float64(0) {
		t.Fatalf("wrong group: %+v", wrong)
	}
	_, down := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["PHOTO-09"]})
	if down["code"] == float64(0) {
		t.Fatalf("down tool: %+v", down)
	}
	_, tracked := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["WET-01"]})
	if tracked["code"] != float64(0) || tracked["data"].(map[string]any)["lot"].(map[string]any)["status"] != "running" {
		t.Fatalf("track in: %+v", tracked)
	}
	_, again := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["WET-01"]})
	if again["code"] == float64(0) {
		t.Fatalf("double track in: %+v", again)
	}
	_, aborted := doJSON(http.MethodPost, "/api/v1/wipMove/abort", token, map[string]any{"lotId": lotID, "reason": "wrong recipe"})
	if aborted["code"] != float64(0) || aborted["data"].(map[string]any)["lot"].(map[string]any)["status"] != "waiting" {
		t.Fatalf("abort: %+v", aborted)
	}
	_, tracked = doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["WET-01"]})
	if tracked["code"] != float64(0) {
		t.Fatalf("track in again: %+v", tracked)
	}
	_, mismatch := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 9, "qtyScrap": 0})
	if mismatch["code"] == float64(0) {
		t.Fatalf("qty mismatch: %+v", mismatch)
	}
	_, noReason := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 8, "qtyScrap": 2})
	if noReason["code"] == float64(0) {
		t.Fatalf("scrap without reason: %+v", noReason)
	}
	_, out := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 8, "qtyScrap": 2, "scrapReasonCode": "PARTICLE"})
	outLot := out["data"].(map[string]any)["lot"].(map[string]any)
	if out["code"] != float64(0) || outLot["currentNodeKey"] != "photo" || int(outLot["quantity"].(float64)) != 8 {
		t.Fatalf("track out: %+v", out)
	}

	step := func(group string, code string) {
		t.Helper()
		_, in := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools[code]})
		if in["code"] != float64(0) {
			t.Fatalf("track in %s: %+v", group, in)
		}
		_, done := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 8})
		if done["code"] != float64(0) {
			t.Fatalf("track out %s: %+v", group, done)
		}
	}
	step("PHOTO", "PHOTO-01")
	step("METRO", "METRO-01")
	_, missing := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID})
	if missing["code"] == float64(0) {
		t.Fatalf("decide without inspection: %+v", missing)
	}
	_, rework := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID, "inspectionResult": "fail"})
	if rework["code"] != float64(0) || rework["data"].(map[string]any)["lot"].(map[string]any)["currentNodeKey"] != "photo" {
		t.Fatalf("rework: %+v", rework)
	}

	_, heldLot := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotIDStr+"/hold", token, map[string]any{"reasonCode": "QA", "reason": "check"})
	if heldLot["code"] != float64(0) {
		t.Fatalf("hold: %+v", heldLot)
	}
	_, heldIn := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools["PHOTO-01"]})
	if heldIn["code"] == float64(0) {
		t.Fatalf("track in while hold: %+v", heldIn)
	}

	_, station := doJSON(http.MethodGet, "/api/v1/wipMove/station?lotNo="+lotNo, token, nil)
	if station["code"] != float64(0) || station["data"].(map[string]any)["nodeType"] == "" {
		t.Fatalf("station: %+v", station)
	}
	_, listed := doJSON(http.MethodPost, "/api/v1/wipMove/list", token, map[string]any{
		"page": 0, "limit": 20, "columns": []map[string]any{{"name": "lot_no", "exp": "=", "value": lotNo}},
	})
	moves := listed["data"].(map[string]any)["wipMoves"].([]any)
	if listed["code"] != float64(0) || len(moves) == 0 {
		t.Fatalf("history: %+v", listed)
	}
	_, overview := doJSON(http.MethodGet, "/api/v1/wipMove/overview", token, nil)
	if overview["code"] != float64(0) || len(overview["data"].(map[string]any)["holds"].([]any)) == 0 {
		t.Fatalf("overview: %+v", overview)
	}

	_, user := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "viewerwip", "password": "viewer123", "realName": "Viewer", "status": 1,
	})
	userID := toID(user["data"].(map[string]any)["id"])
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{4}})
	viewer := login(t, "viewerwip", "viewer123")
	_, forbidden := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", viewer, map[string]any{"lotId": lotID, "equipmentId": tools["PHOTO-01"]})
	if forbidden["code"] != float64(40003) {
		t.Fatalf("viewer track in: %+v", forbidden)
	}
	_, allowed := doJSON(http.MethodGet, "/api/v1/wipMove/overview", viewer, nil)
	if allowed["code"] != float64(0) {
		t.Fatalf("viewer overview: %+v", allowed)
	}
}
