package config

import (
	"io"
	"log"
	"net/http"
	"os"
)

var CompanyConfig = struct {
	Name         string
	Address1     string
	Address2     string
	Email        string
	LogoURL      string
	LogoFilePath string
}{
	Name:         "LAPTIX",
	Address1:     "Laptix Ecom Pvt.Ltd.KINFRA SDF Building,",
	Address2:     "Kakkanchery,Malapuram, 673634",
	Email:        "laptixinfo@gmail.com",
	LogoURL:      "https://res.cloudinary.com/dghzlcoco/image/upload/v1740498507/text-1740498489427_ir9mat.png",
	LogoFilePath: "company_logo.png",
}

func DownloadLogo() {
	resp, err := http.Get(CompanyConfig.LogoURL)
	if err != nil {
		log.Printf("Error downloading logo: %v", err)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(CompanyConfig.LogoFilePath)
	if err != nil {
		log.Printf("Error creating logo file: %v", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Printf("Error saving logo: %v", err)
		return
	}

	log.Println("Logo downloaded successfully")
}
