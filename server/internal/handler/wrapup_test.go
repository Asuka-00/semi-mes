package handler_test

import (
	"net/http"
	"testing"

	"semi-mes/server/internal/config"
)

func TestDefectReworkFollowsRoute(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, eqp := doJSON(http.MethodGet, "/api/v1/eqpEquipment", token, nil)
	tools := map[string]uint64{}
	for _, row := range eqp["data"].(map[string]any)["equipment"].([]any) {
		item := row.(map[string]any)
		tools[item["equipmentCode"].(string)] = uint64(item["id"].(float64))
	}
	lotID, _ := startCleanLot(t, token, "WO-REWORK-MOVE")
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
	rework := func() map[string]any {
		t.Helper()
		_, resp := doJSON(http.MethodPost, "/api/v1/qcDefect", token, map[string]any{
			"lotId": lotID, "defectCode": "PATTERN", "quantity": 1, "disposition": "rework", "note": "图形返工",
		})
		return resp
	}
	first := rework()
	if first["code"] == float64(0) {
		t.Fatalf("clean has no rework edge: %+v", first)
	}
	track("WET-01", 0)
	track("PHOTO-01", 0)
	track("METRO-01", 600)
	_, atDecide := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotID), token, nil)
	if atDecide["data"].(map[string]any)["lot"].(map[string]any)["currentNodeKey"] != "decide" {
		t.Fatalf("expected decide: %+v", atDecide["data"])
	}
	moved := rework()
	if moved["code"] != float64(0) {
		t.Fatalf("rework from decide: %+v", moved)
	}
	_, after := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotID), token, nil)
	lot := after["data"].(map[string]any)["lot"].(map[string]any)
	history := after["data"].(map[string]any)["history"].([]any)
	if lot["currentNodeKey"] != "photo" || lot["status"] != "waiting" || !historyHop(history, "rework", "decide", "photo") {
		t.Fatalf("rework should return to photo: %+v", after["data"])
	}
	track("PHOTO-01", 0)
	track("METRO-01", 610)
	if second := rework(); second["code"] != float64(0) {
		t.Fatalf("second rework: %+v", second)
	}
	track("PHOTO-01", 0)
	track("METRO-01", 615)
	third := rework()
	if third["code"] != float64(0) {
		t.Fatalf("rework over limit: %+v", third)
	}
	_, held := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotID), token, nil)
	heldLot := held["data"].(map[string]any)["lot"].(map[string]any)
	if heldLot["status"] != "hold" || heldLot["currentNodeKey"] != "decide" || heldLot["holdReasonCode"] != "REWORK_LIMIT" {
		t.Fatalf("max rework should hold on decide: %+v", heldLot)
	}
	if !hasEvent(held["data"].(map[string]any)["history"].([]any), "hold") {
		t.Fatalf("hold history missing: %+v", held["data"])
	}
}

func historyHop(history []any, event, from, to string) bool {
	for _, raw := range history {
		row := raw.(map[string]any)
		if row["eventType"] == event && row["fromNodeKey"] == from && row["toNodeKey"] == to {
			return true
		}
	}
	return false
}

func TestSPCFlagOffSkipsMenusAndReactions(t *testing.T) {
	config.Get().Features.SPC = false
	t.Cleanup(func() { config.Get().Features.SPC = true })
	token := login(t, "admin", "admin123")
	_, features := doJSON(http.MethodGet, "/api/v1/features", token, nil)
	if features["code"] != float64(0) || features["data"].(map[string]any)["spc"] != false {
		t.Fatalf("features: %+v", features)
	}
	_, chart := doJSON(http.MethodGet, "/api/v1/qcSpc/chart?param=CD", token, nil)
	if chart["code"] != float64(40003) {
		t.Fatalf("spc api should be forbidden: %+v", chart)
	}
	_, routes := doJSON(http.MethodGet, "/api/v1/route/getUserRoutes", token, nil)
	if routeNamed(routes["data"].(map[string]any)["routes"].([]any), "quality_spc") {
		t.Fatalf("spc menu should be hidden: %+v", routes["data"])
	}
	if !routeNamed(routes["data"].(map[string]any)["routes"].([]any), "quality_defect") {
		t.Fatalf("other quality pages stay: %+v", routes["data"])
	}
	_, plans := doJSON(http.MethodGet, "/api/v1/qcInspectPlan", token, nil)
	var operationID uint64
	for _, raw := range plans["data"].(map[string]any)["plans"].([]any) {
		operationID = uint64(raw.(map[string]any)["operationID"].(float64))
	}
	lotID, _ := startCleanLot(t, token, "WO-SPC-OFF")
	_, measured := doJSON(http.MethodPost, "/api/v1/qcMeasurement", token, map[string]any{
		"lotId": lotID, "operationId": operationID,
		"samples": []map[string]any{{"paramCode": "CD", "values": []float64{900}}},
	})
	if measured["code"] != float64(0) {
		t.Fatalf("measure while spc off: %+v", measured)
	}
	_, detail := doJSON(http.MethodGet, "/api/v1/wipLot/"+uintToID(lotID), token, nil)
	if detail["data"].(map[string]any)["lot"].(map[string]any)["status"] != "waiting" {
		t.Fatalf("spc off must not hold: %+v", detail["data"].(map[string]any)["lot"])
	}
}

func routeNamed(routes []any, name string) bool {
	for _, raw := range routes {
		node := raw.(map[string]any)
		if node["name"] == name {
			return true
		}
		children, _ := node["children"].([]any)
		if routeNamed(children, name) {
			return true
		}
	}
	return false
}

