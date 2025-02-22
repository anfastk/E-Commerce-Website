package controllers

import (
	"net/http"
	"strings"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
)

func UserHome(c *gin.Context) {
	
	keyboard, _ := helper.RelatedProducts(2)
	laptop, _ := helper.RelatedProducts(3)
	mouse, _ := helper.RelatedProducts(4)

	c.HTML(http.StatusOK, "userHome.html", gin.H{
		"Keyboard": keyboard,
		"Laptop":   laptop,
		"Mouse":    mouse,
	})
}

type ProductVariantResponse struct {
	ID           uint     `json:"id"`
	ProductName  string   `json:"product_name"`
	RegularPrice float64  `json:"regular_price"`
	SalePrice    float64  `json:"sale_price"`
	Images       []string `json:"images"`
}

func ShowProducts(c *gin.Context) {
	var Brand []string
	var Category []string

	if err := config.DB.Model(&models.ProductDetail{}).Distinct("brand_name").Where("is_deleted =?",false).Pluck("brand_name", &Brand).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch brand name", "Something Went Wrong", "")
		return
	}
	if err := config.DB.Model(&models.Categories{}).Where("is_deleted =?",false).Pluck("name", &Category).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch category name", "Something Went Wrong", "")
		return
	}

	var variants []models.ProductVariantDetails

	result := config.DB.Preload("VariantsImages", "is_deleted = ?", false).
		Preload("Category").
		Preload("Product").
		Where("is_deleted = ? AND stock_quantity>0", false).
		Find(&variants)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch product variants",
		})
		return
	}

	var response []ProductVariantResponse
	for _, variant := range variants {
		var images []string
		for _, img := range variant.VariantsImages {
			images = append(images, img.ProductVariantsImages)
		}

		response = append(response, ProductVariantResponse{
			ID:           variant.ID,
			ProductName:  variant.ProductName,
			RegularPrice: variant.RegularPrice,
			SalePrice:    variant.SalePrice,
			Images:       images,
		})
	}

	c.HTML(http.StatusFound, "productpage.html", gin.H{
		"status":   true,
		"message":  "Product variants fetched successfully",
		"data":     response,
		"Brand":    Brand,
		"Category": Category,
	})
}

type FilterRequest struct {
	Search            string   `json:"search"`
	Sort              string   `json:"sort"`
	Categories        []string `json:"categories"`
	PriceRanges       []string `json:"priceRanges"`
	Discounts         []int    `json:"discounts"`
	Brands            []string `json:"brands"`
	IncludeOutOfStock bool     `json:"includeOutOfStock"`
}

func FilterProducts(c *gin.Context) {
	var filter FilterRequest
	if err := c.ShouldBindJSON(&filter); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid filter parameters", err.Error(), "")
		return
	}

	query := config.DB.Model(&models.ProductVariantDetails{}).
		Preload("VariantsImages", "is_deleted = ?", false).
		Preload("Category").
		Preload("Product").
		Where("product_variant_details.is_deleted = ?", false)

	query = query.Joins("JOIN product_details ON product_variant_details.product_id = product_details.id")

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Joins("JOIN categories ON product_variant_details.category_id = categories.id").
			Where("product_variant_details.product_name ILIKE ? OR product_details.brand_name ILIKE ? OR categories.name ILIKE ?",
				searchTerm, searchTerm, searchTerm)
	} else if len(filter.Categories) > 0 {
		query = query.Joins("JOIN categories ON product_variant_details.category_id = categories.id").
			Where("categories.name IN ?", filter.Categories)
	}

	if len(filter.Brands) > 0 {
		query = query.Where("product_details.brand_name IN ?", filter.Brands)
	}

	if len(filter.PriceRanges) > 0 {
		priceConditions := []string{}
		for _, pr := range filter.PriceRanges {
			switch pr {
			case "1":
				priceConditions = append(priceConditions, "(product_variant_details.sale_price BETWEEN 1000 AND 50000)")
			case "2":
				priceConditions = append(priceConditions, "(product_variant_details.sale_price BETWEEN 50000 AND 100000)")
			case "3":
				priceConditions = append(priceConditions, "(product_variant_details.sale_price BETWEEN 100000 AND 500000)")
			}
		}
		if len(priceConditions) > 0 {
			query = query.Where(strings.Join(priceConditions, " OR "))
		}
	}

	if len(filter.Discounts) > 0 {
		query = query.Where("((product_variant_details.regular_price - product_variant_details.sale_price) / product_variant_details.regular_price * 100) >= ?",
			filter.Discounts[len(filter.Discounts)-1])
	}

	if !filter.IncludeOutOfStock {
		query = query.Where("product_variant_details.stock_quantity > 0")
	}

	switch filter.Sort {
	/* case "popularity":
	query = query.Order("sales_count DESC") */
	case "price-low":
		query = query.Order("product_variant_details.sale_price ASC")
	case "price-high":
		query = query.Order("product_variant_details.sale_price DESC")
	case "newest":
		query = query.Order("product_variant_details.created_at DESC")
	default:
		query = query.Order("product_variant_details.created_at DESC")
	}

	var variants []models.ProductVariantDetails
	if err := query.Find(&variants).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch products", err.Error(), "")
		return
	}

	var response []ProductVariantResponse
	for _, variant := range variants {
		var images []string
		for _, img := range variant.VariantsImages {
			images = append(images, img.ProductVariantsImages)
		}

		response = append(response, ProductVariantResponse{
			ID:           variant.ID,
			ProductName:  variant.ProductName,
			RegularPrice: variant.RegularPrice,
			SalePrice:    variant.SalePrice,
			Images:       images,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Products filtered successfully",
		"data":    response,
	})
}

