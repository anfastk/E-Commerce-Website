package controllers

import (
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func ShowWalletManagement(c *gin.Context) {
	logger.Log.Info("Requested wallet management page")

	type walletResponce struct {
		ID            uint
		TransactionID string
		Date          string
		ProfilePic    string
		Name          string
		Email         string
		Types         string
		Amount        float64
	}

	var walletTransactions []models.WalletTransaction
	if err := config.DB.Order("created_at DESC").Preload("UserAuth").Find(&walletTransactions).Error; err != nil {
		logger.Log.Error("Failed to fetch wallet transactions", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch wallet Transactions", "Something Went Wrong", "")
		return
	}

	var transaction []walletResponce
	for _, row := range walletTransactions {
		responce := walletResponce{
			ID:            row.ID,
			TransactionID: row.TransactionID[0:12],
			Date:          row.CreatedAt.Format("Jan 02, 2006"),
			ProfilePic:    row.UserAuth.ProfilePic,
			Name:          row.UserAuth.FullName,
			Email:         row.UserAuth.Email,
			Types:         row.Type,
			Amount:        row.Amount,
		}
		transaction = append(transaction, responce)
	}

	logger.Log.Info("Wallet transactions fetched successfully",
		zap.Int("transactionCount", len(walletTransactions)))
	c.HTML(http.StatusOK, "walletManagement.html", gin.H{
		"status": "success",
		"Data":   transaction,
		"code":   http.StatusOK,
	})
}

func ShowTransactionDetails(c *gin.Context) {
	logger.Log.Info("Requested transaction details")

	txnID := c.Param("id")

	type TXNDetails struct {
		TXNID       string
		TXNDate     string
		Amount      float64
		Type        string
		OrderItemID uint
		OrderUID    string
		Receipt     string
		Description string
	}

	var transactionDetail models.WalletTransaction
	if err := config.DB.First(&transactionDetail, txnID).Error; err != nil {
		logger.Log.Error("Failed to fetch transaction details",
			zap.String("transactionID", txnID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch details", "Something Went Wrong", "")
		return
	}

	isOrderAvailable := true
	var order models.Order
	if err := config.DB.First(&order, "order_uid = ?", transactionDetail.OrderId).Error; err != nil {
		logger.Log.Debug("Order not found by UID",
			zap.String("orderUID", transactionDetail.OrderId),
			zap.Error(err))
		isOrderAvailable = false
	}

	type orderItemsDetails struct {
		ID       uint
		OrderUID string
	}

	var orderItem []models.OrderItem
	if transactionDetail.Type != "Deposit" {
		if isOrderAvailable {
			if err := config.DB.Find(&orderItem, "order_id = ?", order.ID).Error; err != nil {
				logger.Log.Error("Failed to fetch order items by order ID",
					zap.Uint("orderID", order.ID),
					zap.Error(err))
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch details", "Something Went Wrong", "")
				return
			}
		} else {
			if err := config.DB.First(&orderItem, "order_uid = ?", transactionDetail.OrderId).Error; err != nil {
				logger.Log.Error("Failed to fetch order items by order UID",
					zap.String("orderUID", transactionDetail.OrderId),
					zap.Error(err))
				helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch details", "Something Went Wrong", "")
				return
			}
		}
	}

	var itmDetails []orderItemsDetails
	if transactionDetail.Type != "Deposit" {
		for _, itm := range orderItem {
			data := orderItemsDetails{
				ID:       itm.ID,
				OrderUID: itm.OrderUID,
			}
			itmDetails = append(itmDetails, data)
		}
	}

	txnDetails := TXNDetails{
		TXNID:       transactionDetail.TransactionID[0:8],
		TXNDate:     transactionDetail.CreatedAt.Format("January 02, 2006 at 03:04 PM"),
		Amount:      transactionDetail.Amount,
		Type:        transactionDetail.Type,
		OrderUID:    transactionDetail.OrderId[0:8],
		Receipt:     transactionDetail.Receipt[0:18],
		Description: transactionDetail.Description,
	}

	var userDetails models.UserAuth
	if err := config.DB.
		Preload("UserProfile").
		First(&userDetails, transactionDetail.UserID).Error; err != nil {
		logger.Log.Error("Failed to fetch user details",
			zap.Uint("userID", transactionDetail.UserID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch details", "Something Went Wrong", "")
		return
	}

	logger.Log.Info("Transaction details fetched successfully",
		zap.String("transactionID", txnID),
		zap.Int("orderItemCount", len(itmDetails)))
	c.HTML(http.StatusOK, "walletDetail.html", gin.H{
		"status":     "success",
		"TxnDetails": txnDetails,
		"User":       userDetails,
		"OrderItems": itmDetails,
		"code":       http.StatusOK,
	})
}
