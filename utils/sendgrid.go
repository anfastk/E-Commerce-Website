package utils

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func SendOTPToEmail(email, otp string) error {
	from := mail.NewEmail("LAPTIX E-Commerce", "laptixinfo@gmail.com")
	subject := "Your OTP Code"
	to := mail.NewEmail("User", email)
	plainTextContent := fmt.Sprintf("Your OTP is: %s", otp)
	htmlContent := fmt.Sprintf("<strong>Your OTP is: %s</strong>", otp)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	return err
}