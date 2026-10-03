package handler_test

import (
	"net/http"
	"testing"
)

func TestCarrierWaferAndTrace(t *testing.T) {
	token := login(t, "admin", "admin123")
	_, created := doJSON(http.MethodPost, "/api/v1/wipCarrier", token, map[string]any{
		"carrierType": "FOUP", "capacity": 25, "status": "empty", "location": "STOCKER", "cleanCount": 1, "cleanLimit": 10,
	})
	if created["code"] != float64(0) {
		t.Fatalf("create carrier: %+v", created)
	}
	carrierID := toID(created["data"].(map[string]any)["id"])
	carrierNo := created["data"].(map[string]any)["carrierNo"].(string)
	if len(carrierNo) < 4 {
		t.Fatalf("carrier number: %+v", created)
	}
	_, listed := doJSON(http.MethodPost, "/api/v1/wipCarrier/list", token, map[string]any{
		"page": 0, "limit": 10,
		"columns": []map[string]any{{"name": "carrier_no", "exp": "=", "value": carrierNo, "logic": "and"}},
	})
	if listed["data"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("carrier list: %+v", listed)
	}

	version := releasedCMOS(t, token)
	orderID := createEntity(t, token, "/api/v1/wipWorkOrder", map[string]any{
		"orderNo": "WO-CARRIER", "productID": uint64(version["productID"].(float64)),
		"routeVersionID": uint64(version["versionID"].(float64)), "plannedQty": 8,
	})
	if _, resp := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/release", token, map[string]any{}); resp["code"] != float64(0) {
		t.Fatalf("release: %+v", resp)
	}
	_, started := doJSON(http.MethodPost, "/api/v1/wipWorkOrder/"+orderID+"/start", token, map[string]any{"quantity": 4, "lotType": "production"})
	if started["code"] != float64(0) {
		t.Fatalf("start: %+v", started)
	}
	lotID := toID(started["data"].(map[string]any)["id"])
	_, wafers := doJSON(http.MethodGet, "/api/v1/wipLot/"+lotID+"/wafers", token, nil)
	rows := wafers["data"].(map[string]any)["wafers"].([]any)
	if wafers["code"] != float64(0) || len(rows) != 4 {
		t.Fatalf("wafers: %+v", wafers)
	}
	first := rows[0].(map[string]any)
	second := rows[1].(map[string]any)
	if _, resp := doJSON(http.MethodPost, "/api/v1/wipCarrier/"+carrierID+"/bind", token, map[string]any{"lotId": uint64(started["data"].(map[string]any)["id"].(float64))}); resp["code"] != float64(0) {
		t.Fatalf("bind: %+v", resp)
	}
	_, slots := doJSON(http.MethodGet, "/api/v1/wipCarrier/"+carrierID+"/slots", token, nil)
	cells := slots["data"].(map[string]any)["slots"].([]any)
	if len(cells) != 25 || cells[0].(map[string]any)["waferNo"] == "" {
		t.Fatalf("slots: %+v", slots["data"])
	}
	_, split := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/split", token, map[string]any{
		"waferIds": [][]uint64{{uint64(first["id"].(float64))}},
	})
	if split["code"] != float64(0) {
		t.Fatalf("split wafer: %+v", split)
	}
	child := split["data"].(map[string]any)["children"].([]any)[0].(map[string]any)
	childID := toID(child["id"])
	_, childWafers := doJSON(http.MethodGet, "/api/v1/wipLot/"+childID+"/wafers", token, nil)
	if len(childWafers["data"].(map[string]any)["wafers"].([]any)) != 1 {
		t.Fatalf("child wafers: %+v", childWafers)
	}
	_, scrap := doJSON(http.MethodPost, "/api/v1/wipLot/"+lotID+"/scrap", token, map[string]any{
		"waferIds": []uint64{uint64(second["id"].(float64))}, "reasonCode": "BROKEN",
	})
	if scrap["code"] != float64(0) || int(scrap["data"].(map[string]any)["lot"].(map[string]any)["quantity"].(float64)) != 2 {
		t.Fatalf("scrap wafer: %+v", scrap)
	}
	waferNo := first["waferNo"].(string)
	_, traced := doJSON(http.MethodPost, "/api/v1/wipLot/trace", token, map[string]any{"waferNo": waferNo})
	if traced["code"] != float64(0) {
		t.Fatalf("wafer trace: %+v", traced)
	}
	report := traced["data"].(map[string]any)
	found := report["wafers"].([]any)
	if len(found) != 1 || found[0].(map[string]any)["waferNo"] != waferNo || len(found[0].(map[string]any)["history"].([]any)) == 0 {
		t.Fatalf("trace wafers: %+v", found)
	}

	if _, resp := doJSON(http.MethodPost, "/api/v1/wipMove/pass", token, map[string]any{"lotId": uint64(started["data"].(map[string]any)["id"].(float64))}); resp["code"] != float64(0) {
		t.Fatalf("pass: %+v", resp)
	}
	_, eqp := doJSON(http.MethodGet, "/api/v1/eqpEquipment", token, nil)
	var wet uint64
	for _, row := range eqp["data"].(map[string]any)["equipment"].([]any) {
		item := row.(map[string]any)
		if item["equipmentCode"] == "WET-01" {
			wet = uint64(item["id"].(float64))
		}
	}
	_, bad := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{
		"lotId": uint64(started["data"].(map[string]any)["id"].(float64)), "equipmentId": wet, "carrierNo": "MISSING",
	})
	if bad["code"] == float64(0) {
		t.Fatalf("wrong carrier should fail: %+v", bad)
	}
	_, tracked := doJSON(http.MethodPost, "/api/v1/wipMove/trackIn", token, map[string]any{
		"lotId": uint64(started["data"].(map[string]any)["id"].(float64)), "equipmentId": wet, "carrierNo": carrierNo,
	})
	if tracked["code"] != float64(0) {
		t.Fatalf("track in with carrier: %+v", tracked)
	}
	lotIDNum := uint64(started["data"].(map[string]any)["id"].(float64))
	if _, resp := doJSON(http.MethodPost, "/api/v1/wipMove/abort", token, map[string]any{"lotId": lotIDNum, "reason": "test"}); resp["code"] != float64(0) {
		t.Fatalf("abort: %+v", resp)
	}
}
