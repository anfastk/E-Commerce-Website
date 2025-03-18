package config

import (
	"log"

	models "github.com/anfastk/E-Commerce-Website/models"
)

func SyncDatabase() {
	err := DB.AutoMigrate(
		&models.AdminModel{}, &models.UserAuth{}, &models.Categories{}, &models.ProductDetail{}, &models.ProductImage{},
		&models.ProductOffer{}, models.ProductDescription{}, &models.ProductVariantsImage{}, models.ProductVariantDetails{}, &models.ProductSpecification{},
		&models.ReservedStock{}, &models.ReservedCoupon{}, &models.Otp{}, &models.UserProfile{}, &models.UserAddress{}, &models.Cart{}, &models.CartItem{},
		&models.Coupon{}, &models.OfferByCategory{}, &models.Order{}, &models.OrderItem{}, &models.Rating{},
		&models.Review{}, &models.ShippingAddress{}, &models.Wallet{}, &models.WalletGiftCard{}, &models.Wishlist{},
		&models.WishlistItem{}, &models.PaymentDetail{}, &models.WalletTransaction{}, &models.ReferralAccount{}, &models.ReferalHistory{}, &models.ReturnRequest{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate models: %v", err)
	}
	log.Println("Models migrated")
	DB.Exec("CREATE INDEX idx_order_items_created_at_product_variant_id ON order_items (created_at, product_variant_id)")
	DownloadLogo()
}
