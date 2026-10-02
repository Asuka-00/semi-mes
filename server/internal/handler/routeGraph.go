package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/i18n"
	"semi-mes/server/internal/routegraph"
)

type graphNodeDTO struct {
	NodeKey        string  `json:"nodeKey"`
	NodeType       string  `json:"nodeType"`
	Name           string  `json:"name"`
	OperationID    uint64  `json:"operationID"`
	RecipeID       uint64  `json:"recipeID"`
	EquipmentGroup string  `json:"equipmentGroup"`
	PosX           float64 `json:"posX"`
	PosY           float64 `json:"posY"`
}

type graphEdgeDTO struct {
	EdgeKey   string                `json:"edgeKey"`
	FromKey   string                `json:"fromKey"`
	ToKey     string                `json:"toKey"`
	EdgeKind  string                `json:"edgeKind"`
	IsDefault bool                  `json:"isDefault"`
	Priority  int                   `json:"priority"`
	Condition *routegraph.Condition `json:"condition"`
	MaxRework int                   `json:"maxRework"`
	OnExceed  string                `json:"onExceed"`
	Label     string                `json:"label"`
}

type saveGraphBody struct {
	Nodes []graphNodeDTO `json:"nodes"`
	Edges []graphEdgeDTO `json:"edges"`
}

type createVersionBody struct {
	Note     string `json:"note"`
	CopyFrom uint64 `json:"copyFrom"`
}

type resolveBody struct {
	CurrentNodeKey string `json:"currentNodeKey"`
	Context        struct {
		Inspection struct {
			Result string `json:"result"`
			Grade  string `json:"grade"`
		} `json:"inspection"`
		Defect struct {
			Code string `json:"code"`
		} `json:"defect"`
		Lot struct {
			ProductCode string  `json:"productCode"`
			Priority    float64 `json:"priority"`
			Type        string  `json:"type"`
		} `json:"lot"`
		ReworkCounts map[string]int `json:"reworkCounts"`
	} `json:"context"`
}

func (h *baseProcessRouteHandler) ListVersions(c *gin.Context) {
	routeID, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	rows, err := dao.ListRouteVersions(database.GetDB(), routeID)
	if err != nil {
		writeRouteErr(c, err)
		return
	}
	response.Success(c, gin.H{"versions": rows})
}

func (h *baseProcessRouteHandler) CreateVersion(c *gin.Context) {
	routeID, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := createVersionBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	row, err := dao.CreateRouteVersion(database.GetDB(), routeID, body.Note, body.CopyFrom)
	if err != nil {
		writeRouteErr(c, err)
		return
	}
	response.Success(c, gin.H{"version": row})
}

func (h *baseProcessRouteHandler) GetVersion(c *gin.Context) {
	routeID, versionID, ok := routeVersionIDs(c)
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	ver, graph, err := loadVersionGraph(routeID, versionID)
	if err != nil {
		writeRouteErr(c, err)
		return
	}
	response.Success(c, gin.H{"version": ver, "nodes": nodesDTO(graph), "edges": edgesDTO(graph)})
}

func (h *baseProcessRouteHandler) SaveGraph(c *gin.Context) {
	routeID, versionID, ok := routeVersionIDs(c)
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := saveGraphBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.SaveRouteGraph(database.GetDB(), routeID, versionID, body.toGraph()); err != nil {
		writeRouteErr(c, err)
		return
	}
	response.Success(c, gin.H{"saved": true})
}

func (h *baseProcessRouteHandler) ValidateVersion(c *gin.Context) {
	routeID, versionID, ok := routeVersionIDs(c)
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	_, graph, err := loadVersionGraph(routeID, versionID)
	if err != nil {
		writeRouteErr(c, err)
		return
	}
	issues := localizeIssues(c, routegraph.Validate(graph))
	response.Success(c, gin.H{"valid": len(issues) == 0, "issues": issues})
}

func (h *baseProcessRouteHandler) ReleaseVersion(c *gin.Context) {
	routeID, versionID, ok := routeVersionIDs(c)
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	ver, issues, err := dao.ReleaseRouteVersion(database.GetDB(), routeID, versionID)
	if err != nil {
		writeRouteErr(c, err)
		return
	}
	localized := localizeIssues(c, issues)
	response.Success(c, gin.H{"released": len(localized) == 0, "version": ver, "issues": localized})
}

