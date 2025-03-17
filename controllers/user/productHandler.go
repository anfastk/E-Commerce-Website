package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/middleware"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func UserHome(c *gin.Context) {
	tokenString, err := c.Cookie("jwtTokensUser")
	isLoggedIn := false
	var userID uint

	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.GetJwtKey(), nil
		})

		if err == nil && token.Valid && claims.Role == "User" {
			isLoggedIn = true
			userID = claims.UserId

			var user models.UserAuth
			if err := config.DB.First(&user, userID).Error; err != nil || user.IsBlocked || user.IsDeleted {
				c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
				isLoggedIn = false
			}
		}
	}

	keyboard, _ := helper.RelatedProducts(2)
	laptop, _ := helper.RelatedProducts(3)
	mouse, _ := helper.RelatedProducts(4)

	if isLoggedIn {
		keyboard = helper.CheckCartAndWishlist(keyboard, userID)
		laptop = helper.CheckCartAndWishlist(laptop, userID)
		mouse = helper.CheckCartAndWishlist(mouse, userID)
	}

	c.HTML(http.StatusOK, "userHome.html", gin.H{
		"Keyboard":   keyboard,
		"Laptop":     laptop,
		"Mouse":      mouse,
		"IsLoggedIn": isLoggedIn,
	})
}

type ProductVariantResponse struct {
	ID              uint    `json:"id"`
	ProductName     string  `json:"product_name"`
	RegularPrice    float64 `json:"regular_price"`
	SalePrice       float64 `json:"sale_price"`
	OfferPercentage int     `json:"offer_persentage"`
	Images          string  `json:"images"`
	IsInCart        bool    `json:"is_in_cart"`
	IsInWishlist    bool    `json:"is_in_wishlist"`
	IsInStock       bool    `json:"is_in_stock"`
}

func ShowProducts(c *gin.Context) {
	tokenString, err := c.Cookie("jwtTokensUser")
	var userID uint

	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.GetJwtKey(), nil
		})

		if err == nil && token.Valid && claims.Role == "User" {
			userID = claims.UserId

			var user models.UserAuth
			if err := config.DB.First(&user, userID).Error; err != nil || user.IsBlocked || user.IsDeleted {
				c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
			}
		}
	}

	var cartItems []models.CartItem
	var wishlistItems []models.WishlistItem

	config.DB.Where("cart_id = (SELECT id FROM carts WHERE user_id = ?)", userID).Find(&cartItems)

	config.DB.Where("wishlist_id = (SELECT id FROM wishlists WHERE user_id = ?)", userID).Find(&wishlistItems)

	cartMap := make(map[uint]bool)
	wishlistMap := make(map[uint]bool)

	for _, item := range cartItems {
		cartMap[item.ProductVariantID] = true
	}

	for _, item := range wishlistItems {
		wishlistMap[item.ProductVariantID] = true
	}

	var Brand []string
	var Category []string

	if err := config.DB.Model(&models.ProductDetail{}).Distinct("brand_name").Where("is_deleted =?", false).Pluck("brand_name", &Brand).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch brand name", "Something Went Wrong", "")
		return
	}
	if err := config.DB.Model(&models.Categories{}).Where("is_deleted =?", false).Pluck("name", &Category).Error; err != nil {
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
		discountAmount, TotalPercentage, disErr := helper.DiscountCalculation(variant.ProductID, variant.CategoryID, variant.RegularPrice, variant.SalePrice)
		if disErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Discount Calculation Failed", "Something Went Wrong", "")
			return
		}
		responses := ProductVariantResponse{
			ID:              variant.ID,
			ProductName:     variant.ProductName,
			RegularPrice:    variant.RegularPrice,
			SalePrice:       variant.SalePrice - discountAmount,
			OfferPercentage: int(TotalPercentage),
			Images:          variant.VariantsImages[0].ProductVariantsImages,
			IsInStock:       variant.StockQuantity > 0,
		}
		if cartMap[variant.ID] {
			responses.IsInCart = true
		}
		if wishlistMap[variant.ID] {
			responses.IsInWishlist = true
		}
		response = append(response, responses)
	}

	c.HTML(http.StatusFound, "productpage.html", gin.H{
		"status":   true,
		"message":  "Product variants fetched successfully",
		"data":     response,
		"Brand":    Brand,
		"Category": Category,
	})
}

