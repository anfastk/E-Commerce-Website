package helper

import (
	"errors"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
)

func DiscountCalculation(productID uint, categoryID uint, regularPrice float64, salePrice float64) (float64, float64, error) {
	if regularPrice <= 0 || salePrice < 0 {
		return 0, 0, errors.New("invalid price values")
	}
 
	var productOffer models.ProductOffer
	var categoryOffer models.OfferByCategory

	if err := config.DB.First(&productOffer, "product_id = ? AND status = ?", productID, "Active").Error; err != nil {
		productOffer.OfferPercentage = 0 
	}

	if err := config.DB.First(&categoryOffer, "category_id = ? AND offer_status = ?", categoryID, "Active").Error; err != nil {
		categoryOffer.CategoryOfferPercentage = 0 
	}

	discountPercentage := productOffer.OfferPercentage
	if categoryOffer.CategoryOfferPercentage > productOffer.OfferPercentage {
		discountPercentage = categoryOffer.CategoryOfferPercentage
	}

	discountAmount := 0.0
	if discountPercentage > 0 {
		discountAmount = regularPrice * discountPercentage / 100
	}

	difference := regularPrice - salePrice
	productDiscountPercentage := 0.0
	if regularPrice > 0 {
		productDiscountPercentage = (difference / regularPrice) * 100
	}

	totalPercentage := productDiscountPercentage + discountPercentage

	return discountAmount, totalPercentage, nil
}
