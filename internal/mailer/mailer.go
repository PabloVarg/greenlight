package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"math"
	"time"

	"gopkg.in/gomail.v2"
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer gomail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	return Mailer{
		dialer: *gomail.NewDialer(host, port, username, password),
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return err
	}

	email := gomail.NewMessage()
	email.SetHeader("To", recipient)
	email.SetHeader("From", m.sender)
	email.SetHeader("Subject", subject.String())
	email.SetBody("text/plain", plainBody.String())
	email.AddAlternative("text/html", htmlBody.String())

	for i := range 3 {
		err := m.dialer.DialAndSend(email)

		if nil == err {
			return nil
		}

		time.Sleep(1*time.Second + (time.Duration(math.Pow(500, float64(i))) * time.Millisecond))
	}

	return nil
}