func (h *baseProcessRouteHandler) ResolveNext(c *gin.Context) {
	routeID, versionID, ok := routeVersionIDs(c)
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := resolveBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.CurrentNodeKey == "" {
		response.Error(c, ecode.InvalidParams)
		return
	}
	_, graph, err := loadVersionGraph(routeID, versionID)
	if err != nil {
		writeRouteErr(c, err)
		return
	}
	result, err := routegraph.Resolve(graph, body.CurrentNodeKey, routegraph.LotContext{
		InspectionResult: body.Context.Inspection.Result,
		InspectionGrade:  body.Context.Inspection.Grade,
		DefectCode:       body.Context.Defect.Code,
		ProductCode:      body.Context.Lot.ProductCode,
		Priority:         body.Context.Lot.Priority,
		LotType:          body.Context.Lot.Type,
		ReworkCounts:     body.Context.ReworkCounts,
	})
	if errors.Is(err, routegraph.ErrNodeNotFound) {
		c.JSON(200, gin.H{"code": 40004, "msg": i18n.T(c, "error.route.not_found"), "data": struct{}{}})
		return
	}
	if errors.Is(err, routegraph.ErrNoPath) {
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.route.no_path"), "data": struct{}{}})
		return
	}
	if err != nil {
		c.JSON(200, gin.H{"code": 50000, "msg": i18n.T(c, "error.server.internal"), "data": struct{}{}})
		return
	}
	response.Success(c, result)
}

func (b saveGraphBody) toGraph() routegraph.Graph {
	g := routegraph.Graph{Nodes: make([]routegraph.Node, 0, len(b.Nodes)), Edges: make([]routegraph.Edge, 0, len(b.Edges))}
	for _, n := range b.Nodes {
		g.Nodes = append(g.Nodes, routegraph.Node{
			Key: n.NodeKey, Type: n.NodeType, Name: n.Name, OperationID: n.OperationID,
			RecipeID: n.RecipeID, EquipmentGroup: n.EquipmentGroup, PosX: n.PosX, PosY: n.PosY,
		})
	}
	for _, e := range b.Edges {
		g.Edges = append(g.Edges, routegraph.Edge{
			Key: e.EdgeKey, From: e.FromKey, To: e.ToKey, Kind: e.EdgeKind, IsDefault: e.IsDefault,
			Priority: e.Priority, Condition: e.Condition, MaxRework: e.MaxRework, OnExceed: e.OnExceed, Label: e.Label,
		})
	}
	return g
}

func nodesDTO(g routegraph.Graph) []graphNodeDTO {
	out := make([]graphNodeDTO, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		out = append(out, graphNodeDTO{
			NodeKey: n.Key, NodeType: n.Type, Name: n.Name, OperationID: n.OperationID,
			RecipeID: n.RecipeID, EquipmentGroup: n.EquipmentGroup, PosX: n.PosX, PosY: n.PosY,
		})
	}
	return out
}

func edgesDTO(g routegraph.Graph) []graphEdgeDTO {
	out := make([]graphEdgeDTO, 0, len(g.Edges))
	for _, e := range g.Edges {
		out = append(out, graphEdgeDTO{
			EdgeKey: e.Key, FromKey: e.From, ToKey: e.To, EdgeKind: e.Kind, IsDefault: e.IsDefault,
			Priority: e.Priority, Condition: e.Condition, MaxRework: e.MaxRework, OnExceed: e.OnExceed, Label: e.Label,
		})
	}
	return out
}

func loadVersionGraph(routeID, versionID uint64) (any, routegraph.Graph, error) {
	ver, err := dao.GetRouteVersion(database.GetDB(), routeID, versionID)
	if err != nil {
		return nil, routegraph.Graph{}, err
	}
	graph, err := dao.LoadRouteGraph(database.GetDB(), versionID)
	return ver, graph, err
}

func writeRouteErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, dao.ErrRouteVersionNotFound):
		c.JSON(200, gin.H{"code": 40004, "msg": i18n.T(c, "error.route.not_found"), "data": struct{}{}})
	case errors.Is(err, dao.ErrRouteNotDraft):
		c.JSON(200, gin.H{"code": 40009, "msg": i18n.T(c, "error.route.not_draft"), "data": struct{}{}})
	default:
		c.JSON(200, gin.H{"code": 50000, "msg": i18n.T(c, "error.server.internal"), "data": struct{}{}})
	}
}

func localizeIssues(c *gin.Context, issues []routegraph.Issue) []routegraph.Issue {
	lang := "zh"
	if strings.Contains(strings.ToLower(c.GetHeader("Accept-Language")), "en") {
		lang = "en"
	}
	out := make([]routegraph.Issue, len(issues))
	for i, issue := range issues {
		issue.Message = routegraph.IssueMessage(issue.Code, lang)
		out[i] = issue
	}
	if out == nil {
		out = []routegraph.Issue{}
	}
	return out
}

func routeVersionIDs(c *gin.Context) (uint64, uint64, bool) {
	routeID, ok := pathUint(c, "id")
	if !ok {
		return 0, 0, false
	}
	versionID, ok := pathUint(c, "versionId")
	return routeID, versionID, ok
}

func pathUint(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	return id, err == nil && id > 0
}
