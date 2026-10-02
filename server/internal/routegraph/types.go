// Package routegraph evaluates process-route flowcharts: conditions, release checks, and the next step.
package routegraph

// Node and edge kinds stored on a route version.
const (
	NodeStart     = "start"
	NodeEnd       = "end"
	NodeOperation = "operation"
	NodeDecision  = "decision"

	EdgeNormal = "normal"
	EdgeRework = "rework"

	StateDraft    = "draft"
	StateReleased = "released"
	StateObsolete = "obsolete"

	ActionMove = "move"
	ActionHold = "hold"
	ActionEnd  = "end"

	ReasonMatched        = "matched"
	ReasonDefault        = "default"
	ReasonReworkExceeded = "rework_exceeded"
	ReasonAlreadyEnd     = "already_end"

	OnExceedHold = "hold"
)

// Node is one vertex of a route version.
type Node struct {
	Key            string
	Type           string
	Name           string
	OperationID    uint64
	RecipeID       uint64
	EquipmentGroup string
	PosX           float64
	PosY           float64
}

// Edge connects two nodes. A non-default edge carries a declarative condition.
type Edge struct {
	Key              string
	From             string
	To               string
	Kind             string
	IsDefault        bool
	Priority         int
	Condition        *Condition
	ConditionInvalid bool
	MaxRework        int
	OnExceed         string
	Label            string
}

// Graph is the directed flowchart of one route version.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Condition is a whitelist predicate group. Exactly one of All or Any is set.
// Empty means the edge has no condition (only legal on the default edge).
type Condition struct {
	All []Predicate `json:"all,omitempty"`
	Any []Predicate `json:"any,omitempty"`
}

// Predicate compares one whitelisted field. No code is executed.
type Predicate struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}

// LotContext is the snapshot available when a lot leaves a step.
type LotContext struct {
	InspectionResult string
	InspectionGrade  string
	DefectCode       string
	ProductCode      string
	Priority         float64
	LotType          string
	ReworkCounts     map[string]int
}

// Result is the next-step decision for one track-out.
type Result struct {
	Action       string `json:"action"`
	NextNodeKey  string `json:"nextNodeKey,omitempty"`
	NextNodeType string `json:"nextNodeType,omitempty"`
	NextNodeName string `json:"nextNodeName,omitempty"`
	EdgeKey      string `json:"edgeKey,omitempty"`
	Reason       string `json:"reason"`
}

// Issue is one release-check failure.
type Issue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Ref     string `json:"ref,omitempty"`
}
