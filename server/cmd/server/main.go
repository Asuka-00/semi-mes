package main

import (
	"fmt"
	"log"

	"semi-mes/server/internal/config"
	"semi-mes/server/internal/middleware"
	"semi-mes/server/internal/router"
	"semi-mes/server/migrations"
	"semi-mes/server/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load("./config/config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := utils.InitDB(config.GlobalConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if err := migrations.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := migrations.SeedData(db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	middleware.InitPermissionMiddleware(db)

	if config.GlobalConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := router.SetupRouter(db)

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
