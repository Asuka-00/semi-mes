package routegraph

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	fieldKindString = "string"
	fieldKindNumber = "number"
)

var fieldKind = map[string]string{
	"inspection.result": fieldKindString,
	"inspection.grade":  fieldKindString,
	"defect.code":       fieldKindString,
	"lot.productCode":   fieldKindString,
	"lot.priority":      fieldKindNumber,
	"lot.type":          fieldKindString,
	"rework.count":      fieldKindNumber,
}

var stringOps = map[string]bool{"eq": true, "ne": true, "in": true, "contains": true}
var numberOps = map[string]bool{"eq": true, "ne": true, "gt": true, "gte": true, "lt": true, "lte": true, "in": true}

// Empty reports whether the condition carries no predicates.
func (c *Condition) Empty() bool {
	if c == nil {
		return true
	}
	return len(c.All) == 0 && len(c.Any) == 0
}

// ParseCondition decodes a stored JSON object. Blank text is an empty condition.
func ParseCondition(raw string) (*Condition, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" || raw == "{}" {
		return nil, nil
	}
	var c Condition
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, err
	}
	if c.Empty() {
		return nil, nil
	}
	return &c, nil
}

// MustJSON returns the stored form of a condition.
func MustJSON(c *Condition) string {
	if c == nil || c.Empty() {
		return ""
	}
	buf, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(buf)
}

func conditionIssues(edge Edge) []Issue {
	c := edge.Condition
	if edge.ConditionInvalid {
		return []Issue{{Code: "bad_condition_shape", Ref: edge.Key}}
	}
	if edge.IsDefault {
		if !c.Empty() {
			return []Issue{{Code: "condition_on_default", Ref: edge.Key}}
		}
		return nil
	}
	if c.Empty() {
		return []Issue{{Code: "missing_condition", Ref: edge.Key}}
	}
	if len(c.All) > 0 && len(c.Any) > 0 {
		return []Issue{{Code: "bad_condition_shape", Ref: edge.Key}}
	}
	preds := c.All
	if len(c.Any) > 0 {
		preds = c.Any
	}
	var issues []Issue
	for _, p := range preds {
		kind, ok := fieldKind[p.Field]
		if !ok {
			issues = append(issues, Issue{Code: "bad_field", Ref: edge.Key})
			continue
		}
		ops := stringOps
		if kind == fieldKindNumber {
			ops = numberOps
		}
		if !ops[p.Op] {
			issues = append(issues, Issue{Code: "bad_op", Ref: edge.Key})
			continue
		}
		if !validValue(kind, p.Value, p.Op) {
			issues = append(issues, Issue{Code: "bad_value", Ref: edge.Key})
		}
	}
	return issues
}

func validValue(kind string, value any, op string) bool {
	if op == "in" {
		arr, ok := toSlice(value)
		if !ok || len(arr) == 0 {
			return false
		}
		for _, item := range arr {
			if !validScalar(kind, item) {
				return false
			}
		}
		return true
	}
	return validScalar(kind, value)
}

func validScalar(kind string, value any) bool {
	if kind == fieldKindNumber {
		_, ok := asFloat(value)
		return ok
	}
	_, ok := value.(string)
	return ok
}

func toSlice(value any) ([]any, bool) {
	switch v := value.(type) {
	case []any:
		return v, true
	case []string:
		out := make([]any, len(v))
		for i := range v {
			out[i] = v[i]
		}
		return out, true
	default:
		return nil, false
	}
}

func asFloat(value any) (float64, bool) {
	switch n := value.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func (c *Condition) matches(ctx LotContext, edgeKey string) bool {
	if c.Empty() {
		return false
	}
	preds := c.All
	all := true
	if len(c.Any) > 0 {
		preds = c.Any
		all = false
	}
	if all {
		for _, p := range preds {
			if !p.matches(ctx, edgeKey) {
				return false
			}
		}
		return true
	}
	for _, p := range preds {
		if p.matches(ctx, edgeKey) {
			return true
		}
	}
	return false
}

func (p Predicate) matches(ctx LotContext, edgeKey string) bool {
	actual := fieldValue(ctx, edgeKey, p.Field)
	switch p.Op {
	case "eq":
		return scalarEqual(actual, p.Value)
	case "ne":
		return !scalarEqual(actual, p.Value)
	case "contains":
		return strings.Contains(fmt.Sprint(actual), fmt.Sprint(p.Value))
	case "in":
		arr, ok := toSlice(p.Value)
		if !ok {
			return false
		}
		for _, item := range arr {
			if scalarEqual(actual, item) {
				return true
			}
		}
		return false
	case "gt", "gte", "lt", "lte":
		left, lok := asFloat(actual)
		right, rok := asFloat(p.Value)
		if !lok || !rok {
			return false
		}
		switch p.Op {
		case "gt":
			return left > right
		case "gte":
			return left >= right
		case "lt":
			return left < right
		default:
			return left <= right
		}
	default:
		return false
	}
}

func scalarEqual(actual, expected any) bool {
	af, aok := asFloat(actual)
	bf, bok := asFloat(expected)
	if aok && bok {
		return af == bf
	}
	return fmt.Sprint(actual) == fmt.Sprint(expected)
}

func fieldValue(ctx LotContext, edgeKey, field string) any {
	switch field {
	case "inspection.result":
		return ctx.InspectionResult
	case "inspection.grade":
		return ctx.InspectionGrade
	case "defect.code":
		return ctx.DefectCode
	case "lot.productCode":
		return ctx.ProductCode
	case "lot.priority":
		return ctx.Priority
	case "lot.type":
		return ctx.LotType
	case "rework.count":
		if ctx.ReworkCounts == nil {
			return 0
		}
		return ctx.ReworkCounts[edgeKey]
	default:
		return nil
	}
}
