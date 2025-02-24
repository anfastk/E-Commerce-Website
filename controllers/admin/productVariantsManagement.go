package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ShowProductVariant(c *gin.Context) {
	var images models.ProductImage
	productID := c.Param("id")
	if err := config.DB.Where("product_id = ? AND is_deleted = ?", productID, false).Find(&images).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Image not found", "Image not found", "")
		return
	}
	c.HTML(http.StatusSeeOther, "addProductVariants.html", gin.H{
		"Images": images,
	})
}

func AddProductVariants(c *gin.Context) {
	productID, err := strconv.Atoi(c.PostForm("product_id"))
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	var mainProduct models.ProductDetail

	if err := config.DB.First(&mainProduct, productID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Product not found", "")
		return
	}

	categoryID := mainProduct.CategoryID

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
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatched form data lengths", "Form Data Error", "")
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
			helper.RespondWithError(c, http.StatusUnprocessableEntity, "Invalid price or quantity values", "Validation Error", "")
			return
		}

		productVariant := models.ProductVariantDetails{
			ProductID:      uint(productID),
			ProductName:    productNames[i],
			Size:           sizes[i],
			Colour:         colors[i],
			Ram:            rams[i],
			Storage:        storages[i],
			RegularPrice:   regularPrice,
			SalePrice:      salePrice,
			StockQuantity:  stockQuantity,
			SKU:            skus[i],
			ProductSummary: productSummaries[i],
			CategoryID:     categoryID,
		}

		if err := tx.Create(&productVariant).Error; err != nil {
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product variant", "Database Error", "")
			return
		}

		form, _ := c.MultipartForm()
		if form != nil {
			files := form.File[fmt.Sprintf("product_images[%d][]", i)]
			if files != nil {
				for _, fileHeader := range files {
					file, _ := fileHeader.Open()
					defer file.Close()

					url, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "ProductVariants","")
					if err != nil {
						tx.Rollback()
						helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Upload Error", "")
						return
					}

					variantImage := models.ProductVariantsImage{
						ProductVariantsImages: url,
						ProductVariantID:      productVariant.ID,
					}
					if err := tx.Create(&variantImage).Error; err != nil {
						tx.Rollback()
						helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product image", "Database Error", "")
						return
					}
				}
			}
		}
	}

	tx.Commit()
	redirectURL := "/admin/products/variant/details?product_id=" + strconv.Itoa(int(productID))
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"code":     200,
		"redirect": redirectURL,
		"message":  "Product variants added successfully",
	})
}

func ShowMutiProductVariantDetails(c *gin.Context) {
	productID, err := strconv.Atoi(c.Query("product_id"))
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid Input", "")
		return
	}
	variantDetails, err := services.ShowMultipleProductVariants(uint(productID))
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product details", "Database Error", "")
		return
	}

	var formattedVariantDetails []map[string]interface{}

	for _, variant := range variantDetails {
		formattedVariant := map[string]interface{}{
			"Id":              variant.Id,
			"ProductId":       variant.ProductId,
			"ProductName":     variant.ProductName,
			"BrandName":       variant.BrandName,
			"IsReturnable":    variant.IsReturnable,
			"IsCodAvailable":  variant.IsCodAvailable,
			"CategoryName":    variant.CategoryName,
			"Descriptions":    variant.Descriptions,
			"Images":          variant.Images,
			"OfferName":       variant.OfferName,
			"OfferDetails":    variant.OfferDetails,
			"OfferStartDate":  variant.OfferStartDate,
			"OfferEndDate":    variant.OfferEndDate,
			"OfferPercentage": variant.OfferPercentage,
			"OfferAmount":     variant.OfferAmount,
			"Size":            variant.Size,
			"Colour":          variant.Colour,
			"Ram":             variant.Ram,
			"Storage":         variant.Storage,
			"StockQuantity":   variant.StockQuantity,
			"RegularPrice":    fmt.Sprintf("%.2f", variant.RegularPrice),
			"SalePrice":       fmt.Sprintf("%.2f", variant.SalePrice),
			"SKU":             variant.SKU,
			"ProductSummary":  variant.ProductSummary,
			"Specification":   variant.Specification,
			"IsDeleted":       variant.IsDeleted,
		}

		formattedVariantDetails = append(formattedVariantDetails, formattedVariant)
	}

	c.HTML(http.StatusSeeOther, "productAllVariantsDetails.html", gin.H{
		"Product": formattedVariantDetails,
	})
}

