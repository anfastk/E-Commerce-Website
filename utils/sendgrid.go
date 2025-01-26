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
	htmlContent := strings.ReplaceAll(htmlTemplate,"%OTP%",otp)
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
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background-color: white;
            border-radius: 5px;
            padding: 30px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .logo {
            text-align: center;
            font-size: 24px;
            font-weight: bold;
            color: #2c3e50;
            margin-bottom: 20px;
        }
        .otp-section {
            text-align: center;
            background-color: #f9f9f9;
            padding: 20px;
            border-radius: 5px;
        }
        .otp-code {
            font-size: 28px;
            letter-spacing: 5px;
            color: #3498db;
            margin: 15px 0;
            border: 2px solid #3498db;
            border-radius: 8px;
            display: inline-block;
            padding: 10px 20px;
            background-color: #f0f8ff;
        }
        .footer {
            text-align: center;
            color: #666;
            font-size: 12px;
            margin-top: 20px;
        }
        .warning {
            color: #e74c3c;
            font-size: 14px;
            margin-top: 15px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">LAPTIX</div>
        <div class="otp-section">
            <h2>Verification Code</h2>
            <p>Your One-Time Password is:</p>
            <div class="otp-code">%OTP%</div>
            <p>This code will expire in 10 minutes.</p>
            <p>Do not share this with anyone.</p>
            <p>If you did not request this code, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 LAPTIX. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`