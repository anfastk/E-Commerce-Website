package config

import (
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

    /* dbHost := os.Getenv("DB_HOST")     // "postgres-service"
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dbPort := os.Getenv("DB_PORT") */

    dsn := os.Getenv("DB")

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

