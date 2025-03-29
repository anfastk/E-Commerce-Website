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
	return
}
