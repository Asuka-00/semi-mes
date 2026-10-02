package handler_test

import (
	"net/http"
	"testing"
)

func TestShopDashboard(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, denied := doJSON(http.MethodGet, "/api/v1/dashboard", "", nil)
	if denied["code"] == float64(0) {
		t.Fatalf("unauth: %+v", denied)
	}
	_, resp := doJSON(http.MethodGet, "/api/v1/dashboard", token, nil)
	if resp["code"] != float64(0) {
		t.Fatalf("dashboard: %+v", resp)
	}
	data := resp["data"].(map[string]any)
	if data["wip"] != true || data["equipment"] != true {
		t.Fatalf("flags: %+v", data)
	}
	if data["wipLots"].(float64) < 6 || data["wipQty"].(float64) < 40 {
		t.Fatalf("wip kpi: %+v", data)
	}
	if data["holdLots"].(float64) < 2 || data["runningLots"].(float64) < 1 {
		t.Fatalf("hold/running: %+v", data)
	}
	if data["todayMoves"].(float64) < 4 || data["todayScrap"].(float64) < 1 || data["todayCompleted"].(float64) < 1 {
		t.Fatalf("today: %+v", data)
	}
	trend := data["trend"].([]any)
	if len(trend) != 7 {
		t.Fatalf("trend: %+v", trend)
	}
	steps := data["byStep"].([]any)
	products := data["byProduct"].([]any)
	if len(steps) < 3 || len(products) < 2 {
		t.Fatalf("charts steps=%d products=%d", len(steps), len(products))
	}
	if steps[0].(map[string]any)["label"] == "" {
		t.Fatalf("step label: %+v", steps[0])
	}
	holds := data["holds"].([]any)
	if len(holds) < 2 || holds[0].(map[string]any)["holdSeconds"].(float64) < 3600 {
		t.Fatalf("holds: %+v", holds)
	}
	var sawDown bool
	for _, row := range data["byEquipment"].([]any) {
		if row.(map[string]any)["key"] == "unscheduled_down" && row.(map[string]any)["count"].(float64) >= 1 {
			sawDown = true
		}
	}
	if !sawDown || data["overduePm"].(float64) < 1 {
		t.Fatalf("equipment: %+v overdue %v", data["byEquipment"], data["overduePm"])
	}
	if len(data["notices"].([]any)) == 0 {
		t.Fatalf("notices: %+v", data["notices"])
	}

	_, created := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "dashadmin", "password": "dash12345", "realName": "Sys", "status": 1,
	})
	if created["code"] != float64(0) {
		t.Fatalf("create sys admin: %+v", created)
	}
	userID := toID(created["data"].(map[string]any)["id"])
	_, linked := doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{2}})
	if linked["code"] != float64(0) {
		t.Fatalf("roles: %+v", linked)
	}
	sys := login(t, "dashadmin", "dash12345")
	_, hidden := doJSON(http.MethodGet, "/api/v1/dashboard", sys, nil)
	if hidden["code"] != float64(0) {
		t.Fatalf("sys dashboard: %+v", hidden)
	}
	view := hidden["data"].(map[string]any)
	if view["wip"] != false || view["equipment"] != false {
		t.Fatalf("sys flags: %+v", view)
	}
	if view["wipLots"].(float64) != 0 || view["todayMoves"].(float64) != 0 || view["overduePm"].(float64) != 0 {
		t.Fatalf("sys leaked numbers: %+v", view)
	}
	if len(view["byStep"].([]any)) != 0 || len(view["holds"].([]any)) != 0 || len(view["byEquipment"].([]any)) != 0 {
		t.Fatalf("sys leaked rows: %+v", view)
	}
}