func ShowSingleProductVariantDetail(c *gin.Context) {
	variantID, idErr := strconv.Atoi(c.Query("variant_id"))
	if idErr != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid Input", "")
		return
	}
	variantDetails, err := services.ShowSingleProductVariantDetail(uint(variantID))
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product details", "Database Error", "")
		return
	}
	c.HTML(http.StatusSeeOther, "productVariantDetails.html", gin.H{
		"Variant":       variantDetails,
		"Regular_Price": fmt.Sprintf("%.2f", variantDetails.RegularPrice),
		"Sale_Price":    fmt.Sprintf("%.2f", variantDetails.SalePrice),
	})
}

func AddProductSpecification(c *gin.Context) {
	variantID, err := strconv.Atoi(c.PostForm("variant_id"))
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product variant ID", "Invalid Input", "")
		return
	}
	headings := c.PostFormArray("key[]")
	specification := c.PostFormArray("value[]")
	if len(headings) != len(specification) {
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch in headings and specification", "Validation Error", "")
		return
	}
	for i := 0; i < len(headings); i++ {
		specification := models.ProductSpecification{
			ProductVariantID:   uint(variantID),
			SpecificationKey:   headings[i],
			SpecificationValue: specification[i],
		}
		if err := config.DB.Create(&specification).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save specification", "Database Error", "")
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Specification added successfully",
		"code":    http.StatusOK,
	})
}

func DeleteProductVariant(c *gin.Context) {
	var variant models.ProductVariantDetails

	variantID := c.Param("id")

	if err := config.DB.Unscoped().First(&variant, "ID = ?", variantID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Not Found", "")
		return
	}
	var mainProduct models.ProductDetail
	if err := config.DB.Unscoped().First(&mainProduct, "ID = ?", variant.ProductID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Not Found", "")
		return
	}

	if mainProduct.IsDeleted {
		helper.RespondWithError(c, http.StatusBadRequest, "Cannot recover because the main product is deleted.", "Operation Failed", "")
		return
	}
	if variant.IsDeleted {
		variant.IsDeleted = false
		variant.DeletedAt = gorm.DeletedAt{}

	} else {
		variant.IsDeleted = true
		variant.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	}

	if err := config.DB.Save(&variant).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete or recover product", "Database Error", "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Success",
		"code":    http.StatusOK,
	})
}

func DeleteVariantImage(c *gin.Context) {
	imageID := c.Param("id")

	var variantImage models.ProductVariantsImage

	if err := config.DB.First(&variantImage, "ID = ? AND is_deleted = ?", imageID, false).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Image not found", "Not Found", "")
		return
	}

	publicID, err := helper.ExtractCloudinaryPublicID(variantImage.ProductVariantsImages)
	if err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Processing Error", "")
		return
	}

	cld := config.InitCloudinary()
	if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Upload Error", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&variantImage).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from database", "Database Error", "")
		return
	}

	redirectURL := "/admin/products/variant/detail?variant_id=" + strconv.Itoa(int(variantImage.ProductVariantID))
	c.Redirect(http.StatusFound, redirectURL)
}

func ShowEditProductVariant(c *gin.Context) {
	variantID := c.Param("id")

	var productVariant models.ProductVariantDetails
	if err := config.DB.First(&productVariant, "id = ?", variantID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product variant not found", "Not Found", "")
		return
	}

	var variantImage models.ProductVariantsImage
	if err := config.DB.First(&variantImage, "product_variant_id = ? AND is_deleted = ?", variantID, false).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Image not found", "Not Found", "")
		return
	}

	c.HTML(http.StatusOK, "updateProductVariants.html", gin.H{
		"Details": productVariant,
		"Image":   variantImage,
	})
}

type updateProductVariants struct {
	ProductName    string  `json:"productname"`
	ProductSummary string  `json:"productsummary"`
	Size           string  `json:"size"`
	Colour         string  `json:"colour"`
	Ram            string  `json:"ram"`
	Storage        string  `json:"storage"`
	StockQuantity  int     `json:"stockquantity"`
	RegularPrice   float64 `json:"regularprice"`
	SalePrice      float64 `json:"saleprice"`
	SKU            string  `json:"sku"`
}

