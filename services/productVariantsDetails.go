package services

import (
	"errors"
	"fmt"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
)

type VariantsDetails struct {
	Id              uint
	ProductId       uint
	ProductName     string
	BrandName       string
	IsReturnable    bool
	IsCodAvailable  bool
	CategoryName    string
	Descriptions    []models.ProductDescription
	Images          []models.ProductVariantsImage
	OfferName       string
	OfferDetails    string
	OfferStartDate  string
	OfferEndDate    string
	OfferPercentage float64
	OfferAmount     float64
	Size            string
	Colour          string
	Ram             string
	Storage         string
	StockQuantity   int
	RegularPrice    float64
	SalePrice       float64
	SKU             string
	ProductSummary  string
	IsDeleted       bool
	Specification   []models.ProductSpecification
}

func ShowSingleProductVariantDetail(variantID uint) (VariantsDetails, error) {
	var productVariant models.ProductVariantDetails
	var product models.ProductDetail
	var descriptions []models.ProductDescription
	var images []models.ProductVariantsImage
	var offer models.ProductOffer
	var category models.Categories
	var specification []models.ProductSpecification

	tx := config.DB.Begin()

	if err := tx.Unscoped().Where("ID = ?", variantID).First(&productVariant).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("Product variant not found")
	}

	if err := tx.Unscoped().Where("ID = ?", productVariant.ProductID).First(&product).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("Main product not found")
	}

	if err := tx.First(&category, product.CategoryID).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("category not found")
	}

	if err := tx.Where("product_id = ? AND is_deleted = ?", product.ID, false).Find(&descriptions).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("descriptions not found")
	}

	if err := tx.Where("product_Variant_id = ? AND is_deleted = ?", variantID, false).Find(&images).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("images not found")
	}

	if err := tx.Where("product_id = ? AND is_valid = true", product.ID).First(&offer).Error; err != nil {
		fmt.Println("No valid offer found:", err)
		offer = models.ProductOffer{}
	}

	if err := tx.Where("product_variant_id = ? AND is_deleted = ?", variantID, false).Find(&specification).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("Specification not found")
	}

	tx.Commit()
	return VariantsDetails{
		Id:              productVariant.ID,
		ProductId:       product.ID,
		ProductName:     product.ProductName,
		BrandName:       product.BrandName,
		IsReturnable:    product.IsReturnable,
		IsCodAvailable:  product.IsCODAvailable,
		CategoryName:    category.Name,
		Descriptions:    descriptions,
		Images:          images,
		OfferName:       offer.OfferName,
		OfferDetails:    offer.OfferDetails,
		OfferStartDate:  offer.StartDate.Format("02-01-2006"),
		OfferEndDate:    offer.EndDate.Format("02-01-2006"),
		OfferPercentage: offer.OfferPercentage,
		OfferAmount:     offer.OfferAmount,
		Size:            productVariant.Size,
		Colour:          productVariant.Colour,
		Ram:             productVariant.Ram,
		Storage:         productVariant.Storage,
		StockQuantity:   productVariant.StockQuantity,
		RegularPrice:    productVariant.RegularPrice,
		SalePrice:       productVariant.SalePrice,
		SKU:             productVariant.SKU,
		ProductSummary:  productVariant.ProductSummary,
		Specification:   specification,
		IsDeleted:       productVariant.IsDeleted,
	}, nil
}

func ShowMultipleProductVariants(productID uint) ([]VariantsDetails, error) {
	var product models.ProductDetail
	var productVariants []models.ProductVariantDetails
	var descriptions []models.ProductDescription
	var offers models.ProductOffer
	var category models.Categories
	var result []VariantsDetails

	tx := config.DB.Begin()

	if err := tx.Unscoped().Where("ID = ?", productID).First(&product).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("Main product not found")
	}

	if err := tx.First(&category, product.CategoryID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("category not found")
	}

	if err := tx.Where("product_id = ? AND is_deleted = ?", product.ID, false).Find(&descriptions).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("descriptions not found")
	}

	if err := tx.Where("product_id = ? AND is_valid = true", product.ID).First(&offers).Error; err != nil {
		fmt.Println("No valid offer found:", err)
		offers = models.ProductOffer{}
	}

	if err := tx.Unscoped().Where("product_id = ?", productID).Find(&productVariants).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("Variants not found")
	}
	for _, variant := range productVariants {
		var images []models.ProductVariantsImage
		var specification []models.ProductSpecification

		if err := tx.Where("product_variant_id = ? AND is_deleted = ?", variant.ID, false).Find(&images).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("Images not found for variant")
		}

		if err := tx.Where("product_variant_id = ? AND is_deleted = ?", variant.ID, false).Find(&specification).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("Specifications not found for variant")
		}

		result = append(result, VariantsDetails{
			Id:              variant.ID,
			ProductId:       product.ID,
			ProductName:     product.ProductName,
			BrandName:       product.BrandName,
			IsReturnable:    product.IsReturnable,
			IsCodAvailable:  product.IsCODAvailable,
			CategoryName:    category.Name,
			Descriptions:    descriptions,
			Images:          images,
			OfferName:       offers.OfferName,
			OfferDetails:    offers.OfferDetails,
			OfferStartDate:  offers.StartDate.Format("02-01-2006"),
			OfferEndDate:    offers.EndDate.Format("02-01-2006"),
			OfferPercentage: offers.OfferPercentage,
			OfferAmount:     offers.OfferAmount,
			Size:            variant.Size,
			Colour:          variant.Colour,
			Ram:             variant.Ram,
			Storage:         variant.Storage,
			StockQuantity:   variant.StockQuantity,
			RegularPrice:    variant.RegularPrice,
			SalePrice:       variant.SalePrice,
			SKU:             variant.SKU,
			ProductSummary:  variant.ProductSummary,
			Specification:   specification,
			IsDeleted:       variant.IsDeleted,
		})
	}

	tx.Commit()
	return result, nil
}