type ProductDetailResponse struct {
	ID             uint                    `json:"id"`
	ProductName    string                  `json:"product_name"`
	CategoryName   string                  `json:"category_name"`
	CategoryID     uint                    `json:"category_id"`
	RegularPrice   float64                 `json:"regular_price"`
	SalePrice      float64                 `json:"sale_price"`
	Images         []string                `json:"images"`
	Size           string                  `json:"size"`
	Color          string                  `json:"color"`
	Ram            string                  `json:"ram"`
	Storage        string                  `json:"storage"`
	Stock          int                     `json:"stock"`
	Summary        string                  `json:"summary"`
	Specifications []SpecificationResponse `json:"specifications"`
	Description    []DescriptionResponse   `json:"description"`
}

type DescriptionResponse struct {
	Heading     string `json:"heading"`
	Description string `json:"description "`
}

type SpecificationResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func ShowProductDetail(c *gin.Context) {
	productID := c.Param("id")

	var variant models.ProductVariantDetails

	result := config.DB.Preload("VariantsImages", "is_deleted = ?", false).
		Preload("Category", "is_deleted = ?", false).
		Preload("Specification", "is_deleted = ?", false).
		Preload("Product.Descriptions", "is_deleted = ?", false).
		Where("id = ? AND is_deleted = ?", productID, false).
		First(&variant)

	if result.Error != nil {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	var images []string
	for _, img := range variant.VariantsImages {
		images = append(images, img.ProductVariantsImages)
	}

	var specs []SpecificationResponse
	for _, spec := range variant.Specification {
		specs = append(specs, SpecificationResponse{
			Key:   spec.SpecificationKey,
			Value: spec.SpecificationValue,
		})
	}

	var description []DescriptionResponse
	for _, descrip := range variant.Product.Descriptions {
		description = append(description, DescriptionResponse{
			Heading:     descrip.Heading,
			Description: descrip.Description,
		})
	}

	var relatedProducts []models.ProductVariantDetails
	config.DB.Preload("VariantsImages", "is_deleted = ?", false).
		Where("category_id = ? AND id != ? AND is_deleted = ?", variant.CategoryID, variant.ID, false).
		Limit(20).
		Find(&relatedProducts)

	type RelatedProductsResponce struct {
		ID             uint     `json:"id"`
		ProductName    string   `json:"product_name"`
		ProductSummary string   `json:"product_summary"`
		SalePrice      float64  `json:"sale_price "`
		Images         []string `json:"images"`
	}
	var relatedProductsResponce []RelatedProductsResponce

	for _, product := range relatedProducts {
		var images []string
		for _, image := range product.VariantsImages {
			images = append(images, image.ProductVariantsImages)
		}
		relatedProductsResponce = append(relatedProductsResponce, RelatedProductsResponce{
			ID:             product.ID,
			ProductName:    product.ProductName,
			ProductSummary: product.ProductSummary,
			SalePrice:      product.SalePrice,
			Images:         images,
		})
	}

	product := ProductDetailResponse{
		ID:             variant.ID,
		ProductName:    variant.ProductName,
		CategoryName:   variant.Category.Name,
		CategoryID:     variant.CategoryID,
		RegularPrice:   variant.RegularPrice,
		SalePrice:      variant.SalePrice,
		Images:         images,
		Size:           variant.Size,
		Color:          variant.Colour,
		Ram:            variant.Ram,
		Storage:        variant.Storage,
		Stock:          variant.StockQuantity,
		Summary:        variant.ProductSummary,
		Specifications: specs,
		Description:    description,
	}

	c.HTML(http.StatusFound, "productDetails.html", gin.H{
		"product":         product,
		"relatedProducts": relatedProductsResponce,
	})
}
