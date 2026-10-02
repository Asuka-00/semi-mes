package routegraph

import "testing"

func sample() Graph {
	return Sample(map[string]uint64{"CLEAN": 1, "PHOTO": 2, "INSPECT": 3, "ETCH": 4, "ENG_REVIEW": 5})
}

func TestSampleValidates(t *testing.T) {
	issues := Validate(sample())
	if len(issues) > 0 {
		t.Fatalf("sample issues: %+v", issues)
	}
}

func TestResolveBranches(t *testing.T) {
	g := sample()

	fail := LotContext{InspectionResult: "fail", LotType: "production"}
	got, err := Resolve(g, KeyDecide, fail)
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != ActionMove || got.NextNodeKey != KeyPhoto || got.EdgeKey != EdgeReworkPhoto || got.Reason != ReasonMatched {
		t.Fatalf("fail branch: %+v", got)
	}

	both := LotContext{InspectionResult: "fail", LotType: "engineering"}
	got, err = Resolve(g, KeyDecide, both)
	if err != nil {
		t.Fatal(err)
	}
	if got.NextNodeKey != KeyPhoto || got.Reason != ReasonMatched {
		t.Fatalf("priority should prefer rework: %+v", got)
	}

	eng := LotContext{InspectionResult: "pass", LotType: "engineering"}
	got, err = Resolve(g, KeyDecide, eng)
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != ActionMove || got.NextNodeKey != KeyEng || got.EdgeKey != EdgeEng || got.Reason != ReasonMatched {
		t.Fatalf("engineering branch: %+v", got)
	}

	prod := LotContext{InspectionResult: "pass", LotType: "production"}
	got, err = Resolve(g, KeyDecide, prod)
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != ActionMove || got.NextNodeKey != KeyEtch || got.EdgeKey != EdgeDefaultEtch || got.Reason != ReasonDefault {
		t.Fatalf("default path: %+v", got)
	}
}

