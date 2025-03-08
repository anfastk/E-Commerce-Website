package config

import (
	"log"

	models "github.com/anfastk/E-Commerce-Website/models"
)

func SyncDatabase() {
	err:=DB.AutoMigrate(
		&models.AdminModel{},&models.UserAuth{},&models.Categories{},&models.ProductDetail{},&models.ProductImage{},
		&models.ProductOffer{},models.ProductDescription{},&models.ProductVariantsImage{},models.ProductVariantDetails{},&models.ProductSpecification{},
		&models.ReservedStock{},&models.ReservedCoupon{},&models.Otp{},&models.UserProfile{},&models.UserAddress{},&models.Cart{},&models.CartItem{},
		&models.Coupon{},&models.OfferByCategory{},&models.Order{},&models.OrderItem{},&models.Rating{},
		&models.Review{},&models.Sale{},&models.SalesProductItem{},&models.ShippingAddress{},&models.Wallet{}, 
		&models.WalletGiftCard{},&models.Wishlist{},&models.WishlistItem{},&models.PaymentDetail{},&models.WalletTransaction{},
		&models.ReferralAccount{},&models.ReferalHistory{},&models.ReturnRequest{},
	)
	if err != nil{
		log.Fatalf("Failed to migrate models: %v",err)
	}
	log.Println("Models migrated")
}