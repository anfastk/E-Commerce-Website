package controllers

import (
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
)

type ProductVariantResponse struct {
    ID             uint     `json:"ID"`
    ProductName    string   `json:"ProductName"`
    CategoryName   string   `json:"CategoryName"`
    RegularPrice   float64  `json:"RegularPrice"`
    SalePrice      float64  `json:"SalePrice"`
    ProductSummary string   `json:"ProductSummary"`
    Images         []string `json:"Images"`
    Status         bool     `json:"Status"`
}

func ShowProductsAdmin(c *gin.Context) {
    logger.Log.Info("Requested to show products for admin")
    
    searchQuery := c.Query("search")
    categoryID := c.Query("category")
    
    var variants []models.ProductVariantDetails
    query := config.DB.Unscoped().
        Preload("VariantsImages").
        Preload("Category").
        Preload("Product")
        
    if searchQuery != "" {
        query = query.Where("product_name ILIKE ?", "%"+searchQuery+"%")
    }
    if categoryID != "" {
        query = query.Where("category_id = ?", categoryID)
    }
    
    result := query.Order("created_at DESC").Find(&variants)
    if result.Error != nil {
        logger.Log.Error("Failed to fetch product variants", zap.Error(result.Error))
        helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product variants", "Failed to fetch product variants", "")
        return
    }

    var response []ProductVariantResponse
    for _, variant := range variants {
        var images []string
        for _, img := range variant.VariantsImages {
            images = append(images, img.ProductVariantsImages)
        }

        categoryName := "Uncategorized"
        if variant.Category.ID != 0 {
            categoryName = variant.Category.Name
        }

        productName := variant.ProductName
        if productName == "" && variant.Product.ID != 0 {
            productName = variant.ProductName
        }

        response = append(response, ProductVariantResponse{
            ID:             variant.ID,
            ProductName:    productName,
            CategoryName:   categoryName,
            RegularPrice:   variant.RegularPrice,
            SalePrice:      variant.SalePrice,
            ProductSummary: variant.ProductSummary,
            Images:         images,
            Status:         variant.IsDeleted,
        })
    } 

    if c.Request.Header.Get("X-Requested-With") != "XMLHttpRequest" {
        c.HTML(http.StatusOK, "productPageAdmin.html", gin.H{
            "status":  true,
            "message": "Product variants fetched successfully",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  true,
        "message": "Product variants fetched successfully",
        "data":    response,
    })
}

func GetCategories(c *gin.Context) {
    var categories []models.Categories
    if err := config.DB.Find(&categories).Error; err != nil {
        logger.Log.Error("Failed to fetch categories", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{
            "status":  false,
            "message": "Failed to fetch categories",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "status":  true,
        "message": "Categories fetched successfully",
        "data":    categories,
    })
}

func ShowAddMainProduct(c *gin.Context) {
	logger.Log.Info("Requested to show add main product page")
	
	var categories []models.Categories
	if err := config.DB.Where("is_deleted = ? AND status = ?", false, "Active").Find(&categories).Error; err != nil {
		logger.Log.Error("Failed to fetch categories", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories", "Failed to fetch categories", "")
		return
	}

	logger.Log.Info("Add main product page loaded successfully")
	c.HTML(http.StatusOK, "addNewMainProductDetails.html", gin.H{
		"categories": categories,
	})
}

func AddMainProductDetails(c *gin.Context) {
	logger.Log.Info("Requested to add main product")
	
	tx := config.DB.Begin()
	categoryID, err := strconv.ParseInt(c.PostForm("category"), 10, 64)
	if err != nil {
		logger.Log.Error("Invalid category ID", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Invalid category ID", "Invalid category ID", "")
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
		logger.Log.Error("Failed to save product", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save product", "Failed to save product", "")
		return
	}

	cld := config.InitCloudinary()
	form, _ := c.MultipartForm()
	if form != nil {
		if productImage, ok := form.File["product_image"]; ok && len(productImage) > 0 {
			fileHeader := productImage[0]
			file, _ := fileHeader.Open()
			defer file.Close()

			url, err := utils.UploadImageToCloudinary(file, fileHeader, cld, "products", "")
			if err != nil {
				logger.Log.Error("Failed to upload product image to Cloudinary", zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Failed to upload product image", "")
				return
			}

			image := models.ProductImage{
				ProductImages: url,
				ProductID:     product.ID,
			}
			if err := tx.Create(&image).Error; err != nil {
				logger.Log.Error("Failed to save product image", zap.Error(err))
				tx.Rollback()
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Failed to upload product image", "")
				return
			}
		}
	}

	tx.Commit()
	logger.Log.Info("Main product added successfully", zap.Uint("productID", product.ID))
	redirectURL := "/admin/products/main/details?product_id=" + strconv.Itoa(int(product.ID))
	c.Redirect(http.StatusFound, redirectURL)
}

func ShowMainProductDetails(c *gin.Context) {
	productIDStr := c.Query("product_id")
	logger.Log.Info("Requested to show main product details", zap.String("productID", productIDStr))
	
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		logger.Log.Error("Invalid product ID", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	productDetails, err := services.ShowMainProductsDetails(uint(productID))
	if err != nil {
		logger.Log.Error("Failed to fetch product details", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch product details", "Failed to fetch product details", "")
		return
	}

	var product []models.ProductVariantDetails
	if err := config.DB.Preload("VariantsImages").Find(&product, "product_id = ?", productID).Error; err != nil {
		logger.Log.Error("Failed to fetch product variants", zap.Error(err))
	}

	logger.Log.Info("Main product details fetched successfully", zap.String("productID", productIDStr))
	c.HTML(http.StatusSeeOther, "mainProductDetails.html", gin.H{
		"Product":  productDetails,
		"Products": product,
	})
}

func AddProductDescription(c *gin.Context) {
	productIDStr := c.PostForm("product_id")
	logger.Log.Info("Requested to add product description", zap.String("productID", productIDStr))
	
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		logger.Log.Error("Invalid product ID", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	headings := c.PostFormArray("heading[]")
	descriptions := c.PostFormArray("description[]")
	if len(headings) != len(descriptions) {
		logger.Log.Error("Mismatch in headings and descriptions", zap.Int("headings", len(headings)), zap.Int("descriptions", len(descriptions)))
		helper.RespondWithError(c, http.StatusBadRequest, "Mismatch in headings and descriptions", "Mismatch in headings and descriptions", "")
		return
	}

	for i := 0; i < len(headings); i++ {
		description := models.ProductDescription{
			ProductID:   uint(productID),
			Heading:     headings[i],
			Description: descriptions[i],
		}
		if err := config.DB.Create(&description).Error; err != nil {
			logger.Log.Error("Failed to save description", zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to save description", "Failed to save description", "")
			return
		}
	}

	logger.Log.Info("Product description added successfully", zap.String("productID", productIDStr))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Product updated successfully",
		"code":    http.StatusOK,
	})
}

func ShowEditMainProduct(c *gin.Context) {
	productID := c.Param("id")
	logger.Log.Info("Requested to show edit main product", zap.String("productID", productID))
	
	var mainProduct models.ProductDetail
	if err := config.DB.First(&mainProduct, productID).Error; err != nil {
		logger.Log.Error("Product not found", zap.String("productID", productID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Product not found", "")
		return
	}

	var productCategoryName models.Categories
	if err := config.DB.Where("is_deleted = ? AND status = ? AND id = ?", false, "Active", mainProduct.CategoryID).Find(&productCategoryName).Error; err != nil {
		logger.Log.Error("Failed to fetch product category", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories", "Failed to fetch categories", "")
		return
	}

	var categories []models.Categories
	if err := config.DB.Where("is_deleted = ? AND status = ?", false, "Active").Find(&categories).Error; err != nil {
		logger.Log.Error("Failed to fetch categories", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories", "Failed to fetch categories", "")
		return
	}

	logger.Log.Info("Edit main product page loaded successfully", zap.String("productID", productID))
	c.HTML(http.StatusOK, "mainProductUpdate.html", gin.H{
		"Details":      mainProduct,
		"CategoryName": productCategoryName,
		"categories":   categories,
	})
}

type updateProduct struct {
	ProductName    string `json:"productname"`
	CategoryID     string `json:"categoryid"`
	BrandName      string `json:"brandname"`
	IsCodAvailable bool   `json:"iscodavailable"`
	IsReturnable   bool   `json:"isreturnable"`
}

func EditMainProduct(c *gin.Context) {
	productID := c.Param("id")
	logger.Log.Info("Requested to edit main product", zap.String("productID", productID))
	
	tx := config.DB.Begin()
	var existingProduct models.ProductDetail
	if err := tx.First(&existingProduct, productID).Error; err != nil {
		logger.Log.Error("Product not found", zap.String("productID", productID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product not found", "Product not found", "")
		return
	}

	var updateData updateProduct
	if err := c.ShouldBindJSON(&updateData); err != nil {
		logger.Log.Error("Invalid request data", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, err.Error(), err.Error(), "")
		return
	}

	categoryID, err := strconv.ParseUint(updateData.CategoryID, 10, 32)
	if err != nil {
		logger.Log.Error("Invalid category ID format", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid category ID format", "Invalid category ID format", "")
		return
	}

	if err := tx.Model(&existingProduct).Updates(models.ProductDetail{
		ProductName:    updateData.ProductName,
		CategoryID:     uint(categoryID),
		BrandName:      updateData.BrandName,
		IsCODAvailable: updateData.IsCodAvailable,
		IsReturnable:   updateData.IsReturnable,
	}).Error; err != nil {
		logger.Log.Error("Failed to update product", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, err.Error(), err.Error(), "")
		return
	}

	var variants []models.ProductVariantDetails
	if err := tx.Unscoped().Find(&variants, "product_id = ?", existingProduct.ID).Error; err != nil {
		logger.Log.Error("Failed to find product variants", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product variants not found", "Product variants not found", "")
		return
	}

	if err := tx.Model(&variants).Update("category_id", existingProduct.CategoryID).Error; err != nil {
		logger.Log.Error("Failed to update product variants", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, err.Error(), err.Error(), "")
		return
	}

	tx.Commit()
	logger.Log.Info("Main product updated successfully", zap.String("productID", productID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "Product updated successfully",
		"code":    http.StatusOK,
	})
}

func DeleteMainProduct(c *gin.Context) {
	productIDStr := c.Param("id")
	logger.Log.Info("Requested to delete main product", zap.String("productID", productIDStr))
	
	productID, err := strconv.ParseUint(productIDStr, 10, 64)
	if err != nil {
		logger.Log.Error("Invalid product ID", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	tx := config.DB.Begin()
	var product models.ProductDetail
	if err := tx.Unscoped().First(&product, productID).Error; err != nil {
		logger.Log.Error("Product not found", zap.Uint64("productID", productID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid input data", "Invalid input data", "")
		return
	}

	var category models.Categories
	if err := config.DB.Unscoped().First(&category, "ID = ?", product.CategoryID).Error; err != nil {
		logger.Log.Error("Category not found", zap.Uint("categoryID", product.CategoryID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Category not found", "Category not found", "")
		return
	}

	if category.IsDeleted {
		logger.Log.Error("Cannot recover - category is deleted", zap.Uint("categoryID", product.CategoryID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Cannot recover because the category is deleted.", "Cannot recover because the category is deleted.", "")
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
		logger.Log.Error("Failed to update product status", zap.Uint64("productID", productID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product", "Failed to update product", "")
		return
	}

	if err := tx.Unscoped().Model(&models.ProductVariantDetails{}).Where("product_id = ?", productID).Updates(updateData).Error; err != nil {
		logger.Log.Error("Failed to update product variants status", zap.Uint64("productID", productID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update product variants", "Failed to update product variants", "")
		return
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Transaction commit failed", zap.Uint64("productID", productID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Transaction commit failed", "Transaction commit failed", "")
		return
	}

	logger.Log.Info("Product status updated successfully", zap.Uint64("productID", productID), zap.Bool("isDeleted", newDeleteStatus))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Product status updated successfully",
		"code":    http.StatusOK,
	})
}

func DeleteDescription(c *gin.Context) {
	descriptionID := c.Param("id")
	logger.Log.Info("Requested to delete description", zap.String("descriptionID", descriptionID))
	
	var description models.ProductDescription
	if err := config.DB.First(&description, descriptionID).Error; err != nil {
		logger.Log.Error("Description not found", zap.String("descriptionID", descriptionID), zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Description not found", "Description not found", "")
		return
	}

	if err := config.DB.Unscoped().Delete(&description).Error; err != nil {
		logger.Log.Error("Failed to delete description", zap.String("descriptionID", descriptionID), zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete description", "Failed to delete description", "")
		return
	}

	logger.Log.Info("Description deleted successfully", zap.String("descriptionID", descriptionID))
	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Descriptions deleted successfully",
		"code":    200,
	})
}

func ReplaceMainProductImage(c *gin.Context) {
	productIDStr := c.PostForm("product_id")
	logger.Log.Info("Requested to replace main product image", zap.String("productID", productIDStr))
	
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		logger.Log.Error("Invalid product ID", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid product ID", "Invalid product ID", "")
		return
	}

	tx := config.DB.Begin()
	var productImage models.ProductImage
	if err := tx.First(&productImage, "product_id = ?", productID).Error; err != nil {
		logger.Log.Error("Product image not found", zap.Int("productID", productID), zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Product image not found", "Product image not found", "")
		return
	}

	oldImage := productImage.ProductImages

	form, err := c.FormFile("product_image")
	if err != nil {
		logger.Log.Error("No file uploaded", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "No file uploaded", "No file uploaded", "")
		return
	}

	cld := config.InitCloudinary()
	file, _ := form.Open()
	url, uploadErr := utils.UploadImageToCloudinary(file, form, cld, "products", "")
	if uploadErr != nil {
		logger.Log.Error("Failed to upload new image to Cloudinary", zap.Error(uploadErr))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to upload product image", "Failed to upload product image", "")
		return
	}

	if err := tx.Model(&productImage).Update("product_images", url).Error; err != nil {
		logger.Log.Error("Failed to update image in database", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update image", "Failed to update image", "")
		return
	}

	publicID, err := helper.ExtractCloudinaryPublicID(oldImage)
	if err != nil {
		logger.Log.Error("Failed to extract Cloudinary public ID", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to extract Cloudinary public ID", "Failed to extract Cloudinary public ID", "")
		return
	}

	if err := utils.DeleteCloudinaryImage(cld, publicID, c); err != nil {
		logger.Log.Error("Failed to delete old image from Cloudinary", zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to delete image from Cloudinary", "Failed to delete image from Cloudinary", "")
		return
	}

	tx.Commit()
	logger.Log.Info("Main product image replaced successfully", zap.Int("productID", productID))
	c.JSON(http.StatusOK, gin.H{
		"status":   "Success",
		"filename": url,
		"code":     http.StatusOK,
	})
}