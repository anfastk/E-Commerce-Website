package config

import (
	"os"

	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)
 
var (
	RAZORPAY_KEY_ID     string
	RAZORPAY_KEY_SECRET string
)

func LoadEnvFile() {
	err := godotenv.Load(".env")
	if err != nil {
		logger.Log.Error("Error loading .env fil", zap.Error(err))
		IsConfigErr = false
		ConfigErr = err
		return
	}

	RAZORPAY_KEY_ID = os.Getenv("RAZORPAY_KEY_ID")
	RAZORPAY_KEY_SECRET = os.Getenv("RAZORPAY_KEY_SECRET")
	IsConfigErr = true
	ConfigErr = nil
}
