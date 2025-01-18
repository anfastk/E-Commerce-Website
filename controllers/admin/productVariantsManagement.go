package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShowProductVariant(c *gin.Context) {
	var images models.ProductImage
	productID := c.Param("id")
	if err := config.DB.Where("product_id = ? AND is_deleted = ?", productID, false).First(&images).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Image not find",
			"code":   500,
		})
		return
	}
	c.HTML(http.StatusSeeOther, "addProductVariants.html", gin.H{
		"Images": images,
	})
}

func AddProductVariants(c *gin.Context) {
	productID, err := strconv.Atoi(c.PostForm("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}
	var mainProduct models.ProductDetail

	if err:=config.DB.First(&mainProduct,productID).Error;err!=nil{
		c.JSON(http.StatusNotFound, gin.H{
            "status": "Not Found",
            "error":  "Product not found",
            "code":   404,
        })
        return
	}

	categoryID:=mainProduct.CategoryID

	productNames := c.PostFormArray("product-name[]")
	productSummaries := c.PostFormArray("product-summary[]")
	sizes := c.PostFormArray("size[]")
	colors := c.PostFormArray("color[]")
	rams := c.PostFormArray("ram[]")
	storages := c.PostFormArray("storage[]")
	regularPrices := c.PostFormArray("regular-price[]")
	salePrices := c.PostFormArray("sale-price[]")
	stockQuantities := c.PostFormArray("stock-quantity[]")
	skus := c.PostFormArray("sku[]")

	formLength := len(productNames)
	if formLength == 0 || formLength != len(productSummaries) || formLength != len(sizes) || formLength != len(colors) ||
		formLength != len(rams) || formLength != len(storages) || formLength != len(regularPrices) ||
		formLength != len(salePrices) || formLength != len(stockQuantities) || formLength != len(skus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Mismatched form data lengths",
			"code":   400,
		})
		return
	}

	tx := config.DB.Begin()
	cld := config.InitCloudinary()

	for i := 0; i < formLength; i++ {
		regularPrice, err := strconv.ParseFloat(regularPrices[i], 64)
		salePrice, err2 := strconv.ParseFloat(salePrices[i], 64)
		stockQuantity, err3 := strconv.Atoi(stockQuantities[i])
		if err != nil || err2 != nil || err3 != nil {
			tx.Rollback()
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"status": "Unprocessable Entity",
				"error":  "Invalid price or quantity values",
				"code":   422,
			})
			return
		}

		productVariant := models.ProductVariantDetails{
			ProductID:      uint(productID),
			ProductName: productNames[i],
			Size:           sizes[i],
			Colour:         colors[i],
			Ram:            rams[i],
			Storage:        storages[i],
			RegularPrice:   regularPrice,
			SalePrice:      salePrice,
			StockQuantity:  stockQuantity,
			SKU:            skus[i],
			ProductSummary: productSummaries[i],
			CategoryID: categoryID,
		}

		if err := tx.Create(&productVariant).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal Server Error",
				"error":  "Failed to save product variant",
				"code":   500,
			})
			return
		}

		form, _ := c.MultipartForm()
		if form != nil {
			files, ok := form.File["product_images[]"]
			if ok {
				for _, fileHeader := range files {
					file, _ := fileHeader.Open()
					defer file.Close()

					url, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "ProductVariants")
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

					variantImage := models.ProductVariantsImage{
						ProductVariantsImages: url,
						ProductVariantID:      productVariant.ID,
					}
					if err := tx.Create(&variantImage).Error; err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{
							"status":  "Internal Server Error",
							"message": "Failed to save product image",
							"error":   err.Error(),
							"code":    500,
						})
						return
					}
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{
					"status":  "Bad Request",
					"message": "No files found in the request",
					"code":    400,
				})
				return
			}
		}
	}

	tx.Commit()
	redirectURL := "/admin/products/variant/details?product_id=" + strconv.Itoa(int(productID))
	c.Redirect(http.StatusFound, redirectURL)
}

func ShowMutiProductVariantDetails(c *gin.Context) {
	productID, err := strconv.Atoi(c.Query("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}
	variantDetails, err := services.ShowMultipleProductVariants(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to fetch product details",
			"code":   500,
		})
		return
	}

	c.HTML(http.StatusSeeOther, "productAllVariantsDetails.html", gin.H{
		"Product": variantDetails,
	})
}

func ShowSingleProductVariantDetail(c *gin.Context) {
	variantID, idErr := strconv.Atoi(c.Query("variant_id"))
	if idErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}
	variantDetails, err := services.ShowSingleProductVariantDetail(uint(variantID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to fetch product details",
			"code":   500,
		})
		return
	}
	c.HTML(http.StatusSeeOther, "productVariantDetails.html", gin.H{
		"Variant": variantDetails,
	})
}

func AddProductSpecification(c *gin.Context) {
	variantID, err := strconv.Atoi(c.PostForm("variant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product variant ID",
			"code":   400,
		})
		return
	}
	headings := c.PostFormArray("key[]")
	specification := c.PostFormArray("value[]")
	if len(headings) != len(specification) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Mismatch in headings and specification",
			"code":   400,
		})
		return
	}
	for i := 0; i < len(headings); i++ {
		specification := models.ProductSpecification{
			ProductVariantID:   uint(variantID),
			SpecificationKey:   headings[i],
			SpecificationValue: specification[i],
		}
		if err := config.DB.Create(&specification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal Server Error",
				"error":  "Failed to save specification",
				"code":   500,
			})
			return
		}
	}
	redirectURL := "/admin/products/variant/detail?variant_id=" + strconv.Itoa(int(variantID))
	c.Redirect(http.StatusFound, redirectURL)
}

func DeleteProductVariant(c *gin.Context) {
	var variant models.ProductVariantDetails

	variantID := c.Param("id")

	if err := config.DB.First(&variant, "ID = ? AND is_deleted = ?", variantID, false).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Product not found",
			"code":   500,
		})
		return
	}

	variant.IsDeleted = true
	variant.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := config.DB.Save(&variant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "InternalServerError",
			"error":  "Failed to delete product",
			"code":   500,
		})
		return
	}
	redirectURL := "/admin/products/variant/details?product_id=" + strconv.Itoa(int(variant.ProductID))
	c.Redirect(http.StatusFound, redirectURL)
}
func DeleteVariantImage(c *gin.Context) {
	imageID := c.Param("id")

	var variantImage models.ProductVariantsImage

	if err := config.DB.First(&variantImage, "ID = ? AND is_deleted = ?", imageID, false).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Image not found",
			"code":   500,
		})
		return
	}

	variantImage.IsDeleted = true
	variantImage.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	if err := config.DB.Save(&variantImage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "InternalServerError",
			"error":  "Failed to delete image",
			"code":   500,
		})
		return
	}
	redirectURL := "/admin/products/variant/details?product_id=" + strconv.Itoa(int(variantImage.ProductVariantID))
	c.Redirect(http.StatusFound, redirectURL)
}
