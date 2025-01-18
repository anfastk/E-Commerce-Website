package services

import (
	"errors"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
)

type MainProductDetails struct {
	Id              uint
	ProductName     string
	BrandName       string
	IsReturnable    bool
	IsCodAvailable  bool
	CategoryName    string
	Descriptions    []models.ProductDescription
	Images          string
	OfferName       string
	OfferDetails    string
	OfferStartDate  string
	OfferEndDate    string
	OfferPercentage float64
	OfferAmount     float64
}

func ShowMainProductsDetails(productID uint) (MainProductDetails, error) {
	var product models.ProductDetail
	var descriptions []models.ProductDescription
	var images models.ProductImage
	var offer models.ProductOffer
	var category models.Categories

	tx := config.DB.Begin()

	if err := tx.Where("ID = ? AND is_deleted = ? ", productID, false).First(&product).Error; err != nil {
		tx.Rollback()
		return MainProductDetails{}, errors.New("Product not found")
	}

	if err := tx.First(&category, product.CategoryID).Error; err != nil {
		tx.Rollback()
		return MainProductDetails{}, errors.New("category not found")
	}

	if err := tx.Where("product_id = ? AND is_deleted = ?", productID, false).Find(&descriptions).Error; err != nil {
		tx.Rollback()
		return MainProductDetails{}, errors.New("descriptions not found")
	}

	if err := tx.Where("product_id = ? AND is_deleted = ?", productID, false).First(&images).Error; err != nil {
		tx.Rollback()
		return MainProductDetails{}, errors.New("images not found")
	}

	if err := tx.Where("product_id = ? AND is_valid = true", productID).First(&offer).Error; err != nil {
		offer = models.ProductOffer{}
	}
	tx.Commit()
	return MainProductDetails{
		Id:              product.ID,
		ProductName:     product.ProductName,
		BrandName:       product.BrandName,
		IsReturnable:    product.IsReturnable,
		IsCodAvailable:  product.IsCODAvailable,
		CategoryName:    category.Name,
		Descriptions:    descriptions,
		Images:          images.ProductImages,
		OfferName:       offer.OfferName,
		OfferDetails:    offer.OfferDetails,
		OfferStartDate:  offer.StartDate.Format("02-01-2006"),
		OfferEndDate:    offer.EndDate.Format("02-01-2006"),
		OfferPercentage: offer.OfferPercentage,
		OfferAmount:     offer.OfferAmount,
	}, nil
}
