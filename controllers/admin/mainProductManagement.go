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
)

type ProductVariantResponse struct {
	ID             uint     `json:"id"`
	ProductName    string   `json:"productname"`
	CategoryName   string   `json:"category_name"`
	RegularPrice   float64  `json:"regular_price"`
	SalePrice      float64  `json:"sale_price"`
	ProductSummary string   `json:"product_summary"`
	IsDeleted      bool     `json:"isdeleted"`
	Images         []string `json:"images"`
}

func ShowProductsAdmin(c *gin.Context) {
	var variants []models.ProductVariantDetails

	result := config.DB.Unscoped().
		Preload("VariantsImages").
		Preload("Category").
		Preload("Product").
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
			IsDeleted:      variant.IsDeleted,
		})
	}
	var formattedResponceDetails []map[string]interface{}
	for _, variant := range response {
		formattedVariant := map[string]interface{}{
			"ID":             variant.ID,
			"ProductName":    variant.ProductName,
			"CategoryName":   variant.CategoryName,
			"RegularPrice":   variant.RegularPrice,
			"SalePrice":      fmt.Sprintf("%.2f", variant.SalePrice),
			"ProductSummary": variant.ProductSummary,
			"Images":         variant.Images,
			"Status":         variant.IsDeleted,
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

func DeleteMainProductImage(c *gin.Context) {
	imageID := c.Param("id")
	var productImage models.ProductImage
	if err := config.DB.First(&productImage, imageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Image not found",
			"code":   http.StatusNotFound,
		})
		return
	}

	publicID, err := helper.ExtractCloudinaryPublicID(productImage.ProductImages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "InternalServerError",
			"error":  "Failed to extract Cloudinary public ID",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	cld := config.InitCloudinary()
	if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "InternalServerError",
			"error":  "Failed to delete image from Cloudinary",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	if err := config.DB.Unscoped().Delete(&productImage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "InternalServerError",
			"error":  "Failed to delete image from database",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(int(productImage.ProductID))
	c.Redirect(http.StatusFound, redirectURL)
}

func ShowEditMainProduct(c *gin.Context) {
	productID := c.Param("id")
	var mainProduct models.ProductDetail
	if err := config.DB.First(&mainProduct, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Product variant not found",
			"code":   http.StatusNotFound,
		})
		return
	}
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

	c.HTML(http.StatusOK, "mainProductUpdate.html", gin.H{
		"Details":    mainProduct,
		"categories": categories,
	})
}

type updateProduct struct {
	ProductName    string `json:"productname"`
	Category       string `json:"category"`
	BrandName      string `json:"brandname"`
	IsCodAvailable bool   `json:"iscodavailable"`
	IsReturnable   bool   `json:"isreturnable"`
}

func EditMainProduct(c *gin.Context) {
	productID := c.Param("id")

	var existingProduct models.ProductDetail
	if err := config.DB.First(&existingProduct, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status": "Not Found",
			"error":  "Product not found",
			"code":   http.StatusNotFound,
		})
		return
	}

	var updateData updateProduct
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid input data",
			"code":   http.StatusBadRequest,
		})
		return
	}

	if err := config.DB.Model(&existingProduct).Updates(updateProduct{
		ProductName:    updateData.ProductName,
		Category:       updateData.Category,
		BrandName:      updateData.BrandName,
		IsCodAvailable: updateData.IsCodAvailable,
		IsReturnable:   updateData.IsReturnable,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Internal Server Error",
			"error":  "Failed to save data",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Product updated successfully",
		"code":    http.StatusOK,
	})
}

func DeleteMainProduct(c *gin.Context) {

	productID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   http.StatusBadRequest,
		})
		return
	}

	tx := config.DB.Begin()

	var product models.ProductDetail
	if err := tx.Unscoped().First(&product, productID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid input data",
			"code":   http.StatusBadRequest,
		})
		return
	}

	newDeleteStatus := !product.IsDeleted
	updateData := map[string]interface{}{
		"is_deleted": newDeleteStatus,
	}
	if newDeleteStatus {
		updateData["deleted_at"] = time.Now()
	} else {
		updateData["deleted_at"] = nil
	}

	if err := tx.Unscoped().Model(&models.ProductDetail{}).Where("id = ?", productID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Error",
			"error":  "Failed to update product",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	if err := tx.Unscoped().Model(&models.ProductVariantDetails{}).Where("product_id = ?", productID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Error",
			"error":  "Failed to update product variants",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "Error",
			"error":  "Transaction commit failed",
			"code":   http.StatusInternalServerError,
		})
		return
	}

	/* c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Product status updated successfully",
		"code":    http.StatusOK,
	})  */
	redirectURL := "/admin/products/main/details?product_id=" + strconv.FormatUint(productID, 10)
	c.Redirect(http.StatusFound, redirectURL)
}

func DeleteDescription(c *gin.Context) {
	descriptionID := c.Param("id")
	var description models.ProductDescription
	if err := config.DB.First(&description, descriptionID).Error; err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Description not found")
		return
	}

	if err := config.DB.Unscoped().Delete(&description).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete description")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Descriptions deleted successfully",
		"code":    200,
	})
}

func UpdateProductDescription(c *gin.Context) {
	type UpdateDescription struct {
		DescriptionIDs []string `json:"description_id"`
		Headings       []string `json:"heading"`
		Descriptions   []string `json:"description"`
	}

	var updateData UpdateDescription
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid request payload",
			"code":   400,
		})
		return
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Invalid product ID",
			"code":   400,
		})
		return
	}

	if len(updateData.DescriptionIDs) != len(updateData.Headings) ||
		len(updateData.Headings) != len(updateData.Descriptions) {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "Bad Request",
			"error":  "Mismatch in description IDs, headings, and descriptions",
			"code":   400,
		})
		return
	}

	for i := 0; i < len(updateData.DescriptionIDs); i++ {
		descID, err := strconv.Atoi(updateData.DescriptionIDs[i])
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "Bad Request",
				"error":  "Invalid description ID",
				"code":   400,
			})
			return
		result := config.DB.Model(&models.ProductDescription{}).
			Where("id = ? AND product_id = ?", descID, productID).
			Updates(map[string]interface{}{
				"heading":     updateData.Headings[i],
				"description": updateData.Descriptions[i],
			})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal Server Error",
				"error":  "Failed to update description",
				"code":   500,
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"status": "Not Found",
				"error":  "Description not found",
				"code":   404,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Descriptions updated successfully",
		"code":    200,
	})
}
