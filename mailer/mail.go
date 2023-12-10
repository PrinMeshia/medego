package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	apimail "github.com/ainsleyclark/go-mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

// ListenForMail listens to the mail channel and sends mail
func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs
		err := m.Send(msg)
		if err != nil {
			m.Results <- Result{false, err}
		} else {
			m.Results <- Result{true, nil}
		}
	}
}

// Send sends an email message using the correct method
func (m *Mail) Send(msg Message) error {
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		return m.ChooseAPI(msg)
	}
	return m.SendSMTPMessage(msg)
}

// ChooseAPI chooses API to use
func (m *Mail) ChooseAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid":
		return m.SendUsingAPI(msg)
	default:
		return fmt.Errorf("unknown API %s; only mailgun, sparkpost or sendgrid accepted", m.API)
	}
}

// SendUsingAPI sends a message using the appropriate API
func (m *Mail) SendUsingAPI(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(m.API, cfg)
	if err != nil {
		return err
	}

	formattedMessage, plainMessage, err := m.buildMessages(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	return err
}

// addAPIAttachments adds attachments to mail being sent via API
func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []apimail.Attachment

		for _, x := range msg.Attachments {
			content, err := os.ReadFile(x)
			if err != nil {
				return err
			}

			fileName := filepath.Base(x)
			attachments = append(attachments, apimail.Attachment{Bytes: content, Filename: fileName})
		}

		tx.Attachments = attachments
	}

	return nil
}

// SendSMTPMessage builds and sends an email message using SMTP
func (m *Mail) SendSMTPMessage(msg Message) error {
	formattedMessage, plainMessage, err := m.buildMessages(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG().
		SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextHTML, formattedMessage).
		AddAlternative(mail.TextPlain, plainMessage)

	for _, x := range msg.Attachments {
		email.AddAttachment(x)
	}

	return email.Send(smtpClient)
}

// buildMessages creates both the HTML and plaintext versions of the message
func (m *Mail) buildMessages(msg Message) (string, string, error) {
	htmlTemplate := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)
	plainTemplate := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	htmlMessage, err := m.buildMessage(htmlTemplate, msg)
	if err != nil {
		return "", "", err
	}

	plainMessage, err := m.buildMessage(plainTemplate, msg)
	if err != nil {
		return "", "", err
	}

	return htmlMessage, plainMessage, nil
}

// buildMessage creates either the HTML or plaintext version of the message
func (m *Mail) buildMessage(templatePath string, msg Message) (string, error) {
	t, err := template.New("email-template").ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	message := tpl.String()

	if templatePath == "html" {
		message, err = m.inlineCSS(message)
		if err != nil {
			return "", err
		}
	}

	return message, nil
}

// getEncryption returns the appropriate encryption type based on a string value
func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

// inlineCSS takes HTML input as a string and inlines CSS where possible
func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}