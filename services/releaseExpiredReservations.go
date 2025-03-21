package services

import (
	"errors"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func ReleaseExpiredReservations(db *gorm.DB) {
	logger.Log.Info("Starting expired reservations cleanup")

	tx := config.DB.Begin()

	var expiredReservations []models.ReservedStock
	err := tx.Where("is_confirmed = ? AND reserve_till < ?", false, time.Now()).
		Find(&expiredReservations).Error
	if err != nil {
		logger.Log.Error("Error finding expired reservations",
			zap.Error(err))
		tx.Rollback()
		return
	}

	for _, reservation := range expiredReservations {
		logger.Log.Debug("Processing expired reservation",
			zap.Uint("reservationID", reservation.ID),
			zap.Uint("productVariantID", reservation.ProductVariantID))

		var coupon models.ReservedCoupon
		if err := tx.First(&coupon, reservation.ReservedCouponID).Error; err != nil {
			logger.Log.Warn("Reserved coupon not found",
				zap.Uint("reservedCouponID", reservation.ReservedCouponID),
				zap.Error(err))
		} else {
			if err := tx.Exec("UPDATE coupons SET users_used_count = users_used_count + ? WHERE id = ?", 1, coupon.CouponID).Error; err != nil {
				logger.Log.Error("Failed to update coupon users_used_count",
					zap.Uint("couponID", coupon.CouponID),
					zap.Error(err))
				tx.Rollback()
				return
			}
			if err := tx.Unscoped().Delete(&coupon).Error; err != nil {
				logger.Log.Error("Failed to delete reserved coupon",
					zap.Uint("reservedCouponID", coupon.ID),
					zap.Error(err))
				tx.Rollback()
				return
			}
		}

		if err := tx.Exec(
			"UPDATE product_variant_details SET stock_quantity = stock_quantity + ? WHERE id = ?",
			reservation.Quantity, reservation.ProductVariantID,
		).Error; err != nil {
			logger.Log.Error("Failed to update stock quantity for reservation",
				zap.Uint("productVariantID", reservation.ProductVariantID),
				zap.Int("quantity", reservation.Quantity),
				zap.Error(err))
			tx.Rollback()
			return
		}

		if err := tx.Unscoped().Delete(&reservation).Error; err != nil {
			logger.Log.Error("Failed to delete reservation",
				zap.Uint("reservationID", reservation.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}
	}

	var failedOrderItems []models.OrderItem
	if err := tx.Where("order_status = ? AND created_at < ?", "Order Not Placed", time.Now().Add(-30*time.Minute)).
		Find(&failedOrderItems).Error; err != nil {
		logger.Log.Error("Error finding failed order items",
			zap.Error(err))
		tx.Rollback()
		return
	}

	for _, item := range failedOrderItems {
		logger.Log.Debug("Processing failed order item",
			zap.Uint("orderItemID", item.ID),
			zap.Uint("productVariantID", item.ProductVariantID))

		if err := tx.Exec(
			"UPDATE product_variant_details SET stock_quantity = stock_quantity + ? WHERE id = ?",
			item.Quantity, item.ProductVariantID,
		).Error; err != nil {
			logger.Log.Error("Failed to update stock quantity for failed order",
				zap.Uint("productVariantID", item.ProductVariantID),
				zap.Int("quantity", item.Quantity),
				zap.Error(err))
			tx.Rollback()
			return
		}

		item.OrderStatus = "Failed"
		if err := tx.Save(&item).Error; err != nil {
			logger.Log.Error("Failed to update order status to Failed",
				zap.Uint("orderItemID", item.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}

		var paymentDetail models.PaymentDetail
		if err := tx.First(&paymentDetail, "order_item_id = ?", item.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Log.Warn("Payment detail not found for order item",
					zap.Uint("orderItemID", item.ID))
				continue
			}
			logger.Log.Error("Error fetching payment detail",
				zap.Uint("orderItemID", item.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}

		paymentDetail.PaymentStatus = "Cancelled"
		if err := tx.Save(&paymentDetail).Error; err != nil {
			logger.Log.Error("Failed to update payment status to Cancelled",
				zap.Uint("paymentDetailID", paymentDetail.ID),
				zap.Error(err))
			tx.Rollback()
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		logger.Log.Error("Failed to commit transaction",
			zap.Error(err))
		return
	}

	logger.Log.Info("Expired reservations released",
		zap.Int("reservationCount", len(expiredReservations)),
		zap.Int("failedOrderCount", len(failedOrderItems)))
}

func StartReservationCleanupTask(db *gorm.DB) {
	logger.Log.Info("Starting reservation cleanup task")
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			ReleaseExpiredReservations(db)
		}
	}()
}
