package config

import (
	"io"
	"net/http"
	"os"

	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"go.uber.org/zap"
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
		logger.Log.Error("Error downloading logo", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(CompanyConfig.LogoFilePath)
	if err != nil {
		logger.Log.Error("Error creating logo file", zap.Error(err))
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		logger.Log.Error("Error saving log", zap.Error(err))
		return
	}
	IsConfigErr = true
	ConfigErr = nil
	logger.Log.Info("Logo downloaded successfully")
}
