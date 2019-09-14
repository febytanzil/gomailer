package gomailer

import (
	"context"
	"encoding/base64"
	"github.com/keighl/postmark"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type postmarkClient struct {
	pClient  *postmark.Client
	config   *Config
	m        sync.Mutex
	isClosed int32
}

func newPostmark(c *Config) *postmarkClient {
	pc := postmark.NewClient(c.ServerToken, c.AccountToken)
	pc.HTTPClient = &http.Client{Timeout: 10 * time.Second}

	return &postmarkClient{
		pClient: pc,
		config:  c,
	}
}

func (p *postmarkClient) Send(msg *Message) error {
	emails := make([]postmark.Email, 0)
	for _, to := range msg.SendTo {
		attchs := make([]postmark.Attachment, 0)
		if 0 < len(msg.Attachments) {
			for _, att := range msg.Attachments {
				a := postmark.Attachment{
					Content: base64.StdEncoding.EncodeToString(att.Byte),
					Name:    att.Filename,
				}
				attchs = append(attchs, a)
			}
		}

		if "" == msg.From {
			msg.From = p.config.FromEmail
		}
		email := postmark.Email{
			From:        msg.From,
			To:          to,
			Subject:     msg.Title,
			TrackOpens:  true,
			Cc:          strings.Join(msg.CC, ","),
			Bcc:         strings.Join(msg.BCC, ","),
			Attachments: attchs,
		}
		if "" == strings.TrimSpace(msg.ContentType) {
			email.HtmlBody = msg.Body
		} else {
			email.TextBody = msg.Body
		}
		emails = append(emails, email)
	}

	_, err := p.pClient.SendEmailBatch(emails)

	return err
}

func (p *postmarkClient) SendAsync(msg *Message) error {
	return p.Send(msg)
}

func (p *postmarkClient) SendContext(ctx context.Context, msg *Message) error {
	// TODO implement ctx & worker pool
	return p.Send(msg)
}

func (p *postmarkClient) Close() error {
	// TODO for worker pool
	if nil == p.pClient {
		return nil
	}

	p.m.Lock()
	defer p.m.Unlock()

	if !atomic.CompareAndSwapInt32(&p.isClosed, 0, stateClosed) {
		return nil
	}

	p.pClient = nil

	return nil
}
