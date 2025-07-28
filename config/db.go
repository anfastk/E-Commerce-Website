package config

import (
	"fmt"
	"os"

	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB          *gorm.DB
	IsConfigErr bool
	ConfigErr   error
)

func DBconnect() {
	var err error

	/* dsn := os.Getenv("DB") */
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Error("Database connection failed", zap.Error(err))
		IsConfigErr = false
		ConfigErr = err
		return
	}

	logger.Log.Info("Database connected successfully")
	IsConfigErr = true
	ConfigErr = nil
}
