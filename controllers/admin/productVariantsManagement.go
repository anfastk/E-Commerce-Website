package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/services"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ShowProductVariant(c *gin.Context) {
	logger.Log.Info("Requested to show product variant page")

	productID := c.Param("id")
	var images models.ProductImage
	if err := config.DB.Where("product_id = ? AND is_deleted = ?", productID, false).Find(&images).Error; err != nil {
		logger.Log.Error("Failed to fetch product images", zap.String("productID", productID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Image not found", "Image not found", "")
		return
	}

	logger.Log.Info("Product variant page loaded successfully", zap.String("productID", productID))
	c.HTML(http.StatusSeeOther, "addProductVariants.html", gin.H{
		"Images": images,
	})
}

func AddProductVariants(c *gin.Context) {
	logger.Log.Info("Requested to add product variants")

	productIDStr := c.PostForm("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		logger.Log.Error("Invalid product ID", zap.String("productID", productIDStr), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	var mainProduct models.ProductDetail
	if err := config.DB.First(&mainProduct, productID).Error; err != nil {
		logger.Log.Error("Product not found", zap.Int("productID", productID), zap.Error(err))
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
		logger.Log.Error("Mismatched form data lengths", zap.Int("formLength", formLength))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatched form data lengths", "Form Data Error", "")
		return
	}

	tx := config.DB.Begin()
	cld := config.InitCloudinary()

	form, err := c.MultipartForm()
	if err != nil {
		logger.Log.Error("Failed to parse multipart form", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Failed to parse form", "Form Error", "")
		return
	}

	for i := 0; i < formLength; i++ {
		regularPrice, err := strconv.ParseFloat(regularPrices[i], 64)
		salePrice, err2 := strconv.ParseFloat(salePrices[i], 64)
		stockQuantity, err3 := strconv.Atoi(stockQuantities[i])
		if err != nil || err2 != nil || err3 != nil {
			logger.Log.Error("Invalid price or quantity values",
				zap.String("regularPrice", regularPrices[i]),
				zap.String("salePrice", salePrices[i]),
				zap.String("stockQuantity", stockQuantities[i]),
				zap.Error(err))
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
			logger.Log.Error("Failed to save product variant", zap.Int("index", i), zap.Error(err))
			tx.Rollback()
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product variant", "Database Error", "")
			return
		}

		files := form.File[fmt.Sprintf("product_images[%d][]", i)]
		logger.Log.Info("Processing variant", zap.Int("index", i), zap.Int("fileCount", len(files)))

		if len(files) == 0 {
			logger.Log.Warn("No images uploaded for variant", zap.Int("index", i))
		}

		for j, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				logger.Log.Error("Failed to open file", zap.Int("variantIndex", i), zap.Int("fileIndex", j), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to process image", "File Error", "")
				return
			}
			defer file.Close()

			url, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "ProductVariants", "")
			if err != nil {
				logger.Log.Error("Failed to upload product image", zap.Int("variantIndex", i), zap.Int("imageIndex", j), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Upload Error", "")
				return
			}

			variantImage := models.ProductVariantsImage{
				ProductVariantsImages: url,
				ProductVariantID:      productVariant.ID,
			}
			if err := tx.Create(&variantImage).Error; err != nil {
				logger.Log.Error("Failed to save product image", zap.Uint("variantID", productVariant.ID), zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product image", "Database Error", "")
				return
			}
		}
	}

	tx.Commit()
	redirectURL := "/admin/products/variant/details?product_id=" + strconv.Itoa(int(productID))
	logger.Log.Info("Product variants added successfully", zap.Int("productID", productID), zap.Int("variantCount", formLength))
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"code":     200,
		"redirect": redirectURL,
		"message":  "Product variants added successfully",
	})
}

func ShowSingleProductVariantDetail(c *gin.Context) {
	logger.Log.Info("Requested to show single product variant detail")

	variantIDStr := c.Query("variant_id")
	variantID, idErr := strconv.Atoi(variantIDStr)
	if idErr != nil {
		logger.Log.Error("Invalid variant ID", zap.String("variantID", variantIDStr), zap.Error(idErr))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid Input", "")
		return
	}

	variantDetails, err := services.ShowSingleProductVariantDetail(uint(variantID))
	if err != nil {
		logger.Log.Error("Failed to fetch variant details", zap.Int("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product details", "Database Error", "")
		return
	}

	logger.Log.Info("Single product variant detail fetched successfully", zap.Int("variantID", variantID))
	c.HTML(http.StatusSeeOther, "productVariantDetails.html", gin.H{
		"Variant":       variantDetails,
		"Regular_Price": fmt.Sprintf("%.2f", variantDetails.RegularPrice),
		"Sale_Price":    fmt.Sprintf("%.2f", variantDetails.SalePrice),
	})
}

func AddProductSpecification(c *gin.Context) {
	logger.Log.Info("Requested to add product specification")

	variantIDStr := c.PostForm("variant_id")
	variantID, err := strconv.Atoi(variantIDStr)
	if err != nil {
		logger.Log.Error("Invalid variant ID", zap.String("variantID", variantIDStr), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product variant ID", "Invalid Input", "")
		return
	}

	headings := c.PostFormArray("key[]")
	specification := c.PostFormArray("value[]")
	if len(headings) != len(specification) {
		logger.Log.Error("Mismatch in headings and specification", zap.Int("headings", len(headings)), zap.Int("specifications", len(specification)))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch in headings and specification", "Validation Error", "")
		return
	}

	for i := 0; i < len(headings); i++ {
		spec := models.ProductSpecification{
			ProductVariantID:   uint(variantID),
			SpecificationKey:   headings[i],
			SpecificationValue: specification[i],
		}
		if err := config.DB.Create(&spec).Error; err != nil {
			logger.Log.Error("Failed to save specification", zap.Int("variantID", variantID), zap.Int("index", i), zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save specification", "Database Error", "")
			return
		}
	}

	logger.Log.Info("Product specification added successfully", zap.Int("variantID", variantID), zap.Int("count", len(headings)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Specification added successfully",
		"code":    http.StatusOK,
	})
}

func DeleteProductVariant(c *gin.Context) {
	logger.Log.Info("Requested to delete product variant")

	variantID := c.Param("id")
	var variant models.ProductVariantDetails

	if err := config.DB.Unscoped().First(&variant, "ID = ?", variantID).Error; err != nil {
		logger.Log.Error("Variant not found", zap.String("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Not Found", "")
		return
	}

	var mainProduct models.ProductDetail
	if err := config.DB.Unscoped().First(&mainProduct, "ID = ?", variant.ProductID).Error; err != nil {
		logger.Log.Error("Main product not found", zap.Uint("productID", variant.ProductID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Not Found", "")
		return
	}

	if mainProduct.IsDeleted {
		logger.Log.Error("Cannot recover - main product is deleted", zap.Uint("productID", variant.ProductID))
		helper.RespondWithError(c, http.StatusBadRequest, "Cannot recover because the main product is deleted.", "Operation Failed", "")
		return
	}

	newDeleteStatus := !variant.IsDeleted
	if newDeleteStatus {
		variant.IsDeleted = true
		variant.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	} else {
		variant.IsDeleted = false
		variant.DeletedAt = gorm.DeletedAt{}
	}

	if err := config.DB.Save(&variant).Error; err != nil {
		logger.Log.Error("Failed to update variant delete status", zap.String("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete or recover product", "Database Error", "")
		return
	}

	logger.Log.Info("Product variant delete status updated",
		zap.String("variantID", variantID),
		zap.Bool("isDeleted", newDeleteStatus))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Success",
		"code":    http.StatusOK,
	})
}

func DeleteVariantImage(c *gin.Context) {
	logger.Log.Info("Requested to delete variant image")

	imageID := c.Param("id")
	var variantImage models.ProductVariantsImage

	if err := config.DB.First(&variantImage, "ID = ? AND is_deleted = ?", imageID, false).Error; err != nil {
		logger.Log.Error("Image not found", zap.String("imageID", imageID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Image not found", "Not Found", "")
		return
	}

	publicID, err := helper.ExtractCloudinaryPublicID(variantImage.ProductVariantsImages)
	if err != nil {
		logger.Log.Error("Failed to extract Cloudinary public ID", zap.String("imageID", imageID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Processing Error", "")
		return
	}

	cld := config.InitCloudinary()
	if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
		logger.Log.Error("Failed to delete image from Cloudinary", zap.String("publicID", publicID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Upload Error", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&variantImage).Error; err != nil {
		logger.Log.Error("Failed to delete image from database", zap.String("imageID", imageID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from database", "Database Error", "")
		return
	}

	redirectURL := "/admin/products/variant/detail?variant_id=" + strconv.Itoa(int(variantImage.ProductVariantID))
	logger.Log.Info("Variant image deleted successfully", zap.String("imageID", imageID))
	c.Redirect(http.StatusFound, redirectURL)
}

func ShowEditProductVariant(c *gin.Context) {
	logger.Log.Info("Requested to show edit product variant")

	variantID := c.Param("id")
	var productVariant models.ProductVariantDetails
	if err := config.DB.First(&productVariant, "id = ?", variantID).Error; err != nil {
		logger.Log.Error("Product variant not found", zap.String("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product variant not found", "Not Found", "")
		return
	}

	var variantImage models.ProductVariantsImage
	if err := config.DB.First(&variantImage, "product_variant_id = ? AND is_deleted = ?", variantID, false).Error; err != nil {
		logger.Log.Error("Variant image not found", zap.String("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Image not found", "Not Found", "")
		return
	}

	logger.Log.Info("Edit product variant page loaded successfully", zap.String("variantID", variantID))
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
	logger.Log.Info("Requested to edit product variant")

	variantID := c.Param("id")
	var existingVariant models.ProductVariantDetails
	if err := config.DB.First(&existingVariant, variantID).Error; err != nil {
		logger.Log.Error("Product variant not found", zap.String("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product variant not found", "Not Found", "")
		return
	}

	var updateData updateProductVariants
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logger.Log.Error("Invalid input data", zap.Error(err))
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
		logger.Log.Error("Failed to save variant updates", zap.String("variantID", variantID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save data", "Database Error", "")
		return
	}

	logger.Log.Info("Product variant updated successfully", zap.String("variantID", variantID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Product updated successfully",
		"code":    http.StatusOK,
	})
}

func DeleteSpecification(c *gin.Context) {
	logger.Log.Info("Requested to delete specification")

	specificationID := c.Param("id")
	var specification models.ProductSpecification
	if err := config.DB.First(&specification, specificationID).Error; err != nil {
		logger.Log.Error("Specification not found", zap.String("specificationID", specificationID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Specification not found", "Specification not found", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&specification).Error; err != nil {
		logger.Log.Error("Failed to delete specification", zap.String("specificationID", specificationID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete specification", "Delete Specification Failed", "")
		return
	}

	logger.Log.Info("Specification deleted successfully", zap.String("specificationID", specificationID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Specification deleted successfully",
		"code":    200,
	})
}

func UpdateProductSpecification(c *gin.Context) {
	logger.Log.Info("Requested to update product specification")

	type UpdateSpecification struct {
		SpecificationIDs []string `json:"specification_id"`
		SpecificationKey []string `json:"specification_key"`
		Specification    []string `json:"specification"`
	}

	var updateData UpdateSpecification
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logger.Log.Error("Invalid request payload", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request payload", "Update Specification Failed", "")
		return
	}

	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		logger.Log.Error("Invalid product ID", zap.String("productID", productIDStr), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Update Specification Failed", "")
		return
	}

	if len(updateData.SpecificationIDs) != len(updateData.SpecificationKey) ||
		len(updateData.SpecificationKey) != len(updateData.Specification) {
		logger.Log.Error("Mismatch in specification arrays",
			zap.Int("ids", len(updateData.SpecificationIDs)),
			zap.Int("keys", len(updateData.SpecificationKey)),
			zap.Int("values", len(updateData.Specification)))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch in specification IDs, headings, and specifications", "Update Specification Failed", "")
		return
	}

	for i := 0; i < len(updateData.SpecificationIDs); i++ {
		descID, err := strconv.Atoi(updateData.SpecificationIDs[i])
		if err != nil {
			logger.Log.Error("Invalid specification ID", zap.String("specID", updateData.SpecificationIDs[i]), zap.Error(err))
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
			logger.Log.Error("Failed to update specification", zap.Int("specID", descID), zap.Error(result.Error))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update specification", "Update Specification Failed", "")
			return
		}

		if result.RowsAffected == 0 {
			logger.Log.Error("Specification not found", zap.Int("specID", descID))
			helper.RespondWithError(c, http.StatusNotFound, "Specification not found", "Update Specification Failed", "")
			return
		}
	}

	logger.Log.Info("Product specification updated successfully", zap.Int("productID", productID), zap.Int("count", len(updateData.SpecificationIDs)))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Specification updated successfully",
		"code":    200,
	})
}

func ReplaceVariantProductImage(c *gin.Context) {
	logger.Log.Info("Requested to replace variant product image")

	imageIDStr := c.PostForm("image_id")
	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		logger.Log.Error("Invalid image ID", zap.String("imageID", imageIDStr), zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Replace Image Failed", "")
		return
	}

	tx := config.DB.Begin()
	var variantImage models.ProductVariantsImage
	if err := tx.First(&variantImage, imageID).Error; err != nil {
		logger.Log.Error("Product image not found", zap.Int("imageID", imageID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product image not found", "Replace Image Failed", "")
		return
	}

	oldImage := variantImage.ProductVariantsImages
	form, err := c.FormFile("product_image")
	if err != nil {
		logger.Log.Error("No file uploaded", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "No file uploaded", "Replace Image Failed", err.Error())
		return
	}

	cld := config.InitCloudinary()
	file, _ := form.Open()
	url, uploadErr := utils.UploadImageToCloudinary(file, form, cld, "ProductVariants", "")
	if uploadErr != nil {
		logger.Log.Error("Failed to upload new image", zap.Error(uploadErr))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Replace Image Failed", "")
		return
	}

	if err := tx.Model(&variantImage).Update("product_variants_images", url).Error; err != nil {
		logger.Log.Error("Failed to update image in database", zap.Int("imageID", imageID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update image", "Replace Image Failed", "")
		return
	}

	publicID, err := helper.ExtractCloudinaryPublicID(oldImage)
	if err != nil {
		logger.Log.Error("Failed to extract Cloudinary public ID", zap.String("oldImage", oldImage), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Replace Image Failed", "")
		return
	}

	if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
		logger.Log.Error("Failed to delete old image from Cloudinary", zap.String("publicID", publicID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Replace Image Failed", "")
		return
	} 

	tx.Commit()
	logger.Log.Info("Variant product image replaced successfully", zap.Int("imageID", imageID))
	c.JSON(http.StatusOK, gin.H{
		"status":   "Success",
		"filename": url,
		"code":     http.StatusOK,
	})
}
