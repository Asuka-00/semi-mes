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
	"semi-mes/server/internal/model"
)

func writeDictErr(c *gin.Context, err error) {
	if errors.Is(err, dao.ErrDict) || errors.Is(err, dao.ErrNumber) {
		response.Error(c, ecode.InvalidParams)
		return
	}
	response.Error(c, ecode.InternalServerError)
}

func ListDictTypes(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListDictTypes(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"sysDictTypes": rows, "total": total})
}

func CreateDictType(c *gin.Context) {
	row := model.SysDictType{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = 0
	if err := dao.SaveDictType(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID})
}

func UpdateDictType(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row := model.SysDictType{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = id
	if err := dao.SaveDictType(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func DeleteDictType(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	if err := dao.DeleteDictType(database.GetDB(), id); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func ListDictItems(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListDictItems(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"sysDictItems": rows, "total": total})
}

func CreateDictItem(c *gin.Context) {
	row := model.SysDictItem{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = 0
	if err := dao.SaveDictItem(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID})
}

func UpdateDictItem(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row := model.SysDictItem{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = id
	if err := dao.SaveDictItem(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func DeleteDictItem(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	res := database.GetDB().Delete(&model.SysDictItem{}, id)
	if res.Error != nil || res.RowsAffected == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func PublicDictItems(c *gin.Context) {
	rows, err := dao.EnabledDictItems(database.GetDB(), c.Query("typeCode"))
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"items": rows})
}

func ListReasonCodes(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListReasonCodes(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"sysReasonCodes": rows, "total": total})
}

func CreateReasonCode(c *gin.Context) {
	row := model.MesReasonCode{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = 0
	if err := dao.SaveReasonCode(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID})
}

func UpdateReasonCode(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row := model.MesReasonCode{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = id
	if err := dao.SaveReasonCode(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func DeleteReasonCode(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	res := database.GetDB().Delete(&model.MesReasonCode{}, id)
	if res.Error != nil || res.RowsAffected == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func PublicReasons(c *gin.Context) {
	rows, err := dao.EnabledReasons(database.GetDB(), c.Query("category"))
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"reasons": rows})
}

func ListNumberRules(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListNumberRules(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"sysNumberRules": rows, "total": total})
}

func CreateNumberRule(c *gin.Context) {
	row := model.SysNumberRule{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = 0
	if err := dao.SaveNumberRule(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": row.ID})
}

func UpdateNumberRule(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row := model.SysNumberRule{}
	if err := c.ShouldBindJSON(&row); err != nil {
		response.Error(c, ecode.InvalidParams)
		return
	}
	row.ID = id
	if err := dao.SaveNumberRule(database.GetDB(), &row); err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func DeleteNumberRule(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		response.Error(c, ecode.InvalidParams)
		return
	}
	res := database.GetDB().Delete(&model.SysNumberRule{}, id)
	if res.Error != nil || res.RowsAffected == 0 {
		response.Error(c, ecode.InvalidParams)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func PreviewNumberRule(c *gin.Context) {
	var body struct {
		RuleCode string `json:"ruleCode"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.RuleCode) == "" {
		response.Error(c, ecode.InvalidParams)
		return
	}
	text, err := dao.PreviewNumber(database.GetDB(), body.RuleCode, time.Now())
	if err != nil {
		writeDictErr(c, err)
		return
	}
	response.Success(c, gin.H{"preview": text})
}
