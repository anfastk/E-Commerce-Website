package services

import (
	"errors"

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

	if err := tx.Order("created_at ASC").Where("product_Variant_id = ? AND is_deleted = ?", variantID, false).Find(&images).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("images not found")
	}

	if err := tx.Where("product_variant_id = ? AND is_deleted = ?", variantID, false).Find(&specification).Error; err != nil {
		tx.Rollback()
		return VariantsDetails{}, errors.New("Specification not found")
	}

	if err := tx.Where("product_id = ?", product.ID).First(&offer).Error; err != nil {
		offer = models.ProductOffer{}
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
