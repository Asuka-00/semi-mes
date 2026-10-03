package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		pref := group.Group("/sysPref")
		pref.GET("/:page", handler.GetPref)
		pref.PUT("/:page", handler.SavePref)
		group.POST("/batch", handler.Batch)
		group.POST("/sysAudit/list", handler.ListAuditLogs)
	})
}