func FilterProducts(c *gin.Context) {
	tokenString, err := c.Cookie("jwtTokensUser")
	var userID uint

	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.GetJwtKey(), nil
		})

		if err == nil && token.Valid && claims.Role == "User" {
			userID = claims.UserId

			var user models.UserAuth
			if err := config.DB.First(&user, userID).Error; err != nil || user.IsBlocked || user.IsDeleted {
				c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
			}
		}
	}

	var cartItems []models.CartItem
	var wishlistItems []models.WishlistItem

	config.DB.Where("cart_id = (SELECT id FROM carts WHERE user_id = ?)", userID).Find(&cartItems)

	config.DB.Where("wishlist_id = (SELECT id FROM wishlists WHERE user_id = ?)", userID).Find(&wishlistItems)

	cartMap := make(map[uint]bool)
	wishlistMap := make(map[uint]bool)

	for _, item := range cartItems {
		cartMap[item.ProductVariantID] = true
	}

	for _, item := range wishlistItems {
		wishlistMap[item.ProductVariantID] = true
	}

	search := c.Query("search")
	sort := c.Query("sort")
	categories := c.QueryArray("categories")
	priceRanges := c.QueryArray("priceRanges")
	discountsStr := c.QueryArray("discounts")
	brands := c.QueryArray("brands")
	includeOutOfStock, _ := strconv.ParseBool(c.DefaultQuery("includeOutOfStock", "false"))

	var discounts []int
	for _, d := range discountsStr {
		if val, err := strconv.Atoi(d); err == nil {
			discounts = append(discounts, val)
		}
	}

	query := config.DB.Model(&models.ProductVariantDetails{}).
		Preload("VariantsImages", "is_deleted = ?", false).
		Preload("Category").
		Preload("Product").
		Where("product_variant_details.is_deleted = ?", false)

	query = query.Joins("JOIN product_details ON product_variant_details.product_id = product_details.id")

	if search != "" {
		searchTerm := "%" + search + "%"
		query = query.Joins("JOIN categories ON product_variant_details.category_id = categories.id").
			Where("product_variant_details.product_name ILIKE ? OR product_details.brand_name ILIKE ? OR categories.name ILIKE ?",
				searchTerm, searchTerm, searchTerm)
	} else if len(categories) > 0 {
		query = query.Joins("JOIN categories ON product_variant_details.category_id = categories.id").
			Where("categories.name IN ?", categories)
	}

	if len(brands) > 0 {
		query = query.Where("product_details.brand_name IN ?", brands)
	}

	if len(priceRanges) > 0 {
		priceConditions := []string{}
		for _, pr := range priceRanges {
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

	if len(discounts) > 0 {
		query = query.Where("((product_variant_details.regular_price - product_variant_details.sale_price) / product_variant_details.regular_price * 100) >= ?",
			discounts[len(discounts)-1])
	}

	if !includeOutOfStock {
		query = query.Where("product_variant_details.stock_quantity > 0")
	}

	switch sort {
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
		discountAmount, TotalPercentage, disErr := helper.DiscountCalculation(variant.ProductID, variant.CategoryID, variant.RegularPrice, variant.SalePrice)
		if disErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Discount Calculation Failed", "Something Went Wrong", "")
			return
		}
		responses := ProductVariantResponse{
			ID:              variant.ID,
			ProductName:     variant.ProductName,
			RegularPrice:    variant.RegularPrice,
			SalePrice:       variant.SalePrice - discountAmount,
			OfferPercentage: int(TotalPercentage),
			Images:          variant.VariantsImages[0].ProductVariantsImages,
			IsInStock:       variant.StockQuantity > 0,
		}
		if cartMap[variant.ID] {
			responses.IsInCart = true
		}
		if wishlistMap[variant.ID] {
			responses.IsInWishlist = true
		}
		response = append(response, responses)
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Products filtered successfully",
		"data":    response,
	})
}

