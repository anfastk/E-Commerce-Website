package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/gin-gonic/gin"
)

type ProductVariantResponse struct {
	ID             uint     `json:"id"`
	ProductName    string   `json:"productname"`
	CategoryName   string   `json:"category_name"`
	RegularPrice   float64  `json:"regular_price"`
	SalePrice      float64  `json:"sale_price"`
	ProductSummary string   `json:"product_summary"`
	Images         []string `json:"images"`
}

func ShowProductsAdmin(c *gin.Context) {
	var variants []models.ProductVariantDetails

	result := config.DB.Preload("VariantsImages", "is_deleted = ?", false).
		Preload("Category", "is_deleted = ? AND status = ?", false, "Active").
		Preload("Product").
		Where("product_variant_details.is_deleted = ?", false).
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

		categoryName := "Uncategorized" // Default value
		if variant.Category.ID != 0 {   // Check if category exists
			categoryName = variant.Category.Name
		}

		productName := variant.ProductName
        if productName == "" && variant.Product.ID != 0 {
            productName = variant.ProductName 
        }

		response = append(response, ProductVariantResponse{
			ID:             variant.ID,
			ProductName:    variant.ProductName,
			CategoryName:   categoryName,
			RegularPrice:   variant.RegularPrice,
			SalePrice:      variant.SalePrice,
			ProductSummary: variant.ProductSummary,
			Images:         images,
		})
	}
	var formattedResponceDetails []map[string]interface{}
	for  _,variant := range response {
		formattedVariant:= map[string]interface{}{
			"ID":             variant.ID,
			"ProductName":    variant.ProductName,
			"CategoryName":   variant.CategoryName,
			"RegularPrice":   variant.RegularPrice,
			"SalePrice":      fmt.Sprintf("%.2f",variant.SalePrice),
			"ProductSummary": variant.ProductSummary,
			"Images":         variant.Images,
		}
		formattedResponceDetails = append(formattedResponceDetails, formattedVariant)
	}

	c.HTML(http.StatusFound, "productPageAdmin.html", gin.H{
		"status":  true,
		"message": "Product variants fetched successfully",
		"data":    formattedResponceDetails,
	})

}

func ShowAddMainProduct(c *gin.Context) {
	var categories []models.Categories
	if err := config.DB.Where("is_deleted = ? AND status = ?", false, "Active").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "Failed to fetch categories",
			"error":   err.Error(),
			"code":    500,
		})
		return
	}

	c.HTML(http.StatusOK, "addNewMainProductDetails.html", gin.H{
		"categories": categories,
	})
}
func AddMainProductDetails(c *gin.Context) {
	tx := config.DB.Begin()
	categoryID, err := strconv.ParseInt(c.PostForm("category"), 10, 64)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "Invalid category ID",
			"error":   err.Error(),
			"code":    500,
		})
		return
	}
	product := models.ProductDetail{
		ProductName:    c.PostForm("product_name"),
		CategoryID:     uint(categoryID),
		BrandName:      c.PostForm("brand_name"),
		IsCODAvailable: c.PostForm("cod_available") == "YES",
		IsReturnable:   c.PostForm("return_available") == "YES",
	}
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "Internal Server Error",
			"message": "Failed to save product",
			"error":   err.Error(),
			"code":    500,
		})
		return
	}
	cld := config.InitCloudinary()

	form, _ := c.MultipartForm()
	if form != nil {
		if productImage, ok := form.File["product_image"]; ok && len(productImage) > 0 {
			fileHeader := productImage[0]
			file, _ := fileHeader.Open()
			defer file.Close()

			url, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "products")
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "Internal Server Error",
					"message": "Failed to upload product image",
					"error":   err.Error(),
					"code":    500,
				})
				return
			}
			image := models.ProductImage{
				ProductImages: url,
				ProductID:     product.ID,
			}
			if err := tx.Create(&image).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "Internal Server Error",
					"message": "Failed to upload product image",
					"error":   err.Error(),
					"code":    500,
				})
				return
			}
		}
	}
	tx.Commit()
	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(int(product.ID))
	c.Redirect(http.StatusFound, redirectURL)

}

func ShowMainProductDetails(c *gin.Context) {
	productID, err := strconv.Atoi(c.Query("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}

	productDetails, err := services.ShowMainProductsDetails(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to fetch product details",
			"code":   500,
		})
		return
	}

	c.HTML(http.StatusSeeOther, "mainProductDetails.html", gin.H{
		"Product": productDetails,
	})
}

func AddProductDescription(c *gin.Context) {
	productID, err := strconv.Atoi(c.PostForm("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}
	headings := c.PostFormArray("heading[]")
	descriptions := c.PostFormArray("description[]")
	if len(headings) != len(descriptions) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Mismatch in headings and descriptions",
			"code":   400,
		})
		return
	}
	for i := 0; i < len(headings); i++ {
		description := models.ProductDescription{
			ProductID:   uint(productID),
			Heading:     headings[i],
			Description: descriptions[i],
		}
		if err := config.DB.Create(&description).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal Server Error",
				"error":  "Failed to save description",
				"code":   500,
			})
			return
		}
	}
	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(int(productID))
	c.Redirect(http.StatusFound, redirectURL)
}