func EditProductVariant(c *gin.Context) {
	variantID := c.Param("id")

	var existingVariant models.ProductVariantDetails
	if err := config.DB.First(&existingVariant, variantID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Product variant not found", "Not Found", "")
		return
	}

	var updateData updateProductVariants
	if err := c.ShouldBindJSON(&updateData); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Validation Error", "")
		return
	}

	if err := config.DB.Model(&existingVariant).Updates(map[string]interface{}{
		"product_name":    updateData.ProductName,
		"product_summary": updateData.ProductSummary,
		"size":            updateData.Size,
		"colour":          updateData.Colour,
		"ram":             updateData.Ram,
		"storage":         updateData.Storage,
		"stock_quantity":  updateData.StockQuantity,
		"regular_price":   updateData.RegularPrice,
		"sale_price":      updateData.SalePrice,
		"sku":             updateData.SKU,
	}).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save data", "Database Error", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Product updated successfully",
		"code":    http.StatusOK,
	})
}
func DeleteSpecification(c *gin.Context) {
	specificationID := c.Param("id")
	var specification models.ProductSpecification
	if err := config.DB.First(&specification, specificationID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Specification not found", "Specification not found", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&specification).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete specification", "Delete Specification Failed", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Specification deleted successfully",
		"code":    200,
	})
}

func UpdateProductSpecification(c *gin.Context) {
	type UpdateSpecification struct {
		SpecificationIDs []string `json:"specification_id"`
		SpecificationKey []string `json:"specification_key"`
		Specification    []string `json:"specification"`
	}

	var updateData UpdateSpecification
	if err := c.ShouldBindJSON(&updateData); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Update Specification Failed", "")
		return
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Update Specification Failed", "")
		return
	}

	if len(updateData.SpecificationIDs) != len(updateData.SpecificationKey) ||
		len(updateData.SpecificationKey) != len(updateData.Specification) {
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch in specification IDs, headings, and specifications", "Update Specification Failed", "")
		return
	}

	for i := 0; i < len(updateData.SpecificationIDs); i++ {
		descID, err := strconv.Atoi(updateData.SpecificationIDs[i])
		if err != nil {
			helper.RespondWithError(c, http.StatusBadRequest, "Invalid specification ID", "Update Specification Failed", "")
			return
		}
		result := config.DB.Model(&models.ProductSpecification{}).
			Where("id = ? AND Product_variant_id = ?", descID, productID).
			Updates(map[string]interface{}{
				"SpecificationKey":   updateData.SpecificationKey[i],
				"SpecificationValue": updateData.Specification[i],
			})

		if result.Error != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update specification", "Update Specification Failed", "")
			return
		}

		if result.RowsAffected == 0 {
			helper.RespondWithError(c, http.StatusNotFound, "Specification not found", "Update Specification Failed", "")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Specification updated successfully",
		"code":    200,
	})
}

func ReplaceVariantProductImage(c *gin.Context) {
	imageID, err := strconv.Atoi(c.PostForm("image_id"))
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Replace Image Failed", "")
		return
	}

	tx := config.DB.Begin()
	var variantImage models.ProductVariantsImage
	if err := tx.First(&variantImage, imageID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product image not found", "Replace Image Failed", "")
		return
	}

	oldImage := variantImage.ProductVariantsImages

	form, err := c.FormFile("product_image")
	if err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "No file uploaded", "Replace Image Failed", err.Error())
		return
	}

	cld := config.InitCloudinary()
	file, _ := form.Open()
	url, uploadErr := utils.UploadImageToCloudinary(file, form, cld, "ProductVariants","")
	if uploadErr != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Replace Image Failed", "")
		return
	}

	if err := tx.Model(&variantImage).Update("product_variants_images", url).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update image", "Replace Image Failed", "")
		return
	}

	publicID, err := helper.ExtractCloudinaryPublicID(oldImage)
	if err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Replace Image Failed", "")
		return
	}

	if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Replace Image Failed", "")
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":   "Success",
		"filename": url,
		"code":     http.StatusOK,
	})
}
