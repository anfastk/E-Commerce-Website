package config

import (
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

func InitCloudinary() *cloudinary.Cloudinary {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	Cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)

	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	return Cld
}
