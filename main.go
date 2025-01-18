package main

import (
	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/routes"
	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func init() {
	r = gin.Default()
	r.Static("static","./static")
	r.LoadHTMLGlob("views/**/*")
	config.LoadEnvFile()
	config.DBconnect()
	config.SyncDatabase()
	config.InitializeGoogleOAuth()
}

func main(){
	routes.AdminRoutes(r)
	routes.UserRouter(r)
	r.Run()
}