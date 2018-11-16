package email

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"io"
)

type gomailHandler struct {
	sender gomail.SendCloser
	config *Config
}

func newGomail(emailConfig *Config) (*gomailHandler, error) {
	dialer := gomail.NewPlainDialer("smtp.gmail.com", 587, emailConfig.Email, emailConfig.Password)
	s, err := dialer.Dial()
	if nil != err {
		return nil, err
	}

	return &gomailHandler{
		sender: s,
		config: emailConfig,
	}, nil
}

func (h *gomailHandler) Send(sendEmailTo *MailDetail) error {
	m := gomail.NewMessage()
	m.SetHeader("From", h.config.Email)
	m.SetHeader("To", sendEmailTo.SendTo...)
	if len(sendEmailTo.SendToCc) > 0 {
		m.SetHeader("Cc", sendEmailTo.SendToCc...)
	}
	m.SetHeader("Subject", sendEmailTo.Title)
	m.SetBody("text/html", sendEmailTo.Body)

	for _, element := range sendEmailTo.Attachment {
		fileByte := element.Byte
		m.Attach(element.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(fileByte)
			return err
		}))
	}
	if err := gomail.Send(h.sender, m); err != nil {
		return fmt.Errorf("error sending email: %s", err)
	}

	return nil
}

func (h *gomailHandler) Close() error {
	if h.sender == nil {
		return nil
	}

	return h.sender.Close()
}
