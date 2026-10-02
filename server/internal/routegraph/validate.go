package routegraph

// Validate checks a route version before it can be released.
// A lot is one physical unit, so the graph has no parallel split or merge.
func Validate(g Graph) []Issue {
	var issues []Issue
	nodes := map[string]Node{}
	for _, n := range g.Nodes {
		if n.Key == "" {
			issues = append(issues, Issue{Code: "empty_key"})
			continue
		}
		if _, ok := nodes[n.Key]; ok {
			issues = append(issues, Issue{Code: "duplicate_node", Ref: n.Key})
			continue
		}
		switch n.Type {
		case NodeStart, NodeEnd, NodeOperation, NodeDecision:
		default:
			issues = append(issues, Issue{Code: "unknown_node_type", Ref: n.Key})
		}
		if n.Type == NodeOperation && n.OperationID == 0 {
			issues = append(issues, Issue{Code: "operation_missing", Ref: n.Key})
		}
		nodes[n.Key] = n
	}

	edges := map[string]Edge{}
	out := map[string][]Edge{}
	in := map[string]int{}
	for _, e := range g.Edges {
		if e.Key == "" {
			issues = append(issues, Issue{Code: "empty_key"})
			continue
		}
		if _, ok := edges[e.Key]; ok {
			issues = append(issues, Issue{Code: "duplicate_edge", Ref: e.Key})
			continue
		}
		edges[e.Key] = e
		if e.Kind != EdgeNormal && e.Kind != EdgeRework {
			issues = append(issues, Issue{Code: "unknown_edge_kind", Ref: e.Key})
		}
		if !hasNode(nodes, e.From) || !hasNode(nodes, e.To) {
			issues = append(issues, Issue{Code: "dangling_edge", Ref: e.Key})
			continue
		}
		out[e.From] = append(out[e.From], e)
		in[e.To]++
		issues = append(issues, conditionIssues(e)...)
		if e.Kind == EdgeRework && e.MaxRework < 1 {
			issues = append(issues, Issue{Code: "rework_max", Ref: e.Key})
		}
		if e.OnExceed != "" && e.OnExceed != OnExceedHold {
			issues = append(issues, Issue{Code: "bad_on_exceed", Ref: e.Key})
		}
	}

	var starts []Node
	var ends []Node
	for _, n := range g.Nodes {
		if n.Key == "" {
			continue
		}
		if n.Type == NodeStart {
			starts = append(starts, n)
		}
		if n.Type == NodeEnd {
			ends = append(ends, n)
		}
	}
	if len(starts) != 1 {
		issues = append(issues, Issue{Code: "start_count"})
	}
	if len(ends) == 0 {
		issues = append(issues, Issue{Code: "end_missing"})
	}
	for _, n := range starts {
		if in[n.Key] > 0 {
			issues = append(issues, Issue{Code: "start_incoming", Ref: n.Key})
		}
	}
	for _, n := range ends {
		if len(out[n.Key]) > 0 {
			issues = append(issues, Issue{Code: "end_outgoing", Ref: n.Key})
		}
	}

	for _, n := range g.Nodes {
		if n.Key == "" || n.Type == NodeEnd {
			continue
		}
		outgoing := out[n.Key]
		if len(outgoing) == 0 {
			issues = append(issues, Issue{Code: "missing_outgoing", Ref: n.Key})
			continue
		}
		defaults := 0
		for _, e := range outgoing {
			if e.IsDefault {
				defaults++
			}
		}
		if len(outgoing) == 1 && defaults != 1 {
			issues = append(issues, Issue{Code: "single_not_default", Ref: n.Key})
		}
		if len(outgoing) >= 2 && defaults != 1 {
			if defaults == 0 {
				issues = append(issues, Issue{Code: "missing_default", Ref: n.Key})
			} else {
				issues = append(issues, Issue{Code: "multiple_default", Ref: n.Key})
			}
		}
		if n.Type == NodeDecision && len(outgoing) < 2 {
			issues = append(issues, Issue{Code: "decision_outgoing", Ref: n.Key})
		}
	}

	if len(starts) == 1 {
		seen := reachable(out, starts[0].Key)
		for _, n := range g.Nodes {
			if n.Key == "" {
				continue
			}
			if !seen[n.Key] {
				issues = append(issues, Issue{Code: "unreachable", Ref: n.Key})
			}
		}
		reverse := reverseAdj(g.Edges, nodes)
		endSeen := map[string]bool{}
		for _, n := range ends {
			for k, ok := range reachable(reverse, n.Key) {
				if ok {
					endSeen[k] = true
				}
			}
		}
		for _, n := range g.Nodes {
			if n.Key == "" || n.Type == NodeEnd {
				continue
			}
			if !endSeen[n.Key] {
				issues = append(issues, Issue{Code: "cannot_reach_end", Ref: n.Key})
			}
		}
	}

	normal := map[string][]Edge{}
	for _, e := range g.Edges {
		if e.Kind == EdgeNormal && hasNode(nodes, e.From) && hasNode(nodes, e.To) {
			normal[e.From] = append(normal[e.From], e)
		}
	}
	for _, e := range g.Edges {
		if e.Kind != EdgeRework || !hasNode(nodes, e.From) || !hasNode(nodes, e.To) {
			continue
		}
		if e.From != e.To && !canReach(normal, e.To, e.From) {
			issues = append(issues, Issue{Code: "rework_not_upstream", Ref: e.Key})
		}
	}

	var capped []Edge
	for _, e := range g.Edges {
		if !hasNode(nodes, e.From) || !hasNode(nodes, e.To) {
			continue
		}
		if e.Kind == EdgeRework && e.MaxRework >= 1 {
			continue
		}
		capped = append(capped, e)
	}
	if !isDAG(nodeKeys(g.Nodes), capped) {
		issues = append(issues, Issue{Code: "unbounded_cycle"})
	}
	return issues
}

