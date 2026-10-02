package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/logger"

	"semi-mes/server/internal/bootstrap"
	"semi-mes/server/internal/config"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/routers"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	dir, err := os.MkdirTemp("", "mes-flow")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	config.Set(&config.Config{
		App: config.App{Name: "mes", Env: "test", Host: "127.0.0.1"},
		Logger: config.Logger{Level: "error", Format: "console"},
		Database: config.Database{
			Driver: "sqlite",
			Sqlite: config.Sqlite{DBFile: filepath.Join(dir, "mes.db"), MaxIdleConns: 2, MaxOpenConns: 5, ConnMaxLifetime: 10},
		},
		Jwt: config.Jwt{SignKey: "mes-jwt-secret-key-change-in-production", ExpireHours: 24},
	})
	if _, err = logger.Init(logger.WithLevel("error"), logger.WithFormat("console")); err != nil {
		panic(err)
	}
	database.InitDB()
	if err = bootstrap.Seed(database.GetDB()); err != nil {
		panic(err)
	}
	testRouter = routers.NewRouter()
	os.Exit(m.Run())
}

func doJSON(method, path, token string, body any) (int, map[string]any) {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp == nil {
		resp = map[string]any{"_status": w.Code, "_raw": w.Body.String()}
	}
	return w.Code, resp
}

func login(t *testing.T, username, password string) string {
	t.Helper()
	_, resp := doJSON(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"userName": username, "password": password})
	if code, _ := resp["code"].(float64); code != 0 {
		t.Fatalf("login failed: %+v", resp)
	}
	data := resp["data"].(map[string]any)
	token, _ := data["token"].(string)
	if token == "" {
		t.Fatal("empty token")
	}
	return token
}

func TestAuthAndRBAC(t *testing.T) {
	_, bad := doJSON(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"userName": "admin", "password": "wrong"})
	if code, _ := bad["code"].(float64); code == 0 {
		t.Fatalf("expected invalid login, got %+v", bad)
	}

	token := login(t, "admin", "admin123")
	_, info := doJSON(http.MethodGet, "/api/v1/auth/getUserInfo", token, nil)
	data := info["data"].(map[string]any)
	if data["userName"] != "admin" {
		t.Fatalf("user info %+v", info)
	}
	buttons := data["buttons"].([]any)
	if len(buttons) == 0 {
		t.Fatal("expected button permissions")
	}
	_, routes := doJSON(http.MethodGet, "/api/v1/route/getUserRoutes", token, nil)
	routeData := routes["data"].(map[string]any)
	if routeData["home"] != "home" {
		t.Fatalf("routes %+v", routes)
	}
	if _, ok := routeData["routes"].([]any); !ok {
		t.Fatalf("routes missing: %+v", routes)
	}

	_, denied := doJSON(http.MethodPost, "/api/v1/baseFactory", "", map[string]any{"factoryCode": "X", "factoryName": "X", "status": 1})
	if denied["code"] == float64(0) {
		t.Fatalf("unauthenticated create should fail: %+v", denied)
	}

	_, createdUser := doJSON(http.MethodPost, "/api/v1/sysUser", token, map[string]any{
		"username": "viewer1", "password": "viewer123", "realName": "Viewer", "status": 1,
	})
	createdData, ok := createdUser["data"].(map[string]any)
	if !ok {
		t.Fatalf("create user response: %+v", createdUser)
	}
	userID := createdData["id"]
	_, _ = doJSON(http.MethodPut, "/api/v1/sysUser/"+toID(userID)+"/roles", token, map[string]any{"roleIds": []uint64{4}})
	viewerToken := login(t, "viewer1", "viewer123")
	_, forbidden := doJSON(http.MethodPost, "/api/v1/baseFactory", viewerToken, map[string]any{"factoryCode": "NO", "factoryName": "NO", "status": 1})
	if forbidden["code"] != float64(40003) {
		t.Fatalf("expected forbidden, got %+v", forbidden)
	}
}

