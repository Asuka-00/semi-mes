package routegraph

// Node keys of the seeded CMOS route.
const (
	KeyStart   = "start"
	KeyClean   = "clean"
	KeyPhoto   = "photo"
	KeyInspect = "inspect"
	KeyDecide  = "decide"
	KeyEtch    = "etch"
	KeyEng     = "eng"
	KeyEndMain = "end_main"
	KeyEndEng  = "end_eng"

	EdgeReworkPhoto = "e_rework"
	EdgeEng         = "e_eng"
	EdgeDefaultEtch = "e_etch"
)

// Sample is a lithography route: inspection failure reworks photo (max 2, then Hold),
// engineering lots branch to a review step, and every other lot takes the default etch path.
// Operation IDs come from the operation master; missing IDs leave the operation node unset.
func Sample(ops map[string]uint64) Graph {
	return Graph{
		Nodes: []Node{
			{Key: KeyStart, Type: NodeStart, Name: "开始 Start", PosX: 40, PosY: 200},
			{Key: KeyClean, Type: NodeOperation, Name: "清洗 Clean", OperationID: ops["CLEAN"], EquipmentGroup: "WET", PosX: 240, PosY: 200},
			{Key: KeyPhoto, Type: NodeOperation, Name: "光刻 Photo", OperationID: ops["PHOTO"], EquipmentGroup: "PHOTO", PosX: 460, PosY: 200},
			{Key: KeyInspect, Type: NodeOperation, Name: "检测 Inspect", OperationID: ops["INSPECT"], EquipmentGroup: "METRO", PosX: 680, PosY: 200},
			{Key: KeyDecide, Type: NodeDecision, Name: "判定 Decision", PosX: 900, PosY: 200},
			{Key: KeyEtch, Type: NodeOperation, Name: "刻蚀 Etch", OperationID: ops["ETCH"], EquipmentGroup: "ETCH", PosX: 1140, PosY: 60},
			{Key: KeyEndMain, Type: NodeEnd, Name: "结束 End", PosX: 1360, PosY: 60},
			{Key: KeyEng, Type: NodeOperation, Name: "工程评审 Eng Review", OperationID: ops["ENG_REVIEW"], EquipmentGroup: "ENG", PosX: 1140, PosY: 340},
			{Key: KeyEndEng, Type: NodeEnd, Name: "结束 End", PosX: 1360, PosY: 340},
		},
		Edges: []Edge{
			{Key: "e_start", From: KeyStart, To: KeyClean, Kind: EdgeNormal, IsDefault: true, Priority: 10},
			{Key: "e_clean", From: KeyClean, To: KeyPhoto, Kind: EdgeNormal, IsDefault: true, Priority: 10},
			{Key: "e_photo", From: KeyPhoto, To: KeyInspect, Kind: EdgeNormal, IsDefault: true, Priority: 10},
			{Key: "e_inspect", From: KeyInspect, To: KeyDecide, Kind: EdgeNormal, IsDefault: true, Priority: 10},
			{
				Key: EdgeReworkPhoto, From: KeyDecide, To: KeyPhoto, Kind: EdgeRework, Priority: 1, MaxRework: 2, OnExceed: OnExceedHold, Label: "fail rework",
				Condition: &Condition{All: []Predicate{{Field: "inspection.result", Op: "eq", Value: "fail"}}},
			},
			{
				Key: EdgeEng, From: KeyDecide, To: KeyEng, Kind: EdgeNormal, Priority: 2, Label: "engineering",
				Condition: &Condition{All: []Predicate{{Field: "lot.type", Op: "eq", Value: "engineering"}}},
			},
			{Key: EdgeDefaultEtch, From: KeyDecide, To: KeyEtch, Kind: EdgeNormal, IsDefault: true, Priority: 100, Label: "else"},
			{Key: "e_etch_end", From: KeyEtch, To: KeyEndMain, Kind: EdgeNormal, IsDefault: true, Priority: 10},
			{Key: "e_eng_end", From: KeyEng, To: KeyEndEng, Kind: EdgeNormal, IsDefault: true, Priority: 10},
		},
	}
}

// Skeleton is the empty draft: start, end, and one default edge.
func Skeleton() Graph {
	return Graph{
		Nodes: []Node{
			{Key: "start", Type: NodeStart, Name: "开始 Start", PosX: 80, PosY: 180},
			{Key: "end", Type: NodeEnd, Name: "结束 End", PosX: 420, PosY: 180},
		},
		Edges: []Edge{
			{Key: "e_start", From: "start", To: "end", Kind: EdgeNormal, IsDefault: true, Priority: 10},
		},
	}
}
