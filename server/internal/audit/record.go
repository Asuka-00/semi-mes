package audit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/jwt"
	"github.com/go-dev-frame/sponge/pkg/logger"

	"semi-mes/server/internal/database"
	"semi-mes/server/internal/model"
)

type captureWriter struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *captureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *captureWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w *captureWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

type classified struct {
	skip       bool
	module     string
	action     string
	entityType string
	table      string
	id         uint64
}

// Middleware records successful mutations and login attempts. Failures of the audit insert do not fail the request.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/v1/") {
			c.Next()
			return
		}
		body, _ := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		item := classify(c.Request.Method, path)
		if item.skip {
			c.Next()
			return
		}
		if item.action == "create" && item.entityType == "qcDefect" {
			item.action = "defect"
			if strings.EqualFold(stringField(body, "disposition"), "rework") {
				item.action = "rework"
			}
		}
		if item.action == "create" && item.entityType == "qcMeasurement" {
			item.action = "measure"
		}
		if item.id == 0 {
			item.id = idFromBody(body)
		}
		before := ""
		if item.table != "" && item.id > 0 && (c.Request.Method == http.MethodPut || c.Request.Method == http.MethodDelete || isMutationRead(item.action)) {
			before = loadRow(item.table, item.id)
		}
		writer := &captureWriter{ResponseWriter: c.Writer, status: http.StatusOK}
		c.Writer = writer
		c.Next()
		code := responseCode(writer.body.Bytes())
		if item.action == "login" && code != 0 {
			item.action = "login_failed"
		} else if code != 0 {
			return
		}
		if item.id == 0 {
			item.id = idFromBody(writer.body.Bytes())
		}
		after := ""
		if item.table != "" && item.id > 0 && item.action != "delete" {
			after = loadRow(item.table, item.id)
		}
		if after == "" && item.action != "delete" && item.action != "login" && item.action != "login_failed" && item.action != "logout" {
			after = clip(redactRaw(body), 8000)
		}
		if item.action == "login" || item.action == "login_failed" || item.action == "logout" {
			before, after = "", ""
		}
		userID, username := actor(c)
		if username == "" && (item.action == "login" || item.action == "login_failed") {
			username = stringField(body, "userName", "username")
		}
		if (item.action == "login" || item.action == "logout") && item.id == 0 && username != "" {
			item.id = userIDByName(username)
		}
		if username == "" && userID > 0 {
			username = usernameOf(userID)
		}
		codeName := entityCode(before, after)
		if codeName == "" {
			codeName = nestedString(writer.body.Bytes(), "lotNo")
		}
		if codeName == "" && (item.action == "login" || item.action == "login_failed" || item.action == "logout") {
			codeName = username
		}
		summary := item.module + " " + item.action
		if lotNo := nestedString(writer.body.Bytes(), "lotNo"); lotNo != "" && item.action == "start" {
			summary = "lot start " + lotNo
		}
		now := time.Now()
		row := model.SysAuditLog{
			UserID: userID, Username: username, IP: c.ClientIP(),
			Module: item.module, Action: item.action, EntityType: item.entityType, EntityID: item.id,
			EntityCode: codeName, Summary: summary,
			BeforeJSON: before, AfterJSON: after, DiffJSON: diffJSON(before, after), CreatedAt: &now,
		}
		if err := database.GetDB().Create(&row).Error; err != nil {
			logger.Warn("audit insert failed", logger.Err(err))
		}
	}
}

func isMutationRead(action string) bool {
	switch action {
	case "hold", "release", "split", "merge", "track_in", "track_out", "abort", "state", "start", "close", "release_order", "complete", "pass", "advance", "defect", "rework", "measure":
		return true
	default:
		return false
	}
}

