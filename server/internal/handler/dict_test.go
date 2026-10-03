package handler_test

import (
	"net/http"
	"strings"
	"sync"
	"testing"
)

func TestDictReasonAndNumber(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, items := doJSON(http.MethodGet, "/api/v1/dict/items?typeCode=lot_status", token, nil)
	if items["code"] != float64(0) || len(items["data"].(map[string]any)["items"].([]any)) < 4 {
		t.Fatalf("dict items: %+v", items)
	}
	_, reasons := doJSON(http.MethodGet, "/api/v1/dict/reasons?category=hold", token, nil)
	if reasons["code"] != float64(0) || len(reasons["data"].(map[string]any)["reasons"].([]any)) == 0 {
		t.Fatalf("reasons: %+v", reasons)
	}
	_, badHold := doJSON(http.MethodPost, "/api/v1/wipLot/1/hold", token, map[string]any{"reasonCode": "NOT_A_CODE", "reason": "x"})
	if badHold["code"] == float64(0) {
		t.Fatalf("unknown hold code should fail: %+v", badHold)
	}

	typeID := createEntity(t, token, "/api/v1/sysDictType", map[string]any{"typeCode": "demo_flag", "typeName": "演示", "status": 1})
	createEntity(t, token, "/api/v1/sysDictItem", map[string]any{
		"typeID": mustNum(typeID), "itemCode": "on", "labelZh": "开", "labelEn": "On", "color": "#18a058", "tag": "success", "sortOrder": 1, "status": 1,
	})
	_, listed := doJSON(http.MethodPost, "/api/v1/sysDictItem/list", token, map[string]any{
		"page": 0, "limit": 10,
		"columns": []map[string]any{{"name": "item_code", "exp": "eq", "value": "on", "logic": "and"}},
	})
	if listed["data"].(map[string]any)["total"].(float64) < 1 {
		t.Fatalf("item list: %+v", listed)
	}

	_, preview := doJSON(http.MethodPost, "/api/v1/sysNumberRule/preview", token, map[string]any{"ruleCode": "work_order"})
	if preview["code"] != float64(0) || !strings.HasPrefix(preview["data"].(map[string]any)["preview"].(string), "WO-") {
		t.Fatalf("preview: %+v", preview)
	}
	_, versions := doJSON(http.MethodGet, "/api/v1/wipWorkOrder/releasedVersions", token, nil)
	version := versions["data"].(map[string]any)["versions"].([]any)[0].(map[string]any)
	_, created := doJSON(http.MethodPost, "/api/v1/wipWorkOrder", token, map[string]any{
		"productID": uint64(version["productID"].(float64)), "routeVersionID": uint64(version["versionID"].(float64)), "plannedQty": 4,
	})
	if created["code"] != float64(0) {
		t.Fatalf("auto order: %+v", created)
	}
	orderID := toID(created["data"].(map[string]any)["id"])
	_, got := doJSON(http.MethodGet, "/api/v1/wipWorkOrder/"+orderID, token, nil)
	orderNo := ""
	if data, ok := got["data"].(map[string]any); ok {
		if row, ok := data["wipWorkOrder"].(map[string]any); ok {
			orderNo, _ = row["orderNo"].(string)
		}
		if orderNo == "" {
			orderNo, _ = data["orderNo"].(string)
		}
	}
	if !strings.HasPrefix(orderNo, "WO-") {
		_, list := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/list", token, map[string]any{
			"page": 0, "limit": 5, "columns": []map[string]any{{"name": "id", "exp": "eq", "value": orderID, "logic": "and"}},
		})
		rows := list["data"].(map[string]any)["wipWorkOrders"].([]any)
		if len(rows) == 0 || !strings.HasPrefix(rows[0].(map[string]any)["orderNo"].(string), "WO-") {
			t.Fatalf("generated order number: get=%+v list=%+v", got, list)
		}
	}

	viewID := createEntity(t, token, "/api/v1/sysUser", map[string]any{"username": "dictview", "password": "dictview123", "realName": "Dict", "status": 1})
	if _, resp := doJSON(http.MethodPut, "/api/v1/sysUser/"+viewID+"/roles", token, map[string]any{"roleIds": []uint64{4}}); resp["code"] != float64(0) {
		t.Fatalf("role: %+v", resp)
	}
	viewToken := login(t, "dictview", "dictview123")
	if _, resp := doJSON(http.MethodGet, "/api/v1/dict/items?typeCode=lot_type", viewToken, nil); resp["code"] != float64(0) {
		t.Fatalf("viewer dict read: %+v", resp)
	}
	if _, resp := doJSON(http.MethodPost, "/api/v1/sysDictType", viewToken, map[string]any{"typeCode": "nope", "typeName": "no", "status": 1}); resp["code"] != float64(40003) {
		t.Fatalf("viewer dict write: %+v", resp)
	}

	const n = 6
	var wg sync.WaitGroup
	codes := make(chan float64, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, resp := doJSON(http.MethodPost, "/api/v1/wipWorkOrder", token, map[string]any{
				"productID": uint64(version["productID"].(float64)), "routeVersionID": uint64(version["versionID"].(float64)), "plannedQty": 1,
			})
			code, _ := resp["code"].(float64)
			codes <- code
		}()
	}
	wg.Wait()
	close(codes)
	for code := range codes {
		if code != 0 {
			t.Fatalf("concurrent order code %v", code)
		}
	}
}
