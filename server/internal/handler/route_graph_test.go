package handler_test

import (
	"net/http"
	"testing"

	"semi-mes/server/internal/routegraph"
)

func TestRouteGraphFlow(t *testing.T) {
	token := login(t, "admin", "admin123")
	productID := createEntity(t, token, "/api/v1/baseProduct", map[string]any{
		"productCode": "P-GRAPH", "productName": "Graph", "productType": "IC", "status": 1,
	})
	ops := map[string]uint64{}
	for _, code := range []string{"CLEAN", "PHOTO", "INSPECT", "ETCH", "ENG_REVIEW"} {
		id := createEntity(t, token, "/api/v1/baseOperation", map[string]any{
			"operationCode": code + "-G", "operationName": code, "operationType": code, "status": 1,
		})
		ops[code] = mustNum(id)
	}
	routeID := createEntity(t, token, "/api/v1/baseProcessRoute", map[string]any{
		"productID": mustNum(productID), "routeCode": "R-GRAPH", "routeName": "Graph", "isDefault": 0, "status": 1,
	})

	_, created := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions", token, map[string]any{"note": "draft"})
	if created["code"] != float64(0) {
		t.Fatalf("create version: %+v", created)
	}
	version := created["data"].(map[string]any)["version"].(map[string]any)
	versionID := toID(version["id"])
	if version["state"] != "draft" {
		t.Fatalf("state %+v", version)
	}

	graph := routegraph.Sample(ops)
	_, saved := doJSON(http.MethodPut, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+versionID+"/graph", token, graphBody(graph))
	if saved["code"] != float64(0) {
		t.Fatalf("save: %+v", saved)
	}

	_, validated := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+versionID+"/validate", token, map[string]any{})
	if validated["code"] != float64(0) || validated["data"].(map[string]any)["valid"] != true {
		t.Fatalf("validate: %+v", validated)
	}

	resolve := func(body map[string]any) map[string]any {
		t.Helper()
		_, resp := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+versionID+"/resolve", token, body)
		if resp["code"] != float64(0) {
			t.Fatalf("resolve %+v -> %+v", body, resp)
		}
		return resp["data"].(map[string]any)
	}
	fail := resolve(map[string]any{"currentNodeKey": routegraph.KeyDecide, "context": map[string]any{"inspection": map[string]any{"result": "fail"}, "lot": map[string]any{"type": "production"}}})
	if fail["action"] != "move" || fail["nextNodeKey"] != routegraph.KeyPhoto || fail["reason"] != "matched" {
		t.Fatalf("fail: %+v", fail)
	}
	eng := resolve(map[string]any{"currentNodeKey": routegraph.KeyDecide, "context": map[string]any{"inspection": map[string]any{"result": "pass"}, "lot": map[string]any{"type": "engineering"}}})
	if eng["nextNodeKey"] != routegraph.KeyEng {
		t.Fatalf("eng: %+v", eng)
	}
	def := resolve(map[string]any{"currentNodeKey": routegraph.KeyDecide, "context": map[string]any{"inspection": map[string]any{"result": "pass"}, "lot": map[string]any{"type": "production"}}})
	if def["nextNodeKey"] != routegraph.KeyEtch || def["reason"] != "default" {
		t.Fatalf("default: %+v", def)
	}
	held := resolve(map[string]any{
		"currentNodeKey": routegraph.KeyDecide,
		"context": map[string]any{
			"inspection":   map[string]any{"result": "fail"},
			"reworkCounts": map[string]any{routegraph.EdgeReworkPhoto: 2},
		},
	})
	if held["action"] != "hold" || held["reason"] != "rework_exceeded" {
		t.Fatalf("hold: %+v", held)
	}

	_, released := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+versionID+"/release", token, map[string]any{})
	if released["code"] != float64(0) || released["data"].(map[string]any)["released"] != true {
		t.Fatalf("release: %+v", released)
	}
	_, blocked := doJSON(http.MethodPut, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+versionID+"/graph", token, graphBody(graph))
	if blocked["code"] == float64(0) {
		t.Fatalf("released graph must be immutable: %+v", blocked)
	}

	_, copied := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions", token, map[string]any{"copyFrom": mustNum(versionID)})
	if copied["code"] != float64(0) {
		t.Fatalf("copy: %+v", copied)
	}
	copyID := toID(copied["data"].(map[string]any)["version"].(map[string]any)["id"])
	_, got := doJSON(http.MethodGet, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+copyID, token, nil)
	nodes := got["data"].(map[string]any)["nodes"].([]any)
	if len(nodes) != len(graph.Nodes) {
		t.Fatalf("copied nodes %+v", got)
	}

	userID := createEntity(t, token, "/api/v1/sysUser", map[string]any{
		"username": "viewergraph", "password": "viewer123", "realName": "Viewer", "status": 1,
	})
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+userID+"/roles", token, map[string]any{"roleIds": []uint64{4}})
	viewer := login(t, "viewergraph", "viewer123")
	_, forbidden := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+copyID+"/release", viewer, map[string]any{})
	if forbidden["code"] != float64(40003) {
		t.Fatalf("viewer release: %+v", forbidden)
	}
	_, allowed := doJSON(http.MethodPost, "/api/v1/baseProcessRoute/"+routeID+"/versions/"+copyID+"/validate", viewer, map[string]any{})
	if allowed["code"] != float64(0) {
		t.Fatalf("viewer validate: %+v", allowed)
	}
}

func graphBody(g routegraph.Graph) map[string]any {
	nodes := make([]map[string]any, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		nodes = append(nodes, map[string]any{
			"nodeKey": n.Key, "nodeType": n.Type, "name": n.Name, "operationID": n.OperationID,
			"recipeID": n.RecipeID, "equipmentGroup": n.EquipmentGroup, "posX": n.PosX, "posY": n.PosY,
		})
	}
	edges := make([]map[string]any, 0, len(g.Edges))
	for _, e := range g.Edges {
		edges = append(edges, map[string]any{
			"edgeKey": e.Key, "fromKey": e.From, "toKey": e.To, "edgeKind": e.Kind, "isDefault": e.IsDefault,
			"priority": e.Priority, "condition": e.Condition, "maxRework": e.MaxRework, "onExceed": e.OnExceed, "label": e.Label,
		})
	}
	return map[string]any{"nodes": nodes, "edges": edges}
}