func classify(method, path string) classified {
	if strings.Contains(path, "/list") || strings.Contains(path, "/sysPref") || strings.Contains(path, "/dashboard") ||
		strings.Contains(path, "/trace") || strings.Contains(path, "/getUserInfo") || strings.Contains(path, "/getUserRoutes") ||
		strings.Contains(path, "/isRouteExist") || strings.Contains(path, "/getConstantRoutes") || strings.HasSuffix(path, "/features") ||
		strings.Contains(path, "/sysNotice") || strings.Contains(path, "/refreshToken") {
		return classified{skip: true}
	}
	if path == "/api/v1/auth/login" {
		return classified{module: "auth", action: "login", entityType: "user"}
	}
	if path == "/api/v1/auth/logout" {
		return classified{module: "auth", action: "logout", entityType: "user"}
	}
	if method == http.MethodGet {
		return classified{skip: true}
	}
	resource := strings.TrimPrefix(path, "/api/v1/")
	parts := strings.Split(resource, "/")
	name := parts[0]
	module, table := resourceMeta(name)
	action := methodAction(method)
	id := uint64(0)
	for i, part := range parts {
		if n, err := strconv.ParseUint(part, 10, 64); err == nil {
			id = n
			if i+1 < len(parts) {
				action = suffixAction(parts[i+1], action)
			}
		}
	}
	if len(parts) > 1 && id == 0 {
		action = suffixAction(parts[len(parts)-1], action)
	}
	if action == "" {
		return classified{skip: true}
	}
	return classified{module: module, action: action, entityType: name, table: table, id: id}
}

func methodAction(method string) string {
	switch method {
	case http.MethodPost:
		return "create"
	case http.MethodPut:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return ""
	}
}

func suffixAction(suffix, fallback string) string {
	switch suffix {
	case "hold":
		return "hold"
	case "releaseHold":
		return "release"
	case "split":
		return "split"
	case "merge":
		return "merge"
	case "trackIn":
		return "track_in"
	case "trackOut":
		return "track_out"
	case "abort":
		return "abort"
	case "state":
		return "state"
	case "start":
		return "start"
	case "close":
		return "close"
	case "release":
		return "release_order"
	case "complete":
		return "complete"
	case "pass":
		return "pass"
	case "advance":
		return "advance"
	case "batch":
		return "batch"
	case "generate":
		return "generate"
	default:
		if suffix == "" || strings.EqualFold(suffix, fallback) {
			return fallback
		}
		if _, err := strconv.ParseUint(suffix, 10, 64); err == nil {
			return fallback
		}
		return fallback
	}
}

func resourceMeta(name string) (string, string) {
	switch name {
	case "baseFactory":
		return "base", "base_factory"
	case "baseWorkshop":
		return "base", "base_workshop"
	case "baseProductionLine":
		return "base", "base_production_line"
	case "baseProduct":
		return "base", "base_product"
	case "baseProcessRoute":
		return "base", "base_process_route"
	case "baseOperation":
		return "base", "base_operation"
	case "baseRecipe":
		return "base", "base_recipe"
	case "wipWorkOrder":
		return "wo", "wip_work_order"
	case "wipLot":
		return "lot", "wip_lot"
	case "wipMove":
		return "wip", "wip_lot"
	case "eqpEquipment":
		return "eqp", "eqp_equipment"
	case "eqpPmPlan":
		return "eqp", "eqp_pm_plan"
	case "eqpPmTask":
		return "eqp", "eqp_pm_task"
	case "qcInspectPlan":
		return "qc", "qc_inspect_plan"
	case "qcMeasurement":
		return "qc", "wip_lot"
	case "qcDefect":
		return "qc", "wip_lot"
	case "qcDefectCode":
		return "qc", "qc_defect_code"
	case "sysUser":
		return "system", "sys_user"
	case "sysRole":
		return "system", "sys_role"
	case "sysMenu":
		return "system", "sys_menu"
	case "sysDictType":
		return "system", "sys_dict_type"
	case "sysDictItem":
		return "system", "sys_dict_item"
	case "sysReasonCode":
		return "system", "mes_reason_code"
	case "sysNumberRule":
		return "system", "sys_number_rule"
	default:
		return name, ""
	}
}

func loadRow(table string, id uint64) string {
	if table == "" || id == 0 {
		return ""
	}
	var rows []map[string]any
	err := database.GetDB().Table(table).Where("id = ?", id).Limit(1).Find(&rows).Error
	if err != nil || len(rows) == 0 {
		return ""
	}
	redactMap(rows[0])
	raw, err := json.Marshal(rows[0])
	if err != nil {
		return ""
	}
	return clip(string(raw), 8000)
}

