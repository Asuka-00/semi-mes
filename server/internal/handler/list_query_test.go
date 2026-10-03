package handler_test

import (
	"net/http"
	"testing"
)

func TestListQueryAndBatch(t *testing.T) {
	token := login(t, "admin", "admin123")
	code := "LQ-" + toID(float64(len(token)))
	idA := createEntity(t, token, "/api/v1/baseFactory", map[string]any{"factoryCode": code + "A", "factoryName": "Query Alpha", "status": 1})
	idB := createEntity(t, token, "/api/v1/baseFactory", map[string]any{"factoryCode": code + "B", "factoryName": "Query Beta", "status": 1})

	_, like := doJSON(http.MethodPost, "/api/v1/baseFactory/list", token, map[string]any{
		"page": 0, "limit": 20,
		"columns": []map[string]any{{"name": "factory_code", "exp": "like", "value": code, "logic": "and"}},
	})
	if like["code"] != float64(0) || int(like["data"].(map[string]any)["total"].(float64)) != 2 {
		t.Fatalf("like filter: %+v", like)
	}

	_, multi := doJSON(http.MethodPost, "/api/v1/baseFactory/list", token, map[string]any{
		"page": 0, "limit": 20,
		"columns": []map[string]any{
			{"name": "factory_code", "exp": "like", "value": code, "logic": "and"},
			{"name": "status", "exp": "in", "value": "1,2", "logic": "and"},
		},
	})
	if multi["code"] != float64(0) || int(multi["data"].(map[string]any)["total"].(float64)) != 2 {
		t.Fatalf("in filter: %+v", multi)
	}

	_, ignored := doJSON(http.MethodPost, "/api/v1/baseFactory/list", token, map[string]any{
		"page": 0, "limit": 5,
		"columns": []map[string]any{{"name": "password", "exp": "=", "value": "x", "logic": "and"}},
	})
	if ignored["code"] == float64(0) {
		t.Fatalf("unknown column should be rejected: %+v", ignored)
	}

	_, pref := doJSON(http.MethodPut, "/api/v1/sysPref/baseFactory", token, map[string]any{"payload": `{"hidden":["address"]}`})
	if pref["code"] != float64(0) {
		t.Fatalf("save pref: %+v", pref)
	}
	_, loaded := doJSON(http.MethodGet, "/api/v1/sysPref/baseFactory", token, nil)
	if loaded["data"].(map[string]any)["payload"] != `{"hidden":["address"]}` {
		t.Fatalf("load pref: %+v", loaded)
	}
	_, badKey := doJSON(http.MethodPut, "/api/v1/sysPref/bad!key", token, map[string]any{"payload": "{}"})
	if badKey["code"] == float64(0) {
		t.Fatalf("bad page key should fail: %+v", badKey)
	}

	_, disabled := doJSON(http.MethodPost, "/api/v1/batch", token, map[string]any{
		"resource": "baseFactory", "action": "disable", "ids": []uint64{mustNum(idA), mustNum(idB)},
	})
	if disabled["code"] != float64(0) || len(disabled["data"].(map[string]any)["ok"].([]any)) != 2 {
		t.Fatalf("batch disable: %+v", disabled)
	}
	_, onlyOff := doJSON(http.MethodPost, "/api/v1/baseFactory/list", token, map[string]any{
		"page": 0, "limit": 20,
		"columns": []map[string]any{
			{"name": "factory_code", "exp": "like", "value": code, "logic": "and"},
			{"name": "status", "exp": "=", "value": 2, "logic": "and"},
		},
	})
	if int(onlyOff["data"].(map[string]any)["total"].(float64)) != 2 {
		t.Fatalf("disabled rows: %+v", onlyOff)
	}

	_, removed := doJSON(http.MethodPost, "/api/v1/batch", token, map[string]any{
		"resource": "baseFactory", "action": "delete", "ids": []uint64{mustNum(idA), mustNum(idB)},
	})
	if removed["code"] != float64(0) {
		t.Fatalf("batch delete: %+v", removed)
	}

	_, protect := doJSON(http.MethodPost, "/api/v1/batch", token, map[string]any{
		"resource": "sysUser", "action": "delete", "ids": []uint64{1},
	})
	failed := protect["data"].(map[string]any)["failed"].([]any)
	if protect["code"] != float64(0) || len(failed) != 1 {
		t.Fatalf("protect admin: %+v", protect)
	}

	_, lots := doJSON(http.MethodPost, "/api/v1/wipLot/list", token, map[string]any{
		"page": 0, "limit": 5,
		"columns": []map[string]any{{"name": "lot_no", "exp": "like", "value": "DEMO-WIP-001", "logic": "and"}},
	})
	lotRows := lots["data"].(map[string]any)["wipLots"].([]any)
	if len(lotRows) == 0 {
		t.Fatalf("demo lot missing: %+v", lots)
	}
	lotID := uint64(lotRows[0].(map[string]any)["id"].(float64))
	_, held := doJSON(http.MethodPost, "/api/v1/batch", token, map[string]any{
		"resource": "wipLot", "action": "hold", "ids": []uint64{lotID}, "reasonCode": "ENG_HOLD", "reason": "batch",
	})
	if held["code"] != float64(0) || len(held["data"].(map[string]any)["ok"].([]any)) != 1 {
		t.Fatalf("batch hold: %+v", held)
	}
	_, released := doJSON(http.MethodPost, "/api/v1/batch", token, map[string]any{
		"resource": "wipLot", "action": "release", "ids": []uint64{lotID}, "reasonCode": "RELEASE", "reason": "batch",
	})
	if released["code"] != float64(0) || len(released["data"].(map[string]any)["ok"].([]any)) != 1 {
		t.Fatalf("batch release: %+v", released)
	}

	_, createdUser := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "lqsys", "password": "viewer123", "realName": "Sys", "status": 1,
	})
	userID := toID(createdUser["data"].(map[string]any)["id"])
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{2}})
	sysToken := login(t, "lqsys", "viewer123")
	_, denied := doJSON(http.MethodPost, "/api/v1/batch", sysToken, map[string]any{
		"resource": "baseFactory", "action": "delete", "ids": []uint64{1},
	})
	if denied["code"] != float64(40003) {
		t.Fatalf("sys admin batch should be forbidden: %+v", denied)
	}
}
