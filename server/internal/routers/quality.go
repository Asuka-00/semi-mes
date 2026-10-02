package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		plans := group.Group("/qcInspectPlan")
		plans.GET("", handler.ListInspectPlans)
		plans.POST("", handler.SaveInspectPlan)
		plans.PUT("/:id", handler.SaveInspectPlan)
		plans.DELETE("/:id", handler.DeleteInspectPlan)

		measure := group.Group("/qcMeasurement")
		measure.POST("", handler.RecordMeasurement)

		codes := group.Group("/qcDefectCode")
		codes.GET("", handler.ListDefectCodes)
		codes.POST("", handler.SaveDefectCode)
		codes.PUT("/:id", handler.SaveDefectCode)

		defects := group.Group("/qcDefect")
		defects.GET("", handler.ListDefects)
		defects.GET("/pareto", handler.DefectPareto)
		defects.POST("", handler.RecordDefect)

		spc := group.Group("/qcSpc")
		spc.GET("/chart", handler.SpcChart)
		spc.GET("/policy", handler.ListSpcPolicies)
		spc.PUT("/policy", handler.SaveSpcPolicy)
		spc.PUT("/limit", handler.SaveSpcLimit)
	})
}
