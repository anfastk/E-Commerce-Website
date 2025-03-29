package config

import (
	"os"

	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/cloudinary/cloudinary-go/v2"
	"go.uber.org/zap"
)

func InitCloudinary() *cloudinary.Cloudinary {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	Cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)

	if err != nil {
		logger.Log.Error("Failed to initialize Cloudinary", zap.Error(err))
		IsConfigErr = false
		ConfigErr = err
		return nil
	}
	IsConfigErr = true
	ConfigErr = nil
	return Cld
} 
