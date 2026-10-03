package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/model"
)

type trackInBody struct {
	LotID       uint64 `json:"lotId"`
	EquipmentID uint64 `json:"equipmentId"`
	RecipeID    uint64 `json:"recipeId"`
}

type sampleBody struct {
	ParamCode string    `json:"paramCode"`
	Values    []float64 `json:"values"`
}

type trackOutBody struct {
	LotID            uint64       `json:"lotId"`
	QtyOut           int          `json:"qtyOut"`
	QtyScrap         int          `json:"qtyScrap"`
	ScrapReasonCode  string       `json:"scrapReasonCode"`
	InspectionResult string       `json:"inspectionResult"`
	InspectionGrade  string       `json:"inspectionGrade"`
	DefectCode       string       `json:"defectCode"`
	Measurements     []sampleBody `json:"measurements"`
}

type passBody struct {
	LotID            uint64 `json:"lotId"`
	InspectionResult string `json:"inspectionResult"`
	InspectionGrade  string `json:"inspectionGrade"`
	DefectCode       string `json:"defectCode"`
}

type abortBody struct {
	LotID  uint64 `json:"lotId"`
	Reason string `json:"reason"`
}

// Station loads the lot, current step, and allowed equipment.
func Station(c *gin.Context) {
	lotNo := c.Query("lotNo")
	if lotNo == "" {
		response.Error(c, ecode.InvalidParams)
		return
	}
	view, err := dao.GetStation(database.GetDB(), lotNo)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, view)
}

// TrackIn checks a lot into equipment.
func TrackIn(c *gin.Context) {
	body := trackInBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 || body.EquipmentID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	lot, move, err := dao.TrackIn(database.GetDB(), dao.TrackInInput{
		LotID: body.LotID, EquipmentID: body.EquipmentID, RecipeID: body.RecipeID, OperatorID: op,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot, "move": move})
}

// TrackOut finishes the step and moves the lot.
func TrackOut(c *gin.Context) {
	body := trackOutBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	samples := make([]dao.SampleInput, 0, len(body.Measurements))
	for _, sample := range body.Measurements {
		samples = append(samples, dao.SampleInput{ParamCode: sample.ParamCode, Values: sample.Values})
	}
	lot, result, err := dao.TrackOut(database.GetDB(), dao.TrackOutInput{
		LotID: body.LotID, QtyOut: body.QtyOut, QtyScrap: body.QtyScrap, ScrapReasonCode: body.ScrapReasonCode,
		InspectionResult: body.InspectionResult, InspectionGrade: body.InspectionGrade, DefectCode: body.DefectCode,
		Measurements: samples, OperatorID: op,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot, "result": result})
}

// AbortTrackIn cancels an open check-in.
func AbortTrackIn(c *gin.Context) {
	body := abortBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	lot, err := dao.AbortTrackIn(database.GetDB(), body.LotID, op, body.Reason)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot})
}

// PassNode leaves a start or decision node.
func PassNode(c *gin.Context) {
	body := passBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	op, _ := currentUserID(c)
	lot, result, err := dao.PassNode(database.GetDB(), dao.PassInput{
		LotID: body.LotID, InspectionResult: body.InspectionResult, InspectionGrade: body.InspectionGrade,
		DefectCode: body.DefectCode, OperatorID: op,
	})
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot, "result": result})
}

// ListMoves returns searchable move history.
func ListMoves(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListMoves(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if rows == nil {
		rows = []dao.MoveView{}
	}
	response.Success(c, gin.H{"wipMoves": rows, "total": total})
}

// WIPOverview returns counts and the hold list.
func WIPOverview(c *gin.Context) {
	view, err := dao.WIPOverview(database.GetDB())
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, view)
}

// ListEquipment returns the equipment stub.
func ListEquipment(c *gin.Context) {
	rows, err := dao.ListEquipment(database.GetDB(), c.Query("group"))
	if err != nil {
		writeWipErr(c, err)
		return
	}
	if rows == nil {
		rows = []model.EqpEquipment{}
	}
	response.Success(c, gin.H{"equipment": rows})
}
