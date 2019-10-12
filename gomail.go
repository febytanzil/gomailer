package gomailer

import (
	"context"
	"gopkg.in/gomail.v2"
	"io"
	"log"
	"strings"
	"sync"
	"sync/atomic"
)

type goMail struct {
	sender      gomail.SendCloser
	dialer      *gomail.Dialer
	config      *Config
	isConnected int32
	isClosed    int32
	messagePool chan *poolMessage
	m           sync.Mutex
}

type poolMessage struct {
	future chan futureError
	done   <-chan struct{}
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
		dialer:      gomail.NewPlainDialer(emailConfig.Host, emailConfig.Port, emailConfig.Username, emailConfig.Password),
	}
}

func (h *goMail) SendContext(ctx context.Context, msg *Message) error {
	ftr, err := h.send(ctx, msg)
	if nil != err {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case result := <-ftr:
		return result.err
	}
}

func (h *goMail) SendAsync(msg *Message) error {
	_, err := h.send(context.Background(), msg)
	return err
}

func (h *goMail) Send(msg *Message) error {
	ftr, err := h.send(context.Background(), msg)
	if nil != err {
		return err
	}

	result := <-ftr

	return result.err
}

func (h *goMail) send(ctx context.Context, msg *Message) (chan futureError, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		ftr := make(chan futureError)

		m := &poolMessage{
			future: ftr,
			msg:    msg,
			done:   ctx.Done(),
		}
		err := h.sendToPool(ctx, m)
		if nil != err {
			return nil, err
		}

		return ftr, nil
	}
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
		select {
		case <-task.done:
			close(task.future)
		default:
			msg := task.msg
			m := gomail.NewMessage()
			if "" == msg.From {
				msg.From = h.config.FromEmail
			}
			m.SetHeader("From", msg.From)
			m.SetHeader("To", msg.SendTo...)
			if len(msg.CC) > 0 {
				m.SetHeader("Cc", msg.CC...)
			}
			if len(msg.BCC) > 0 {
				m.SetHeader("Bcc", msg.BCC...)
			}
			m.SetHeader("Subject", msg.Title)
			if "" != strings.TrimSpace(msg.ContentType) {
				m.SetBody(msg.ContentType, msg.Body)
			} else {
				m.SetBody("text/html", msg.Body)
			}

			for _, element := range msg.Attachments {
				fileByte := element.Byte
				m.Attach(element.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
					_, err := w.Write(fileByte)
					return err
				}))
			}

			err := h.dialer.DialAndSend(m)

			go func(e error, task *poolMessage) {
				select {
				case <-task.done:
					log.Println("worker finished but task context was done with err", e)
				default:
					task.future <- futureError{
						err: e,
					}
				}
				close(task.future)
			}(err, task)
		}
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

	dialer := gomail.NewPlainDialer(h.config.Host, h.config.Port, h.config.Username, h.config.Password)
	s, err := dialer.Dial()
	if nil != err {
		return err
	}

	h.sender = s
	atomic.StoreInt32(&h.isConnected, stateConnected)
	go h.listen()

	return nil
}

func (h *goMail) sendToPool(ctx context.Context, task *poolMessage) error {
	/*if stateConnected != atomic.LoadInt32(&h.isConnected) {
		if err := h.connect(); nil != err {
			return err
		}
	}*/

	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			close(task.future)
			return
		default:
			h.messagePool <- task
		}
	}(ctx)

	return nil
}
