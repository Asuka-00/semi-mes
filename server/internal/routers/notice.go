package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		group.GET("/features", handler.Features)
		n := group.Group("/sysNotice")
		n.GET("", handler.ListNotices)
		n.POST("/readAll", handler.MarkAllNoticesRead)
		n.POST("/:id/read", handler.MarkNoticeRead)
	})
}
