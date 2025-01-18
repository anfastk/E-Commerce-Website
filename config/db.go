package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func DBconnect(){
	var err error
	dsn:= os.Getenv("DB")
	DB,err = gorm.Open(postgres.Open(dsn),&gorm.Config{})
	if err != nil{
		log.Fatalf("Failed to connect database: %v",err)
	}
	log.Println("Database connected successfully")
}