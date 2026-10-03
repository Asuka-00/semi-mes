package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
	"semi-mes/server/internal/model"
)

func ListCarriers(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListCarriers(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"wipCarriers": rows, "total": total})
}

func CreateCarrier(c *gin.Context) {
	row := model.WipCarrier{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = 0
	if err := dao.SaveCarrier(database.GetDB(), &row); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID, "carrierNo": row.CarrierNo})
}

func UpdateCarrier(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row := model.WipCarrier{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = id
	if err := dao.SaveCarrier(database.GetDB(), &row); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func DeleteCarrier(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	res := database.GetDB().Delete(&model.WipCarrier{}, id)
	if res.Error != nil {
		writeWipErr(c, res.Error)
		return
	}
	if res.RowsAffected == 0 {
		response.Error(c, ecode.NotFound)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func CarrierSlots(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	carrier, slots, lotNo, err := dao.SlotMap(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"carrier": carrier, "slots": slots, "lotNo": lotNo})
}

type bindBody struct {
	LotID uint64 `json:"lotId"`
}

func BindCarrier(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := bindBody{}
	if err := c.ShouldBindJSON(&body); err != nil || body.LotID == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.BindCarrier(database.GetDB(), id, body.LotID); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func UnbindCarrier(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.UnbindCarrier(database.GetDB(), id); err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func LotWafers(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	rows, err := dao.ListLotWafers(database.GetDB(), id)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"wafers": rows})
}

type scrapWaferBody struct {
	WaferIDs   []uint64 `json:"waferIds"`
	ReasonCode string   `json:"reasonCode"`
}

func ScrapWafers(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	body := scrapWaferBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	lot, err := dao.ScrapLotWafers(database.GetDB(), id, body.WaferIDs, body.ReasonCode)
	if err != nil {
		writeWipErr(c, err)
		return
	}
	response.Success(c, gin.H{"lot": lot})
}
