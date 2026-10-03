package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		carriers := group.Group("/wipCarrier")
		carriers.POST("/list", handler.ListCarriers)
		carriers.POST("", handler.CreateCarrier)
		carriers.PUT("/:id", handler.UpdateCarrier)
		carriers.DELETE("/:id", handler.DeleteCarrier)
		carriers.GET("/:id/slots", handler.CarrierSlots)
		carriers.POST("/:id/bind", handler.BindCarrier)
		carriers.POST("/:id/unbind", handler.UnbindCarrier)

		lots := group.Group("/wipLot")
		lots.GET("/:id/wafers", handler.LotWafers)
		lots.POST("/:id/scrap", handler.ScrapWafers)
	})
}
