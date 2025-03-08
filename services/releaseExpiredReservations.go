package services

import (
	"errors"
	"log"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"gorm.io/gorm"
)

func ReleaseExpiredReservations(db *gorm.DB) {
	tx := config.DB.Begin()

	var expiredReservations []models.ReservedStock
	err := tx.Where("is_confirmed = ? AND reserve_till < ?", false, time.Now()).
		Find(&expiredReservations).Error
	if err != nil {
		tx.Rollback()
		log.Println("Error finding expired reservations:", err)
		return
	}

	for _, reservation := range expiredReservations {
		var coupon models.ReservedCoupon
		tx.First(&coupon, reservation.ReservedCouponID)
		tx.Exec("UPDATE coupons SET users_used_count = users_used_count + ? WHERE id = ?", 1, coupon.CouponID)
		tx.Unscoped().Delete(&coupon)
		tx.Exec(
			"UPDATE product_variant_details SET stock_quantity = stock_quantity + ? WHERE id = ?",
			reservation.Quantity, reservation.ProductVariantID,
		)
		tx.Unscoped().Delete(&reservation)
	}

	var failedOrderItems []models.OrderItem

	if err := tx.Where("order_status = ? AND created_at < ?", "Order Not Placed", time.Now().Add(-30*time.Minute)).
		Find(&failedOrderItems).Error; err != nil {
		tx.Rollback()
		log.Println("Error finding expired reservations:", err)
		return
	}

	for _, item := range failedOrderItems {
		if err := tx.Exec(
			"UPDATE product_variant_details SET stock_quantity = stock_quantity + ? WHERE id = ?",
			item.Quantity, item.ProductVariantID,
		).Error; err != nil {
			tx.Rollback()
			log.Println("Error updating stock quantity:", err)
			return
		}

		item.OrderStatus = "Failed"
		if err := tx.Save(&item).Error; err != nil {
			tx.Rollback()
			log.Println("Error updating order status:", err)
			return
		}

		var paymentDetail models.PaymentDetail
		if err := tx.First(&paymentDetail, "order_item_id = ?", item.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Println("Payment detail not found for order item:", item.ID)
				continue
			}
			tx.Rollback()
			log.Println("Error fetching payment detail:", err)
			return
		}

		paymentDetail.PaymentStatus = "Cancelled"
		if err := tx.Save(&paymentDetail).Error; err != nil {
			tx.Rollback()
			log.Println("Error updating payment status:", err)
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Println("Error committing transaction:", err)
		return
	}
	log.Println("Expired reservations released.")
}

func StartReservationCleanupTask(db *gorm.DB) {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			ReleaseExpiredReservations(db)
		}
	}()
}
