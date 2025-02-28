package main

import (
	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/routes"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func init() {
	config.LoadEnvFile()
	r = gin.Default()
	r.Static("static","./static")
	r.LoadHTMLGlob("views/**/*")
	config.DBconnect()
	config.SyncDatabase()
	config.InitializeGoogleOAuth()
}

func main(){
	routes.AdminRoutes(r)
	routes.UserRouter(r)
	services.StartReservationCleanupTask(config.DB)
	r.Run()
}