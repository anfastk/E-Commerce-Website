package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)
var (
	RAZORPAY_KEY_ID string
	RAZORPAY_KEY_SECRET string
)

func LoadEnvFile() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file %v", err)
	}
    
    RAZORPAY_KEY_ID = os.Getenv("RAZORPAY_KEY_ID")
    RAZORPAY_KEY_SECRET = os.Getenv("RAZORPAY_KEY_SECRET")

}
