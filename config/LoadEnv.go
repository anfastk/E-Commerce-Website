package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

func LoadEnvFile() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading .env file %v",err)
	}
}