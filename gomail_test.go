package gomailer

import (
	"context"
	"fmt"
	"gopkg.in/gomail.v2"
	"io"
	"testing"
)

type gomailSenderMock struct {
	gomail.SendCloser
}

func (m *gomailSenderMock) Close() error {
	fmt.Println("client closed")
	return nil
}

func (m *gomailSenderMock) Send(from string, to []string, msg io.WriterTo) error {
	for _, r := range to {
		fmt.Println(from, "send email to ", r)
	}

	return nil
}

func TestGoMail_Send(t *testing.T) {
	c := &goMail{
		sender: &gomailSenderMock{},
		config: &Config{
			Host:      "",
			Port:      587,
			FromEmail: "",
			Username:  "",
			Password:  "",
		},
	}
	err := c.Send(&Message{
		Body:   "body",
		SendTo: []string{"test1@mail.com", "test2@mail.com"},
		Title:  "title",
	})
	if nil == err {
		t.Error("err should not be nil")
	}
}

func TestGoMail_Close(t *testing.T) {
	c := &goMail{
		sender: &gomailSenderMock{},
		config: &Config{
			Host:      "",
			Port:      587,
			FromEmail: "",
			Username:  "",
			Password:  "",
		},
	}
	err := c.Close()
	if nil != err {
		t.Error("err should be nil, but got", err)
	}
}

func TestGoMail_SendContext(t *testing.T) {
	c := &goMail{
		sender: &gomailSenderMock{},
		config: &Config{
			Host:      "",
			Port:      587,
			FromEmail: "",
			Username:  "",
			Password:  "",
		},
	}
	err := c.SendContext(context.Background(), &Message{
		Body:   "body",
		SendTo: []string{"test1@mail.com", "test2@mail.com"},
		Title:  "title",
	})
	if nil == err {
		t.Error("err should not be nil")
	}
}
