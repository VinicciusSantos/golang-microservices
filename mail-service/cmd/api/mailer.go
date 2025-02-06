package main

import (
	"bytes"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) (err error) {
	var (
		client *mail.SMTPClient
		formattedMessage string
		plainMessage string
	)

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	msg.DataMap = map[string]any{
		"message": msg.Data,
	}

	if formattedMessage, err = m.buildHTMLMessage(msg); err != nil {
		return err
	}

	if plainMessage, err = m.buildPlainMessage(msg); err != nil {
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

	if client, err = server.Connect(); err != nil {
		return err
	}

	email := mail.
		NewMSG().
		SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextHTML, formattedMessage).
		AddAlternative(mail.TextPlain, plainMessage)

	for _, attachment := range msg.Attachments {
		email.AddAttachment(attachment)
	}

	if err := email.Send(client); err != nil {
		return err
	}

	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (formattedMessage string, err error) {
	var (
		templateToRender = "./templates/mail.html.gohtml"
		tpl              bytes.Buffer
		t                *template.Template
	)

	if t, err = template.New("email-html").ParseFiles(templateToRender); err != nil {
		return "", err
	}

	if err := t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	if formattedMessage, err = m.inlineCSS(tpl.String()); err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) inlineCSS(s string) (formattedMessage string, err error) {
	var prem premailer.Premailer

	if prem, err = premailer.NewPremailerFromString(s, &premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}); err != nil {
		return "", err
	}

	return prem.Transform()
}

func (m *Mail) buildPlainMessage(msg Message) (plainMessage string, err error) {
	var (
		templateToRender = "./templates/mail.plain.gohtml"
		tpl              bytes.Buffer
		t                *template.Template
	)

	if t, err = template.New("email-plain").ParseFiles(templateToRender); err != nil {
		return "", err
	}

	if err := t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *Mail) getEncryption(encryption string) mail.Encryption {
	switch encryption {
	case "ssl":
		return mail.EncryptionSSL
	case "tls":
		return mail.EncryptionTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
