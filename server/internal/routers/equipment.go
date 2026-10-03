package routers

import (
	"github.com/gin-gonic/gin"

	"semi-mes/server/internal/handler"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		eqp := group.Group("/eqpEquipment")
		eqp.POST("/list", handler.ListEquipmentPage)
		eqp.GET("/board", handler.EquipmentBoard)
		eqp.POST("/:id/state", handler.ChangeEquipmentState)
		eqp.GET("/:id", handler.GetEquipment)
		eqp.PUT("/:id", handler.UpdateEquipment)
		eqp.DELETE("/:id", handler.DeleteEquipment)
		eqp.POST("", handler.CreateEquipment)

		plans := group.Group("/eqpPmPlan")
		plans.GET("", handler.ListPmPlans)
		plans.POST("/list", handler.ListPmPlanPage)
		plans.POST("", handler.SavePmPlan)
		plans.PUT("/:id", handler.SavePmPlan)
		plans.DELETE("/:id", handler.DeletePmPlan)

		tasks := group.Group("/eqpPmTask")
		tasks.POST("/list", handler.ListPmTasks)
		tasks.POST("/generate", handler.GeneratePmTasks)
		tasks.POST("/:id/start", handler.StartPmTask)
		tasks.POST("/:id/complete", handler.CompletePmTask)
	})
}