func hasNode(nodes map[string]Node, key string) bool {
	_, ok := nodes[key]
	return ok
}

func nodeKeys(nodes []Node) []string {
	keys := make([]string, 0, len(nodes))
	seen := map[string]bool{}
	for _, n := range nodes {
		if n.Key == "" || seen[n.Key] {
			continue
		}
		seen[n.Key] = true
		keys = append(keys, n.Key)
	}
	return keys
}

func reachable(adj map[string][]Edge, start string) map[string]bool {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(key string) {
		if seen[key] {
			return
		}
		seen[key] = true
		for _, e := range adj[key] {
			walk(e.To)
		}
	}
	walk(start)
	return seen
}

func reverseAdj(edges []Edge, nodes map[string]Node) map[string][]Edge {
	rev := map[string][]Edge{}
	for _, e := range edges {
		if !hasNode(nodes, e.From) || !hasNode(nodes, e.To) {
			continue
		}
		rev[e.To] = append(rev[e.To], Edge{From: e.To, To: e.From})
	}
	return rev
}

func canReach(adj map[string][]Edge, from, to string) bool {
	return reachable(adj, from)[to]
}

func isDAG(keys []string, edges []Edge) bool {
	indeg := map[string]int{}
	next := map[string][]string{}
	for _, key := range keys {
		indeg[key] = 0
	}
	for _, e := range edges {
		if _, ok := indeg[e.From]; !ok {
			continue
		}
		if _, ok := indeg[e.To]; !ok {
			continue
		}
		next[e.From] = append(next[e.From], e.To)
		indeg[e.To]++
	}
	var q []string
	for _, key := range keys {
		if indeg[key] == 0 {
			q = append(q, key)
		}
	}
	seen := 0
	for len(q) > 0 {
		cur := q[0]
		q = q[1:]
		seen++
		for _, nxt := range next[cur] {
			indeg[nxt]--
			if indeg[nxt] == 0 {
				q = append(q, nxt)
			}
		}
	}
	return seen == len(keys)
}
