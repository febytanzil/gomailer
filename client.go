package gomailer

import (
	"context"
	"errors"
)

type Implementation int

const (
	Gomail = Implementation(iota)
	Postmark
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
	Host string
	Port int

	// Email configures the sender's email
	Email    string
	Password string

	// Postmark Settings
	ServerToken  string
	AccountToken string
}

var ErrClosed = errors.New("connection has been closed")

// New email return email handler struct
func NewClient(impl Implementation, emailConfig *Config) (Client, error) {
	switch impl {
	case Gomail:
		return newGomail(emailConfig), nil
	case Postmark:
		return newPostmark(emailConfig), nil
	default:
		return nil, errors.New("no email implementations found")
	}
}
