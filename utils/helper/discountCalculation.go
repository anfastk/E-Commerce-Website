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

	difference := regularPrice - salePrice
	productDiscountPercentage := (difference / regularPrice) * 100

	discountPercentage := productOffer.OfferPercentage + productDiscountPercentage
	if categoryOffer.CategoryOfferPercentage > discountPercentage {
		discountPercentage = categoryOffer.CategoryOfferPercentage
	}
	if discountPercentage > 100 {
		discountPercentage = 100
	}

	discountAmount := (regularPrice * discountPercentage / 100) - difference

	return discountAmount, discountPercentage, nil
}