func TestBaseDataCRUD(t *testing.T) {
	token := login(t, "admin", "admin123")

	factoryID := createEntity(t, token, "/api/v1/baseFactory", map[string]any{"factoryCode": "F1", "factoryName": "Fab", "status": 1})
	assertList(t, token, "/api/v1/baseFactory/list", "baseFactorys")
	updateEntity(t, token, "/api/v1/baseFactory/"+factoryID, map[string]any{"factoryCode": "F1", "factoryName": "Fab-2", "status": 1})

	workshopID := createEntity(t, token, "/api/v1/baseWorkshop", map[string]any{"factoryID": mustNum(factoryID), "workshopCode": "W1", "workshopName": "Wafer", "workshopType": "FAB", "status": 1})
	assertList(t, token, "/api/v1/baseWorkshop/list", "baseWorkshops")
	updateEntity(t, token, "/api/v1/baseWorkshop/"+workshopID, map[string]any{"factoryID": mustNum(factoryID), "workshopCode": "W1", "workshopName": "Wafer-2", "status": 1})

	lineID := createEntity(t, token, "/api/v1/baseProductionLine", map[string]any{"workshopID": mustNum(workshopID), "lineCode": "L1", "lineName": "Line", "capacity": 0, "status": 1})
	assertList(t, token, "/api/v1/baseProductionLine/list", "baseProductionLines")
	updateEntity(t, token, "/api/v1/baseProductionLine/"+lineID, map[string]any{"workshopID": mustNum(workshopID), "lineCode": "L1", "lineName": "Line-2", "capacity": 10, "status": 1})

	productID := createEntity(t, token, "/api/v1/baseProduct", map[string]any{"productCode": "P1", "productName": "Chip", "productType": "IC", "version": "A", "status": 1})
	assertList(t, token, "/api/v1/baseProduct/list", "baseProducts")
	updateEntity(t, token, "/api/v1/baseProduct/"+productID, map[string]any{"productCode": "P1", "productName": "Chip-2", "status": 1})

	routeID := createEntity(t, token, "/api/v1/baseProcessRoute", map[string]any{"productID": mustNum(productID), "routeCode": "R1", "routeName": "Route", "isDefault": 1, "status": 1})
	assertList(t, token, "/api/v1/baseProcessRoute/list", "baseProcessRoutes")
	updateEntity(t, token, "/api/v1/baseProcessRoute/"+routeID, map[string]any{"productID": mustNum(productID), "routeCode": "R1", "routeName": "Route-2", "isDefault": 0, "status": 1})

	opID := createEntity(t, token, "/api/v1/baseOperation", map[string]any{"routeID": mustNum(routeID), "operationCode": "OP1", "operationName": "Photo", "operationType": "PHOTO", "sequence": 1, "standardTime": 0, "status": 1})
	assertList(t, token, "/api/v1/baseOperation/list", "baseOperations")
	updateEntity(t, token, "/api/v1/baseOperation/"+opID, map[string]any{"routeID": mustNum(routeID), "operationCode": "OP1", "operationName": "Photo-2", "sequence": 2, "standardTime": 30, "status": 1})

	recipeID := createEntity(t, token, "/api/v1/baseRecipe", map[string]any{"operationID": mustNum(opID), "recipeCode": "RC1", "recipeName": "Recipe", "isDefault": 1, "parameters": "{}", "status": 1})
	assertList(t, token, "/api/v1/baseRecipe/list", "baseRecipes")
	updateEntity(t, token, "/api/v1/baseRecipe/"+recipeID, map[string]any{"operationID": mustNum(opID), "recipeCode": "RC1", "recipeName": "Recipe-2", "isDefault": 0, "status": 1})

	deleteEntity(t, token, "/api/v1/baseRecipe/"+recipeID)
	deleteEntity(t, token, "/api/v1/baseOperation/"+opID)
	deleteEntity(t, token, "/api/v1/baseProcessRoute/"+routeID)
	deleteEntity(t, token, "/api/v1/baseProduct/"+productID)
	deleteEntity(t, token, "/api/v1/baseProductionLine/"+lineID)
	deleteEntity(t, token, "/api/v1/baseWorkshop/"+workshopID)
	deleteEntity(t, token, "/api/v1/baseFactory/"+factoryID)
}

func createEntity(t *testing.T, token, path string, body map[string]any) string {
	t.Helper()
	_, resp := doJSON(http.MethodPost, path, token, body)
	if resp["code"] != float64(0) {
		t.Fatalf("create %s failed: %+v", path, resp)
	}
	id := resp["data"].(map[string]any)["id"]
	return toID(id)
}

func updateEntity(t *testing.T, token, path string, body map[string]any) {
	t.Helper()
	_, resp := doJSON(http.MethodPut, path, token, body)
	if resp["code"] != float64(0) {
		t.Fatalf("update %s failed: %+v", path, resp)
	}
}

func deleteEntity(t *testing.T, token, path string) {
	t.Helper()
	_, resp := doJSON(http.MethodDelete, path, token, nil)
	if resp["code"] != float64(0) {
		t.Fatalf("delete %s failed: %+v", path, resp)
	}
}

func assertList(t *testing.T, token, path, key string) {
	t.Helper()
	_, resp := doJSON(http.MethodPost, path, token, map[string]any{"page": 0, "limit": 10})
	if resp["code"] != float64(0) {
		t.Fatalf("list %s failed: %+v", path, resp)
	}
	data := resp["data"].(map[string]any)
	if _, ok := data[key]; !ok {
		t.Fatalf("list %s missing %s: %+v", path, key, data)
	}
}

func toID(v any) string {
	switch n := v.(type) {
	case float64:
		return strconvFormat(uint64(n))
	case string:
		return n
	default:
		return ""
	}
}

func mustNum(id string) uint64 {
	var n uint64
	for _, ch := range id {
		n = n*10 + uint64(ch-'0')
	}
	return n
}

func strconvFormat(n uint64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
