package media

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"text/template"

	"github.com/jordan-wright/email"
)

func SendEmailVerification(
	toEmail string, 
	businessName string, 
	otp string,
) (error) {

	fromEmail := os.Getenv("GMAIL_USERNAME")
	password := os.Getenv("GMAIL_PASSWORD")

	e := email.NewEmail()
	e.From = fmt.Sprintf("Usual <%s>", fromEmail)
	
	e.To = []string{toEmail}
	e.Subject = "Email Verification"

	var body bytes.Buffer
	t, err := template.ParseFiles("./assets/html/verify_email.html")
	if err != nil {
		return err
	}

	err = t.Execute(&body, struct {
		Name    string
		OTP string
	}{ Name:  businessName, OTP: otp })
	
	if err != nil {
		return err
	}

	e.HTML = body.Bytes()

	b, err := os.ReadFile("./assets/images/logo2.png") // just pass the file name
    if err != nil {
        return err
    }

	_, err = e.Attach(bytes.NewReader(b), "image1", "image/png")
	if err != nil {
		return err
	}


	err = e.Send("smtp.gmail.com:587", 
		smtp.PlainAuth("", fromEmail, password, "smtp.gmail.com"))
	if err != nil {
		return err
	}

	return nil
}