type ProductDetailResponse struct {
	ID              uint                    `json:"id"`
	ProductName     string                  `json:"product_name"`
	CategoryName    string                  `json:"category_name"`
	CategoryID      uint                    `json:"category_id"`
	RegularPrice    float64                 `json:"regular_price"`
	SalePrice       float64                 `json:"sale_price"`
	Images          []string                `json:"images"`
	Size            string                  `json:"size"`
	Color           string                  `json:"color"`
	Ram             string                  `json:"ram"`
	Storage         string                  `json:"storage"`
	Stock           int                     `json:"stock"`
	OfferPercentage int                     `json:"offer_percentage"`
	Summary         string                  `json:"summary"`
	IsInCart        bool                    `json:"is_in_cart"`
	IsInWishlist    bool                    `json:"is_in_wishlist"`
	Specifications  []SpecificationResponse `json:"specifications"`
	Description     []DescriptionResponse   `json:"description"`
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

	tokenString, err := c.Cookie("jwtTokensUser")
	var userID uint

	if err == nil && tokenString != "" {
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return middleware.GetJwtKey(), nil
		})

		if err == nil && token.Valid && claims.Role == "User" {
			userID = claims.UserId

			var user models.UserAuth
			if err := config.DB.First(&user, userID).Error; err != nil || user.IsBlocked || user.IsDeleted {
				c.SetCookie("jwtTokensUser", "", -1, "/", "", false, true)
			}
		}
	}

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

	var cartItems models.CartItem
	var wishlistItems models.WishlistItem
	IsInCart := true
	IsInWishlist := true

	if err := config.DB.Where("cart_id = (SELECT id FROM carts WHERE user_id = ?) AND product_variant_id = ?", userID, variant.ID).First(&cartItems).Error; err != nil {
		IsInCart = false
	}

	if err := config.DB.Where("wishlist_id = (SELECT id FROM wishlists WHERE user_id = ?) AND product_variant_id = ?", userID, variant.ID).First(&wishlistItems).Error; err != nil {
		IsInWishlist = false
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

	discountAmount, TotalPercentage, disErr := helper.DiscountCalculation(variant.ProductID, variant.CategoryID, variant.RegularPrice, variant.SalePrice)
	if disErr != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Discount Calculation Failed", "Something Went Wrong", "")
		return
	}
	product := ProductDetailResponse{
		ID:              variant.ID,
		ProductName:     variant.ProductName,
		CategoryName:    variant.Category.Name,
		CategoryID:      variant.CategoryID,
		RegularPrice:    variant.RegularPrice,
		SalePrice:       variant.SalePrice - discountAmount,
		Images:          images,
		Size:            variant.Size,
		Color:           variant.Colour,
		Ram:             variant.Ram,
		Storage:         variant.Storage,
		OfferPercentage: int(TotalPercentage),
		Stock:           variant.StockQuantity,
		Summary:         variant.ProductSummary,
		Specifications:  specs,
		Description:     description,
		IsInCart:        IsInCart,
		IsInWishlist:    IsInWishlist,
	}

	type otherVariantDetail struct {
		ID              uint
		Colour          string
		Ram             string
		Rom             string
		RegularPrice    float64
		SalePrice       float64
		OfferPersentage int
		Image           string
		Stock           int
	}

	var otherVariant []models.ProductVariantDetails
	if err := config.DB.Preload("VariantsImages").Where("id !=? AND product_id = ?", variant.ID, variant.ProductID).Find(&otherVariant).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Variants Not Found", "Something Went Wrong", "")
		return
	}

	var otherVariantDetails []otherVariantDetail
	for _, row := range otherVariant {
		discountAmount, TotalPercentage, disErr := helper.DiscountCalculation(row.ProductID, row.CategoryID, row.RegularPrice, row.SalePrice)
		if disErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Discount Calculation Failed", "Something Went Wrong", "")
			return
		}
		itm := otherVariantDetail{
			ID:              row.ID,
			Colour:          row.Colour,
			Ram:             row.Ram,
			Rom:             row.Storage,
			RegularPrice:    row.RegularPrice,
			SalePrice:       row.SalePrice - discountAmount,
			OfferPersentage: int(TotalPercentage),
			Image:           row.VariantsImages[0].ProductVariantsImages,
			Stock:           row.StockQuantity,
		}
		otherVariantDetails = append(otherVariantDetails, itm)
	}

	var relatedProducts []models.ProductVariantDetails
	config.DB.Preload("VariantsImages", "is_deleted = ?", false).
		Where("category_id = ? AND id != ? AND is_deleted = ?", variant.CategoryID, variant.ID, false).
		Limit(20).
		Find(&relatedProducts)

	type RelatedProductsResponce struct {
		ID              uint     `json:"id"`
		ProductName     string   `json:"product_name"`
		ProductSummary  string   `json:"product_summary"`
		OfferPercentage int      `json:"offer_percentage"`
		SalePrice       float64  `json:"sale_price "`
		Images          []string `json:"images"`
	}
	var relatedProductsResponce []RelatedProductsResponce

	for _, product := range relatedProducts {
		var images []string
		for _, image := range product.VariantsImages {
			images = append(images, image.ProductVariantsImages)
		}
		discountAmount, TotalPercentage, disErr := helper.DiscountCalculation(product.ProductID, product.CategoryID, product.RegularPrice, product.SalePrice)
		if disErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Discount Calculation Failed", "Something Went Wrong", "")
			return
		}
		relatedProductsResponce = append(relatedProductsResponce, RelatedProductsResponce{
			ID:              product.ID,
			ProductName:     product.ProductName,
			ProductSummary:  product.ProductSummary,
			OfferPercentage: int(TotalPercentage),
			SalePrice:       product.SalePrice - discountAmount,
			Images:          images,
		})
	}

	c.HTML(http.StatusFound, "productDetails.html", gin.H{
		"product":             product,
		"relatedProducts":     relatedProductsResponce,
		"OtherVariantDetails": otherVariantDetails,
	})
}
