package dao

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"semi-mes/server/internal/model"
	"semi-mes/server/internal/routegraph"
)

// ErrRouteVersionNotFound is returned when the version is missing or belongs to another route.
var ErrRouteVersionNotFound = errors.New("route version not found")

// ErrRouteNotDraft is returned when a mutation targets a released or obsolete version.
var ErrRouteNotDraft = errors.New("route version is not draft")

// ListRouteVersions returns versions of a route, newest first.
func ListRouteVersions(db *gorm.DB, routeID uint64) ([]model.BaseRouteVersion, error) {
	var rows []model.BaseRouteVersion
	err := db.Where("route_id = ?", routeID).Order("version_no desc").Find(&rows).Error
	return rows, err
}

// GetRouteVersion loads one version and checks the route id.
func GetRouteVersion(db *gorm.DB, routeID, versionID uint64) (*model.BaseRouteVersion, error) {
	var row model.BaseRouteVersion
	err := db.Where("id = ? AND route_id = ?", versionID, routeID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRouteVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// CreateRouteVersion inserts the next draft, copying an existing graph or starting from a skeleton.
func CreateRouteVersion(db *gorm.DB, routeID uint64, note string, copyFrom uint64) (*model.BaseRouteVersion, error) {
	var route model.BaseProcessRoute
	if err := db.First(&route, routeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRouteVersionNotFound
		}
		return nil, err
	}
	var maxNo int
	if err := db.Model(&model.BaseRouteVersion{}).Where("route_id = ?", routeID).Select("COALESCE(MAX(version_no),0)").Scan(&maxNo).Error; err != nil {
		return nil, err
	}
	graph := routegraph.Skeleton()
	if copyFrom > 0 {
		src, err := GetRouteVersion(db, routeID, copyFrom)
		if err != nil {
			return nil, err
		}
		graph, err = LoadRouteGraph(db, src.ID)
		if err != nil {
			return nil, err
		}
	}
	row := &model.BaseRouteVersion{RouteID: routeID, VersionNo: maxNo + 1, State: routegraph.StateDraft, Note: note}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		return saveRouteGraph(tx, row.ID, graph)
	}); err != nil {
		return nil, err
	}
	return row, nil
}

// SaveRouteGraph replaces nodes and edges of a draft version.
func SaveRouteGraph(db *gorm.DB, routeID, versionID uint64, g routegraph.Graph) error {
	ver, err := GetRouteVersion(db, routeID, versionID)
	if err != nil {
		return err
	}
	if ver.State != routegraph.StateDraft {
		return ErrRouteNotDraft
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return saveRouteGraph(tx, versionID, g)
	})
}

func saveRouteGraph(tx *gorm.DB, versionID uint64, g routegraph.Graph) error {
	if err := tx.Unscoped().Where("version_id = ?", versionID).Delete(&model.BaseRouteNode{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("version_id = ?", versionID).Delete(&model.BaseRouteEdge{}).Error; err != nil {
		return err
	}
	if len(g.Nodes) > 0 {
		rows := make([]model.BaseRouteNode, 0, len(g.Nodes))
		for _, n := range g.Nodes {
			rows = append(rows, model.BaseRouteNode{
				VersionID: versionID, NodeKey: n.Key, NodeType: n.Type, Name: n.Name,
				OperationID: n.OperationID, RecipeID: n.RecipeID, EquipmentGroup: n.EquipmentGroup,
				PosX: n.PosX, PosY: n.PosY,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(g.Edges) == 0 {
		return nil
	}
	rows := make([]model.BaseRouteEdge, 0, len(g.Edges))
	for _, e := range g.Edges {
		flag := 0
		if e.IsDefault {
			flag = 1
		}
		onExceed := e.OnExceed
		if e.Kind == routegraph.EdgeRework && onExceed == "" {
			onExceed = routegraph.OnExceedHold
		}
		rows = append(rows, model.BaseRouteEdge{
			VersionID: versionID, EdgeKey: e.Key, FromKey: e.From, ToKey: e.To, EdgeKind: e.Kind,
			IsDefault: flag, Priority: e.Priority, ConditionJSON: routegraph.MustJSON(e.Condition),
			MaxRework: e.MaxRework, OnExceed: onExceed, Label: e.Label,
		})
	}
	return tx.Create(&rows).Error
}

// LoadRouteGraph reads the flowchart of a version.
func LoadRouteGraph(db *gorm.DB, versionID uint64) (routegraph.Graph, error) {
	var nodes []model.BaseRouteNode
	var edges []model.BaseRouteEdge
	if err := db.Where("version_id = ?", versionID).Order("id asc").Find(&nodes).Error; err != nil {
		return routegraph.Graph{}, err
	}
	if err := db.Where("version_id = ?", versionID).Order("id asc").Find(&edges).Error; err != nil {
		return routegraph.Graph{}, err
	}
	g := routegraph.Graph{Nodes: make([]routegraph.Node, 0, len(nodes)), Edges: make([]routegraph.Edge, 0, len(edges))}
	for _, n := range nodes {
		g.Nodes = append(g.Nodes, routegraph.Node{
			Key: n.NodeKey, Type: n.NodeType, Name: n.Name, OperationID: n.OperationID,
			RecipeID: n.RecipeID, EquipmentGroup: n.EquipmentGroup, PosX: n.PosX, PosY: n.PosY,
		})
	}
	for _, e := range edges {
		cond, err := routegraph.ParseCondition(e.ConditionJSON)
		edge := routegraph.Edge{
			Key: e.EdgeKey, From: e.FromKey, To: e.ToKey, Kind: e.EdgeKind, IsDefault: e.IsDefault == 1,
			Priority: e.Priority, Condition: cond, MaxRework: e.MaxRework, OnExceed: e.OnExceed, Label: e.Label,
		}
		if err != nil {
			edge.ConditionInvalid = true
		}
		g.Edges = append(g.Edges, edge)
	}
	return g, nil
}

// ReleaseRouteVersion validates a draft and publishes it, obsoleting the previous release.
func ReleaseRouteVersion(db *gorm.DB, routeID, versionID uint64) (*model.BaseRouteVersion, []routegraph.Issue, error) {
	ver, err := GetRouteVersion(db, routeID, versionID)
	if err != nil {
		return nil, nil, err
	}
	if ver.State != routegraph.StateDraft {
		return nil, nil, ErrRouteNotDraft
	}
	graph, err := LoadRouteGraph(db, versionID)
	if err != nil {
		return nil, nil, err
	}
	issues := routegraph.Validate(graph)
	if len(issues) > 0 {
		return ver, issues, nil
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.BaseRouteVersion{}).
			Where("route_id = ? AND state = ? AND id <> ?", routeID, routegraph.StateReleased, versionID).
			Update("state", routegraph.StateObsolete).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.BaseRouteVersion{}).Where("id = ?", versionID).Update("state", routegraph.StateReleased).Error; err != nil {
			return err
		}
		return tx.Model(&model.BaseProcessRoute{}).Where("id = ?", routeID).Update("version", strconv.Itoa(ver.VersionNo)).Error
	})
	if err != nil {
		return nil, nil, err
	}
	ver.State = routegraph.StateReleased
	return ver, nil, nil
}
