package email

import (
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func SendEmail(to string, subject string, body string) error {

	err := godotenv.Load()

	if err != nil {
		return err
	}

	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))

	if err != nil {
		return err
	}

	return nil
}
