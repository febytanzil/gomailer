package gomailer

import (
	"context"
	"errors"
)

type Implementation int

const (
	Gomail = Implementation(iota)
)

type Client interface {
	// Send will send email
	Send(msg *Message) error
	// SendContext provide context to send function
	SendContext(ctx context.Context, msg *Message) error
	// Close permanently close client connection
	Close() error
}

type Message struct {
	Attachments []*Attachment
	SendTo      []string
	CC          []string
	BCC         []string
	Title       string
	Body        string
}

type Attachment struct {
	Filename string
	Byte     []byte
}

type Config struct {
	Host     string
	Port     int
	Email    string
	Password string
}

var ErrClosed = errors.New("connection has been closed")

// New email return email handler struct
func NewClient(impl Implementation, emailConfig *Config) (Client, error) {
	if Gomail == impl {
		return newGomail(emailConfig), nil
	}

	return nil, errors.New("no email implementations found")
}
