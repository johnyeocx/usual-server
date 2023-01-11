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
) {

	// fromEmail := os.Getenv("GMAIL_USERNAME")
	// password := os.Getenv("GMAIL_PASSWORD")

	

	e := email.NewEmail()
	e.From = "Usual <team@usual.ltd>"
	e.To = []string{toEmail}
	e.Subject = "Email Verification"

	var body bytes.Buffer
	t, _ := template.ParseFiles("./assets/html/verify_email.html")
	fmt.Println(businessName, otp)
	t.Execute(&body, struct {
		Name    string
		OTP string
	}{
		Name:  businessName,
		OTP: otp,
	})

	e.HTML = body.Bytes()

	b, err := os.ReadFile("./assets/images/logo2.png") // just pass the file name
    if err != nil {
        fmt.Print(err)
    }

	e.Attach(bytes.NewReader(b), "image1", "image/png")

	e.Send("smtp.gmail.com:587", smtp.PlainAuth("", "team@usual.ltd", "fbpizupqwbshvgcb", "smtp.gmail.com"))
}