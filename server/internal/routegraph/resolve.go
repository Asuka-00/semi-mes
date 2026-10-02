package routegraph

import (
	"errors"
	"sort"
)

// ErrNodeNotFound means the current node is not on this version.
var ErrNodeNotFound = errors.New("node not found")

// ErrNoPath means the current node has no usable outgoing edge.
var ErrNoPath = errors.New("no outgoing path")

// Resolve picks the next node when a lot leaves currentKey.
// Non-default edges are tested by ascending priority; the default edge is last.
// A rework edge whose taken count has reached MaxRework yields Hold and does not traverse.
func Resolve(g Graph, currentKey string, ctx LotContext) (Result, error) {
	var current *Node
	index := map[string]Node{}
	for _, n := range g.Nodes {
		index[n.Key] = n
		if n.Key == currentKey {
			node := n
			current = &node
		}
	}
	if current == nil {
		return Result{}, ErrNodeNotFound
	}
	if current.Type == NodeEnd {
		return Result{Action: ActionEnd, NextNodeKey: current.Key, NextNodeType: current.Type, NextNodeName: current.Name, Reason: ReasonAlreadyEnd}, nil
	}

	var outgoing []Edge
	for _, e := range g.Edges {
		if e.From == currentKey {
			outgoing = append(outgoing, e)
		}
	}
	sort.SliceStable(outgoing, func(i, j int) bool {
		if outgoing[i].IsDefault != outgoing[j].IsDefault {
			return !outgoing[i].IsDefault
		}
		return outgoing[i].Priority < outgoing[j].Priority
	})

	var chosen *Edge
	reason := ReasonMatched
	for i := range outgoing {
		e := &outgoing[i]
		if e.IsDefault {
			continue
		}
		if e.Condition.matches(ctx, e.Key) {
			chosen = e
			reason = ReasonMatched
			break
		}
	}
	if chosen == nil {
		for i := range outgoing {
			if outgoing[i].IsDefault {
				chosen = &outgoing[i]
				reason = ReasonDefault
				break
			}
		}
	}
	if chosen == nil {
		return Result{}, ErrNoPath
	}
	if chosen.Kind == EdgeRework {
		taken := 0
		if ctx.ReworkCounts != nil {
			taken = ctx.ReworkCounts[chosen.Key]
		}
		if chosen.MaxRework > 0 && taken >= chosen.MaxRework {
			return Result{Action: ActionHold, EdgeKey: chosen.Key, Reason: ReasonReworkExceeded}, nil
		}
	}
	next, ok := index[chosen.To]
	if !ok {
		return Result{}, ErrNoPath
	}
	return Result{
		Action:       ActionMove,
		NextNodeKey:  next.Key,
		NextNodeType: next.Type,
		NextNodeName: next.Name,
		EdgeKey:      chosen.Key,
		Reason:       reason,
	}, nil
}
