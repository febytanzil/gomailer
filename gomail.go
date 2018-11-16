package gomailer

import (
	"gopkg.in/gomail.v2"
	"io"
	"sync"
	"sync/atomic"
)

type goMail struct {
	sender      gomail.SendCloser
	config      *Config
	isConnected int32
	isClosed    int32
	messagePool chan *poolMessage
	m           sync.Mutex
}

type poolMessage struct {
	future chan futureError
	msg    *Message
}

type futureError struct {
	err error
}

const (
	stateDisconnected = iota
	stateConnected
	stateClosed
)

func newGomail(emailConfig *Config) *goMail {
	return &goMail{
		config:      emailConfig,
		messagePool: make(chan *poolMessage),
	}
}

func (h *goMail) Send(msg *Message) error {
	ftr := make(chan futureError)
	defer close(ftr)

	m := &poolMessage{
		future: ftr,
		msg:    msg,
	}
	err := h.sendToPool(m)
	if nil != err {
		return err
	}
	f := <-ftr

	return f.err
}

func (h *goMail) disconnect() error {
	if h.sender == nil {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&h.isConnected, stateConnected, stateDisconnected) {
		return nil
	}

	return h.sender.Close()
}

func (h *goMail) Close() error {
	if h.sender == nil {
		return nil
	}
	h.m.Lock()
	defer h.m.Unlock()

	if !atomic.CompareAndSwapInt32(&h.isClosed, 0, stateClosed) {
		return nil
	}

	if err := h.sender.Close(); nil != err {
		return err
	}

	return nil
}

func (h *goMail) listen() {
	for task := range h.messagePool {
		msg := task.msg
		m := gomail.NewMessage()
		m.SetHeader("From", h.config.Email)
		m.SetHeader("To", msg.SendTo...)
		if len(msg.CC) > 0 {
			m.SetHeader("Cc", msg.CC...)
		}
		m.SetHeader("Subject", msg.Title)
		m.SetBody("text/html", msg.Body)

		for _, element := range msg.Attachments {
			fileByte := element.Byte
			m.Attach(element.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(fileByte)
				return err
			}))
		}

		err := gomail.Send(h.sender, m)
		if err != nil {
			h.disconnect()
		}

		go func(e error, task *poolMessage) {
			task.future <- futureError{
				err: e,
			}
		}(err, task)
	}
}

func (h *goMail) connect() error {
	h.m.Lock()
	defer h.m.Unlock()

	if stateClosed == atomic.LoadInt32(&h.isClosed) {
		return ErrClosed
	}
	if stateConnected == atomic.LoadInt32(&h.isConnected) {
		return nil
	}

	dialer := gomail.NewPlainDialer(h.config.Host, h.config.Port, h.config.Email, h.config.Password)
	s, err := dialer.Dial()
	if nil != err {
		return err
	}

	h.sender = s
	atomic.StoreInt32(&h.isConnected, stateConnected)
	go h.listen()

	return nil
}

func (h *goMail) sendToPool(task *poolMessage) error {
	if stateConnected != atomic.LoadInt32(&h.isConnected) {
		if err := h.connect(); nil != err {
			return err
		}
	}

	go func() {
		h.messagePool <- task
	}()

	return nil
}
