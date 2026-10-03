package handler_test

import (
	"net/http"
	"strings"
	"testing"
)

func fmtLot(value any) string {
	text, _ := value.(string)
	return text
}

func TestWorkOrderAndLot(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, versions := doJSON(http.MethodGet, "/api/v1/wipWorkOrder/releasedVersions", token, nil)
	if versions["code"] != float64(0) {
		t.Fatalf("versions: %+v", versions)
	}
	list := versions["data"].(map[string]any)["versions"].([]any)
	if len(list) == 0 {
		t.Fatal("expected seeded released route")
	}
	version := list[0].(map[string]any)
	versionID := uint64(version["versionID"].(float64))
	productID := uint64(version["productID"].(float64))

	_, denied := doJSON(http.MethodPost, "/api/v1/wipWorkOrder", token, map[string]any{
		"orderNo": "WO-DRAFT", "productID": productID, "routeVersionID": 999999, "plannedQty": 10,
	})
	if denied["code"] == float64(0) {
		t.Fatalf("draft or missing version should fail: %+v", denied)
	}

	orderID := createEntity(t, token, "/api/v1/wipWorkOrder", map[string]any{
		"orderNo": "WO-CMOS", "productID": productID, "routeVersionID": versionID,
		"plannedQty": 20, "priority": 3, "dueDate": "2026-12-31", "note": "demo",
	})
	_, early := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 5, "lotType": "production"})
	if early["code"] == float64(0) {
		t.Fatalf("start before release: %+v", early)
	}
	_, released := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/release", token, map[string]any{})
	if released["code"] != float64(0) {
		t.Fatalf("release order: %+v", released)
	}

	_, started := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 10, "lotType": "production"})
	if started["code"] != float64(0) {
		t.Fatalf("start: %+v", started)
	}
	lotData := started["data"].(map[string]any)
	if !strings.HasPrefix(fmtLot(lotData["lotNo"]), "LOT-") || lotData["currentNodeKey"] != "start" {
		t.Fatalf("lot identity: %+v", lotData)
	}
	lotID := toID(lotData["id"])

	_, eng := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 6, "lotType": "engineering"})
	engNo := fmtLot(eng["data"].(map[string]any)["lotNo"])
	if eng["code"] != float64(0) || !strings.HasPrefix(engNo, "LOT-") || engNo == fmtLot(lotData["lotNo"]) {
		t.Fatalf("engineering lot: %+v", eng)
	}
	engID := toID(eng["data"].(map[string]any)["id"])

	_, overflow := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 10, "lotType": "production"})
	if overflow["code"] == float64(0) {
		t.Fatalf("overflow should fail: %+v", overflow)
	}

	walk := []string{"clean", "photo", "inspect", "decide"}
	for _, next := range walk {
		_, resp := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{"inspectionResult": "pass"})
		if resp["code"] != float64(0) {
			t.Fatalf("advance to %s: %+v", next, resp)
		}
		lot := resp["data"].(map[string]any)["lot"].(map[string]any)
		if lot["currentNodeKey"] != next {
			t.Fatalf("node %s got %+v", next, lot)
		}
	}
	_, rework := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{"inspectionResult": "fail"})
	reworkLot := rework["data"].(map[string]any)["lot"].(map[string]any)
	if rework["code"] != float64(0) || reworkLot["currentNodeKey"] != "photo" {
		t.Fatalf("rework: %+v", rework)
	}
	for _, next := range []string{"inspect", "decide"} {
		_, resp := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{"inspectionResult": "pass"})
		lot := resp["data"].(map[string]any)["lot"].(map[string]any)
		if lot["currentNodeKey"] != next {
			t.Fatalf("back to %s: %+v", next, resp)
		}
	}
	_, rework2 := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{"inspectionResult": "fail"})
	if rework2["data"].(map[string]any)["lot"].(map[string]any)["currentNodeKey"] != "photo" {
		t.Fatalf("second rework: %+v", rework2)
	}
	for _, next := range []string{"inspect", "decide"} {
		_, resp := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{})
		if resp["data"].(map[string]any)["lot"].(map[string]any)["currentNodeKey"] != next {
			t.Fatalf("third loop %s: %+v", next, resp)
		}
	}
	_, held := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{"inspectionResult": "fail"})
	heldLot := held["data"].(map[string]any)["lot"].(map[string]any)
	if held["code"] != float64(0) || heldLot["status"] != "hold" {
		t.Fatalf("rework hold: %+v", held)
	}
	_, resume := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/releaseHold", token, map[string]any{"reasonCode": "ENG_OK", "reason": "continue"})
	if resume["code"] != float64(0) {
		t.Fatalf("release hold: %+v", resume)
	}
	_, etch := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{"inspectionResult": "pass"})
	if etch["data"].(map[string]any)["result"].(map[string]any)["nextNodeKey"] != "etch" {
		t.Fatalf("default etch: %+v", etch)
	}
	_, done := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/advance", token, map[string]any{})
	doneLot := done["data"].(map[string]any)["lot"].(map[string]any)
	if doneLot["status"] != "completed" || doneLot["currentNodeKey"] != "end_main" {
		t.Fatalf("complete: %+v", done)
	}

	_, split := doJSON(http.MethodPost, "/api/v1/wipLot/"+engID+"/split", token, map[string]any{"quantities": []int{2, 1}})
	if split["code"] != float64(0) {
		t.Fatalf("split: %+v", split)
	}
	parent := split["data"].(map[string]any)["lot"].(map[string]any)
	if int(parent["quantity"].(float64)) != 3 {
		t.Fatalf("remainder: %+v", parent)
	}
	children := split["data"].(map[string]any)["children"].([]any)
	childID := toID(children[0].(map[string]any)["id"])
	_, merged := doJSON(http.MethodPost, "/api/v1/wipLot/merge", token, map[string]any{
		"targetId": mustNum(engID), "sourceIds": []uint64{mustNum(childID)},
	})
	if merged["code"] != float64(0) || int(merged["data"].(map[string]any)["lot"].(map[string]any)["quantity"].(float64)) != 5 {
		t.Fatalf("merge: %+v", merged)
	}

	_, detail := doJSON(http.MethodGet, "/api/v1/wipLot/"+engID, token, nil)
	detailData := detail["data"].(map[string]any)
	if detail["code"] != float64(0) || len(detailData["history"].([]any)) == 0 || len(detailData["nodes"].([]any)) == 0 {
		t.Fatalf("detail: %+v", detail)
	}

	_, user := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "viewerlot", "password": "viewer123", "realName": "Viewer", "status": 1,
	})
	userID := toID(user["data"].(map[string]any)["id"])
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{4}})
	viewer := login(t, "viewerlot", "viewer123")
	_, forbidden := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/close", viewer, map[string]any{})
	if forbidden["code"] != float64(40003) {
		t.Fatalf("viewer close: %+v", forbidden)
	}
	_, listed := doJSON(http.MethodPost, "/api/v1/wipLot/list", viewer, map[string]any{"page": 0, "limit": 10})
	if listed["code"] != float64(0) {
		t.Fatalf("viewer list: %+v", listed)
	}
}
