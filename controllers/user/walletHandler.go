package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/pkg/logger"
	"github.com/anfastk/E-Commerce-Website/utils"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func CreateWallet(c *gin.Context, userID uint) {
	logger.Log.Info("Creating wallet", zap.Uint("userID", userID))

	var wallet models.Wallet
	if err := config.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
		userWallet := models.Wallet{
			UserID:  userID,
			Balance: 0,
		}
		if createErr := config.DB.Create(&userWallet).Error; createErr != nil {
			logger.Log.Error("Failed to create wallet",
				zap.Uint("userID", userID),
				zap.Error(createErr))
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Creation Failed", "Something Went Wrong", "")
			return
		}
		logger.Log.Info("Wallet created successfully",
			zap.Uint("userID", userID),
			zap.Uint("walletID", userWallet.ID))
	} else {
		logger.Log.Debug("Wallet already exists",
			zap.Uint("userID", userID),
			zap.Uint("walletID", wallet.ID))
	}
}

func WalletHandler(c *gin.Context) {
	logger.Log.Info("Fetching wallet details")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	type TransactionHistoryWallet struct {
		Date        string  `json:"date"`
		Description string  `json:"description"`
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	CreateWallet(c, userID)

	var walletDetails models.Wallet
	if err := config.DB.First(&walletDetails, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Wallet not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Something Went Wrong", "")
		return
	}

	var walletTransactions []models.WalletTransaction
	if err := config.DB.Order("created_at DESC").Find(&walletTransactions, "user_id = ? AND wallet_id = ?", userID, walletDetails.ID).Error; err != nil {
		logger.Log.Error("Failed to fetch wallet transactions",
			zap.Uint("userID", userID),
			zap.Uint("walletID", walletDetails.ID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Something Went Wrong", "")
		return
	}

	var history []TransactionHistoryWallet
	for _, his := range walletTransactions {
		row := TransactionHistoryWallet{
			Date:        his.CreatedAt.Format("Jan 02, 2006"),
			Description: his.Description,
			Type:        his.Type,
			Amount:      his.Amount,
		}
		history = append(history, row)
	}

	var referralDetails models.ReferralAccount
	if err := config.DB.First(&referralDetails, "user_id = ?", userID).Error; err != nil {
		logger.Log.Warn("Referral account not found, creating new",
			zap.Uint("userID", userID))
		CreateReferralAccount(c, userauth.ID)
		// Re-fetch after creation
		if err := config.DB.First(&referralDetails, "user_id = ?", userID).Error; err != nil {
			logger.Log.Error("Failed to fetch newly created referral account",
				zap.Uint("userID", userID),
				zap.Error(err))
		}
	}

	CheckForReferrer(c)
	CheckForJoinee(c)

	logger.Log.Info("Wallet details loaded",
		zap.Uint("userID", userID),
		zap.Uint("walletID", walletDetails.ID),
		zap.Int("transactionCount", len(history)))
	c.HTML(http.StatusOK, "profileWallet.html", gin.H{
		"status":             "success",
		"message":            "Order details fetched successfully",
		"User":               userauth,
		"Wallet":             walletDetails,
		"WalletTransactions": history,
		"ReferralDetails":    referralDetails,
	})
}

func AddMoneyTOWalltet(c *gin.Context) {
	logger.Log.Info("Adding money to wallet")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var addMoneyInput struct {
		PaymentMethod string  `json:"paymentMethod"`
		Amount        float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&addMoneyInput); err != nil {
		logger.Log.Error("Failed to bind add money data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Enter All Fields Correctly", "")
		return
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	if addMoneyInput.PaymentMethod == "" || addMoneyInput.Amount == 0 {
		logger.Log.Warn("Invalid payment method or amount",
			zap.String("paymentMethod", addMoneyInput.PaymentMethod),
			zap.Float64("amount", addMoneyInput.Amount))
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Enter All Fields Correctly", "/profile/order/details")
		return
	}

	switch addMoneyInput.PaymentMethod {
	case "Razorpay":
		razorpayOrder, err := CreateRazorpayOrder(c, addMoneyInput.Amount)
		if err != nil {
			logger.Log.Error("Failed to create Razorpay order",
				zap.Uint("userID", userID),
				zap.Float64("amount", addMoneyInput.Amount),
				zap.Error(err))
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create Razorpay order", "Something Went Wrong", "/profile/order/details")
			return
		}

		razorpayOrderID, ok := razorpayOrder["id"].(string)
		if !ok {
			logger.Log.Error("Failed to extract Razorpay order ID",
				zap.Any("razorpayResponse", razorpayOrder))
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Razorpay response", "Something Went Wrong", "/checkout")
			return
		}

		logger.Log.Info("Razorpay order created",
			zap.Uint("userID", userID),
			zap.String("orderID", razorpayOrderID),
			zap.Float64("amount", addMoneyInput.Amount))
		c.JSON(http.StatusOK, gin.H{
			"status":   "OK",
			"order_id": razorpayOrderID,
			"amount":   addMoneyInput.Amount * 100,
			"currency": "INR",
			"key_id":   config.RAZORPAY_KEY_ID,
			"prefill": gin.H{
				"name":  userauth.FullName,
				"email": userauth.Email,
			},
		})

	default:
		logger.Log.Warn("Unsupported payment method",
			zap.String("paymentMethod", addMoneyInput.PaymentMethod),
			zap.Uint("userID", userID))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/profile/order/details")
		return
	}
}

func VerifyAddTOWalletRazorpayPayment(c *gin.Context) {
	logger.Log.Info("Verifying Razorpay payment for wallet")

	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var verifyRequest struct {
		PaymentID string  `json:"razorpay_payment_id"`
		OrderID   string  `json:"razorpay_order_id"`
		Signature string  `json:"razorpay_signature"`
		Amount    float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		logger.Log.Error("Failed to bind verification data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}

	tx := config.DB.Begin()

	var userDetails models.UserAuth
	if err := tx.First(&userDetails, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/profile/order/details")
		return
	}

	data := verifyRequest.OrderID + "|" + verifyRequest.PaymentID
	expectedSignature := hmac.New(sha256.New, []byte(config.RAZORPAY_KEY_SECRET))
	expectedSignature.Write([]byte(data))
	calculatedSignature := hex.EncodeToString(expectedSignature.Sum(nil))

	if calculatedSignature != verifyRequest.Signature {
		logger.Log.Warn("Invalid payment signature",
			zap.Uint("userID", userID),
			zap.String("orderID", verifyRequest.OrderID),
			zap.String("paymentID", verifyRequest.PaymentID))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid signature", "Payment verification failed", "/profile/order/details")
		return
	}

	var wallet models.Wallet
	if err := tx.First(&wallet, "user_id = ?", userID).Error; err != nil {
		logger.Log.Error("Wallet not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Wallet not found", "")
		return
	}

	actualAmount := verifyRequest.Amount / 100
	lastBalance := wallet.Balance
	wallet.Balance += actualAmount
	if err := tx.Save(&wallet).Error; err != nil {
		logger.Log.Error("Failed to update wallet balance",
			zap.Uint("walletID", wallet.ID),
			zap.Float64("amount", actualAmount),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Amount Adding Failed", "Something Went Wrong", "")
		return
	}

	receiptID := "rcpt_" + uuid.New().String()
	createHistory := models.WalletTransaction{
		UserID:        userID,
		WalletID:      wallet.ID,
		Amount:        actualAmount,
		Description:   "Added funds via RazorPay",
		Type:          "Deposit",
		Receipt:       receiptID,
		OrderId:       verifyRequest.OrderID,
		TransactionID: verifyRequest.PaymentID,
		LastBalance:   lastBalance,
		PaymentMethod: "RazorPay",
	}

	if err := tx.Create(&createHistory).Error; err != nil {
		logger.Log.Error("Failed to create wallet transaction",
			zap.Uint("walletID", wallet.ID),
			zap.String("transactionID", verifyRequest.PaymentID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Adding Failed", "Something Went Wrong", "")
		return
	}

	tx.Commit()
	logger.Log.Info("Wallet payment verified and updated",
		zap.Uint("userID", userID),
		zap.Uint("walletID", wallet.ID),
		zap.String("paymentID", verifyRequest.PaymentID),
		zap.Float64("amount", actualAmount))
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment verified successfully",
	})
}

func GenerateGiftCardCode() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	hexStr := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("LPTX-%s-%s-%s-%s", hexStr[0:4], hexStr[4:8], hexStr[8:12], hexStr[12:16])
}

func SendGiftCard(c *gin.Context) {
	logger.Log.Info("Sending Gift Card")
	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var Details struct {
		RecipientName  string `json:"recipient_name"`
		RecipientEmail string `json:"recipient_email"`
		Amount         int    `json:"amount"`
		Message        string `json:"message"`
	}

	if err := c.ShouldBindJSON(&Details); err != nil {
		logger.Log.Error("Failed to bind details data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}

	fmt.Println(Details)
	tx := config.DB.Begin()

	var userDetails models.UserAuth
	if err := tx.First(&userDetails, userID).Error; err != nil {
		logger.Log.Error("User not found",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/profile/order/details")
		return
	}

	var walletDetails models.Wallet
	if err := tx.First(&walletDetails, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		logger.Log.Error("Wallet Not Found", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Wallet Not Found", "Something Went Wrong", "")
		return
	}

	if walletDetails.Balance < float64(Details.Amount) {
		tx.Rollback()
		logger.Log.Error("Wallet Balance Is Lessthan Entered Amount")
		helper.RespondWithError(c, http.StatusBadRequest, "Wallet Balance Is Lessthan Entered Amount", "Wallet Balance Is Lessthan Entered Amount", "")
		return
	}

	GiftCode := GenerateGiftCardCode()
	transactionID := fmt.Sprintf("%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	cleanedGiftCode := strings.ReplaceAll(GiftCode, " ", "")
	cleanedGiftCode = strings.ReplaceAll(cleanedGiftCode, "-", "")

	data := models.WalletGiftCard{
		GiftCardCode:   cleanedGiftCode,
		GiftCardValue:  float64(Details.Amount),
		ExpDate:        time.Now().AddDate(1, 0, 0),
		UserID:         userID,
		RecipientName:  strings.ToUpper(Details.RecipientName),
		RecipientEmail: Details.RecipientEmail,
		Message:        Details.Message,
		Status:         "Active",
		PaymentMethod:  "Wallet",
		TransactionID:  transactionID,
	}

	if err := tx.Create(&data).Error; err != nil {
		tx.Rollback()
		logger.Log.Error("Failed to create gift card", zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create gift card", "Something WEnt Wrong", "")
		return
	}
	formattedExpDate := data.ExpDate.Format("January 02, 2006")
	giftCardValueStr := fmt.Sprintf("%.2f", data.GiftCardValue)
	if err := utils.SendGiftCardToEmail(data.RecipientName, userDetails.ProfilePic, data.Message, data.RecipientEmail, giftCardValueStr, GiftCode, formattedExpDate); err != nil {
		logger.Log.Error("Failed to send gift card to email",
			zap.String("email", data.RecipientEmail),
			zap.Error(err))
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to send Gift Card", "Failed To Gift Card , Please Try Again ", "")
		return
	}

	lastBalance := walletDetails.Balance
	walletDetails.Balance -= data.GiftCardValue
	if err := tx.Save(&walletDetails).Error; err != nil {
		logger.Log.Error("Failed to update wallet balance",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update wallet details", "Failed to update order", "")
		return
	}

	walletReceiptID := "rcpt-" + uuid.New().String()
	rand.Seed(time.Now().UnixNano())
	walletTransactionID := fmt.Sprintf("-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	orderUID := helper.GenerateOrderID()

	walletHistory := models.WalletTransaction{
		UserID:        userID,
		WalletID:      walletDetails.ID,
		Amount:        data.GiftCardValue,
		LastBalance:   lastBalance,
		Description:   "Send A Gift Card To (" + data.RecipientName + "). Email ID :" + data.RecipientEmail,
		Type:          "Gift Send",
		Receipt:       walletReceiptID,
		OrderId:       orderUID,
		TransactionID: walletTransactionID,
		PaymentMethod: "Wallet",
	}
	if err := tx.Create(&walletHistory).Error; err != nil {
		logger.Log.Error("Failed to create wallet transaction",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Creation Failed", "Something Went Wrong", "/cart")
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Gift Card Send Successfully",
		"Code":    http.StatusOK,
	})
}

func RedeemGiftCard(c *gin.Context) {
	logger.Log.Info("Redeeming Gift Card")
	userID := helper.FetchUserID(c)
	logger.Log.Debug("Fetched user ID", zap.Uint("userID", userID))

	var userInput struct {
		GiftCode string `json:"code"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		logger.Log.Error("Failed to bind details data", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}

	cleanedGiftCode := strings.ReplaceAll(userInput.GiftCode, " ", "")
	cleanedGiftCode = strings.ReplaceAll(cleanedGiftCode, "-", "")

	tx := config.DB.Begin()
	var giftCardDetails models.WalletGiftCard
	if err := tx.First(&giftCardDetails, "gift_card_code = ?", cleanedGiftCode).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Invalid Code Entered", "Invalid Code Entered", "")
		return
	}

	if giftCardDetails.GiftCardCode != cleanedGiftCode {
		helper.RespondWithError(c, http.StatusNotFound, "Invalid Code Entered", "Invalid Code Entered", "")
		return
	} else if giftCardDetails.UserID == userID {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Usage", "You cannot use your own gift code.", "")
		return
	} else if giftCardDetails.Status == "Redeemed" || giftCardDetails.RedeemedUserID != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Code Already redeemed", "Code Already redeemed", "")
		return
	} else if giftCardDetails.ExpDate.Before(time.Now()) {
		giftCardDetails.Status = "Expired"
		if err := config.DB.Save(&giftCardDetails).Error; err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update code", "Something Went Wrong", "")
			return
		}
		helper.RespondWithError(c, http.StatusBadRequest, "Code Expired", "Code Expired", "")
		return
	}

	var walletDetails models.Wallet
	if err := tx.First(&walletDetails, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		logger.Log.Error("Wallet Not Found", zap.Error(err))
		helper.RespondWithError(c, http.StatusBadRequest, "Wallet Not Found", "Something Went Wrong", "")
		return
	}

	lastBalance := walletDetails.Balance
	walletDetails.Balance += giftCardDetails.GiftCardValue
	if err := tx.Save(&walletDetails).Error; err != nil {
		logger.Log.Error("Failed to update wallet balance",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update wallet details", "Failed to update order", "")
		return
	}

	walletReceiptID := "rcpt-" + uuid.New().String()
	rand.Seed(time.Now().UnixNano())
	walletTransactionID := fmt.Sprintf("-%d-%d", time.Now().UnixNano(), rand.Intn(10000))
	orderUID := helper.GenerateOrderID()

	walletHistory := models.WalletTransaction{
		UserID:        userID,
		WalletID:      walletDetails.ID,
		Amount:        giftCardDetails.GiftCardValue,
		LastBalance:   lastBalance,
		Description:   "Redeemed A Gift Card ",
		Type:          "Gift Redeemed",
		Receipt:       walletReceiptID,
		OrderId:       orderUID,
		TransactionID: walletTransactionID,
		PaymentMethod: "Gift Card",
	}
	if err := tx.Create(&walletHistory).Error; err != nil {
		logger.Log.Error("Failed to create wallet transaction",
			zap.Uint("userID", userID),
			zap.Error(err))
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Creation Failed", "Something Went Wrong", "/cart")
		return
	}

	giftCardDetails.Status = "Redeemed"
	giftCardDetails.RedeemedUserID = &userID
	if err := config.DB.Save(&giftCardDetails).Error; err != nil {
		helper.RespondWithError(c, http.StatusInternalServerError, "Failed to update code", "Something Went Wrong", "")
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  "Success",
		"message": "Gift Card Redeemed Successfully",
		"Code":    http.StatusOK,
	})
}
