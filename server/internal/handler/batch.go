package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/rbac"
)

type batchBody struct {
	Resource   string   `json:"resource"`
	Action     string   `json:"action"`
	IDs        []uint64 `json:"ids"`
	ReasonCode string   `json:"reasonCode"`
	Reason     string   `json:"reason"`
}

// Batch runs a permission-checked action on several rows.
func Batch(c *gin.Context) {
	body := batchBody{}
	if err := c.ShouldBindJSON(&body); err != nil || len(body.IDs) == 0 || len(body.IDs) > 200 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	code, ok := batchPermission(body.Resource, body.Action)
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, 40001, "error.auth.unauthorized")
		return
	}
	allowed, err := rbac.UserHasPermission(userID, code)
	if err != nil {
		response.Error(c, ecode.InternalServerError)
		return
	}
	if !allowed {
		fail(c, 40003, "error.auth.forbidden")
		return
	}
	result, err := dao.RunBatch(database.GetDB(), body.Resource, body.Action, body.IDs, body.ReasonCode, body.Reason)
	if err != nil {
		if errors.Is(err, dao.ErrBatch) {
			response.Error(c, ecode.InvalidParams)
			return
		}
		response.Error(c, ecode.InternalServerError)
		return
	}
	response.Success(c, result)
}

func batchPermission(resource, action string) (string, bool) {
	prefix := map[string]string{
		"baseFactory":        "base:factory",
		"baseWorkshop":       "base:workshop",
		"baseProductionLine": "base:line",
		"baseProduct":        "base:product",
		"baseProcessRoute":   "base:route",
		"baseOperation":      "base:operation",
		"baseRecipe":         "base:recipe",
		"sysUser":            "system:user",
		"sysRole":            "system:role",
		"sysMenu":            "system:menu",
		"wipWorkOrder":       "wo:order",
		"wipLot":             "lot:lot",
		"eqpEquipment":       "eqp:equipment",
		"eqpPmPlan":          "eqp:pm",
		"qcInspectPlan":      "qc:plan",
		"qcDefectCode":       "qc:defect",
	}
	base, ok := prefix[resource]
	if !ok {
		return "", false
	}
	switch action {
	case "delete":
		if resource == "wipLot" {
			return "", false
		}
		return base + ":delete", true
	case "enable", "disable":
		switch resource {
		case "wipLot", "wipWorkOrder", "eqpEquipment":
			return "", false
		}
		return base + ":edit", true
	case "hold", "release":
		if resource != "wipLot" {
			return "", false
		}
		return base + ":edit", true
	default:
		return "", false
	}
}
