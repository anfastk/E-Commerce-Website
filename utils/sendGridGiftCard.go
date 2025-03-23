package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendGiftCardToEmail(senderName,senderProdile, message, email, amount, giftCode, expDate string) error {
	from := mail.NewEmail("LAPTIX E-Commerce", "laptixinfo@gmail.com")
	subject := "You've Received a Gift Card!"
	to := mail.NewEmail("Recipient", email)
	
	// Prepare plain text content
	plainTextContent := fmt.Sprintf("You've received a gift card!\nAmount: %s\nCode: %s\nExpires: %s\nFrom: %s\nMessage: %s",
		amount, giftCode, expDate, senderName, message)
	
	// Replace placeholders in HTML template
	htmlContent := giftCardTemplate
	htmlContent = strings.ReplaceAll(htmlContent, "%SENDER_NAME%", senderName)
	htmlContent = strings.ReplaceAll(htmlContent, "%MESSAGE%", message)
	htmlContent = strings.ReplaceAll(htmlContent, "%AMOUNT%", amount)
	htmlContent = strings.ReplaceAll(htmlContent, "%GIFT_CODE%", giftCode)
	htmlContent = strings.ReplaceAll(htmlContent, "%EXP_DATE%", expDate)
	htmlContent = strings.ReplaceAll(htmlContent, "%SENDER_INITIALS%", senderProdile)

	messageEmail := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(messageEmail)
	return err
}

const giftCardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Your Gift Card</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: 'Helvetica Neue', Arial, sans-serif;
            background-color: #121212;
            color: #f5f5f5;
        }
        .container {
            max-width: 100%;
            margin: 0 auto;
            padding: 20px;
            background-color: #1e1e1e;
        }
        .header {
            text-align: center;
            padding: 20px 0;
            border-bottom: 1px solid #333;
        }
        .logo-container {
            margin-bottom: 20px;
            text-align: center;
        }
        .logo-image {
            max-width: 180px;
            height: auto;
        }
        .gift-card-container {
            background: linear-gradient(135deg, #2d2d2d 0%, #1a1a1a 100%);
            border-radius: 12px;
            padding: 30px;
            margin: 25px 0;
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.4);
            text-align: center;
            position: relative;
            overflow: hidden;
        }
        .gift-card-container::before {
            content: "";
            position: absolute;
            top: -50px;
            left: -50px;
            width: 100px;
            height: 100px;
            background: rgba(187, 134, 252, 0.1);
            border-radius: 50%;
        }
        .gift-card-container::after {
            content: "";
            position: absolute;
            bottom: -50px;
            right: -50px;
            width: 100px;
            height: 100px;
            background: rgba(187, 134, 252, 0.1);
            border-radius: 50%;
        }
        .gift-card-title {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 15px;
            color: #e0e0e0;
        }
        .gift-card-message {
            margin-bottom: 25px;
            font-size: 16px;
            line-height: 1.5;
            color: #b0b0b0;
        }
        .gift-card-code-container {
            position: relative;
            margin: 30px 0;
        }
        .gift-card-code {
            background-color: #2a2a2a;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Courier New', monospace;
            font-size: 22px;
            letter-spacing: 2px;
            color: #fff;
            border: 1px dashed #444;
            position: relative;
            z-index: 1;
        }
        .copy-hint {
            font-size: 12px;
            color: #bb86fc;
            margin-top: 5px;
            font-style: italic;
        }
        .amount {
            font-size: 36px;
            font-weight: bold;
            color: #bb86fc;
            margin: 15px 0;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }
        .expiry-date {
            background-color: #2a2a2a;
            border-radius: 20px;
            padding: 8px 15px;
            display: inline-block;
            margin: 15px 0;
            font-size: 14px;
            border-left: 3px solid #bb86fc;
        }
        .expiry-date strong {
            color: #bb86fc;
        }
        .divider {
            height: 1px;
            background: linear-gradient(to right, transparent, #444, transparent);
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #777;
            border-top: 1px solid #333;
            margin-top: 20px;
        }
        .button {
            display: inline-block;
            background-color: #bb86fc;
            color: #121212;
            text-decoration: none;
            padding: 12px 30px;
            border-radius: 25px;
            font-weight: bold;
            margin: 20px 0;
            transition: background-color 0.3s, transform 0.2s;
            box-shadow: 0 4px 6px rgba(0,0,0,0.2);
        }
        .button:hover {
            background-color: #a370d8;
            transform: translateY(-2px);
            box-shadow: 0 6px 8px rgba(0,0,0,0.3);
        }
        .terms {
            font-size: 11px;
            color: #666;
            max-width: 450px;
            margin: 15px auto;
            line-height: 1.4;
        }
        .contact-info {
            margin-top: 15px;
            color: #888;
        }
        .highlight {
            color: #bb86fc;
        }
        .card-decoration {
            position: absolute;
            width: 120px;
            height: 120px;
            border-radius: 60px;
            background: linear-gradient(45deg, rgba(187, 134, 252, 0.05), transparent);
            top: -30px;
            right: -30px;
            z-index: 0;
        }
        .sender-details {
            background-color: #252525;
            border-radius: 8px;
            padding: 15px;
            margin: 20px 0;
            text-align: left;
            border-left: 3px solid #bb86fc;
        }
        .sender-details h3 {
            margin-top: 0;
            margin-bottom: 10px;
            color: #e0e0e0;
        }
        .sender-message {
            font-style: italic;
            color: #b0b0b0;
            line-height: 1.5;
        }
        .sender-signature {
            margin-top: 15px;
            font-weight: bold;
            color: #e0e0e0;
        }
        .avatar-container {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
        }
        .avatar {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background-color: #bb86fc;
            margin-right: 15px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: bold;
            color: #1e1e1e;
        }
        .sender-name {
            font-weight: bold;
            color: #e0e0e0;
        }
        .sender-relation {
            font-size: 12px;
            color: #b0b0b0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo-container">
                <a href="https://www.laptix.com">
                    <img src="https://res.cloudinary.com/dghzlcoco/image/upload/v1742683014/text-1742682998645_rtgwj1.png" alt="LAPTIX" class="logo-image">
                </a>
            </div>
            <h1>Your Gift Card Has Arrived!</h1>
        </div>
        
        <div class="sender-details">
            <div class="avatar-container">
                <img src="%SENDER_INITIALS%" class="avatar">
                <div>
                    <div class="sender-name">%SENDER_NAME%</div>
                    <div class="sender-relation">Your Friend</div>
                </div>
            </div>
            <div class="sender-message">
                %MESSAGE%
            </div>
            <div class="sender-signature">Cheers, %SENDER_NAME%</div>
        </div>
        
        <div class="gift-card-container">
            <div class="card-decoration"></div>
            <div class="gift-card-title">Special Gift Just For You</div>
            <div class="gift-card-message">
                Thank you for being an amazing friend. Here's a gift card to show my appreciation!
            </div>
            <div class="amount">&#8377;%AMOUNT%</div>
            
            <div class="gift-card-code-container">
                <div class="gift-card-code">%GIFT_CODE%</div>
                <div class="copy-hint">Click to copy</div>
            </div>
            
            <div class="expiry-date">
                Valid until: <strong>%EXP_DATE%</strong>
            </div>
            
            <p>Use this code at checkout to redeem your gift.</p>
            <a href="#" class="button">Shop Now</a>
            
            <div class="divider"></div>
            
            <p>This gift card can be used for any product on our website.</p>
        </div>
        
        <div class="footer">
            <p>Need help? Contact our <span class="highlight">customer support</span></p>
            <div class="contact-info">
                support@laptix.com | (555) 123-4567
            </div>
            <div class="terms">
                Terms & Conditions: Gift card expires on the date shown. Cannot be combined with other promotions. No cash value. Unused balance remains on card. Lost or stolen cards cannot be replaced.
            </div>
            <p>© 2025 LAPTIX. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`