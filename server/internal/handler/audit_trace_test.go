package handler_test

import (
	"net/http"
	"strings"
	"testing"
)

func TestAuditAndTrace(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, bad := doJSON(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"userName": "admin", "password": "nope"})
	if bad["code"] == float64(0) {
		t.Fatal("expected failed login")
	}

	_, audits := doJSON(http.MethodPost, "/api/v1/sysAudit/list", token, map[string]any{
		"page": 0, "limit": 20, "sort": "-id",
		"columns": []map[string]any{
			{"name": "username", "exp": "eq", "value": "admin", "logic": "and"},
			{"name": "action", "exp": "eq", "value": "login", "logic": "and"},
		},
	})
	if audits["code"] != float64(0) {
		t.Fatalf("audit list: %+v", audits)
	}
	auditData := audits["data"].(map[string]any)
	if auditData["total"].(float64) < 1 {
		t.Fatalf("expected login audit: %+v", audits)
	}
	first := auditData["sysAudits"].([]any)[0].(map[string]any)
	if strings.Contains(fmtString(first["afterJson"]), "password") {
		t.Fatalf("login audit kept a password: %+v", first)
	}

	_, failed := doJSON(http.MethodPost, "/api/v1/sysAudit/list", token, map[string]any{
		"page": 0, "limit": 5,
		"columns": []map[string]any{{"name": "action", "exp": "eq", "value": "login_failed", "logic": "and"}},
	})
	if failed["data"].(map[string]any)["total"].(float64) < 1 {
		t.Fatalf("expected login_failed: %+v", failed)
	}

	createEntity(t, token, "/api/v1/baseFactory", map[string]any{"factoryCode": "AUDIT-FAB", "factoryName": "Audit Fab", "status": 1})
	_, created := doJSON(http.MethodPost, "/api/v1/sysAudit/list", token, map[string]any{
		"page": 0, "limit": 5,
		"columns": []map[string]any{
			{"name": "entity_code", "exp": "like", "value": "AUDIT-FAB", "logic": "and"},
			{"name": "action", "exp": "eq", "value": "create", "logic": "and"},
		},
	})
	rows := created["data"].(map[string]any)["sysAudits"].([]any)
	if len(rows) == 0 {
		t.Fatalf("factory create was not audited: %+v", created)
	}
	row := rows[0].(map[string]any)
	if !strings.Contains(fmtString(row["afterJson"]), "factory_code") {
		t.Fatalf("missing factory snapshot: %+v", row)
	}

	_, versions := doJSON(http.MethodGet, "/api/v1/wipWorkOrder/releasedVersions", token, nil)
	version := versions["data"].(map[string]any)["versions"].([]any)[0].(map[string]any)
	orderID := createEntity(t, token, "/api/v1/wipWorkOrder", map[string]any{
		"orderNo": "AUDIT-WO", "productID": uint64(version["productID"].(float64)), "routeVersionID": uint64(version["versionID"].(float64)),
		"plannedQty": 8, "priority": 2,
	})
	if _, released := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/release", token, map[string]any{}); released["code"] != float64(0) {
		t.Fatalf("release: %+v", released)
	}
	_, started := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 6, "lotType": "production"})
	if started["code"] != float64(0) {
		t.Fatalf("start: %+v", started)
	}
	lotID := toID(started["data"].(map[string]any)["id"])
	_, held := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/hold", token, map[string]any{"reasonCode": "AUDIT_HOLD", "reason": "audit"})
	if held["code"] != float64(0) {
		t.Fatalf("hold: %+v", held)
	}
	_, holdAudit := doJSON(http.MethodPost, "/api/v1/sysAudit/list", token, map[string]any{
		"page": 0, "limit": 5,
		"columns": []map[string]any{
			{"name": "action", "exp": "eq", "value": "hold", "logic": "and"},
			{"name": "entity_id", "exp": "eq", "value": lotID, "logic": "and"},
		},
	})
	holdRows := holdAudit["data"].(map[string]any)["sysAudits"].([]any)
	if len(holdRows) == 0 || !strings.Contains(fmtString(holdRows[0].(map[string]any)["diffJson"])+fmtString(holdRows[0].(map[string]any)["afterJson"]), "hold") {
		t.Fatalf("hold audit: %+v", holdAudit)
	}

	_, split := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/split", token, map[string]any{"quantities": []int{2}})
	if split["code"] == float64(0) {
		t.Fatal("held lot should not split")
	}
	_, releasedHold := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/releaseHold", token, map[string]any{})
	if releasedHold["code"] != float64(0) {
		t.Fatalf("release hold: %+v", releasedHold)
	}
	_, split = doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/split", token, map[string]any{"quantities": []int{2}})
	if split["code"] != float64(0) {
		t.Fatalf("split: %+v", split)
	}
	children := split["data"].(map[string]any)["children"].([]any)
	childNo := children[0].(map[string]any)["lotNo"].(string)

	_, trace := doJSON(http.MethodPost, "/api/v1/wipLot/trace", token, map[string]any{"lotNo": "AUDIT-WO-001"})
	if trace["code"] != float64(0) {
		t.Fatalf("trace: %+v", trace)
	}
	forward := trace["data"].(map[string]any)["forward"].([]any)
	if len(forward) == 0 || forward[0].(map[string]any)["lotNo"] != childNo {
		t.Fatalf("forward genealogy: %+v", trace["data"])
	}
	_, back := doJSON(http.MethodGet, "/api/v1/wipLot/"+toID(children[0].(map[string]any)["id"])+"/trace", token, nil)
	backward := back["data"].(map[string]any)["backward"].([]any)
	if len(backward) == 0 || backward[0].(map[string]any)["lotNo"] != "AUDIT-WO-001" {
		t.Fatalf("backward genealogy: %+v", back["data"])
	}
	history := trace["data"].(map[string]any)["history"].([]any)
	if len(history) == 0 {
		t.Fatal("expected lot history")
	}

	_, demo := doJSON(http.MethodPost, "/api/v1/wipLot/trace", token, map[string]any{"lotNo": "DEMO-WIP-003"})
	moves := demo["data"].(map[string]any)["moves"].([]any)
	if len(moves) == 0 || moves[0].(map[string]any)["equipmentCode"] != "WET-01" {
		t.Fatalf("demo moves: %+v", demo["data"])
	}
	_, defects := doJSON(http.MethodPost, "/api/v1/wipLot/trace", token, map[string]any{"lotNo": "DEMO-WIP-001"})
	if len(defects["data"].(map[string]any)["defects"].([]any)) == 0 {
		t.Fatalf("expected defects: %+v", defects)
	}
	_, reverse := doJSON(http.MethodPost, "/api/v1/wipLot/trace/reverse", token, map[string]any{"equipmentCode": "WET-01"})
	lots := reverse["data"].(map[string]any)["lots"].([]any)
	foundWet := false
	for _, item := range lots {
		if item.(map[string]any)["lotNo"] == "DEMO-WIP-003" {
			foundWet = true
		}
	}
	if !foundWet {
		t.Fatalf("reverse lookup: %+v", reverse)
	}
	_, emptyReverse := doJSON(http.MethodPost, "/api/v1/wipLot/trace/reverse", token, map[string]any{})
	if emptyReverse["code"] == float64(0) {
		t.Fatalf("reverse without filter should fail: %+v", emptyReverse)
	}

	_, loggedOut := doJSON(http.MethodPost, "/api/v1/auth/logout", token, map[string]any{})
	if loggedOut["code"] != float64(0) {
		t.Fatalf("logout: %+v", loggedOut)
	}
	token = login(t, "admin", "admin123")
	_, logoutAudit := doJSON(http.MethodPost, "/api/v1/sysAudit/list", token, map[string]any{
		"page": 0, "limit": 5,
		"columns": []map[string]any{{"name": "action", "exp": "eq", "value": "logout", "logic": "and"}},
	})
	if logoutAudit["data"].(map[string]any)["total"].(float64) < 1 {
		t.Fatalf("logout audit: %+v", logoutAudit)
	}

	sysID := createEntity(t, token, "/api/v1/sysUser", map[string]any{"username": "audsys", "password": "audsys123", "realName": "Audit", "status": 1})
	if _, resp := doJSON(http.MethodPut, "/api/v1/sysUser/"+sysID+"/roles", token, map[string]any{"roleIds": []uint64{2}}); resp["code"] != float64(0) {
		t.Fatalf("role: %+v", resp)
	}
	viewID := createEntity(t, token, "/api/v1/sysUser", map[string]any{"username": "audview", "password": "audview123", "realName": "View", "status": 1})
	if _, resp := doJSON(http.MethodPut, "/api/v1/sysUser/"+viewID+"/roles", token, map[string]any{"roleIds": []uint64{4}}); resp["code"] != float64(0) {
		t.Fatalf("role: %+v", resp)
	}
	sysToken := login(t, "audsys", "audsys123")
	if _, resp := doJSON(http.MethodPost, "/api/v1/sysAudit/list", sysToken, map[string]any{"page": 0, "limit": 1}); resp["code"] != float64(0) {
		t.Fatalf("sys admin audit: %+v", resp)
	}
	viewToken := login(t, "audview", "audview123")
	if _, resp := doJSON(http.MethodPost, "/api/v1/sysAudit/list", viewToken, map[string]any{"page": 0, "limit": 1}); resp["code"] != float64(40003) {
		t.Fatalf("viewer audit: %+v", resp)
	}
	if _, resp := doJSON(http.MethodPost, "/api/v1/wipLot/trace", viewToken, map[string]any{"lotNo": "DEMO-WIP-001"}); resp["code"] != float64(0) {
		t.Fatalf("viewer trace: %+v", resp)
	}
}

func fmtString(value any) string {
	text, _ := value.(string)
	return text
}