func diffJSON(before, after string) string {
	if before == "" && after == "" {
		return ""
	}
	var left, right map[string]any
	_ = json.Unmarshal([]byte(before), &left)
	_ = json.Unmarshal([]byte(after), &right)
	if left == nil {
		left = map[string]any{}
	}
	if right == nil {
		right = map[string]any{}
	}
	changed := map[string]any{}
	seen := map[string]struct{}{}
	for key := range left {
		seen[key] = struct{}{}
	}
	for key := range right {
		seen[key] = struct{}{}
	}
	for key := range seen {
		if key == "updated_at" || key == "created_at" || key == "password" {
			continue
		}
		if reflect.DeepEqual(left[key], right[key]) {
			continue
		}
		changed[key] = map[string]any{"from": left[key], "to": right[key]}
	}
	if len(changed) == 0 {
		return ""
	}
	raw, err := json.Marshal(changed)
	if err != nil {
		return ""
	}
	return clip(string(raw), 8000)
}

func responseCode(body []byte) int {
	var payload struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return -1
	}
	return payload.Code
}

func idFromBody(body []byte) uint64 {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0
	}
	if data, ok := payload["data"].(map[string]any); ok {
		if id := anyID(data["id"]); id > 0 {
			return id
		}
	}
	for _, key := range []string{"id", "lotId", "lotID", "equipmentID", "equipmentId"} {
		if id := anyID(payload[key]); id > 0 {
			return id
		}
	}
	return 0
}

func anyID(value any) uint64 {
	switch n := value.(type) {
	case float64:
		return uint64(n)
	case string:
		id, _ := strconv.ParseUint(n, 10, 64)
		return id
	default:
		return 0
	}
}

func nestedString(body []byte, key string) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if data, ok := payload["data"].(map[string]any); ok {
		if text, ok := data[key].(string); ok {
			return text
		}
	}
	if text, ok := payload[key].(string); ok {
		return text
	}
	return ""
}

func userIDByName(username string) uint64 {
	var id uint64
	_ = database.GetDB().Table("sys_user").Where("username = ?", username).Pluck("id", &id).Error
	return id
}

func stringField(body []byte, keys ...string) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	for _, key := range keys {
		if text, ok := payload[key].(string); ok && text != "" {
			return text
		}
	}
	return ""
}

func actor(c *gin.Context) (uint64, string) {
	raw, ok := c.Get("claims")
	if !ok {
		return 0, ""
	}
	claims, ok := raw.(*jwt.Claims)
	if !ok || claims == nil {
		return 0, ""
	}
	id, err := strconv.ParseUint(claims.UID, 10, 64)
	if err != nil {
		return 0, ""
	}
	return id, usernameOf(id)
}

func usernameOf(id uint64) string {
	var name string
	_ = database.GetDB().Table("sys_user").Where("id = ?", id).Pluck("username", &name).Error
	return name
}

func entityCode(before, after string) string {
	for _, raw := range []string{after, before} {
		var row map[string]any
		if json.Unmarshal([]byte(raw), &row) != nil {
			continue
		}
		for _, key := range []string{"lot_no", "order_no", "factory_code", "equipment_code", "username", "product_code", "defect_code", "role_code", "recipe_code", "workshop_code", "line_code", "plan_name", "task_no", "type_code", "item_code", "reason_code", "rule_code"} {
			if text, ok := row[key].(string); ok && text != "" {
				return text
			}
		}
	}
	return ""
}

func redactRaw(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	redactAny(payload)
	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(raw)
}

func redactMap(row map[string]any) {
	for key := range row {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "password") || lower == "token" || lower == "refreshtoken" {
			row[key] = "***"
		}
	}
}

func redactAny(value any) {
	switch row := value.(type) {
	case map[string]any:
		redactMap(row)
		for _, child := range row {
			redactAny(child)
		}
	case []any:
		for _, child := range row {
			redactAny(child)
		}
	}
}

func clip(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}