func TestNoticesAndSoftDeletedUsername(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, list := doJSON(http.MethodGet, "/api/v1/sysNotice", token, nil)
	if list["code"] != float64(0) {
		t.Fatalf("notices: %+v", list)
	}
	notices := list["data"].(map[string]any)["notices"].([]any)
	if len(notices) == 0 || list["data"].(map[string]any)["unread"].(float64) < 1 {
		t.Fatalf("seed should create pm notices: %+v", list["data"])
	}
	sawPM := false
	for _, raw := range notices {
		kind := raw.(map[string]any)["kind"]
		if kind == "pm_overdue" || kind == "pm_due" {
			sawPM = true
			if raw.(map[string]any)["titleZh"] == "" || raw.(map[string]any)["titleEn"] == "" {
				t.Fatalf("bilingual notice: %+v", raw)
			}
		}
	}
	if !sawPM {
		t.Fatalf("missing pm notice: %+v", notices)
	}
	id := uintToID(uint64(notices[0].(map[string]any)["id"].(float64)))
	_, marked := doJSON(http.MethodPost, "/api/v1/sysNotice/"+id+"/read", token, map[string]any{})
	if marked["code"] != float64(0) {
		t.Fatalf("mark read: %+v", marked)
	}
	_, all := doJSON(http.MethodPost, "/api/v1/sysNotice/readAll", token, map[string]any{})
	if all["code"] != float64(0) {
		t.Fatalf("read all: %+v", all)
	}
	_, empty := doJSON(http.MethodGet, "/api/v1/sysNotice", token, nil)
	if empty["data"].(map[string]any)["unread"].(float64) != 0 {
		t.Fatalf("unread should be zero: %+v", empty["data"])
	}

	lotID, _ := startCleanLot(t, token, "WO-NOTICE-HOLD")
	_, held := doJSON(http.MethodPost, "/api/v1/wipLot/"+uintToID(lotID)+"/hold", token, map[string]any{"reasonCode": "QA", "reason": "抽检"})
	if held["code"] != float64(0) {
		t.Fatalf("hold: %+v", held)
	}
	_, again := doJSON(http.MethodGet, "/api/v1/sysNotice", token, nil)
	if !noticeKind(again["data"].(map[string]any)["notices"].([]any), "lot_hold") {
		t.Fatalf("hold notice: %+v", again["data"])
	}

	_, down := doJSON(http.MethodPost, "/api/v1/eqpEquipment", token, map[string]any{
		"equipmentCode": "EQP-DOWN-NOTE", "equipmentName": "停机通知台", "equipmentGroup": "WET", "status": "standby", "capacity": 1,
	})
	if down["code"] != float64(0) {
		t.Fatalf("create eqp: %+v", down)
	}
	eqpID := toID(down["data"].(map[string]any)["id"])
	_, state := doJSON(http.MethodPost, "/api/v1/eqpEquipment/"+eqpID+"/state", token, map[string]any{"toState": "unscheduled_down", "reasonCode": "BREAKDOWN", "reason": "真空异常"})
	if state["code"] != float64(0) {
		t.Fatalf("down: %+v", state)
	}
	_, downNotes := doJSON(http.MethodGet, "/api/v1/sysNotice", token, nil)
	if !noticeKind(downNotes["data"].(map[string]any)["notices"].([]any), "eqp_unscheduled_down") {
		t.Fatalf("down notice: %+v", downNotes["data"])
	}

	_, created := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "reuse-name", "password": "reuse123", "realName": "Reuse", "status": 1,
	})
	if created["code"] != float64(0) {
		t.Fatalf("create user: %+v", created)
	}
	userID := toID(created["data"].(map[string]any)["id"])
	_, deleted := doJSON(http.MethodDelete, "/api/v1/sysUser/"+userID, token, nil)
	if deleted["code"] != float64(0) {
		t.Fatalf("delete user: %+v", deleted)
	}
	_, gone := doJSON(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"userName": "reuse-name", "password": "reuse123"})
	if gone["code"] == float64(0) {
		t.Fatalf("deleted user must not log in: %+v", gone)
	}
	_, againUser := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "reuse-name", "password": "reuse456", "realName": "Reuse 2", "status": 1,
	})
	if againUser["code"] != float64(0) {
		t.Fatalf("username should be reusable: %+v", againUser)
	}
	_, dup := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "reuse-name", "password": "reuse789", "realName": "Reuse 3", "status": 1,
	})
	if dup["code"] == float64(0) {
		t.Fatalf("active username must stay unique: %+v", dup)
	}
	relogin := login(t, "reuse-name", "reuse456")
	if relogin == "" {
		t.Fatal("new user should log in")
	}
	_, role := doJSON(http.MethodPost, "/api/v1/sysRole", token, map[string]any{
		"roleCode": "reuse_role", "roleName": "Reuse Role", "status": 1,
	})
	if role["code"] != float64(0) {
		t.Fatalf("create role: %+v", role)
	}
	roleID := toID(role["data"].(map[string]any)["id"])
	_, _ = doJSON(http.MethodDelete, "/api/v1/sysRole/"+roleID, token, nil)
	_, role2 := doJSON(http.MethodPost, "/api/v1/sysRole", token, map[string]any{
		"roleCode": "reuse_role", "roleName": "Reuse Role 2", "status": 1,
	})
	if role2["code"] != float64(0) {
		t.Fatalf("role code should be reusable: %+v", role2)
	}
}

func noticeKind(notices []any, kind string) bool {
	for _, raw := range notices {
		if raw.(map[string]any)["kind"] == kind {
			return true
		}
	}
	return false
}
