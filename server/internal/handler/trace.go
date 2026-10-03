package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
)

type traceBody struct {
	LotID         uint64 `json:"lotId"`
	LotNo         string `json:"lotNo"`
	EquipmentID   uint64 `json:"equipmentId"`
	EquipmentCode string `json:"equipmentCode"`
	RecipeID      uint64 `json:"recipeId"`
	OperationID   uint64 `json:"operationId"`
	NodeKey       string `json:"nodeKey"`
	From          string `json:"from"`
	To            string `json:"to"`
}

// TraceLot loads a lot by id or lot number and returns its trace report.
func TraceLot(c *gin.Context) {
	body := traceBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	id := body.LotID
	if id == 0 && strings.TrimSpace(body.LotNo) != "" {
		found, err := dao.FindLotByNo(database.GetDB(), body.LotNo)
		if err != nil {
			writeWipErr(c, err)
			return
		}
		id = found
	}
	if id == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	writeTrace(c, id)
}

// TraceLotByID loads the trace report for the path id.
func TraceLotByID(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	writeTrace(c, id)
}

func writeTrace(c *gin.Context, id uint64) {
	report, err := dao.LotTrace(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, report)
}

// TraceReverse lists lots that match an equipment, recipe, operation, or time window.
func TraceReverse(c *gin.Context) {
	body := traceBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	toExclusive := ""
	if strings.TrimSpace(body.To) != "" {
		day, err := time.Parse("2006-01-02", strings.TrimSpace(body.To))
		if err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
		toExclusive = day.AddDate(0, 0, 1).Format("2006-01-02")
	}
	rows, err := dao.ReverseTrace(database.GetDB(), dao.TraceReverseQuery{
		EquipmentID: body.EquipmentID, EquipmentCode: body.EquipmentCode, RecipeID: body.RecipeID,
		OperationID: body.OperationID, NodeKey: body.NodeKey, From: strings.TrimSpace(body.From), ToExclusive: toExclusive,
	})
	if err != nil {
		if errors.Is(err, dao.ErrTraceFilter) {
			response.Error(c, ecode.InvalidParams)
			return
		}
		response.Error(c, ecode.InternalServerError)
		return
	}
	response.Success(c, gin.H{"lots": rows})
}
