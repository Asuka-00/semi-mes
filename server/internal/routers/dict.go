package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		group.GET("/dict/items", handler.PublicDictItems)
		group.GET("/dict/reasons", handler.PublicReasons)

		types := group.Group("/sysDictType")
		types.POST("/list", handler.ListDictTypes)
		types.POST("", handler.CreateDictType)
		types.PUT("/:id", handler.UpdateDictType)
		types.DELETE("/:id", handler.DeleteDictType)

		items := group.Group("/sysDictItem")
		items.POST("/list", handler.ListDictItems)
		items.POST("", handler.CreateDictItem)
		items.PUT("/:id", handler.UpdateDictItem)
		items.DELETE("/:id", handler.DeleteDictItem)

		reasons := group.Group("/sysReasonCode")
		reasons.POST("/list", handler.ListReasonCodes)
		reasons.POST("", handler.CreateReasonCode)
		reasons.PUT("/:id", handler.UpdateReasonCode)
		reasons.DELETE("/:id", handler.DeleteReasonCode)

		rules := group.Group("/sysNumberRule")
		rules.POST("/list", handler.ListNumberRules)
		rules.POST("/preview", handler.PreviewNumberRule)
		rules.POST("", handler.CreateNumberRule)
		rules.PUT("/:id", handler.UpdateNumberRule)
		rules.DELETE("/:id", handler.DeleteNumberRule)
	})
}
