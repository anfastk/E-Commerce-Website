package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendOTPToEmail(email, otp string) error {
	from := mail.NewEmail("LAPTIX E-Commerce", "laptixinfo@gmail.com")
	subject := "Your OTP Code"
	to := mail.NewEmail("User", email)
	plainTextContent := fmt.Sprintf("Your OTP is: %s", otp)
	htmlContent := strings.ReplaceAll(htmlTemplate, "%OTP%", otp)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	return err
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LAPTIX Verification</title>
    <style>
        body {
            margin: 0;
            padding: 0;
            font-family: 'Helvetica Neue', Arial, sans-serif;
            background-color: #121212;
            color: #f5f5f5;
            width: 100%;
        }
        .container {
            width: 100%;
            margin: 0 auto;
            padding: 20px;
            background-color: #1e1e1e;
            box-sizing: border-box;
        }
        .header {
            text-align: center;
            padding: 20px 0;
            border-bottom: 1px solid #333;
            width: 100%;
        }
        .logo-container {
            margin-bottom: 20px;
            text-align: center;
        }
        .logo-image {
            max-width: 180px;
            height: auto;
        }
        .otp-container {
            background: linear-gradient(135deg, #2d2d2d 0%, #1a1a1a 100%);
            border-radius: 12px;
            padding: 30px;
            margin: 25px 0;
            box-shadow: 0 4px 15px rgba(0,0,0,0.4);
            text-align: center;
            position: relative;
            overflow: hidden;
            width: 100%;
            box-sizing: border-box;
        }
        .otp-container::before {
            content: "";
            position: absolute;
            top: -50px;
            left: -50px;
            width: 100px;
            height: 100px;
            background: rgba(187, 134, 252, 0.1);
            border-radius: 50%;
        }
        .otp-container::after {
            content: "";
            position: absolute;
            bottom: -50px;
            right: -50px;
            width: 100px;
            height: 100px;
            background: rgba(187, 134, 252, 0.1);
            border-radius: 50%;
        }
        .otp-title {
            font-size: 24px;
            font-weight: bold;
            margin-bottom: 15px;
            color: #e0e0e0;
        }
        .otp-message {
            margin-bottom: 25px;
            font-size: 16px;
            line-height: 1.5;
            color: #b0b0b0;
        }
        .otp-code-container {
            position: relative;
            margin: 30px 0;
        }
        .otp-code {
            background-color: #2a2a2a;
            padding: 15px 25px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
            font-size: 28px;
            letter-spacing: 8px;
            color: #bb86fc;
            border: 1px dashed #444;
            position: relative;
            z-index: 1;
            display: inline-block;
            font-weight: bold;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }
        .expiry-info {
            background-color: #2a2a2a;
            border-radius: 20px;
            padding: 8px 15px;
            display: inline-block;
            margin: 15px 0;
            font-size: 14px;
            border-left: 3px solid #bb86fc;
        }
        .expiry-info strong {
            color: #bb86fc;
        }
        .divider {
            height: 1px;
            background: linear-gradient(to right, transparent, #444, transparent);
            margin: 20px 0;
            width: 100%;
        }
        .footer {
            text-align: center;
            padding: 20px;
            font-size: 12px;
            color: #777;
            border-top: 1px solid #333;
            margin-top: 20px;
            width: 100%;
            box-sizing: border-box;
        }
        .warning {
            color: #ff5c5c;
            font-weight: bold;
            margin: 15px 0;
            font-size: 14px;
            background-color: rgba(255, 92, 92, 0.1);
            padding: 10px;
            border-radius: 6px;
            border-left: 3px solid #ff5c5c;
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
    </style>
</head>
<body>
    <div class="container">
        <div class="header"> 
            <div class="logo-container">
                <!-- Replace with your actual logo image and website URL -->
                <a href="https://www.laptix.com">
                    <img src="https://res.cloudinary.com/dghzlcoco/image/upload/v1742683014/text-1742682998645_rtgwj1.png" alt="LAPTIX" class="logo-image">
                </a>
            </div>
            <h1>Verification Required</h1>
        </div>
        
        <div class="otp-container">
            <div class="card-decoration"></div>
            <div class="otp-title">Verification Code</div>
            <div class="otp-message">
                Please use the following one-time password (OTP) to complete your verification.
            </div>
            
            <div class="otp-code-container">
                <div class="otp-code">%OTP%</div>
            </div>
            
            <div class="expiry-info">
                This code will expire in <strong>10 minutes</strong>
            </div>
            
            <div class="divider"></div>
            
            <div class="warning">
                Do not share this code with anyone, including LAPTIX staff.
            </div>
            
            <p>If you did not request this code, please ignore this email or contact support.</p>
        </div>
        
        <div class="footer">
            <p>Need help? Contact our <span class="highlight">customer support</span></p>
            <div class="contact-info">
                support@laptix.com | (555) 123-4567
            </div>
            <p>&copy; 2025 LAPTIX. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`