func TestResolveReworkLimit(t *testing.T) {
	g := sample()
	ctx := LotContext{InspectionResult: "fail", LotType: "production", ReworkCounts: map[string]int{EdgeReworkPhoto: 1}}
	got, err := Resolve(g, KeyDecide, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != ActionMove || got.NextNodeKey != KeyPhoto {
		t.Fatalf("under limit: %+v", got)
	}

	ctx.ReworkCounts[EdgeReworkPhoto] = 2
	got, err = Resolve(g, KeyDecide, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != ActionHold || got.Reason != ReasonReworkExceeded || got.NextNodeKey != "" {
		t.Fatalf("max rework: %+v", got)
	}
}

func TestResolveEndAndMissing(t *testing.T) {
	g := sample()
	got, err := Resolve(g, KeyEndMain, LotContext{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Action != ActionEnd || got.Reason != ReasonAlreadyEnd {
		t.Fatalf("end: %+v", got)
	}
	if _, err = Resolve(g, "missing", LotContext{}); err != ErrNodeNotFound {
		t.Fatalf("missing node err=%v", err)
	}
}

func TestConditionOperators(t *testing.T) {
	g := Graph{
		Nodes: []Node{
			{Key: "start", Type: NodeStart},
			{Key: "a", Type: NodeOperation, OperationID: 1},
			{Key: "b", Type: NodeOperation, OperationID: 2},
			{Key: "end", Type: NodeEnd},
		},
		Edges: []Edge{
			{Key: "s", From: "start", To: "a", Kind: EdgeNormal, IsDefault: true},
			{
				Key: "hi", From: "a", To: "b", Kind: EdgeNormal, Priority: 1,
				Condition: &Condition{Any: []Predicate{
					{Field: "lot.priority", Op: "gte", Value: 3},
					{Field: "defect.code", Op: "in", Value: []any{"D1", "D2"}},
				}},
			},
			{Key: "def", From: "a", To: "end", Kind: EdgeNormal, IsDefault: true, Priority: 10},
			{Key: "be", From: "b", To: "end", Kind: EdgeNormal, IsDefault: true},
		},
	}
	// decision-like branch on an operation is allowed; validate separately
	got, err := Resolve(g, "a", LotContext{Priority: 4})
	if err != nil || got.NextNodeKey != "b" || got.Reason != ReasonMatched {
		t.Fatalf("priority gte: %+v %v", got, err)
	}
	got, err = Resolve(g, "a", LotContext{DefectCode: "D2", Priority: 1})
	if err != nil || got.NextNodeKey != "b" {
		t.Fatalf("in: %+v %v", got, err)
	}
	got, err = Resolve(g, "a", LotContext{DefectCode: "DX", Priority: 1})
	if err != nil || got.NextNodeKey != "end" || got.Reason != ReasonDefault {
		t.Fatalf("else: %+v %v", got, err)
	}

	contains := g
	contains.Edges = append([]Edge{}, g.Edges...)
	contains.Edges[1].Condition = &Condition{All: []Predicate{{Field: "lot.productCode", Op: "contains", Value: "CMOS"}}}
	got, err = Resolve(contains, "a", LotContext{ProductCode: "DEMO-CMOS-1"})
	if err != nil || got.NextNodeKey != "b" {
		t.Fatalf("contains: %+v %v", got, err)
	}
}

func TestValidateRejectsUnsafeGraphs(t *testing.T) {
	assertHas := func(g Graph, code string) {
		t.Helper()
		for _, issue := range Validate(g) {
			if issue.Code == code {
				return
			}
		}
		t.Fatalf("expected %s in %+v", code, Validate(g))
	}

	assertHas(Graph{Nodes: []Node{{Key: "end", Type: NodeEnd}}}, "start_count")

	assertHas(Graph{
		Nodes: []Node{{Key: "s", Type: NodeStart}, {Key: "a", Type: NodeOperation, OperationID: 1}, {Key: "e", Type: NodeEnd}},
		Edges: []Edge{
			{Key: "1", From: "s", To: "a", Kind: EdgeNormal, IsDefault: true},
			{Key: "2", From: "a", To: "e", Kind: EdgeNormal, IsDefault: true},
			{Key: "3", From: "a", To: "e", Kind: EdgeNormal, IsDefault: true, Condition: &Condition{All: []Predicate{{Field: "lot.type", Op: "eq", Value: "production"}}}},
		},
	}, "multiple_default")

	assertHas(Graph{
		Nodes: []Node{{Key: "s", Type: NodeStart}, {Key: "d", Type: NodeDecision}, {Key: "e", Type: NodeEnd}},
		Edges: []Edge{
			{Key: "1", From: "s", To: "d", Kind: EdgeNormal, IsDefault: true},
			{Key: "2", From: "d", To: "e", Kind: EdgeNormal, IsDefault: true},
		},
	}, "decision_outgoing")

	forward := Graph{
		Nodes: []Node{
			{Key: "s", Type: NodeStart},
			{Key: "a", Type: NodeOperation, OperationID: 1},
			{Key: "b", Type: NodeOperation, OperationID: 2},
			{Key: "e", Type: NodeEnd},
		},
		Edges: []Edge{
			{Key: "1", From: "s", To: "a", Kind: EdgeNormal, IsDefault: true},
			{Key: "2", From: "a", To: "b", Kind: EdgeNormal, IsDefault: true},
			{Key: "3", From: "b", To: "e", Kind: EdgeNormal, IsDefault: true},
			{Key: "rw", From: "a", To: "b", Kind: EdgeRework, MaxRework: 1, Condition: &Condition{All: []Predicate{{Field: "inspection.result", Op: "eq", Value: "fail"}}}},
		},
	}
	assertHas(forward, "rework_not_upstream")

	loop := Graph{
		Nodes: []Node{{Key: "s", Type: NodeStart}, {Key: "a", Type: NodeOperation, OperationID: 1}, {Key: "e", Type: NodeEnd}},
		Edges: []Edge{
			{Key: "1", From: "s", To: "a", Kind: EdgeNormal, IsDefault: true},
			{Key: "2", From: "a", To: "a", Kind: EdgeNormal, IsDefault: true},
			{Key: "3", From: "a", To: "e", Kind: EdgeNormal},
		},
	}
	assertHas(loop, "unbounded_cycle")

	badField := sample()
	badField.Edges[4].Condition = &Condition{All: []Predicate{{Field: "lot.script", Op: "eq", Value: "x"}}}
	assertHas(badField, "bad_field")

	noOp := sample()
	noOp.Nodes[1].OperationID = 0
	assertHas(noOp, "operation_missing")
}

func TestReworkCountCondition(t *testing.T) {
	g := Graph{
		Nodes: []Node{
			{Key: "s", Type: NodeStart},
			{Key: "a", Type: NodeOperation, OperationID: 1},
			{Key: "b", Type: NodeOperation, OperationID: 1},
			{Key: "e", Type: NodeEnd},
		},
		Edges: []Edge{
			{Key: "1", From: "s", To: "a", Kind: EdgeNormal, IsDefault: true},
			{Key: "2", From: "a", To: "b", Kind: EdgeNormal, IsDefault: true},
			{
				Key: "rw", From: "b", To: "a", Kind: EdgeRework, Priority: 1, MaxRework: 3, OnExceed: OnExceedHold,
				Condition: &Condition{All: []Predicate{
					{Field: "inspection.result", Op: "eq", Value: "fail"},
					{Field: "rework.count", Op: "lt", Value: 2},
				}},
			},
			{Key: "def", From: "b", To: "e", Kind: EdgeNormal, IsDefault: true, Priority: 10},
		},
	}
	if issues := Validate(g); len(issues) > 0 {
		t.Fatalf("expected valid rework loop: %+v", issues)
	}
	got, err := Resolve(g, "b", LotContext{InspectionResult: "fail", ReworkCounts: map[string]int{"rw": 1}})
	if err != nil || got.NextNodeKey != "a" {
		t.Fatalf("count 1: %+v %v", got, err)
	}
	got, err = Resolve(g, "b", LotContext{InspectionResult: "fail", ReworkCounts: map[string]int{"rw": 2}})
	if err != nil || got.NextNodeKey != "e" || got.Reason != ReasonDefault {
		t.Fatalf("count blocks condition, default continues: %+v %v", got, err)
	}
}
