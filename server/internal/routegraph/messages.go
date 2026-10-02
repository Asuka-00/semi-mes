package routegraph

var issueText = map[string][2]string{
	"empty_key":            {"节点或连线缺少标识", "A node or edge is missing its key"},
	"duplicate_node":       {"节点标识重复", "Duplicate node key"},
	"duplicate_edge":       {"连线标识重复", "Duplicate edge key"},
	"unknown_node_type":    {"节点类型不合法", "Unknown node type"},
	"unknown_edge_kind":    {"连线类型不合法", "Unknown edge kind"},
	"dangling_edge":        {"连线指向不存在的节点", "Edge references a missing node"},
	"operation_missing":    {"工序节点必须选择工序", "Operation node requires an operation"},
	"start_count":          {"必须有且仅有一个开始节点", "Exactly one start node is required"},
	"end_missing":          {"至少需要一个结束节点", "At least one end node is required"},
	"start_incoming":       {"开始节点不能有入边", "Start node cannot have incoming edges"},
	"end_outgoing":         {"结束节点不能有出边", "End node cannot have outgoing edges"},
	"missing_outgoing":     {"节点缺少出边，批次会停住", "Node has no outgoing edge"},
	"single_not_default":   {"唯一出边必须标为默认路径", "The only outgoing edge must be the default"},
	"missing_default":      {"分支必须有一条默认（否则）路径", "Branch needs one default (else) edge"},
	"multiple_default":     {"同一节点只能有一条默认路径", "Only one default edge is allowed"},
	"decision_outgoing":    {"判定节点至少需要两条出边", "Decision node needs at least two outgoing edges"},
	"unreachable":          {"节点从开始不可达", "Node is not reachable from start"},
	"cannot_reach_end":     {"节点无法到达结束", "Node cannot reach an end"},
	"rework_not_upstream":  {"返工边必须回到上游节点", "Rework edge must target an upstream node"},
	"rework_max":           {"返工边必须设置大于 0 的最大次数", "Rework edge needs a max count greater than zero"},
	"bad_on_exceed":        {"超出返工次数时只支持 Hold", "Only Hold is supported when rework is exceeded"},
	"unbounded_cycle":      {"存在没有返工次数上限的环", "Cycle has no capped rework edge"},
	"bad_condition_shape":  {"条件必须是 all 或 any 其中一种", "Condition must be either all or any"},
	"condition_on_default": {"默认路径不能带条件", "Default edge cannot carry a condition"},
	"missing_condition":    {"非默认路径必须填写条件", "Non-default edge needs a condition"},
	"bad_field":            {"条件字段不在允许列表中", "Condition field is not allowed"},
	"bad_op":               {"该字段不支持此比较符", "Operator is not allowed for this field"},
	"bad_value":            {"条件取值类型不正确", "Condition value has the wrong type"},
}

// IssueMessage returns a zh-CN or en-US explanation for a validation code.
func IssueMessage(code, lang string) string {
	pair, ok := issueText[code]
	if !ok {
		return code
	}
	if lang == "en" {
		return pair[1]
	}
	return pair[0]
}
