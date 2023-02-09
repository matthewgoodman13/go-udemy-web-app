package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed templates
var emailTemplatesFS embed.FS

// SendMail sends an email
func (app *application) SendMail(from, to, subject, tmpl string, data interface{}) error {
	templateToRender := fmt.Sprintf("templates/%s.html.tmpl", tmpl)

	// Parse the template
	t, err := template.New("email-html").ParseFS(emailTemplatesFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	// Create a new buffer to write our email template to
	var tpl bytes.Buffer
	if err := t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}
	formattedMessage := tpl.String()

	// Prepare Plain Text message
	templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)
	t, err = template.New("email-plain").ParseFS(emailTemplatesFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	if err := t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}
	plainMessage := tpl.String()

	// Send the mail
	server := mail.NewSMTPClient()
	server.Host = app.config.smtp.host
	server.Port = app.config.smtp.port
	server.Username = app.config.smtp.username
	server.Password = app.config.smtp.password
	server.Encryption = mail.EncryptionSTARTTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smptClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(subject)
	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if err := email.Send(smptClient); err != nil {
		app.errorLog.Println(err)
		return err
	}

	app.infoLog.Println("Email sent successfully")
	return nil
}
