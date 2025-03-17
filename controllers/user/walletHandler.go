package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"

	"github.com/anfastk/E-Commerce-Website/config"
	"github.com/anfastk/E-Commerce-Website/models"
	"github.com/anfastk/E-Commerce-Website/utils/helper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateWallet(c *gin.Context, userID uint) {
	var wallet models.Wallet
	if err := config.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
		userWallet := models.Wallet{
			UserID:  userID,
			Balance: 0,
		}
		if createErr := config.DB.Create(&userWallet).Error; createErr != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Creation Failed", "Something Went Wrong", "")
			return
		}
	}
}

func WalletHandler(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	type TransactionHistoryWallet struct {
		Date        string  `json:"date"`
		Description string  `json:"description"`
		Type        string  `json:"type"`
		Amount      float64 `json:"amount"`
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}
	CreateWallet(c, userID)
	var walletDetails models.Wallet
	if err := config.DB.First(&walletDetails, "user_id = ?", userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Something Went Wrong", "")
		return
	}

	var walletTransactons []models.WalletTransaction
	if err := config.DB.Order("created_at DESC").Find(&walletTransactons, "user_id = ? AND wallet_id = ?", userID, walletDetails.ID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Something Went Wrong", "")
		return
	}

	var history []TransactionHistoryWallet
	for _, his := range walletTransactons {
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
		CreateCart(c, userauth.ID)
	}

	CheckForReferrer(c)
	CheckForJoinee(c)

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
	
	userID := helper.FetchUserID(c)

	var addMoneyInput struct {
		PaymentMethod string  `json:"paymentMethod"`
		Amount        float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&addMoneyInput); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid data", "Enter All Fields Correctly", "")
		return
	}

	var userauth models.UserAuth
	if err := config.DB.First(&userauth, userID).Error; err != nil {
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	/* amount, err := strconv.Atoi(addMoneyInput.Amount)
	if err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Amount Enterd", "Enter Correct Amount", "")
		return
	} */

	if addMoneyInput.PaymentMethod == "" || addMoneyInput.Amount == 0 {
		helper.RespondWithError(c, http.StatusBadRequest, "Request Not Found", "Enter All Fields Correctly", "/profile/order/details")
		return
	}

	switch addMoneyInput.PaymentMethod {
	case "Razorpay":
		razorpayOrder, err := CreateRazorpayOrder(c, addMoneyInput.Amount)
		if err != nil {
			helper.RespondWithError(c, http.StatusInternalServerError, "Failed to create Razorpay order", "Something Went Wrong", "/profile/order/details")
			return
		}

		razorpayOrderID, ok := razorpayOrder["id"].(string)
		if !ok {
			log.Println("Failed to extract order ID from Razorpay response:", razorpayOrder)
			helper.RespondWithError(c, http.StatusInternalServerError, "Invalid Razorpay response", "Something Went Wrong", "/checkout")
			return
		}

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
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid Payment Method", "Invalid Payment Method", "/profile/order/details")
		return

	}
}

func VerifyAddTOWalletRazorpayPayment(c *gin.Context) {
	
	userID := helper.FetchUserID(c)

	var verifyRequest struct {
		PaymentID string  `json:"razorpay_payment_id"`
		OrderID   string  `json:"razorpay_order_id"`
		Signature string  `json:"razorpay_signature"`
		Amount    float64 `json:"amount"`
	}

	if err := c.ShouldBindJSON(&verifyRequest); err != nil {
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid request", "Invalid request format", "")
		return
	}
	tx := config.DB.Begin()

	var userDetails models.UserAuth

	if err := tx.First(&userDetails, userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "User not found", "User not found", "/profile/order/details")
		return
	}

	data := verifyRequest.OrderID + "|" + verifyRequest.PaymentID
	expectedSignature := hmac.New(sha256.New, []byte(config.RAZORPAY_KEY_SECRET))
	expectedSignature.Write([]byte(data))
	calculatedSignature := hex.EncodeToString(expectedSignature.Sum(nil))

	if calculatedSignature != verifyRequest.Signature {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusBadRequest, "Invalid signature", "Payment verification failed", "/profile/order/details")
		return
	}

	var wallet models.Wallet

	if err := tx.First(&wallet, "user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusNotFound, "Wallet not found", "Wallet not found", "")
		return
	}

	actualAmount := verifyRequest.Amount / 100

	lastBalace := wallet.Balance
	wallet.Balance += actualAmount
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Amount Addung Failed", "Something Went Wrong", "")
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
		LastBalance:   lastBalace,
		PaymentMethod: "RazorPay",
	}

	if err := tx.Create(&createHistory).Error; err != nil {
		tx.Rollback()
		helper.RespondWithError(c, http.StatusInternalServerError, "Wallet Transaction Adding Failed", "Something Went Wrong", "")
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Payment verified successfully",
	})
}
