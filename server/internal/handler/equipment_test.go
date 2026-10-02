package handler_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestEquipmentStatePMAndPermissions(t *testing.T) {
	token := login(t, "admin", "admin123")
	id := createEntity(t, token, "/api/v1/eqpEquipment", map[string]any{
		"equipmentCode": "EQP-TEST", "equipmentName": "测试槽", "equipmentGroup": "WET", "equipmentType": "WET",
		"status": "standby", "capacity": 1, "chamberCount": 1, "manufacturer": "Demo", "modelName": "T-1",
		"serialNo": "SN-1", "location": "Bay T",
	})
	_, bad := doJSON(http.MethodPost, "/api/v1/eqpEquipment/"+id+"/state", token, map[string]any{"toState": "productive", "reasonCode": "PROD_START"})
	if bad["code"] == float64(0) {
		t.Fatalf("manual productive should fail: %+v", bad)
	}
	_, ns := doJSON(http.MethodPost, "/api/v1/eqpEquipment/"+id+"/state", token, map[string]any{"toState": "non_scheduled", "reasonCode": "SHIFT_END", "reason": "no shift"})
	if ns["code"] != float64(0) || ns["data"].(map[string]any)["status"] != "non_scheduled" {
		t.Fatalf("to non_scheduled: %+v", ns)
	}
	_, nope := doJSON(http.MethodPost, "/api/v1/eqpEquipment/"+id+"/state", token, map[string]any{"toState": "unscheduled_down", "reasonCode": "BREAKDOWN"})
	if nope["code"] == float64(0) {
		t.Fatalf("non_scheduled to unscheduled_down should fail: %+v", nope)
	}
	_, back := doJSON(http.MethodPost, "/api/v1/eqpEquipment/"+id+"/state", token, map[string]any{"toState": "standby", "reasonCode": "SHIFT_START"})
	if back["code"] != float64(0) {
		t.Fatalf("back to standby: %+v", back)
	}

	lotID, _ := startCleanLot(t, token, "WO-EQP")
	_, tracked := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": mustNum(id)})
	if tracked["code"] != float64(0) {
		t.Fatalf("track in: %+v", tracked)
	}
	_, detail := doJSON(http.MethodGet, "/api/v1/eqpEquipment/"+id, token, nil)
	eqp := detail["data"].(map[string]any)["equipment"].(map[string]any)
	if detail["code"] != float64(0) || eqp["status"] != "productive" || int(detail["data"].(map[string]any)["openLots"].(float64)) != 1 {
		t.Fatalf("detail after track in: %+v", detail)
	}
	lotID2, _ := startCleanLot(t, token, "WO-EQP2")
	_, full := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID2, "equipmentId": mustNum(id)})
	if full["code"] == float64(0) {
		t.Fatalf("capacity should block: %+v", full)
	}
	_, out := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 4})
	if out["code"] != float64(0) {
		t.Fatalf("track out: %+v", out)
	}
	_, after := doJSON(http.MethodGet, "/api/v1/eqpEquipment/"+id, token, nil)
	if after["data"].(map[string]any)["equipment"].(map[string]any)["status"] != "standby" {
		t.Fatalf("released: %+v", after["data"].(map[string]any)["equipment"])
	}
	logs := after["data"].(map[string]any)["stateLogs"].([]any)
	if len(logs) < 2 {
		t.Fatalf("state history: %+v", logs)
	}

	due := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	_, plan := doJSON(http.MethodPost, "/api/v1/eqpPmPlan", token, map[string]any{
		"equipmentID": mustNum(id), "planName": "测试保养", "triggerType": "time", "intervalDays": 7,
		"checklist": []string{"检查", "复位"}, "blockTrackIn": true, "nextDueAt": due, "enabled": true,
	})
	if plan["code"] != float64(0) {
		t.Fatalf("plan: %+v", plan)
	}
	_, tasks := doJSON(http.MethodPost, "/api/v1/eqpPmTask/list", token, map[string]any{"equipmentID": mustNum(id), "limit": 10})
	list := tasks["data"].(map[string]any)["tasks"].([]any)
	if tasks["code"] != float64(0) || len(list) == 0 || list[0].(map[string]any)["status"] != "overdue" {
		t.Fatalf("overdue task: %+v", tasks)
	}
	taskID := toID(list[0].(map[string]any)["id"])
	_, blocked := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID2, "equipmentId": mustNum(id)})
	if blocked["code"] == float64(0) {
		t.Fatalf("overdue PM should block track in: %+v", blocked)
	}
	_, started := doJSON(http.MethodPost, "/api/v1/eqpPmTask/"+taskID+"/start", token, map[string]any{})
	if started["code"] != float64(0) || started["data"].(map[string]any)["status"] != "in_progress" {
		t.Fatalf("start pm: %+v", started)
	}
	_, down := doJSON(http.MethodGet, "/api/v1/eqpEquipment/"+id, token, nil)
	if down["data"].(map[string]any)["equipment"].(map[string]any)["status"] != "scheduled_down" {
		t.Fatalf("scheduled down: %+v", down["data"].(map[string]any)["equipment"])
	}
	_, done := doJSON(http.MethodPost, "/api/v1/eqpPmTask/"+taskID+"/complete", token, map[string]any{
		"result": "pass", "note": "ok",
		"items": []map[string]any{{"name": "检查", "result": "pass"}, {"name": "复位", "result": "pass"}},
	})
	if done["code"] != float64(0) {
		t.Fatalf("complete pm: %+v", done)
	}
	_, ready := doJSON(http.MethodGet, "/api/v1/eqpEquipment/"+id, token, nil)
	if ready["data"].(map[string]any)["equipment"].(map[string]any)["status"] != "standby" {
		t.Fatalf("after pm: %+v", ready["data"].(map[string]any)["equipment"])
	}
	_, again := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID2, "equipmentId": mustNum(id)})
	if again["code"] != float64(0) {
		t.Fatalf("track in after pm: %+v", again)
	}

	_, board := doJSON(http.MethodGet, "/api/v1/eqpEquipment/board", token, nil)
	if board["code"] != float64(0) || len(board["data"].(map[string]any)["tools"].([]any)) == 0 {
		t.Fatalf("board: %+v", board)
	}
	_, user := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "viewerEqp", "password": "viewer123", "realName": "Viewer", "status": 1,
	})
	userID := toID(user["data"].(map[string]any)["id"])
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{4}})
	viewer := login(t, "viewerEqp", "viewer123")
	_, forbidden := doJSON(http.MethodPost, "/api/v1/eqpEquipment/"+id+"/state", viewer, map[string]any{"toState": "engineering", "reasonCode": "ENG_SETUP"})
	if forbidden["code"] != float64(40003) {
		t.Fatalf("viewer state: %+v", forbidden)
	}
	_, allowed := doJSON(http.MethodGet, "/api/v1/eqpEquipment/board", viewer, nil)
	if allowed["code"] != float64(0) {
		t.Fatalf("viewer board: %+v", allowed)
	}
}

