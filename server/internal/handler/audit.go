package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"semi-mes/server/internal/dao"
	"semi-mes/server/internal/database"
	"semi-mes/server/internal/ecode"
)

// ListAuditLogs returns audit rows for the unified list query.
func ListAuditLogs(c *gin.Context) {
	body := wipListBody{}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			response.Error(c, ecode.InvalidParams)
			return
		}
	}
	page, limit, sort, cols := body.query()
	rows, total, err := dao.ListAudit(database.GetDB(), page, limit, sort, cols)
	if err != nil {
		response.Error(c, ecode.InternalServerError)
		return
	}
	response.Success(c, gin.H{"sysAudits": rows, "total": total})
}
