package helper

import (
	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
)

type productsResponse struct {
	ID              uint    `json:"id"`
	ProductName     string  `json:"product_name"`
	ProductSummary  string  `json:"product_summary"`
	SalePrice       float64 `json:"sale_price"`
	RegularPrice    float64 `json:"regular_price"`
	OfferPercentage int     `json:"offer_percentage"`
	Images          string  `json:"images"`
	IsInCart        bool    `json:"is_in_cart"`
	IsInWishlist    bool    `json:"is_in_wishlist"`
}

func RelatedProducts(categoryID uint) ([]productsResponse, error) {
	var products []models.ProductVariantDetails
	if err := config.DB.Preload("VariantsImages", "is_deleted = ?", false).
		Where("category_id = ? AND is_deleted = ? AND stock_quantity>0", categoryID, false).
		Limit(20).
		Find(&products).Error; err != nil {
		return nil, err
	}

	var response []productsResponse
	for _, product := range products {
		discountAmount, TotalPercentage, _ := DiscountCalculation(product.ID, product.CategoryID, product.RegularPrice, product.SalePrice)
		response = append(response, productsResponse{
			ID:              product.ID,
			ProductName:     product.ProductName,
			ProductSummary:  product.ProductSummary,
			SalePrice:       product.SalePrice - discountAmount,
			RegularPrice:    product.RegularPrice,
			OfferPercentage: int(TotalPercentage),
			Images:          product.VariantsImages[0].ProductVariantsImages,
		})
	}
	return response, nil
}

func CheckCartAndWishlist(products []productsResponse, userID uint) []productsResponse {
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

	for i, product := range products {
		if cartMap[product.ID] {
			products[i].IsInCart = true
		}
		if wishlistMap[product.ID] {
			products[i].IsInWishlist = true
		}
	}

	return products
}