func TestTrackOutReworkLimitHold(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, eqp := doJSON(http.MethodGet, "/api/v1/eqpEquipment", token, nil)
	tools := map[string]uint64{}
	for _, row := range eqp["data"].(map[string]any)["equipment"].([]any) {
		item := row.(map[string]any)
		tools[item["equipmentCode"].(string)] = uint64(item["id"].(float64))
	}
	lotID, _ := startCleanLot(t, token, "WO-REWORK")
	track := func(code string) {
		t.Helper()
		_, in := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{"lotId": lotID, "equipmentId": tools[code]})
		if in["code"] != float64(0) {
			t.Fatalf("track in %s: %+v", code, in)
		}
		_, out := doJSON(http.MethodPost, "/api/v1/wipMove/trackOut", token, map[string]any{"lotId": lotID, "qtyOut": 4})
		if out["code"] != float64(0) {
			t.Fatalf("track out %s: %+v", code, out)
		}
	}
	track("WET-01")
	for i := 0; i < 2; i++ {
		track("PHOTO-01")
		track("METRO-01")
		_, rework := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID, "inspectionResult": "fail"})
		lot := rework["data"].(map[string]any)["lot"].(map[string]any)
		if rework["code"] != float64(0) || lot["currentNodeKey"] != "photo" || lot["status"] == "hold" {
			t.Fatalf("rework %d: %+v", i+1, rework)
		}
	}
	track("PHOTO-01")
	track("METRO-01")
	_, held := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID, "inspectionResult": "fail"})
	if held["code"] != float64(0) {
		t.Fatalf("limit: %+v", held)
	}
	data := held["data"].(map[string]any)
	lot := data["lot"].(map[string]any)
	result := data["result"].(map[string]any)
	if lot["status"] != "hold" || lot["currentNodeKey"] != "decide" || result["action"] != "hold" {
		t.Fatalf("expected hold on decide: %+v", held)
	}
}

func startCleanLot(t *testing.T, token, orderNo string) (uint64, string) {
	t.Helper()
	_, versions := doJSON(http.MethodGet, "/api/v1/wipWorkOrder/releasedVersions", token, nil)
	version := versions["data"].(map[string]any)["versions"].([]any)[0].(map[string]any)
	orderID := createEntity(t, token, "/api/v1/wipWorkOrder", map[string]any{
		"orderNo": orderNo, "productID": uint64(version["productID"].(float64)),
		"routeVersionID": uint64(version["versionID"].(float64)), "plannedQty": 20, "priority": 2,
	})
	_, _ = doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/release", token, map[string]any{})
	_, started := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 4, "lotType": "production"})
	if started["code"] != float64(0) {
		t.Fatalf("start %s: %+v", orderNo, started)
	}
	lotID := uint64(started["data"].(map[string]any)["id"].(float64))
	_, passed := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": lotID})
	if passed["code"] != float64(0) || passed["data"].(map[string]any)["lot"].(map[string]any)["currentNodeKey"] != "clean" {
		t.Fatalf("pass start: %+v", passed)
	}
	return lotID, fmt.Sprint(orderNo)
}
