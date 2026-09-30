package router

import (
	"semi-mes/server/internal/handler"
	"semi-mes/server/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.Use(middleware.CORS())

	authHandler := handler.NewAuthHandler(db)
	userHandler := handler.NewUserHandler(db)
	roleHandler := handler.NewRoleHandler(db)
	menuHandler := handler.NewMenuHandler(db)
	baseDataHandler := handler.NewBaseDataHandler(db)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.JWTAuth(), authHandler.Logout)
			auth.GET("/user-info", middleware.JWTAuth(), authHandler.GetUserInfo)
			auth.GET("/menus", middleware.JWTAuth(), authHandler.GetMenus)
		}

		users := api.Group("/users")
		users.Use(middleware.JWTAuth())
		{
			users.GET("", middleware.RequirePermission("system:user:query"), userHandler.List)
			users.POST("", middleware.RequirePermission("system:user:add"), userHandler.Create)
			users.PUT("/:id", middleware.RequirePermission("system:user:edit"), userHandler.Update)
			users.DELETE("/:id", middleware.RequirePermission("system:user:delete"), userHandler.Delete)
		}

		roles := api.Group("/roles")
		roles.Use(middleware.JWTAuth())
		{
			roles.GET("", middleware.RequirePermission("system:role:query"), roleHandler.List)
			roles.POST("", middleware.RequirePermission("system:role:add"), roleHandler.Create)
			roles.PUT("/:id", middleware.RequirePermission("system:role:edit"), roleHandler.Update)
			roles.DELETE("/:id", middleware.RequirePermission("system:role:delete"), roleHandler.Delete)
			roles.GET("/:id/menus", middleware.RequirePermission("system:role:query"), roleHandler.GetMenus)
		}

		menus := api.Group("/menus")
		menus.Use(middleware.JWTAuth())
		{
			menus.GET("", middleware.RequirePermission("system:menu:query"), menuHandler.List)
			menus.POST("", middleware.RequirePermission("system:menu:add"), menuHandler.Create)
			menus.PUT("/:id", middleware.RequirePermission("system:menu:edit"), menuHandler.Update)
			menus.DELETE("/:id", middleware.RequirePermission("system:menu:delete"), menuHandler.Delete)
		}

		baseData := api.Group("/base-data")
		baseData.Use(middleware.JWTAuth())
		{
			baseData.GET("/factories", middleware.RequirePermission("base:factory:query"), baseDataHandler.ListFactories)
			baseData.POST("/factories", middleware.RequirePermission("base:factory:add"), baseDataHandler.CreateFactory)
			baseData.PUT("/factories/:id", middleware.RequirePermission("base:factory:edit"), baseDataHandler.UpdateFactory)
			baseData.DELETE("/factories/:id", middleware.RequirePermission("base:factory:delete"), baseDataHandler.DeleteFactory)

			baseData.GET("/workshops", middleware.RequirePermission("base:workshop:query"), baseDataHandler.ListWorkshops)
			baseData.POST("/workshops", middleware.RequirePermission("base:workshop:add"), baseDataHandler.CreateWorkshop)
			baseData.PUT("/workshops/:id", middleware.RequirePermission("base:workshop:edit"), baseDataHandler.UpdateWorkshop)
			baseData.DELETE("/workshops/:id", middleware.RequirePermission("base:workshop:delete"), baseDataHandler.DeleteWorkshop)

			baseData.GET("/production-lines", middleware.RequirePermission("base:line:query"), baseDataHandler.ListProductionLines)
			baseData.POST("/production-lines", middleware.RequirePermission("base:line:add"), baseDataHandler.CreateProductionLine)
			baseData.PUT("/production-lines/:id", middleware.RequirePermission("base:line:edit"), baseDataHandler.UpdateProductionLine)
			baseData.DELETE("/production-lines/:id", middleware.RequirePermission("base:line:delete"), baseDataHandler.DeleteProductionLine)

			baseData.GET("/products", middleware.RequirePermission("base:product:query"), baseDataHandler.ListProducts)
			baseData.POST("/products", middleware.RequirePermission("base:product:add"), baseDataHandler.CreateProduct)
			baseData.PUT("/products/:id", middleware.RequirePermission("base:product:edit"), baseDataHandler.UpdateProduct)
			baseData.DELETE("/products/:id", middleware.RequirePermission("base:product:delete"), baseDataHandler.DeleteProduct)

			baseData.GET("/process-routes", middleware.RequirePermission("base:route:query"), baseDataHandler.ListProcessRoutes)
			baseData.POST("/process-routes", middleware.RequirePermission("base:route:add"), baseDataHandler.CreateProcessRoute)
			baseData.PUT("/process-routes/:id", middleware.RequirePermission("base:route:edit"), baseDataHandler.UpdateProcessRoute)
			baseData.DELETE("/process-routes/:id", middleware.RequirePermission("base:route:delete"), baseDataHandler.DeleteProcessRoute)

			baseData.GET("/operations", middleware.RequirePermission("base:operation:query"), baseDataHandler.ListOperations)
			baseData.POST("/operations", middleware.RequirePermission("base:operation:add"), baseDataHandler.CreateOperation)
			baseData.PUT("/operations/:id", middleware.RequirePermission("base:operation:edit"), baseDataHandler.UpdateOperation)
			baseData.DELETE("/operations/:id", middleware.RequirePermission("base:operation:delete"), baseDataHandler.DeleteOperation)

			baseData.GET("/recipes", middleware.RequirePermission("base:recipe:query"), baseDataHandler.ListRecipes)
			baseData.POST("/recipes", middleware.RequirePermission("base:recipe:add"), baseDataHandler.CreateRecipe)
			baseData.PUT("/recipes/:id", middleware.RequirePermission("base:recipe:edit"), baseDataHandler.UpdateRecipe)
			baseData.DELETE("/recipes/:id", middleware.RequirePermission("base:recipe:delete"), baseDataHandler.DeleteRecipe)
		}
	}

	return r
}
