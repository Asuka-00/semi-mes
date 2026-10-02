package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		moves := group.Group("/wipMove")
		moves.GET("/station", handler.Station)
		moves.GET("/overview", handler.WIPOverview)
		moves.POST("/list", handler.ListMoves)
		moves.POST("/trackIn", handler.TrackIn)
		moves.POST("/trackOut", handler.TrackOut)
		moves.POST("/abort", handler.AbortTrackIn)
		moves.POST("/pass", handler.PassNode)

		group.GET("/eqpEquipment", handler.ListEquipment)
	})
}
