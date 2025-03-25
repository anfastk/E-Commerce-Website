package main

import (
	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/routes"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func init() {
	logger.InitLogger()
	config.LoadEnvFile()
	r = gin.Default()
	r.Static("static", "./static")
	r.LoadHTMLGlob("views/**/*")
	config.DBconnect()
	r.Use(middleware.DBRecoveryMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(middleware.NoCacheMiddleware())
	config.SyncDatabase()
	config.InitializeGoogleOAuth() 
}

func main() {
	routes.AdminRoutes(r)
	routes.UserRouter(r)
	services.StartReservationCleanupTask(config.DB)
	logger.Log.Info("E-commerce website started!")
	r.Run()
}
