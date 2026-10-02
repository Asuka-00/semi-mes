package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		orders := group.Group("/wipWorkOrder")
		orders.GET("/releasedVersions", handler.ReleasedRoutes)
		orders.POST("/list", handler.ListWorkOrders)
		orders.POST("", handler.CreateWorkOrder)
		orders.PUT("/:id", handler.UpdateWorkOrder)
		orders.DELETE("/:id", handler.DeleteWorkOrder)
		orders.POST("/:id/release", handler.ReleaseWorkOrder)
		orders.POST("/:id/close", handler.CloseWorkOrder)
		orders.POST("/:id/start", handler.StartLot)

		lots := group.Group("/wipLot")
		lots.POST("/list", handler.ListLots)
		lots.POST("/merge", handler.MergeLots)
		lots.GET("/:id", handler.GetLot)
		lots.POST("/:id/hold", handler.HoldLot)
		lots.POST("/:id/releaseHold", handler.ReleaseHoldLot)
		lots.POST("/:id/split", handler.SplitLot)
		lots.POST("/:id/advance", handler.AdvanceLot)
	})
}
