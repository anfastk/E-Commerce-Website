package config

import (
	"log"

	models "github.com/anfastk/E-Commerce-Website/models"
)

func SyncDatabase() {
	err:=DB.AutoMigrate(
		&models.AdminModel{},&models.UserAuth{},&models.Categories{},&models.ProductDetail{},&models.ProductImage{},
		&models.ProductOffer{},models.ProductDescription{},&models.ProductVariantsImage{},models.ProductVariantDetails{},&models.ProductSpecification{},
		&models.Otp{},&models.UserProfile{},&models.UserAddress{},
	)
	if err != nil{
		log.Fatalf("Failed to migrate models: %v",err)
	}
	log.Println("Models migrated")
}